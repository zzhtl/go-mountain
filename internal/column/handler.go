package column

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// Column 栏目模型
type Column struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	SortOrder   int       `db:"sort_order" json:"sort_order"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// Handler 栏目管理处理器
type Handler struct {
	db *sqlx.DB
}

// NewHandler 创建栏目处理器
func NewHandler(db *sqlx.DB) *Handler {
	h := &Handler{db: db}
	// 创建栏目表
	migration := `CREATE TABLE IF NOT EXISTS columns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		sort_order INTEGER DEFAULT 0,
		created_at DATETIME,
		updated_at DATETIME
	)`
	h.db.Exec(h.db.Rebind(migration))
	return h
}

// RegisterAdminRoutes 注册后台管理路由
func (h *Handler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.GET("/", h.List)
	rg.POST("/", h.Create)
	rg.GET("/:id", h.Get)
	rg.PUT("/:id", h.Update)
	rg.DELETE("/:id", h.Delete)
}

// RegisterMPRoutes 注册小程序路由
func (h *Handler) RegisterMPRoutes(rg *gin.RouterGroup) {
	rg.GET("/", h.ListForMP)
}

// List 获取所有栏目（后台管理）
func (h *Handler) List(c *gin.Context) {
	var columns []Column
	query := h.db.Rebind("SELECT id, name, description, sort_order, created_at, updated_at FROM columns ORDER BY sort_order, id")
	if err := h.db.Select(&columns, query); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, columns)
}

// Create 创建栏目
func (h *Handler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		SortOrder   int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	query := h.db.Rebind("INSERT INTO columns (name, description, sort_order, created_at, updated_at) VALUES (?, ?, ?, ?, ?)")
	result, err := h.db.Exec(query, req.Name, req.Description, req.SortOrder, now, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	column := Column{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	c.JSON(http.StatusCreated, column)
}

// Get 获取单个栏目
func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var column Column
	query := h.db.Rebind("SELECT id, name, description, sort_order, created_at, updated_at FROM columns WHERE id = ?")
	if err := h.db.Get(&column, query, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "column not found"})
		return
	}
	c.JSON(http.StatusOK, column)
}

// Update 更新栏目
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		SortOrder   int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	query := h.db.Rebind("UPDATE columns SET name = ?, description = ?, sort_order = ?, updated_at = ? WHERE id = ?")
	if _, err := h.db.Exec(query, req.Name, req.Description, req.SortOrder, now, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// Delete 删除栏目
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// 检查是否有文章关联
	var count int
	if err := h.db.Get(&count, h.db.Rebind("SELECT COUNT(*) FROM articles WHERE column_id = ?"), id); err == nil && count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete column with articles"})
		return
	}

	query := h.db.Rebind("DELETE FROM columns WHERE id = ?")
	if _, err := h.db.Exec(query, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ListForMP 小程序获取栏目列表
func (h *Handler) ListForMP(c *gin.Context) {
	var columns []Column
	query := h.db.Rebind("SELECT id, name, description FROM columns ORDER BY sort_order, id")
	if err := h.db.Select(&columns, query); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, columns)
}
