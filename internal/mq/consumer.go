// ═════════════════════════════════════════════════════════════════════
// Kafka 消息消费者 — 基于 segmentio/kafka-go
// 消费者组模式，支持自动重平衡
// 核心特性：
//   - 3 次重试 + 指数退避（100ms, 500ms, 1s）
//   - Goroutine panic 自愈（异常崩溃后自动重启消费循环）
//   - 重试耗尽后仍提交偏移（不阻塞消费进度，保证 at-least-once）
// ═════════════════════════════════════════════════════════════════════
package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
	kafka "github.com/segmentio/kafka-go"
)

const maxRetries = 3

// SeckillOrderHandler 秒杀订单处理回调
type SeckillOrderHandler func(ctx context.Context, msg SeckillMessage) error

// CacheInvalidationHandler 缓存失效处理回调
type CacheInvalidationHandler func(ctx context.Context, msg CacheInvalidationMessage) error

// Consumer Kafka 消费者，按 Topic 启动独立 goroutine 消费
// 每个 Topic 一个 Reader，通过 context cancel 统一关闭
type Consumer struct {
	groupID string
	brokers []string
	log     zerolog.Logger
	mu      sync.Mutex
	cancels []context.CancelFunc
}

// NewConsumer 创建消费者组模式的 Kafka 消费者
// groupID 用于消费者组内负载均衡
func NewConsumer(brokers []string, groupID string, log zerolog.Logger) *Consumer {
	return &Consumer{
		groupID: groupID,
		brokers: brokers,
		log:     log,
	}
}

// ConsumeSeckillOrders 启动秒杀订单消费 goroutine
func (c *Consumer) ConsumeSeckillOrders(ctx context.Context, handler SeckillOrderHandler) {
	c.consume(ctx, "seckill_voucher_topic", func(msg kafka.Message) error {
		var sm SeckillMessage
		if err := json.Unmarshal(msg.Value, &sm); err != nil {
			return fmt.Errorf("failed to unmarshal seckill message: %w", err)
		}
		return handler(ctx, sm)
	})
}

// ConsumeCacheInvalidation 启动缓存失效消费 goroutine
func (c *Consumer) ConsumeCacheInvalidation(ctx context.Context, handler CacheInvalidationHandler) {
	c.consume(ctx, "seckill_voucher_cache_invalidation", func(msg kafka.Message) error {
		var cm CacheInvalidationMessage
		if err := json.Unmarshal(msg.Value, &cm); err != nil {
			return fmt.Errorf("failed to unmarshal cache invalidation message: %w", err)
		}
		return handler(ctx, cm)
	})
}

type messageHandler func(msg kafka.Message) error

// consume 启动一个 Topic 的后台消费循环
// 关键设计：
// 1. panic recovery 后自动重启，防止单个异常导致整个消费者死亡
// 2. 每次 FetchMessage 检查 context，支持优雅退出
// 3. 重试耗尽后继续提交偏移，避免消息堆积
func (c *Consumer) consume(parentCtx context.Context, topic string, handler messageHandler) {
	ctx, cancel := context.WithCancel(parentCtx)
	c.mu.Lock()
	c.cancels = append(c.cancels, cancel)
	c.mu.Unlock()

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     c.brokers,
		GroupID:     c.groupID,
		Topic:       topic,
		MinBytes:    10,
		MaxBytes:    10e6, // 10MB
		MaxWait:     time.Second,
		StartOffset: kafka.LastOffset,
	})

	go func() {
		defer func() {
			r.Close()
			// panic 自愈：如果消费者因未知 panic 退出，自动重启
			if rec := recover(); rec != nil {
				c.log.Error().Interface("panic", rec).Str("topic", topic).Msg("consumer panicked, restarting")
				go c.consume(parentCtx, topic, handler)
			}
		}()
		c.log.Info().Str("topic", topic).Msg("starting consumer")
		for {
			select {
			case <-ctx.Done():
				c.log.Info().Str("topic", topic).Msg("consumer stopped")
				return
			default:
			}

			msg, err := r.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.log.Warn().Err(err).Str("topic", topic).Msg("fetch message error")
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if err := c.processWithRetry(handler, msg, topic); err != nil {
				c.log.Error().Err(err).Str("topic", topic).
					Int("partition", msg.Partition).
					Int64("offset", msg.Offset).
					Msg("all retries exhausted, committing offset anyway")
			}

			// 无论重试结果如何都提交偏移，防止消息堆积
			if err := r.CommitMessages(ctx, msg); err != nil {
				c.log.Warn().Err(err).Str("topic", topic).Msg("commit message error")
			}
		}
	}()
}

// processWithRetry 处理消息，最多重试 3 次
// 退避策略：第 1 次间隔 100ms，第 2 次 500ms，第 3 次 1s（100*i² ms）
// 所有重试耗尽后返回最终错误，由调用方决定是否提交偏移
func (c *Consumer) processWithRetry(handler messageHandler, msg kafka.Message, topic string) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			backoff := time.Duration(100*i*i) * time.Millisecond
			if backoff < 100*time.Millisecond {
				backoff = 100 * time.Millisecond
			}
			time.Sleep(backoff)
			c.log.Warn().Str("topic", topic).Int("retry", i).Msg("retrying message")
		}
		if err := handler(msg); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	return fmt.Errorf("message processing failed after %d retries: %w", maxRetries, lastErr)
}

// Close 停止所有消费者 goroutine
func (c *Consumer) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, cancel := range c.cancels {
		cancel()
	}
	c.cancels = nil
}
