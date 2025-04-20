package mp

import (
	"github.com/gin-gonic/gin"
	"github.com/zzhtl/go-mountain/internal/user"
)

// RegisterRoutes 为小程序注册用户相关 API，路径前缀由上层提供
func RegisterRoutes(rg *gin.RouterGroup, userHandler *user.Handler) {
	userHandler.RegisterMPRoutes(rg)
}
