// ═════════════════════════════════════════════════════════════════════
// JWT 认证中间件 — 用户身份校验与上下文注入
// 提供两种模式：
//   - AuthRequired：强制校验，无有效 Token 直接拒绝
//   - OptionalAuth：可选校验，有 Token 则解析，没有也能通过
// 前端直接传裸 Token 在 Authorization 头，兼容 Bearer 格式
// ═════════════════════════════════════════════════════════════════════
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/javaup/flashsale-system/internal/jwt"
	"github.com/javaup/flashsale-system/pkg"
)

const (
	// CtxUser 存储当前用户的 Claims（含 userId、nickName、icon）
	CtxUser = "currentUser"
	// CtxUserID 单独提取 userId 方便后续业务直接获取
	CtxUserID = "currentUserId"
)

// AuthRequired 强制认证中间件
// 提取并验证 JWT → 注入用户信息到 Context → 放行
// 验证失败返回 401 + "请先登录"，不继续执行后续 handler
func AuthRequired(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, pkg.FailWithMsg("请先登录"))
			c.Abort()
			return
		}

		claims, err := jwtManager.ParseToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, pkg.FailWithMsg("请先登录"))
			c.Abort()
			return
		}

		c.Set(CtxUser, claims)
		c.Set(CtxUserID, claims.UserID)
		c.Next()
	}
}

// OptionalAuth 可选认证中间件
// 有 Token 就解析并注入用户信息，没有也正常放行
// 用于"登录后可获得更好体验但不是必须"的场景
func OptionalAuth(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.Next()
			return
		}
		claims, err := jwtManager.ParseToken(tokenStr)
		if err == nil {
			c.Set(CtxUser, claims)
			c.Set(CtxUserID, claims.UserID)
		}
		c.Next()
	}
}

// GetUserID 从 Gin Context 提取当前登录用户 ID
// 由 AuthRequired/OptionalAuth 注入，未登录时返回 0
func GetUserID(c *gin.Context) int64 {
	id, exists := c.Get(CtxUserID)
	if !exists {
		return 0
	}
	return id.(int64)
}

// extractToken 从 HTTP Authorization 头中提取 JWT 字符串
// 前端的 Vue3 项目直接传裸 Token（不带 Bearer 前缀）
// 为兼容通用客户端也支持 "Bearer xxx" 格式
func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
		return parts[1]
	}
	return parts[0]
}
