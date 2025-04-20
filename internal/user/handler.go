package user

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/zzhtl/go-mountain/internal/config"
)

// User 表示注册用户模型
// 将来可根据需求扩展字段
// ... existing code ...

type User struct {
	ID        int64     `db:"id" json:"id"`
	Phone     string    `db:"phone" json:"phone"`
	OpenID    string    `db:"open_id" json:"openid"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Handler 封装用户管理及微信登录处理器
// 包含注册与后台管理接口

type Handler struct {
	db        *sqlx.DB
	wechatCfg config.WechatConfig
}

// NewHandler 创建用户模块处理器，初始化表并持有微信配置
func NewHandler(db *sqlx.DB, wechatCfg config.WechatConfig) *Handler {
	h := &Handler{db: db, wechatCfg: wechatCfg}
	// 创建 users 表
	migration := `CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        phone TEXT,
        open_id TEXT,
        name TEXT,
        created_at DATETIME,
        updated_at DATETIME
    )`
	h.db.Exec(h.db.Rebind(migration))
	return h
}

// RegisterMPRoutes 注册给小程序调用的用户 API，路径前缀由上层提供
func (h *Handler) RegisterMPRoutes(rg *gin.RouterGroup) {
	rg.POST("/register", h.Register)
	rg.POST("/login", h.WechatLogin)
}

// RegisterAdminRoutes 注册后台管理用户 API，需附加 JWT 鉴权
func (h *Handler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.GET("/", h.List)
	rg.GET("/:id", h.Get)
	rg.PUT("/:id", h.Update)
	rg.DELETE("/:id", h.Delete)
}

// Register 小程序通过手机号或微信 OpenID 注册
func (h *Handler) Register(c *gin.Context) {
	var req struct {
		Phone  string `json:"phone"`
		OpenID string `json:"openid"`
		Name   string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	now := time.Now()
	res, err := h.db.Exec(h.db.Rebind("INSERT INTO users (phone, open_id, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?)"),
		req.Phone, req.OpenID, req.Name, now, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := res.LastInsertId()
	user := User{ID: id, Phone: req.Phone, OpenID: req.OpenID, Name: req.Name, CreatedAt: now, UpdatedAt: now}
	c.JSON(http.StatusCreated, user)
}

// List 列出所有注册用户
func (h *Handler) List(c *gin.Context) {
	var users []User
	if err := h.db.Select(&users,
		h.db.Rebind("SELECT id, phone, open_id AS open_id, name, created_at, updated_at FROM users")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// Get 根据 ID 获取用户信息
func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var u User
	if err := h.db.Get(&u,
		h.db.Rebind("SELECT id, phone, open_id AS open_id, name, created_at, updated_at FROM users WHERE id = ?"), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, u)
}

// Update 更新用户信息
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Phone  string `json:"phone"`
		OpenID string `json:"openid"`
		Name   string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	now := time.Now()
	_, err = h.db.Exec(h.db.Rebind("UPDATE users SET phone = ?, open_id = ?, name = ?, updated_at = ? WHERE id = ?"),
		req.Phone, req.OpenID, req.Name, now, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

// Delete 删除用户
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if _, err := h.db.Exec(h.db.Rebind("DELETE FROM users WHERE id = ?"), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// WechatLogin 小程序微信 Code2Session 登录，返回用户信息
func (h *Handler) WechatLogin(c *gin.Context) {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 调用微信 Code2Session
	mp, err := miniProgram.NewMiniProgram(&miniProgram.UserConfig{
		AppID:     h.wechatCfg.AppID,
		Secret:    h.wechatCfg.Secret,
		HttpDebug: false,
		Debug:     false,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	session, err := mp.Auth.Session(context.Background(), req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	openid := session.OpenID
	// 查询或创建用户
	var u User
	err = h.db.Get(&u, h.db.Rebind("SELECT id, phone, open_id, name, created_at, updated_at FROM users WHERE open_id = ?"), openid)
	if err != nil {
		now := time.Now()
		res, err := h.db.Exec(h.db.Rebind("INSERT INTO users (open_id, created_at, updated_at) VALUES (?, ?, ?)"), openid, now, now)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		id, _ := res.LastInsertId()
		u = User{ID: id, OpenID: openid, CreatedAt: now, UpdatedAt: now}
	}
	c.JSON(http.StatusOK, u)
}
