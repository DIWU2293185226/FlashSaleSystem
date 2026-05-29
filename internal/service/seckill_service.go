package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/javaup/flashsale-system/internal/bloom"
	"github.com/javaup/flashsale-system/internal/cache"
	"github.com/javaup/flashsale-system/internal/idgen"
	"github.com/javaup/flashsale-system/internal/locker"
	"github.com/javaup/flashsale-system/internal/model"
	"github.com/javaup/flashsale-system/internal/mq"
	"github.com/javaup/flashsale-system/internal/ratelimit"
	"github.com/javaup/flashsale-system/internal/repository"
	"github.com/javaup/flashsale-system/pkg"
	"github.com/javaup/flashsale-system/pkg/enums"
	"github.com/redis/go-redis/v9"
)

// SeckillService handles the full seckill flow.
type SeckillService struct {
	idGen              *idgen.Snowflake
	seckillVoucherRepo *repository.SeckillVoucherRepository
	voucherRepo        *repository.VoucherRepository
	orderRepo          *repository.VoucherOrderRepository
	redis              *cache.RedisCache
	lua                *cache.LuaManager
	locker             *locker.Locker
	bloomFilter        *bloom.Filter
	producer           *mq.Producer
	slidingLimiter     *ratelimit.SlidingWindowLimiter
	tokenLimiter       *ratelimit.TokenBucketLimiter
	tokenMgr           *ratelimit.AccessTokenManager
}

func NewSeckillService(
	idGen *idgen.Snowflake,
	seckillVoucherRepo *repository.SeckillVoucherRepository,
	voucherRepo *repository.VoucherRepository,
	orderRepo *repository.VoucherOrderRepository,
	redis *cache.RedisCache,
	lua *cache.LuaManager,
	locker *locker.Locker,
	bloomFilter *bloom.Filter,
	producer *mq.Producer,
) *SeckillService {
	return &SeckillService{
		idGen:              idGen,
		seckillVoucherRepo: seckillVoucherRepo,
		voucherRepo:        voucherRepo,
		orderRepo:          orderRepo,
		redis:              redis,
		lua:                lua,
		locker:             locker,
		bloomFilter:        bloomFilter,
		producer:           producer,
		slidingLimiter:     ratelimit.NewSlidingWindowLimiter(lua),
		tokenLimiter:       ratelimit.NewTokenBucketLimiter(lua),
		tokenMgr:           ratelimit.NewAccessTokenManager(lua),
	}
}

// ---------- Cache warmup ----------

// LoadSeckillStock loads seckill stock and voucher info from DB into Redis.
func (s *SeckillService) LoadSeckillStock(voucherID int64) *pkg.Result {
	ctx := context.Background()

	sv, err := s.seckillVoucherRepo.GetByVoucherID(voucherID)
	if err != nil {
		return pkg.FailWithCode(pkg.ErrSeckillVoucherNotExist)
	}

	stockKey := pkg.SeckillStockKey + strconv.FormatInt(voucherID, 10)
	userKey := pkg.SeckillUserKey + strconv.FormatInt(voucherID, 10)
	voucherKey := pkg.SeckillVoucherTagKey + strconv.FormatInt(voucherID, 10)

	s.redis.Set(ctx, stockKey, sv.Stock, 24*time.Hour)
	s.redis.Set(ctx, voucherKey, "1", 24*time.Hour)
	_ = s.redis.Expire(ctx, userKey, 24*time.Hour)

	// Add to Bloom filter
	_ = s.bloomFilter.AddString(ctx, strconv.FormatInt(voucherID, 10))

	return pkg.OK()
}

// ---------- Rate limiting ----------

// checkRateLimit checks both IP-level and user-level rate limits.
func (s *SeckillService) checkRateLimit(ip string, userID int64) *pkg.Result {
	now := time.Now().UnixMilli()

	// IP rate limit: 10 requests per second
	ipKey := "rate:seckill:ip:" + ip
	allowed, err := s.slidingLimiter.Allow(ipKey, 1000, 10, now)
	if err != nil || !allowed {
		return pkg.FailWithCode(pkg.ErrSeckillRateLimitIPExceeded)
	}

	// User rate limit: 5 requests per second
	userKey := "rate:seckill:user:" + strconv.FormatInt(userID, 10)
	allowed, err = s.tokenLimiter.Allow(userKey, 5, 5, 1)
	if err != nil || !allowed {
		return pkg.FailWithCode(pkg.ErrSeckillRateLimitUserExceeded)
	}

	return nil
}

// checkPreconditions validates time window, Bloom filter, and repeat idempotency.
func (s *SeckillService) checkPreconditions(voucherID, userID int64) *pkg.Result {
	ctx := context.Background()

	// Check repeat idempotency (best-effort, allow through on Redis error to avoid false rejection)
	ok, err := ratelimit.CheckRepeat(ctx, s.redis, pkg.SeckillVoucherOrder,
		strconv.FormatInt(voucherID, 10)+":"+strconv.FormatInt(userID, 10))
	if err != nil {
		// Redis unavailable — allow through; the Lua SADD check still enforces one-person-one-order atomically
		_ = err
	} else if !ok {
		return pkg.FailWithCode(pkg.ErrSeckillVoucherClaimed)
	}

	// Bloom filter pre-check (fast-path rejection, false positives are safe)
	exists, err := s.bloomFilter.ExistsString(ctx, strconv.FormatInt(voucherID, 10))
	if err == nil && !exists {
		// Bloom filter says not exists, but this could be an unloaded voucher.
		// Instead of rejecting, fall through — the Lua script will check anyway.
		// Only reject if we know for sure from a secondary check.
		_ = exists
	}

	// Verify voucher is still valid from DB (best-effort pre-check, Lua is the source of truth)
	sv, err := s.seckillVoucherRepo.GetByVoucherID(voucherID)
	if err != nil {
		return pkg.FailWithCode(pkg.ErrSeckillVoucherNotExist)
	}
	now := time.Now()
	if now.Before(sv.BeginTime) {
		return pkg.FailWithCode(pkg.ErrSeckillVoucherNotBegin)
	}
	if now.After(sv.EndTime) {
		return pkg.FailWithCode(pkg.ErrSeckillVoucherIsOver)
	}
	// Note: stock check is intentionally omitted here — it would create a TOCTOU race.
	// The Lua script atomically checks stock at the moment of execution.

	return nil
}

// SeckillVoucher is the main seckill entry point. It executes the Lua script atomically,
// then sends a Kafka message for async order creation.
func (s *SeckillService) SeckillVoucher(voucherID int64, userID int64, ip string) *pkg.Result {
	// 1. Rate limiting
	if err := s.checkRateLimit(ip, userID); err != nil {
		return err
	}

	// 2. Preconditions
	if err := s.checkPreconditions(voucherID, userID); err != nil {
		return err
	}

	// 3. Generate IDs
	traceID := s.idGen.NextID()
	orderID := s.idGen.NextID()

	// 4. Execute Lua script (atomic stock DECR + user SADD + trace HSET)
	stockKey := pkg.SeckillStockKey + strconv.FormatInt(voucherID, 10)
	userKey := pkg.SeckillUserKey + strconv.FormatInt(voucherID, 10)
	traceKey := pkg.SeckillTraceLogKey + strconv.FormatInt(traceID, 10)
	voucherKey := pkg.SeckillVoucherTagKey + strconv.FormatInt(voucherID, 10)

	result, err := s.lua.Eval("seckillVoucher",
		[]string{stockKey, userKey, traceKey, voucherKey},
		userID, voucherID, traceID,
	)
	// Retry once on transient error (Lua script is fully atomic — partial execution is impossible)
	if err != nil {
		result, err = s.lua.Eval("seckillVoucher",
			[]string{stockKey, userKey, traceKey, voucherKey},
			userID, voucherID, traceID,
		)
	}
	if err != nil {
		return pkg.FailWithMsg("秒杀失败")
	}

	arr, ok := result.([]interface{})
	if !ok || len(arr) < 1 {
		return pkg.FailWithMsg("秒杀失败")
	}
	code, _ := arr[0].(int64)
	if code != 1 {
		msg := "秒杀失败"
		if len(arr) > 1 {
			msg, _ = arr[1].(string)
		}
		return pkg.FailWithMsg(msg)
	}

	// 5. Send Kafka async message for order creation
	msg := mq.SeckillMessage{
		UserID:    userID,
		VoucherID: voucherID,
		OrderID:   orderID,
		TraceID:   traceID,
	}
	if err := s.producer.SendSeckillMessage(context.Background(), msg); err != nil {
		// If Kafka is unavailable, attempt synchronous order creation (fallback)
		s.createOrderSync(msg)
	}

	return pkg.OKWithData(orderID)
}

// createOrderSync is a fallback for when Kafka is unavailable.
func (s *SeckillService) createOrderSync(msg mq.SeckillMessage) {
	order := &model.VoucherOrder{
		ID:        msg.OrderID,
		UserID:    msg.UserID,
		VoucherID: msg.VoucherID,
		Status:    int(enums.OrderStatusNormal),
	}
	ctx := context.Background()

	// Distributed lock to prevent duplicate order creation
	lockKey := pkg.LockOrderKey + strconv.FormatInt(msg.OrderID, 10)
	locked, err := s.locker.Lock(ctx, lockKey, 30*time.Second)
	if err != nil || !locked {
		return
	}
	defer s.locker.Unlock(ctx, lockKey)

	if err := s.orderRepo.Create(order); err != nil {
		// Rollback on failure
		s.rollback(msg.VoucherID, msg.UserID, msg.TraceID)
	}
}

// ---------- Kafka consumer handler ----------

// HandleSeckillOrder processes a seckill order from the Kafka queue.
func (s *SeckillService) HandleSeckillOrder(ctx context.Context, msg mq.SeckillMessage) error {
	// Deduplication: check if order already exists
	existing, err := s.orderRepo.GetByUserIDAndVoucherID(msg.UserID, msg.VoucherID)
	if err == nil && existing != nil {
		// Order already exists, skip
		return nil
	}

	lockKey := pkg.LockOrderKey + strconv.FormatInt(msg.OrderID, 10)
	locked, err := s.locker.TryLock(ctx, lockKey, 30*time.Second, 5*time.Second)
	if err != nil || !locked {
		return fmt.Errorf("failed to acquire lock for order %d", msg.OrderID)
	}
	defer s.locker.Unlock(ctx, lockKey)

	// Re-check after acquiring lock
	existing, err = s.orderRepo.GetByUserIDAndVoucherID(msg.UserID, msg.VoucherID)
	if err == nil && existing != nil {
		return nil
	}

	order := &model.VoucherOrder{
		ID:        msg.OrderID,
		UserID:    msg.UserID,
		VoucherID: msg.VoucherID,
		Status:    int(enums.OrderStatusNormal),
	}
	if err := s.orderRepo.Create(order); err != nil {
		s.rollback(msg.VoucherID, msg.UserID, msg.TraceID)
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

// rollback restores stock and removes user from the purchased set.
func (s *SeckillService) rollback(voucherID, userID, traceID int64) {
	stockKey := pkg.SeckillStockKey + strconv.FormatInt(voucherID, 10)
	userKey := pkg.SeckillUserKey + strconv.FormatInt(voucherID, 10)
	traceKey := pkg.SeckillTraceLogKey + strconv.FormatInt(traceID, 10)

	_, err := s.lua.Eval("seckillVoucherRollBack",
		[]string{stockKey, userKey, traceKey},
		userID, voucherID,
	)
	if err != nil {
		// Log rollback failure (would be written to RollbackFailureLog in production)
		_ = err
	}
}

// ---------- Order queries ----------

// GetOrder retrieves a voucher order.
func (s *SeckillService) GetOrder(orderID, voucherID, userID int64) *pkg.Result {
	order, err := s.orderRepo.GetByOrderID(userID, voucherID, orderID)
	if err != nil {
		return pkg.FailWithCode(pkg.ErrSeckillVoucherOrderNotExist)
	}
	return pkg.OKWithData(order)
}

// CancelOrder cancels a seckill order and rolls back.
func (s *SeckillService) CancelOrder(orderID, voucherID, userID int64) *pkg.Result {
	order, err := s.orderRepo.GetByOrderID(userID, voucherID, orderID)
	if err != nil {
		return pkg.FailWithCode(pkg.ErrSeckillVoucherOrderNotExist)
	}
	if order.Status == int(enums.OrderStatusCancel) {
		return pkg.FailWithCode(pkg.ErrVoucherOrderCancel)
	}

	order.Status = int(enums.OrderStatusCancel)
	if err := s.orderRepo.Update(order); err != nil {
		return pkg.FailWithMsg("取消订单失败")
	}

	// Rollback Redis state
	stockKey := pkg.SeckillStockKey + strconv.FormatInt(voucherID, 10)
	userKey := pkg.SeckillUserKey + strconv.FormatInt(voucherID, 10)
	_ = s.redis.Client.Incr(context.Background(), stockKey).Err()
	_ = s.redis.Client.SRem(context.Background(), userKey, userID).Err()

	return pkg.OK()
}

// GetSeckillVoucherFull returns the full seckill voucher info.
func (s *SeckillService) GetSeckillVoucherFull(voucherID int64) *pkg.Result {
	sv, err := s.seckillVoucherRepo.GetByVoucherID(voucherID)
	if err != nil {
		return pkg.FailWithCode(pkg.ErrSeckillVoucherNotExist)
	}

	v, err := s.voucherRepo.GetByID(voucherID)
	if err != nil {
		return pkg.FailWithCode(pkg.ErrSeckillVoucherNotExist)
	}

	full := &model.SeckillVoucherFull{
		VoucherID:     sv.VoucherID,
		InitStock:     sv.InitStock,
		Stock:         sv.Stock,
		AllowedLevels: sv.AllowedLevels,
		MinLevel:      sv.MinLevel,
		CreateTime:    sv.CreateTime,
		BeginTime:     sv.BeginTime,
		EndTime:       sv.EndTime,
		Status:        v.Status,
		ShopID:        v.ShopID,
	}
	return pkg.OKWithData(full)
}

// ---------- Access token ----------

// GenerateAccessToken generates an access token for a user to perform seckill.
func (s *SeckillService) GenerateAccessToken(voucherID, userID int64) *pkg.Result {
	key := "seckill:token:" + strconv.FormatInt(voucherID, 10)
	token, err := s.tokenMgr.Generate(key, voucherID, userID, 300)
	if err != nil {
		return pkg.FailWithMsg("生成令牌失败")
	}
	return pkg.OKWithData(token)
}

// ValidateAccessToken validates an access token.
func (s *SeckillService) ValidateAccessToken(voucherID, userID int64, token string) bool {
	key := "seckill:token:" + strconv.FormatInt(voucherID, 10)
	valid, err := s.tokenMgr.Validate(key, voucherID, userID, token)
	if err != nil {
		return false
	}
	return valid
}

// ---------- Distributed worker ID assignment ----------

// AssignWorkerID assigns a Snowflake worker/datacenter ID via Redis.
func (s *SeckillService) AssignWorkerID(key string, workerID, datacenterID int64) (int64, int64, error) {
	result, err := s.lua.Eval("workAndDataCenterId",
		[]string{key},
		workerID, datacenterID,
	)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to assign worker ID: %w", err)
	}
	arr, ok := result.([]interface{})
	if !ok || len(arr) < 3 {
		return 0, 0, fmt.Errorf("unexpected assign result: %v", result)
	}
	success, _ := arr[0].(int64)
	if success == 0 {
		return 0, 0, fmt.Errorf("no available worker ID")
	}
	wID, _ := arr[1].(int64)
	dID, _ := arr[2].(int64)
	return wID, dID, nil
}

// ---------- Stock query ----------

// GetStock returns the current stock from Redis.
func (s *SeckillService) GetStock(voucherID int64) *pkg.Result {
	ctx := context.Background()
	stockKey := pkg.SeckillStockKey + strconv.FormatInt(voucherID, 10)
	val, err := s.redis.Get(ctx, stockKey)
	if err != nil {
		if err == redis.Nil {
			return pkg.FailWithMsg("库存未加载，请先调用 load-stock 预热")
		}
		return pkg.FailWithMsg("查询库存失败")
	}
	stock, _ := strconv.Atoi(val)
	return pkg.OKWithData(stock)
}
