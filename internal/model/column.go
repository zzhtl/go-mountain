package model

// Column 栏目/分类
type Column struct {
	BaseModel
	Name        string `gorm:"type:text;not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	SortOrder   int    `gorm:"default:0" json:"sort_order"`
}

func (Column) TableName() string {
	return "columns"
}
