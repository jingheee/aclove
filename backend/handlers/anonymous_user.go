package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"io.lazydoge/aclove/logger"
	"io.lazydoge/aclove/middleware"
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
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

	c.JSON(http.StatusOK, resp)
}

type BanRequest struct {
	UserID int64     `json:"user_id" binding:"required"`
	Until  time.Time `json:"until" binding:"required"`
	Reason string    `json:"reason"`
}

func (h *AnonymousUserHandler) BanUser(c *gin.Context) {
	var req BanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	if req.Until.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "封禁时间必须在未来"})
		return
	}

	if err := h.userService.BanUser(c.Request.Context(), req.UserID, req.Until, req.Reason); err != nil {
		logger.Error("封禁用户失败", "error", err, "user_id", req.UserID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "封禁用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户已封禁"})
}

type UnbanRequest struct {
	UserID int64 `json:"user_id" binding:"required"`
}

func (h *AnonymousUserHandler) UnbanUser(c *gin.Context) {
	var req UnbanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	if err := h.userService.UnbanUser(c.Request.Context(), req.UserID); err != nil {
		logger.Error("解封用户失败", "error", err, "user_id", req.UserID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解封用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户已解封"})
}

type CooldownRequest struct {
	UserID   int64         `json:"user_id" binding:"required"`
	Duration time.Duration `json:"duration" binding:"required"`
}

func (h *AnonymousUserHandler) SetCooldown(c *gin.Context) {
	var req CooldownRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	if req.Duration <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "冷却时长必须大于0"})
		return
	}

	if err := h.userService.SetCooldown(c.Request.Context(), req.UserID, req.Duration); err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		logger.Error("设置冷却期失败", "error", err, "user_id", req.UserID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "设置冷却期失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "冷却期已设置"})
}
