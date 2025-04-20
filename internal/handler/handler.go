package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// Item 表示一个简单的数据模型
type Item struct {
	ID          int64  `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`
}

// Handler 封装数据库连接
type Handler struct {
	db *sqlx.DB
}

// NewHandler 创建 Handler 并执行必要的迁移
func NewHandler(db *sqlx.DB) *Handler {
	h := &Handler{db: db}
	// 自动创建 items 表
	migration := `CREATE TABLE IF NOT EXISTS items (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT,
        description TEXT
    )`
	if _, err := h.db.Exec(h.db.Rebind(migration)); err != nil {
		panic(fmt.Errorf("创建 items 表失败: %w", err))
	}
	return h
}

// RegisterRoutes 注册 HTTP 路由
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/ping", h.Ping)
	r.GET("/items", h.ListItems)
	r.POST("/items", h.CreateItem)
	r.GET("/items/:id", h.GetItem)
	r.PUT("/items/:id", h.UpdateItem)
	r.DELETE("/items/:id", h.DeleteItem)
}

// Ping 测试接口
func (h *Handler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

// ListItems 列出所有 items
func (h *Handler) ListItems(c *gin.Context) {
	var items []Item
	query := h.db.Rebind("SELECT id, name, description FROM items")
	if err := h.db.Select(&items, query); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// CreateItem 创建一个新的 item
func (h *Handler) CreateItem(c *gin.Context) {
	var item Item
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query := h.db.Rebind("INSERT INTO items (name, description) VALUES (?, ?)")
	result, err := h.db.Exec(query, item.Name, item.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	item.ID = id
	c.JSON(http.StatusCreated, item)
}

// GetItem 根据 ID 获取 item
func (h *Handler) GetItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var item Item
	query := h.db.Rebind("SELECT id, name, description FROM items WHERE id = ?")
	if err := h.db.Get(&item, query, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// UpdateItem 更新指定 ID 的 item
func (h *Handler) UpdateItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var item Item
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query := h.db.Rebind("UPDATE items SET name = ?, description = ? WHERE id = ?")
	if _, err := h.db.Exec(query, item.Name, item.Description, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	item.ID = id
	c.JSON(http.StatusOK, item)
}

// DeleteItem 删除指定 ID 的 item
func (h *Handler) DeleteItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	query := h.db.Rebind("DELETE FROM items WHERE id = ?")
	if _, err := h.db.Exec(query, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
