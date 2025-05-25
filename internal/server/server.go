package server

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/zzhtl/go-mountain/internal/article"
	"github.com/zzhtl/go-mountain/internal/column"
	"github.com/zzhtl/go-mountain/internal/config"
	"github.com/zzhtl/go-mountain/internal/handler"
	"github.com/zzhtl/go-mountain/internal/middleware"
	adminRoutes "github.com/zzhtl/go-mountain/internal/routes/admin"
	mpRoutes "github.com/zzhtl/go-mountain/internal/routes/mp"
	"github.com/zzhtl/go-mountain/internal/upload"
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

	// 设置最大文件上传大小
	engine.MaxMultipartMemory = 50 << 20 // 50 MB

	// 静态文件服务
	engine.Static("/uploads", "./uploads")

	// API 路由分组 - 必须在静态文件路由之前定义
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

	// 小程序栏目 API
	mpColumn := mp.Group("/columns")
	columnHandler := column.NewHandler(dbConn)
	columnHandler.RegisterMPRoutes(mpColumn)

	// 小程序文章 API
	mpArticle := mp.Group("/articles")
	articleHandler := article.NewHandler(dbConn)
	articleHandler.RegisterMPRoutes(mpArticle)

	// 后台管理登录 API
	auth := apiGroup.Group("/admin/auth")
	adminRoutes.RegisterAuthRoutes(auth, cfg)

	// 后台用户管理 API
	adminUsers := apiGroup.Group("/admin/users")
	adminRoutes.RegisterUserRoutes(adminUsers, mpHandler, cfg.JWT.Secret)

	// 后台栏目管理 API（需要JWT认证）
	adminColumns := apiGroup.Group("/admin/columns")
	adminColumns.Use(middleware.JWTAuth(cfg.JWT.Secret))
	columnHandler.RegisterAdminRoutes(adminColumns)

	// 后台文章管理 API（需要JWT认证）
	adminArticles := apiGroup.Group("/admin/articles")
	adminArticles.Use(middleware.JWTAuth(cfg.JWT.Secret))
	articleHandler.RegisterAdminRoutes(adminArticles)

	// 文件上传 API（需要JWT认证）
	adminUpload := apiGroup.Group("/admin/upload")
	adminUpload.Use(middleware.JWTAuth(cfg.JWT.Secret))
	uploadHandler := upload.NewHandler()
	uploadHandler.RegisterRoutes(adminUpload)

	// Admin UI 静态文件服务 - 使用不同的路径避免与API冲突
	engine.StaticFS("/web", gin.Dir("frontend-admin/dist", false))

	// SPA回退处理 - 所有/web/*路径都返回index.html
	engine.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/web/") {
			c.File("frontend-admin/dist/index.html")
			return
		}
		c.JSON(404, gin.H{"error": "Not found"})
	})

	// 添加重定向，从 /admin 到 /web
	engine.GET("/admin", func(c *gin.Context) {
		c.Redirect(302, "/web/")
	})
	engine.GET("/admin/", func(c *gin.Context) {
		c.Redirect(302, "/web/")
	})

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
