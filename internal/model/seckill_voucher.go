// ═════════════════════════════════════════════════════════════════════
// 秒杀优惠券相关模型
// tb_seckill_voucher 是 tb_voucher 的扩展表，存储秒杀专属字段
// SeckillVoucherFull 聚合了主表和扩展表信息，用于前端展示
// ═════════════════════════════════════════════════════════════════════
package model

import "time"

// SeckillVoucher 秒杀优惠券扩展表，对应 tb_seckill_voucher
// 与 Voucher 表是一对一关系（通过 VoucherID 关联）
// InitStock=初始库存, Stock=当前库存（秒杀过程中 Redis 为主, DB 为辅）
type SeckillVoucher struct {
	ID            int64     `gorm:"column:id;primaryKey" json:"id"`
	VoucherID     int64     `gorm:"column:voucher_id" json:"voucherId"`
	InitStock     int       `gorm:"column:init_stock" json:"initStock"`       // 初始库存总量
	Stock         int       `gorm:"column:stock" json:"stock"`                // 当前库存余量（DB 兜底，Redis 实时）
	AllowedLevels string    `gorm:"column:allowed_levels;size:512" json:"allowedLevels"` // 允许购买的用户等级
	MinLevel      int       `gorm:"column:min_level" json:"minLevel"`         // 最低购买等级
	CreateTime    time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	BeginTime     time.Time `gorm:"column:begin_time" json:"beginTime"`       // 秒杀开始时间
	EndTime       time.Time `gorm:"column:end_time" json:"endTime"`           // 秒杀结束时间
	UpdateTime    time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (SeckillVoucher) TableName() string {
	return "tb_seckill_voucher"
}

// SeckillVoucherFull 完整秒杀券视图（Voucher + SeckillVoucher 聚合）
// 被 GetSeckillVoucherFull API 使用，返回前端所需全部信息
type SeckillVoucherFull struct {
	ID            int64     `json:"id"`
	VoucherID     int64     `json:"voucherId"`
	InitStock     int       `json:"initStock"`
	Stock         int       `json:"stock"`
	AllowedLevels string    `json:"allowedLevels"`
	MinLevel      int       `json:"minLevel"`
	CreateTime    time.Time `json:"createTime"`
	BeginTime     time.Time `json:"beginTime"`
	EndTime       time.Time `json:"endTime"`
	Status        int       `json:"status"`  // 取自主表 Voucher.Status
	ShopID        int64     `json:"shopId"`  // 取自主表 Voucher.ShopID
}
