package main

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"time"

	"github.com/zzhtl/go-mountain/internal/config"
	"github.com/zzhtl/go-mountain/internal/db"
)

// generateRandomPassword 生成随机密码
func generateRandomPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	b := make([]byte, length)
	for i := range b {
		randomByte := make([]byte, 1)
		rand.Read(randomByte)
		b[i] = charset[randomByte[0]%byte(len(charset))]
	}
	return string(b)
}

// hashPassword 对密码进行哈希处理
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", hash)
}

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 初始化数据库
	database, err := db.Init(cfg.Database)
	if err != nil {
		log.Fatal("初始化数据库失败:", err)
	}
	defer database.Close()

	// 创建后台用户表
	migration := `CREATE TABLE IF NOT EXISTS backend_users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		role TEXT DEFAULT 'editor',
		status INTEGER DEFAULT 1,
		created_at DATETIME,
		updated_at DATETIME
	)`
	if _, err := database.Exec(database.Rebind(migration)); err != nil {
		log.Fatal("创建表失败:", err)
	}

	// 检查是否已有管理员用户
	var count int
	if err := database.Get(&count, database.Rebind("SELECT COUNT(*) FROM backend_users WHERE role = 'admin'")); err != nil {
		log.Fatal("检查管理员用户失败:", err)
	}

	if count > 0 {
		fmt.Println("管理员用户已存在，无需创建")
		return
	}

	// 生成管理员用户
	adminUsername := "admin"
	adminEmail := "admin@example.com"
	adminPassword := generateRandomPassword()
	hashedPassword := hashPassword(adminPassword)

	now := time.Now()
	query := database.Rebind(`
		INSERT INTO backend_users (username, email, password, role, status, created_at, updated_at) 
		VALUES (?, ?, ?, 'admin', 1, ?, ?)
	`)

	_, err = database.Exec(query, adminUsername, adminEmail, hashedPassword, now, now)
	if err != nil {
		log.Fatal("创建管理员用户失败:", err)
	}

	fmt.Println("============================================")
	fmt.Println("✅ 管理员用户创建成功！")
	fmt.Println("============================================")
	fmt.Printf("用户名: %s\n", adminUsername)
	fmt.Printf("邮箱: %s\n", adminEmail)
	fmt.Printf("密码: %s\n", adminPassword)
	fmt.Println("============================================")
	fmt.Println("⚠️  请妥善保存上述信息，并在首次登录后修改密码！")
	fmt.Println("============================================")
}
