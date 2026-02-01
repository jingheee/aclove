package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/rueidis"
	"io.lazydoge/aclove/jsonutil"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusBanned   Status = "banned"
	StatusCooldown Status = "cooldown"
)

type UserInfo struct {
	ID              jsonutil.Int64 `json:"id"`
	Cookie          string         `json:"cookie"`
	FingerprintHash string         `json:"fingerprint_hash"`
	IP              string         `json:"ip"`
	Status          Status         `json:"status"`
	StatusReason    string         `json:"status_reason,omitempty"`
	BannedUntil     *time.Time     `json:"banned_until,omitempty"`
	CooldownUntil   *time.Time     `json:"cooldown_until,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
}

func (u *UserInfo) IsBanned() bool {
	if u.Status != StatusBanned {
		return false
	}
	if u.BannedUntil != nil && u.BannedUntil.Before(time.Now()) {
		return false
	}
	return true
}

func (u *UserInfo) IsCooldown() bool {
	if u.Status != StatusCooldown {
		return false
	}
	if u.CooldownUntil != nil && u.CooldownUntil.Before(time.Now()) {
		return false
	}
	return true
}

type Session struct {
	ID        string         `json:"id"`
	UserID    jsonutil.Int64 `json:"user_id"`
	UserInfo  *UserInfo      `json:"user_info,omitempty"`
	IP        string         `json:"ip"`
	CreatedAt time.Time      `json:"created_at"`
	ExpiresAt time.Time      `json:"expires_at"`
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

type StoreConfig struct {
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	KeyPrefix       string
	SessionTTL      time.Duration
	CleanupInterval time.Duration
}

func (c *StoreConfig) setDefaults() {
	if c.KeyPrefix == "" {
		c.KeyPrefix = "session:"
	}
	if c.SessionTTL == 0 {
		c.SessionTTL = 30 * 24 * time.Hour
	}
	if c.CleanupInterval == 0 {
		c.CleanupInterval = 5 * time.Minute
	}
}

type Store struct {
	client rueidis.Client
	config StoreConfig
}

func NewStore(cfg StoreConfig) (*Store, error) {
	cfg.setDefaults()

	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{cfg.RedisAddr},
		Password:    cfg.RedisPassword,
		SelectDB:    cfg.RedisDB,
	})
	if err != nil {
		return nil, fmt.Errorf("连接Redis失败: %w", err)
	}

	return &Store{
		client: client,
		config: cfg,
	}, nil
}

func (s *Store) Get(ctx context.Context, sessionID string) (*Session, error) {
	key := s.config.KeyPrefix + sessionID

	data, err := s.client.Do(ctx, s.client.B().Get().Key(key).Build()).ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("获取session失败: %w", err)
	}

	var session Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("解析session失败: %w", err)
	}

	if session.IsExpired() {
		_ = s.Delete(ctx, sessionID)
		return nil, ErrSessionExpired
	}

	return &session, nil
}

func (s *Store) Set(ctx context.Context, session *Session) error {
	key := s.config.KeyPrefix + session.ID

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("序列化session失败: %w", err)
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		ttl = s.config.SessionTTL
	}

	if err := s.client.Do(ctx, s.client.B().Set().Key(key).Value(string(data)).Ex(ttl).Build()).Error(); err != nil {
		return fmt.Errorf("保存session失败: %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, sessionID string) error {
	key := s.config.KeyPrefix + sessionID

	if err := s.client.Do(ctx, s.client.B().Del().Key(key).Build()).Error(); err != nil {
		return fmt.Errorf("删除session失败: %w", err)
	}

	return nil
}

func (s *Store) Refresh(ctx context.Context, sessionID string) error {
	key := s.config.KeyPrefix + sessionID

	if err := s.client.Do(ctx, s.client.B().Expire().Key(key).Seconds(int64(s.config.SessionTTL.Seconds())).Build()).Error(); err != nil {
		return fmt.Errorf("刷新session失败: %w", err)
	}

	return nil
}

// Create 创建新的 session
func (s *Store) Create(ctx context.Context, sessionID string, userInfo *UserInfo, clientIP string) (*Session, error) {
	now := time.Now()
	session := &Session{
		ID:        sessionID,
		UserID:    userInfo.ID,
		UserInfo:  userInfo,
		IP:        clientIP,
		CreatedAt: now,
		ExpiresAt: now.Add(s.config.SessionTTL),
	}

	if err := s.Set(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// Update 更新 session
func (s *Store) Update(ctx context.Context, sessionID string, updateFn func(*Session) error) (*Session, error) {
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if err := updateFn(session); err != nil {
		return nil, fmt.Errorf("更新session失败: %w", err)
	}

	if err := s.Set(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// RefreshTTL 刷新 session 的 TTL
func (s *Store) RefreshTTL(ctx context.Context, sessionID string) error {
	return s.Refresh(ctx, sessionID)
}

// GetByUserID 获取指定用户的所有 sessions
func (s *Store) GetByUserID(ctx context.Context, userID int64) ([]*Session, error) {
	// 这里简化实现，实际应该使用索引
	// 目前返回空列表
	return []*Session{}, nil
}

func (s *Store) Close() {
	s.client.Close()
}
