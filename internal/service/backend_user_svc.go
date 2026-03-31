package service

import (
	"context"

	"gorm.io/gorm"

	"github.com/zzhtl/go-mountain/internal/model"
	"github.com/zzhtl/go-mountain/internal/pkg/crypto"
	"github.com/zzhtl/go-mountain/internal/pkg/errcode"
	"github.com/zzhtl/go-mountain/internal/repository"
)

// BackendUserService 后台用户管理服务
type BackendUserService struct {
	repo *repository.BaseRepo[model.BackendUser]
	db   *gorm.DB
}

// NewBackendUserService 创建后台用户管理服务
func NewBackendUserService(db *gorm.DB) *BackendUserService {
	return &BackendUserService{
		repo: repository.NewBaseRepo[model.BackendUser](db),
		db:   db,
	}
}

// BackendUserListItem 后台用户列表项（含角色信息）
type BackendUserListItem struct {
	model.BackendUser
	RoleName    string `json:"role_name"`
	RoleDisplay string `json:"role_display"`
}

// List 获取后台用户列表
func (s *BackendUserService) List(ctx context.Context, page, pageSize int) ([]BackendUserListItem, int64, error) {
	var (
		list  []BackendUserListItem
		total int64
	)

	db := s.db.WithContext(ctx).Model(&model.BackendUser{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := s.db.WithContext(ctx).
		Table("backend_users").
		Select("backend_users.*, roles.name as role_name, roles.display_name as role_display").
		Joins("LEFT JOIN roles ON backend_users.role_id = roles.id").
		Where("backend_users.deleted_at IS NULL").
		Order("backend_users.created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&list).Error

	return list, total, err
}

// Get 获取单个后台用户
func (s *BackendUserService) Get(ctx context.Context, id int64) (*BackendUserListItem, error) {
	var item BackendUserListItem
	err := s.db.WithContext(ctx).
		Table("backend_users").
		Select("backend_users.*, roles.name as role_name, roles.display_name as role_display").
		Joins("LEFT JOIN roles ON backend_users.role_id = roles.id").
		Where("backend_users.id = ? AND backend_users.deleted_at IS NULL", id).
		First(&item).Error
	if err != nil {
		return nil, errcode.ErrNotFound
	}
	return &item, nil
}

// Create 创建后台用户，返回明文密码
func (s *BackendUserService) Create(ctx context.Context, username, email string, roleID int64) (*model.BackendUser, string, error) {
	// 验证角色是否存在
	var roleCount int64
	s.db.WithContext(ctx).Model(&model.Role{}).Where("id = ? AND status = 1", roleID).Count(&roleCount)
	if roleCount == 0 {
		return nil, "", errcode.ErrInvalidParam
	}

	password := GenerateRandomPassword(8)
	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	user := &model.BackendUser{
		Username:        username,
		Email:           email,
		Password:        hashedPassword,
		RoleID:          roleID,
		PasswordVersion: 2,
		Status:          1,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, "", err
	}

	return user, password, nil
}

// Update 更新后台用户信息
func (s *BackendUserService) Update(ctx context.Context, id int64, username, email string, roleID int64) error {
	// 验证角色
	var roleCount int64
	s.db.WithContext(ctx).Model(&model.Role{}).Where("id = ? AND status = 1", roleID).Count(&roleCount)
	if roleCount == 0 {
		return errcode.ErrInvalidParam
	}

	return s.repo.Update(ctx, id, map[string]any{
		"username": username,
		"email":    email,
		"role_id":  roleID,
	})
}

// UpdateStatus 更新用户状态
func (s *BackendUserService) UpdateStatus(ctx context.Context, id int64, status int) error {
	return s.repo.Update(ctx, id, map[string]any{"status": status})
}

// ResetPassword 重置用户密码，返回新的明文密码
func (s *BackendUserService) ResetPassword(ctx context.Context, id int64) (string, error) {
	password := GenerateRandomPassword(8)
	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		return "", err
	}

	err = s.repo.Update(ctx, id, map[string]any{
		"password":         hashedPassword,
		"password_version": 2,
	})
	if err != nil {
		return "", err
	}

	return password, nil
}

// Delete 删除后台用户
func (s *BackendUserService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// GetCurrentUserMenus 获取当前用户的菜单权限树
func (s *BackendUserService) GetCurrentUserMenus(ctx context.Context, userID int64) ([]*model.Menu, error) {
	var user model.BackendUser
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, errcode.ErrNotFound
	}

	// 检查是否是 admin 角色（拥有所有菜单）
	var role model.Role
	s.db.WithContext(ctx).First(&role, user.RoleID)

	var menus []model.Menu
	if role.Name == "admin" {
		s.db.WithContext(ctx).Where("status = 1").Order("sort, id").Find(&menus)
	} else {
		s.db.WithContext(ctx).
			Joins("INNER JOIN role_menus ON menus.id = role_menus.menu_id").
			Where("role_menus.role_id = ? AND menus.status = 1", user.RoleID).
			Order("menus.sort, menus.id").
			Find(&menus)
	}

	return buildMenuTree(menus, 0), nil
}

// buildMenuTree 构建菜单树
func buildMenuTree(menus []model.Menu, parentID int64) []*model.Menu {
	var tree []*model.Menu
	for i := range menus {
		if menus[i].ParentID == parentID {
			menu := &menus[i]
			menu.Children = buildMenuTree(menus, menu.ID)
			tree = append(tree, menu)
		}
	}
	return tree
}
