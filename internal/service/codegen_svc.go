package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
	"unicode"

	"gorm.io/gorm"

	"github.com/zzhtl/go-mountain/internal/model"
	"github.com/zzhtl/go-mountain/internal/repository"
)

// CodegenService 代码生成服务
type CodegenService struct {
	repo *repository.BaseRepo[model.CodegenConfig]
	db   *gorm.DB
}

// NewCodegenService 创建代码生成服务
func NewCodegenService(db *gorm.DB) *CodegenService {
	return &CodegenService{
		repo: repository.NewBaseRepo[model.CodegenConfig](db),
		db:   db,
	}
}

// TableInfo 数据库表信息
type TableInfo struct {
	TableName string `json:"table_name"`
	Comment   string `json:"comment"`
}

// ColumnInfo 数据库列信息
type ColumnInfo struct {
	Field      string `json:"field"`
	Type       string `json:"type"`
	Nullable   bool   `json:"nullable"`
	Default    string `json:"default"`
	Comment    string `json:"comment"`
	IsPrimary  bool   `json:"is_primary"`
	GoField    string `json:"go_field"`
	GoType     string `json:"go_type"`
	JsonTag    string `json:"json_tag"`
	FormType   string `json:"form_type"`
}

// GeneratedCode 生成的代码
type GeneratedCode struct {
	ModelCode    string `json:"model_code"`
	ServiceCode  string `json:"service_code"`
	HandlerCode  string `json:"handler_code"`
	RouterCode   string `json:"router_code"`
	APICode      string `json:"api_code"`
}

// 需要排除的系统表
var systemTables = map[string]bool{
	"schema_migrations": true,
}

// GetTables 获取所有用户表
func (s *CodegenService) GetTables(ctx context.Context) ([]TableInfo, error) {
	var tables []TableInfo
	err := s.db.WithContext(ctx).Raw(`
		SELECT c.relname AS table_name,
		       COALESCE(pg_catalog.obj_description(c.oid, 'pg_class'), '') AS comment
		FROM pg_catalog.pg_class c
		JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relkind = 'r'
		  AND n.nspname = 'public'
		ORDER BY c.relname
	`).Scan(&tables).Error
	if err != nil {
		return nil, err
	}

	// 过滤系统表
	filtered := make([]TableInfo, 0, len(tables))
	for _, t := range tables {
		if !systemTables[t.TableName] {
			filtered = append(filtered, t)
		}
	}
	return filtered, nil
}

// GetTableColumns 获取表的列信息
func (s *CodegenService) GetTableColumns(ctx context.Context, tableName string) ([]ColumnInfo, error) {
	var columns []ColumnInfo
	err := s.db.WithContext(ctx).Raw(`
		SELECT
			c.column_name AS field,
			c.data_type AS type,
			(c.is_nullable = 'YES') AS nullable,
			COALESCE(c.column_default, '') AS "default",
			COALESCE(pgd.description, '') AS comment,
			COALESCE(tc.constraint_type = 'PRIMARY KEY', false) AS is_primary
		FROM information_schema.columns c
		LEFT JOIN pg_catalog.pg_statio_all_tables st
			ON st.schemaname = c.table_schema AND st.relname = c.table_name
		LEFT JOIN pg_catalog.pg_description pgd
			ON pgd.objoid = st.relid AND pgd.objsubid = c.ordinal_position
		LEFT JOIN information_schema.key_column_usage kcu
			ON kcu.table_schema = c.table_schema
			AND kcu.table_name = c.table_name
			AND kcu.column_name = c.column_name
		LEFT JOIN information_schema.table_constraints tc
			ON tc.constraint_name = kcu.constraint_name
			AND tc.constraint_type = 'PRIMARY KEY'
		WHERE c.table_schema = 'public' AND c.table_name = ?
		ORDER BY c.ordinal_position
	`, tableName).Scan(&columns).Error
	if err != nil {
		return nil, err
	}

	// 跳过 BaseModel 管理的字段，自动填充 Go 类型和表单类型
	baseFields := map[string]bool{"id": true, "created_at": true, "updated_at": true, "deleted_at": true}
	result := make([]ColumnInfo, 0, len(columns))
	for _, col := range columns {
		if baseFields[col.Field] {
			continue
		}
		col.GoField = snakeToPascal(col.Field)
		col.GoType = pgTypeToGoType(col.Type)
		col.JsonTag = col.Field
		col.FormType = inferFormType(col.Field, col.Type)
		result = append(result, col)
	}

	return result, nil
}

// List 获取代码生成配置列表
func (s *CodegenService) List(ctx context.Context, page, pageSize int) ([]model.CodegenConfig, int64, error) {
	scope := func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}
	return s.repo.List(ctx, page, pageSize, scope)
}

// Get 获取单个配置
func (s *CodegenService) Get(ctx context.Context, id int64) (*model.CodegenConfig, error) {
	return s.repo.GetByID(ctx, id)
}

// Create 创建配置
func (s *CodegenService) Create(ctx context.Context, config *model.CodegenConfig) error {
	return s.repo.Create(ctx, config)
}

// Update 更新配置
func (s *CodegenService) Update(ctx context.Context, id int64, updates map[string]any) error {
	return s.repo.Update(ctx, id, updates)
}

// Delete 删除配置
func (s *CodegenService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// Preview 预览生成的代码（不写文件）
func (s *CodegenService) Preview(ctx context.Context, id int64) (*GeneratedCode, error) {
	config, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("配置不存在")
	}

	return s.generateCode(config)
}

// Generate 生成代码并写入文件 + 创建菜单
func (s *CodegenService) Generate(ctx context.Context, id int64) (*GeneratedCode, error) {
	config, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("配置不存在")
	}

	code, err := s.generateCode(config)
	if err != nil {
		return nil, err
	}

	// 写入文件
	files := map[string]string{
		fmt.Sprintf("internal/model/gen_%s.go", config.ModuleName):          code.ModelCode,
		fmt.Sprintf("internal/service/gen_%s_svc.go", config.ModuleName):    code.ServiceCode,
		fmt.Sprintf("internal/handler/gen_%s_handler.go", config.ModuleName): code.HandlerCode,
	}

	for path, content := range files {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("创建目录失败 %s: %w", dir, err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("写入文件失败 %s: %w", path, err)
		}
	}

	// 创建菜单项
	if err := s.createMenus(ctx, config); err != nil {
		return nil, fmt.Errorf("创建菜单失败: %w", err)
	}

	// 更新配置状态
	now := time.Now()
	s.repo.Update(ctx, id, map[string]any{
		"generated":    true,
		"generated_at": &now,
	})

	return code, nil
}

// createMenus 自动创建菜单项
func (s *CodegenService) createMenus(ctx context.Context, config *model.CodegenConfig) error {
	parentID := config.ParentMenuID
	moduleName := config.ModuleName
	routeName := strings.ReplaceAll(moduleName, "_", "-")

	// 检查菜单是否已存在
	var count int64
	s.db.WithContext(ctx).Model(&model.Menu{}).
		Where("name = ? AND is_generated = true", "gen-"+routeName).
		Count(&count)
	if count > 0 {
		return nil // 已存在则跳过
	}

	// 如果没有指定父菜单，创建一个目录菜单
	if parentID == 0 {
		dir := model.Menu{
			ParentID:    0,
			Name:        "gen-" + routeName + "-dir",
			Title:       config.DisplayName,
			Icon:        "Grid",
			Sort:        100,
			Type:        1,
			Status:      1,
			IsGenerated: true,
		}
		if err := s.db.WithContext(ctx).Create(&dir).Error; err != nil {
			return err
		}
		parentID = dir.ID
	}

	// 创建列表菜单
	listMenu := model.Menu{
		ParentID:    parentID,
		Name:        "gen-" + routeName,
		Title:       config.DisplayName,
		Path:        "/admin/gen-" + routeName,
		Component:   "gen/" + routeName,
		Icon:        "List",
		Sort:        1,
		Type:        2,
		Status:      1,
		IsGenerated: true,
	}
	if err := s.db.WithContext(ctx).Create(&listMenu).Error; err != nil {
		return err
	}

	// 创建 CRUD 权限按钮
	perms := []struct {
		name   string
		title  string
		perm   string
		method string
	}{
		{"list", "查看列表", "gen-" + routeName + ":list", "GET"},
		{"get", "查看详情", "gen-" + routeName + ":get", "GET"},
		{"create", "创建", "gen-" + routeName + ":create", "POST"},
		{"update", "编辑", "gen-" + routeName + ":update", "PUT"},
		{"delete", "删除", "gen-" + routeName + ":delete", "DELETE"},
	}
	for _, p := range perms {
		btn := model.Menu{
			ParentID:    listMenu.ID,
			Name:        "gen-" + routeName + "-" + p.name,
			Title:       p.title,
			Permission:  p.perm,
			Method:      p.method,
			Sort:        0,
			Type:        3,
			Status:      1,
			IsGenerated: true,
		}
		s.db.WithContext(ctx).Create(&btn)
	}

	// 为 admin 角色分配新菜单
	var adminRole model.Role
	if err := s.db.WithContext(ctx).Where("name = ?", "admin").First(&adminRole).Error; err == nil {
		var allNew []model.Menu
		s.db.WithContext(ctx).Where("is_generated = true AND name LIKE ?", "gen-"+routeName+"%").Find(&allNew)
		for _, m := range allNew {
			s.db.WithContext(ctx).FirstOrCreate(&model.RoleMenu{}, model.RoleMenu{RoleID: adminRole.ID, MenuID: m.ID})
		}
		// 如果创建了目录菜单，也分配
		if config.ParentMenuID == 0 {
			s.db.WithContext(ctx).FirstOrCreate(&model.RoleMenu{}, model.RoleMenu{RoleID: adminRole.ID, MenuID: parentID})
		}
	}

	return nil
}

// generateCode 根据配置生成所有代码
func (s *CodegenService) generateCode(config *model.CodegenConfig) (*GeneratedCode, error) {
	var columns []model.ColumnConfig
	if err := json.Unmarshal(config.ColumnsConfig, &columns); err != nil {
		return nil, fmt.Errorf("解析字段配置失败: %w", err)
	}

	data := &templateData{
		TableName:    config.TblName,
		ModuleName:   config.ModuleName,
		DisplayName:  config.DisplayName,
		StructName:   snakeToPascal(config.ModuleName),
		RouteName:    strings.ReplaceAll(config.ModuleName, "_", "-"),
		Columns:      columns,
		HasSearch:     hasSearchable(columns),
		SearchColumns: getSearchable(columns),
	}

	modelCode, err := renderTemplate(modelTpl, data)
	if err != nil {
		return nil, fmt.Errorf("生成 Model 失败: %w", err)
	}

	serviceCode, err := renderTemplate(serviceTpl, data)
	if err != nil {
		return nil, fmt.Errorf("生成 Service 失败: %w", err)
	}

	handlerCode, err := renderTemplate(handlerTpl, data)
	if err != nil {
		return nil, fmt.Errorf("生成 Handler 失败: %w", err)
	}

	routerCode, err := renderTemplate(routerSnippetTpl, data)
	if err != nil {
		return nil, fmt.Errorf("生成 Router 代码片段失败: %w", err)
	}

	apiCode, err := renderTemplate(apiSnippetTpl, data)
	if err != nil {
		return nil, fmt.Errorf("生成 API 代码片段失败: %w", err)
	}

	return &GeneratedCode{
		ModelCode:   modelCode,
		ServiceCode: serviceCode,
		HandlerCode: handlerCode,
		RouterCode:  routerCode,
		APICode:     apiCode,
	}, nil
}

// ==================== 模板数据结构 ====================

type templateData struct {
	TableName     string
	ModuleName    string
	DisplayName   string
	StructName    string
	RouteName     string
	Columns       []model.ColumnConfig
	HasSearch     bool
	SearchColumns []model.ColumnConfig
}

// ==================== 辅助函数 ====================

func snakeToPascal(s string) string {
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		runes := []rune(p)
		runes[0] = unicode.ToUpper(runes[0])
		b.WriteString(string(runes))
	}
	return b.String()
}

func pgTypeToGoType(pgType string) string {
	pgType = strings.ToLower(pgType)
	switch {
	case strings.Contains(pgType, "int8"), strings.Contains(pgType, "bigint"):
		return "int64"
	case strings.Contains(pgType, "int"), strings.Contains(pgType, "serial"):
		return "int"
	case strings.Contains(pgType, "numeric"), strings.Contains(pgType, "decimal"),
		strings.Contains(pgType, "real"), strings.Contains(pgType, "double"), strings.Contains(pgType, "float"):
		return "float64"
	case strings.Contains(pgType, "bool"):
		return "bool"
	case strings.Contains(pgType, "timestamp"), strings.Contains(pgType, "date"):
		return "time.Time"
	case strings.Contains(pgType, "json"):
		return "json.RawMessage"
	default:
		return "string"
	}
}

func inferFormType(field, pgType string) string {
	pgType = strings.ToLower(pgType)
	field = strings.ToLower(field)

	switch {
	case strings.Contains(field, "content") || strings.Contains(field, "description") || strings.Contains(field, "remark"):
		return "textarea"
	case strings.Contains(field, "image") || strings.Contains(field, "thumbnail") || strings.Contains(field, "avatar") || strings.Contains(field, "cover"):
		return "image"
	case strings.Contains(field, "status") || strings.Contains(field, "is_") || strings.Contains(field, "enabled"):
		if strings.Contains(pgType, "bool") {
			return "switch"
		}
		return "select"
	case strings.Contains(pgType, "bool"):
		return "switch"
	case strings.Contains(pgType, "int"), strings.Contains(pgType, "numeric"), strings.Contains(pgType, "decimal"),
		strings.Contains(pgType, "float"), strings.Contains(pgType, "double"), strings.Contains(pgType, "real"):
		return "number"
	case strings.Contains(pgType, "timestamp"):
		return "datetime"
	case strings.Contains(pgType, "date"):
		return "date"
	case strings.Contains(pgType, "json"):
		return "textarea"
	default:
		return "input"
	}
}

func hasSearchable(columns []model.ColumnConfig) bool {
	for _, c := range columns {
		if c.Searchable {
			return true
		}
	}
	return false
}

func getSearchable(columns []model.ColumnConfig) []model.ColumnConfig {
	var result []model.ColumnConfig
	for _, c := range columns {
		if c.Searchable {
			result = append(result, c)
		}
	}
	return result
}

func renderTemplate(tplStr string, data *templateData) (string, error) {
	funcMap := template.FuncMap{
		"snakeToPascal": snakeToPascal,
		"goType": func(col model.ColumnConfig) string {
			return col.Type
		},
		"gormTag": func(col model.ColumnConfig) string {
			tag := "type:text"
			switch col.Type {
			case "int", "int64":
				tag = ""
			case "float64":
				tag = "type:decimal(10,2)"
			case "bool":
				tag = "default:false"
			case "time.Time":
				tag = ""
			case "json.RawMessage":
				tag = "type:jsonb"
			}
			if col.Required {
				if tag != "" {
					tag += ";not null"
				} else {
					tag = "not null"
				}
			}
			return tag
		},
		"isString": func(t string) bool {
			return t == "string"
		},
		"isSelect": func(ft string) bool {
			return ft == "select"
		},
		"isTextLike": func(ft string) bool {
			return ft == "textarea" || ft == "richtext"
		},
		"needsTimeImport": func(cols []model.ColumnConfig) bool {
			for _, c := range cols {
				if c.Type == "time.Time" {
					return true
				}
			}
			return false
		},
		"needsJSONImport": func(cols []model.ColumnConfig) bool {
			for _, c := range cols {
				if c.Type == "json.RawMessage" {
					return true
				}
			}
			return false
		},
	}

	tpl, err := template.New("codegen").Funcs(funcMap).Parse(tplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ==================== 代码模板 ====================

var modelTpl = `package model
{{if or (needsTimeImport .Columns) (needsJSONImport .Columns)}}
import (
{{- if needsTimeImport .Columns}}
	"time"
{{- end}}
{{- if needsJSONImport .Columns}}
	"encoding/json"
{{- end}}
)
{{end}}
// {{.StructName}} {{.DisplayName}}
type {{.StructName}} struct {
	BaseModel
{{- range .Columns}}
	{{snakeToPascal .Field}} {{.Type}} ` + "`" + `gorm:"{{gormTag .}}" json:"{{.Field}}"` + "`" + `
{{- end}}
}

func ({{.StructName}}) TableName() string {
	return "{{.TableName}}"
}
`

var serviceTpl = `package service

import (
	"context"

	"gorm.io/gorm"

	"github.com/zzhtl/go-mountain/internal/model"
	"github.com/zzhtl/go-mountain/internal/repository"
)

// {{.StructName}}Service {{.DisplayName}}服务
type {{.StructName}}Service struct {
	repo *repository.BaseRepo[model.{{.StructName}}]
	db   *gorm.DB
}

// New{{.StructName}}Service 创建{{.DisplayName}}服务
func New{{.StructName}}Service(db *gorm.DB) *{{.StructName}}Service {
	return &{{.StructName}}Service{
		repo: repository.NewBaseRepo[model.{{.StructName}}](db),
		db:   db,
	}
}

// List 获取{{.DisplayName}}列表
func (s *{{.StructName}}Service) List(ctx context.Context, page, pageSize int{{if .HasSearch}}, keyword string{{end}}) ([]model.{{.StructName}}, int64, error) {
	scope := func(db *gorm.DB) *gorm.DB {
		q := db.Order("created_at DESC")
{{- if .HasSearch}}
		if keyword != "" {
			like := "%" + keyword + "%"
			q = q.Where(
{{- $first := true}}
{{- range .SearchColumns}}
{{- if $first}}
				"{{.Field}} LIKE ?"{{$first = false}}
{{- else}}
				+" OR {{.Field}} LIKE ?"
{{- end}}
{{- end}},
{{- range $i, $c := .SearchColumns}}
{{- if eq $i 0}}
				like
{{- else}}, like
{{- end}}
{{- end}},
			)
		}
{{- end}}
		return q
	}
	return s.repo.List(ctx, page, pageSize, scope)
}

// Get 获取{{.DisplayName}}详情
func (s *{{.StructName}}Service) Get(ctx context.Context, id int64) (*model.{{.StructName}}, error) {
	return s.repo.GetByID(ctx, id)
}

// Create 创建{{.DisplayName}}
func (s *{{.StructName}}Service) Create(ctx context.Context, entity *model.{{.StructName}}) error {
	return s.repo.Create(ctx, entity)
}

// Update 更新{{.DisplayName}}
func (s *{{.StructName}}Service) Update(ctx context.Context, id int64, updates map[string]any) error {
	return s.repo.Update(ctx, id, updates)
}

// Delete 删除{{.DisplayName}}
func (s *{{.StructName}}Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
`

var handlerTpl = `package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zzhtl/go-mountain/internal/model"
	"github.com/zzhtl/go-mountain/internal/pkg/response"
	"github.com/zzhtl/go-mountain/internal/service"
)

// {{.StructName}}Handler {{.DisplayName}}处理器
type {{.StructName}}Handler struct {
	svc *service.{{.StructName}}Service
}

// New{{.StructName}}Handler 创建{{.DisplayName}}处理器
func New{{.StructName}}Handler(svc *service.{{.StructName}}Service) *{{.StructName}}Handler {
	return &{{.StructName}}Handler{svc: svc}
}

// List 获取{{.DisplayName}}列表
func (h *{{.StructName}}Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
{{- if .HasSearch}}
	keyword := c.Query("keyword")
	list, total, err := h.svc.List(c.Request.Context(), page, pageSize, keyword)
{{- else}}
	list, total, err := h.svc.List(c.Request.Context(), page, pageSize)
{{- end}}
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.PageOK(c, list, total, page, pageSize)
}

// Get 获取{{.DisplayName}}详情
func (h *{{.StructName}}Handler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	item, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "{{.DisplayName}}不存在")
		return
	}
	response.OK(c, item)
}

// Create 创建{{.DisplayName}}
func (h *{{.StructName}}Handler) Create(c *gin.Context) {
	var entity model.{{.StructName}}
	if err := c.ShouldBindJSON(&entity); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.Create(c.Request.Context(), &entity); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Created(c, entity)
}

// Update 更新{{.DisplayName}}
func (h *{{.StructName}}Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	var updates map[string]any
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 排除不可更新的字段
	delete(updates, "id")
	delete(updates, "created_at")
	delete(updates, "updated_at")
	delete(updates, "deleted_at")

	if err := h.svc.Update(c.Request.Context(), id, updates); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id})
}

// Delete 删除{{.DisplayName}}
func (h *{{.StructName}}Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.NoContent(c)
}
`

var routerSnippetTpl = `// ===== 请将以下代码添加到 internal/router/router.go 的 Setup 函数中 =====

// 创建 service 和 handler（添加到 "创建 services" 区域）
gen{{.StructName}}Svc := service.New{{.StructName}}Service(db)
gen{{.StructName}}Handler := handler.New{{.StructName}}Handler(gen{{.StructName}}Svc)

// 注册路由（添加到 adminAuth 路由组中）
gen{{.RouteName}} := adminAuth.Group("/gen-{{.RouteName}}")
gen{{.RouteName}}.GET("/", gen{{.StructName}}Handler.List)
gen{{.RouteName}}.POST("/", gen{{.StructName}}Handler.Create)
gen{{.RouteName}}.GET("/:id", gen{{.StructName}}Handler.Get)
gen{{.RouteName}}.PUT("/:id", gen{{.StructName}}Handler.Update)
gen{{.RouteName}}.DELETE("/:id", gen{{.StructName}}Handler.Delete)
`

var apiSnippetTpl = `// ===== 请将以下代码添加到 frontend-admin/src/api/index.js =====

export const gen{{.StructName}}Api = {
  list: params => request.get('/api/admin/gen-{{.RouteName}}/', { params }),
  get: id => request.get(` + "`" + `/api/admin/gen-{{.RouteName}}/${id}` + "`" + `),
  create: data => request.post('/api/admin/gen-{{.RouteName}}/', data),
  update: (id, data) => request.put(` + "`" + `/api/admin/gen-{{.RouteName}}/${id}` + "`" + `, data),
  delete: id => request.delete(` + "`" + `/api/admin/gen-{{.RouteName}}/${id}` + "`" + `)
}
`
