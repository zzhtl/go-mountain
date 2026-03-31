package model

// Article 文章
type Article struct {
	BaseModel
	ColumnID  int64  `gorm:"not null;index" json:"column_id"`
	Title     string `gorm:"type:text;not null" json:"title"`
	Thumbnail string `gorm:"type:text" json:"thumbnail"`
	Content   string `gorm:"type:text" json:"content"`
	Author    string `gorm:"type:text" json:"author"`
	Status    int    `gorm:"default:0" json:"status"` // 0:草稿 1:已发布 2:下架
	ViewCount int    `gorm:"default:0" json:"view_count"`

	// 关联
	Column *Column `gorm:"foreignKey:ColumnID" json:"column,omitempty"`
}

func (Article) TableName() string {
	return "articles"
}
