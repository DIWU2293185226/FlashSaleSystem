// ═════════════════════════════════════════════════════════════════════
// 优惠券数据模型
// Type 字段区分普通券和秒杀券，秒杀券关联 SeckillVoucher 扩展表
// Stock/BeginTime/EndTime 为秒杀券特有字段，通过关联查询填充
// ═════════════════════════════════════════════════════════════════════
package model

import "time"

// Voucher 优惠券主表，对应 tb_voucher
// 秒杀券的库存和时间字段存储在 tb_seckill_voucher 扩展表中
type Voucher struct {
	ID          int64     `gorm:"column:id;primaryKey" json:"id"`
	ShopID      int64     `gorm:"column:shop_id" json:"shopId"`
	Title       string    `gorm:"column:title;size:255" json:"title"`
	SubTitle    string    `gorm:"column:sub_title;size:255" json:"subTitle"`
	Rules       string    `gorm:"column:rules;size:1024" json:"rules"`
	PayValue    int64     `gorm:"column:pay_value" json:"payValue"`       // 支付金额（分）
	ActualValue int64     `gorm:"column:actual_value" json:"actualValue"` // 实际价值（分）
	Type        int       `gorm:"column:type" json:"type"`                // 0:普通券 1:秒杀券
	Status      int       `gorm:"column:status" json:"status"`            // 1:上架 2:下架 3:过期
	CreateTime  time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime  time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`

	// 秒杀券扩展字段（非数据库直存，由 service 层拼装）
	Stock     int       `gorm:"-" json:"stock,omitempty"`
	BeginTime time.Time `gorm:"-" json:"beginTime,omitempty"`
	EndTime   time.Time `gorm:"-" json:"endTime,omitempty"`
}

func (Voucher) TableName() string {
	return "tb_voucher"
}
