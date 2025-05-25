package upload

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler 文件上传处理器
type Handler struct {
	uploadDir string
}

// NewHandler 创建上传处理器
func NewHandler() *Handler {
	uploadDir := "uploads"
	// 确保上传目录存在
	os.MkdirAll(uploadDir, 0755)
	return &Handler{uploadDir: uploadDir}
}

// RegisterRoutes 注册上传路由
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/image", h.UploadImage)
	rg.POST("/video", h.UploadVideo)
}

// UploadImage 上传图片
func (h *Handler) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取文件失败"})
		return
	}
	defer file.Close()

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的图片格式"})
		return
	}

	// 验证文件大小（最大5MB）
	if header.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "图片大小不能超过5MB"})
		return
	}

	// 生成文件名
	filename := h.generateFilename("image", ext)
	filePath := filepath.Join(h.uploadDir, "images", filename)

	// 确保目录存在
	os.MkdirAll(filepath.Dir(filePath), 0755)

	// 保存文件
	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建文件失败"})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	// 返回文件URL
	url := fmt.Sprintf("/uploads/images/%s", filename)
	c.JSON(http.StatusOK, gin.H{
		"url":      url,
		"filename": filename,
		"size":     header.Size,
	})
}

// UploadVideo 上传视频
func (h *Handler) UploadVideo(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取文件失败"})
		return
	}
	defer file.Close()

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{
		".mp4":  true,
		".avi":  true,
		".mov":  true,
		".wmv":  true,
		".flv":  true,
		".webm": true,
	}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的视频格式"})
		return
	}

	// 验证文件大小（最大50MB）
	if header.Size > 50*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "视频大小不能超过50MB"})
		return
	}

	// 生成文件名
	filename := h.generateFilename("video", ext)
	filePath := filepath.Join(h.uploadDir, "videos", filename)

	// 确保目录存在
	os.MkdirAll(filepath.Dir(filePath), 0755)

	// 保存文件
	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建文件失败"})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	// 返回文件URL
	url := fmt.Sprintf("/uploads/videos/%s", filename)
	c.JSON(http.StatusOK, gin.H{
		"url":      url,
		"filename": filename,
		"size":     header.Size,
		"duration": 0, // 视频时长需要额外处理
	})
}

// generateFilename 生成唯一文件名
func (h *Handler) generateFilename(prefix, ext string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s_%d%s", prefix, timestamp, ext)
}
