package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/ArtisanCloud/PowerLibs/v3/object"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/payment"
	orderRequest "github.com/ArtisanCloud/PowerWeChat/v3/src/payment/order/request"
	refundRequest "github.com/ArtisanCloud/PowerWeChat/v3/src/payment/refund/request"
	"gorm.io/gorm"

	"github.com/zzhtl/go-mountain/internal/model"
	"github.com/zzhtl/go-mountain/internal/pkg/errcode"
	"github.com/zzhtl/go-mountain/internal/repository"
)

// PaymentService 支付服务
type PaymentService struct {
	repo            *repository.BaseRepo[model.Payment]
	db              *gorm.DB
	systemConfigSvc *SystemConfigService
}

// NewPaymentService 创建支付服务
func NewPaymentService(db *gorm.DB, systemConfigSvc *SystemConfigService) *PaymentService {
	return &PaymentService{
		repo:            repository.NewBaseRepo[model.Payment](db),
		db:              db,
		systemConfigSvc: systemConfigSvc,
	}
}

// GetDB 返回数据库连接（供 handler 创建关联 service 使用）
func (s *PaymentService) GetDB() *gorm.DB {
	return s.db
}

// GetPaymentApp 根据数据库中的系统配置创建 PowerWeChat Payment 实例
func (s *PaymentService) GetPaymentApp(ctx context.Context) (*payment.Payment, error) {
	cfg := s.systemConfigSvc.GetWechatPayConfig(ctx)

	appID := cfg["wechat.app_id"]
	mchID := cfg["wechat.mch_id"]
	mchApiV3Key := cfg["wechat.mch_api_v3_key"]
	serialNo := cfg["wechat.mch_serial_no"]
	privateKey := cfg["wechat.mch_private_key"]
	notifyURL := cfg["wechat.notify_url"]

	if appID == "" || mchID == "" || mchApiV3Key == "" || privateKey == "" {
		return nil, fmt.Errorf("微信支付配置不完整，请在后台【系统配置】中完善微信支付相关配置")
	}

	// 将私钥内容写入临时文件，PowerWeChat 需要文件路径
	keyFile, err := os.CreateTemp("", "wechat_mch_key_*.pem")
	if err != nil {
		return nil, fmt.Errorf("创建临时密钥文件失败: %w", err)
	}
	if _, err := keyFile.WriteString(privateKey); err != nil {
		keyFile.Close()
		os.Remove(keyFile.Name())
		return nil, fmt.Errorf("写入密钥文件失败: %w", err)
	}
	keyFile.Close()
	// 延迟清理临时文件（给 PowerWeChat 足够时间读取）
	go func() {
		time.Sleep(10 * time.Second)
		os.Remove(keyFile.Name())
	}()

	app, err := payment.NewPayment(&payment.UserConfig{
		AppID:       appID,
		MchID:       mchID,
		MchApiV3Key: mchApiV3Key,
		KeyPath:     keyFile.Name(),
		SerialNo:    serialNo,
		NotifyURL:   notifyURL,
	})
	if err != nil {
		return nil, fmt.Errorf("初始化微信支付实例失败: %w", err)
	}

	return app, nil
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

// GenerateRefundNo 生成退款单号
func (s *PaymentService) GenerateRefundNo() string {
	return fmt.Sprintf("R%s%06d", time.Now().Format("20060102150405"), time.Now().UnixNano()%1000000)
}

// CreatePrepayOrder 创建预支付订单，调用微信 JSAPI 下单接口
// 返回前端小程序拉起支付所需的参数
func (s *PaymentService) CreatePrepayOrder(ctx context.Context, userID int64, amount float64, bizType string, bizID int64, openID string, description string) (*model.Payment, *object.StringMap, error) {
	orderNo := s.GenerateOrderNo()

	// 获取 PowerWeChat 支付实例
	app, err := s.GetPaymentApp(ctx)
	if err != nil {
		return nil, nil, err
	}

	// 金额转为分（微信支付以分为单位）
	totalCents := int(math.Round(amount * 100))

	// 调用微信 JSAPI 统一下单
	result, err := app.Order.JSAPITransaction(ctx, &orderRequest.RequestJSAPIPrepay{
		Description: description,
		OutTradeNo:  orderNo,
		TimeExpire:  time.Now().Add(30 * time.Minute).Format(time.RFC3339),
		Amount: &orderRequest.JSAPIAmount{
			Total:    totalCents,
			Currency: "CNY",
		},
		Payer: &orderRequest.JSAPIPayer{
			OpenID: openID,
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("微信下单失败: %w", err)
	}

	if result.PrepayID == "" {
		return nil, nil, fmt.Errorf("微信下单失败: 未获取到 prepay_id")
	}

	prepayID := result.PrepayID

	// 创建支付记录
	pay := &model.Payment{
		OrderNo:  orderNo,
		UserID:   userID,
		Amount:   amount,
		PayType:  "wechat_jsapi",
		Status:   0,
		BizType:  bizType,
		BizID:    bizID,
		PrepayID: prepayID,
	}
	if err := s.repo.Create(ctx, pay); err != nil {
		return nil, nil, err
	}

	// 生成前端小程序拉起支付所需的参数
	payParams, err := app.JSSDK.BridgeConfig(prepayID, false)
	if err != nil {
		return nil, nil, fmt.Errorf("生成支付参数失败: %w", err)
	}

	return pay, payParams.(*object.StringMap), nil
}

// HandleNotify 处理微信支付回调（由 handler 在验签解密后调用）
func (s *PaymentService) HandleNotify(ctx context.Context, orderNo string, transactionID string, notifyData []byte) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 查询支付记录
		var pay model.Payment
		if err := tx.Where("order_no = ?", orderNo).First(&pay).Error; err != nil {
			return errcode.ErrPaymentNotFound
		}

		// 幂等处理
		if pay.Status == 1 {
			return nil
		}

		now := time.Now()
		// 更新支付记录
		if err := tx.Model(&pay).Updates(map[string]any{
			"transaction_id": transactionID,
			"status":         1,
			"paid_at":        &now,
			"notify_data":    notifyData,
		}).Error; err != nil {
			return err
		}

		// 更新关联业务状态
		if pay.BizType == "registration" {
			if err := tx.Model(&model.Registration{}).
				Where("id = ?", pay.BizID).
				Updates(map[string]any{
					"status":     1,
					"payment_id": pay.ID,
				}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// RefundOrder 退款，调用微信退款接口
func (s *PaymentService) RefundOrder(ctx context.Context, paymentID int64) error {
	var pay model.Payment
	if err := s.db.WithContext(ctx).First(&pay, paymentID).Error; err != nil {
		return errcode.ErrPaymentNotFound
	}

	if pay.Status != 1 {
		return errcode.ErrRefundFailed
	}

	// 获取 PowerWeChat 支付实例
	app, err := s.GetPaymentApp(ctx)
	if err != nil {
		return err
	}

	// 金额转为分
	totalCents := int(math.Round(pay.Amount * 100))
	refundNo := s.GenerateRefundNo()

	// 调用微信退款接口
	_, err = app.Refund.Refund(ctx, &refundRequest.RequestRefund{
		TransactionID: pay.TransactionID,
		OutRefundNo:   refundNo,
		Reason:        "管理员操作退款",
		Amount: &refundRequest.RefundAmount{
			Refund:   totalCents,
			Total:    totalCents,
			Currency: "CNY",
		},
	})
	if err != nil {
		return fmt.Errorf("微信退款失败: %w", err)
	}

	now := time.Now()
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新支付状态
		if err := tx.Model(&pay).Updates(map[string]any{
			"status":    2,
			"refund_at": &now,
		}).Error; err != nil {
			return err
		}

		// 更新关联业务状态
		if pay.BizType == "registration" {
			if err := tx.Model(&model.Registration{}).
				Where("id = ?", pay.BizID).
				Update("status", 3).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetByOrderNo 根据订单号查询支付记录
func (s *PaymentService) GetByOrderNo(ctx context.Context, orderNo string) (*model.Payment, error) {
	var pay model.Payment
	if err := s.db.WithContext(ctx).Where("order_no = ?", orderNo).First(&pay).Error; err != nil {
		return nil, errcode.ErrPaymentNotFound
	}
	return &pay, nil
}

// MarshalTransaction 将交易通知序列化为 JSON 用于存储
func MarshalTransaction(v any) []byte {
	data, _ := json.Marshal(v)
	return data
}
