// ═════════════════════════════════════════════════════════════════════
// 库存对账 HTTP 处理器 — 检查 Redis 与 DB 库存一致性 / 修复不一致
// 秒杀场景下 Redis 库存和 DB 库存可能因异常（回滚失败、消息丢失）而不一致
// 对账功能可以发现并修复这些差异
// ═════════════════════════════════════════════════════════════════════
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/javaup/flashsale-system/internal/service"
	"github.com/javaup/flashsale-system/pkg"
)

// ReconciliationHandler 库存对账 HTTP 处理器
type ReconciliationHandler struct {
	svc *service.ReconciliationService
}

func NewReconciliationHandler(svc *service.ReconciliationService) *ReconciliationHandler {
	return &ReconciliationHandler{svc: svc}
}

// CheckStock 检查 Redis 库存与 DB 库存是否一致
// GET /reconcile/check/:voucherId
// 返回 Redis 库存和 DB 库存的对比结果，以及差异值
func (h *ReconciliationHandler) CheckStock(c *gin.Context) {
	voucherID, err := strconv.ParseInt(c.Param("voucherId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	result, err := h.svc.CheckStock(voucherID)
	if err != nil {
		c.JSON(http.StatusOK, pkg.FailWithMsg("检查失败"))
		return
	}
	c.JSON(http.StatusOK, pkg.OKWithData(result))
}

// FixStock 将 Redis 库存修复为 DB 库存的值
// POST /reconcile/fix/:voucherId
// 当对账发现不一致时，以 DB 为准覆盖 Redis 中的库存
func (h *ReconciliationHandler) FixStock(c *gin.Context) {
	voucherID, err := strconv.ParseInt(c.Param("voucherId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	if err := h.svc.FixStock(voucherID); err != nil {
		c.JSON(http.StatusOK, pkg.FailWithMsg("修复失败"))
		return
	}
	c.JSON(http.StatusOK, pkg.OK())
}
