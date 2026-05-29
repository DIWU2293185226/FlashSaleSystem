// ═════════════════════════════════════════════════════════════════════
// 订阅/候补状态枚举
// 用于用户对已售罄秒杀券的候补订阅流程
// ═════════════════════════════════════════════════════════════════════
package enums

// SubscribeStatus 订阅状态
type SubscribeStatus int

const (
	SubscribeStatusUnsubscribed SubscribeStatus = 0 // 未订阅 — 未登记候补或已取消
	SubscribeStatusSubscribed   SubscribeStatus = 1 // 已订阅 — 正在候补队列中等待
	SubscribeStatusSuccess      SubscribeStatus = 2 // 已成功 — 候补成功并已创建订单
)

func (s SubscribeStatus) Code() int {
	return int(s)
}

func (s SubscribeStatus) String() string {
	switch s {
	case SubscribeStatusUnsubscribed:
		return "未订阅"
	case SubscribeStatusSubscribed:
		return "已订阅"
	case SubscribeStatusSuccess:
		return "已成功"
	default:
		return "未知"
	}
}
