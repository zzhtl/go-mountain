package model

import "encoding/json"

// Registration 报名记录
type Registration struct {
	BaseModel
	ActivityID int64           `gorm:"not null;index" json:"activity_id"`
	UserID     int64           `gorm:"not null;index" json:"user_id"`
	Name       string          `gorm:"type:text;not null" json:"name"`
	Phone      string          `gorm:"type:text;not null" json:"phone"`
	IDCard     string          `gorm:"type:text" json:"id_card"`
	ExtraInfo  json.RawMessage `gorm:"type:jsonb" json:"extra_info,omitempty"`
	Status     int             `gorm:"default:0" json:"status"` // 0:待支付 1:已支付 2:已取消 3:已退款
	PaymentID  *int64          `json:"payment_id,omitempty"`

	// 关联
	Activity *Activity `gorm:"foreignKey:ActivityID" json:"activity,omitempty"`
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Registration) TableName() string {
	return "registrations"
}
