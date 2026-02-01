package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/rueidis"
)

type Cache struct {
	client    rueidis.Client
	prefix    string
	defaultTT time.Duration
}

type Config struct {
	Addr        string
	Password    string
	DB          int
	Prefix      string
	DefaultTTLD time.Duration
}

func New(cfg Config) (*Cache, error) {
	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{cfg.Addr},
		Password:    cfg.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("连接Redis失败: %w", err)
	}

	ttl := cfg.DefaultTTLD
	if ttl == 0 {
		ttl = 24 * time.Hour
	}

	return &Cache{
		client:    client,
		prefix:    cfg.Prefix,
		defaultTT: ttl,
	}, nil
}

func (c *Cache) key(categoryID int64) string {
	return fmt.Sprintf("%scategory:%d", c.prefix, categoryID)
}

func (c *Cache) Get(ctx context.Context, categoryID int64, dest interface{}) error {
	cmd := c.client.B().Get().Key(c.key(categoryID)).Build()
	res := c.client.Do(ctx, cmd)
	if err := res.Error(); err != nil {
		if err == rueidis.Nil {
			return nil
		}
		return fmt.Errorf("获取缓存失败: %w", err)
	}

	val, err := res.ToString()
	if err != nil {
		return fmt.Errorf("获取缓存值失败: %w", err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("反序列化缓存失败: %w", err)
	}

	return nil
}

func (c *Cache) Set(ctx context.Context, categoryID int64, value interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.defaultTT
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化缓存失败: %w", err)
	}

	cmd := c.client.B().Set().Key(c.key(categoryID)).Value(string(data)).Ex(ttl).Build()
	if err := c.client.Do(ctx, cmd).Error(); err != nil {
		return fmt.Errorf("设置缓存失败: %w", err)
	}

	return nil
}

func (c *Cache) Delete(ctx context.Context, categoryID int64) error {
	cmd := c.client.B().Del().Key(c.key(categoryID)).Build()
	if err := c.client.Do(ctx, cmd).Error(); err != nil {
		return fmt.Errorf("删除缓存失败: %w", err)
	}

	return nil
}

func (c *Cache) Close() {
	c.client.Close()
}
