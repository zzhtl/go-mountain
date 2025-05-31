package menu

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// Menu 菜单模型
type Menu struct {
	ID        int64     `db:"id" json:"id"`
	ParentID  int64     `db:"parent_id" json:"parent_id"`
	Name      string    `db:"name" json:"name"`
	Title     string    `db:"title" json:"title"`
	Path      string    `db:"path" json:"path"`
	Component string    `db:"component" json:"component"`
	Icon      string    `db:"icon" json:"icon"`
	Sort      int       `db:"sort" json:"sort"`
	Type      int       `db:"type" json:"type"`     // 1:菜单 2:按钮
	Status    int       `db:"status" json:"status"` // 0:禁用 1:启用
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	Children  []*Menu   `json:"children,omitempty"`
}

// Handler 菜单管理处理器
type Handler struct {
	db *sqlx.DB
}

// NewHandler 创建菜单处理器
func NewHandler(db *sqlx.DB) *Handler {
	h := &Handler{db: db}
	h.createTables()
	return h
}

// createTables 创建菜单表
func (h *Handler) createTables() {
	menuTable := `CREATE TABLE IF NOT EXISTS menus (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		parent_id INTEGER DEFAULT 0,
		name TEXT NOT NULL,
		title TEXT NOT NULL,
		path TEXT,
		component TEXT,
		icon TEXT,
		sort INTEGER DEFAULT 0,
		type INTEGER DEFAULT 1,
		status INTEGER DEFAULT 1,
		created_at DATETIME,
		updated_at DATETIME
	)`
	h.db.Exec(h.db.Rebind(menuTable))

	// 插入默认菜单
	h.insertDefaultMenus()
}

// insertDefaultMenus 插入默认菜单
func (h *Handler) insertDefaultMenus() {
	var count int
	h.db.Get(&count, h.db.Rebind("SELECT COUNT(*) FROM menus"))
	if count > 0 {
		return
	}

	now := time.Now()
	defaultMenus := []Menu{
		{ParentID: 0, Name: "articles", Title: "文章管理", Path: "/admin/articles", Component: "ArticleList", Icon: "Document", Sort: 1, Type: 1, Status: 1, CreatedAt: now, UpdatedAt: now},
		{ParentID: 0, Name: "columns", Title: "栏目管理", Path: "/admin/columns", Component: "ColumnList", Icon: "Menu", Sort: 2, Type: 1, Status: 1, CreatedAt: now, UpdatedAt: now},
		{ParentID: 0, Name: "mp-users", Title: "小程序用户", Path: "/admin/users", Component: "UserList", Icon: "User", Sort: 3, Type: 1, Status: 1, CreatedAt: now, UpdatedAt: now},
		{ParentID: 0, Name: "backend-users", Title: "用户管理", Path: "/admin/backend-users", Component: "BackendUserList", Icon: "UserFilled", Sort: 4, Type: 1, Status: 1, CreatedAt: now, UpdatedAt: now},
		{ParentID: 0, Name: "roles", Title: "角色管理", Path: "/admin/roles", Component: "RoleList", Icon: "Key", Sort: 5, Type: 1, Status: 1, CreatedAt: now, UpdatedAt: now},
		{ParentID: 0, Name: "menus", Title: "菜单管理", Path: "/admin/menus", Component: "MenuList", Icon: "Grid", Sort: 6, Type: 1, Status: 1, CreatedAt: now, UpdatedAt: now},
	}

	for _, menu := range defaultMenus {
		h.db.Exec(h.db.Rebind("INSERT INTO menus (parent_id, name, title, path, component, icon, sort, type, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"),
			menu.ParentID, menu.Name, menu.Title, menu.Path, menu.Component, menu.Icon, menu.Sort, menu.Type, menu.Status, menu.CreatedAt, menu.UpdatedAt)
	}
}

// RegisterAdminRoutes 注册菜单管理路由
func (h *Handler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.GET("/", h.List)
	rg.GET("/tree", h.Tree)
	rg.POST("/", h.Create)
	rg.GET("/:id", h.Get)
	rg.PUT("/:id", h.Update)
	rg.DELETE("/:id", h.Delete)
	rg.PUT("/:id/status", h.UpdateStatus)
}

// List 获取菜单列表
func (h *Handler) List(c *gin.Context) {
	var menus []Menu
	query := h.db.Rebind("SELECT id, parent_id, name, title, path, component, icon, sort, type, status, created_at, updated_at FROM menus ORDER BY sort, id")
	if err := h.db.Select(&menus, query); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, menus)
}

// Tree 获取菜单树形结构
func (h *Handler) Tree(c *gin.Context) {
	var menus []Menu
	query := h.db.Rebind("SELECT id, parent_id, name, title, path, component, icon, sort, type, status, created_at, updated_at FROM menus WHERE status = 1 ORDER BY sort, id")
	if err := h.db.Select(&menus, query); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tree := h.buildMenuTree(menus, 0)
	c.JSON(http.StatusOK, tree)
}

// buildMenuTree 构建菜单树
func (h *Handler) buildMenuTree(menus []Menu, parentID int64) []*Menu {
	var tree []*Menu
	for i := range menus {
		if menus[i].ParentID == parentID {
			menu := &menus[i]
			menu.Children = h.buildMenuTree(menus, menu.ID)
			tree = append(tree, menu)
		}
	}
	return tree
}

// Create 创建菜单
func (h *Handler) Create(c *gin.Context) {
	var req struct {
		ParentID  int64  `json:"parent_id"`
		Name      string `json:"name" binding:"required"`
		Title     string `json:"title" binding:"required"`
		Path      string `json:"path"`
		Component string `json:"component"`
		Icon      string `json:"icon"`
		Sort      int    `json:"sort"`
		Type      int    `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.Type == 0 {
		req.Type = 1
	}

	now := time.Now()
	query := h.db.Rebind("INSERT INTO menus (parent_id, name, title, path, component, icon, sort, type, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)")
	result, err := h.db.Exec(query, req.ParentID, req.Name, req.Title, req.Path, req.Component, req.Icon, req.Sort, req.Type, now, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	menu := Menu{
		ID:        id,
		ParentID:  req.ParentID,
		Name:      req.Name,
		Title:     req.Title,
		Path:      req.Path,
		Component: req.Component,
		Icon:      req.Icon,
		Sort:      req.Sort,
		Type:      req.Type,
		Status:    1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	c.JSON(http.StatusCreated, menu)
}

// Get 获取单个菜单
func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var menu Menu
	query := h.db.Rebind("SELECT id, parent_id, name, title, path, component, icon, sort, type, status, created_at, updated_at FROM menus WHERE id = ?")
	if err := h.db.Get(&menu, query, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "菜单不存在"})
		return
	}
	c.JSON(http.StatusOK, menu)
}

// Update 更新菜单
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		ParentID  int64  `json:"parent_id"`
		Name      string `json:"name" binding:"required"`
		Title     string `json:"title" binding:"required"`
		Path      string `json:"path"`
		Component string `json:"component"`
		Icon      string `json:"icon"`
		Sort      int    `json:"sort"`
		Type      int    `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	query := h.db.Rebind("UPDATE menus SET parent_id = ?, name = ?, title = ?, path = ?, component = ?, icon = ?, sort = ?, type = ?, updated_at = ? WHERE id = ?")
	if _, err := h.db.Exec(query, req.ParentID, req.Name, req.Title, req.Path, req.Component, req.Icon, req.Sort, req.Type, now, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// UpdateStatus 更新菜单状态
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
	query := h.db.Rebind("UPDATE menus SET status = ?, updated_at = ? WHERE id = ?")
	if _, err := h.db.Exec(query, req.Status, now, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// Delete 删除菜单
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// 检查是否有子菜单
	var count int
	if err := h.db.Get(&count, h.db.Rebind("SELECT COUNT(*) FROM menus WHERE parent_id = ?"), id); err == nil && count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "存在子菜单，无法删除"})
		return
	}

	// 删除角色菜单关联
	h.db.Exec(h.db.Rebind("DELETE FROM role_menus WHERE menu_id = ?"), id)

	// 删除菜单
	query := h.db.Rebind("DELETE FROM menus WHERE id = ?")
	if _, err := h.db.Exec(query, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
