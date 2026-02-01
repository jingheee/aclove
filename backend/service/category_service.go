package service

import (
	"context"
	"errors"

	"io.lazydoge/aclove/jsonutil"
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
	var parentID *int64
	if req.ParentID != nil {
		pid := int64(*req.ParentID)
		parentID = &pid
	}

	category := &models.Category{
		ID:          models.GenerateSnowflakeID(),
		Name:        req.Name,
		Description: req.Description,
		ParentID:    parentID,
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
		pid := int64(*req.ParentID)
		if err := s.ValidateParentID(ctx, pid); err != nil {
			return nil, err
		}
		category.ParentID = &pid
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
	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	targetCategoryID := int64(0)
	if category.ParentID != nil {
		targetCategoryID = *category.ParentID
	} else {
		firstCategory, err := s.repo.GetFirstAvailableCategory(ctx)
		if err != nil {
			return err
		}
		targetCategoryID = firstCategory.ID
	}

	if targetCategoryID == id {
		return errors.New("无法删除唯一的分类，请先创建其他分类")
	}

	if err := s.repo.UpdatePostsCategoryID(ctx, id, targetCategoryID); err != nil {
		return err
	}

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

func (s *CategoryService) GetCategoryTree(ctx context.Context) ([]*models.CategoryResponse, error) {
	allCategories, err := s.repo.List(ctx, 0, -1)
	if err != nil {
		return nil, err
	}

	if len(allCategories) == 0 {
		return []*models.CategoryResponse{}, nil
	}

	childrenMap := make(map[int64][]*models.Category)
	for _, category := range allCategories {
		if category.ParentID != nil {
			childrenMap[*category.ParentID] = append(childrenMap[*category.ParentID], category)
		}
	}

	rootCategories := make([]*models.Category, 0, len(allCategories))
	for _, category := range allCategories {
		if category.ParentID == nil {
			rootCategories = append(rootCategories, category)
		}
	}

	return models.CategoriesToResponseWithChildren(rootCategories, childrenMap), nil
}

func (s *CategoryService) ValidateParentID(ctx context.Context, parentID int64) error {
	if parentID == 0 {
		return nil
	}
	_, err := s.repo.GetByID(ctx, parentID)
	if err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			return errors.New("父分类不存在")
		}
		return err
	}
	return nil
}

// Int64PtrToJSONInt64 将 *int64 转换为 *jsonutil.Int64
func Int64PtrToJSONInt64(ptr *int64) *jsonutil.Int64 {
	if ptr == nil {
		return nil
	}
	v := jsonutil.Int64(*ptr)
	return &v
}

// JSONInt64PtrToInt64 将 *jsonutil.Int64 转换为 *int64
func JSONInt64PtrToInt64(ptr *jsonutil.Int64) *int64 {
	if ptr == nil {
		return nil
	}
	v := int64(*ptr)
	return &v
}
