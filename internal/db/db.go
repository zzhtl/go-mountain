package db

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zzhtl/go-mountain/internal/config"
)

// Init 根据配置初始化数据库连接
func Init(cfg config.DatabaseConfig) (*sqlx.DB, error) {
	var (
		driver = cfg.Driver
		dsn    = cfg.DSN
		db     *sqlx.DB
		err    error
	)

	switch driver {
	case "sqlite3":
		db, err = sqlx.Open("sqlite3", dsn)
	case "postgres":
		db, err = sqlx.Open("postgres", dsn)
	case "mysql":
		db, err = sqlx.Open("mysql", dsn)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	if err != nil {
		return nil, err
	}
	// 校验连接
	if err = db.Ping(); err != nil {
		return nil, err
	}
	// 可根据需要设置连接池等参数
	return db, nil
}
