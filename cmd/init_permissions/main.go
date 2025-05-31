package main

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zzhtl/go-mountain/internal/config"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("无法加载配置: %v", err)
	}

	// 连接数据库
	db, err := sqlx.Connect(cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("无法连接数据库: %v", err)
	}
	defer db.Close()

	// 初始化权限
	if err := initializePermissions(db); err != nil {
		log.Fatalf("初始化权限失败: %v", err)
	}

	log.Println("权限初始化完成!")
}

func initializePermissions(db *sqlx.DB) error {
	// 获取所有菜单ID
	var menuIDs []int64
	err := db.Select(&menuIDs, "SELECT id FROM menus WHERE status = 1")
	if err != nil {
		return err
	}

	// 获取admin角色ID
	var adminRoleID int64
	err = db.Get(&adminRoleID, "SELECT id FROM roles WHERE name = 'admin' LIMIT 1")
	if err != nil {
		return err
	}

	// 为admin角色分配所有菜单权限
	for _, menuID := range menuIDs {
		_, err = db.Exec("INSERT OR IGNORE INTO role_menus (role_id, menu_id) VALUES (?, ?)", adminRoleID, menuID)
		if err != nil {
			return err
		}
	}

	// 获取editor角色ID
	var editorRoleID int64
	err = db.Get(&editorRoleID, "SELECT id FROM roles WHERE name = 'editor' LIMIT 1")
	if err != nil {
		log.Printf("编辑员角色不存在，跳过权限分配")
		return nil
	}

	// 为editor角色分配部分菜单权限（文章管理、栏目管理）
	editorMenus := []string{"articles", "columns"}
	for _, menuName := range editorMenus {
		var menuID int64
		err = db.Get(&menuID, "SELECT id FROM menus WHERE name = ? LIMIT 1", menuName)
		if err != nil {
			log.Printf("菜单 %s 不存在，跳过", menuName)
			continue
		}
		_, err = db.Exec("INSERT OR IGNORE INTO role_menus (role_id, menu_id) VALUES (?, ?)", editorRoleID, menuID)
		if err != nil {
			return err
		}
	}

	return nil
}
