// ═════════════════════════════════════════════════════════════════════
// 优惠券相关 DTO
// SeckillVoucherDto 是创建秒杀券的请求体
// 订阅/提醒 DTO 用于候补和延迟通知功能
// ═════════════════════════════════════════════════════════════════════
package dto

import "time"

// SeckillVoucherDto 创建秒杀优惠券请求体
// 同时写入 tb_voucher（基础信息）和 tb_seckill_voucher（秒杀扩展信息）
type SeckillVoucherDto struct {
	ShopID        int64     `json:"shopId" binding:"required"`
	Title         string    `json:"title" binding:"required"`
	SubTitle      string    `json:"subTitle"`
	Rules         string    `json:"rules"`
	PayValue      int64     `json:"payValue" binding:"required"`      // 支付金额（分）
	ActualValue   int64     `json:"actualValue" binding:"required"`   // 实际价值（分）
	Stock         int       `json:"stock" binding:"required"`         // 秒杀库存量
	BeginTime     time.Time `json:"beginTime" binding:"required"`     // 秒杀开始时间
	EndTime       time.Time `json:"endTime" binding:"required"`       // 秒杀结束时间
	AllowedLevels string    `json:"allowedLevels"`                    // 允许的用户等级
	MinLevel      int       `json:"minLevel"`                         // 最低等级
}

// VoucherSubscribeDto 单个订阅请求
type VoucherSubscribeDto struct {
	VoucherID int64 `json:"voucherId" binding:"required"`
}

// VoucherSubscribeBatchDto 批量订阅请求
type VoucherSubscribeBatchDto struct {
	VoucherIDList []int64 `json:"voucherIdList" binding:"required"`
}

// DelayVoucherReminderDto 设置延迟提醒请求
// 用于秒杀开始前提醒用户
type DelayVoucherReminderDto struct {
	VoucherID    int64 `json:"voucherId" binding:"required"`
	DelaySeconds int   `json:"delaySeconds" binding:"required"` // 延时时长（秒）
}
