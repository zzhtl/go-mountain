package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zzhtl/go-mountain/internal/pkg/response"
	"github.com/zzhtl/go-mountain/internal/service"
)

// BackendUserHandler 后台用户管理处理器
type BackendUserHandler struct {
	svc *service.BackendUserService
}

// NewBackendUserHandler 创建后台用户管理处理器
func NewBackendUserHandler(svc *service.BackendUserService) *BackendUserHandler {
	return &BackendUserHandler{svc: svc}
}

// List 获取后台用户列表
func (h *BackendUserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := h.svc.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.PageOK(c, list, total, page, pageSize)
}

// Get 获取单个后台用户
func (h *BackendUserHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	item, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	response.OK(c, item)
}

// Create 创建后台用户
func (h *BackendUserHandler) Create(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		RoleID   int64  `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, password, err := h.svc.Create(c.Request.Context(), req.Username, req.Email, req.RoleID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Created(c, gin.H{
		"user":     user,
		"password": password,
	})
}

// Update 更新后台用户信息
func (h *BackendUserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		RoleID   int64  `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.Update(c.Request.Context(), id, req.Username, req.Email, req.RoleID); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.OK(c, gin.H{"id": id})
}

// UpdateStatus 更新用户状态
func (h *BackendUserHandler) UpdateStatus(c *gin.Context) {
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

// ResetPassword 重置用户密码
func (h *BackendUserHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	password, err := h.svc.ResetPassword(c.Request.Context(), id)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.OK(c, gin.H{
		"id":       id,
		"password": password,
	})
}

// Delete 删除后台用户
func (h *BackendUserHandler) Delete(c *gin.Context) {
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

// GetCurrentUserMenus 获取当前用户菜单权限
func (h *BackendUserHandler) GetCurrentUserMenus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未授权")
		return
	}

	uid := int64(userID.(float64))
	menus, err := h.svc.GetCurrentUserMenus(c.Request.Context(), uid)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.OK(c, menus)
}
