package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/javaup/flashsale-system/internal/cache"
	"github.com/javaup/flashsale-system/internal/model"
	"github.com/javaup/flashsale-system/internal/repository"
	"github.com/javaup/flashsale-system/pkg"
)

// ReconciliationResult holds the result of a reconciliation check.
type ReconciliationResult struct {
	VoucherID  int64 `json:"voucherId"`
	RedisStock int   `json:"redisStock"`
	DBStock    int   `json:"dbStock"`
	Matched    bool  `json:"matched"`
}

// ReconciliationService detects and fixes inconsistencies between Redis and DB.
type ReconciliationService struct {
	seckillVoucherRepo *repository.SeckillVoucherRepository
	reconcileLogRepo   *repository.VoucherReconcileLogRepository
	redis              *cache.RedisCache
}

func NewReconciliationService(
	seckillVoucherRepo *repository.SeckillVoucherRepository,
	reconcileLogRepo *repository.VoucherReconcileLogRepository,
	redis *cache.RedisCache,
) *ReconciliationService {
	return &ReconciliationService{
		seckillVoucherRepo: seckillVoucherRepo,
		reconcileLogRepo:   reconcileLogRepo,
		redis:              redis,
	}
}

// CheckStock compares Redis stock with DB stock for a specific voucher.
func (s *ReconciliationService) CheckStock(voucherID int64) (*ReconciliationResult, error) {
	ctx := context.Background()

	sv, err := s.seckillVoucherRepo.GetByVoucherID(voucherID)
	if err != nil {
		return nil, fmt.Errorf("failed to get seckill voucher %d: %w", voucherID, err)
	}

	stockKey := pkg.SeckillStockKey + strconv.FormatInt(voucherID, 10)
	redisVal, err := s.redis.Get(ctx, stockKey)
	redisStock := sv.Stock // default to DB stock if Redis key doesn't exist
	if err == nil {
		if val, parseErr := strconv.Atoi(redisVal); parseErr == nil {
			redisStock = val
		}
	}

	matched := redisStock == sv.Stock
	return &ReconciliationResult{
		VoucherID:  voucherID,
		RedisStock: redisStock,
		DBStock:    sv.Stock,
		Matched:    matched,
	}, nil
}

// FixStock fixes the Redis stock to match the DB stock for a voucher.
func (s *ReconciliationService) FixStock(voucherID int64) error {
	ctx := context.Background()

	sv, err := s.seckillVoucherRepo.GetByVoucherID(voucherID)
	if err != nil {
		return fmt.Errorf("failed to get seckill voucher %d: %w", voucherID, err)
	}

	stockKey := pkg.SeckillStockKey + strconv.FormatInt(voucherID, 10)
	beforeStr, _ := s.redis.Get(ctx, stockKey)
	beforeQty, _ := strconv.Atoi(beforeStr)

	// Set Redis stock to DB stock
	if err := s.redis.Set(ctx, stockKey, sv.Stock, 0); err != nil {
		return fmt.Errorf("failed to fix stock for voucher %d: %w", voucherID, err)
	}

	afterQty := sv.Stock
	changeQty := afterQty - beforeQty

	// Log the reconciliation
	log := &model.VoucherReconcileLog{
		VoucherID:           voucherID,
		Detail:              fmt.Sprintf("stock fixed: %d -> %d (change: %d)", beforeQty, afterQty, changeQty),
		BeforeQty:           beforeQty,
		ChangeQty:           changeQty,
		AfterQty:            afterQty,
		LogType:             1, // stock fix
		BusinessType:        1, // seckill
		ReconciliationStatus: 1,
	}
	_ = s.reconcileLogRepo.Create(log)

	return nil
}

// CheckAllActive checks all active seckill vouchers for consistency.
func (s *ReconciliationService) CheckAllActive(voucherIDs []int64) ([]*ReconciliationResult, error) {
	var results []*ReconciliationResult
	for _, vid := range voucherIDs {
		result, err := s.CheckStock(vid)
		if err != nil {
			continue
		}
		results = append(results, result)
	}
	return results, nil
}

// LogInconsistency logs a detected inconsistency without fixing it.
func (s *ReconciliationService) LogInconsistency(voucherID int64, redisStock, dbStock int, detail string) {
	log := &model.VoucherReconcileLog{
		VoucherID:           voucherID,
		Detail:              detail,
		BeforeQty:           redisStock,
		ChangeQty:           dbStock - redisStock,
		AfterQty:            dbStock,
		LogType:             2, // inconsistency detected
		BusinessType:        1, // seckill
		ReconciliationStatus: 1,
	}
	_ = s.reconcileLogRepo.Create(log)
}
