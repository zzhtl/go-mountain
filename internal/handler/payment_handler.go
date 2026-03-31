package handler

import (
	"io"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zzhtl/go-mountain/internal/pkg/response"
	"github.com/zzhtl/go-mountain/internal/service"
)

// PaymentHandler 支付处理器
type PaymentHandler struct {
	svc *service.PaymentService
}

// NewPaymentHandler 创建支付处理器
func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

// List 获取支付记录列表（后台）
func (h *PaymentHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	bizType := c.Query("biz_type")

	list, total, err := h.svc.List(c.Request.Context(), page, pageSize, status, bizType)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.PageOK(c, list, total, page, pageSize)
}

// Get 获取支付详情（后台）
func (h *PaymentHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	payment, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "支付记录不存在")
		return
	}

	response.OK(c, payment)
}

// Refund 退款（后台操作）
func (h *PaymentHandler) Refund(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.svc.RefundOrder(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, gin.H{"id": id})
}

// WechatNotify 微信支付回调（不需要JWT认证）
func (h *PaymentHandler) WechatNotify(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(500, gin.H{"code": "FAIL", "message": "读取请求体失败"})
		return
	}

	// TODO: 实际接入时需要：
	// 1. 验证微信签名（从 Header 获取 Wechatpay-Signature 等）
	// 2. 解密通知数据（AES-256-GCM）
	// 3. 提取 out_trade_no 和 transaction_id

	// 临时处理：从请求中解析订单信息
	orderNo := c.Query("out_trade_no")
	transactionID := c.Query("transaction_id")

	if orderNo == "" {
		c.JSON(400, gin.H{"code": "FAIL", "message": "缺少订单号"})
		return
	}

	if err := h.svc.HandleNotify(c.Request.Context(), orderNo, transactionID, body); err != nil {
		c.JSON(500, gin.H{"code": "FAIL", "message": err.Error()})
		return
	}

	// 微信要求返回此格式表示成功
	c.JSON(200, gin.H{"code": "SUCCESS", "message": "成功"})
}

// CreateOrder 创建支付订单（小程序端）
func (h *PaymentHandler) CreateOrder(c *gin.Context) {
	var req struct {
		RegistrationID int64 `json:"registration_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := int64(c.GetFloat64("user_id"))
	openID, _ := c.Get("openid")

	// 查询报名记录获取金额信息
	regSvc := service.NewRegistrationService(h.svc.GetDB())
	reg, err := regSvc.Get(c.Request.Context(), req.RegistrationID)
	if err != nil {
		response.NotFound(c, "报名记录不存在")
		return
	}

	if reg.UserID != userID {
		response.Forbidden(c, "无权操作")
		return
	}

	if reg.Status != 0 {
		response.BadRequest(c, "该报名记录状态不支持支付")
		return
	}

	// 获取活动价格
	actSvc := service.NewActivityService(h.svc.GetDB())
	activity, err := actSvc.Get(c.Request.Context(), reg.ActivityID)
	if err != nil {
		response.NotFound(c, "活动不存在")
		return
	}

	openIDStr, _ := openID.(string)
	payment, payParams, err := h.svc.CreatePrepayOrder(
		c.Request.Context(), userID, activity.Price, "registration", req.RegistrationID, openIDStr,
	)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.OK(c, gin.H{
		"payment":    payment,
		"pay_params": payParams,
	})
}

// QueryOrder 查询支付状态（小程序端轮询）
func (h *PaymentHandler) QueryOrder(c *gin.Context) {
	orderNo := c.Query("order_no")
	if orderNo == "" {
		response.BadRequest(c, "缺少订单号")
		return
	}

	payment, err := h.svc.GetByOrderNo(c.Request.Context(), orderNo)
	if err != nil {
		response.NotFound(c, "支付记录不存在")
		return
	}

	response.OK(c, payment)
}
