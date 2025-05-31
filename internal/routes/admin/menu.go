package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/zzhtl/go-mountain/internal/menu"
)

// RegisterMenuRoutes 注册菜单管理路由
func RegisterMenuRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	handler := menu.NewHandler(db)
	menuGroup := router.Group("/menus")
	handler.RegisterAdminRoutes(menuGroup)
}
