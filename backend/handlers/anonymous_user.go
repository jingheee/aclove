package handlers

import (
	"time"

	"github.com/gin-gonic/gin"

	"io.lazydoge/aclove/logger"
	"io.lazydoge/aclove/middleware"
	"io.lazydoge/aclove/models"
	"io.lazydoge/aclove/service"
)

type AnonymousUserHandler struct {
	userService *service.AnonymousUserService
}

func NewAnonymousUserHandler(userService *service.AnonymousUserService) *AnonymousUserHandler {
	return &AnonymousUserHandler{userService: userService}
}

type UserResponse struct {
	ID            int64      `json:"id"`
	Status        string     `json:"status"`
	StatusReason  string     `json:"status_reason,omitempty"`
	BannedUntil   *time.Time `json:"banned_until,omitempty"`
	CooldownUntil *time.Time `json:"cooldown_until,omitempty"`
}

func (h *AnonymousUserHandler) GetCurrentUser(c *gin.Context) {
	user, exists := middleware.GetAnonymousUser(c)
	if !exists {
		models.JSONUnauthorized(c, "未找到用户信息")
		return
	}

	resp := UserResponse{
		ID:     user.ID,
		Status: string(user.Status),
	}

	if user.StatusReason != "" {
		resp.StatusReason = user.StatusReason
	}

	if user.BannedUntil != nil {
		resp.BannedUntil = user.BannedUntil
	}

	if user.CooldownUntil != nil {
		resp.CooldownUntil = user.CooldownUntil
	}

	models.JSONSuccess(c, resp)
}

type BanRequest struct {
	UserID int64     `json:"user_id" binding:"required"`
	Until  time.Time `json:"until" binding:"required"`
	Reason string    `json:"reason"`
}

func (h *AnonymousUserHandler) BanUser(c *gin.Context) {
	var req BanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.JSONBadRequest(c, "无效的请求参数")
		return
	}

	if req.Until.Before(time.Now()) {
		models.JSONBadRequest(c, "封禁时间必须在未来")
		return
	}

	if err := h.userService.BanUser(c.Request.Context(), req.UserID, req.Until, req.Reason); err != nil {
		logger.Error("封禁用户失败", "error", err, "user_id", req.UserID)
		models.JSONInternalError(c, "封禁用户失败")
		return
	}

	models.JSONSuccessWithMsg(c, "用户已封禁", nil)
}

type UnbanRequest struct {
	UserID int64 `json:"user_id" binding:"required"`
}

func (h *AnonymousUserHandler) UnbanUser(c *gin.Context) {
	var req UnbanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.JSONBadRequest(c, "无效的请求参数")
		return
	}

	if err := h.userService.UnbanUser(c.Request.Context(), req.UserID); err != nil {
		logger.Error("解封用户失败", "error", err, "user_id", req.UserID)
		models.JSONInternalError(c, "解封用户失败")
		return
	}

	models.JSONSuccessWithMsg(c, "用户已解封", nil)
}

type CooldownRequest struct {
	UserID   int64         `json:"user_id" binding:"required"`
	Duration time.Duration `json:"duration" binding:"required"`
}

func (h *AnonymousUserHandler) SetCooldown(c *gin.Context) {
	var req CooldownRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.JSONBadRequest(c, "无效的请求参数")
		return
	}

	if req.Duration <= 0 {
		models.JSONBadRequest(c, "冷却时长必须大于0")
		return
	}

	if err := h.userService.SetCooldown(c.Request.Context(), req.UserID, req.Duration); err != nil {
		if err == service.ErrUserNotFound {
			models.JSONNotFound(c, "用户不存在")
			return
		}
		logger.Error("设置冷却期失败", "error", err, "user_id", req.UserID)
		models.JSONInternalError(c, "设置冷却期失败")
		return
	}

	models.JSONSuccessWithMsg(c, "冷却期已设置", nil)
}
