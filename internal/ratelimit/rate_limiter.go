// ═════════════════════════════════════════════════════════════════════
// 限流器 — 秒杀系统的第一道防线
// 提供三层限流策略：
// 1. 滑动窗口（IP 粒度）— 基于 ZSet，清理过期记录 → 计数 → 写入
// 2. 令牌桶（用户粒度）— 基于 Hash，按时间差自动补充令牌
// 3. 访问令牌（细粒度控制）— 基于 Hash，预先生成，每次消费一个
//
// 所有限流逻辑使用 Lua 脚本在 Redis 中原子执行
// ═════════════════════════════════════════════════════════════════════
package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/javaup/flashsale-system/internal/cache"
)

// SlidingWindowLimiter 滑动窗口限流器
// 在指定时间窗口内限制最大请求次数
// 窗口过期记录自动清除，内存友好
type SlidingWindowLimiter struct {
	lua *cache.LuaManager
}

// NewSlidingWindowLimiter 创建滑动窗口限流器
func NewSlidingWindowLimiter(lua *cache.LuaManager) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{lua: lua}
}

// Allow 检查请求是否被允许
// windowMs: 时间窗口（毫秒），maxRequests: 窗口内最大请求数
// Lua 实现：ZREMRANGEBYSCORE 清理窗口外记录 → ZCARD 计数 → ZADD 写入
func (s *SlidingWindowLimiter) Allow(key string, windowMs int64, maxRequests int64, nowMs int64) (bool, error) {
	result, err := s.lua.Eval("rateLimitSliding",
		[]string{key},
		windowMs, maxRequests, nowMs,
	)
	if err != nil {
		return false, fmt.Errorf("rate limit eval error: %w", err)
	}
	arr, ok := result.([]interface{})
	if !ok || len(arr) < 1 {
		return false, fmt.Errorf("unexpected rate limit result: %v", result)
	}
	code, _ := arr[0].(int64)
	return code == 1, nil
}

// TokenBucketLimiter 令牌桶限流器
// 按时间差计算应补充的令牌数，支持突发流量
// rate: 令牌产生速率（个/秒），capacity: 桶容量
type TokenBucketLimiter struct {
	lua *cache.LuaManager
}

// NewTokenBucketLimiter 创建令牌桶限流器
func NewTokenBucketLimiter(lua *cache.LuaManager) *TokenBucketLimiter {
	return &TokenBucketLimiter{lua: lua}
}

// Allow 消耗指定数量的令牌
// Lua 实现：读取上次时间和可用令牌 → 按时间差补充 → 判断是否足够 → 扣减
func (t *TokenBucketLimiter) Allow(key string, rate float64, capacity float64, tokensToConsume float64) (bool, error) {
	now := time.Now().UnixMilli()
	result, err := t.lua.Eval("tokenBucket",
		[]string{key},
		rate, capacity, now, tokensToConsume,
	)
	if err != nil {
		return false, fmt.Errorf("token bucket eval error: %w", err)
	}
	arr, ok := result.([]interface{})
	if !ok || len(arr) < 1 {
		return false, fmt.Errorf("unexpected token bucket result: %v", result)
	}
	code, _ := arr[0].(int64)
	return code == 1, nil
}

// AccessTokenManager 访问令牌管理器
// 秒杀前预生成一批令牌，用户拿到令牌后才允许进入秒杀
// 适用于"预约 → 抢购"模式的准入控制
type AccessTokenManager struct {
	lua *cache.LuaManager
}

// NewAccessTokenManager 创建访问令牌管理器
func NewAccessTokenManager(lua *cache.LuaManager) *AccessTokenManager {
	return &AccessTokenManager{lua: lua}
}

// Generate 为用户生成一个访问令牌，存入 Redis Hash
// ttlSeconds: 令牌有效期（秒），到期自动失效
func (a *AccessTokenManager) Generate(key string, voucherID int64, userID int64, ttlSeconds int) (string, error) {
	token := fmt.Sprintf("tk_%d_%d_%d", voucherID, userID, time.Now().UnixNano())
	result, err := a.lua.Eval("seckill_access_token",
		[]string{key},
		"generate", strconv.FormatInt(voucherID, 10), strconv.FormatInt(userID, 10), token, ttlSeconds,
	)
	if err != nil {
		return "", fmt.Errorf("access token generate error: %w", err)
	}
	arr, ok := result.([]interface{})
	if !ok || len(arr) < 2 {
		return "", fmt.Errorf("unexpected generate result: %v", result)
	}
	return fmt.Sprintf("%v", arr[1]), nil
}

// Validate 校验并消耗一个访问令牌（一次性，校验后删除）
func (a *AccessTokenManager) Validate(key string, voucherID int64, userID int64, token string) (bool, error) {
	result, err := a.lua.Eval("seckill_access_token",
		[]string{key},
		"validate", strconv.FormatInt(voucherID, 10), strconv.FormatInt(userID, 10), token,
	)
	if err != nil {
		return false, fmt.Errorf("access token validate error: %w", err)
	}
	arr, ok := result.([]interface{})
	if !ok || len(arr) < 1 {
		return false, fmt.Errorf("unexpected validate result: %v", result)
	}
	code, _ := arr[0].(int64)
	return code == 1, nil
}

// CheckRepeat 防重复提交（SETNX 实现）
// 在秒杀链路中作为第一层幂等控制
// 重要：Redis 故障时允许通过，由 Lua SADD 兜底一人一单
func CheckRepeat(ctx context.Context, redis *cache.RedisCache, prefix, key string) (bool, error) {
	redisKey := "repeat:" + prefix + ":" + key
	ok, err := redis.Client.SetNX(ctx, redisKey, "1", 10*time.Second).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

var bg = context.Background()
