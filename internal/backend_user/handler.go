package backend_user

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jmoiron/sqlx"
	"github.com/zzhtl/go-mountain/internal/config"
)

// BackendUser 后台用户模型
type BackendUser struct {
	ID        int64     `db:"id" json:"id"`
	Username  string    `db:"username" json:"username"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password" json:"-"`    // 密码不返回给前端
	Role      string    `db:"role" json:"role"`     // admin, editor
	Status    int       `db:"status" json:"status"` // 0:禁用 1:启用
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Handler 后台用户管理处理器
type Handler struct {
	db     *sqlx.DB
	jwtCfg config.JWTConfig
}

// NewHandler 创建后台用户处理器
func NewHandler(db *sqlx.DB, jwtCfg config.JWTConfig) *Handler {
	h := &Handler{db: db, jwtCfg: jwtCfg}
	// 创建后台用户表
	migration := `CREATE TABLE IF NOT EXISTS backend_users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		role TEXT DEFAULT 'editor',
		status INTEGER DEFAULT 1,
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
	rg.PUT("/:id/reset-password", h.ResetPassword)
}

// RegisterAuthRoutes 注册认证路由
func (h *Handler) RegisterAuthRoutes(rg *gin.RouterGroup) {
	rg.POST("/login", h.Login)
	rg.PUT("/change-password", h.ChangePassword)
}

// List 获取后台用户列表
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	// 查询总数
	var total int
	if err := h.db.Get(&total, h.db.Rebind("SELECT COUNT(*) FROM backend_users")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查询列表
	var users []BackendUser
	query := h.db.Rebind(`
		SELECT id, username, email, role, status, created_at, updated_at 
		FROM backend_users 
		ORDER BY created_at DESC 
		LIMIT ? OFFSET ?
	`)
	if err := h.db.Select(&users, query, pageSize, offset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":      users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Create 创建后台用户
func (h *Handler) Create(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Role     string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证角色
	if req.Role != "admin" && req.Role != "editor" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "角色只能是admin或editor"})
		return
	}

	// 生成随机密码
	password := h.generateRandomPassword()
	hashedPassword := h.hashPassword(password)

	now := time.Now()
	query := h.db.Rebind(`
		INSERT INTO backend_users (username, email, password, role, status, created_at, updated_at) 
		VALUES (?, ?, ?, ?, 1, ?, ?)
	`)
	result, err := h.db.Exec(query, req.Username, req.Email, hashedPassword, req.Role, now, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	user := BackendUser{
		ID:        id,
		Username:  req.Username,
		Email:     req.Email,
		Role:      req.Role,
		Status:    1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":     user,
		"password": password, // 返回明文密码给管理员
	})
}

// Get 获取单个后台用户
func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var user BackendUser
	query := h.db.Rebind(`
		SELECT id, username, email, role, status, created_at, updated_at 
		FROM backend_users WHERE id = ?
	`)
	if err := h.db.Get(&user, query, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// Update 更新后台用户信息
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Role     string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证角色
	if req.Role != "admin" && req.Role != "editor" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "角色只能是admin或editor"})
		return
	}

	now := time.Now()
	query := h.db.Rebind(`
		UPDATE backend_users 
		SET username = ?, email = ?, role = ?, updated_at = ? 
		WHERE id = ?
	`)
	if _, err := h.db.Exec(query, req.Username, req.Email, req.Role, now, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// UpdateStatus 更新用户状态
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
	query := h.db.Rebind("UPDATE backend_users SET status = ?, updated_at = ? WHERE id = ?")
	if _, err := h.db.Exec(query, req.Status, now, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// ResetPassword 管理员重置用户密码
func (h *Handler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// 生成新的随机密码
	newPassword := h.generateRandomPassword()
	hashedPassword := h.hashPassword(newPassword)

	now := time.Now()
	query := h.db.Rebind("UPDATE backend_users SET password = ?, updated_at = ? WHERE id = ?")
	if _, err := h.db.Exec(query, hashedPassword, now, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       id,
		"password": newPassword, // 返回新密码给管理员
	})
}

// Delete 删除后台用户
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	query := h.db.Rebind("DELETE FROM backend_users WHERE id = ?")
	if _, err := h.db.Exec(query, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// Login 后台用户登录
func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查询用户
	var user BackendUser
	query := h.db.Rebind("SELECT id, username, email, password, role, status FROM backend_users WHERE username = ?")
	if err := h.db.Get(&user, query, req.Username); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 检查用户状态
	if user.Status != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "账户已被禁用"})
		return
	}

	// 验证密码
	if !h.verifyPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 生成JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(h.jwtCfg.Secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// ChangePassword 用户修改密码
func (h *Handler) ChangePassword(c *gin.Context) {
	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查询当前用户
	var user BackendUser
	query := h.db.Rebind("SELECT id, password FROM backend_users WHERE id = ?")
	if err := h.db.Get(&user, query, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 验证旧密码
	if !h.verifyPassword(req.OldPassword, user.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "原密码错误"})
		return
	}

	// 更新密码
	hashedPassword := h.hashPassword(req.NewPassword)
	now := time.Now()
	updateQuery := h.db.Rebind("UPDATE backend_users SET password = ?, updated_at = ? WHERE id = ?")
	if _, err := h.db.Exec(updateQuery, hashedPassword, now, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// generateRandomPassword 生成随机密码
func (h *Handler) generateRandomPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	b := make([]byte, length)
	for i := range b {
		randomByte := make([]byte, 1)
		rand.Read(randomByte)
		b[i] = charset[randomByte[0]%byte(len(charset))]
	}
	return string(b)
}

// hashPassword 对密码进行哈希处理
func (h *Handler) hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", hash)
}

// verifyPassword 验证密码
func (h *Handler) verifyPassword(password, hashedPassword string) bool {
	return h.hashPassword(password) == hashedPassword
}
