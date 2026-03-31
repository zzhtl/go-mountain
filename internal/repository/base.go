package repository

import (
	"context"

	"gorm.io/gorm"
)

// BaseRepo 泛型基础 Repository，提供通用 CRUD 操作
type BaseRepo[T any] struct {
	DB *gorm.DB
}

// NewBaseRepo 创建基础 Repository
func NewBaseRepo[T any](db *gorm.DB) *BaseRepo[T] {
	return &BaseRepo[T]{DB: db}
}

// Create 创建记录
func (r *BaseRepo[T]) Create(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).Create(entity).Error
}

// GetByID 根据 ID 查询
func (r *BaseRepo[T]) GetByID(ctx context.Context, id int64) (*T, error) {
	var entity T
	if err := r.DB.WithContext(ctx).First(&entity, id).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

// Update 更新记录（仅更新非零值字段）
func (r *BaseRepo[T]) Update(ctx context.Context, id int64, updates map[string]any) error {
	var entity T
	return r.DB.WithContext(ctx).Model(&entity).Where("id = ?", id).Updates(updates).Error
}

// Save 保存记录（全量更新）
func (r *BaseRepo[T]) Save(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).Save(entity).Error
}

// Delete 软删除
func (r *BaseRepo[T]) Delete(ctx context.Context, id int64) error {
	var entity T
	return r.DB.WithContext(ctx).Delete(&entity, id).Error
}

// List 分页查询
func (r *BaseRepo[T]) List(ctx context.Context, page, pageSize int, scopes ...func(*gorm.DB) *gorm.DB) ([]T, int64, error) {
	var (
		list  []T
		total int64
	)

	db := r.DB.WithContext(ctx).Model(new(T))
	for _, scope := range scopes {
		db = scope(db)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// FindAll 查询所有（不分页）
func (r *BaseRepo[T]) FindAll(ctx context.Context, scopes ...func(*gorm.DB) *gorm.DB) ([]T, error) {
	var list []T
	db := r.DB.WithContext(ctx).Model(new(T))
	for _, scope := range scopes {
		db = scope(db)
	}
	if err := db.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// Count 统计数量
func (r *BaseRepo[T]) Count(ctx context.Context, scopes ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	db := r.DB.WithContext(ctx).Model(new(T))
	for _, scope := range scopes {
		db = scope(db)
	}
	if err := db.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// Exists 检查记录是否存在
func (r *BaseRepo[T]) Exists(ctx context.Context, query string, args ...any) (bool, error) {
	var count int64
	if err := r.DB.WithContext(ctx).Model(new(T)).Where(query, args...).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
