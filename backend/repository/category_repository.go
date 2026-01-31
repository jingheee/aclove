package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"io.lazydoge/aclove/models/query"
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

func (r *CategoryRepository) Create(ctx context.Context, category *query.CategoryDO) error {
	result := r.db.WithContext(ctx).Create(category)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (*query.CategoryDO, error) {
	var category query.CategoryDO
	result := r.db.WithContext(ctx).First(&category, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, result.Error
	}
	return &category, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *query.CategoryDO) error {
	result := r.db.WithContext(ctx).Save(category)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&query.CategoryDO{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func (r *CategoryRepository) List(ctx context.Context, offset, limit int) ([]*query.CategoryDO, error) {
	var categories []*query.CategoryDO
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

func (r *CategoryRepository) GetByParentID(ctx context.Context, parentID *int64) ([]*query.CategoryDO, error) {
	var categories []*query.CategoryDO
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
	result := r.db.WithContext(ctx).Model(&query.CategoryDO{}).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}

func (r *CategoryRepository) UpdatePostsCategoryID(ctx context.Context, oldCategoryID, newCategoryID int64) error {
	result := r.db.WithContext(ctx).
		Model(&query.PostDO{}).
		Where("category_id = ?", oldCategoryID).
		Update("category_id", newCategoryID)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *CategoryRepository) GetFirstAvailableCategory(ctx context.Context) (*query.CategoryDO, error) {
	var category query.CategoryDO
	result := r.db.WithContext(ctx).
		Order("position ASC, created_at ASC").
		First(&category)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, result.Error
	}
	return &category, nil
}
