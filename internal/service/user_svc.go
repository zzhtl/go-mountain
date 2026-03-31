package service

import (
	"context"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram"
	"gorm.io/gorm"

	"github.com/zzhtl/go-mountain/internal/model"
	"github.com/zzhtl/go-mountain/internal/pkg/errcode"
	"github.com/zzhtl/go-mountain/internal/repository"
)

// UserService 小程序用户管理服务
type UserService struct {
	repo      *repository.BaseRepo[model.User]
	db        *gorm.DB
	appID     string
	appSecret string
}

// NewUserService 创建小程序用户服务
func NewUserService(db *gorm.DB, appID, appSecret string) *UserService {
	return &UserService{
		repo:      repository.NewBaseRepo[model.User](db),
		db:        db,
		appID:     appID,
		appSecret: appSecret,
	}
}

// WechatLogin 微信小程序登录
func (s *UserService) WechatLogin(ctx context.Context, code string) (*model.User, error) {
	mp, err := miniProgram.NewMiniProgram(&miniProgram.UserConfig{
		AppID:     s.appID,
		Secret:    s.appSecret,
		HttpDebug: false,
		Debug:     false,
	})
	if err != nil {
		return nil, err
	}

	session, err := mp.Auth.Session(ctx, code)
	if err != nil {
		return nil, err
	}

	openID := session.OpenID

	// 查询或创建用户
	var user model.User
	err = s.db.WithContext(ctx).Where("open_id = ?", openID).First(&user).Error
	if err != nil {
		user = model.User{OpenID: openID, Status: 1}
		if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
			return nil, err
		}
	}

	return &user, nil
}

// Register 注册/绑定手机号
func (s *UserService) Register(ctx context.Context, phone, openID, name string) (*model.User, error) {
	var user model.User

	// 通过 openid 查找用户
	err := s.db.WithContext(ctx).Where("open_id = ?", openID).First(&user).Error
	if err != nil {
		// 不存在则创建
		user = model.User{
			OpenID: openID,
			Phone:  phone,
			Name:   name,
			Status: 1,
		}
		if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	}

	// 已存在则更新
	s.db.WithContext(ctx).Model(&user).Updates(map[string]any{
		"phone": phone,
		"name":  name,
	})

	return &user, nil
}

// List 获取小程序用户列表（后台）
func (s *UserService) List(ctx context.Context, page, pageSize int) ([]model.User, int64, error) {
	scope := func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}
	return s.repo.List(ctx, page, pageSize, scope)
}

// Get 获取用户详情
func (s *UserService) Get(ctx context.Context, id int64) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errcode.ErrNotFound
	}
	return user, nil
}

// Update 更新用户信息
func (s *UserService) Update(ctx context.Context, id int64, updates map[string]any) error {
	return s.repo.Update(ctx, id, updates)
}

// Delete 删除用户
func (s *UserService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
