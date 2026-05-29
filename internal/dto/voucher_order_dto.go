// ═════════════════════════════════════════════════════════════════════
// 订单查询/操作 DTO
// ═════════════════════════════════════════════════════════════════════
package dto

// GetVoucherOrderDto 按订单 ID 查询
type GetVoucherOrderDto struct {
	OrderID int64 `json:"orderId" binding:"required"`
}

// GetVoucherOrderByVoucherIdDto 按优惠券 ID 查询
type GetVoucherOrderByVoucherIdDto struct {
	VoucherID int64 `json:"voucherId" binding:"required"`
}

// CancelVoucherOrderDto 取消订单请求
type CancelVoucherOrderDto struct {
	VoucherID int64 `json:"voucherId" binding:"required"`
}
