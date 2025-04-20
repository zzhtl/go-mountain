package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/zzhtl/go-mountain/internal/middleware"
	"github.com/zzhtl/go-mountain/internal/user"
)

// RegisterUserRoutes 注册后台用户管理 API，添加 JWT 中间件
func RegisterUserRoutes(rg *gin.RouterGroup, userHandler *user.Handler, jwtSecret string) {
	rg.Use(middleware.JWTAuth(jwtSecret))
	userHandler.RegisterAdminRoutes(rg)
}
