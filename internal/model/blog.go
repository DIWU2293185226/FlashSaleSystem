// ═════════════════════════════════════════════════════════════════════
// 社交模块数据模型
// Blog 包含点赞数、评论数，支持热门排序和关注推送
// Icon/Name/IsLike 是运行时填充的非 DB 字段
// ═════════════════════════════════════════════════════════════════════
package model

import "time"

// Blog 探店博客，对应 tb_blog
// 用户可发布探店图文，其他用户点赞、评论
// Liked 为数据库计数器，同时配合 Redis Set 做点赞去重
type Blog struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	ShopID     int64     `gorm:"column:shop_id" json:"shopId"`
	UserID     int64     `gorm:"column:user_id" json:"userId"`
	Title      string    `gorm:"column:title;size:255" json:"title"`
	Images     string    `gorm:"column:images;size:2048" json:"images"`
	Content    string    `gorm:"column:content;size:2048" json:"content"`
	Liked      int       `gorm:"column:liked" json:"liked"`   // 点赞数（DB 兜底 + Redis 实时）
	Comments   int       `gorm:"column:comments" json:"comments"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`

	// 非 DB 字段，由查询时按需拼装
	Icon   string `gorm:"-" json:"icon,omitempty"`   // 博主头像
	Name   string `gorm:"-" json:"name,omitempty"`   // 博主昵称
	IsLike bool   `gorm:"-" json:"isLike,omitempty"` // 当前登录用户是否已点赞
}

func (Blog) TableName() string {
	return "tb_blog"
}
