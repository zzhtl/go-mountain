package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/zzhtl/go-mountain/internal/role"
)

// RegisterRoleRoutes 注册角色管理路由
func RegisterRoleRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	handler := role.NewHandler(db)
	roleGroup := router.Group("/roles")
	handler.RegisterAdminRoutes(roleGroup)
}
