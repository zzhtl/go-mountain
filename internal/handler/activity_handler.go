package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/zzhtl/go-mountain/internal/model"
	"github.com/zzhtl/go-mountain/internal/pkg/response"
	"github.com/zzhtl/go-mountain/internal/service"
)

// ActivityHandler 活动处理器
type ActivityHandler struct {
	svc *service.ActivityService
}

// NewActivityHandler 创建活动处理器
func NewActivityHandler(svc *service.ActivityService) *ActivityHandler {
	return &ActivityHandler{svc: svc}
}

type activityRequest struct {
	Title           string     `json:"title" binding:"required"`
	Description     string     `json:"description"`
	Content         string     `json:"content"`
	Thumbnail       string     `json:"thumbnail"`
	Location        string     `json:"location"`
	StartTime       time.Time  `json:"start_time" binding:"required"`
	EndTime         time.Time  `json:"end_time" binding:"required"`
	RegStartTime    *time.Time `json:"reg_start_time"`
	RegEndTime      *time.Time `json:"reg_end_time"`
	MaxParticipants int        `json:"max_participants"`
	Price           float64    `json:"price"`
	Status          int        `json:"status"`
}

// List 获取活动列表（后台）
func (h *ActivityHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	list, total, err := h.svc.List(c.Request.Context(), page, pageSize, status)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.PageOK(c, list, total, page, pageSize)
}

// Get 获取活动详情
func (h *ActivityHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	item, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "活动不存在")
		return
	}

	response.OK(c, item)
}

// Create 创建活动
func (h *ActivityHandler) Create(c *gin.Context) {
	var req activityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")

	activity := &model.Activity{
		Title:           req.Title,
		Description:     req.Description,
		Content:         req.Content,
		Thumbnail:       req.Thumbnail,
		Location:        req.Location,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		RegStartTime:    req.RegStartTime,
		RegEndTime:      req.RegEndTime,
		MaxParticipants: req.MaxParticipants,
		Price:           req.Price,
		Status:          req.Status,
		CreatedBy:       int64(userID.(float64)),
	}

	if err := h.svc.Create(c.Request.Context(), activity); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Created(c, activity)
}

// Update 更新活动
func (h *ActivityHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req activityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updates := map[string]any{
		"title":            req.Title,
		"description":      req.Description,
		"content":          req.Content,
		"thumbnail":        req.Thumbnail,
		"location":         req.Location,
		"start_time":       req.StartTime,
		"end_time":         req.EndTime,
		"reg_start_time":   req.RegStartTime,
		"reg_end_time":     req.RegEndTime,
		"max_participants": req.MaxParticipants,
		"price":            req.Price,
		"status":           req.Status,
	}

	if err := h.svc.Update(c.Request.Context(), id, updates); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.OK(c, gin.H{"id": id})
}

// UpdateStatus 更新活动状态
func (h *ActivityHandler) UpdateStatus(c *gin.Context) {
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

// Delete 删除活动
func (h *ActivityHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.NoContent(c)
}

// ListForMP 获取活动列表（小程序）
func (h *ActivityHandler) ListForMP(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.svc.ListForMP(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.PageOK(c, list, total, page, pageSize)
}

// GetForMP 获取活动详情（小程序）
func (h *ActivityHandler) GetForMP(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	item, err := h.svc.GetForMP(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "活动不存在")
		return
	}

	response.OK(c, item)
}
