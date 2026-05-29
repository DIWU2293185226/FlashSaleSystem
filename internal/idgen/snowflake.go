// ═════════════════════════════════════════════════════════════════════
// Snowflake ID 生成器 — 分布式唯一 ID
// 组成：41bit 时间戳 + 5bit workerID + 5bit datacenterID + 12bit 序列号
// 自定义 epoch: 2023-11-14（避开 Twitter 原始 epoch，延长使用寿命）
// 每毫秒可生成 4096 个唯一 ID
// 线程安全（sync.Mutex），支持解析 ID 回推时间/机器信息
// ═════════════════════════════════════════════════════════════════════
package idgen

import (
	"sync"
	"time"
)

const (
	epoch           = 1700000000000 // 自定义起始时间戳 (2023-11-14 13:33:20 UTC)
	workerBits      = 5             // Worker ID 位数
	datacenterBits  = 5             // 数据中心 ID 位数
	seqBits         = 12            // 序列号位数

	maxWorker     = -1 ^ (-1 << workerBits)      // 最大 worker ID (31)
	maxDatacenter = -1 ^ (-1 << datacenterBits)  // 最大 datacenter ID (31)
	maxSeq        = -1 ^ (-1 << seqBits)         // 最大序列号 (4095)

	workerShift      = seqBits                                    // 序列号偏移
	datacenterShift  = seqBits + workerBits                       // 数据中心偏移
	timestampShift   = seqBits + workerBits + datacenterBits      // 时间戳偏移
)

// Snowflake 分布式 ID 生成器
// Worker/Datacenter ID 通过 Redis Lua 脚本自动分配（见 workAndDataCenterId.lua）
type Snowflake struct {
	mu            sync.Mutex
	workerID      int64
	datacenterID  int64
	sequence      int64
	lastStamp     int64
}

// NewSnowflake 创建 Snowflake ID 生成器
// workerID/datacenterID 超出范围（0-31）时自动归零
func NewSnowflake(workerID, datacenterID int64) *Snowflake {
	if workerID < 0 || workerID > maxWorker {
		workerID = 0
	}
	if datacenterID < 0 || datacenterID > maxDatacenter {
		datacenterID = 0
	}
	return &Snowflake{
		workerID:     workerID,
		datacenterID: datacenterID,
		sequence:     0,
		lastStamp:    -1,
	}
}

// NextID 生成下一个唯一 ID
// 处理两种竞争场景：
// 1. 同一毫秒内：序列号自增，溢出则等待下一毫秒
// 2. 时钟回拨：自旋等待直到追上之前的时间戳
func (s *Snowflake) NextID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()
	// 时钟回拨保护：等待系统时间追上来
	if now < s.lastStamp {
		for now < s.lastStamp {
			now = time.Now().UnixMilli()
		}
	}

	if now == s.lastStamp {
		s.sequence = (s.sequence + 1) & maxSeq
		// 同一毫秒内序列号耗尽，等待下一毫秒
		if s.sequence == 0 {
			for now <= s.lastStamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		s.sequence = 0
	}

	s.lastStamp = now
	return (now-epoch)<<timestampShift |
		(s.datacenterID << datacenterShift) |
		(s.workerID << workerShift) |
		s.sequence
}

// ParseID 从 ID 中解析出时间戳、workerID、datacenterID 和序列号
// 可用于全链路追踪时反推 ID 生成时间和节点
func ParseID(id int64) (timestamp int64, workerID int64, datacenterID int64, sequence int64) {
	timestamp = (id >> timestampShift) + epoch
	datacenterID = (id >> datacenterShift) & maxDatacenter
	workerID = (id >> workerShift) & maxWorker
	sequence = id & maxSeq
	return
}
