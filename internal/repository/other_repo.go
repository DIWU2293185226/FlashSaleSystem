// ═════════════════════════════════════════════════════════════════════
// 扩展仓库 — 对账日志/回滚失败记录/订阅候补
// 这些表属于辅助功能模块（库存对账、异常回滚、缺货登记），
// 数据量相对较小，统一使用广播表方式存放在 hmdp_0
// ═════════════════════════════════════════════════════════════════════
package repository

import (
	"github.com/javaup/flashsale-system/internal/model"
	"github.com/javaup/flashsale-system/internal/sharding"
)

// VoucherOrderRouterRepository tb_voucher_order_router 表的数据访问
// 存储订单路由信息，仅作为分片路由的持久化参考，实际路由逻辑由 sharding.Router 完成
type VoucherOrderRouterRepository struct {
	dm     *DatabaseManager
	router *sharding.Router
}

func NewVoucherOrderRouterRepository(dm *DatabaseManager, router *sharding.Router) *VoucherOrderRouterRepository {
	return &VoucherOrderRouterRepository{dm: dm, router: router}
}

// VoucherReconcileLogRepository tb_voucher_reconcile_log 表的数据访问
// 库存对账日志：记录每次对账操作的结果，用于追踪 Redis 和 DB 之间的库存一致性
type VoucherReconcileLogRepository struct {
	dm     *DatabaseManager
	router *sharding.Router
}

func NewVoucherReconcileLogRepository(dm *DatabaseManager, router *sharding.Router) *VoucherReconcileLogRepository {
	return &VoucherReconcileLogRepository{dm: dm, router: router}
}

// Create 写入对账日志
func (r *VoucherReconcileLogRepository) Create(log *model.VoucherReconcileLog) error {
	db := r.dm.GetDB(broadcastDB)
	return db.Create(log).Error
}

// GetByVoucherID 查询指定优惠券的所有对账记录，按时间倒序
func (r *VoucherReconcileLogRepository) GetByVoucherID(voucherID int64) ([]model.VoucherReconcileLog, error) {
	db := r.dm.GetDB(broadcastDB)
	var logs []model.VoucherReconcileLog
	err := db.Where("voucher_id = ?", voucherID).Order("create_time DESC").Find(&logs).Error
	return logs, err
}

// RollbackFailureLogRepository tb_rollback_failure_log 表的数据访问
// 秒杀回滚失败记录：当 Kafka 消费失败导致库存回滚无法完成时，
// 将失败信息记录到此表，供运维手工或定时任务补偿处理
type RollbackFailureLogRepository struct {
	dm *DatabaseManager
}

func NewRollbackFailureLogRepository(dm *DatabaseManager) *RollbackFailureLogRepository {
	return &RollbackFailureLogRepository{dm: dm}
}

func (r *RollbackFailureLogRepository) Create(log *model.RollbackFailureLog) error {
	db := r.dm.GetDB(broadcastDB)
	return db.Create(log).Error
}

// ListByVoucherID 查询某优惠券的回滚失败记录，分页倒序
func (r *RollbackFailureLogRepository) ListByVoucherID(voucherID int64, offset, limit int) ([]model.RollbackFailureLog, error) {
	db := r.dm.GetDB(broadcastDB)
	var logs []model.RollbackFailureLog
	err := db.Where("voucher_id = ?", voucherID).
		Order("create_time DESC").
		Offset(offset).Limit(limit).
		Find(&logs).Error
	return logs, err
}

// SubscribeRepository tb_voucher_subscribe 表的数据访问
// 订阅候补登记：当秒杀券售罄后，用户可以"缺货登记"，
// 一旦有库存释放（退单），系统会按登记顺序通知用户
type SubscribeRepository struct {
	dm *DatabaseManager
}

func NewSubscribeRepository(dm *DatabaseManager) *SubscribeRepository {
	return &SubscribeRepository{dm: dm}
}

func (r *SubscribeRepository) Create(sub *model.VoucherSubscribe) error {
	db := r.dm.GetDB(broadcastDB)
	return db.Create(sub).Error
}

// GetByUserIDAndVoucherID 查询用户是否已登记某优惠券的候补
// 用于防止重复登记
func (r *SubscribeRepository) GetByUserIDAndVoucherID(userID, voucherID int64) (*model.VoucherSubscribe, error) {
	db := r.dm.GetDB(broadcastDB)
	var sub model.VoucherSubscribe
	err := db.Where("user_id = ? AND voucher_id = ?", userID, voucherID).First(&sub).Error
	return &sub, err
}

// ListByUserID 分页查询用户的候补登记历史
func (r *SubscribeRepository) ListByUserID(userID int64, page, size int) ([]model.VoucherSubscribe, int64, error) {
	db := r.dm.GetDB(broadcastDB)
	var subs []model.VoucherSubscribe
	var total int64
	query := db.Model(&model.VoucherSubscribe{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * size
	err := query.Order("create_time DESC").Offset(offset).Limit(size).Find(&subs).Error
	return subs, total, err
}
