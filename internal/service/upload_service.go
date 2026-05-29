// ═════════════════════════════════════════════════════════════════════
// 文件上传服务 — 博客图片上传/删除
// 按日期分目录存储（YYYYMMDD），文件名使用纳秒时间戳防冲突
// 返回相对 URL 路径，由 Gin 静态文件服务提供访问
// ═════════════════════════════════════════════════════════════════════
package service

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/javaup/flashsale-system/pkg"
)

// UploadService 文件上传服务
type UploadService struct {
	uploadDir string
}

// NewUploadService 创建上传服务，默认目录 ./uploads/blog
func NewUploadService(uploadDir string) *UploadService {
	if uploadDir == "" {
		uploadDir = "./uploads/blog"
	}
	return &UploadService{uploadDir: uploadDir}
}

// UploadBlogImage 保存上传的博客图片并返回 URL 路径
// 文件名使用纳秒时间戳防止重复和冲突
func (s *UploadService) UploadBlogImage(data []byte, filename string) *pkg.Result {
	dir := filepath.Join(s.uploadDir, time.Now().Format("20060102"))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return pkg.FailWithMsg("上传失败")
	}
	ext := filepath.Ext(filename)
	savedName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(dir, savedName)
	if err := os.WriteFile(savePath, data, 0644); err != nil {
		return pkg.FailWithMsg("上传失败")
	}
	url := fmt.Sprintf("/uploads/blog/%s/%s", time.Now().Format("20060102"), savedName)
	return pkg.OKWithData(url)
}

// DeleteBlogImage 删除已上传的图片
func (s *UploadService) DeleteBlogImage(name string) *pkg.Result {
	path := filepath.Join(s.uploadDir, name)
	if err := os.Remove(path); err != nil {
		return pkg.FailWithMsg("删除失败")
	}
	return pkg.OK()
}
