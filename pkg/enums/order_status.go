// ═════════════════════════════════════════════════════════════════════
// 订单状态枚举
// ═════════════════════════════════════════════════════════════════════
package enums

// OrderStatus 订单状态
type OrderStatus int

const (
	OrderStatusNormal OrderStatus = 1 // 正常 — 订单创建成功，有效
	OrderStatusCancel OrderStatus = 2 // 取消 — 用户主动取消或超时取消
)

func (s OrderStatus) Code() int {
	return int(s)
}

func (s OrderStatus) String() string {
	switch s {
	case OrderStatusNormal:
		return "正常"
	case OrderStatusCancel:
		return "取消"
	default:
		return "未知"
	}
}
