// ═════════════════════════════════════════════════════════════════════
// 数据模型定义 — 与 MySQL 表结构一一对应
// 使用 GORM 标签映射列名、主键、自动时间戳
// 密码字段标记 json:"-" 防止序列化泄露
// ═════════════════════════════════════════════════════════════════════
package model

import "time"

// User 用户账号模型，对应 tb_user 表
// Password 使用 bcrypt 加密存储，不返回给前端
type User struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	Phone      string    `gorm:"column:phone;size:11" json:"phone"`
	Password   string    `gorm:"column:password;size:128" json:"-"`
	NickName   string    `gorm:"column:nick_name;size:32" json:"nickName"`
	Icon       string    `gorm:"column:icon;size:255" json:"icon"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (User) TableName() string {
	return "tb_user"
}
