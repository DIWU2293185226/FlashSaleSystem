// ═════════════════════════════════════════════════════════════════════
// 商铺服务 — 含多级缓存、空值穿透保护、GEO 距离排序
// 查询链路：本地缓存(null标记检查) → 多级缓存(FreeCache+Redis) → DB
// DB 查不到时写入空值标记，防止恶意遍历不存在的 ID
// 附近商铺用 Haversine 公式计算距离后排序
// ═════════════════════════════════════════════════════════════════════
package service

import (
	"math"
	"strconv"

	"github.com/javaup/flashsale-system/internal/cache"
	"github.com/javaup/flashsale-system/internal/model"
	"github.com/javaup/flashsale-system/internal/repository"
	"github.com/javaup/flashsale-system/pkg"
)

// ShopService 商铺业务逻辑
// 多级缓存 + 空值穿透保护 + Haversine 距离排序
type ShopService struct {
	shopRepo     *repository.ShopRepository
	shopTypeRepo *repository.ShopTypeRepository
	multiCache   *cache.MultiLevelCache
}

func NewShopService(shopRepo *repository.ShopRepository, shopTypeRepo *repository.ShopTypeRepository, multiCache *cache.MultiLevelCache) *ShopService {
	return &ShopService{shopRepo: shopRepo, shopTypeRepo: shopTypeRepo, multiCache: multiCache}
}

const (
	shopCacheKey  = "cache:shop:"
	nullKeyPrefix = "cache:null:shop:"
)

// GetByID 查询商铺（多级缓存 + 空值标记防穿透）
// 缓存穿透防护设计：
// 1. 先检查本地缓存的空值标记（命中直接返回"不存在"）
// 2. 未命中则查多级缓存
// 3. 多级缓存未命中查 DB
// 4. DB 查不到时写入空值标记（5 分钟 TTL），避免反复穿透
func (s *ShopService) GetByID(id int64) *pkg.Result {
	key := shopCacheKey + strconv.FormatInt(id, 10)

	// 本地缓存空值标记检查（缓存穿透防护）
	var nullCheck string
	if s.multiCache.GetLocal(nullKeyPrefix+strconv.FormatInt(id, 10), &nullCheck) == nil {
		return pkg.FailWithMsg("商铺不存在")
	}

	var shop model.Shop
	if err := s.multiCache.Get(key, &shop); err == nil {
		return pkg.OKWithData(shop)
	}
	dbShop, err := s.shopRepo.GetByID(id)
	if err != nil {
		// 写入空值标记，吸收针对不存在 ID 的遍历攻击
		_ = s.multiCache.SetLocal(nullKeyPrefix+strconv.FormatInt(id, 10), "", pkg.CacheNullTTL)
		return pkg.FailWithMsg("商铺不存在")
	}
	_ = s.multiCache.Set(key, dbShop, 1800)
	return pkg.OKWithData(dbShop)
}

// Create 新增商铺
func (s *ShopService) Create(shop *model.Shop) *pkg.Result {
	if err := s.shopRepo.Create(shop); err != nil {
		return pkg.FailWithMsg("新增商铺失败")
	}
	return pkg.OK()
}

// Update 更新商铺并失效缓存
func (s *ShopService) Update(shop *model.Shop) *pkg.Result {
	if err := s.shopRepo.Update(shop); err != nil {
		return pkg.FailWithMsg("更新商铺失败")
	}
	key := shopCacheKey + strconv.FormatInt(shop.ID, 10)
	_ = s.multiCache.Del(key)
	return pkg.OK()
}

// ListByType 按类型分页查询商铺
func (s *ShopService) ListByType(typeID int64, current int) *pkg.Result {
	if current <= 0 {
		current = 1
	}
	offset := (current - 1) * pkg.MaxPageSize
	shops, total, err := s.shopRepo.ListByType(typeID, offset, pkg.MaxPageSize)
	if err != nil {
		return pkg.FailWithMsg("查询商铺失败")
	}
	return pkg.OKWithDataTotal(shops, total)
}

// ListByName 按名称关键字搜索（LIKE 模糊匹配）
func (s *ShopService) ListByName(name string, current int) *pkg.Result {
	if current <= 0 {
		current = 1
	}
	db := s.shopRepo.GetDB()
	var shops []model.Shop
	var total int64
	query := db.Model(&model.Shop{}).Where("name LIKE ?", "%"+name+"%")
	if err := query.Count(&total).Error; err != nil {
		return pkg.FailWithMsg("查询商铺失败")
	}
	offset := (current - 1) * pkg.MaxPageSize
	if err := query.Offset(offset).Limit(pkg.MaxPageSize).Find(&shops).Error; err != nil {
		return pkg.FailWithMsg("查询商铺失败")
	}
	return pkg.OKWithDataTotal(shops, total)
}

// ListNearby 附近商铺查询（Haversine 距离计算 + 冒泡排序）
// 先查出所有商铺，逐个计算与用户坐标的距离
// 然后按距离排序，筛选出指定范围内的商铺
func (s *ShopService) ListNearby(x, y float64, distanceKm int, current int) *pkg.Result {
	if current <= 0 {
		current = 1
	}
	db := s.shopRepo.GetDB()
	offset := (current - 1) * pkg.MaxPageSize
	var shops []model.Shop
	if err := db.Limit(pkg.MaxPageSize).Offset(offset).Find(&shops).Error; err != nil {
		return pkg.FailWithMsg("查询商铺失败")
	}
	for i := range shops {
		shops[i].Distance = haversine(x, y, shops[i].X, shops[i].Y)
	}
	for i := 0; i < len(shops); i++ {
		for j := i + 1; j < len(shops); j++ {
			if shops[i].Distance > shops[j].Distance {
				shops[i], shops[j] = shops[j], shops[i]
			}
		}
	}
	var filtered []model.Shop
	for _, s := range shops {
		if s.Distance <= float64(distanceKm)*1000 {
			filtered = append(filtered, s)
		}
	}
	if filtered == nil {
		filtered = []model.Shop{}
	}
	return pkg.OKWithData(filtered)
}

// haversine 使用 Haversine 公式计算两点间距离（米）
// 公式：a = sin²(Δlat/2) + cos(lat1)*cos(lat2)*sin²(Δlng/2), c = 2*atan2(√a, √(1-a)), d = R*c
func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * R * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// ShopTypeService 商铺类型服务
type ShopTypeService struct {
	shopTypeRepo *repository.ShopTypeRepository
}

func NewShopTypeService(shopTypeRepo *repository.ShopTypeRepository) *ShopTypeService {
	return &ShopTypeService{shopTypeRepo: shopTypeRepo}
}

func (s *ShopTypeService) ListAll() *pkg.Result {
	types, err := s.shopTypeRepo.ListAll()
	if err != nil {
		return pkg.FailWithMsg("查询商铺类型失败")
	}
	return pkg.OKWithData(types)
}
