package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/zzhtl/go-mountain/internal/config"
	"github.com/zzhtl/go-mountain/internal/router"
)

// Server 封装 HTTP 服务器
type Server struct {
	engine *gin.Engine
	db     *gorm.DB
	cfg    *config.Config
}

// NewServer 创建服务器实例
func NewServer(db *gorm.DB, cfg *config.Config) *Server {
	engine := gin.Default()
	router.Setup(engine, db, cfg)

	return &Server{
		engine: engine,
		db:     db,
		cfg:    cfg,
	}
}

// Run 启动 HTTP 服务器（支持优雅关闭）
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%d", s.cfg.Server.Port)

	srv := &http.Server{
		Addr:    addr,
		Handler: s.engine,
	}

	// 创建可取消的上下文监听系统信号
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 启动服务器
	go func() {
		log.Printf("服务器启动在 %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待关闭信号
	<-ctx.Done()
	log.Println("收到关闭信号，正在优雅关闭...")

	// 给 5 秒时间处理完当前请求
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("服务器关闭失败: %w", err)
	}

	// 关闭数据库连接
	if sqlDB, err := s.db.DB(); err == nil {
		sqlDB.Close()
	}

	log.Println("服务器已关闭")
	return nil
}
