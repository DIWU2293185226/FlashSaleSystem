// ═════════════════════════════════════════════════════════════════════
// 业务错误码体系
// 统一管理所有业务异常，避免 handler/service 层直接写中文字符串
// 每个错误码由数字编号 + 中文描述组成，方便前端根据 code 做多语言
// ═════════════════════════════════════════════════════════════════════
package pkg

import "fmt"

// BaseCodeError 带业务码的错误类型，实现 error 接口
// 相比直接返回字符串，携带 code 方便调用方区分错误类别
type BaseCodeError struct {
	Code int    // 业务错误码
	Msg  string // 中文错误描述
}

func (e *BaseCodeError) Error() string {
	return e.Msg
}

func NewBaseCodeError(code int, msg string) *BaseCodeError {
	return &BaseCodeError{Code: code, Msg: msg}
}

// 预定义业务错误码
// 按模块划分区间：秒杀 1xxxx, 用户 2xxxx
var (
	ErrSuccess                          = NewBaseCodeError(0, "OK")
	ErrSeckillVoucherNotExist           = NewBaseCodeError(10001, "秒杀优惠券不存在")
	ErrSeckillVoucherNotBegin           = NewBaseCodeError(10002, "秒杀优惠券未开始")
	ErrSeckillVoucherIsOver             = NewBaseCodeError(10003, "秒杀优惠券已结束")
	ErrSeckillVoucherStockNotExist      = NewBaseCodeError(10004, "秒杀优惠券库存不存在")
	ErrSeckillVoucherStockInsufficient  = NewBaseCodeError(10005, "秒杀优惠券库存不足")
	ErrSeckillVoucherClaimed            = NewBaseCodeError(10006, "秒杀优惠券已领取")
	ErrSeckillRateLimitIPExceeded       = NewBaseCodeError(10007, "请求过于频繁，请稍后再试")
	ErrSeckillRateLimitUserExceeded     = NewBaseCodeError(10008, "操作过于频繁，请稍后再试")
	ErrSeckillVoucherOrderNotExist      = NewBaseCodeError(10009, "优惠券订单不存在")
	ErrSeckillStockNotNegative          = NewBaseCodeError(10010, "修改后的剩余库存不能为负数")
	ErrVoucherUnavailable               = NewBaseCodeError(10011, "优惠券已下架")
	ErrVoucherExpired                   = NewBaseCodeError(10012, "优惠券已过期")
	ErrVoucherOrderExist                = NewBaseCodeError(10013, "优惠券订单已存在")
	ErrVoucherOrderCancel               = NewBaseCodeError(10014, "优惠券订单已取消")
	ErrUserNotExist                     = NewBaseCodeError(20000, "用户不存在")
	ErrUserAlreadyPurchase              = NewBaseCodeError(20001, "用户已经购买")
)

// GetBaseCode 根据 code 反查预定义错误（兜底返回"未知错误码"）
func GetBaseCode(code int) *BaseCodeError {
	switch code {
	case 0:
		return ErrSuccess
	case 10001:
		return ErrSeckillVoucherNotExist
	case 10002:
		return ErrSeckillVoucherNotBegin
	case 10003:
		return ErrSeckillVoucherIsOver
	case 10004:
		return ErrSeckillVoucherStockNotExist
	case 10005:
		return ErrSeckillVoucherStockInsufficient
	case 10006:
		return ErrSeckillVoucherClaimed
	case 10007:
		return ErrSeckillRateLimitIPExceeded
	case 10008:
		return ErrSeckillRateLimitUserExceeded
	case 10009:
		return ErrSeckillVoucherOrderNotExist
	case 10010:
		return ErrSeckillStockNotNegative
	case 10011:
		return ErrVoucherUnavailable
	case 10012:
		return ErrVoucherExpired
	case 10013:
		return ErrVoucherOrderExist
	case 10014:
		return ErrVoucherOrderCancel
	case 20000:
		return ErrUserNotExist
	case 20001:
		return ErrUserAlreadyPurchase
	default:
		return NewBaseCodeError(code, fmt.Sprintf("未知错误码:%d", code))
	}
}
