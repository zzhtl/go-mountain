package service

import (
	"context"

	"gorm.io/gorm"

	"github.com/zzhtl/go-mountain/internal/model"
	"github.com/zzhtl/go-mountain/internal/pkg/errcode"
	"github.com/zzhtl/go-mountain/internal/repository"
)

// ActivityService 活动管理服务
type ActivityService struct {
	repo *repository.BaseRepo[model.Activity]
	db   *gorm.DB
}

// NewActivityService 创建活动服务
func NewActivityService(db *gorm.DB) *ActivityService {
	return &ActivityService{
		repo: repository.NewBaseRepo[model.Activity](db),
		db:   db,
	}
}

// ActivityListItem 活动列表项（含报名人数）
type ActivityListItem struct {
	model.Activity
	RegCount int64 `json:"reg_count"`
}

// List 获取活动列表（后台管理）
func (s *ActivityService) List(ctx context.Context, page, pageSize int, status int) ([]ActivityListItem, int64, error) {
	var (
		list  []ActivityListItem
		total int64
	)

	db := s.db.WithContext(ctx).Table("activities").
		Select("activities.*, (SELECT COUNT(*) FROM registrations WHERE registrations.activity_id = activities.id AND registrations.status IN (0,1) AND registrations.deleted_at IS NULL) as reg_count").
		Where("activities.deleted_at IS NULL")

	if status >= 0 {
		db = db.Where("activities.status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Order("activities.created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// Get 获取活动详情
func (s *ActivityService) Get(ctx context.Context, id int64) (*ActivityListItem, error) {
	var item ActivityListItem
	err := s.db.WithContext(ctx).Table("activities").
		Select("activities.*, (SELECT COUNT(*) FROM registrations WHERE registrations.activity_id = activities.id AND registrations.status IN (0,1) AND registrations.deleted_at IS NULL) as reg_count").
		Where("activities.id = ? AND activities.deleted_at IS NULL", id).
		First(&item).Error
	if err != nil {
		return nil, errcode.ErrNotFound
	}
	return &item, nil
}

// Create 创建活动
func (s *ActivityService) Create(ctx context.Context, activity *model.Activity) error {
	return s.repo.Create(ctx, activity)
}

// Update 更新活动
func (s *ActivityService) Update(ctx context.Context, id int64, updates map[string]any) error {
	return s.repo.Update(ctx, id, updates)
}

// UpdateStatus 更新活动状态
func (s *ActivityService) UpdateStatus(ctx context.Context, id int64, status int) error {
	return s.repo.Update(ctx, id, map[string]any{"status": status})
}

// Delete 删除活动
func (s *ActivityService) Delete(ctx context.Context, id int64) error {
	// 检查是否有有效报名
	var count int64
	s.db.WithContext(ctx).Model(&model.Registration{}).
		Where("activity_id = ? AND status IN (0,1) AND deleted_at IS NULL", id).
		Count(&count)
	if count > 0 {
		return errcode.ErrActivityHasRegistrations
	}
	return s.repo.Delete(ctx, id)
}

// GetForMP 获取活动详情（小程序端）
func (s *ActivityService) GetForMP(ctx context.Context, id int64) (*ActivityListItem, error) {
	var item ActivityListItem
	err := s.db.WithContext(ctx).Table("activities").
		Select("activities.*, (SELECT COUNT(*) FROM registrations WHERE registrations.activity_id = activities.id AND registrations.status IN (0,1) AND registrations.deleted_at IS NULL) as reg_count").
		Where("activities.id = ? AND activities.status IN (1,2,3) AND activities.deleted_at IS NULL", id).
		First(&item).Error
	if err != nil {
		return nil, errcode.ErrNotFound
	}
	return &item, nil
}

// ListForMP 获取活动列表（小程序端，只显示非草稿的活动）
func (s *ActivityService) ListForMP(ctx context.Context, page, pageSize int) ([]ActivityListItem, int64, error) {
	var (
		list  []ActivityListItem
		total int64
	)

	db := s.db.WithContext(ctx).Table("activities").
		Select("activities.*, (SELECT COUNT(*) FROM registrations WHERE registrations.activity_id = activities.id AND registrations.status IN (0,1) AND registrations.deleted_at IS NULL) as reg_count").
		Where("activities.status IN (1,2,3,4) AND activities.deleted_at IS NULL")

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Order("activities.start_time DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}
