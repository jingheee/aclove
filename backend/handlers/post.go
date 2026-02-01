package handlers

import (
	"errors"
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
		models.JSONUnauthorized(c, "请先获取会话")
		return
	}

	var req models.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.JSONBadRequest(c, "无效的请求数据")
		return
	}

	clientIP := middleware.ExtractClientIP(c.Request.RemoteAddr)
	userAgent := c.Request.UserAgent()

	post, err := h.postService.CreatePost(c.Request.Context(), int64(user.ID), &req, clientIP, userAgent)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPostCooldown):
			models.JSONTooManyRequests(c, "发帖过于频繁，请30秒后再试")
		case errors.Is(err, service.ErrDailyPostLimit):
			models.JSONTooManyRequests(c, "今日发帖已达上限（50帖）")
		case errors.Is(err, service.ErrCategoryNotFound):
			models.JSONBadRequest(c, "所选分类不存在")
		default:
			logger.Error("创建帖子失败", "error", err, "user_id", user.ID)
			models.JSONInternalError(c, "创建帖子失败")
		}
		return
	}

	models.JSONCreated(c, models.ToPostDetail(post, ""))
}

func (h *PostHandler) Get(c *gin.Context) {
	id, err := parsePostID(c)
	if err != nil {
		models.JSONBadRequest(c, "无效的帖子ID")
		return
	}

	post, err := h.postService.GetPostDetail(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			models.JSONNotFound(c, "帖子不存在")
			return
		}
		logger.Error("获取帖子详情失败", "error", err, "post_id", id)
		models.JSONInternalError(c, "获取帖子详情失败")
		return
	}

	models.JSONSuccess(c, post)
}

func (h *PostHandler) List(c *gin.Context) {
	var query models.ListPostsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		models.JSONBadRequest(c, "无效的查询参数")
		return
	}

	response, err := h.postService.ListPosts(c.Request.Context(), &query)
	if err != nil {
		logger.Error("获取帖子列表失败", "error", err)
		models.JSONInternalError(c, "获取帖子列表失败")
		return
	}

	models.JSONSuccess(c, response)
}

func (h *PostHandler) Update(c *gin.Context) {
	user, exists := middleware.GetAnonymousUser(c)
	if !exists {
		models.JSONUnauthorized(c, "请先获取会话")
		return
	}

	id, err := parsePostID(c)
	if err != nil {
		models.JSONBadRequest(c, "无效的帖子ID")
		return
	}

	var req models.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.JSONBadRequest(c, "无效的请求数据")
		return
	}

	post, err := h.postService.UpdatePost(c.Request.Context(), id, int64(user.ID), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPostNotFound):
			models.JSONNotFound(c, "帖子不存在")
		case errors.Is(err, service.ErrNotPostAuthor):
			models.JSONForbidden(c, "只有作者可以编辑帖子")
		default:
			logger.Error("更新帖子失败", "error", err, "post_id", id, "user_id", user.ID)
			models.JSONInternalError(c, "更新帖子失败")
		}
		return
	}

	models.JSONSuccess(c, models.ToPostDetail(post, ""))
}

func (h *PostHandler) Delete(c *gin.Context) {
	user, exists := middleware.GetAnonymousUser(c)
	if !exists {
		models.JSONUnauthorized(c, "请先获取会话")
		return
	}

	id, err := parsePostID(c)
	if err != nil {
		models.JSONBadRequest(c, "无效的帖子ID")
		return
	}

	var req models.DeletePostRequest
	c.ShouldBindJSON(&req)

	isAdmin := false

	err = h.postService.DeletePost(c.Request.Context(), id, int64(user.ID), isAdmin, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPostNotFound):
			models.JSONNotFound(c, "帖子不存在")
		case errors.Is(err, service.ErrNotPostAuthor):
			models.JSONForbidden(c, "只有作者可以删除帖子")
		default:
			logger.Error("删除帖子失败", "error", err, "post_id", id, "user_id", user.ID)
			models.JSONInternalError(c, "删除帖子失败")
		}
		return
	}

	models.JSONSuccessWithMsg(c, "删除成功", nil)
}

func parsePostID(c *gin.Context) (int64, error) {
	idStr := c.Param("id")
	if idStr == "" {
		return 0, errors.New("empty id")
	}
	return strconv.ParseInt(idStr, 10, 64)
}
