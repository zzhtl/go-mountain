package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/zzhtl/go-mountain/internal/config"
	"github.com/zzhtl/go-mountain/internal/handler"
	adminRoutes "github.com/zzhtl/go-mountain/internal/routes/admin"
	mpRoutes "github.com/zzhtl/go-mountain/internal/routes/mp"
	"github.com/zzhtl/go-mountain/internal/user"
)

// Server 封装 Gin 引擎和数据库连接和全局配置
// 包含启动方法

type Server struct {
	engine *gin.Engine
	db     *sqlx.DB
	cfg    *config.Config
}

// NewServer 创建一个新的 Server 实例，初始化路由，接收完整配置
func NewServer(dbConn *sqlx.DB, cfg *config.Config) *Server {
	engine := gin.Default()
	// Admin UI 静态文件服务，前端位于 frontend-admin/dist
	engine.Static("/admin", "frontend-admin/dist")
	// SPA 刷新路由回退到 index.html
	engine.GET("/admin/*any", func(c *gin.Context) {
		c.File("frontend-admin/dist/index.html")
	})
	// API 路由分组
	apiGroup := engine.Group("/api")
	// 业务基础接口分组
	biz := apiGroup.Group("")
	bizHandler := handler.NewHandler(dbConn)
	biz.GET("/ping", bizHandler.Ping)
	biz.GET("/items", bizHandler.ListItems)
	biz.POST("/items", bizHandler.CreateItem)
	biz.GET("/items/:id", bizHandler.GetItem)
	biz.PUT("/items/:id", bizHandler.UpdateItem)
	biz.DELETE("/items/:id", bizHandler.DeleteItem)
	// 小程序用户 API
	mp := apiGroup.Group("/mp")
	mpHandler := user.NewHandler(dbConn, cfg.Wechat)
	mpRoutes.RegisterRoutes(mp, mpHandler)
	// 后台管理登录 API
	auth := apiGroup.Group("/admin/auth")
	adminRoutes.RegisterAuthRoutes(auth, cfg)
	// 后台用户管理 API
	adminUsers := apiGroup.Group("/admin/users")
	adminRoutes.RegisterUserRoutes(adminUsers, mpHandler, cfg.JWT.Secret)
	return &Server{
		engine: engine,
		db:     dbConn,
		cfg:    cfg,
	}
}

// Run 启动 HTTP 服务器
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%d", s.cfg.Server.Port)
	return s.engine.Run(addr)
}
