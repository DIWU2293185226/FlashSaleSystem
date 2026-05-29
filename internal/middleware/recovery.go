// ═════════════════════════════════════════════════════════════════════
// Panic 恢复中间件
// 任一 handler 出现 panic 时捕获并记录日志，防止整个进程崩溃
// 返回 500 JSON 响应，避免连接悬挂
// ═════════════════════════════════════════════════════════════════════
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/javaup/flashsale-system/pkg"
	"github.com/rs/zerolog"
)

// Recovery 返回一个 panic 恢复中间件
// defer recover 捕获运行时异常，记录出错路径和 panic 信息
func Recovery(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error().
					Interface("panic", err).
					Str("path", c.Request.URL.Path).
					Msg("panic recovered")

				c.AbortWithStatusJSON(http.StatusInternalServerError, pkg.FailWithMsg("服务器内部错误"))
			}
		}()
		c.Next()
	}
}
