// ═════════════════════════════════════════════════════════════════════
// 滚动分页结果 — 用于 Feed 流等需要游标分页的场景
// 相比传统 offset/limit，滚动分页在数据频繁插入时不会导致重复或跳变
// ═════════════════════════════════════════════════════════════════════
package dto

// ScrollResult 滚动分页响应
// MinTime 是当前页的最小时间戳，Offset 是相同时间戳的偏移量
// 前端下次请求时传入这两个值实现无缝翻页
type ScrollResult struct {
	List    []interface{} `json:"list"`    // 数据列表
	MinTime int64         `json:"minTime"` // 当前页最小时间戳（下次请求传入）
	Offset  int           `json:"offset"`  // 当前页最后一条的偏移量
}
