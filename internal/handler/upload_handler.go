package handler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/zzhtl/go-mountain/internal/pkg/response"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	uploadDir string
}

// NewUploadHandler 创建上传处理器
func NewUploadHandler() *UploadHandler {
	uploadDir := "uploads"
	os.MkdirAll(uploadDir, 0755)
	return &UploadHandler{uploadDir: uploadDir}
}

// UploadImage 上传图片
func (h *UploadHandler) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "获取文件失败")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedExts[ext] {
		response.BadRequest(c, "不支持的图片格式")
		return
	}

	if header.Size > 5*1024*1024 {
		response.BadRequest(c, "图片大小不能超过5MB")
		return
	}

	filename := h.generateFilename("image", ext)
	filePath := filepath.Join(h.uploadDir, "images", filename)
	os.MkdirAll(filepath.Dir(filePath), 0755)

	out, err := os.Create(filePath)
	if err != nil {
		response.ServerError(c, "创建文件失败")
		return
	}
	defer out.Close()

	if _, err = io.Copy(out, file); err != nil {
		response.ServerError(c, "保存文件失败")
		return
	}

	url := fmt.Sprintf("/uploads/images/%s", filename)
	response.OK(c, gin.H{
		"url":      url,
		"filename": filename,
		"size":     header.Size,
	})
}

// UploadVideo 上传视频
func (h *UploadHandler) UploadVideo(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "获取文件失败")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{".mp4": true, ".avi": true, ".mov": true, ".wmv": true, ".flv": true, ".webm": true}
	if !allowedExts[ext] {
		response.BadRequest(c, "不支持的视频格式")
		return
	}

	if header.Size > 50*1024*1024 {
		response.BadRequest(c, "视频大小不能超过50MB")
		return
	}

	filename := h.generateFilename("video", ext)
	filePath := filepath.Join(h.uploadDir, "videos", filename)
	os.MkdirAll(filepath.Dir(filePath), 0755)

	out, err := os.Create(filePath)
	if err != nil {
		response.ServerError(c, "创建文件失败")
		return
	}
	defer out.Close()

	if _, err = io.Copy(out, file); err != nil {
		response.ServerError(c, "保存文件失败")
		return
	}

	url := fmt.Sprintf("/uploads/videos/%s", filename)
	response.OK(c, gin.H{
		"url":      url,
		"filename": filename,
		"size":     header.Size,
	})
}

func (h *UploadHandler) generateFilename(prefix, ext string) string {
	return fmt.Sprintf("%s_%d%s", prefix, time.Now().UnixNano(), ext)
}
