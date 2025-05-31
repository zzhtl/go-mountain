package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/zzhtl/go-mountain/internal/backend_user"
	"github.com/zzhtl/go-mountain/internal/middleware"
)

// RegisterBackendUserRoutes 注册后台用户管理路由
func RegisterBackendUserRoutes(rg *gin.RouterGroup, backendUserHandler *backend_user.Handler, jwtSecret string) {
	rg.Use(middleware.JWTAuth(jwtSecret))
	backendUserHandler.RegisterAdminRoutes(rg)
}

// RegisterBackendAuthRoutes 注册后台用户认证路由
func RegisterBackendAuthRoutes(rg *gin.RouterGroup, backendUserHandler *backend_user.Handler, jwtSecret string) {
	// 登录接口不需要认证
	rg.POST("/login", backendUserHandler.Login)

	// 修改密码需要认证
	authGroup := rg.Group("")
	authGroup.Use(middleware.JWTAuth(jwtSecret))
	authGroup.PUT("/change-password", backendUserHandler.ChangePassword)
}
