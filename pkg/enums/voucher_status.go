// ═════════════════════════════════════════════════════════════════════
// 优惠券状态枚举
// ═════════════════════════════════════════════════════════════════════
package enums

// VoucherStatus 优惠券状态
type VoucherStatus int

const (
	VoucherStatusAvailable   VoucherStatus = 1 // 上架 — 可正常使用/秒杀
	VoucherStatusUnavailable VoucherStatus = 2 // 下架 — 商家主动下架
	VoucherStatusExpired     VoucherStatus = 3 // 过期 — 超过有效时间自动过期
)

func (s VoucherStatus) Code() int {
	return int(s)
}

func (s VoucherStatus) String() string {
	switch s {
	case VoucherStatusAvailable:
		return "上架"
	case VoucherStatusUnavailable:
		return "下架"
	case VoucherStatusExpired:
		return "过期"
	default:
		return "未知"
	}
}
