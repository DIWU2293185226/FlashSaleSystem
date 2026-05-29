// ═════════════════════════════════════════════════════════════════════
// 应用层分片路由 — 替代 ShardingSphere
// 原 Java 项目依赖 ShardingSphere 做分库分表，Go 端无法直接复用
// 因此在应用层实现分片路由逻辑：根据分片键动态计算 DB 和表名
//
// 分片策略：
//   - User/Voucher 系列：int64 分片键通过 intdiv（除后再模）均匀分布
//   - UserPhone：手机号字符串通过 FNV32a 哈希后取模
//   - VoucherOrder：user_id 分库，voucher_id 分表（跨库关联查询需要聚合）
// ═════════════════════════════════════════════════════════════════════
package sharding

import (
	"fmt"
	"hash/fnv"
	"math"

	"github.com/javaup/flashsale-system/internal/config"
)

// Router 分片路由器，根据 key 计算目标数据库和表名
// 所有方法返回字符串，由 Repository 层拼接 SQL 时使用
type Router struct {
	dbCount    int // 数据库总数
	tableCount int // 每库分表数
}

// NewRouter 根据配置创建分片路由器
func NewRouter(cfg *config.ShardConfig) *Router {
	return &Router{
		dbCount:    cfg.DbCount,
		tableCount: cfg.TableCount,
	}
}

// DBForMod 按分片键取模选择数据库
// 例：key=5, dbCount=2 → hmdp_1
func (r *Router) DBForMod(shardKey int64) string {
	dbIdx := shardKey % int64(r.dbCount)
	return fmt.Sprintf("hmdp_%d", dbIdx)
}

// TableForDivMod (key/n)%n 分片策略
// 匹配 Java 的 intdiv 模式，比直接取模分布更均匀
// 适用于自增主键或不均匀的 key 分布
func (r *Router) TableForDivMod(prefix string, shardKey int64) string {
	tableIdx := (shardKey / int64(r.tableCount)) % int64(r.tableCount)
	return fmt.Sprintf("%s_%d", prefix, tableIdx)
}

// TableForMod 简单取模分片
// 适用于均匀分布的 key（如哈希后的值）
func (r *Router) TableForMod(prefix string, shardKey int64) string {
	tableIdx := shardKey % int64(r.tableCount)
	return fmt.Sprintf("%s_%d", prefix, tableIdx)
}

// HashMod FNV32a 哈希后取绝对值，用于字符串类型的分片键（如手机号）
func (r *Router) HashMod(key string) int64 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int64(math.Abs(float64(int32(h.Sum32()))))
}

// TableForHashMod 哈希取模分片
func (r *Router) TableForHashMod(prefix string, key string) string {
	tableIdx := r.HashMod(key) % int64(r.tableCount)
	return fmt.Sprintf("%s_%d", prefix, tableIdx)
}

// DBAndTableForMod 同时计算库名和表名
// 一次调用返回两个值，减少重复计算
func (r *Router) DBAndTableForMod(dbPrefix string, tablePrefix string, shardKey int64) (string, string) {
	db := r.DBForMod(shardKey)
	table := r.TableForMod(tablePrefix, shardKey)
	return db, table
}

// ──────────── 各实体的具体路由方法 ────────────
// 每种实体有其对应的分片策略，封装为独立方法
// Repository 层直接调用即可得到目标库表名

func (r *Router) UserTable(userID int64) string     { return r.TableForDivMod("tb_user", userID) }
func (r *Router) UserDB(userID int64) string         { return r.DBForMod(userID) }

func (r *Router) VoucherTable(voucherID int64) string     { return r.TableForDivMod("tb_voucher", voucherID) }
func (r *Router) VoucherDB(voucherID int64) string         { return r.DBForMod(voucherID) }

func (r *Router) SeckillVoucherTable(voucherID int64) string {
	return r.TableForDivMod("tb_seckill_voucher", voucherID)
}
func (r *Router) SeckillVoucherDB(voucherID int64) string { return r.DBForMod(voucherID) }

// VoucherOrder 特殊：按 user_id 分库、voucher_id 分表
// 这样同一用户的订单集中在一个库，便于分页查询
func (r *Router) VoucherOrderTable(voucherID int64) string { return r.TableForMod("tb_voucher_order", voucherID) }
func (r *Router) VoucherOrderDB(userID int64) string        { return r.DBForMod(userID) }

func (r *Router) UserInfoTable(userID int64) string  { return r.TableForDivMod("tb_user_info", userID) }
func (r *Router) UserInfoDB(userID int64) string      { return r.DBForMod(userID) }

// UserPhone 分片键是字符串手机号，需要使用哈希取模
func (r *Router) UserPhoneTable(phone string) string {
	return r.TableForHashMod("tb_user_phone", phone)
}
func (r *Router) UserPhoneDB(phone string) string {
	dbIdx := r.HashMod(phone) % int64(r.dbCount)
	return fmt.Sprintf("hmdp_%d", dbIdx)
}
