package role

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// Role 角色模型
type Role struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	DisplayName string    `db:"display_name" json:"display_name"`
	Description string    `db:"description" json:"description"`
	Status      int       `db:"status" json:"status"` // 0:禁用 1:启用
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// RoleMenu 角色菜单关联
type RoleMenu struct {
	RoleID int64 `db:"role_id" json:"role_id"`
	MenuID int64 `db:"menu_id" json:"menu_id"`
}

// Handler 角色管理处理器
type Handler struct {
	db *sqlx.DB
}

// NewHandler 创建角色处理器
func NewHandler(db *sqlx.DB) *Handler {
	h := &Handler{db: db}
	h.createTables()
	return h
}

// createTables 创建相关表
func (h *Handler) createTables() {
	// 创建角色表
	roleTable := `CREATE TABLE IF NOT EXISTS roles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		display_name TEXT NOT NULL,
		description TEXT,
		status INTEGER DEFAULT 1,
		created_at DATETIME,
		updated_at DATETIME
	)`
	h.db.Exec(h.db.Rebind(roleTable))

	// 创建角色菜单关联表
	roleMenuTable := `CREATE TABLE IF NOT EXISTS role_menus (
		role_id INTEGER NOT NULL,
		menu_id INTEGER NOT NULL,
		PRIMARY KEY (role_id, menu_id)
	)`
	h.db.Exec(h.db.Rebind(roleMenuTable))

	// 插入默认角色
	h.insertDefaultRoles()
}

// insertDefaultRoles 插入默认角色
func (h *Handler) insertDefaultRoles() {
	var count int
	h.db.Get(&count, h.db.Rebind("SELECT COUNT(*) FROM roles"))
	if count > 0 {
		return
	}

	now := time.Now()
	defaultRoles := []Role{
		{Name: "admin", DisplayName: "超级管理员", Description: "拥有所有权限", Status: 1, CreatedAt: now, UpdatedAt: now},
		{Name: "editor", DisplayName: "编辑员", Description: "文章编辑权限", Status: 1, CreatedAt: now, UpdatedAt: now},
		{Name: "viewer", DisplayName: "查看员", Description: "只读权限", Status: 1, CreatedAt: now, UpdatedAt: now},
	}

	for _, role := range defaultRoles {
		h.db.Exec(h.db.Rebind("INSERT INTO roles (name, display_name, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)"),
			role.Name, role.DisplayName, role.Description, role.Status, role.CreatedAt, role.UpdatedAt)
	}
}

// RegisterAdminRoutes 注册角色管理路由
func (h *Handler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.GET("/", h.List)
	rg.POST("/", h.Create)
	rg.GET("/:id", h.Get)
	rg.PUT("/:id", h.Update)
	rg.DELETE("/:id", h.Delete)
	rg.PUT("/:id/status", h.UpdateStatus)
	rg.GET("/:id/menus", h.GetRoleMenus)
	rg.PUT("/:id/menus", h.UpdateRoleMenus)
}

// List 获取角色列表
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	// 查询总数
	var total int
	if err := h.db.Get(&total, h.db.Rebind("SELECT COUNT(*) FROM roles")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查询列表
	var roles []Role
	query := h.db.Rebind("SELECT id, name, display_name, description, status, created_at, updated_at FROM roles ORDER BY id LIMIT ? OFFSET ?")
	if err := h.db.Select(&roles, query, pageSize, offset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":      roles,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Create 创建角色
func (h *Handler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		DisplayName string `json:"display_name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	query := h.db.Rebind("INSERT INTO roles (name, display_name, description, status, created_at, updated_at) VALUES (?, ?, ?, 1, ?, ?)")
	result, err := h.db.Exec(query, req.Name, req.DisplayName, req.Description, now, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	role := Role{
		ID:          id,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Status:      1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	c.JSON(http.StatusCreated, role)
}

// Get 获取单个角色
func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var role Role
	query := h.db.Rebind("SELECT id, name, display_name, description, status, created_at, updated_at FROM roles WHERE id = ?")
	if err := h.db.Get(&role, query, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "角色不存在"})
		return
	}
	c.JSON(http.StatusOK, role)
}

// Update 更新角色
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		DisplayName string `json:"display_name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	query := h.db.Rebind("UPDATE roles SET name = ?, display_name = ?, description = ?, updated_at = ? WHERE id = ?")
	if _, err := h.db.Exec(query, req.Name, req.DisplayName, req.Description, now, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// UpdateStatus 更新角色状态
func (h *Handler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Status int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Status != 0 && req.Status != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "状态只能是0或1"})
		return
	}

	now := time.Now()
	query := h.db.Rebind("UPDATE roles SET status = ?, updated_at = ? WHERE id = ?")
	if _, err := h.db.Exec(query, req.Status, now, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// Delete 删除角色
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// 检查是否有用户使用此角色
	var count int
	if err := h.db.Get(&count, h.db.Rebind("SELECT COUNT(*) FROM backend_users WHERE role_id = ?"), id); err == nil && count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该角色正在被用户使用，无法删除"})
		return
	}

	// 删除角色菜单关联
	h.db.Exec(h.db.Rebind("DELETE FROM role_menus WHERE role_id = ?"), id)

	// 删除角色
	query := h.db.Rebind("DELETE FROM roles WHERE id = ?")
	if _, err := h.db.Exec(query, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// GetRoleMenus 获取角色的菜单权限
func (h *Handler) GetRoleMenus(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var menuIDs []int64
	query := h.db.Rebind("SELECT menu_id FROM role_menus WHERE role_id = ?")
	if err := h.db.Select(&menuIDs, query, roleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"menu_ids": menuIDs})
}

// UpdateRoleMenus 更新角色的菜单权限
func (h *Handler) UpdateRoleMenus(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		MenuIDs []int64 `json:"menu_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 删除现有关联
	h.db.Exec(h.db.Rebind("DELETE FROM role_menus WHERE role_id = ?"), roleID)

	// 添加新关联
	for _, menuID := range req.MenuIDs {
		h.db.Exec(h.db.Rebind("INSERT INTO role_menus (role_id, menu_id) VALUES (?, ?)"), roleID, menuID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "权限更新成功"})
}
