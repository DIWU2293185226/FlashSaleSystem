// ═════════════════════════════════════════════════════════════════════
// 博客模块 HTTP 处理器 — 发布/点赞/热门/关注推送
// 点赞使用 Redis Set 做去重（SIsMember 判断 + SAdd/SRem 切换），DB 存计数兜底
// 关注推送使用 scroll 分页，通过 lastId 游标避免数据重复/跳变
// ═════════════════════════════════════════════════════════════════════
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/javaup/flashsale-system/internal/middleware"
	"github.com/javaup/flashsale-system/internal/model"
	"github.com/javaup/flashsale-system/internal/service"
	"github.com/javaup/flashsale-system/pkg"
)

// BlogHandler 博客模块 HTTP 处理器
type BlogHandler struct {
	svc *service.BlogService
}

func NewBlogHandler(svc *service.BlogService) *BlogHandler {
	return &BlogHandler{svc: svc}
}

// Create 发布博客
// POST /blog [AuthRequired]
func (h *BlogHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, pkg.FailWithMsg("请先登录"))
		return
	}
	var blog model.Blog
	if err := c.ShouldBindJSON(&blog); err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.Create(&blog, userID))
}

// Like 点赞/取消点赞切换 [AuthRequired]
// PUT /blog/like/{id}
// 已点赞则取消，未点赞则点赞，通过 Redis SIsMember 判断当前状态
func (h *BlogHandler) Like(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, pkg.FailWithMsg("请先登录"))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.Like(id, userID))
}

// GetByID 博客详情（含博主信息 + 当前用户是否点赞）
// GET /blog/{id}
func (h *BlogHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	userID := middleware.GetUserID(c)
	c.JSON(http.StatusOK, h.svc.GetByID(id, userID))
}

// ListLikes 查询点赞某博客的用户列表
// GET /blog/likes/{id}
func (h *BlogHandler) ListLikes(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.ListLikes(id))
}

// ListHot 热门博客列表（按点赞数 + 发布时间排序）
// GET /blog/hot?current=1
func (h *BlogHandler) ListHot(c *gin.Context) {
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	c.JSON(http.StatusOK, h.svc.ListHot(current))
}

// ListByUserID 查询指定用户的博客列表
// GET /blog/of/user?id=xxx&current=1
func (h *BlogHandler) ListByUserID(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Query("id"), 10, 64)
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	c.JSON(http.StatusOK, h.svc.ListByUserID(userID, current))
}

// ListFollowBlog 关注推送 Feed 流 [AuthRequired]
// GET /blog/of/follow?lastId=xxx&offset=0
// 使用 scroll 分页（通过 lastId 游标），避免传统分页在频繁插入时的重复和跳变
func (h *BlogHandler) ListFollowBlog(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, pkg.FailWithMsg("请先登录"))
		return
	}
	lastID, _ := strconv.ParseInt(c.Query("lastId"), 10, 64)
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	c.JSON(http.StatusOK, h.svc.ListFollowBlog(userID, lastID, offset))
}

// ListMyBlogs 查询当前用户的博客 [AuthRequired]
// GET /blog/of/me?current=1
func (h *BlogHandler) ListMyBlogs(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, pkg.FailWithMsg("请先登录"))
		return
	}
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	c.JSON(http.StatusOK, h.svc.ListByUserID(userID, current))
}
