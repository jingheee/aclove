package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/rueidis"
	"gorm.io/datatypes"

	"io.lazydoge/aclove/jsonutil"
	"io.lazydoge/aclove/logger"
	"io.lazydoge/aclove/models"
	"io.lazydoge/aclove/models/query"
	"io.lazydoge/aclove/repository"
)

const (
	postCooldownKey      = "post:cooldown:%d"
	postDailyCountKey    = "post:daily:%d"
	postCacheKey         = "post:%d"
	postCooldownDuration = 30 * time.Second
	postDailyLimit       = 50
	cacheTTL             = 5 * time.Minute
)

type CategoryRepo interface {
	GetByID(ctx context.Context, id int64) (*models.Category, error)
}

type PostService struct {
	postRepo     query.PostRepo
	categoryRepo CategoryRepo
	redis        rueidis.Client
}

func NewPostService(postRepo query.PostRepo, categoryRepo CategoryRepo, redis rueidis.Client) *PostService {
	return &PostService{
		postRepo:     postRepo,
		categoryRepo: categoryRepo,
		redis:        redis,
	}
}

func (s *PostService) CreatePost(ctx context.Context, userID int64, req *models.CreatePostRequest, clientIP, userAgent string) (*models.Post, error) {
	if err := s.checkPostCooldown(ctx, userID); err != nil {
		return nil, err
	}

	categoryID := int64(req.CategoryID)
	if err := s.validateCategory(ctx, categoryID); err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}

	mediaJSON, err := s.marshalMediaAttachments(req.MediaAttachments)
	if err != nil {
		return nil, err
	}

	editorType := "markdown"
	viewCount := 0
	replyCount := 0
	upvoteCount := 0
	downvoteCount := 0
	score := 0.0
	editCount := 0

	post := &models.Post{
		ID:               models.GenerateSnowflakeID(),
		UserID:           userID,
		CategoryID:       categoryID,
		Title:            req.Title,
		Content:          req.Content,
		MediaAttachments: mediaJSON,
		IP:               &clientIP,
		UserAgent:        &userAgent,
		EditorType:       &editorType,
		ViewCount:        &viewCount,
		ReplyCount:       &replyCount,
		UpvoteCount:      &upvoteCount,
		DownvoteCount:    &downvoteCount,
		Score:            &score,
		EditCount:        &editCount,
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		logger.Error("创建帖子失败", "error", err, "user_id", userID)
		return nil, fmt.Errorf("创建帖子失败: %w", err)
	}

	s.setPostCooldown(ctx, userID)
	s.incrementDailyPostCount(ctx, userID)

	logger.Info("创建帖子成功", "post_id", post.ID, "user_id", userID, "category_id", categoryID)
	return post, nil
}

func (s *PostService) GetPost(ctx context.Context, id int64) (*models.Post, error) {
	cacheKey := fmt.Sprintf(postCacheKey, id)

	if s.redis != nil {
		cmd := s.redis.B().Get().Key(cacheKey).Build()
		res := s.redis.Do(ctx, cmd)
		if val, err := res.ToString(); err == nil {
			var cachedPost models.Post
			if err := json.Unmarshal([]byte(val), &cachedPost); err == nil && cachedPost.ID != 0 {
				return &cachedPost, nil
			}
		}
	}

	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, query.ErrPostNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	if post.DeletedAt.Valid {
		return nil, ErrPostNotFound
	}

	if s.redis != nil {
		data, _ := json.Marshal(post)
		s.redis.Do(ctx, s.redis.B().Set().Key(cacheKey).Value(string(data)).Ex(cacheTTL).Build())
	}

	return post, nil
}

func (s *PostService) GetPostDetail(ctx context.Context, id int64) (*models.PostDetail, error) {
	post, err := s.GetPost(ctx, id)
	if err != nil {
		return nil, err
	}

	category, err := s.categoryRepo.GetByID(ctx, post.CategoryID)
	if err != nil {
		logger.Warn("获取分类信息失败", "error", err, "category_id", post.CategoryID)
		category = &models.Category{ID: post.CategoryID, Name: "未知分类"}
	}

	s.postRepo.IncrementViewCount(ctx, id)

	return models.ToPostDetail(post, category.Name), nil
}

func (s *PostService) ListPosts(ctx context.Context, query *models.ListPostsQuery) (*models.PostListResponse, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 50 {
		query.PageSize = 20
	}

	offset := (query.Page - 1) * query.PageSize

	var posts []*models.Post
	var total int64
	var err error

	if query.CategoryID != nil {
		categoryID := int64(*query.CategoryID)
		posts, err = s.postRepo.ListByCategory(ctx, categoryID, offset, query.PageSize)
		if err != nil {
			return nil, err
		}
		total, err = s.postRepo.CountByCategory(ctx, categoryID)
	} else {
		switch query.Sort {
		case "hot":
			posts, err = s.listByHot(ctx, offset, query.PageSize)
		case "top":
			posts, err = s.listByTop(ctx, offset, query.PageSize)
		default:
			posts, err = s.postRepo.List(ctx, offset, query.PageSize)
		}
		if err != nil {
			return nil, err
		}
		total, err = s.postRepo.Count(ctx)
	}

	if err != nil {
		return nil, err
	}

	items := make([]*models.PostListItem, len(posts))
	for i, post := range posts {
		categoryName := ""
		if category, err := s.categoryRepo.GetByID(ctx, post.CategoryID); err == nil {
			categoryName = category.Name
		}
		items[i] = models.ToPostListItem(post, categoryName)
	}

	return &models.PostListResponse{
		Items:    items,
		Total:    jsonutil.Int64(total),
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func (s *PostService) listByHot(ctx context.Context, offset, limit int) ([]*models.Post, error) {
	var posts []*models.Post
	result := s.postRepo.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("score DESC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&posts)
	return posts, result.Error
}

func (s *PostService) listByTop(ctx context.Context, offset, limit int) ([]*models.Post, error) {
	var posts []*models.Post
	result := s.postRepo.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("upvote_count DESC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&posts)
	return posts, result.Error
}

func (s *PostService) UpdatePost(ctx context.Context, postID, userID int64, req *models.UpdatePostRequest) (*models.Post, error) {
	post, err := s.GetPost(ctx, postID)
	if err != nil {
		return nil, err
	}

	if post.UserID != userID {
		return nil, ErrNotPostAuthor
	}

	if req.Title != nil {
		post.Title = *req.Title
	}
	if req.Content != nil {
		post.Content = *req.Content
	}
	if len(req.MediaAttachments) > 0 {
		mediaJSON, err := s.marshalMediaAttachments(req.MediaAttachments)
		if err != nil {
			return nil, err
		}
		post.MediaAttachments = mediaJSON
	}

	now := time.Now()
	post.LastEditedAt = &now
	editCount := 0
	if post.EditCount != nil {
		editCount = *post.EditCount
	}
	editCount++
	post.EditCount = &editCount

	if err := s.postRepo.Update(ctx, post); err != nil {
		logger.Error("更新帖子失败", "error", err, "post_id", postID)
		return nil, fmt.Errorf("更新帖子失败: %w", err)
	}

	s.invalidatePostCache(ctx, postID)

	logger.Info("更新帖子成功", "post_id", postID, "user_id", userID)
	return post, nil
}

func (s *PostService) DeletePost(ctx context.Context, postID, userID int64, isAdmin bool, reason string) error {
	post, err := s.GetPost(ctx, postID)
	if err != nil {
		return err
	}

	if !isAdmin && post.UserID != userID {
		return ErrNotPostAuthor
	}

	if isAdmin && reason != "" {
		if err := s.postRepo.DeleteWithReason(ctx, postID, reason); err != nil {
			logger.Error("删除帖子失败", "error", err, "post_id", postID)
			return fmt.Errorf("删除帖子失败: %w", err)
		}
	} else {
		if err := s.postRepo.Delete(ctx, postID); err != nil {
			logger.Error("删除帖子失败", "error", err, "post_id", postID)
			return fmt.Errorf("删除帖子失败: %w", err)
		}
	}

	s.invalidatePostCache(ctx, postID)

	logger.Info("删除帖子成功", "post_id", postID, "user_id", userID, "is_admin", isAdmin)
	return nil
}

func (s *PostService) checkPostCooldown(ctx context.Context, userID int64) error {
	if s.redis == nil {
		return nil
	}

	cooldownKey := fmt.Sprintf(postCooldownKey, userID)
	exists, err := s.redis.Do(ctx, s.redis.B().Exists().Key(cooldownKey).Build()).AsBool()
	if err != nil {
		logger.Warn("检查发帖冷却期失败", "error", err, "user_id", userID)
		return nil
	}
	if exists {
		return ErrPostCooldown
	}

	dailyKey := fmt.Sprintf(postDailyCountKey, userID)
	count, err := s.redis.Do(ctx, s.redis.B().Get().Key(dailyKey).Build()).AsInt64()
	if err != nil && !rueidis.IsRedisNil(err) {
		logger.Warn("获取日发帖数失败", "error", err, "user_id", userID)
		return nil
	}
	if count >= postDailyLimit {
		return ErrDailyPostLimit
	}

	return nil
}

func (s *PostService) setPostCooldown(ctx context.Context, userID int64) {
	if s.redis == nil {
		return
	}

	cooldownKey := fmt.Sprintf(postCooldownKey, userID)
	s.redis.Do(ctx, s.redis.B().Set().Key(cooldownKey).Value("1").Ex(postCooldownDuration).Build())
}

func (s *PostService) incrementDailyPostCount(ctx context.Context, userID int64) {
	if s.redis == nil {
		return
	}

	dailyKey := fmt.Sprintf(postDailyCountKey, userID)
	s.redis.Do(ctx, s.redis.B().Incr().Key(dailyKey).Build())
	s.redis.Do(ctx, s.redis.B().Expire().Key(dailyKey).Seconds(86400).Build())
}

func (s *PostService) invalidatePostCache(ctx context.Context, postID int64) {
	if s.redis != nil {
		cacheKey := fmt.Sprintf(postCacheKey, postID)
		s.redis.Do(ctx, s.redis.B().Del().Key(cacheKey).Build())
	}
}

func (s *PostService) validateCategory(ctx context.Context, categoryID int64) error {
	_, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}
	return nil
}

func (s *PostService) marshalMediaAttachments(attachments []models.MediaAttachment) (datatypes.JSON, error) {
	if len(attachments) == 0 {
		return datatypes.JSON("[]"), nil
	}
	data, err := json.Marshal(attachments)
	if err != nil {
		return nil, fmt.Errorf("序列化媒体附件失败: %w", err)
	}
	return datatypes.JSON(data), nil
}

var (
	ErrPostNotFound     = errors.New("帖子不存在")
	ErrNotPostAuthor    = errors.New("不是帖子作者")
	ErrPostCooldown     = errors.New("发帖过于频繁，请稍后再试")
	ErrDailyPostLimit   = errors.New("今日发帖已达上限")
	ErrCategoryNotFound = errors.New("分类不存在")
)
