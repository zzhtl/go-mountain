package model

// SystemConfig 系统配置
type SystemConfig struct {
	BaseModel
	Key       string `gorm:"type:text;uniqueIndex;not null" json:"key"`
	Value     string `gorm:"type:text;not null" json:"value"`
	Type      string `gorm:"type:text;default:string" json:"type"` // string/number/json/boolean
	GroupName string `gorm:"type:text" json:"group_name"`
	Remark    string `gorm:"type:text" json:"remark"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}
