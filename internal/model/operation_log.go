package model

import (
	"encoding/json"
	"time"
)

// OperationLog 操作日志
type OperationLog struct {
	ID         int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64           `gorm:"index" json:"user_id"`
	Username   string          `gorm:"type:text" json:"username"`
	Module     string          `gorm:"type:text" json:"module"`
	Action     string          `gorm:"type:text" json:"action"` // create/update/delete/login/export
	TargetType string          `gorm:"type:text" json:"target_type"`
	TargetID   int64           `json:"target_id"`
	IP         string          `gorm:"type:text" json:"ip"`
	UserAgent  string          `gorm:"type:text" json:"user_agent"`
	Detail     json.RawMessage `gorm:"type:jsonb" json:"detail,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

func (OperationLog) TableName() string {
	return "operation_logs"
}
