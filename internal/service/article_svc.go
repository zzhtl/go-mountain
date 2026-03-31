package service

import (
	"context"

	"gorm.io/gorm"

	"github.com/zzhtl/go-mountain/internal/model"
	"github.com/zzhtl/go-mountain/internal/pkg/errcode"
	"github.com/zzhtl/go-mountain/internal/repository"
)

// ArticleService 文章管理服务
type ArticleService struct {
	repo *repository.BaseRepo[model.Article]
	db   *gorm.DB
}

// NewArticleService 创建文章服务
func NewArticleService(db *gorm.DB) *ArticleService {
	return &ArticleService{
		repo: repository.NewBaseRepo[model.Article](db),
		db:   db,
	}
}

// ArticleListItem 文章列表项（含栏目名）
type ArticleListItem struct {
	model.Article
	ColumnName string `json:"column_name"`
}

// List 获取文章列表（后台管理）
func (s *ArticleService) List(ctx context.Context, page, pageSize int, columnID int64, status int) ([]ArticleListItem, int64, error) {
	var (
		list  []ArticleListItem
		total int64
	)

	db := s.db.WithContext(ctx).Table("articles").
		Select("articles.*, columns.name as column_name").
		Joins("LEFT JOIN columns ON articles.column_id = columns.id").
		Where("articles.deleted_at IS NULL")

	if columnID > 0 {
		db = db.Where("articles.column_id = ?", columnID)
	}
	if status >= 0 {
		db = db.Where("articles.status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Order("articles.created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// Get 获取文章详情（后台）
func (s *ArticleService) Get(ctx context.Context, id int64) (*ArticleListItem, error) {
	var item ArticleListItem
	err := s.db.WithContext(ctx).Table("articles").
		Select("articles.*, columns.name as column_name").
		Joins("LEFT JOIN columns ON articles.column_id = columns.id").
		Where("articles.id = ? AND articles.deleted_at IS NULL", id).
		First(&item).Error
	if err != nil {
		return nil, errcode.ErrNotFound
	}
	return &item, nil
}

// Create 创建文章
func (s *ArticleService) Create(ctx context.Context, article *model.Article) error {
	return s.repo.Create(ctx, article)
}

// Update 更新文章
func (s *ArticleService) Update(ctx context.Context, id int64, updates map[string]any) error {
	return s.repo.Update(ctx, id, updates)
}

// UpdateStatus 更新文章状态
func (s *ArticleService) UpdateStatus(ctx context.Context, id int64, status int) error {
	return s.repo.Update(ctx, id, map[string]any{"status": status})
}

// Delete 删除文章
func (s *ArticleService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// ListByColumn 根据栏目获取已发布文章（小程序端）
func (s *ArticleService) ListByColumn(ctx context.Context, columnID int64, page, pageSize int) ([]model.Article, int64, error) {
	scope := func(db *gorm.DB) *gorm.DB {
		return db.Where("column_id = ? AND status = 1", columnID).Order("created_at DESC")
	}
	return s.repo.List(ctx, page, pageSize, scope)
}

// GetForMP 获取文章详情（小程序端，增加浏览量）
func (s *ArticleService) GetForMP(ctx context.Context, id int64) (*ArticleListItem, error) {
	// 增加浏览量
	s.db.WithContext(ctx).Model(&model.Article{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1"))

	var item ArticleListItem
	err := s.db.WithContext(ctx).Table("articles").
		Select("articles.*, columns.name as column_name").
		Joins("LEFT JOIN columns ON articles.column_id = columns.id").
		Where("articles.id = ? AND articles.status = 1 AND articles.deleted_at IS NULL", id).
		First(&item).Error
	if err != nil {
		return nil, errcode.ErrNotFound
	}
	return &item, nil
}
