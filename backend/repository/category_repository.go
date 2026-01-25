package repository

import (
	"context"
	"errors"

	"io.lazydoge/aclove/models"

	"gorm.io/gorm"
)

var (
	ErrCategoryNotFound = errors.New("分类不存在")
	ErrInvalidInput     = errors.New("无效的输入参数")
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, category *models.Category) error {
	result := r.db.WithContext(ctx).Create(category)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (*models.Category, error) {
	var category models.Category
	result := r.db.WithContext(ctx).First(&category, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, result.Error
	}
	return &category, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *models.Category) error {
	result := r.db.WithContext(ctx).Save(category)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&models.Category{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func (r *CategoryRepository) List(ctx context.Context, offset, limit int) ([]*models.Category, error) {
	var categories []*models.Category
	result := r.db.WithContext(ctx).
		Order("position ASC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&categories)
	if result.Error != nil {
		return nil, result.Error
	}
	return categories, nil
}

func (r *CategoryRepository) GetByParentID(ctx context.Context, parentID *int64) ([]*models.Category, error) {
	var categories []*models.Category
	query := r.db.WithContext(ctx).Order("position ASC, created_at DESC")
	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}
	result := query.Find(&categories)
	if result.Error != nil {
		return nil, result.Error
	}
	return categories, nil
}

func (r *CategoryRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&models.Category{}).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}
