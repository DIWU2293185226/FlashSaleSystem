// ═════════════════════════════════════════════════════════════════════
// 秒杀模块 HTTP 处理器 — 秒杀下单/订单查询/取消/库存管理/令牌生成
// 秒杀流程：限流检查 → 布隆过滤器 → Lua 原子操作 → Kafka 异步落库
// 完整链路涉及限流、防重、一人一单、库存扣减、异步下单、回滚等
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

// SeckillHandler 秒杀订单 HTTP 处理器
type SeckillHandler struct {
	svc *service.SeckillService
}

func NewSeckillHandler(svc *service.SeckillService) *SeckillHandler {
	return &SeckillHandler{svc: svc}
}

// Seckill 执行秒杀操作 [AuthRequired]
// POST /voucher-order/seckill
// 请求体: {"voucherId": 123, "token": "xxx"}（token 可选，用于访问令牌验证）
// 完整流程：限流 → 布隆过滤器 → 幂等防重 → Lua 原子扣库存 → Kafka 异步创建订单
func (h *SeckillHandler) Seckill(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, pkg.FailWithMsg("请先登录"))
		return
	}

	var req struct {
		VoucherID int64  `json:"voucherId" binding:"required"`
		Token     string `json:"token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}

	// 如果提供了访问令牌，先校验令牌有效性
	// 令牌模式用于前端防刷：先申请令牌，秒杀时携带令牌，服务端校验
	if req.Token != "" {
		if !h.svc.ValidateAccessToken(req.VoucherID, userID, req.Token) {
			c.JSON(http.StatusForbidden, pkg.FailWithMsg("令牌无效"))
			return
		}
	}

	ip := c.ClientIP()
	c.JSON(http.StatusOK, h.svc.SeckillVoucher(req.VoucherID, userID, ip))
}

// GetOrder 查询秒杀订单详情
// GET /voucher-order/:orderId?voucherId=X&userId=Y
func (h *SeckillHandler) GetOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("orderId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	voucherID, _ := strconv.ParseInt(c.Query("voucherId"), 10, 64)
	userID, _ := strconv.ParseInt(c.Query("userId"), 10, 64)

	c.JSON(http.StatusOK, h.svc.GetOrder(orderID, voucherID, userID))
}

// CancelOrder 取消秒杀订单并回滚库存 [AuthRequired]
// PUT /voucher-order/cancel/:orderId
// 取消后会自动恢复 Redis 库存和 DB 库存，同时从 Redis 已购集合中移除用户
func (h *SeckillHandler) CancelOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, pkg.FailWithMsg("请先登录"))
		return
	}

	orderID, err := strconv.ParseInt(c.Param("orderId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}

	var req struct {
		VoucherID int64 `json:"voucherId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}

	c.JSON(http.StatusOK, h.svc.CancelOrder(orderID, req.VoucherID, userID))
}

// LoadStock 将秒杀券的 DB 库存预热到 Redis
// POST /voucher-order/load-stock
// 通常在秒杀活动开始前调用，将库存从 DB 加载到 Redis 以提升秒杀性能
func (h *SeckillHandler) LoadStock(c *gin.Context) {
	var req struct {
		VoucherID int64 `json:"voucherId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.LoadSeckillStock(req.VoucherID))
}

// GetStock 查询 Redis 中的秒杀库存
// GET /voucher-order/stock/:voucherId
func (h *SeckillHandler) GetStock(c *gin.Context) {
	voucherID, err := strconv.ParseInt(c.Param("voucherId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.GetStock(voucherID))
}

// GetSeckillVoucherFull 查询完整的秒杀券信息（含商品信息和秒杀信息）
// GET /voucher-order/voucher/:voucherId
func (h *SeckillHandler) GetSeckillVoucherFull(c *gin.Context) {
	voucherID, err := strconv.ParseInt(c.Param("voucherId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.GetSeckillVoucherFull(voucherID))
}

// GenerateToken 生成秒杀访问令牌 [AuthRequired]
// POST /voucher-order/token
// 令牌机制：用户先申请令牌（限流检查通过才发放），秒杀时携带令牌
// 这样可以缓解秒杀接口的直接压力，将限流前置到令牌申请阶段
func (h *SeckillHandler) GenerateToken(c *gin.Context) {
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
	c.JSON(http.StatusOK, h.svc.GenerateAccessToken(req.VoucherID, userID))
}
