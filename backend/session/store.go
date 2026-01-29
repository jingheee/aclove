package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/rueidis"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusBanned   Status = "banned"
	StatusCooldown Status = "cooldown"
)

type UserInfo struct {
	ID              int64      `json:"id"`
	Cookie          string     `json:"cookie"`
	FingerprintHash string     `json:"fingerprint_hash"`
	IP              string     `json:"ip"`
	Status          Status     `json:"status"`
	StatusReason    string     `json:"status_reason,omitempty"`
	BannedUntil     *time.Time `json:"banned_until,omitempty"`
	CooldownUntil   *time.Time `json:"cooldown_until,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
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
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	UserInfo  *UserInfo `json:"user_info,omitempty"`
	IP        string    `json:"ip"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
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

func (s *Store) sessionKey(sessionID string) string {
	return fmt.Sprintf("%s%s", s.config.KeyPrefix, sessionID)
}

func (s *Store) userKey(userID int64) string {
	return fmt.Sprintf("%suser:%d", s.config.KeyPrefix, userID)
}

func (s *Store) Create(ctx context.Context, sessionID string, userInfo *UserInfo, clientIP string) (*Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID不能为空")
	}

	session := &Session{
		ID:        sessionID,
		UserID:    userInfo.ID,
		UserInfo:  userInfo,
		IP:        clientIP,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(s.config.SessionTTL),
	}

	data, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("序列化session失败: %w", err)
	}

	ttl := s.config.SessionTTL
	cmd := s.client.B().Set().Key(s.sessionKey(sessionID)).Value(string(data)).Ex(ttl).Build()
	if err := s.client.Do(ctx, cmd).Error(); err != nil {
		return nil, fmt.Errorf("保存session到Redis失败: %w", err)
	}

	userSessionKey := s.userKey(userInfo.ID)
	userCmd := s.client.B().Hset().Key(userSessionKey).FieldValue().FieldValue(sessionID, string(data)).Build()
	if err := s.client.Do(ctx, userCmd).Error(); err != nil {
		return nil, fmt.Errorf("保存用户session索引失败: %w", err)
	}

	expireCmd := s.client.B().Expire().Key(userSessionKey).Seconds(int64(ttl.Seconds())).Build()
	s.client.Do(ctx, expireCmd)

	return session, nil
}

func (s *Store) Get(ctx context.Context, sessionID string) (*Session, error) {
	if sessionID == "" {
		return nil, ErrSessionNotFound
	}

	cmd := s.client.B().Get().Key(s.sessionKey(sessionID)).Build()
	res := s.client.Do(ctx, cmd)

	if err := res.Error(); err != nil {
		if err == rueidis.Nil {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("从Redis获取session失败: %w", err)
	}

	val, err := res.ToString()
	if err != nil {
		return nil, fmt.Errorf("读取session值失败: %w", err)
	}

	var session Session
	if err := json.Unmarshal([]byte(val), &session); err != nil {
		return nil, fmt.Errorf("反序列化session失败: %w", err)
	}

	if session.IsExpired() {
		s.Delete(ctx, sessionID)
		return nil, ErrSessionNotFound
	}

	return &session, nil
}

func (s *Store) GetByUserID(ctx context.Context, userID int64) ([]*Session, error) {
	userKey := s.userKey(userID)
	cmd := s.client.B().Hgetall().Key(userKey).Build()
	res := s.client.Do(ctx, cmd)

	if err := res.Error(); err != nil {
		if err == rueidis.Nil {
			return []*Session{}, nil
		}
		return nil, fmt.Errorf("获取用户session列表失败: %w", err)
	}

	fields, err := res.AsStrMap()
	if err != nil {
		return nil, fmt.Errorf("解析session数据失败: %w", err)
	}

	sessions := make([]*Session, 0, len(fields))
	for _, val := range fields {
		var session Session
		if err := json.Unmarshal([]byte(val), &session); err != nil {
			continue
		}
		if !session.IsExpired() {
			sessions = append(sessions, &session)
		}
	}

	return sessions, nil
}

func (s *Store) Update(ctx context.Context, sessionID string, updateFn func(*Session) error) (*Session, error) {
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if err := updateFn(session); err != nil {
		return nil, fmt.Errorf("更新session失败: %w", err)
	}

	data, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("序列化session失败: %w", err)
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		ttl = s.config.SessionTTL
		session.ExpiresAt = time.Now().Add(ttl)
	}

	cmd := s.client.B().Set().Key(s.sessionKey(sessionID)).Value(string(data)).Ex(ttl).Build()
	if err := s.client.Do(ctx, cmd).Error(); err != nil {
		return nil, fmt.Errorf("更新session到Redis失败: %w", err)
	}

	userKey := s.userKey(session.UserID)
	userCmd := s.client.B().Hset().Key(userKey).FieldValue().FieldValue(sessionID, string(data)).Build()
	s.client.Do(ctx, userCmd)

	return session, nil
}

func (s *Store) Delete(ctx context.Context, sessionID string) error {
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		if err == ErrSessionNotFound {
			return nil
		}
		return err
	}

	cmd := s.client.B().Del().Key(s.sessionKey(sessionID)).Build()
	if err := s.client.Do(ctx, cmd).Error(); err != nil {
		return fmt.Errorf("删除session失败: %w", err)
	}

	userKey := s.userKey(session.UserID)
	hdelCmd := s.client.B().Hdel().Key(userKey).Field(sessionID).Build()
	s.client.Do(ctx, hdelCmd)

	return nil
}

func (s *Store) DeleteByUserID(ctx context.Context, userID int64) error {
	sessions, err := s.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	keys := make([]string, 0, len(sessions))
	for _, session := range sessions {
		keys = append(keys, s.sessionKey(session.ID))
	}

	if len(keys) > 0 {
		cmd := s.client.B().Del().Key(keys...).Build()
		if err := s.client.Do(ctx, cmd).Error(); err != nil {
			return fmt.Errorf("批量删除session失败: %w", err)
		}
	}

	userKey := s.userKey(userID)
	delCmd := s.client.B().Del().Key(userKey).Build()
	s.client.Do(ctx, delCmd)

	return nil
}

func (s *Store) RefreshTTL(ctx context.Context, sessionID string) error {
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	session.ExpiresAt = time.Now().Add(s.config.SessionTTL)
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("序列化session失败: %w", err)
	}

	cmd := s.client.B().Set().Key(s.sessionKey(sessionID)).Value(string(data)).Ex(s.config.SessionTTL).Build()
	if err := s.client.Do(ctx, cmd).Error(); err != nil {
		return fmt.Errorf("刷新session TTL失败: %w", err)
	}

	return nil
}

func (s *Store) Close() error {
	s.client.Close()
	return nil
}
