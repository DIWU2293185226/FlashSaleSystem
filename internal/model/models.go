// ═════════════════════════════════════════════════════════════════════
// 辅助业务数据模型
// 包含关注、评论、用户扩展信息、对账日志、订阅、回滚失败记录等
// 这些实体统一放在此文件而非分散多个文件，减少包内文件数量
// ═════════════════════════════════════════════════════════════════════
package model

import "time"

// Follow 关注关系，对应 tb_follow
// UserID 关注 FollowUserID，用于社交 Feed 流推送
type Follow struct {
	ID           int64     `gorm:"column:id;primaryKey" json:"id"`
	UserID       int64     `gorm:"column:user_id" json:"userId"`
	FollowUserID int64     `gorm:"column:follow_user_id" json:"followUserId"`
	CreateTime   time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
}

func (Follow) TableName() string { return "tb_follow" }

// BlogComments 博客评论，对应 tb_blog_comments
// ParentID/AnswerID 支持嵌套回复结构
type BlogComments struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	UserID     int64     `gorm:"column:user_id" json:"userId"`
	BlogID     int64     `gorm:"column:blog_id" json:"blogId"`
	ParentID   int64     `gorm:"column:parent_id" json:"parentId"` // 父评论 ID（用于回复）
	AnswerID   int64     `gorm:"column:answer_id" json:"answerId"` // 被回复评论 ID
	Content    string    `gorm:"column:content;size:255" json:"content"`
	Liked      int       `gorm:"column:liked" json:"liked"`
	Status     int       `gorm:"column:status" json:"status"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (BlogComments) TableName() string { return "tb_blog_comments" }

// UserInfo 用户扩展信息，对应 tb_user_info
// 与 User 表一对一，存储城市、简介、粉丝数、等级等
type UserInfo struct {
	ID         int64      `gorm:"column:id;primaryKey" json:"id"`
	UserID     int64      `gorm:"column:user_id" json:"userId"`
	City       string     `gorm:"column:city;size:64" json:"city"`
	Introduce  string     `gorm:"column:introduce;size:128" json:"introduce"`
	Fans       int        `gorm:"column:fans" json:"fans"`         // 粉丝数
	Followee   int        `gorm:"column:followee" json:"followee"`  // 关注数
	Gender     int        `gorm:"column:gender" json:"gender"`
	Birthday   *time.Time `gorm:"column:birthday" json:"birthday"`
	Credits    int        `gorm:"column:credits" json:"credits"`   // 积分
	Level      int        `gorm:"column:level" json:"level"`       // 用户等级（影响秒杀权限）
	CreateTime time.Time  `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime time.Time  `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (UserInfo) TableName() string { return "tb_user_info" }

// UserPhone 用户手机号表，对应 tb_user_phone
// 手机号通过 FNV32a 哈希分片，支持多手机号绑定
type UserPhone struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	UserID     int64     `gorm:"column:user_id" json:"userId"`
	Phone      string    `gorm:"column:phone;size:512" json:"phone"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (UserPhone) TableName() string { return "tb_user_phone" }

// VoucherOrderRouter 分片路由器辅助表，对应 tb_voucher_order_router
// 记录订单的路由信息，用于跨分片查询时定位
type VoucherOrderRouter struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	OrderID    int64     `gorm:"column:order_id" json:"orderId"`
	UserID     int64     `gorm:"column:user_id" json:"userId"`
	VoucherID  int64     `gorm:"column:voucher_id" json:"voucherId"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (VoucherOrderRouter) TableName() string { return "tb_voucher_order_router" }

// VoucherReconcileLog 库存对账日志，对应 tb_voucher_reconcile_log
// 记录 Redis 和 DB 库存差异及修复动作，用于监控和排查
type VoucherReconcileLog struct {
	ID                   int64     `gorm:"column:id;primaryKey" json:"id"`
	OrderID              int64     `gorm:"column:order_id" json:"orderId"`
	UserID               int64     `gorm:"column:user_id" json:"userId"`
	VoucherID            int64     `gorm:"column:voucher_id" json:"voucherId"`
	MessageID            string    `gorm:"column:message_id;size:64" json:"messageId"`
	Detail               string    `gorm:"column:detail;size:1024" json:"detail"`
	BeforeQty            int       `gorm:"column:before_qty" json:"beforeQty"`       // 修复前 Redis 库存
	ChangeQty            int       `gorm:"column:change_qty" json:"changeQty"`       // 变更数量
	AfterQty             int       `gorm:"column:after_qty" json:"afterQty"`         // 修复后 Redis 库存
	TraceID              int64     `gorm:"column:trace_id" json:"traceId"`
	LogType              int       `gorm:"column:log_type" json:"logType"`
	BusinessType         int       `gorm:"column:business_type" json:"businessType"`
	ReconciliationStatus int       `gorm:"column:reconciliation_status" json:"reconciliationStatus"`
	CreateTime           time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime           time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (VoucherReconcileLog) TableName() string { return "tb_voucher_reconcile_log" }

// VoucherSubscribe 用户订阅/候补记录，对应 tb_voucher_subscribe
// 秒杀售罄后用户可登记候补，有库存释放时自动通知
type VoucherSubscribe struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	UserID     int64     `gorm:"column:user_id" json:"userId"`
	VoucherID  int64     `gorm:"column:voucher_id" json:"voucherId"`
	Status     int       `gorm:"column:status" json:"status"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (VoucherSubscribe) TableName() string { return "tb_voucher_subscribe" }

// RollbackFailureLog 回滚失败记录，对应 tb_rollback_failure_log
// 当秒杀回滚（补偿库存）操作失败时记录，供人工介入处理
type RollbackFailureLog struct {
	ID           int64     `gorm:"column:id;primaryKey" json:"id"`
	VoucherID    int64     `gorm:"column:voucher_id" json:"voucherId"`
	UserID       int64     `gorm:"column:user_id" json:"userId"`
	OrderID      int64     `gorm:"column:order_id" json:"orderId"`
	TraceID      int64     `gorm:"column:trace_id" json:"traceId"`
	Detail       string    `gorm:"column:detail;size:1024" json:"detail"`
	ResultCode   int       `gorm:"column:result_code" json:"resultCode"`
	RetryAttempts int      `gorm:"column:retry_attempts" json:"retryAttempts"` // 已重试次数
	Source       string    `gorm:"column:source;size:64" json:"source"`
	CreateTime   time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime   time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (RollbackFailureLog) TableName() string { return "tb_rollback_failure_log" }
