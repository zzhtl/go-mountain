package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"

	"github.com/zzhtl/go-mountain/internal/model"
	"github.com/zzhtl/go-mountain/internal/pkg/crypto"
	"github.com/zzhtl/go-mountain/internal/pkg/errcode"
)

// AuthService 认证服务
type AuthService struct {
	db        *gorm.DB
	jwtSecret string
}

// NewAuthService 创建认证服务
func NewAuthService(db *gorm.DB, jwtSecret string) *AuthService {
	return &AuthService{db: db, jwtSecret: jwtSecret}
}

// LoginResult 登录结果
type LoginResult struct {
	Token string         `json:"token"`
	User  LoginUserInfo  `json:"user"`
}

// LoginUserInfo 登录用户信息
type LoginUserInfo struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	RoleID      int64  `json:"role_id"`
	RoleName    string `json:"role"`
	RoleDisplay string `json:"role_display"`
}

// Login 后台用户登录
func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	var user model.BackendUser
	if err := s.db.WithContext(ctx).Preload("Role").
		Where("username = ?", username).First(&user).Error; err != nil {
		return nil, errcode.ErrInvalidPassword
	}

	if user.Status != 1 {
		return nil, errcode.ErrAccountDisabled
	}

	if !crypto.VerifyPassword(password, user.Password, user.PasswordVersion) {
		return nil, errcode.ErrInvalidPassword
	}

	// 渐进式密码迁移：SHA256 → bcrypt
	if crypto.NeedsMigration(user.PasswordVersion) {
		if newHash, err := crypto.HashPassword(password); err == nil {
			s.db.WithContext(ctx).Model(&user).Updates(map[string]any{
				"password":         newHash,
				"password_version": 2,
			})
		}
	}

	// 更新最后登录时间
	now := time.Now()
	s.db.WithContext(ctx).Model(&user).Update("last_login", now)

	// 生成 JWT
	token, err := s.generateToken(&user)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}

	roleName := ""
	roleDisplay := ""
	if user.Role != nil {
		roleName = user.Role.Name
		roleDisplay = user.Role.DisplayName
	}

	return &LoginResult{
		Token: token,
		User: LoginUserInfo{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			RoleID:      user.RoleID,
			RoleName:    roleName,
			RoleDisplay: roleDisplay,
		},
	}, nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	var user model.BackendUser
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return errcode.ErrNotFound
	}

	if !crypto.VerifyPassword(oldPassword, user.Password, user.PasswordVersion) {
		return fmt.Errorf("原密码错误")
	}

	newHash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.db.WithContext(ctx).Model(&user).Updates(map[string]any{
		"password":         newHash,
		"password_version": 2,
	}).Error
}

// generateToken 生成 JWT 令牌
func (s *AuthService) generateToken(user *model.BackendUser) (string, error) {
	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}

	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role_id":  user.RoleID,
		"role":     roleName,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// GenerateRandomPassword 生成随机密码
func GenerateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		randomByte := make([]byte, 1)
		rand.Read(randomByte)
		b[i] = charset[randomByte[0]%byte(len(charset))]
	}
	return string(b)
}
