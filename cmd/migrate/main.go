package main

import (
	"context"
	"fmt"
	"log"

	"github.com/zzhtl/go-mountain/internal/config"
	"github.com/zzhtl/go-mountain/internal/db"
	"github.com/zzhtl/go-mountain/internal/model"
	"github.com/zzhtl/go-mountain/internal/pkg/crypto"
	"github.com/zzhtl/go-mountain/internal/service"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	database, err := db.Init(cfg.Database)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	log.Println("开始数据库迁移...")

	// 自动迁移所有表结构
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
	log.Println("表结构迁移完成")

	ctx := context.Background()

	// 初始化默认角色
	roleSvc := service.NewRoleService(database)
	if err := roleSvc.InitDefaultRoles(ctx); err != nil {
		log.Printf("初始化默认角色失败: %v", err)
	} else {
		log.Println("默认角色初始化完成")
	}

	// 初始化默认菜单
	menuSvc := service.NewMenuService(database)
	if err := menuSvc.InitDefaultMenus(ctx); err != nil {
		log.Printf("初始化默认菜单失败: %v", err)
	} else {
		log.Println("默认菜单初始化完成")
	}

	// 创建默认管理员账号（如果不存在）
	var adminCount int64
	database.Model(&model.BackendUser{}).Where("username = ?", "admin").Count(&adminCount)
	if adminCount == 0 {
		var adminRole model.Role
		if err := database.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
			log.Printf("未找到 admin 角色: %v", err)
		} else {
			password := service.GenerateRandomPassword(10)
			hashedPassword, err := crypto.HashPassword(password)
			if err != nil {
				log.Fatalf("密码加密失败: %v", err)
			}

			admin := &model.BackendUser{
				Username:        "admin",
				Email:           "admin@example.com",
				Password:        hashedPassword,
				PasswordVersion: 2,
				RoleID:          adminRole.ID,
				Status:          1,
			}

			if err := database.Create(admin).Error; err != nil {
				log.Fatalf("创建管理员账号失败: %v", err)
			}

			fmt.Println("========================================")
			fmt.Printf("管理员账号创建成功\n")
			fmt.Printf("用户名: admin\n")
			fmt.Printf("密码: %s\n", password)
			fmt.Println("请妥善保管密码！")
			fmt.Println("========================================")
		}
	} else {
		log.Println("管理员账号已存在，跳过创建")
	}

	log.Println("数据库迁移全部完成！")
}
