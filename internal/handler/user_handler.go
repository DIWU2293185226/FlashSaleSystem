// ═════════════════════════════════════════════════════════════════════
// 用户模块 HTTP 处理器 — 登录/登出/信息查询/签到
// 认证路由通过 AuthRequired 中间件保护，userID 从 JWT Claims 中提取
// ═════════════════════════════════════════════════════════════════════
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/javaup/flashsale-system/internal/dto"
	"github.com/javaup/flashsale-system/internal/middleware"
	"github.com/javaup/flashsale-system/internal/service"
	"github.com/javaup/flashsale-system/pkg"
)

// UserHandler 用户模块 HTTP 处理器
type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// SendCode 发送手机验证码
// GET /user/code?phone=138xxxx
func (h *UserHandler) SendCode(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("手机号不能为空"))
		return
	}
	c.JSON(http.StatusOK, h.svc.SendCode(phone))
}

// Login 用户登录（验证码登录/密码登录）
// POST /user/login
// 登录成功后返回 JWT token，前端存储在 header 中用于后续请求认证
func (h *UserHandler) Login(c *gin.Context) {
	var form dto.LoginFormDTO
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.Login(&form))
}

// Logout 用户登出
// 当前实现为无操作 — JWT 无状态 token 的失效由前端清除 token 完成
// 如需严格登出可在 Redis 维护黑名单
func (h *UserHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, pkg.OK())
}

// GetMe 获取当前登录用户信息
// [AuthRequired] 从 JWT 中提取 userID 查询
func (h *UserHandler) GetMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, pkg.FailWithMsg("请先登录"))
		return
	}
	c.JSON(http.StatusOK, h.svc.GetByID(userID))
}

// GetByID 根据用户 ID 查询用户
// 用于在其他模块中展示用户信息（如博客作者信息）
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.GetByID(id))
}

// GetInfoByID 查询用户详细信息（包括个性签名、城市等）
func (h *UserHandler) GetInfoByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.GetUserInfo(id))
}

// Sign 用户签到（每日签到，Redis Bitmap 存储）
// [AuthRequired] 每个用户每天只能签到一次
func (h *UserHandler) Sign(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, pkg.FailWithMsg("请先登录"))
		return
	}
	c.JSON(http.StatusOK, h.svc.Sign(userID))
}

// SignCount 查询本月连续签到天数
// [AuthRequired] 使用 BITFIELD 命令从右往左统计连续的 1
func (h *UserHandler) SignCount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, pkg.FailWithMsg("请先登录"))
		return
	}
	c.JSON(http.StatusOK, h.svc.SignCount(userID))
}
