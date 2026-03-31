package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zzhtl/go-mountain/internal/pkg/response"
	"github.com/zzhtl/go-mountain/internal/service"
)

// UserHandler 小程序用户处理器
type UserHandler struct {
	svc *service.UserService
}

// NewUserHandler 创建小程序用户处理器
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// WechatLogin 微信小程序登录
func (h *UserHandler) WechatLogin(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.svc.WechatLogin(c.Request.Context(), req.Code)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.OK(c, user)
}

// Register 注册/绑定手机号
func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Phone  string `json:"phone"`
		OpenID string `json:"openid"`
		Name   string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.svc.Register(c.Request.Context(), req.Phone, req.OpenID, req.Name)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Created(c, user)
}

// List 获取小程序用户列表（后台）
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := h.svc.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.PageOK(c, list, total, page, pageSize)
}

// Get 获取用户详情（后台）
func (h *UserHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	user, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	response.OK(c, user)
}

// Update 更新用户信息（后台）
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req struct {
		Phone  string `json:"phone"`
		OpenID string `json:"openid"`
		Name   string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updates := map[string]any{
		"phone":   req.Phone,
		"open_id": req.OpenID,
		"name":    req.Name,
	}

	if err := h.svc.Update(c.Request.Context(), id, updates); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.OK(c, gin.H{"id": id})
}

// Delete 删除用户（后台）
func (h *UserHandler) Delete(c *gin.Context) {
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
