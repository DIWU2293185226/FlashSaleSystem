// ═════════════════════════════════════════════════════════════════════
// 优惠券订单模型
// 秒杀成功后由 Kafka 消费者异步创建订单记录
// PayTime/UseTime/RefundTime 使用 *time.Time 指针类型，未发生时为 nil
// ═════════════════════════════════════════════════════════════════════
package model

import "time"

// VoucherOrder 优惠券订单表，对应 tb_voucher_order（分片表）
// 秒杀全流程的核心产出：库存扣减成功 → 异步创建此订单
type VoucherOrder struct {
	ID                   int64      `gorm:"column:id;primaryKey" json:"id"`
	UserID               int64      `gorm:"column:user_id" json:"userId"`
	VoucherID            int64      `gorm:"column:voucher_id" json:"voucherId"`
	PayType              int        `gorm:"column:pay_type" json:"payType"`
	Status               int        `gorm:"column:status" json:"status"`                              // 1:正常 2:取消
	ReconciliationStatus int        `gorm:"column:reconciliation_status" json:"reconciliationStatus"` // 1:待处理
	CreateTime           time.Time  `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	PayTime              *time.Time `gorm:"column:pay_time" json:"payTime"`       // 支付时间（未支付时为 nil）
	UseTime              *time.Time `gorm:"column:use_time" json:"useTime"`       // 使用时间（未使用时为 nil）
	RefundTime           *time.Time `gorm:"column:refund_time" json:"refundTime"` // 退款时间（未退款时为 nil）
	UpdateTime           time.Time  `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (VoucherOrder) TableName() string {
	return "tb_voucher_order"
}
