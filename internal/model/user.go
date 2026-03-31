package model

// User 小程序用户
type User struct {
	BaseModel
	OpenID  string `gorm:"type:text;uniqueIndex" json:"openid"`
	UnionID string `gorm:"type:text" json:"union_id"`
	Phone   string `gorm:"type:text" json:"phone"`
	Name    string `gorm:"type:text" json:"name"`
	Avatar  string `gorm:"type:text" json:"avatar"`
	Gender  int    `gorm:"default:0" json:"gender"`
	Status  int    `gorm:"default:1" json:"status"`
}

func (User) TableName() string {
	return "users"
}
