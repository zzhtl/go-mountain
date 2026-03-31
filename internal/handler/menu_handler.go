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

// MenuHandler 菜单处理器
type MenuHandler struct {
	svc *service.MenuService
}

// NewMenuHandler 创建菜单处理器
func NewMenuHandler(svc *service.MenuService) *MenuHandler {
	return &MenuHandler{svc: svc}
}

// List 获取菜单列表
func (h *MenuHandler) List(c *gin.Context) {
	menus, err := h.svc.List(c.Request.Context())
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, menus)
}

// Tree 获取菜单树形结构
func (h *MenuHandler) Tree(c *gin.Context) {
	tree, err := h.svc.Tree(c.Request.Context())
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, tree)
}

// Get 获取单个菜单
func (h *MenuHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	menu, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "菜单不存在")
		return
	}

	response.OK(c, menu)
}

// Create 创建菜单
func (h *MenuHandler) Create(c *gin.Context) {
	var req struct {
		ParentID   int64  `json:"parent_id"`
		Name       string `json:"name" binding:"required"`
		Title      string `json:"title" binding:"required"`
		Path       string `json:"path"`
		Component  string `json:"component"`
		Icon       string `json:"icon"`
		Sort       int    `json:"sort"`
		Type       int    `json:"type"`
		Permission string `json:"permission"`
		Method     string `json:"method"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	menu := &model.Menu{
		ParentID:   req.ParentID,
		Name:       req.Name,
		Title:      req.Title,
		Path:       req.Path,
		Component:  req.Component,
		Icon:       req.Icon,
		Sort:       req.Sort,
		Type:       req.Type,
		Permission: req.Permission,
		Method:     req.Method,
	}

	if err := h.svc.Create(c.Request.Context(), menu); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Created(c, menu)
}

// Update 更新菜单
func (h *MenuHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req struct {
		ParentID   int64  `json:"parent_id"`
		Name       string `json:"name" binding:"required"`
		Title      string `json:"title" binding:"required"`
		Path       string `json:"path"`
		Component  string `json:"component"`
		Icon       string `json:"icon"`
		Sort       int    `json:"sort"`
		Type       int    `json:"type"`
		Permission string `json:"permission"`
		Method     string `json:"method"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updates := map[string]any{
		"parent_id":  req.ParentID,
		"name":       req.Name,
		"title":      req.Title,
		"path":       req.Path,
		"component":  req.Component,
		"icon":       req.Icon,
		"sort":       req.Sort,
		"type":       req.Type,
		"permission": req.Permission,
		"method":     req.Method,
	}

	if err := h.svc.Update(c.Request.Context(), id, updates); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.OK(c, gin.H{"id": id})
}

// UpdateStatus 更新菜单状态
func (h *MenuHandler) UpdateStatus(c *gin.Context) {
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

	if req.Status != 0 && req.Status != 1 {
		response.BadRequest(c, "状态只能是0或1")
		return
	}

	if err := h.svc.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.OK(c, gin.H{"id": id})
}

// Delete 删除菜单
func (h *MenuHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, errcode.ErrMenuHasChildren) {
			response.BadRequest(c, err.Error())
			return
		}
		response.ServerError(c, err.Error())
		return
	}

	response.NoContent(c)
}
