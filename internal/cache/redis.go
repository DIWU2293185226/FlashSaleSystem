// ═════════════════════════════════════════════════════════════════════
// Redis 缓存封装 — 基于 go-redis 的基础操作层
// 提供常用的 Redis 数据类型操作方法（String/Set/Hash/ZSet）
// 额外封装 Lua 脚本相关操作（Eval/EvalSha/ScriptLoad），供 LuaManager 使用
// ═════════════════════════════════════════════════════════════════════
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/javaup/flashsale-system/internal/config"
	"github.com/redis/go-redis/v9"
)

// RedisCache 封装 go-redis 客户端，提供业务常用的 Redis 操作
// 直接暴露 Client 字段，特殊操作可通过 Client 直接调用
type RedisCache struct {
	Client *redis.Client
}

// InitRedis 根据配置创建 Redis 连接，超时 5 秒
// 连接建立后执行 PING 验证连通性
func InitRedis(cfg *config.RedisConfig) (*RedisCache, error) {
	if cfg == nil {
		return nil, fmt.Errorf("redis config is nil")
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.Database,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: 10,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisCache{Client: client}, nil
}

// MustInitRedis 初始化 Redis，失败时 panic（启动阶段快速失败）
func MustInitRedis(cfg *config.RedisConfig) *RedisCache {
	rc, err := InitRedis(cfg)
	if err != nil {
		panic(fmt.Sprintf("redis init failed: %v", err))
	}
	return rc
}

// Get 获取字符串值
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// Set 设置键值对，支持过期时间
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

// Del 删除一个或多个键
func (r *RedisCache) Del(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.Client.Exists(ctx, key).Result()
	return n > 0, err
}

// Expire 设置键的过期时间
func (r *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.Client.Expire(ctx, key, expiration).Err()
}

// SAdd 向集合添加成员
func (r *RedisCache) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return r.Client.SAdd(ctx, key, members...).Result()
}

// SIsMember 检查成员是否在集合中
func (r *RedisCache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.Client.SIsMember(ctx, key, member).Result()
}

// SRem 从集合移除成员
func (r *RedisCache) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return r.Client.SRem(ctx, key, members...).Result()
}

// HSet 设置哈希字段
func (r *RedisCache) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.Client.HSet(ctx, key, values...).Err()
}

// HGet 获取哈希字段
func (r *RedisCache) HGet(ctx context.Context, key, field string) (string, error) {
	return r.Client.HGet(ctx, key, field).Result()
}

// ZAdd 向有序集合添加成员
func (r *RedisCache) ZAdd(ctx context.Context, key string, member string, score float64) error {
	return r.Client.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Err()
}

// ZRangeByScore 按分数范围查询有序集合成员
func (r *RedisCache) ZRangeByScore(ctx context.Context, key string, min, max string, offset, count int64) ([]string, error) {
	return r.Client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: offset,
		Count:  count,
	}).Result()
}

// IncrBy 将键的值增加指定量
func (r *RedisCache) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.Client.IncrBy(ctx, key, value).Result()
}

// Eval 执行 Lua 脚本
func (r *RedisCache) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return r.Client.Eval(ctx, script, keys, args...).Result()
}

// EvalSha 通过 SHA 哈希执行已缓存的 Lua 脚本
func (r *RedisCache) EvalSha(ctx context.Context, sha string, keys []string, args ...interface{}) (interface{}, error) {
	return r.Client.EvalSha(ctx, sha, keys, args...).Result()
}

// ScriptLoad 将 Lua 脚本加载到 Redis 并返回 SHA 哈希
func (r *RedisCache) ScriptLoad(ctx context.Context, script string) (string, error) {
	return r.Client.ScriptLoad(ctx, script).Result()
}

// Close 关闭 Redis 连接
func (r *RedisCache) Close() error {
	return r.Client.Close()
}
