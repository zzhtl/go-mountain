package router

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/zzhtl/go-mountain/internal/config"
	"github.com/zzhtl/go-mountain/internal/handler"
	"github.com/zzhtl/go-mountain/internal/middleware"
	"github.com/zzhtl/go-mountain/internal/service"
)

// Setup 配置所有路由
func Setup(engine *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// 全局中间件
	engine.Use(middleware.CORS())

	// 设置最大文件上传大小
	engine.MaxMultipartMemory = 50 << 20 // 50 MB

	// 静态文件服务
	engine.Static("/uploads", "./uploads")

	// API 路由
	api := engine.Group("/api")

	// 健康检查
	api.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// 创建 services
	authSvc := service.NewAuthService(db, cfg.JWT.Secret)
	backendUserSvc := service.NewBackendUserService(db)
	articleSvc := service.NewArticleService(db)
	columnSvc := service.NewColumnService(db)
	roleSvc := service.NewRoleService(db)
	menuSvc := service.NewMenuService(db)
	userSvc := service.NewUserService(db, cfg.Wechat.AppID, cfg.Wechat.Secret)
	activitySvc := service.NewActivityService(db)
	registrationSvc := service.NewRegistrationService(db)
	paymentSvc := service.NewPaymentService(db)
	systemConfigSvc := service.NewSystemConfigService(db)

	// 创建 handlers
	authHandler := handler.NewAuthHandler(authSvc)
	backendUserHandler := handler.NewBackendUserHandler(backendUserSvc)
	articleHandler := handler.NewArticleHandler(articleSvc)
	columnHandler := handler.NewColumnHandler(columnSvc)
	roleHandler := handler.NewRoleHandler(roleSvc)
	menuHandler := handler.NewMenuHandler(menuSvc)
	userHandler := handler.NewUserHandler(userSvc)
	uploadHandler := handler.NewUploadHandler()
	activityHandler := handler.NewActivityHandler(activitySvc)
	registrationHandler := handler.NewRegistrationHandler(registrationSvc)
	paymentHandler := handler.NewPaymentHandler(paymentSvc)
	systemConfigHandler := handler.NewSystemConfigHandler(systemConfigSvc)

	// ==================== 小程序 API ====================
	mp := api.Group("/mp")
	{
		mp.POST("/login", userHandler.WechatLogin)
		mp.POST("/register", userHandler.Register)

		mpColumns := mp.Group("/columns")
		mpColumns.GET("/", columnHandler.ListForMP)

		mpArticles := mp.Group("/articles")
		mpArticles.GET("/column/:columnId", articleHandler.ListByColumn)
		mpArticles.GET("/:id", articleHandler.GetForMP)

		// 活动（公开）
		mpActivities := mp.Group("/activities")
		mpActivities.GET("/", activityHandler.ListForMP)
		mpActivities.GET("/:id", activityHandler.GetForMP)

		// 需要小程序用户认证的接口
		mpAuth := mp.Group("")
		mpAuth.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			// 报名
			mpAuth.POST("/registrations", registrationHandler.Create)
			mpAuth.PUT("/registrations/:id/cancel", registrationHandler.Cancel)
			mpAuth.GET("/registrations/mine", registrationHandler.MyRegistrations)

			// 支付
			mpAuth.POST("/payments/create", paymentHandler.CreateOrder)
			mpAuth.GET("/payments/query", paymentHandler.QueryOrder)
		}
	}

	// 微信支付回调（不需要任何认证）
	api.POST("/payment/wechat/notify", paymentHandler.WechatNotify)

	// ==================== 后台 API ====================
	admin := api.Group("/admin")

	// 认证路由（不需要 JWT）
	auth := admin.Group("/backend-auth")
	{
		auth.POST("/login", authHandler.Login)
	}

	// 需要 JWT 认证 + RBAC 权限校验的路由
	adminAuth := admin.Group("")
	adminAuth.Use(middleware.JWTAuth(cfg.JWT.Secret))
	adminAuth.Use(middleware.RBACAuth(db))
	{
		// 修改密码
		adminAuth.PUT("/backend-auth/change-password", authHandler.ChangePassword)

		// 后台用户管理
		bu := adminAuth.Group("/backend-users")
		bu.GET("/", backendUserHandler.List)
		bu.POST("/", backendUserHandler.Create)
		bu.GET("/current/menus", backendUserHandler.GetCurrentUserMenus)
		bu.GET("/:id", backendUserHandler.Get)
		bu.PUT("/:id", backendUserHandler.Update)
		bu.DELETE("/:id", backendUserHandler.Delete)
		bu.PUT("/:id/status", backendUserHandler.UpdateStatus)
		bu.PUT("/:id/reset-password", backendUserHandler.ResetPassword)

		// 小程序用户管理
		users := adminAuth.Group("/users")
		users.GET("/", userHandler.List)
		users.GET("/:id", userHandler.Get)
		users.PUT("/:id", userHandler.Update)
		users.DELETE("/:id", userHandler.Delete)

		// 文章管理
		articles := adminAuth.Group("/articles")
		articles.GET("/", articleHandler.List)
		articles.POST("/", articleHandler.Create)
		articles.GET("/:id", articleHandler.Get)
		articles.PUT("/:id", articleHandler.Update)
		articles.DELETE("/:id", articleHandler.Delete)
		articles.PUT("/:id/status", articleHandler.UpdateStatus)

		// 栏目管理
		columns := adminAuth.Group("/columns")
		columns.GET("/", columnHandler.List)
		columns.POST("/", columnHandler.Create)
		columns.GET("/:id", columnHandler.Get)
		columns.PUT("/:id", columnHandler.Update)
		columns.DELETE("/:id", columnHandler.Delete)

		// 角色管理
		roles := adminAuth.Group("/roles")
		roles.GET("/", roleHandler.List)
		roles.POST("/", roleHandler.Create)
		roles.GET("/:id", roleHandler.Get)
		roles.PUT("/:id", roleHandler.Update)
		roles.DELETE("/:id", roleHandler.Delete)
		roles.PUT("/:id/status", roleHandler.UpdateStatus)
		roles.GET("/:id/menus", roleHandler.GetRoleMenus)
		roles.PUT("/:id/menus", roleHandler.UpdateRoleMenus)

		// 菜单管理
		menus := adminAuth.Group("/menus")
		menus.GET("/", menuHandler.List)
		menus.GET("/tree", menuHandler.Tree)
		menus.POST("/", menuHandler.Create)
		menus.GET("/:id", menuHandler.Get)
		menus.PUT("/:id", menuHandler.Update)
		menus.DELETE("/:id", menuHandler.Delete)
		menus.PUT("/:id/status", menuHandler.UpdateStatus)

		// 活动管理
		activities := adminAuth.Group("/activities")
		activities.GET("/", activityHandler.List)
		activities.POST("/", activityHandler.Create)
		activities.GET("/:id", activityHandler.Get)
		activities.PUT("/:id", activityHandler.Update)
		activities.DELETE("/:id", activityHandler.Delete)
		activities.PUT("/:id/status", activityHandler.UpdateStatus)

		// 报名管理
		registrations := adminAuth.Group("/registrations")
		registrations.GET("/", registrationHandler.List)
		registrations.GET("/:id", registrationHandler.Get)

		// 支付管理
		payments := adminAuth.Group("/payments")
		payments.GET("/", paymentHandler.List)
		payments.GET("/:id", paymentHandler.Get)
		payments.PUT("/:id/refund", paymentHandler.Refund)

		// 系统配置管理
		sysConfigs := adminAuth.Group("/system-configs")
		sysConfigs.GET("/", systemConfigHandler.List)
		sysConfigs.GET("/groups", systemConfigHandler.GetGroups)
		sysConfigs.POST("/", systemConfigHandler.Save)
		sysConfigs.POST("/batch", systemConfigHandler.BatchSave)
		sysConfigs.DELETE("/", systemConfigHandler.Delete)

		// 文件上传
		upload := adminAuth.Group("/upload")
		upload.POST("/image", uploadHandler.UploadImage)
		upload.POST("/video", uploadHandler.UploadVideo)
	}

	// Admin UI 静态文件
	engine.StaticFS("/web", gin.Dir("frontend-admin/dist", false))

	// SPA 回退
	engine.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/web/") {
			c.File("frontend-admin/dist/index.html")
			return
		}
		c.JSON(404, gin.H{"code": 404, "message": "Not found"})
	})

	// 重定向
	engine.GET("/admin", func(c *gin.Context) { c.Redirect(302, "/web/") })
	engine.GET("/admin/", func(c *gin.Context) { c.Redirect(302, "/web/") })
}
