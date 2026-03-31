package service

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/zzhtl/go-mountain/internal/model"
	"github.com/zzhtl/go-mountain/internal/pkg/errcode"
	"github.com/zzhtl/go-mountain/internal/repository"
)

// PaymentService 支付服务
type PaymentService struct {
	repo *repository.BaseRepo[model.Payment]
	db   *gorm.DB
}

// NewPaymentService 创建支付服务
func NewPaymentService(db *gorm.DB) *PaymentService {
	return &PaymentService{
		repo: repository.NewBaseRepo[model.Payment](db),
		db:   db,
	}
}

// GetDB 返回数据库连接（供 handler 创建关联 service 使用）
func (s *PaymentService) GetDB() *gorm.DB {
	return s.db
}

// PaymentListItem 支付列表项
type PaymentListItem struct {
	model.Payment
	UserName string `json:"user_name"`
}

// List 获取支付记录列表（后台管理）
func (s *PaymentService) List(ctx context.Context, page, pageSize int, status int, bizType string) ([]PaymentListItem, int64, error) {
	var (
		list  []PaymentListItem
		total int64
	)

	db := s.db.WithContext(ctx).Table("payments").
		Select("payments.*").
		Where("payments.deleted_at IS NULL")

	if status >= 0 {
		db = db.Where("payments.status = ?", status)
	}
	if bizType != "" {
		db = db.Where("payments.biz_type = ?", bizType)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Order("payments.created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// Get 获取支付详情
func (s *PaymentService) Get(ctx context.Context, id int64) (*model.Payment, error) {
	return s.repo.GetByID(ctx, id)
}

// GenerateOrderNo 生成订单号
func (s *PaymentService) GenerateOrderNo() string {
	return fmt.Sprintf("P%s%06d", time.Now().Format("20060102150405"), time.Now().UnixNano()%1000000)
}

// CreatePrepayOrder 创建预支付订单
// 调用微信 JSAPI 下单接口，返回前端拉起支付所需参数
func (s *PaymentService) CreatePrepayOrder(ctx context.Context, userID int64, amount float64, bizType string, bizID int64, openID string) (*model.Payment, map[string]string, error) {
	orderNo := s.GenerateOrderNo()

	// 创建支付记录
	payment := &model.Payment{
		OrderNo: orderNo,
		UserID:  userID,
		Amount:  amount,
		PayType: "wechat_jsapi",
		Status:  0,
		BizType: bizType,
		BizID:   bizID,
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, nil, err
	}

	// TODO: 调用微信 JSAPI 统一下单接口
	// 1. 构建请求参数（appid, mchid, description, out_trade_no, amount, payer.openid 等）
	// 2. 签名请求
	// 3. 发送 POST 到 https://api.mch.weixin.qq.com/v3/pay/transactions/jsapi
	// 4. 获取 prepay_id
	// 5. 生成前端支付参数（timeStamp, nonceStr, package, signType, paySign）

	// 临时返回模拟参数，实际接入时替换
	payParams := map[string]string{
		"order_no": orderNo,
		"message":  "微信支付接口尚未接入，请先配置微信支付证书",
	}

	return payment, payParams, nil
}

// HandleNotify 处理微信支付回调
func (s *PaymentService) HandleNotify(ctx context.Context, orderNo string, transactionID string, notifyData []byte) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 查询支付记录
		var payment model.Payment
		if err := tx.Where("order_no = ?", orderNo).First(&payment).Error; err != nil {
			return errcode.ErrPaymentNotFound
		}

		// 幂等处理
		if payment.Status == 1 {
			return nil
		}

		now := time.Now()
		// 更新支付记录
		if err := tx.Model(&payment).Updates(map[string]any{
			"transaction_id": transactionID,
			"status":         1,
			"paid_at":        &now,
			"notify_data":    notifyData,
		}).Error; err != nil {
			return err
		}

		// 更新关联业务状态
		if payment.BizType == "registration" {
			if err := tx.Model(&model.Registration{}).
				Where("id = ?", payment.BizID).
				Updates(map[string]any{
					"status":     1,
					"payment_id": payment.ID,
				}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// RefundOrder 退款
func (s *PaymentService) RefundOrder(ctx context.Context, paymentID int64) error {
	var payment model.Payment
	if err := s.db.WithContext(ctx).First(&payment, paymentID).Error; err != nil {
		return errcode.ErrPaymentNotFound
	}

	if payment.Status != 1 {
		return errcode.ErrRefundFailed
	}

	// TODO: 调用微信退款接口
	// POST https://api.mch.weixin.qq.com/v3/refund/domestic/refunds

	now := time.Now()
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新支付状态
		if err := tx.Model(&payment).Updates(map[string]any{
			"status":    2,
			"refund_at": &now,
		}).Error; err != nil {
			return err
		}

		// 更新关联业务状态
		if payment.BizType == "registration" {
			if err := tx.Model(&model.Registration{}).
				Where("id = ?", payment.BizID).
				Update("status", 3).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetByOrderNo 根据订单号查询支付记录
func (s *PaymentService) GetByOrderNo(ctx context.Context, orderNo string) (*model.Payment, error) {
	var payment model.Payment
	if err := s.db.WithContext(ctx).Where("order_no = ?", orderNo).First(&payment).Error; err != nil {
		return nil, errcode.ErrPaymentNotFound
	}
	return &payment, nil
}
