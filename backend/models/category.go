package models

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID          int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string         `json:"name" gorm:"type:varchar(255);not null"`
	Description string         `json:"description" gorm:"type:text"`
	ParentID    *int64         `json:"parent_id,omitempty"`
	Position    int            `json:"position" gorm:"default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (Category) TableName() string {
	return "ac_categories"
}

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description"`
	ParentID    *int64 `json:"parent_id,omitempty"`
	Position    int    `json:"position"`
}

type UpdateCategoryRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty"`
	ParentID    *int64  `json:"parent_id,omitempty"`
	Position    *int    `json:"position,omitempty"`
}

type CategoryResponse struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	ParentID    *int64     `json:"parent_id,omitempty"`
	Position    int        `json:"position"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (c *Category) ToResponse() *CategoryResponse {
	return &CategoryResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		ParentID:    c.ParentID,
		Position:    c.Position,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}
