package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Session  SessionConfig  `yaml:"session"`
	Storage  StorageConfig  `yaml:"storage"`
}

type StorageConfig struct {
	RustFS RustFSConfig `yaml:"rustfs"`
}

type RustFSConfig struct {
	BaseURL       string `yaml:"base_url"`
	Username      string `yaml:"username"`
	Password      string `yaml:"password"`
	MaxFileSize   int64  `yaml:"max_file_size"`    // 字节
	MaxConcurrent int    `yaml:"max_concurrent"`   // 最大并发上传数
}

type RedisConfig struct {
	Addr                   string `yaml:"addr"`
	Password               string `yaml:"password"`
	DB                     int    `yaml:"db"`
	CachePrefix            string `yaml:"cache_prefix"`
	CacheTTLStr            string `yaml:"cache_ttl"`
	SessionPrefix          string `yaml:"session_prefix"`
	SessionTTLStr          string `yaml:"session_ttl"`
	SessionCleanupIntervalStr string `yaml:"session_cleanup_interval"`
	// 解析后的值
	CacheTTL             time.Duration `yaml:"-"`
	SessionTTL           time.Duration `yaml:"-"`
	SessionCleanupInterval time.Duration `yaml:"-"`
}

type SessionConfig struct {
	CookieName     string `yaml:"cookie_name"`
	CookieDomain   string `yaml:"cookie_domain"`
	CookieSecure   bool   `yaml:"cookie_secure"`
	CookieHttpOnly bool   `yaml:"cookie_http_only"`
	CookieSameSite string `yaml:"cookie_same_site"`
}

type AppConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            string `yaml:"port"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Name            string `yaml:"name"`
	SSLMode         string `yaml:"sslmode"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.Username, d.Password, d.Name, d.SSLMode,
	)
}

func (d *DatabaseConfig) ConnMaxLifetimeDuration() time.Duration {
	return time.Duration(d.ConnMaxLifetime) * time.Second
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	raw := make(map[string]interface{})
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	expandValues(raw)

	expandedData, err := yaml.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("序列化配置文件失败: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(expandedData, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	if err := cfg.parseDurations(); err != nil {
		return nil, fmt.Errorf("解析duration失败: %w", err)
	}

	cfg.setDefaults()

	return cfg, nil
}

func (c *Config) parseDurations() error {
	var err error
	
	if c.Redis.CacheTTLStr != "" {
		c.Redis.CacheTTL, err = time.ParseDuration(c.Redis.CacheTTLStr)
		if err != nil {
			return fmt.Errorf("无效的 cache_ttl: %s", c.Redis.CacheTTLStr)
		}
	}
	
	if c.Redis.SessionTTLStr != "" {
		c.Redis.SessionTTL, err = time.ParseDuration(c.Redis.SessionTTLStr)
		if err != nil {
			return fmt.Errorf("无效的 session_ttl: %s", c.Redis.SessionTTLStr)
		}
	}
	
	if c.Redis.SessionCleanupIntervalStr != "" {
		c.Redis.SessionCleanupInterval, err = time.ParseDuration(c.Redis.SessionCleanupIntervalStr)
		if err != nil {
			return fmt.Errorf("无效的 session_cleanup_interval: %s", c.Redis.SessionCleanupIntervalStr)
		}
	}
	
	return nil
}

func (c *Config) setDefaults() {
	if c.Redis.CachePrefix == "" {
		c.Redis.CachePrefix = "aclove:cache:"
	}
	if c.Redis.CacheTTL == 0 {
		c.Redis.CacheTTL = 24 * time.Hour
	}
	if c.Redis.SessionPrefix == "" {
		c.Redis.SessionPrefix = "aclove:session:"
	}
	if c.Redis.SessionTTL == 0 {
		c.Redis.SessionTTL = 30 * 24 * time.Hour
	}
	if c.Redis.SessionCleanupInterval == 0 {
		c.Redis.SessionCleanupInterval = 5 * time.Minute
	}
	if c.Session.CookieName == "" {
		c.Session.CookieName = "aclove_session"
	}
	if c.Session.CookieSameSite == "" {
		c.Session.CookieSameSite = "Lax"
	}
	// Storage defaults
	if c.Storage.RustFS.MaxFileSize == 0 {
		c.Storage.RustFS.MaxFileSize = 50 * 1024 * 1024 // 50MB
	}
	if c.Storage.RustFS.MaxConcurrent == 0 {
		c.Storage.RustFS.MaxConcurrent = 10
	}
}

func expandValues(m map[string]interface{}) {
	for k, v := range m {
		switch val := v.(type) {
		case string:
			m[k] = expandEnv(val)
		case map[string]interface{}:
			expandValues(val)
		case []interface{}:
			for i, item := range val {
				if item, ok := item.(string); ok {
					val[i] = expandEnv(item)
				}
			}
		}
	}
}

func expandEnv(value string) string {
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		content := strings.TrimPrefix(strings.TrimSuffix(value, "}"), "${")
		if idx := strings.Index(content, ":"); idx != -1 {
			envKey := content[:idx]
			defaultVal := content[idx+1:]
			if envValue := os.Getenv(envKey); envValue != "" {
				return envValue
			}
			return defaultVal
		}
		if envValue := os.Getenv(content); envValue != "" {
			return envValue
		}
	}
	return value
}

func mustParseInt(s string) int {
	if s == "" {
		return 0
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}
