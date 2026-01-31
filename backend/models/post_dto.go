package models

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"io.lazydoge/aclove/models/query"
)

type Post = query.PostDO

type MediaAttachment struct {
	URL  string `json:"url" binding:"required,url"`
	Type string `json:"type" binding:"required,oneof=image video audio file"`
	Size int64  `json:"size" binding:"min=0"`
}

type CreatePostRequest struct {
	CategoryID        int64             `json:"category_id" binding:"required"`
	Title             string            `json:"title" binding:"required,min=5,max=100"`
	Content           string            `json:"content" binding:"required,min=10,max=10000"`
	MediaAttachments  []MediaAttachment `json:"media_attachments,omitempty" binding:"omitempty,max=9,dive"`
}

type UpdatePostRequest struct {
	Title             *string           `json:"title,omitempty" binding:"omitempty,min=5,max=100"`
	Content           *string           `json:"content,omitempty" binding:"omitempty,min=10,max=10000"`
	MediaAttachments  []MediaAttachment `json:"media_attachments,omitempty" binding:"omitempty,max=9,dive"`
}

type PostAuthor struct {
	ID        int64  `json:"id"`
	Anonymous bool   `json:"anonymous"`
}

type PostCategory struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type PostListItem struct {
	ID              int64        `json:"id"`
	Title           string       `json:"title"`
	Summary         string       `json:"summary"`
	Category        PostCategory `json:"category"`
	Author          PostAuthor   `json:"author"`
	ViewCount       int          `json:"view_count"`
	ReplyCount      int          `json:"reply_count"`
	UpvoteCount     int          `json:"upvote_count"`
	MediaCount      int          `json:"media_count"`
	CreatedAt       time.Time    `json:"created_at"`
	LastEditedAt    *time.Time   `json:"last_edited_at,omitempty"`
	EditCount       int          `json:"edit_count"`
}

type PostDetail struct {
	ID                int64             `json:"id"`
	Title             string            `json:"title"`
	Content           string            `json:"content"`
	ContentRendered   string            `json:"content_rendered"`
	MediaAttachments  datatypes.JSON    `json:"media_attachments"`
	Category          PostCategory      `json:"category"`
	Author            PostAuthor        `json:"author"`
	ViewCount         int               `json:"view_count"`
	ReplyCount        int               `json:"reply_count"`
	UpvoteCount       int               `json:"upvote_count"`
	DownvoteCount     int               `json:"downvote_count"`
	EditCount         int               `json:"edit_count"`
	CreatedAt         time.Time         `json:"created_at"`
	LastEditedAt      *time.Time        `json:"last_edited_at,omitempty"`
}

type PostListResponse struct {
	Items      []*PostListItem `json:"items"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
}

type ListPostsQuery struct {
	CategoryID *int64 `form:"category_id"`
	Page       int    `form:"page,default=1" binding:"min=1"`
	PageSize   int    `form:"page_size,default=20" binding:"min=1,max=50"`
	Sort       string `form:"sort,default=newest" binding:"oneof=newest hot top"`
}

type DeletePostRequest struct {
	Reason string `json:"reason,omitempty"`
}

func GenerateSummary(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}

func ToPostListItem(post *Post, categoryName string) *PostListItem {
	viewCount := 0
	if post.ViewCount != nil {
		viewCount = *post.ViewCount
	}
	replyCount := 0
	if post.ReplyCount != nil {
		replyCount = *post.ReplyCount
	}
	upvoteCount := 0
	if post.UpvoteCount != nil {
		upvoteCount = *post.UpvoteCount
	}
	editCount := 0
	if post.EditCount != nil {
		editCount = *post.EditCount
	}

	mediaCount := 0
	if len(post.MediaAttachments) > 0 {
		var attachments []MediaAttachment
		if err := json.Unmarshal(post.MediaAttachments, &attachments); err == nil {
			mediaCount = len(attachments)
		}
	}

	return &PostListItem{
		ID:       post.ID,
		Title:    post.Title,
		Summary:  GenerateSummary(post.Content, 200),
		Category: PostCategory{ID: post.CategoryID, Name: categoryName},
		Author:   PostAuthor{ID: post.UserID, Anonymous: true},
		ViewCount:    viewCount,
		ReplyCount:   replyCount,
		UpvoteCount:  upvoteCount,
		MediaCount:   mediaCount,
		CreatedAt:    post.CreatedAt,
		LastEditedAt: post.LastEditedAt,
		EditCount:    editCount,
	}
}

func ToPostDetail(post *Post, categoryName string) *PostDetail {
	viewCount := 0
	if post.ViewCount != nil {
		viewCount = *post.ViewCount
	}
	replyCount := 0
	if post.ReplyCount != nil {
		replyCount = *post.ReplyCount
	}
	upvoteCount := 0
	if post.UpvoteCount != nil {
		upvoteCount = *post.UpvoteCount
	}
	downvoteCount := 0
	if post.DownvoteCount != nil {
		downvoteCount = *post.DownvoteCount
	}
	editCount := 0
	if post.EditCount != nil {
		editCount = *post.EditCount
	}

	contentRendered := ""
	if post.ContentRendered != nil {
		contentRendered = *post.ContentRendered
	}

	return &PostDetail{
		ID:               post.ID,
		Title:            post.Title,
		Content:          post.Content,
		ContentRendered:  contentRendered,
		MediaAttachments: post.MediaAttachments,
		Category:         PostCategory{ID: post.CategoryID, Name: categoryName},
		Author:           PostAuthor{ID: post.UserID, Anonymous: true},
		ViewCount:        viewCount,
		ReplyCount:       replyCount,
		UpvoteCount:      upvoteCount,
		DownvoteCount:    downvoteCount,
		EditCount:        editCount,
		CreatedAt:        post.CreatedAt,
		LastEditedAt:     post.LastEditedAt,
	}
}
