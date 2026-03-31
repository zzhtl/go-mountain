package service

import (
	"context"

	"gorm.io/gorm"

	"github.com/zzhtl/go-mountain/internal/model"
	"github.com/zzhtl/go-mountain/internal/pkg/errcode"
	"github.com/zzhtl/go-mountain/internal/repository"
)

// RoleService 角色管理服务
type RoleService struct {
	repo *repository.BaseRepo[model.Role]
	db   *gorm.DB
}

// NewRoleService 创建角色服务
func NewRoleService(db *gorm.DB) *RoleService {
	return &RoleService{
		repo: repository.NewBaseRepo[model.Role](db),
		db:   db,
	}
}

// List 获取角色列表
func (s *RoleService) List(ctx context.Context, page, pageSize int) ([]model.Role, int64, error) {
	scope := func(db *gorm.DB) *gorm.DB {
		return db.Order("id")
	}
	return s.repo.List(ctx, page, pageSize, scope)
}

// Get 获取单个角色
func (s *RoleService) Get(ctx context.Context, id int64) (*model.Role, error) {
	return s.repo.GetByID(ctx, id)
}

// Create 创建角色
func (s *RoleService) Create(ctx context.Context, role *model.Role) error {
	role.Status = 1
	return s.repo.Create(ctx, role)
}

// Update 更新角色
func (s *RoleService) Update(ctx context.Context, id int64, updates map[string]any) error {
	return s.repo.Update(ctx, id, updates)
}

// UpdateStatus 更新角色状态
func (s *RoleService) UpdateStatus(ctx context.Context, id int64, status int) error {
	return s.repo.Update(ctx, id, map[string]any{"status": status})
}

// Delete 删除角色
func (s *RoleService) Delete(ctx context.Context, id int64) error {
	// 检查是否有用户使用
	exists, _ := repository.NewBaseRepo[model.BackendUser](s.db).Exists(ctx, "role_id = ?", id)
	if exists {
		return errcode.ErrRoleInUse
	}

	// 删除角色菜单关联
	s.db.WithContext(ctx).Where("role_id = ?", id).Delete(&model.RoleMenu{})

	return s.repo.Delete(ctx, id)
}

// GetRoleMenus 获取角色的菜单 ID 列表
func (s *RoleService) GetRoleMenus(ctx context.Context, roleID int64) ([]int64, error) {
	var menuIDs []int64
	err := s.db.WithContext(ctx).Model(&model.RoleMenu{}).
		Where("role_id = ?", roleID).
		Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

// UpdateRoleMenus 更新角色的菜单权限
func (s *RoleService) UpdateRoleMenus(ctx context.Context, roleID int64, menuIDs []int64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除现有关联
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RoleMenu{}).Error; err != nil {
			return err
		}
		// 添加新关联
		for _, menuID := range menuIDs {
			rm := model.RoleMenu{RoleID: roleID, MenuID: menuID}
			if err := tx.Create(&rm).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// InitDefaultRoles 初始化默认角色
func (s *RoleService) InitDefaultRoles(ctx context.Context) error {
	var count int64
	s.db.WithContext(ctx).Model(&model.Role{}).Count(&count)
	if count > 0 {
		return nil
	}

	roles := []model.Role{
		{Name: "admin", DisplayName: "超级管理员", Description: "拥有所有权限", Status: 1},
		{Name: "editor", DisplayName: "编辑员", Description: "文章编辑权限", Status: 1},
		{Name: "viewer", DisplayName: "查看员", Description: "只读权限", Status: 1},
	}

	for i := range roles {
		if err := s.db.WithContext(ctx).Create(&roles[i]).Error; err != nil {
			return err
		}
	}
	return nil
}
