package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"io.lazydoge/aclove/logger"
	"io.lazydoge/aclove/models"
	"io.lazydoge/aclove/models/query"
)

type ManagerConfig struct {
	StoreConfig
	CookieName     string
	CookieDomain   string
	CookieSecure   bool
	CookieHttpOnly bool
	CookieSameSite string
}

func (c *ManagerConfig) setDefaults() {
	c.StoreConfig.setDefaults()
	if c.CookieName == "" {
		c.CookieName = "aclove_session"
	}
	if c.CookieSameSite == "" {
		c.CookieSameSite = "Lax"
	}
}

type Manager struct {
	store  *Store
	repo   query.AnonymousUserRepo
	config ManagerConfig
}

func NewManager(repo query.AnonymousUserRepo, cfg ManagerConfig) (*Manager, error) {
	cfg.setDefaults()

	store, err := NewStore(cfg.StoreConfig)
	if err != nil {
		return nil, fmt.Errorf("创建session store失败: %w", err)
	}

	return &Manager{
		store:  store,
		repo:   repo,
		config: cfg,
	}, nil
}

func (m *Manager) GenerateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成session ID失败: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func (m *Manager) GetOrCreateSession(ctx context.Context, sessionID string, clientIP string, fingerprint string) (*Session, error) {
	if sessionID != "" {
		session, err := m.store.Get(ctx, sessionID)
		if err == nil && session != nil {
			if err := m.validateSession(session); err != nil {
				return nil, err
			}

			if session.IP != clientIP {
				session, _ = m.store.Update(ctx, sessionID, func(s *Session) error {
					s.IP = clientIP
					return nil
				})
			}

			return session, nil
		}
	}

	return m.createNewSession(ctx, clientIP, fingerprint)
}

func (m *Manager) createNewSession(ctx context.Context, clientIP string, fingerprint string) (*Session, error) {
	newSessionID, err := m.GenerateSessionID()
	if err != nil {
		return nil, err
	}

	var fingerprintHash *string
	if fingerprint != "" {
		hash := m.hashFingerprint(fingerprint)
		fingerprintHash = &hash
	}

	userDO := &query.AnonymousUserDO{
		ID:              models.GenerateSnowflakeID(),
		Cookie:          newSessionID,
		FingerprintHash: fingerprintHash,
		IP:              clientIP,
	}

	if err := m.repo.Create(ctx, userDO); err != nil {
		logger.Error("创建匿名用户失败", "error", err, "ip", clientIP)
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	userInfo := &UserInfo{
		ID:              userDO.ID,
		Cookie:          userDO.Cookie,
		FingerprintHash: m.safeString(userDO.FingerprintHash),
		IP:              userDO.IP,
		Status:          StatusActive,
		CreatedAt:       time.Now(),
	}

	if userDO.Status != nil {
		userInfo.Status = Status(*userDO.Status)
	}
	if userDO.StatusReason != nil {
		userInfo.StatusReason = *userDO.StatusReason
	}
	if userDO.BannedUntil != nil {
		userInfo.BannedUntil = userDO.BannedUntil
	}
	if userDO.CooldownUntil != nil {
		userInfo.CooldownUntil = userDO.CooldownUntil
	}

	session, err := m.store.Create(ctx, newSessionID, userInfo, clientIP)
	if err != nil {
		logger.Error("创建session失败", "error", err, "user_id", userDO.ID)
		return nil, fmt.Errorf("创建session失败: %w", err)
	}

	logger.Info("创建新用户session", "user_id", userDO.ID, "session_id", newSessionID[:16], "ip", clientIP)
	return session, nil
}

func (m *Manager) validateSession(session *Session) error {
	if session.UserInfo == nil {
		return ErrInvalidSession
	}

	if session.UserInfo.IsBanned() {
		return &BannedError{
			Reason:      session.UserInfo.StatusReason,
			BannedUntil: session.UserInfo.BannedUntil,
		}
	}

	if session.UserInfo.IsCooldown() {
		return &CooldownError{
			CooldownUntil: session.UserInfo.CooldownUntil,
		}
	}

	return nil
}

func (m *Manager) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	return m.store.Get(ctx, sessionID)
}

func (m *Manager) RefreshSession(ctx context.Context, sessionID string) (*Session, error) {
	if err := m.store.RefreshTTL(ctx, sessionID); err != nil {
		return nil, err
	}
	return m.store.Get(ctx, sessionID)
}

func (m *Manager) DeleteSession(ctx context.Context, sessionID string) error {
	return m.store.Delete(ctx, sessionID)
}

func (m *Manager) UpdateUserInfo(ctx context.Context, userID int64, updateFn func(*UserInfo) error) error {
	sessions, err := m.store.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, session := range sessions {
		_, err := m.store.Update(ctx, session.ID, func(s *Session) error {
			if s.UserInfo == nil {
				s.UserInfo = &UserInfo{ID: userID}
			}
			return updateFn(s.UserInfo)
		})
		if err != nil {
			logger.Warn("更新session用户信息失败", "error", err, "session_id", session.ID[:16])
		}
	}

	return nil
}

func (m *Manager) Close() error {
	return m.store.Close()
}

func (m *Manager) CookieConfig() (name string, maxAge int, domain string, secure bool, httpOnly bool, sameSite string) {
	return m.config.CookieName,
		int(m.config.SessionTTL.Seconds()),
		m.config.CookieDomain,
		m.config.CookieSecure,
		m.config.CookieHttpOnly,
		m.config.CookieSameSite
}

func (m *Manager) hashFingerprint(fingerprint string) string {
	return fingerprint
}

func (m *Manager) safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

type BannedError struct {
	Reason      string
	BannedUntil *time.Time
}

func (e *BannedError) Error() string {
	if e.BannedUntil != nil {
		return fmt.Sprintf("用户已被封禁至 %s: %s", e.BannedUntil.Format(time.RFC3339), e.Reason)
	}
	return fmt.Sprintf("用户已被永久封禁: %s", e.Reason)
}

type CooldownError struct {
	CooldownUntil *time.Time
}

func (e *CooldownError) Error() string {
	if e.CooldownUntil != nil {
		return fmt.Sprintf("用户处于冷却期，直到 %s", e.CooldownUntil.Format(time.RFC3339))
	}
	return "用户处于冷却期"
}
