// ═════════════════════════════════════════════════════════════════════
// 本地缓存封装 — 基于 FreeCache 的进程内 L1 缓存
// 相比 Redis（L2），本地缓存的读取延迟从毫秒级降至纳秒级
// 适用于热数据缓存，如商铺信息、用户信息等频繁读取但不常变化的数据
// 默认容量 100MB，支持泛型 Get/Set、原始字节存取、空值标记防穿透
// ═════════════════════════════════════════════════════════════════════
package cache

import (
	"encoding/json"

	"github.com/coocood/freecache"
)

// LocalCache 本地进程内缓存，封装 FreeCache
// 作为多级缓存（L1，L2=Redis）的最前端，只存最热的数据
type LocalCache struct {
	cache *freecache.Cache
}

// NewLocalCache 创建本地缓存，size 为字节数，默认 100MB
// FreeCache 在初始化时一次性分配内存，减少 GC 压力
func NewLocalCache(size int) *LocalCache {
	if size <= 0 {
		size = 100 * 1024 * 1024 // 100 MB
	}
	return &LocalCache{
		cache: freecache.NewCache(size),
	}
}

// Get 获取并反序列化值为指定类型
func (c *LocalCache) Get(key string, dest interface{}) error {
	data, err := c.cache.Get([]byte(key))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Set 序列化并存储值，ttlSeconds 为过期秒数
func (c *LocalCache) Set(key string, value interface{}, ttlSeconds int) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.cache.Set([]byte(key), data, ttlSeconds)
}

// Del 删除键
func (c *LocalCache) Del(key string) bool {
	return c.cache.Del([]byte(key))
}

// Clear 清空整个本地缓存
func (c *LocalCache) Clear() {
	c.cache.Clear()
}

// GetRaw 获取原始字节（不反序列化）
func (c *LocalCache) GetRaw(key string) ([]byte, error) {
	return c.cache.Get([]byte(key))
}

// SetRaw 存储原始字节
func (c *LocalCache) SetRaw(key string, data []byte, ttlSeconds int) error {
	return c.cache.Set([]byte(key), data, ttlSeconds)
}

// EntryCount 返回缓存条目数
func (c *LocalCache) EntryCount() int {
	return int(c.cache.EntryCount())
}

// HitRate 返回缓存命中率
func (c *LocalCache) HitRate() float64 {
	return c.cache.HitRate()
}

// Expire 重置键的过期时间
func (c *LocalCache) Expire(key string, ttlSeconds int) bool {
	data, err := c.cache.Get([]byte(key))
	if err != nil {
		return false
	}
	return c.cache.Set([]byte(key), data, ttlSeconds) == nil
}

// defaultLocalTTL 默认本地缓存 TTL = 10 分钟
const defaultLocalTTL = 600

// defaultNullTTL 空值标记 TTL = 5 分钟（与 pkg.CacheNullTTL 保持一致）
const defaultNullTTL = 300

// SetWithDefaults 使用默认 TTL 存储值
func (c *LocalCache) SetWithDefaults(key string, value interface{}) error {
	return c.Set(key, value, defaultLocalTTL)
}

// SetNull 存储空值标记，用于缓存穿透保护
// 当查询 DB 发现数据不存在时，写入此标记
// 后续相同查询直接命中空标记，避免反复穿透到 DB
func (c *LocalCache) SetNull(key string) error {
	return c.SetRaw(key, []byte{}, defaultNullTTL)
}

// IsNull 检查键是否为空值标记
func (c *LocalCache) IsNull(key string) bool {
	data, err := c.cache.Get([]byte(key))
	if err != nil {
		return false
	}
	return len(data) == 0
}

// TTL 返回键的剩余时间（FreeCache 不直接支持 TTL 查询，近似实现）
func (c *LocalCache) TTL(key string) int {
	_, err := c.cache.Get([]byte(key))
	if err != nil {
		return -1
	}
	return 0
}
