package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zzhtl/go-mountain/internal/model"
	"github.com/zzhtl/go-mountain/internal/pkg/errcode"
	"github.com/zzhtl/go-mountain/internal/pkg/response"
	"github.com/zzhtl/go-mountain/internal/service"
)

// ColumnHandler 栏目处理器
type ColumnHandler struct {
	svc *service.ColumnService
}

// NewColumnHandler 创建栏目处理器
func NewColumnHandler(svc *service.ColumnService) *ColumnHandler {
	return &ColumnHandler{svc: svc}
}

// List 获取所有栏目（后台）
func (h *ColumnHandler) List(c *gin.Context) {
	columns, err := h.svc.List(c.Request.Context())
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, columns)
}

// Get 获取单个栏目
func (h *ColumnHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	column, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "栏目不存在")
		return
	}

	response.OK(c, column)
}

// Create 创建栏目
func (h *ColumnHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		SortOrder   int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	column := &model.Column{
		Name:        req.Name,
		Description: req.Description,
		SortOrder:   req.SortOrder,
	}

	if err := h.svc.Create(c.Request.Context(), column); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Created(c, column)
}

// Update 更新栏目
func (h *ColumnHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		SortOrder   int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updates := map[string]any{
		"name":        req.Name,
		"description": req.Description,
		"sort_order":  req.SortOrder,
	}

	if err := h.svc.Update(c.Request.Context(), id, updates); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.OK(c, gin.H{"id": id})
}

// Delete 删除栏目
func (h *ColumnHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, errcode.ErrColumnHasArticles) {
			response.BadRequest(c, err.Error())
			return
		}
		response.ServerError(c, err.Error())
		return
	}

	response.NoContent(c)
}

// ListForMP 小程序获取栏目列表
func (h *ColumnHandler) ListForMP(c *gin.Context) {
	columns, err := h.svc.ListForMP(c.Request.Context())
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, columns)
}
