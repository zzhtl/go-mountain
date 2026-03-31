package model

// Menu 菜单
type Menu struct {
	BaseModel
	ParentID    int64  `gorm:"default:0;index" json:"parent_id"`
	Name        string `gorm:"type:text;not null" json:"name"`
	Title       string `gorm:"type:text;not null" json:"title"`
	Path        string `gorm:"type:text" json:"path"`
	Component   string `gorm:"type:text" json:"component"`
	Icon        string `gorm:"type:text" json:"icon"`
	Sort        int    `gorm:"default:0" json:"sort"`
	Type        int    `gorm:"default:1" json:"type"`                // 1:目录 2:菜单 3:按钮/API
	Permission  string `gorm:"type:text" json:"permission"`  // API权限标识 如 article:list
	Method      string `gorm:"type:text" json:"method"`       // GET/POST/PUT/DELETE
	Status      int    `gorm:"default:1" json:"status"`
	IsGenerated bool   `gorm:"default:false" json:"is_generated"`    // 是否代码生成器创建

	// 非数据库字段
	Children []*Menu `gorm:"-" json:"children,omitempty"`
}

func (Menu) TableName() string {
	return "menus"
}
