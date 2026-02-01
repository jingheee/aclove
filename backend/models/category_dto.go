package models

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"io.lazydoge/aclove/jsonutil"
	"io.lazydoge/aclove/models/query"
)

var node *snowflake.Node

func init() {
	var err error
	node, err = snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
}

func GenerateSnowflakeID() int64 {
	return node.Generate().Int64()
}

type Category = query.CategoryDO

type CreateCategoryRequest struct {
	Name        string          `json:"name" binding:"required,min=1,max=255"`
	Description string          `json:"description"`
	ParentID    *jsonutil.Int64 `json:"parent_id,omitempty"`
	Position    int             `json:"position"`
}

type UpdateCategoryRequest struct {
	Name        *string         `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description *string         `json:"description,omitempty"`
	ParentID    *jsonutil.Int64 `json:"parent_id,omitempty"`
	Position    *int            `json:"position,omitempty"`
}

type CategoryResponse struct {
	ID          jsonutil.Int64      `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	ParentID    *jsonutil.Int64     `json:"parent_id,omitempty"`
	Position    int                 `json:"position"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	Children    []*CategoryResponse `json:"children,omitempty"`
}

func CategoryToResponse(c *Category) *CategoryResponse {
	var parentID *jsonutil.Int64
	if c.ParentID != nil {
		id := jsonutil.Int64(*c.ParentID)
		parentID = &id
	}
	return &CategoryResponse{
		ID:          jsonutil.Int64(c.ID),
		Name:        c.Name,
		Description: c.Description,
		ParentID:    parentID,
		Position:    c.Position,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func CategoryToResponseWithChildren(c *Category, children []*Category) *CategoryResponse {
	resp := CategoryToResponse(c)
	if len(children) > 0 {
		resp.Children = make([]*CategoryResponse, len(children))
		for i, child := range children {
			resp.Children[i] = CategoryToResponse(child)
		}
	}
	return resp
}

func CategoriesToResponseWithChildren(categories []*Category, childrenMap map[int64][]*Category) []*CategoryResponse {
	if len(categories) == 0 {
		return nil
	}
	result := make([]*CategoryResponse, len(categories))
	for i, category := range categories {
		var children []*Category
		if childrenMap != nil {
			children = childrenMap[category.ID]
		}
		result[i] = CategoryToResponseWithChildren(category, children)
	}
	return result
}

// CategoriesToResponse 将分类列表转换为响应列表
func CategoriesToResponse(categories []*Category) []*CategoryResponse {
	if len(categories) == 0 {
		return nil
	}
	result := make([]*CategoryResponse, len(categories))
	for i, category := range categories {
		result[i] = CategoryToResponse(category)
	}
	return result
}
