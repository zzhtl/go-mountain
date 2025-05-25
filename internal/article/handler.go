package article

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// Article 文章模型
type Article struct {
	ID        int64     `db:"id" json:"id"`
	ColumnID  int64     `db:"column_id" json:"column_id"`
	Title     string    `db:"title" json:"title"`
	Thumbnail string    `db:"thumbnail" json:"thumbnail"`
	Content   string    `db:"content" json:"content"`
	Author    string    `db:"author" json:"author"`
	Status    int       `db:"status" json:"status"` // 0:草稿 1:已发布
	ViewCount int       `db:"view_count" json:"view_count"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	// 关联字段
	ColumnName string `db:"column_name" json:"column_name,omitempty"`
}

// Handler 文章管理处理器
type Handler struct {
	db *sqlx.DB
}

// NewHandler 创建文章处理器
func NewHandler(db *sqlx.DB) *Handler {
	h := &Handler{db: db}
	// 创建文章表
	migration := `CREATE TABLE IF NOT EXISTS articles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		column_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		thumbnail TEXT,
		content TEXT,
		author TEXT,
		status INTEGER DEFAULT 0,
		view_count INTEGER DEFAULT 0,
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
	rg.PUT("/:id/status", h.UpdateStatus)
}

// RegisterMPRoutes 注册小程序路由
func (h *Handler) RegisterMPRoutes(rg *gin.RouterGroup) {
	rg.GET("/column/:columnId", h.ListByColumn)
	rg.GET("/:id", h.GetForMP)
}

// List 获取文章列表（后台管理）
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	columnID, _ := strconv.ParseInt(c.Query("column_id"), 10, 64)
	status, _ := strconv.Atoi(c.Query("status"))

	offset := (page - 1) * pageSize

	// 构建查询条件
	where := "WHERE 1=1"
	args := []interface{}{}

	if columnID > 0 {
		where += " AND a.column_id = ?"
		args = append(args, columnID)
	}

	if status >= 0 {
		where += " AND a.status = ?"
		args = append(args, status)
	}

	// 查询总数
	var total int
	countQuery := h.db.Rebind("SELECT COUNT(*) FROM articles a " + where)
	if err := h.db.Get(&total, countQuery, args...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查询列表
	args = append(args, pageSize, offset)
	query := h.db.Rebind(`
		SELECT a.*, c.name as column_name 
		FROM articles a 
		LEFT JOIN columns c ON a.column_id = c.id 
		` + where + ` 
		ORDER BY a.created_at DESC 
		LIMIT ? OFFSET ?
	`)

	var articles []Article
	if err := h.db.Select(&articles, query, args...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":      articles,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Create 创建文章
func (h *Handler) Create(c *gin.Context) {
	var req struct {
		ColumnID  int64  `json:"column_id" binding:"required"`
		Title     string `json:"title" binding:"required"`
		Thumbnail string `json:"thumbnail"`
		Content   string `json:"content"`
		Author    string `json:"author"`
		Status    int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	query := h.db.Rebind(`
		INSERT INTO articles (column_id, title, thumbnail, content, author, status, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	result, err := h.db.Exec(query, req.ColumnID, req.Title, req.Thumbnail, req.Content, req.Author, req.Status, now, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	article := Article{
		ID:        id,
		ColumnID:  req.ColumnID,
		Title:     req.Title,
		Thumbnail: req.Thumbnail,
		Content:   req.Content,
		Author:    req.Author,
		Status:    req.Status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	c.JSON(http.StatusCreated, article)
}

// Get 获取文章详情
func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var article Article
	query := h.db.Rebind(`
		SELECT a.*, c.name as column_name 
		FROM articles a 
		LEFT JOIN columns c ON a.column_id = c.id 
		WHERE a.id = ?
	`)
	if err := h.db.Get(&article, query, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}
	c.JSON(http.StatusOK, article)
}

// Update 更新文章
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		ColumnID  int64  `json:"column_id" binding:"required"`
		Title     string `json:"title" binding:"required"`
		Thumbnail string `json:"thumbnail"`
		Content   string `json:"content"`
		Author    string `json:"author"`
		Status    int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	query := h.db.Rebind(`
		UPDATE articles 
		SET column_id = ?, title = ?, thumbnail = ?, content = ?, author = ?, status = ?, updated_at = ? 
		WHERE id = ?
	`)
	if _, err := h.db.Exec(query, req.ColumnID, req.Title, req.Thumbnail, req.Content, req.Author, req.Status, now, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// Delete 删除文章
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	query := h.db.Rebind("DELETE FROM articles WHERE id = ?")
	if _, err := h.db.Exec(query, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// UpdateStatus 更新文章状态
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

	query := h.db.Rebind("UPDATE articles SET status = ?, updated_at = ? WHERE id = ?")
	if _, err := h.db.Exec(query, req.Status, time.Now(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// ListByColumn 根据栏目获取文章列表（小程序）
func (h *Handler) ListByColumn(c *gin.Context) {
	columnID, err := strconv.ParseInt(c.Param("columnId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid column id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize

	// 只查询已发布的文章
	query := h.db.Rebind(`
		SELECT id, title, thumbnail, author, view_count, created_at 
		FROM articles 
		WHERE column_id = ? AND status = 1 
		ORDER BY created_at DESC 
		LIMIT ? OFFSET ?
	`)

	var articles []Article
	if err := h.db.Select(&articles, query, columnID, pageSize, offset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查询总数
	var total int
	countQuery := h.db.Rebind("SELECT COUNT(*) FROM articles WHERE column_id = ? AND status = 1")
	h.db.Get(&total, countQuery, columnID)

	c.JSON(http.StatusOK, gin.H{
		"list":      articles,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetForMP 获取文章详情（小程序）
func (h *Handler) GetForMP(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// 增加浏览量
	h.db.Exec(h.db.Rebind("UPDATE articles SET view_count = view_count + 1 WHERE id = ?"), id)

	var article Article
	query := h.db.Rebind(`
		SELECT a.*, c.name as column_name 
		FROM articles a 
		LEFT JOIN columns c ON a.column_id = c.id 
		WHERE a.id = ? AND a.status = 1
	`)

	err = h.db.Get(&article, query, id)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, article)
}
