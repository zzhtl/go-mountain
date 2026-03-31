package model

import (
	"encoding/json"
	"time"
)

// Payment 支付记录
type Payment struct {
	BaseModel
	OrderNo       string          `gorm:"type:text;uniqueIndex;not null" json:"order_no"`
	TransactionID string          `gorm:"type:text" json:"transaction_id"`
	UserID        int64           `gorm:"not null;index" json:"user_id"`
	Amount        float64         `gorm:"type:decimal(10,2);not null" json:"amount"`
	PayType       string          `gorm:"type:text;not null" json:"pay_type"` // wechat_jsapi
	Status        int             `gorm:"default:0" json:"status"`                   // 0:待支付 1:已支付 2:已退款 3:支付失败
	BizType       string          `gorm:"type:text;not null" json:"biz_type"` // registration/donation
	BizID         int64           `gorm:"not null" json:"biz_id"`
	PrepayID      string          `gorm:"type:text" json:"prepay_id"`
	PaidAt        *time.Time      `json:"paid_at,omitempty"`
	RefundAt      *time.Time      `json:"refund_at,omitempty"`
	NotifyData    json.RawMessage `gorm:"type:jsonb" json:"notify_data,omitempty"`

	// 关联
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Payment) TableName() string {
	return "payments"
}
