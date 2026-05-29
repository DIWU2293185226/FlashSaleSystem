// ═════════════════════════════════════════════════════════════════════
// 优惠券模块 HTTP 处理器 — 普通券/秒杀券的增删改查
// 秒杀券需要同时写入 tb_voucher 和 tb_seckill_voucher 两张表
// 通过 shopId 查优惠券时需要遍历全部分片再聚合结果
// ═════════════════════════════════════════════════════════════════════
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/javaup/flashsale-system/internal/dto"
	"github.com/javaup/flashsale-system/internal/model"
	"github.com/javaup/flashsale-system/internal/service"
	"github.com/javaup/flashsale-system/pkg"
)

// VoucherHandler 优惠券模块 HTTP 处理器
type VoucherHandler struct {
	svc *service.VoucherService
}

func NewVoucherHandler(svc *service.VoucherService) *VoucherHandler {
	return &VoucherHandler{svc: svc}
}

// AddNormal 新增普通优惠券
// POST /voucher
func (h *VoucherHandler) AddNormal(c *gin.Context) {
	var v model.Voucher
	if err := c.ShouldBindJSON(&v); err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.AddNormal(&v))
}

// AddSeckill 新增秒杀优惠券（同时写入 tb_voucher + tb_seckill_voucher）
// POST /voucher/seckill
func (h *VoucherHandler) AddSeckill(c *gin.Context) {
	var req dto.SeckillVoucherDto
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.AddSeckill(&req))
}

// GetByID 查询优惠券详情，秒杀券附带秒杀信息
// POST /voucher/get?id=xxx
func (h *VoucherHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.GetByID(id))
}

// ListByShopID 查询指定商铺的优惠券列表（跨分片聚合）
// GET /voucher/list/{shopId}
func (h *VoucherHandler) ListByShopID(c *gin.Context) {
	shopID, err := strconv.ParseInt(c.Param("shopId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.ListByShopID(shopID))
}

// UpdateSeckill 更新秒杀优惠券基础信息
// POST /voucher/update/seckill
func (h *VoucherHandler) UpdateSeckill(c *gin.Context) {
	var v model.Voucher
	if err := c.ShouldBindJSON(&v); err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.UpdateSeckill(&v))
}

// UpdateSeckillStock 更新秒杀优惠券库存（增量更新）
// POST /voucher/update/seckill/stock
func (h *VoucherHandler) UpdateSeckillStock(c *gin.Context) {
	var req struct {
		VoucherID int64 `json:"voucherId"`
		Stock     int   `json:"stock"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.UpdateSeckillStock(req.VoucherID, req.Stock))
}
