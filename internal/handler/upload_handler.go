// ═════════════════════════════════════════════════════════════════════
// 文件上传 HTTP 处理器 — 博客图片上传/删除
// 图片按日期分目录存储（YYYYMMDD），文件名使用纳秒时间戳防冲突
// 返回相对 URL 路径，由 Gin 静态文件服务提供访问
// ═════════════════════════════════════════════════════════════════════
package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/javaup/flashsale-system/internal/service"
	"github.com/javaup/flashsale-system/pkg"
)

// UploadHandler 文件上传 HTTP 处理器
type UploadHandler struct {
	svc *service.UploadService
}

func NewUploadHandler(svc *service.UploadService) *UploadHandler {
	return &UploadHandler{svc: svc}
}

// UploadBlog 上传博客图片 [AuthRequired]
// POST /upload/blog
// 从 multipart/form-data 中读取 "file" 字段，保存后返回可访问的 URL
func (h *UploadHandler) UploadBlog(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("请选择文件"))
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("读取文件失败"))
		return
	}
	c.JSON(http.StatusOK, h.svc.UploadBlogImage(data, header.Filename))
}

// DeleteBlog 删除已上传的博客图片 [AuthRequired]
// DELETE /upload/blog/delete?name=xxx
func (h *UploadHandler) DeleteBlog(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, pkg.FailWithMsg("参数错误"))
		return
	}
	c.JSON(http.StatusOK, h.svc.DeleteBlogImage(name))
}
