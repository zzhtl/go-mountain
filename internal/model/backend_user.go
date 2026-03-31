package model

import "time"

// BackendUser 后台管理用户
type BackendUser struct {
	BaseModel
	Username        string    `gorm:"type:text;uniqueIndex;not null" json:"username"`
	Email           string    `gorm:"type:text;uniqueIndex;not null" json:"email"`
	Password        string    `gorm:"type:text;not null" json:"-"`
	Avatar          string    `gorm:"type:text" json:"avatar"`
	RoleID          int64     `gorm:"not null;index" json:"role_id"`
	PasswordVersion int       `gorm:"default:2" json:"-"` // 1=SHA256 2=bcrypt
	Status          int       `gorm:"default:1" json:"status"`
	LastLogin       *time.Time `json:"last_login,omitempty"`

	// 关联
	Role *Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

func (BackendUser) TableName() string {
	return "backend_users"
}
