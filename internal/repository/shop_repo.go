// ═════════════════════════════════════════════════════════════════════
// 商铺相关仓库 — tb_shop / tb_shop_type（广播表，不分片）
// 商铺和商铺类型是基础数据，量小且以读为主，直接放在 hmdp_0
// 所有的查询都走 broadcastDB 常量
// ═════════════════════════════════════════════════════════════════════
package repository

import (
	"github.com/javaup/flashsale-system/internal/model"
	"github.com/javaup/flashsale-system/internal/sharding"
	"gorm.io/gorm"
)

// ShopRepository tb_shop 表的数据访问（广播表，不分片）
type ShopRepository struct {
	dm     *DatabaseManager
	router *sharding.Router
}

func NewShopRepository(dm *DatabaseManager, router *sharding.Router) *ShopRepository {
	return &ShopRepository{dm: dm, router: router}
}

func (r *ShopRepository) GetByID(id int64) (*model.Shop, error) {
	db := r.dm.GetDB(broadcastDB)
	var shop model.Shop
	err := db.First(&shop, id).Error
	return &shop, err
}

func (r *ShopRepository) Create(shop *model.Shop) error {
	db := r.dm.GetDB(broadcastDB)
	return db.Create(shop).Error
}

func (r *ShopRepository) Update(shop *model.Shop) error {
	db := r.dm.GetDB(broadcastDB)
	return db.Where("id = ?", shop.ID).Updates(shop).Error
}

// GetDB 返回广播数据库连接，供上层 service 做跨仓库联合查询
func (r *ShopRepository) GetDB() *gorm.DB {
	return r.dm.GetDB(broadcastDB)
}

func (r *ShopRepository) Delete(id int64) error {
	db := r.dm.GetDB(broadcastDB)
	return db.Delete(&model.Shop{}, id).Error
}

// ListByType 按商铺类型分页查询
func (r *ShopRepository) ListByType(typeID int64, offset, limit int) ([]model.Shop, int64, error) {
	db := r.dm.GetDB(broadcastDB)
	var shops []model.Shop
	var total int64
	query := db.Model(&model.Shop{}).Where("type_id = ?", typeID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Offset(offset).Limit(limit).Find(&shops).Error
	return shops, total, err
}

// ShopTypeRepository tb_shop_type 表的数据访问（广播表，不分片）
// 商铺类型很少变更，可按需配合缓存使用
type ShopTypeRepository struct {
	dm *DatabaseManager
}

func NewShopTypeRepository(dm *DatabaseManager) *ShopTypeRepository {
	return &ShopTypeRepository{dm: dm}
}

// ListAll 查询所有商铺类型，按 sort 字段升序排列
func (r *ShopTypeRepository) ListAll() ([]model.ShopType, error) {
	db := r.dm.GetDB(broadcastDB)
	var types []model.ShopType
	err := db.Order("sort ASC").Find(&types).Error
	return types, err
}
