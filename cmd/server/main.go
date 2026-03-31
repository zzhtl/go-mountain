package main

import (
	"context"
	"log"

	"github.com/zzhtl/go-mountain/internal/config"
	"github.com/zzhtl/go-mountain/internal/db"
	"github.com/zzhtl/go-mountain/internal/model"
	"github.com/zzhtl/go-mountain/internal/server"
	"github.com/zzhtl/go-mountain/internal/service"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	database, err := db.Init(cfg.Database)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 自动迁移表结构
	if err := database.AutoMigrate(
		&model.BackendUser{},
		&model.Role{},
		&model.Menu{},
		&model.RoleMenu{},
		&model.Article{},
		&model.Column{},
		&model.User{},
		&model.Activity{},
		&model.Registration{},
		&model.Payment{},
		&model.CodegenConfig{},
		&model.OperationLog{},
		&model.SystemConfig{},
	); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 初始化默认数据
	ctx := context.Background()
	roleSvc := service.NewRoleService(database)
	if err := roleSvc.InitDefaultRoles(ctx); err != nil {
		log.Printf("初始化默认角色失败: %v", err)
	}

	menuSvc := service.NewMenuService(database)
	if err := menuSvc.InitDefaultMenus(ctx); err != nil {
		log.Printf("初始化默认菜单失败: %v", err)
	}

	systemConfigSvc := service.NewSystemConfigService(database)
	systemConfigSvc.InitDefaultConfigs(ctx)

	// 启动服务器
	srv := server.NewServer(database, cfg)
	if err := srv.Run(); err != nil {
		log.Fatalf("服务器运行失败: %v", err)
	}
}
