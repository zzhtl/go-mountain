package service

import (
	"context"
	"encoding/json"
	"time"

	"gorm.io/gorm"

	"github.com/zzhtl/go-mountain/internal/model"
	"github.com/zzhtl/go-mountain/internal/pkg/errcode"
	"github.com/zzhtl/go-mountain/internal/repository"
)

// RegistrationService 报名服务
type RegistrationService struct {
	repo *repository.BaseRepo[model.Registration]
	db   *gorm.DB
}

// NewRegistrationService 创建报名服务
func NewRegistrationService(db *gorm.DB) *RegistrationService {
	return &RegistrationService{
		repo: repository.NewBaseRepo[model.Registration](db),
		db:   db,
	}
}

// RegistrationListItem 报名列表项（含活动和用户信息）
type RegistrationListItem struct {
	model.Registration
	ActivityTitle string `json:"activity_title"`
	UserName      string `json:"user_name"`
	UserPhone     string `json:"user_phone"`
}

// List 获取报名列表（后台管理）
func (s *RegistrationService) List(ctx context.Context, page, pageSize int, activityID int64, status int) ([]RegistrationListItem, int64, error) {
	var (
		list  []RegistrationListItem
		total int64
	)

	db := s.db.WithContext(ctx).Table("registrations").
		Select("registrations.*, activities.title as activity_title").
		Joins("LEFT JOIN activities ON registrations.activity_id = activities.id").
		Where("registrations.deleted_at IS NULL")

	if activityID > 0 {
		db = db.Where("registrations.activity_id = ?", activityID)
	}
	if status >= 0 {
		db = db.Where("registrations.status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Order("registrations.created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// Get 获取报名详情
func (s *RegistrationService) Get(ctx context.Context, id int64) (*RegistrationListItem, error) {
	var item RegistrationListItem
	err := s.db.WithContext(ctx).Table("registrations").
		Select("registrations.*, activities.title as activity_title").
		Joins("LEFT JOIN activities ON registrations.activity_id = activities.id").
		Where("registrations.id = ? AND registrations.deleted_at IS NULL", id).
		First(&item).Error
	if err != nil {
		return nil, errcode.ErrRegistrationNotFound
	}
	return &item, nil
}

// CreateRegistrationRequest 创建报名请求
type CreateRegistrationRequest struct {
	ActivityID int64           `json:"activity_id" binding:"required"`
	Name       string          `json:"name" binding:"required"`
	Phone      string          `json:"phone" binding:"required"`
	IDCard     string          `json:"id_card"`
	ExtraInfo  json.RawMessage `json:"extra_info"`
}

// Create 创建报名（小程序端调用）
func (s *RegistrationService) Create(ctx context.Context, userID int64, req *CreateRegistrationRequest) (*model.Registration, error) {
	// 查询活动
	var activity model.Activity
	if err := s.db.WithContext(ctx).First(&activity, req.ActivityID).Error; err != nil {
		return nil, errcode.ErrNotFound
	}

	// 校验活动状态
	if activity.Status != 1 {
		return nil, errcode.ErrActivityNotOpen
	}

	// 校验报名时间
	now := time.Now()
	if activity.RegStartTime != nil && now.Before(*activity.RegStartTime) {
		return nil, errcode.ErrActivityNotOpen
	}
	if activity.RegEndTime != nil && now.After(*activity.RegEndTime) {
		return nil, errcode.ErrActivityNotOpen
	}

	// 校验人数上限
	if activity.MaxParticipants > 0 {
		var count int64
		s.db.WithContext(ctx).Model(&model.Registration{}).
			Where("activity_id = ? AND status IN (0,1) AND deleted_at IS NULL", req.ActivityID).
			Count(&count)
		if count >= int64(activity.MaxParticipants) {
			return nil, errcode.ErrActivityFull
		}
	}

	// 校验是否重复报名
	var existCount int64
	s.db.WithContext(ctx).Model(&model.Registration{}).
		Where("activity_id = ? AND user_id = ? AND status IN (0,1) AND deleted_at IS NULL", req.ActivityID, userID).
		Count(&existCount)
	if existCount > 0 {
		return nil, errcode.ErrAlreadyRegistered
	}

	// 创建报名记录
	reg := &model.Registration{
		ActivityID: req.ActivityID,
		UserID:     userID,
		Name:       req.Name,
		Phone:      req.Phone,
		IDCard:     req.IDCard,
		ExtraInfo:  req.ExtraInfo,
		Status:     0, // 待支付（免费活动也先设为0，后面由支付流程或直接确认）
	}

	// 免费活动直接确认报名
	if activity.Price == 0 {
		reg.Status = 1 // 已确认
	}

	if err := s.repo.Create(ctx, reg); err != nil {
		return nil, err
	}

	return reg, nil
}

// Cancel 取消报名
func (s *RegistrationService) Cancel(ctx context.Context, id int64, userID int64) error {
	var reg model.Registration
	if err := s.db.WithContext(ctx).First(&reg, id).Error; err != nil {
		return errcode.ErrRegistrationNotFound
	}

	// 校验是否是本人的报名
	if reg.UserID != userID {
		return errcode.ErrForbidden
	}

	if reg.Status == 2 {
		return errcode.ErrRegistrationCancelled
	}

	return s.repo.Update(ctx, id, map[string]any{"status": 2})
}

// GetByUser 获取用户的报名列表（小程序端）
func (s *RegistrationService) GetByUser(ctx context.Context, userID int64, page, pageSize int) ([]RegistrationListItem, int64, error) {
	var (
		list  []RegistrationListItem
		total int64
	)

	db := s.db.WithContext(ctx).Table("registrations").
		Select("registrations.*, activities.title as activity_title").
		Joins("LEFT JOIN activities ON registrations.activity_id = activities.id").
		Where("registrations.user_id = ? AND registrations.deleted_at IS NULL", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Order("registrations.created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}
