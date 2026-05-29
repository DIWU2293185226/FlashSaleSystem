// ═════════════════════════════════════════════════════════════════════
// 请求日志中间件 — 基于 zerolog 结构化日志
// 记录每次 HTTP 请求的方法、路径、状态码、耗时、客户端 IP
// 根据状态码自动选择日志级别：5xx→Error, 4xx→Warn, 其他→Info
// ═════════════════════════════════════════════════════════════════════
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Logger 返回一个基于 zerolog 的请求日志中间件
func Logger(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		logger := log.Info()
		if status >= 500 {
			logger = log.Error()
		} else if status >= 400 {
			logger = log.Warn()
		}

		logger.
			Str("method", c.Request.Method).
			Str("path", path).
			Str("query", query).
			Int("status", status).
			Dur("latency", latency).
			Str("ip", c.ClientIP()).
			Msg("request")
	}
}
