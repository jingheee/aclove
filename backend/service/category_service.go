package service

import (
	"context"
	"errors"

	"io.lazydoge/aclove/models"
	"io.lazydoge/aclove/repository"
)

type CategoryService struct {
	repo *repository.CategoryRepository
}

func NewCategoryService(repo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) CreateCategory(ctx context.Context, req *models.CreateCategoryRequest) (*models.Category, error) {
	category := &models.Category{
		ID:          models.GenerateSnowflakeID(),
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		Position:    req.Position,
	}
	if err := s.repo.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *CategoryService) GetCategory(ctx context.Context, id int64) (*models.Category, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CategoryService) UpdateCategory(ctx context.Context, id int64, req *models.UpdateCategoryRequest) (*models.Category, error) {
	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.ParentID != nil {
		if err := s.ValidateParentID(ctx, req.ParentID); err != nil {
			return nil, err
		}
		category.ParentID = req.ParentID
	}

	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Description != nil {
		category.Description = *req.Description
	}
	if req.Position != nil {
		category.Position = *req.Position
	}

	if err := s.repo.Update(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *CategoryService) ListCategories(ctx context.Context, page, pageSize int) ([]*models.Category, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	categories, err := s.repo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (s *CategoryService) GetChildren(ctx context.Context, parentID *int64) ([]*models.Category, error) {
	return s.repo.GetByParentID(ctx, parentID)
}

func (s *CategoryService) ValidateParentID(ctx context.Context, parentID *int64) error {
	if parentID == nil {
		return nil
	}
	_, err := s.repo.GetByID(ctx, *parentID)
	if err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			return errors.New("父分类不存在")
		}
		return err
	}
	return nil
}
