package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"io.lazydoge/aclove/logger"
	"io.lazydoge/aclove/middleware"
	"io.lazydoge/aclove/models"
	"io.lazydoge/aclove/service"
)

type PostHandler struct {
	postService *service.PostService
}

func NewPostHandler(postService *service.PostService) *PostHandler {
	return &PostHandler{postService: postService}
}

func (h *PostHandler) Create(c *gin.Context) {
	user, exists := middleware.GetAnonymousUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先获取会话"})
		return
	}

	var req models.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据", "details": err.Error()})
		return
	}

	clientIP := middleware.ExtractClientIP(c.Request.RemoteAddr)
	userAgent := c.Request.UserAgent()

	post, err := h.postService.CreatePost(c.Request.Context(), user.ID, &req, clientIP, userAgent)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPostCooldown):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "发帖过于频繁，请30秒后再试"})
		case errors.Is(err, service.ErrDailyPostLimit):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "今日发帖已达上限（50帖）"})
		case errors.Is(err, service.ErrCategoryNotFound):
			c.JSON(http.StatusBadRequest, gin.H{"error": "所选分类不存在"})
		default:
			logger.Error("创建帖子失败", "error", err, "user_id", user.ID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建帖子失败"})
		}
		return
	}

	c.JSON(http.StatusCreated, models.ToPostDetail(post, ""))
}

func (h *PostHandler) Get(c *gin.Context) {
	id, err := parsePostID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的帖子ID"})
		return
	}

	post, err := h.postService.GetPostDetail(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
			return
		}
		logger.Error("获取帖子详情失败", "error", err, "post_id", id)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取帖子详情失败"})
		return
	}

	c.JSON(http.StatusOK, post)
}

func (h *PostHandler) List(c *gin.Context) {
	var query models.ListPostsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的查询参数", "details": err.Error()})
		return
	}

	response, err := h.postService.ListPosts(c.Request.Context(), &query)
	if err != nil {
		logger.Error("获取帖子列表失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取帖子列表失败"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *PostHandler) Update(c *gin.Context) {
	user, exists := middleware.GetAnonymousUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先获取会话"})
		return
	}

	id, err := parsePostID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的帖子ID"})
		return
	}

	var req models.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据", "details": err.Error()})
		return
	}

	post, err := h.postService.UpdatePost(c.Request.Context(), id, user.ID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPostNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
		case errors.Is(err, service.ErrNotPostAuthor):
			c.JSON(http.StatusForbidden, gin.H{"error": "只有作者可以编辑帖子"})
		default:
			logger.Error("更新帖子失败", "error", err, "post_id", id, "user_id", user.ID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新帖子失败"})
		}
		return
	}

	c.JSON(http.StatusOK, models.ToPostDetail(post, ""))
}

func (h *PostHandler) Delete(c *gin.Context) {
	user, exists := middleware.GetAnonymousUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先获取会话"})
		return
	}

	id, err := parsePostID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的帖子ID"})
		return
	}

	var req models.DeletePostRequest
	c.ShouldBindJSON(&req)

	isAdmin := false

	err = h.postService.DeletePost(c.Request.Context(), id, user.ID, isAdmin, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPostNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
		case errors.Is(err, service.ErrNotPostAuthor):
			c.JSON(http.StatusForbidden, gin.H{"error": "只有作者可以删除帖子"})
		default:
			logger.Error("删除帖子失败", "error", err, "post_id", id, "user_id", user.ID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除帖子失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func parsePostID(c *gin.Context) (int64, error) {
	idStr := c.Param("id")
	if idStr == "" {
		return 0, errors.New("empty id")
	}
	return strconv.ParseInt(idStr, 10, 64)
}
