package model

// Role 角色
type Role struct {
	BaseModel
	Name        string `gorm:"type:text;uniqueIndex;not null" json:"name"`
	DisplayName string `gorm:"type:text;not null" json:"display_name"`
	Description string `gorm:"type:text" json:"description"`
	Status      int    `gorm:"default:1" json:"status"`

	// 关联
	Menus []Menu `gorm:"many2many:role_menus;" json:"menus,omitempty"`
}

func (Role) TableName() string {
	return "roles"
}

// RoleMenu 角色菜单关联表
type RoleMenu struct {
	RoleID int64 `gorm:"primaryKey" json:"role_id"`
	MenuID int64 `gorm:"primaryKey" json:"menu_id"`
}

func (RoleMenu) TableName() string {
	return "role_menus"
}
