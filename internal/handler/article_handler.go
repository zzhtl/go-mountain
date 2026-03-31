package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zzhtl/go-mountain/internal/model"
	"github.com/zzhtl/go-mountain/internal/pkg/response"
	"github.com/zzhtl/go-mountain/internal/service"
)

// ArticleHandler 文章处理器
type ArticleHandler struct {
	svc *service.ArticleService
}

// NewArticleHandler 创建文章处理器
func NewArticleHandler(svc *service.ArticleService) *ArticleHandler {
	return &ArticleHandler{svc: svc}
}

// List 获取文章列表（后台）
func (h *ArticleHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	columnID, _ := strconv.ParseInt(c.Query("column_id"), 10, 64)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	list, total, err := h.svc.List(c.Request.Context(), page, pageSize, columnID, status)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.PageOK(c, list, total, page, pageSize)
}

// Get 获取文章详情（后台）
func (h *ArticleHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	item, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "文章不存在")
		return
	}

	response.OK(c, item)
}

// Create 创建文章
func (h *ArticleHandler) Create(c *gin.Context) {
	var req struct {
		ColumnID  int64  `json:"column_id" binding:"required"`
		Title     string `json:"title" binding:"required"`
		Thumbnail string `json:"thumbnail"`
		Content   string `json:"content"`
		Author    string `json:"author"`
		Status    int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	article := &model.Article{
		ColumnID:  req.ColumnID,
		Title:     req.Title,
		Thumbnail: req.Thumbnail,
		Content:   req.Content,
		Author:    req.Author,
		Status:    req.Status,
	}

	if err := h.svc.Create(c.Request.Context(), article); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Created(c, article)
}

// Update 更新文章
func (h *ArticleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req struct {
		ColumnID  int64  `json:"column_id" binding:"required"`
		Title     string `json:"title" binding:"required"`
		Thumbnail string `json:"thumbnail"`
		Content   string `json:"content"`
		Author    string `json:"author"`
		Status    int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updates := map[string]any{
		"column_id": req.ColumnID,
		"title":     req.Title,
		"thumbnail": req.Thumbnail,
		"content":   req.Content,
		"author":    req.Author,
		"status":    req.Status,
	}

	if err := h.svc.Update(c.Request.Context(), id, updates); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.OK(c, gin.H{"id": id})
}

// UpdateStatus 更新文章状态
func (h *ArticleHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.OK(c, gin.H{"id": id})
}

// Delete 删除文章
func (h *ArticleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.NoContent(c)
}

// ListByColumn 根据栏目获取文章（小程序）
func (h *ArticleHandler) ListByColumn(c *gin.Context) {
	columnID, err := strconv.ParseInt(c.Param("columnId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的栏目ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.svc.ListByColumn(c.Request.Context(), columnID, page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.PageOK(c, list, total, page, pageSize)
}

// GetForMP 获取文章详情（小程序）
func (h *ArticleHandler) GetForMP(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	item, err := h.svc.GetForMP(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "文章不存在")
		return
	}

	response.OK(c, item)
}
