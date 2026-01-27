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
}

type RedisConfig struct {
	Addr        string        `yaml:"addr"`
	Password    string        `yaml:"password"`
	DB          int           `yaml:"db"`
	CachePrefix string        `yaml:"cache_prefix"`
	CacheTTL    time.Duration `yaml:"cache_ttl"`
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

	return cfg, nil
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
