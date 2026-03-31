package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zzhtl/go-mountain/internal/pkg/response"
	"github.com/zzhtl/go-mountain/internal/service"
)

// RegistrationHandler 报名处理器
type RegistrationHandler struct {
	svc *service.RegistrationService
}

// NewRegistrationHandler 创建报名处理器
func NewRegistrationHandler(svc *service.RegistrationService) *RegistrationHandler {
	return &RegistrationHandler{svc: svc}
}

// List 获取报名列表（后台）
func (h *RegistrationHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	activityID, _ := strconv.ParseInt(c.Query("activity_id"), 10, 64)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	list, total, err := h.svc.List(c.Request.Context(), page, pageSize, activityID, status)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.PageOK(c, list, total, page, pageSize)
}

// Get 获取报名详情（后台）
func (h *RegistrationHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	item, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "报名记录不存在")
		return
	}

	response.OK(c, item)
}

// Create 创建报名（小程序端）
func (h *RegistrationHandler) Create(c *gin.Context) {
	var req service.CreateRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := int64(c.GetFloat64("user_id"))

	reg, err := h.svc.Create(c.Request.Context(), userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, reg)
}

// Cancel 取消报名（小程序端）
func (h *RegistrationHandler) Cancel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	userID := int64(c.GetFloat64("user_id"))

	if err := h.svc.Cancel(c.Request.Context(), id, userID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, gin.H{"id": id})
}

// MyRegistrations 获取我的报名列表（小程序端）
func (h *RegistrationHandler) MyRegistrations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	userID := int64(c.GetFloat64("user_id"))

	list, total, err := h.svc.GetByUser(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.PageOK(c, list, total, page, pageSize)
}
