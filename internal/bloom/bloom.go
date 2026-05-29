// ═════════════════════════════════════════════════════════════════════
// 布隆过滤器 — 基于 Redis 位数组
// 用于快速判断 voucher 是否存在，避免每次秒杀请求都查询 DB
// FNV64a 双哈希：h1 + i*h2，减少哈希冲突
// 注意：布隆过滤器有误判率（可配置），但不会漏判
// 在秒杀链路中作为前置过滤，误判只会让不存在的 key 多走一步 Lua
// ═════════════════════════════════════════════════════════════════════
package bloom

import (
	"context"
	"encoding/binary"
	"hash"
	"hash/fnv"
	"math"

	"github.com/javaup/flashsale-system/internal/cache"
)

// Filter Redis 位数组实现的布隆过滤器
// 使用 Redis SETBIT/GETBIT 操作，key 统一前缀 "bloom:{name}"
// bits 和 numHash 根据预估插入量和期望误判率自动计算
type Filter struct {
	name     string           // Redis key 标识（区分不同业务）
	redis    *cache.RedisCache
	bits     uint             // 位数组大小
	numHash  uint             // 哈希函数个数
	hashFn   hash.Hash64      // FNV64a 哈希函数
}

// NewFilter 创建布隆过滤器
// expectedInsertions: 预估插入元素数量
// falseProbability: 期望误判率（越小则位数组越大，通常 0.01 即可）
func NewFilter(name string, expectedInsertions uint, falseProbability float64, redis *cache.RedisCache) *Filter {
	bits, numHash := optimalParams(expectedInsertions, falseProbability)
	return &Filter{
		name:    name,
		redis:   redis,
		bits:    bits,
		numHash: numHash,
		hashFn:  fnv.New64a(),
	}
}

// Add 将元素加入布隆过滤器
func (f *Filter) Add(ctx context.Context, item []byte) error {
	key := "bloom:" + f.name
	for i := uint(0); i < f.numHash; i++ {
		offset := f.hash(item, i) % uint64(f.bits)
		if err := f.redis.Client.SetBit(ctx, key, int64(offset), 1).Err(); err != nil {
			return err
		}
	}
	return nil
}

// AddString 添加字符串元素
func (f *Filter) AddString(ctx context.Context, item string) error {
	return f.Add(ctx, []byte(item))
}

// AddInt64 添加 int64 类型元素（小端编码）
func (f *Filter) AddInt64(ctx context.Context, item int64) error {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(item))
	return f.Add(ctx, b)
}

// Exists 判断元素是否可能存在
// 返回 true: 可能存在（有误判可能）
// 返回 false: 一定不存在
func (f *Filter) Exists(ctx context.Context, item []byte) (bool, error) {
	key := "bloom:" + f.name
	for i := uint(0); i < f.numHash; i++ {
		offset := f.hash(item, i) % uint64(f.bits)
		bit, err := f.redis.Client.GetBit(ctx, key, int64(offset)).Result()
		if err != nil {
			return false, err
		}
		if bit == 0 {
			return false, nil
		}
	}
	return true, nil
}

// ExistsString 判断字符串元素
func (f *Filter) ExistsString(ctx context.Context, item string) (bool, error) {
	return f.Exists(ctx, []byte(item))
}

// ExistsInt64 判断 int64 元素
func (f *Filter) ExistsInt64(ctx context.Context, item int64) (bool, error) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(item))
	return f.Exists(ctx, b)
}

// hash 使用 FNV64a 双哈希算法
// 将 64 位哈希拆分为高 32 位和低 32 位
// 通过 seed 参数生成不同的哈希位置：h1 + seed * h2
func (f *Filter) hash(item []byte, seed uint) uint64 {
	f.hashFn.Reset()
	f.hashFn.Write(item)
	h := f.hashFn.Sum64()
	h1 := h >> 32
	h2 := h & 0xffffffff
	return h1 + uint64(seed)*h2
}

// optimalParams 根据预估插入量和期望误判率计算最优位数组大小和哈希次数
// 公式：
//   bits = -n*ln(p) / (ln2)^2
//   numHash = bits/n * ln2
// 其中 n=插入量, p=误判率
func optimalParams(expectedInsertions uint, falseProbability float64) (bits, numHash uint) {
	if expectedInsertions == 0 {
		expectedInsertions = 1
	}
	bits = uint(math.Ceil(float64(expectedInsertions) * math.Log(falseProbability) / math.Log(1.0/math.Pow(2.0, math.Ln2))))
	numHash = uint(math.Ceil(float64(bits) / float64(expectedInsertions) * math.Ln2))
	if bits < 64 {
		bits = 64
	}
	if numHash < 1 {
		numHash = 1
	}
	return
}
