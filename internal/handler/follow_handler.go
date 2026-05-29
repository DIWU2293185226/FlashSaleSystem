// ═════════════════════════════════════════════════════════════════════
// 关注模块 HTTP 处理器 — 关注/取关/是否已关注/共同关注
// 所有接口需要登录，userID 从 JWT Claims 中提取
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

// FollowHandler 关注模块 HTTP 处理器
type FollowHandler struct {
	svc *service.FollowService
}

func NewFollowHandler(svc *service.FollowService) *FollowHandler {
	return &FollowHandler{svc: svc}
}

// Follow 关注/取关切换 [AuthRequired]
// PUT /follow/{id}/{isFollow}
// isFollow=true 表示关注，false 表示取消关注，不允许关注自己
func (h *FollowHandler) Follow(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, pkg.FailWithMsg("请先登录"))
		return
	}
	followUserID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	isFollow, err := strconv.ParseBool(c.Param("isFollow"))
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.Follow(userID, followUserID, isFollow))
}

// IsFollowed 检查当前用户是否已关注目标用户 [AuthRequired]
// GET /follow/or/not/{id}
func (h *FollowHandler) IsFollowed(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, pkg.FailWithMsg("请先登录"))
		return
	}
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.IsFollowed(userID, targetID))
}

// GetCommon 查询当前用户与目标用户的共同关注 [AuthRequired]
// GET /follow/common/{id}
func (h *FollowHandler) GetCommon(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, pkg.FailWithMsg("请先登录"))
		return
	}
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.GetCommon(userID, targetID))
}
