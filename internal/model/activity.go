package model

import "time"

// Activity 活动
type Activity struct {
	BaseModel
	Title           string     `gorm:"type:text;not null" json:"title"`
	Description     string     `gorm:"type:text" json:"description"`
	Content         string     `gorm:"type:text" json:"content"`
	Thumbnail       string     `gorm:"type:text" json:"thumbnail"`
	Location        string     `gorm:"type:text" json:"location"`
	StartTime       time.Time  `json:"start_time"`
	EndTime         time.Time  `json:"end_time"`
	RegStartTime    *time.Time `json:"reg_start_time,omitempty"`
	RegEndTime      *time.Time `json:"reg_end_time,omitempty"`
	MaxParticipants int        `gorm:"default:0" json:"max_participants"` // 0=不限
	Price           float64    `gorm:"type:decimal(10,2);default:0" json:"price"`
	Status          int        `gorm:"default:0" json:"status"` // 0:草稿 1:报名中 2:报名截止 3:进行中 4:已结束
	CreatedBy       int64      `json:"created_by"`
}

func (Activity) TableName() string {
	return "activities"
}
