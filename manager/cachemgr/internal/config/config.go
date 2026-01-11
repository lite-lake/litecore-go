package config

import (
	"fmt"
	"strings"
	"time"
)

const (
	// 默认值
	DefaultRedisHost            = "localhost"
	DefaultRedisPort            = 6379
	DefaultRedisDB              = 0
	DefaultRedisMaxIdleConns    = 10
	DefaultRedisMaxOpenConns    = 100
	DefaultRedisConnMaxLifetime = 30 * time.Second

	DefaultMemoryMaxSize    = 100                 // MB
	DefaultMemoryMaxAge     = 30 * 24 * time.Hour // 30 天
	DefaultMemoryMaxBackups = 1000
	DefaultMemoryCompress   = false
)

// DefaultConfig 返回默认配置（使用内存缓存驱动）
func DefaultConfig() *CacheConfig {
	return &CacheConfig{
		Driver: "memory",
		RedisConfig: &RedisConfig{
			Host:            DefaultRedisHost,
			Port:            DefaultRedisPort,
			DB:              DefaultRedisDB,
			MaxIdleConns:    DefaultRedisMaxIdleConns,
			MaxOpenConns:    DefaultRedisMaxOpenConns,
			ConnMaxLifetime: DefaultRedisConnMaxLifetime,
		},
		MemoryConfig: &MemoryConfig{
			MaxSize:    DefaultMemoryMaxSize,
			MaxAge:     DefaultMemoryMaxAge,
			MaxBackups: DefaultMemoryMaxBackups,
			Compress:   DefaultMemoryCompress,
		},
	}
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Driver       string        `yaml:"driver"`        // 驱动类型: redis, memory, none
	RedisConfig  *RedisConfig  `yaml:"redis_config"`  // Redis 配置
	MemoryConfig *MemoryConfig `yaml:"memory_config"` // Memory 配置
}

// RedisConfig Redis 缓存配置
type RedisConfig struct {
	Host            string        `yaml:"host"`              // Redis 主机地址
	Port            int           `yaml:"port"`              // Redis 端口
	Password        string        `yaml:"password"`          // Redis 密码
	DB              int           `yaml:"db"`                // Redis 数据库编号
	MaxIdleConns    int           `yaml:"max_idle_conns"`    // 最大空闲连接数
	MaxOpenConns    int           `yaml:"max_open_conns"`    // 最大打开连接数
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"` // 连接最大存活时间
}

// MemoryConfig 内存缓存配置
type MemoryConfig struct {
	MaxSize    int           `yaml:"max_size"`    // 最大缓存大小（MB）
	MaxAge     time.Duration `yaml:"max_age"`     // 最大缓存时间
	MaxBackups int           `yaml:"max_backups"` // 最大备份项数（清理策略相关）
	Compress   bool          `yaml:"compress"`    // 是否压缩
}

// Validate 验证配置
func (c *CacheConfig) Validate() error {
	if c.Driver == "" {
		return fmt.Errorf("driver is required")
	}

	// 标准化驱动名称
	c.Driver = strings.ToLower(strings.TrimSpace(c.Driver))

	// 验证驱动类型
	switch c.Driver {
	case "redis", "memory", "none":
		// 有效驱动
	default:
		return fmt.Errorf("unsupported driver: %s (must be redis, memory, or none)", c.Driver)
	}

	// Redis 驱动需要 Redis 配置
	if c.Driver == "redis" && c.RedisConfig == nil {
		return fmt.Errorf("redis_config is required when driver is redis")
	}

	// Memory 驱动需要 Memory 配置
	if c.Driver == "memory" && c.MemoryConfig == nil {
		return fmt.Errorf("memory_config is required when driver is memory")
	}

	return nil
}

// ParseCacheConfigFromMap 从 ConfigMap 解析缓存配置
func ParseCacheConfigFromMap(cfg map[string]any) (*CacheConfig, error) {
	config := &CacheConfig{
		Driver: "none", // 默认使用 none 驱动
		RedisConfig: &RedisConfig{
			Host:            DefaultRedisHost,
			Port:            DefaultRedisPort,
			DB:              DefaultRedisDB,
			MaxIdleConns:    DefaultRedisMaxIdleConns,
			MaxOpenConns:    DefaultRedisMaxOpenConns,
			ConnMaxLifetime: DefaultRedisConnMaxLifetime,
		},
		MemoryConfig: &MemoryConfig{
			MaxSize:    DefaultMemoryMaxSize,
			MaxAge:     DefaultMemoryMaxAge,
			MaxBackups: DefaultMemoryMaxBackups,
			Compress:   DefaultMemoryCompress,
		},
	}

	if cfg == nil {
		return config, nil
	}

	// 解析 driver
	if driver, ok := cfg["driver"].(string); ok {
		config.Driver = strings.ToLower(strings.TrimSpace(driver))
	}

	// 解析 redis_config
	if redisConfigMap, ok := cfg["redis_config"].(map[string]any); ok {
		redisConfig, err := parseRedisConfig(redisConfigMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse redis_config: %w", err)
		}
		config.RedisConfig = redisConfig
	}

	// 解析 memory_config
	if memoryConfigMap, ok := cfg["memory_config"].(map[string]any); ok {
		memoryConfig, err := parseMemoryConfig(memoryConfigMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse memory_config: %w", err)
		}
		config.MemoryConfig = memoryConfig
	}

	return config, nil
}

// parseRedisConfig 解析 Redis 配置
func parseRedisConfig(cfg map[string]any) (*RedisConfig, error) {
	config := &RedisConfig{
		Host:            DefaultRedisHost,
		Port:            DefaultRedisPort,
		DB:              DefaultRedisDB,
		MaxIdleConns:    DefaultRedisMaxIdleConns,
		MaxOpenConns:    DefaultRedisMaxOpenConns,
		ConnMaxLifetime: DefaultRedisConnMaxLifetime,
	}

	// 解析 host
	if host, ok := cfg["host"].(string); ok {
		config.Host = strings.TrimSpace(host)
	}

	// 解析 port
	if port, ok := cfg["port"]; ok {
		switch v := port.(type) {
		case int:
			config.Port = v
		case int64:
			if v > 0 && v <= 65535 {
				config.Port = int(v)
			}
		case float64:
			if v > 0 && v <= 65535 && v == float64(int64(v)) {
				config.Port = int(v)
			}
		case string:
			// 字符串形式的端口号
			var portNum int
			if _, err := fmt.Sscanf(v, "%d", &portNum); err == nil {
				config.Port = portNum
			}
		}
	}

	// 解析 password
	if password, ok := cfg["password"].(string); ok {
		config.Password = password
	}

	// 解析 db
	if db, ok := cfg["db"]; ok {
		switch v := db.(type) {
		case int:
			if v >= 0 && v <= 15 {
				config.DB = v
			}
		case int64:
			if v >= 0 && v <= 15 {
				config.DB = int(v)
			}
		case float64:
			if v >= 0 && v <= 15 && v == float64(int64(v)) {
				config.DB = int(v)
			}
		}
	}

	// 解析 max_idle_conns
	if maxIdle, ok := cfg["max_idle_conns"]; ok {
		if v, ok := toInt(maxIdle); ok && v > 0 {
			config.MaxIdleConns = v
		}
	}

	// 解析 max_open_conns
	if maxOpen, ok := cfg["max_open_conns"]; ok {
		if v, ok := toInt(maxOpen); ok && v > 0 {
			config.MaxOpenConns = v
		}
	}

	// 解析 conn_max_lifetime
	if connMaxLifetime, ok := cfg["conn_max_lifetime"]; ok {
		if duration, err := parseDuration(connMaxLifetime); err == nil {
			config.ConnMaxLifetime = duration
		}
	}

	return config, nil
}

// parseMemoryConfig 解析 Memory 配置
func parseMemoryConfig(cfg map[string]any) (*MemoryConfig, error) {
	config := &MemoryConfig{
		MaxSize:    DefaultMemoryMaxSize,
		MaxAge:     DefaultMemoryMaxAge,
		MaxBackups: DefaultMemoryMaxBackups,
		Compress:   DefaultMemoryCompress,
	}

	// 解析 max_size
	if maxSize, ok := cfg["max_size"]; ok {
		if v, ok := toInt(maxSize); ok && v > 0 {
			config.MaxSize = v
		}
	}

	// 解析 max_age
	if maxAge, ok := cfg["max_age"]; ok {
		if duration, err := parseDuration(maxAge); err == nil && duration > 0 {
			config.MaxAge = duration
		}
	}

	// 解析 max_backups
	if maxBackups, ok := cfg["max_backups"]; ok {
		if v, ok := toInt(maxBackups); ok && v >= 0 {
			config.MaxBackups = v
		}
	}

	// 解析 compress
	if compress, ok := cfg["compress"].(bool); ok {
		config.Compress = compress
	}

	return config, nil
}

// toInt 将任意类型转换为 int
func toInt(v any) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float64:
		if val == float64(int64(val)) {
			return int(val), true
		}
		return 0, false
	default:
		return 0, false
	}
}

// parseDuration 解析时间长度，支持 int, int64, float64 和字符串
func parseDuration(v any) (time.Duration, error) {
	switch val := v.(type) {
	case int:
		return time.Duration(val) * time.Second, nil
	case int64:
		return time.Duration(val) * time.Second, nil
	case float64:
		return time.Duration(val) * time.Second, nil
	case string:
		// 尝试解析字符串形式的时间长度
		// 支持: "30s", "5m", "1h", "30" (默认为秒)
		s := strings.TrimSpace(val)
		if s == "" {
			return 0, fmt.Errorf("empty duration string")
		}

		// 先尝试解析标准时间格式（如 "5m", "1h"）
		if duration, err := time.ParseDuration(s); err == nil {
			return duration, nil
		}

		// 如果失败，尝试解析纯数字（默认为秒）
		if num, err := parseDurationSeconds(s); err == nil {
			return time.Duration(num) * time.Second, nil
		}

		return 0, fmt.Errorf("invalid duration format: %s", val)
	default:
		return 0, fmt.Errorf("unsupported duration type: %T", v)
	}
}

// parseDurationSeconds 解析秒数形式的字符串
func parseDurationSeconds(s string) (int, error) {
	s = strings.TrimSpace(s)
	var num int
	if _, err := fmt.Sscanf(s, "%d", &num); err != nil {
		return 0, fmt.Errorf("invalid duration: %s", s)
	}
	return num, nil
}
