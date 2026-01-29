package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"io.lazydoge/aclove/logger"
	"io.lazydoge/aclove/models/query"
)

const (
	cookieLength     = 32
	cookieExpiration = 30 * 24 * time.Hour
	cookieName       = "aclove_session"
)

type AnonymousUserStatus string

const (
	StatusActive   AnonymousUserStatus = "active"
	StatusBanned   AnonymousUserStatus = "banned"
	StatusCooldown AnonymousUserStatus = "cooldown"
)

type AnonymousUserService struct {
	repo query.AnonymousUserRepo
}

func NewAnonymousUserService(repo query.AnonymousUserRepo) *AnonymousUserService {
	return &AnonymousUserService{repo: repo}
}

type AnonymousUserInfo struct {
	ID              int64
	Cookie          string
	FingerprintHash string
	IP              string
	Status          AnonymousUserStatus
	StatusReason    string
	BannedUntil     *time.Time
	CooldownUntil   *time.Time
}

func (s *AnonymousUserService) GenerateCookie() (string, error) {
	bytes := make([]byte, cookieLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成随机cookie失败: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func (s *AnonymousUserService) GetOrCreateUser(ctx context.Context, cookie string, clientIP string, fingerprint string) (*AnonymousUserInfo, error) {
	if cookie != "" {
		user, err := s.repo.GetByCookie(ctx, cookie)
		if err == nil {
			if err := s.validateUserStatus(user); err != nil {
				return nil, err
			}
			return s.toUserInfo(user), nil
		}
		if err != query.ErrAnonymousUserNotFound {
			logger.Error("查询匿名用户失败", "error", err, "cookie", cookie)
			return nil, fmt.Errorf("查询用户失败: %w", err)
		}
	}

	newCookie, err := s.GenerateCookie()
	if err != nil {
		return nil, err
	}

	var fingerprintHash *string
	if fingerprint != "" {
		hash := s.hashFingerprint(fingerprint)
		fingerprintHash = &hash
	}

	user := &query.AnonymousUserDO{
		Cookie:          newCookie,
		FingerprintHash: fingerprintHash,
		IP:              clientIP,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		logger.Error("创建匿名用户失败", "error", err, "ip", clientIP)
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	logger.Info("创建新匿名用户", "user_id", user.ID, "ip", clientIP)
	return s.toUserInfo(user), nil
}

func (s *AnonymousUserService) GetUserByCookie(ctx context.Context, cookie string) (*AnonymousUserInfo, error) {
	user, err := s.repo.GetByCookie(ctx, cookie)
	if err != nil {
		if err == query.ErrAnonymousUserNotFound {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	if err := s.validateUserStatus(user); err != nil {
		return nil, err
	}

	return s.toUserInfo(user), nil
}

func (s *AnonymousUserService) validateUserStatus(user *query.AnonymousUserDO) error {
	status := AnonymousUserStatus("active")
	if user.Status != nil {
		status = AnonymousUserStatus(*user.Status)
	}

	now := time.Now()

	if status == StatusBanned {
		if user.BannedUntil != nil && user.BannedUntil.After(now) {
			return &UserBannedError{
				Reason:      s.safeString(user.StatusReason),
				BannedUntil: *user.BannedUntil,
			}
		}
	}

	if status == StatusCooldown && user.CooldownUntil != nil && user.CooldownUntil.After(now) {
		return &UserCooldownError{
			CooldownUntil: *user.CooldownUntil,
		}
	}

	return nil
}

func (s *AnonymousUserService) BanUser(ctx context.Context, userID int64, until time.Time, reason string) error {
	if err := s.repo.Ban(ctx, userID, until); err != nil {
		return fmt.Errorf("封禁用户失败: %w", err)
	}
	logger.Info("封禁匿名用户", "user_id", userID, "until", until, "reason", reason)
	return nil
}

func (s *AnonymousUserService) UnbanUser(ctx context.Context, userID int64) error {
	if err := s.repo.Unban(ctx, userID); err != nil {
		return fmt.Errorf("解封用户失败: %w", err)
	}
	logger.Info("解封匿名用户", "user_id", userID)
	return nil
}

func (s *AnonymousUserService) SetCooldown(ctx context.Context, userID int64, duration time.Duration) error {
	until := time.Now().Add(duration)
	result := s.repo.WithContext(ctx).
		Model(&query.AnonymousUserDO{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"status":         string(StatusCooldown),
			"cooldown_until": until,
		})
	if result.Error != nil {
		return fmt.Errorf("设置冷却期失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	logger.Info("设置匿名用户冷却期", "user_id", userID, "until", until)
	return nil
}

func (s *AnonymousUserService) hashFingerprint(fingerprint string) string {
	return fingerprint
}

func (s *AnonymousUserService) toUserInfo(user *query.AnonymousUserDO) *AnonymousUserInfo {
	info := &AnonymousUserInfo{
		ID:     user.ID,
		Cookie: user.Cookie,
		IP:     user.IP,
	}

	if user.FingerprintHash != nil {
		info.FingerprintHash = *user.FingerprintHash
	}

	if user.Status != nil {
		info.Status = AnonymousUserStatus(*user.Status)
	}

	if user.StatusReason != nil {
		info.StatusReason = *user.StatusReason
	}

	if user.BannedUntil != nil {
		info.BannedUntil = user.BannedUntil
	}

	if user.CooldownUntil != nil {
		info.CooldownUntil = user.CooldownUntil
	}

	return info
}

func (s *AnonymousUserService) safeString(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

var (
	ErrUserNotFound = fmt.Errorf("用户不存在")
)

type UserBannedError struct {
	Reason      string
	BannedUntil time.Time
}

func (e *UserBannedError) Error() string {
	return fmt.Sprintf("用户已被封禁至 %s: %s", e.BannedUntil.Format(time.RFC3339), e.Reason)
}

type UserCooldownError struct {
	CooldownUntil time.Time
}

func (e *UserCooldownError) Error() string {
	return fmt.Sprintf("用户处于冷却期，直到 %s", e.CooldownUntil.Format(time.RFC3339))
}

func ExtractClientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

func CookieConfig() (string, time.Duration) {
	return cookieName, cookieExpiration
}
