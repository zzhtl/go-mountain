package crypto

import (
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 10

// HashPassword 使用 bcrypt 加密密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败: %w", err)
	}
	return string(bytes), nil
}

// VerifyPassword 验证密码，支持渐进式迁移
// passwordVersion: 1=SHA256(旧), 2=bcrypt(新)
func VerifyPassword(password, hashedPassword string, passwordVersion int) bool {
	switch passwordVersion {
	case 1:
		return verifySHA256(password, hashedPassword)
	default:
		return verifyBcrypt(password, hashedPassword)
	}
}

// NeedsMigration 检查密码是否需要迁移到 bcrypt
func NeedsMigration(passwordVersion int) bool {
	return passwordVersion == 1
}

// verifySHA256 验证 SHA256 密码（旧方式，兼容用）
func verifySHA256(password, hashedPassword string) bool {
	hash := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", hash) == hashedPassword
}

// verifyBcrypt 验证 bcrypt 密码
func verifyBcrypt(password, hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
