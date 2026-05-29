// ═════════════════════════════════════════════════════════════════════
// 订阅候补 HTTP 处理器 — 缺货登记/取消登记/状态查询/历史查询
// 当秒杀券售罄后，用户可以登记候补，一旦有库存释放（退单）会收到通知
// ═════════════════════════════════════════════════════════════════════
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/javaup/flashsale-system/internal/middleware"
	"github.com/javaup/flashsale-system/internal/service"
	"github.com/javaup/flashsale-system/pkg"
)

// SubscribeHandler 订阅候补 HTTP 处理器
type SubscribeHandler struct {
	svc *service.SubscribeService
}

func NewSubscribeHandler(svc *service.SubscribeService) *SubscribeHandler {
	return &SubscribeHandler{svc: svc}
}

// Subscribe 用户登记候补（缺货登记） [AuthRequired]
// POST /subscribe
// 当秒杀券卖完时可以调用此接口排队候补
func (h *SubscribeHandler) Subscribe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, pkg.FailWithMsg("请先登录"))
		return
	}
	var req struct {
		VoucherID int64 `json:"voucherId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.Subscribe(req.VoucherID, userID))
}

// Unsubscribe 取消候补登记 [AuthRequired]
// DELETE /subscribe/:voucherId
func (h *SubscribeHandler) Unsubscribe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, pkg.FailWithMsg("请先登录"))
		return
	}
	voucherID, err := strconv.ParseInt(c.Param("voucherId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.Unsubscribe(voucherID, userID))
}

// GetStatus 查询当前用户在某优惠券上的候补状态 [AuthRequired]
// GET /subscribe/status/:voucherId
func (h *SubscribeHandler) GetStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, pkg.FailWithMsg("请先登录"))
		return
	}
	voucherID, err := strconv.ParseInt(c.Param("voucherId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.GetSubscribeStatus(voucherID, userID))
}

// GetHistory 查询当前用户的候补登记历史 [AuthRequired]
// GET /subscribe/history?current=1
func (h *SubscribeHandler) GetHistory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, pkg.FailWithMsg("请先登录"))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	c.JSON(http.StatusOK, h.svc.GetSubscribeHistory(userID, page, pkg.MaxPageSize))
}
