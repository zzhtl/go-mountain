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

// BackendUserV2 后台用户模型（支持动态角色）
type BackendUserV2 struct {
	ID          int64     `db:"id" json:"id"`
	Username    string    `db:"username" json:"username"`
	Email       string    `db:"email" json:"email"`
	Password    string    `db:"password" json:"-"` // 密码不返回给前端
	RoleID      int64     `db:"role_id" json:"role_id"`
	Status      int       `db:"status" json:"status"` // 0:禁用 1:启用
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
	RoleName    string    `db:"role_name" json:"role_name,omitempty"`
	RoleDisplay string    `db:"role_display" json:"role_display,omitempty"`
}

// UserMenus 用户菜单权限
type UserMenus struct {
	UserID   int64   `json:"user_id"`
	Username string  `json:"username"`
	Menus    []*Menu `json:"menus"`
}

// Menu 菜单信息
type Menu struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Title     string  `json:"title"`
	Path      string  `json:"path"`
	Component string  `json:"component"`
	Icon      string  `json:"icon"`
	Sort      int     `json:"sort"`
	Type      int     `json:"type"`
	Children  []*Menu `json:"children,omitempty"`
}

// HandlerV2 后台用户管理处理器V2
type HandlerV2 struct {
	db     *sqlx.DB
	jwtCfg config.JWTConfig
}

// NewHandlerV2 创建后台用户处理器V2
func NewHandlerV2(db *sqlx.DB, jwtCfg config.JWTConfig) *HandlerV2 {
	h := &HandlerV2{db: db, jwtCfg: jwtCfg}
	h.updateTables()
	return h
}

// updateTables 更新数据库表结构
func (h *HandlerV2) updateTables() {
	// 更新后台用户表，添加role_id字段
	migration := `CREATE TABLE IF NOT EXISTS backend_users_v2 (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		role_id INTEGER NOT NULL,
		status INTEGER DEFAULT 1,
		created_at DATETIME,
		updated_at DATETIME
	)`
	h.db.Exec(h.db.Rebind(migration))

	// 迁移数据（如果存在旧表）
	h.migrateFromOldTable()
}

// migrateFromOldTable 从旧表迁移数据
func (h *HandlerV2) migrateFromOldTable() {
	// 检查新表是否有数据
	var count int
	h.db.Get(&count, h.db.Rebind("SELECT COUNT(*) FROM backend_users_v2"))
	if count > 0 {
		return
	}

	// 检查旧表是否存在
	var oldCount int
	err := h.db.Get(&oldCount, h.db.Rebind("SELECT COUNT(*) FROM backend_users"))
	if err != nil {
		return // 旧表不存在
	}

	// 迁移数据
	rows, err := h.db.Query(h.db.Rebind("SELECT id, username, email, password, role, status, created_at, updated_at FROM backend_users"))
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var username, email, password, role string
		var status int
		var createdAt, updatedAt time.Time

		rows.Scan(&id, &username, &email, &password, &role, &status, &createdAt, &updatedAt)

		// 根据角色名称获取角色ID
		var roleID int64 = 1 // 默认editor角色
		if role == "admin" {
			roleID = 1 // admin角色ID为1
		} else if role == "editor" {
			roleID = 2 // editor角色ID为2
		}

		// 插入新表
		h.db.Exec(h.db.Rebind("INSERT INTO backend_users_v2 (username, email, password, role_id, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)"),
			username, email, password, roleID, status, createdAt, updatedAt)
	}
}

// RegisterAdminRoutes 注册后台管理路由
func (h *HandlerV2) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.GET("/", h.List)
	rg.POST("/", h.Create)
	rg.GET("/:id", h.Get)
	rg.PUT("/:id", h.Update)
	rg.DELETE("/:id", h.Delete)
	rg.PUT("/:id/status", h.UpdateStatus)
	rg.PUT("/:id/reset-password", h.ResetPassword)
	rg.GET("/current/menus", h.GetCurrentUserMenus)
}

// RegisterAuthRoutes 注册认证路由
func (h *HandlerV2) RegisterAuthRoutes(rg *gin.RouterGroup) {
	rg.POST("/login", h.Login)
	rg.PUT("/change-password", h.ChangePassword)
}

// List 获取后台用户列表
func (h *HandlerV2) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	// 查询总数
	var total int
	if err := h.db.Get(&total, h.db.Rebind("SELECT COUNT(*) FROM backend_users_v2")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查询列表（关联角色信息）
	var users []BackendUserV2
	query := h.db.Rebind(`
		SELECT u.id, u.username, u.email, u.role_id, u.status, u.created_at, u.updated_at,
		       r.name as role_name, r.display_name as role_display
		FROM backend_users_v2 u 
		LEFT JOIN roles r ON u.role_id = r.id
		ORDER BY u.created_at DESC 
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
func (h *HandlerV2) Create(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		RoleID   int64  `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证角色是否存在
	var roleCount int
	if err := h.db.Get(&roleCount, h.db.Rebind("SELECT COUNT(*) FROM roles WHERE id = ? AND status = 1"), req.RoleID); err != nil || roleCount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色"})
		return
	}

	// 生成随机密码
	password := h.generateRandomPassword()
	hashedPassword := h.hashPassword(password)

	now := time.Now()
	query := h.db.Rebind(`
		INSERT INTO backend_users_v2 (username, email, password, role_id, status, created_at, updated_at) 
		VALUES (?, ?, ?, ?, 1, ?, ?)
	`)
	result, err := h.db.Exec(query, req.Username, req.Email, hashedPassword, req.RoleID, now, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	user := BackendUserV2{
		ID:        id,
		Username:  req.Username,
		Email:     req.Email,
		RoleID:    req.RoleID,
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
func (h *HandlerV2) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var user BackendUserV2
	query := h.db.Rebind(`
		SELECT u.id, u.username, u.email, u.role_id, u.status, u.created_at, u.updated_at,
		       r.name as role_name, r.display_name as role_display
		FROM backend_users_v2 u 
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.id = ?
	`)
	if err := h.db.Get(&user, query, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// Update 更新后台用户信息
func (h *HandlerV2) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		RoleID   int64  `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证角色是否存在
	var roleCount int
	if err := h.db.Get(&roleCount, h.db.Rebind("SELECT COUNT(*) FROM roles WHERE id = ? AND status = 1"), req.RoleID); err != nil || roleCount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色"})
		return
	}

	now := time.Now()
	query := h.db.Rebind(`
		UPDATE backend_users_v2 
		SET username = ?, email = ?, role_id = ?, updated_at = ? 
		WHERE id = ?
	`)
	if _, err := h.db.Exec(query, req.Username, req.Email, req.RoleID, now, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// UpdateStatus 更新用户状态
func (h *HandlerV2) UpdateStatus(c *gin.Context) {
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
	query := h.db.Rebind("UPDATE backend_users_v2 SET status = ?, updated_at = ? WHERE id = ?")
	if _, err := h.db.Exec(query, req.Status, now, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// ResetPassword 管理员重置用户密码
func (h *HandlerV2) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// 生成新的随机密码
	newPassword := h.generateRandomPassword()
	hashedPassword := h.hashPassword(newPassword)

	now := time.Now()
	query := h.db.Rebind("UPDATE backend_users_v2 SET password = ?, updated_at = ? WHERE id = ?")
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
func (h *HandlerV2) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	query := h.db.Rebind("DELETE FROM backend_users_v2 WHERE id = ?")
	if _, err := h.db.Exec(query, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// Login 后台用户登录
func (h *HandlerV2) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查询用户
	var user BackendUserV2
	query := h.db.Rebind(`
		SELECT u.id, u.username, u.email, u.password, u.role_id, u.status,
		       r.name as role_name, r.display_name as role_display
		FROM backend_users_v2 u 
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.username = ?
	`)
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
		"role_id":  user.RoleID,
		"role":     user.RoleName,
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
			"id":           user.ID,
			"username":     user.Username,
			"email":        user.Email,
			"role_id":      user.RoleID,
			"role":         user.RoleName,
			"role_display": user.RoleDisplay,
		},
	})
}

// ChangePassword 用户修改密码
func (h *HandlerV2) ChangePassword(c *gin.Context) {
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
	var user BackendUserV2
	query := h.db.Rebind("SELECT id, password FROM backend_users_v2 WHERE id = ?")
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
	updateQuery := h.db.Rebind("UPDATE backend_users_v2 SET password = ?, updated_at = ? WHERE id = ?")
	if _, err := h.db.Exec(updateQuery, hashedPassword, now, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// GetCurrentUserMenus 获取当前用户的菜单权限
func (h *HandlerV2) GetCurrentUserMenus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取用户角色
	var roleID int64
	query := h.db.Rebind("SELECT role_id FROM backend_users_v2 WHERE id = ?")
	if err := h.db.Get(&roleID, query, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户角色失败"})
		return
	}

	// 获取角色对应的菜单
	menuQuery := h.db.Rebind(`
		SELECT m.id, m.parent_id, m.name, m.title, m.path, m.component, m.icon, m.sort, m.type
		FROM menus m
		INNER JOIN role_menus rm ON m.id = rm.menu_id
		WHERE rm.role_id = ? AND m.status = 1
		ORDER BY m.sort, m.id
	`)
	rows, err := h.db.Query(menuQuery, roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取菜单失败"})
		return
	}
	defer rows.Close()

	menuMap := make(map[int64]*Menu)
	var rootMenus []*Menu

	for rows.Next() {
		var menu Menu
		var parentID int64
		if err := rows.Scan(&menu.ID, &parentID, &menu.Name, &menu.Title, &menu.Path, &menu.Component, &menu.Icon, &menu.Sort, &menu.Type); err != nil {
			continue
		}

		menuMap[menu.ID] = &menu

		if parentID == 0 {
			rootMenus = append(rootMenus, &menu)
		} else {
			if parent, exists := menuMap[parentID]; exists {
				if parent.Children == nil {
					parent.Children = []*Menu{}
				}
				parent.Children = append(parent.Children, &menu)
			}
		}
	}

	c.JSON(http.StatusOK, rootMenus)
}

// 工具函数
func (h *HandlerV2) generateRandomPassword() string {
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

func (h *HandlerV2) hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", hash)
}

func (h *HandlerV2) verifyPassword(password, hashedPassword string) bool {
	return h.hashPassword(password) == hashedPassword
}
