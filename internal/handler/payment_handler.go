package handler

import (
	"io"
	"strconv"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/models"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/payment/notify/request"
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
// 使用 PowerWeChat 进行签名验证和数据解密
func (h *PaymentHandler) WechatNotify(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取 PowerWeChat 支付实例
	app, err := h.svc.GetPaymentApp(ctx)
	if err != nil {
		c.JSON(500, gin.H{"code": "FAIL", "message": "支付配置错误: " + err.Error()})
		return
	}

	// 使用 PowerWeChat 处理支付回调（自动验签 + 解密通知数据）
	resp, err := app.HandlePaidNotify(
		c.Request,
		func(message *request.RequestNotify, transaction *models.Transaction, fail func(message string)) interface{} {
			if transaction == nil {
				fail("交易数据为空")
				return nil
			}

			// 仅处理支付成功的通知
			if transaction.TradeState != "SUCCESS" {
				return true
			}

			// 序列化通知数据用于存档
			notifyData := service.MarshalTransaction(transaction)

			// 更新支付状态和关联业务状态
			if err := h.svc.HandleNotify(ctx, transaction.OutTradeNo, transaction.TransactionID, notifyData); err != nil {
				fail("处理支付通知失败: " + err.Error())
				return nil
			}

			return true
		},
	)

	if err != nil {
		c.JSON(500, gin.H{"code": "FAIL", "message": err.Error()})
		return
	}

	// 将 PowerWeChat 的响应转发给微信服务器
	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", body)
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

	// 获取活动信息
	actSvc := service.NewActivityService(h.svc.GetDB())
	activity, err := actSvc.Get(c.Request.Context(), reg.ActivityID)
	if err != nil {
		response.NotFound(c, "活动不存在")
		return
	}

	openIDStr, _ := openID.(string)
	payment, payParams, err := h.svc.CreatePrepayOrder(
		c.Request.Context(), userID, activity.Price,
		"registration", req.RegistrationID, openIDStr, activity.Title,
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
