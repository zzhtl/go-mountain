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

// RoleHandler 角色处理器
type RoleHandler struct {
	svc *service.RoleService
}

// NewRoleHandler 创建角色处理器
func NewRoleHandler(svc *service.RoleService) *RoleHandler {
	return &RoleHandler{svc: svc}
}

// List 获取角色列表
func (h *RoleHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := h.svc.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.PageOK(c, list, total, page, pageSize)
}

// Get 获取单个角色
func (h *RoleHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	role, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "角色不存在")
		return
	}

	response.OK(c, role)
}

// Create 创建角色
func (h *RoleHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		DisplayName string `json:"display_name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	role := &model.Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
	}

	if err := h.svc.Create(c.Request.Context(), role); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Created(c, role)
}

// Update 更新角色
func (h *RoleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		DisplayName string `json:"display_name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updates := map[string]any{
		"name":         req.Name,
		"display_name": req.DisplayName,
		"description":  req.Description,
	}

	if err := h.svc.Update(c.Request.Context(), id, updates); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.OK(c, gin.H{"id": id})
}

// UpdateStatus 更新角色状态
func (h *RoleHandler) UpdateStatus(c *gin.Context) {
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

// Delete 删除角色
func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, errcode.ErrRoleInUse) {
			response.BadRequest(c, err.Error())
			return
		}
		response.ServerError(c, err.Error())
		return
	}

	response.NoContent(c)
}

// GetRoleMenus 获取角色的菜单权限
func (h *RoleHandler) GetRoleMenus(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	menuIDs, err := h.svc.GetRoleMenus(c.Request.Context(), roleID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.OK(c, gin.H{"menu_ids": menuIDs})
}

// UpdateRoleMenus 更新角色的菜单权限
func (h *RoleHandler) UpdateRoleMenus(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req struct {
		MenuIDs []int64 `json:"menu_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.UpdateRoleMenus(c.Request.Context(), roleID, req.MenuIDs); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.OK(c, gin.H{"message": "权限更新成功"})
}
