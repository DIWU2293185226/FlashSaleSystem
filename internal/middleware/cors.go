// ═════════════════════════════════════════════════════════════════════
// CORS 跨域中间件
// 前端在开发时通过 Vite 代理转发请求本身不跨域，
// 但上线后前端可能部署在不同域名下，仍需要 CORS 支持
// 也方便前后端分离开发时直接用浏览器访问后端地址调试
// ═════════════════════════════════════════════════════════════════════
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS 返回一个允许跨域访问的 Gin 中间件
// 允许所有来源（开发环境）、常见 HTTP 方法和自定义头
// OPTIONS 预检请求直接返回 204，不继续执行后续 handler
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, token")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin")
		c.Header("Access-Control-Max-Age", "3600")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
