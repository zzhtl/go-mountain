package model

import (
	"encoding/json"
	"time"
)

// CodegenConfig 代码生成配置
type CodegenConfig struct {
	BaseModel
	TblName      string          `gorm:"column:table_name;type:text;not null" json:"table_name"`
	ModuleName   string          `gorm:"type:text;not null" json:"module_name"`
	DisplayName  string          `gorm:"type:text;not null" json:"display_name"`
	ParentMenuID int64           `gorm:"default:0" json:"parent_menu_id"`
	ColumnsConfig json.RawMessage `gorm:"type:jsonb;not null" json:"columns_config"`
	Generated    bool            `gorm:"default:false" json:"generated"`
	GeneratedAt  *time.Time      `json:"generated_at,omitempty"`
}

func (CodegenConfig) TableName() string {
	return "codegen_configs"
}

// ColumnConfig 字段配置
type ColumnConfig struct {
	Field       string         `json:"field"`
	Label       string         `json:"label"`
	Type        string         `json:"type"`
	FormType    string         `json:"form_type"` // input/textarea/number/select/date/datetime/image/richtext/switch
	Options     []SelectOption `json:"options,omitempty"`
	ListVisible bool           `json:"list_visible"`
	Searchable  bool           `json:"searchable"`
	Sortable    bool           `json:"sortable"`
	Required    bool           `json:"required"`
}

// SelectOption 下拉选项
type SelectOption struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}
