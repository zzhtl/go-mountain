package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/zzhtl/go-mountain/internal/model"
	"github.com/zzhtl/go-mountain/internal/pkg/response"
	"github.com/zzhtl/go-mountain/internal/service"
)

// SystemConfigHandler 系统配置处理器
type SystemConfigHandler struct {
	svc *service.SystemConfigService
}

// NewSystemConfigHandler 创建系统配置处理器
func NewSystemConfigHandler(svc *service.SystemConfigService) *SystemConfigHandler {
	return &SystemConfigHandler{svc: svc}
}

// List 获取所有配置
func (h *SystemConfigHandler) List(c *gin.Context) {
	group := c.Query("group_name")

	var (
		configs []model.SystemConfig
		err     error
	)

	if group != "" {
		configs, err = h.svc.ListByGroup(c.Request.Context(), group)
	} else {
		configs, err = h.svc.List(c.Request.Context())
	}

	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.OK(c, configs)
}

// GetGroups 获取所有分组
func (h *SystemConfigHandler) GetGroups(c *gin.Context) {
	groups, err := h.svc.GetGroups(c.Request.Context())
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, groups)
}

// Save 保存单个配置
func (h *SystemConfigHandler) Save(c *gin.Context) {
	var req struct {
		Key       string `json:"key" binding:"required"`
		Value     string `json:"value"`
		Type      string `json:"type"`
		GroupName string `json:"group_name"`
		Remark    string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Type == "" {
		req.Type = "string"
	}

	if err := h.svc.SetValue(c.Request.Context(), req.Key, req.Value, req.Type, req.GroupName, req.Remark); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.OK(c, gin.H{"key": req.Key})
}

// BatchSave 批量保存配置
func (h *SystemConfigHandler) BatchSave(c *gin.Context) {
	var req struct {
		Configs []model.SystemConfig `json:"configs" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.BatchSet(c.Request.Context(), req.Configs); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.OK(c, gin.H{"count": len(req.Configs)})
}

// Delete 删除配置项
func (h *SystemConfigHandler) Delete(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		response.BadRequest(c, "缺少 key 参数")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), key); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.NoContent(c)
}
