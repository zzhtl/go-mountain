package main

import (
	"log"

	"github.com/zzhtl/go-mountain/internal/config"
	"github.com/zzhtl/go-mountain/internal/db"
	"github.com/zzhtl/go-mountain/internal/server"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库连接
	dbConn, err := db.Init(cfg.Database)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer dbConn.Close()

	// 启动服务器
	srv := server.NewServer(dbConn, cfg)
	if err := srv.Run(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
