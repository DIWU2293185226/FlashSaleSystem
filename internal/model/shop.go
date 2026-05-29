// ═════════════════════════════════════════════════════════════════════
// 商铺相关数据模型
// Shop 包含经纬度用于 Haversine 距离计算和 GEO 排序
// Distance 字段标记 gorm:"-" 不由 ORM 管理，仅用于 API 响应
// ═════════════════════════════════════════════════════════════════════
package model

import "time"

// Shop 商铺模型，对应 tb_shop 表
// X/Y 为高德坐标经纬度，用于附近商铺距离排序
type Shop struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	Name       string    `gorm:"column:name;size:128" json:"name"`
	TypeID     int64     `gorm:"column:type_id" json:"typeId"`
	Images     string    `gorm:"column:images;size:1024" json:"images"`
	Area       string    `gorm:"column:area;size:128" json:"area"`
	Address    string    `gorm:"column:address;size:255" json:"address"`
	X          float64   `gorm:"column:x" json:"x"`     // 经度
	Y          float64   `gorm:"column:y" json:"y"`     // 纬度
	AvgPrice   int64     `gorm:"column:avg_price" json:"avgPrice"`
	Sold       int       `gorm:"column:sold" json:"sold"`
	Comments   int       `gorm:"column:comments" json:"comments"`
	Score      int       `gorm:"column:score" json:"score"`
	OpenHours  string    `gorm:"column:open_hours;size:32" json:"openHours"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`

	// Distance 与当前用户的距离（非数据库字段，运行时计算）
	Distance float64 `gorm:"-" json:"distance,omitempty"`
}

func (Shop) TableName() string {
	return "tb_shop"
}

// ShopType 商铺分类模型
type ShopType struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	Name       string    `gorm:"column:name;size:32" json:"name"`
	Icon       string    `gorm:"column:icon;size:255" json:"icon"`
	Sort       int       `gorm:"column:sort" json:"sort"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"-"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime" json:"-"`
}

func (ShopType) TableName() string {
	return "tb_shop_type"
}
