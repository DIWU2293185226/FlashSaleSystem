// ═════════════════════════════════════════════════════════════════════
// 多级缓存抽象 — 统一本地缓存（FreeCache）和 Redis 缓存
// 查询顺序：L1 本地 → L2 Redis，Redis 命中后回填 L1
// 写操作：同时写入两级缓存
// 删操作：两级同时失效（或 L1 单独失效用于接收 Redis 失效通知）
// ═════════════════════════════════════════════════════════════════════
package cache

import (
	"context"
	"encoding/json"
	"time"
)

// CacheLevel 缓存层级选择
type CacheLevel int

const (
	CacheLevelAll   CacheLevel = iota // 同时操作 L1 + L2
	CacheLevelLocal                   // 仅操作 L1（本地缓存）
	CacheLevelRedis                   // 仅操作 L2（Redis）
)

var bg = context.Background()

// MultiLevelCache 多级缓存，提供 Local → Redis 两级缓存操作
// 适用于商铺信息等读多写少的热数据
type MultiLevelCache struct {
	local *LocalCache
	redis *RedisCache
}

// NewMultiLevelCache 创建多级缓存，指定本地缓存容量
func NewMultiLevelCache(localSize int, redis *RedisCache) *MultiLevelCache {
	return &MultiLevelCache{
		local: NewLocalCache(localSize),
		redis: redis,
	}
}

// NewMultiLevelCacheWithLocal 使用已有的 LocalCache 创建多级缓存
func NewMultiLevelCacheWithLocal(local *LocalCache, redis *RedisCache) *MultiLevelCache {
	return &MultiLevelCache{
		local: local,
		redis: redis,
	}
}

// Get 按 L1 → L2 顺序读取缓存
// L2 命中后自动回填 L1，下次访问直接命中 L1
func (mc *MultiLevelCache) Get(key string, dest interface{}) error {
	// L1 查询（纳秒级）
	err := mc.local.Get(key, dest)
	if err == nil {
		return nil
	}

	// L2 查询（毫秒级）
	data, err := mc.redis.Get(bg, key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return err
	}

	// L2 命中 → 回填 L1
	if bytes, err := json.Marshal(dest); err == nil {
		_ = mc.local.SetRaw(key, bytes, defaultLocalTTL)
	}

	return nil
}

// Set 同时写入 L1 和 L2
func (mc *MultiLevelCache) Set(key string, value interface{}, ttlSeconds int) error {
	if err := mc.redis.Set(bg, key, value, time.Duration(ttlSeconds)*time.Second); err != nil {
		return err
	}
	_ = mc.local.Set(key, value, ttlSeconds)
	return nil
}

// Del 同时删除 L1 和 L2
func (mc *MultiLevelCache) Del(key string) error {
	_ = mc.redis.Del(bg, key)
	mc.local.Del(key)
	return nil
}

// DelLocal 仅删除 L1（用于收到 Redis 失效通知时同步清除本地缓存）
func (mc *MultiLevelCache) DelLocal(key string) {
	mc.local.Del(key)
}

// GetLocal 仅读取 L1
func (mc *MultiLevelCache) GetLocal(key string, dest interface{}) error {
	return mc.local.Get(key, dest)
}

// SetLocal 仅写入 L1
func (mc *MultiLevelCache) SetLocal(key string, value interface{}, ttlSeconds int) error {
	return mc.local.Set(key, value, ttlSeconds)
}

// GetRedis 仅读取 L2
func (mc *MultiLevelCache) GetRedis(key string) (string, error) {
	return mc.redis.Get(bg, key)
}

// SetRedis 仅写入 L2
func (mc *MultiLevelCache) SetRedis(key string, value interface{}, ttlSeconds int) error {
	return mc.redis.Set(bg, key, value, time.Duration(ttlSeconds)*time.Second)
}

// Clear 清空 L1
func (mc *MultiLevelCache) Clear() error {
	mc.local.Clear()
	return nil
}

// WithLevel 返回指定层级的操作视图
func (mc *MultiLevelCache) WithLevel(level CacheLevel) *ScopedCache {
	return &ScopedCache{
		multi: mc,
		level: level,
	}
}

// ScopedCache 限定层级的缓存操作视图
// 某些场景只需要操作 L1 或 L2（如缓存预热只写 L2）
type ScopedCache struct {
	multi *MultiLevelCache
	level CacheLevel
}

func (s *ScopedCache) Get(key string, dest interface{}) error {
	switch s.level {
	case CacheLevelLocal:
		return s.multi.GetLocal(key, dest)
	case CacheLevelRedis:
		data, err := s.multi.GetRedis(key)
		if err != nil {
			return err
		}
		return json.Unmarshal([]byte(data), dest)
	default:
		return s.multi.Get(key, dest)
	}
}

func (s *ScopedCache) Set(key string, value interface{}, ttlSeconds int) error {
	switch s.level {
	case CacheLevelLocal:
		return s.multi.local.Set(key, value, ttlSeconds)
	case CacheLevelRedis:
		return s.multi.SetRedis(key, value, ttlSeconds)
	default:
		return s.multi.Set(key, value, ttlSeconds)
	}
}

func (s *ScopedCache) Del(key string) error {
	switch s.level {
	case CacheLevelLocal:
		s.multi.local.Del(key)
		return nil
	case CacheLevelRedis:
		return s.multi.redis.Del(bg, key)
	default:
		return s.multi.Del(key)
	}
}

// SetNX 仅当 key 不存在时设置（Redis SETNX）
func (mc *MultiLevelCache) SetNX(key string, value interface{}, ttlSeconds int) (bool, error) {
	return mc.redis.Client.SetNX(bg, key, value, time.Duration(ttlSeconds)*time.Second).Result()
}

// Exists 检查 Redis 中 key 是否存在
func (mc *MultiLevelCache) Exists(key string) (bool, error) {
	n, err := mc.redis.Client.Exists(bg, key).Result()
	return n > 0, err
}
