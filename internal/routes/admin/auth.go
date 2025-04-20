package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/zzhtl/go-mountain/internal/admin"
	"github.com/zzhtl/go-mountain/internal/config"
)

// RegisterAuthRoutes 注册后台管理登录路由
func RegisterAuthRoutes(rg *gin.RouterGroup, cfg *config.Config) {
	h := admin.NewHandler(cfg.Admin, cfg.JWT)
	rg.POST("/login", h.Login)
}
