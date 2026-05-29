// ═════════════════════════════════════════════════════════════════════
// Kafka 消息生产者 — 基于 segmentio/kafka-go
// 使用 Hash 分区器保证相同 key 的消息进入同一分区（利于顺序消费）
// BatchTimeout=10ms 平衡延迟和吞吐量
// 秒杀链路中使用异步消息削峰，将订单创建从同步链路剥离
// ═════════════════════════════════════════════════════════════════════
package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	kafka "github.com/segmentio/kafka-go"
)

// Producer Kafka 消息生产者
// 封装了消息序列化和发送，支持任意消息类型
type Producer struct {
	writer *kafka.Writer
	log    zerolog.Logger
}

// NewProducer 创建 Kafka 生产者
// 使用 Hash 分区器确保同 key 消息进入同一分区
// Async=false 确保发送失败时调用方可感知并做降级处理
func NewProducer(brokers []string, log zerolog.Logger) *Producer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.Hash{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}
	return &Producer{writer: w, log: log}
}

// Send 发送消息到指定 Topic，自动 JSON 序列化
// key 用于分区路由和消息去重
func (p *Producer) Send(ctx context.Context, topic string, key string, msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	err = p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
	})
	if err != nil {
		p.log.Warn().Err(err).Str("topic", topic).Str("key", key).Msg("kafka send failed")
		return err
	}
	p.log.Debug().Str("topic", topic).Str("key", key).Msg("kafka message sent")
	return nil
}

// SendSeckillMessage 发送秒杀订单消息（topic: seckill_voucher_topic）
func (p *Producer) SendSeckillMessage(ctx context.Context, msg SeckillMessage) error {
	key := fmt.Sprintf("%d_%d", msg.VoucherID, msg.UserID)
	return p.Send(ctx, "seckill_voucher_topic", key, msg)
}

// SendCacheInvalidation 发送缓存失效消息
func (p *Producer) SendCacheInvalidation(ctx context.Context, msg CacheInvalidationMessage) error {
	key := fmt.Sprintf("cache_%d", msg.VoucherID)
	return p.Send(ctx, "seckill_voucher_cache_invalidation", key, msg)
}

// Close 关闭生产者，释放连接
func (p *Producer) Close() error {
	return p.writer.Close()
}
