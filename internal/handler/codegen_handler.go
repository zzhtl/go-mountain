package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zzhtl/go-mountain/internal/model"
	"github.com/zzhtl/go-mountain/internal/pkg/response"
	"github.com/zzhtl/go-mountain/internal/service"
)

// CodegenHandler 代码生成处理器
type CodegenHandler struct {
	svc *service.CodegenService
}

// NewCodegenHandler 创建代码生成处理器
func NewCodegenHandler(svc *service.CodegenService) *CodegenHandler {
	return &CodegenHandler{svc: svc}
}

// GetTables 获取数据库所有表
func (h *CodegenHandler) GetTables(c *gin.Context) {
	tables, err := h.svc.GetTables(c.Request.Context())
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, tables)
}

// GetTableColumns 获取指定表的列信息
func (h *CodegenHandler) GetTableColumns(c *gin.Context) {
	tableName := c.Query("table_name")
	if tableName == "" {
		response.BadRequest(c, "缺少 table_name 参数")
		return
	}

	columns, err := h.svc.GetTableColumns(c.Request.Context(), tableName)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, columns)
}

// List 获取代码生成配置列表
func (h *CodegenHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := h.svc.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.PageOK(c, list, total, page, pageSize)
}

// Get 获取单个配置
func (h *CodegenHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	config, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "配置不存在")
		return
	}
	response.OK(c, config)
}

// Create 创建配置
func (h *CodegenHandler) Create(c *gin.Context) {
	var req struct {
		TableName    string              `json:"table_name" binding:"required"`
		ModuleName   string              `json:"module_name" binding:"required"`
		DisplayName  string              `json:"display_name" binding:"required"`
		ParentMenuID int64               `json:"parent_menu_id"`
		Columns      []model.ColumnConfig `json:"columns" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	columnsJSON, _ := json.Marshal(req.Columns)
	config := &model.CodegenConfig{
		TblName:       req.TableName,
		ModuleName:    req.ModuleName,
		DisplayName:   req.DisplayName,
		ParentMenuID:  req.ParentMenuID,
		ColumnsConfig: columnsJSON,
	}

	if err := h.svc.Create(c.Request.Context(), config); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Created(c, config)
}

// Update 更新配置
func (h *CodegenHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req struct {
		TableName    string              `json:"table_name"`
		ModuleName   string              `json:"module_name"`
		DisplayName  string              `json:"display_name"`
		ParentMenuID int64               `json:"parent_menu_id"`
		Columns      []model.ColumnConfig `json:"columns"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updates := map[string]any{}
	if req.TableName != "" {
		updates["table_name"] = req.TableName
	}
	if req.ModuleName != "" {
		updates["module_name"] = req.ModuleName
	}
	if req.DisplayName != "" {
		updates["display_name"] = req.DisplayName
	}
	updates["parent_menu_id"] = req.ParentMenuID
	if req.Columns != nil {
		columnsJSON, _ := json.Marshal(req.Columns)
		updates["columns_config"] = columnsJSON
	}

	if err := h.svc.Update(c.Request.Context(), id, updates); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id})
}

// Delete 删除配置
func (h *CodegenHandler) Delete(c *gin.Context) {
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

// Preview 预览生成代码
func (h *CodegenHandler) Preview(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	code, err := h.svc.Preview(c.Request.Context(), id)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, code)
}

// Generate 执行代码生成
func (h *CodegenHandler) Generate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	code, err := h.svc.Generate(c.Request.Context(), id)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, code)
}
