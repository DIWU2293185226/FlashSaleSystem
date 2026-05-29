// ═════════════════════════════════════════════════════════════════════
// 全局常量定义 — Redis Key、Kafka Topic、系统参数
// 所有字符串常量集中管理，避免各模块硬编码
// 命名前缀区分业务域：seckill:/cache:/lock: 等
// ═════════════════════════════════════════════════════════════════════
package pkg

// Redis Key 前缀常量
// 使用冒号分隔的层级结构，便于 Redis CLI 按前缀扫描和分组
const (
	// ──── 秒杀相关 Key ────
	SeckillStockKey            = "seckill:stock:"          // 秒杀库存 Hash
	SeckillUserKey             = "seckill:user:"           // 已秒杀用户 Set（一人一单校验）
	SeckillTraceLogKey         = "seckill:trace:"          // 秒杀轨迹日志
	SeckillVoucherTagKey       = "seckill:cache:voucher:"  // 秒杀券缓存
	SeckillVoucherNullTagKey   = "seckill:cache:null:"     // 空值标记缓存（防穿透）
	SeckillSubscribeUserKey    = "seckill:subscribe:user:" // 订阅用户 Set
	SeckillSubscribeZSetKey    = "seckill:subscribe:zset:" // 订阅时间序 ZSet
	SeckillSubscribeStatusKey  = "seckill:subscribe:status:" // 订阅状态 Key
	SeckillShopTopBuyersDaily  = "shop:top:buyers:daily:"  // 店铺今日热门买家

	// ──── 缓存 Key ────
	CacheShopKey     = "cache:shop:"      // 商铺信息缓存
	CacheShopNullKey = "cache:shop:null:" // 商铺空值标记（防缓存穿透）

	// ──── 分布式锁 Key ────
	LockSeckillVoucherKey      = "lock:seckill:voucher:"      // 秒杀券操作锁
	LockSeckillVoucherStockKey = "lock:seckill:voucher:stock:" // 秒杀库存操作锁
	LockOrderKey               = "lock:order:"                 // 订单操作锁

	// ──── Kafka Topic 主题 ────
	SeckillVoucherTopic             = "seckill_voucher_topic"              // 秒杀订单创建
	SeckillVoucherCacheInvalidation = "seckill_voucher_cache_invalidation" // 缓存失效通知
	DelayVoucherReminder            = "delay_voucher_reminder"            // 延迟提醒

	// 布隆过滤器名称标识
	BloomFilterHandlerShop    = "shop"
	BloomFilterHandlerVoucher = "voucher"

	// 分页查询最大每页条数（前端列表页统一限制）
	MaxPageSize = 5
	// 缓存空值 TTL（分钟），用于缓存穿透保护
	CacheNullTTL = 5

	// 防重复提交标识前缀
	SeckillVoucherOrder = "seckill_voucher_order"
)
