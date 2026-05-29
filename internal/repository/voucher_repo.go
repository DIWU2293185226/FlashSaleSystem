// ═════════════════════════════════════════════════════════════════════
// 优惠券相关仓库 — tb_voucher / tb_seckill_voucher / tb_voucher_order 的分片 CRUD
// 三张表的分片策略各不相同：
//   - voucher: 按 id 分片（intdiv 策略）
//   - seckill_voucher: 按 voucher_id 分片
//   - voucher_order: user_id 决定数据库、voucher_id 决定表
// 秒杀库存扣减使用乐观锁（WHERE stock > 0），避免超卖
// ═════════════════════════════════════════════════════════════════════
package repository

import (
	"github.com/javaup/flashsale-system/internal/model"
	"github.com/javaup/flashsale-system/internal/sharding"
	"gorm.io/gorm"
)

// VoucherRepository tb_voucher 表的数据访问
// 按优惠券 ID 分片（intdiv 策略）
type VoucherRepository struct {
	dm     *DatabaseManager
	router *sharding.Router
}

func NewVoucherRepository(dm *DatabaseManager, router *sharding.Router) *VoucherRepository {
	return &VoucherRepository{dm: dm, router: router}
}

func (r *VoucherRepository) GetByID(id int64) (*model.Voucher, error) {
	db := r.dm.GetDB(r.router.VoucherDB(id))
	table := r.router.VoucherTable(id)
	var v model.Voucher
	err := db.Table(table).Where("id = ?", id).First(&v).Error
	return &v, err
}

func (r *VoucherRepository) Create(v *model.Voucher) error {
	db := r.dm.GetDB(r.router.VoucherDB(v.ID))
	table := r.router.VoucherTable(v.ID)
	return db.Table(table).Create(v).Error
}

func (r *VoucherRepository) Update(v *model.Voucher) error {
	db := r.dm.GetDB(r.router.VoucherDB(v.ID))
	table := r.router.VoucherTable(v.ID)
	return db.Table(table).Where("id = ?", v.ID).Updates(v).Error
}

// GetDB 按数据源名返回 GORM DB 实例，用于跨分片查询（如 ListByShopID 遍历全部分片）
func (r *VoucherRepository) GetDB(name string) *gorm.DB {
	return r.dm.GetDB(name)
}

// SeckillVoucherRepository tb_seckill_voucher 表的数据访问
// 按 voucher_id 分片，与 tb_voucher 共分片键
type SeckillVoucherRepository struct {
	dm     *DatabaseManager
	router *sharding.Router
}

func NewSeckillVoucherRepository(dm *DatabaseManager, router *sharding.Router) *SeckillVoucherRepository {
	return &SeckillVoucherRepository{dm: dm, router: router}
}

func (r *SeckillVoucherRepository) GetByVoucherID(voucherID int64) (*model.SeckillVoucher, error) {
	db := r.dm.GetDB(r.router.SeckillVoucherDB(voucherID))
	table := r.router.SeckillVoucherTable(voucherID)
	var sv model.SeckillVoucher
	err := db.Table(table).Where("voucher_id = ?", voucherID).First(&sv).Error
	return &sv, err
}

func (r *SeckillVoucherRepository) Create(sv *model.SeckillVoucher) error {
	db := r.dm.GetDB(r.router.SeckillVoucherDB(sv.VoucherID))
	table := r.router.SeckillVoucherTable(sv.VoucherID)
	return db.Table(table).Create(sv).Error
}

// UpdateStock 乐观锁扣减库存（stock - 1）
// WHERE stock > 0 保证不会扣到负数，天然防超卖
// 但如果并发高且库存只剩 1 时大量请求同时进来，只有一个能成功
func (r *SeckillVoucherRepository) UpdateStock(voucherID int64) error {
	db := r.dm.GetDB(r.router.SeckillVoucherDB(voucherID))
	table := r.router.SeckillVoucherTable(voucherID)
	return db.Table(table).
		Where("voucher_id = ? AND stock > 0", voucherID).
		Update("stock", gorm.Expr("stock - 1")).Error
}

// UpdateStockByOffset 按偏移量更新库存（正数增加，负数减少）
// 用于回滚场景：订单取消后需要恢复库存
func (r *SeckillVoucherRepository) UpdateStockByOffset(voucherID int64, offset int) error {
	db := r.dm.GetDB(r.router.SeckillVoucherDB(voucherID))
	table := r.router.SeckillVoucherTable(voucherID)
	return db.Table(table).
		Where("voucher_id = ?", voucherID).
		Update("stock", gorm.Expr("stock + ?", offset)).Error
}

// VoucherOrderRepository tb_voucher_order 表的数据访问
// 分片策略特殊：user_id 决定数据库（用户维度的分片），voucher_id 决定表
// 这样同一个用户的所有订单在同一个数据库，方便事务和查询
type VoucherOrderRepository struct {
	dm     *DatabaseManager
	router *sharding.Router
}

func NewVoucherOrderRepository(dm *DatabaseManager, router *sharding.Router) *VoucherOrderRepository {
	return &VoucherOrderRepository{dm: dm, router: router}
}

func (r *VoucherOrderRepository) GetByOrderID(userID int64, voucherID int64, orderID int64) (*model.VoucherOrder, error) {
	db := r.dm.GetDB(r.router.VoucherOrderDB(userID))
	table := r.router.VoucherOrderTable(voucherID)
	var order model.VoucherOrder
	err := db.Table(table).Where("id = ?", orderID).First(&order).Error
	return &order, err
}

func (r *VoucherOrderRepository) Create(order *model.VoucherOrder) error {
	db := r.dm.GetDB(r.router.VoucherOrderDB(order.UserID))
	table := r.router.VoucherOrderTable(order.VoucherID)
	return db.Table(table).Create(order).Error
}

func (r *VoucherOrderRepository) Update(order *model.VoucherOrder) error {
	db := r.dm.GetDB(r.router.VoucherOrderDB(order.UserID))
	table := r.router.VoucherOrderTable(order.VoucherID)
	return db.Table(table).Where("id = ?", order.ID).Updates(order).Error
}

// GetByUserIDAndVoucherID 查询用户是否已购买过此优惠券（一人一单校验）
// 如果返回记录存在，说明用户已经买过，不能重复秒杀
func (r *VoucherOrderRepository) GetByUserIDAndVoucherID(userID, voucherID int64) (*model.VoucherOrder, error) {
	db := r.dm.GetDB(r.router.VoucherOrderDB(userID))
	table := r.router.VoucherOrderTable(voucherID)
	var order model.VoucherOrder
	err := db.Table(table).Where("user_id = ? AND voucher_id = ?", userID, voucherID).First(&order).Error
	return &order, err
}
