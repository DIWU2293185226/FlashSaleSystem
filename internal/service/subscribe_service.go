package service

import (
	"context"
	"strconv"
	"time"

	"github.com/javaup/flashsale-system/internal/cache"
	"github.com/javaup/flashsale-system/internal/repository"
	"github.com/javaup/flashsale-system/pkg"
	"github.com/javaup/flashsale-system/pkg/enums"
)

// SubscribeService handles voucher subscribe/waitlist functionality.
type SubscribeService struct {
	subscribeRepo *repository.SubscribeRepository
	seckillRepo   *repository.SeckillVoucherRepository
	orderRepo     *repository.VoucherOrderRepository
	redis         *cache.RedisCache
}

func NewSubscribeService(
	subscribeRepo *repository.SubscribeRepository,
	seckillRepo *repository.SeckillVoucherRepository,
	orderRepo *repository.VoucherOrderRepository,
	redis *cache.RedisCache,
) *SubscribeService {
	return &SubscribeService{
		subscribeRepo: subscribeRepo,
		seckillRepo:   seckillRepo,
		orderRepo:     orderRepo,
		redis:         redis,
	}
}

// Subscribe adds a user to the waitlist for a voucher. When stock becomes available,
// subscribed users get notified/auto-created orders.
func (s *SubscribeService) Subscribe(voucherID, userID int64) *pkg.Result {
	ctx := context.Background()
	userKey := pkg.SeckillSubscribeUserKey + strconv.FormatInt(voucherID, 10)
	zsetKey := pkg.SeckillSubscribeZSetKey

	// Check if already subscribed
	isMember, err := s.redis.SIsMember(ctx, userKey, userID)
	if err == nil && isMember {
		return pkg.OKWithData("已订阅")
	}

	// Add to user set
	if _, err := s.redis.SAdd(ctx, userKey, userID); err != nil {
		return pkg.FailWithMsg("订阅失败")
	}

	// Add to global ZSet with timestamp for ordering
	_ = s.redis.ZAdd(ctx, zsetKey, strconv.FormatInt(voucherID, 10)+":"+strconv.FormatInt(userID, 10), float64(time.Now().UnixMilli()))

	// Subscribe status in Redis
	statusKey := pkg.SeckillSubscribeStatusKey + strconv.FormatInt(voucherID, 10) + ":" + strconv.FormatInt(userID, 10)
	_ = s.redis.Set(ctx, statusKey, strconv.Itoa(int(enums.SubscribeStatusSubscribed)), 72*time.Hour)

	_ = s.redis.Expire(ctx, userKey, 72*time.Hour)

	return pkg.OKWithData("订阅成功")
}

// Unsubscribe removes a user from the waitlist.
func (s *SubscribeService) Unsubscribe(voucherID, userID int64) *pkg.Result {
	ctx := context.Background()
	userKey := pkg.SeckillSubscribeUserKey + strconv.FormatInt(voucherID, 10)
	zsetKey := pkg.SeckillSubscribeZSetKey

	_, _ = s.redis.SRem(ctx, userKey, userID)
	_ = s.redis.ZAdd(ctx, zsetKey, strconv.FormatInt(voucherID, 10)+":"+strconv.FormatInt(userID, 10), float64(time.Now().UnixMilli()))

	statusKey := pkg.SeckillSubscribeStatusKey + strconv.FormatInt(voucherID, 10) + ":" + strconv.FormatInt(userID, 10)
	_ = s.redis.Set(ctx, statusKey, strconv.Itoa(int(enums.SubscribeStatusUnsubscribed)), 72*time.Hour)

	return pkg.OKWithData("取消订阅成功")
}

// GetSubscribeStatus returns the subscribe status for a user.
func (s *SubscribeService) GetSubscribeStatus(voucherID, userID int64) *pkg.Result {
	ctx := context.Background()
	statusKey := pkg.SeckillSubscribeStatusKey + strconv.FormatInt(voucherID, 10) + ":" + strconv.FormatInt(userID, 10)
	val, err := s.redis.Get(ctx, statusKey)
	if err != nil {
		return pkg.OKWithData(enums.SubscribeStatusUnsubscribed.Code())
	}
	status, _ := strconv.Atoi(val)
	return pkg.OKWithData(status)
}

// GetSubscribedUsers returns all subscribed user IDs for a voucher.
func (s *SubscribeService) GetSubscribedUsers(voucherID int64) ([]int64, error) {
	ctx := context.Background()
	userKey := pkg.SeckillSubscribeUserKey + strconv.FormatInt(voucherID, 10)
	members, err := s.redis.Client.SMembers(ctx, userKey).Result()
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(members))
	for _, m := range members {
		id, _ := strconv.ParseInt(m, 10, 64)
		if id > 0 {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// NotifySubscribers notifies all subscribed users that stock is available.
// This would typically be called after stock is restocked.
// Currently counts eligible users — a full implementation would push notifications
// or create orders via the seckill flow.
func (s *SubscribeService) NotifySubscribers(voucherID int64) *pkg.Result {
	users, err := s.GetSubscribedUsers(voucherID)
	if err != nil {
		return pkg.FailWithMsg("获取订阅用户失败")
	}

	sv, err := s.seckillRepo.GetByVoucherID(voucherID)
	if err != nil || sv.Stock <= 0 {
		return pkg.FailWithCode(pkg.ErrSeckillVoucherStockInsufficient)
	}

	eligible := int64(0)
	for _, userID := range users {
		// Check one-person-one-order
		existing, _ := s.orderRepo.GetByUserIDAndVoucherID(userID, voucherID)
		if existing != nil {
			// Already purchased, skip
			_ = s.Unsubscribe(voucherID, userID)
			continue
		}
		if sv.Stock <= 0 {
			break
		}
		// TODO: In production, push notification or auto-create order via seckill flow
		eligible++
		sv.Stock--
	}

	return pkg.OKWithData(map[string]int64{
		"subscribed": int64(len(users)),
		"eligible":   eligible,
	})
}

// GetSubscribeHistory returns the subscription history from the database.
func (s *SubscribeService) GetSubscribeHistory(userID int64, page, size int) *pkg.Result {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 50 {
		size = pkg.MaxPageSize
	}
	records, total, err := s.subscribeRepo.ListByUserID(userID, page, size)
	if err != nil {
		return pkg.FailWithMsg("获取订阅记录失败")
	}
	return pkg.OKWithDataTotal(records, total)
}
