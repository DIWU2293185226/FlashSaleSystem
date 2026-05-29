// ═════════════════════════════════════════════════════════════════════
// 分布式锁 — 基于 Redis SETNX
// 在秒杀链路中用于订单创建的互斥控制
// 提供两种释放方式：简单 DEL 和 Lua 值校验（防止误删其他线程的锁）
// TryLock 支持带超时的轮询获取
// ═════════════════════════════════════════════════════════════════════
package locker

import (
	"context"
	"fmt"
	"time"

	"github.com/javaup/flashsale-system/internal/cache"
	"github.com/redis/go-redis/v9"
)

const (
	// DefaultTTL 默认锁的 TTL = 30 秒
	DefaultTTL = 30 * time.Second
	// retryInterval 锁重试间隔 = 100ms
	retryInterval = 100 * time.Millisecond
)

// Locker Redis 分布式锁
// 注意：这不是 Redlock，适用于单 Redis 节点场景
// 锁的正确性依赖于 Redis 主节点可用性
type Locker struct {
	redis *cache.RedisCache
}

// New 创建分布式锁
func New(redis *cache.RedisCache) *Locker {
	return &Locker{redis: redis}
}

// Lock 尝试获取锁（SETNX），单次尝试不重试
// ttl 为锁自动释放时间，防止死锁
func (l *Locker) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	ok, err := l.redis.Client.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock %s: %w", key, err)
	}
	return ok, nil
}

// TryLock 在超时时间内轮询获取锁
// timeout: 最长等待时间，到期未获取返回 false
// 轮询间隔 100ms，适合秒杀订单创建的短时等待场景
func (l *Locker) TryLock(ctx context.Context, key string, ttl time.Duration, timeout time.Duration) (bool, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ok, err := l.Lock(ctx, key, ttl)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(retryInterval):
		}
	}
	return false, nil
}

// Unlock 直接删除 key 释放锁（简单模式）
// 适用于锁的持有者明确唯一的场景
func (l *Locker) Unlock(ctx context.Context, key string) error {
	return l.redis.Client.Del(ctx, key).Err()
}

// UnlockWithScript 通过 Lua 脚本校验锁持有者后释放
// 防止误释放其他线程持有的锁
// Lua 逻辑：GET key == value ? DEL : return 0
func (l *Locker) UnlockWithScript(ctx context.Context, key, value string) (bool, error) {
	script := `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
	`
	result, err := l.redis.Client.Eval(ctx, script, []string{key}, value).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return result.(int64) == 1, nil
}
