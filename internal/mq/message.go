// ═════════════════════════════════════════════════════════════════════
// Kafka 消息类型定义
// 所有通过 MQ 传递的业务消息集中定义在此
// 使用 JSON 序列化，字段名与 Java 版保持兼容
// ═════════════════════════════════════════════════════════════════════
package mq

// SeckillMessage 秒杀成功后的订单创建消息
// Lua 脚本原子执行成功后发送，消费者异步创建订单
// TraceID 用于全链路追踪和回滚操作
type SeckillMessage struct {
	UserID    int64 `json:"userId"`
	VoucherID int64 `json:"voucherId"`
	OrderID   int64 `json:"orderId"`  // Snowflake 提前生成
	TraceID   int64 `json:"traceId"`  // 用于回滚时追踪轨迹
}

// CacheInvalidationMessage 缓存失效通知
// 秒杀券信息变更时发送，各消费者据此清除本地缓存
type CacheInvalidationMessage struct {
	VoucherID int64 `json:"voucherId"`
}

// DelayReminderMessage 延迟提醒消息
// 秒杀开始前通过延迟队列向用户发送提醒
type DelayReminderMessage struct {
	VoucherID    int64 `json:"voucherId"`
	UserID       int64 `json:"userId"`
	DelaySeconds int   `json:"delaySeconds"` // 延时时长（秒）
}

// VoucherCancelMessage 秒杀回滚消息
// 订单创建失败时的补偿操作，恢复 Redis 库存和用户记录
type VoucherCancelMessage struct {
	OrderID   int64 `json:"orderId"`
	VoucherID int64 `json:"voucherId"`
	UserID    int64 `json:"userId"`
}
