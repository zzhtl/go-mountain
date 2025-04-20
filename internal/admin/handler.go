package admin

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/zzhtl/go-mountain/internal/config"
)

// Handler 处理后台管理员登录
// 包含验证配置中的用户名和密码，并生成 JWT Token

type Handler struct {
	adminCfg config.AdminConfig
	jwtCfg   config.JWTConfig
}

// NewHandler 创建 Admin Handler
func NewHandler(adminCfg config.AdminConfig, jwtCfg config.JWTConfig) *Handler {
	return &Handler{adminCfg: adminCfg, jwtCfg: jwtCfg}
}

// Login 后台登录接口，返回 JWT
func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Username != h.adminCfg.Username || req.Password != h.adminCfg.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	// 生成 JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": req.Username,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(h.jwtCfg.Secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
