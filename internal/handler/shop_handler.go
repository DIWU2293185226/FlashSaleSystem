// ═════════════════════════════════════════════════════════════════════
// 商铺模块 HTTP 处理器 — 商铺 CRUD / 分类查询 / 模糊搜索 / 附近 GEO 查询
// 附近商铺使用 Haversine 公式计算距离，支持按距离筛选
// ═════════════════════════════════════════════════════════════════════
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/javaup/flashsale-system/internal/model"
	"github.com/javaup/flashsale-system/internal/service"
	"github.com/javaup/flashsale-system/pkg"
)

// ShopHandler 商铺模块 HTTP 处理器
type ShopHandler struct {
	svc *service.ShopService
}

func NewShopHandler(svc *service.ShopService) *ShopHandler {
	return &ShopHandler{svc: svc}
}

// GetByID 查询商铺详情（走多级缓存）
// GET /shop/{id}
func (h *ShopHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.GetByID(id))
}

// Create 新增商铺
// POST /shop
func (h *ShopHandler) Create(c *gin.Context) {
	var shop model.Shop
	if err := c.ShouldBindJSON(&shop); err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.Create(&shop))
}

// Update 更新商铺信息，同时失效缓存
// PUT /shop
func (h *ShopHandler) Update(c *gin.Context) {
	var shop model.Shop
	if err := c.ShouldBindJSON(&shop); err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.Update(&shop))
}

// ListByType 按商铺类型分页查询
// GET /shop/of/type?typeId=1&current=1
func (h *ShopHandler) ListByType(c *gin.Context) {
	typeID, _ := strconv.ParseInt(c.Query("typeId"), 10, 64)
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	c.JSON(http.StatusOK, h.svc.ListByType(typeID, current))
}

// ListByName 按商铺名称模糊搜索
// GET /shop/of/name?name=华&current=1
func (h *ShopHandler) ListByName(c *gin.Context) {
	name := c.Query("name")
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	c.JSON(http.StatusOK, h.svc.ListByName(name, current))
}

// ListNearby 查询附近的商铺（基于用户坐标）
// GET /shop/of/nearby?x=116.23&y=39.91&distance=5&current=1
// 使用 Haversine 公式计算所有商铺与用户的距离，筛选出 distance 范围内的商铺按距离排序
func (h *ShopHandler) ListNearby(c *gin.Context) {
	x, _ := strconv.ParseFloat(c.Query("x"), 64)
	y, _ := strconv.ParseFloat(c.Query("y"), 64)
	distance, _ := strconv.Atoi(c.DefaultQuery("distance", "5"))
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	c.JSON(http.StatusOK, h.svc.ListNearby(x, y, distance, current))
}

// ShopTypeHandler 商铺类型 HTTP 处理器
type ShopTypeHandler struct {
	svc *service.ShopTypeService
}

func NewShopTypeHandler(svc *service.ShopTypeService) *ShopTypeHandler {
	return &ShopTypeHandler{svc: svc}
}

// ListAll 查询所有商铺类型列表（按 sort 排序）
// GET /shop-type/list
func (h *ShopTypeHandler) ListAll(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.ListAll())
}
