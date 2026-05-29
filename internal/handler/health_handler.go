// ═════════════════════════════════════════════════════════════════════
// 健康检查处理器 — 用于 Kubernetes 探活和负载均衡健康检测
// 提供最简的 ping/pong 接口，不依赖任何外部组件
// ═════════════════════════════════════════════════════════════════════
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/javaup/flashsale-system/pkg"
)

// HealthHandler 健康检查处理器
type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Ping 健康检查接口，返回 "pong" 表示服务正常
func (h *HealthHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, pkg.OKWithData("pong"))
}
