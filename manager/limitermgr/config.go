package limitermgr

import (
	"fmt"
	"strconv"
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

	DefaultMemoryMaxBackups = 1000
)

// DefaultConfig 返回默认配置（使用内存限流驱动）
func DefaultConfig() *LimiterConfig {
	return &LimiterConfig{
		Driver: "memory",
		RedisConfig: &RedisLimiterConfig{
			Host:            DefaultRedisHost,
			Port:            DefaultRedisPort,
			DB:              DefaultRedisDB,
			MaxIdleConns:    DefaultRedisMaxIdleConns,
			MaxOpenConns:    DefaultRedisMaxOpenConns,
			ConnMaxLifetime: DefaultRedisConnMaxLifetime,
		},
		MemoryConfig: &MemoryLimiterConfig{
			MaxBackups: DefaultMemoryMaxBackups,
		},
	}
}

// LimiterConfig 限流配置
type LimiterConfig struct {
	Driver       string               `yaml:"driver"`        // 驱动类型: redis, memory
	RedisConfig  *RedisLimiterConfig  `yaml:"redis_config"`  // Redis 配置
	MemoryConfig *MemoryLimiterConfig `yaml:"memory_config"` // Memory 配置
}

// RedisLimiterConfig Redis 限流配置
type RedisLimiterConfig struct {
	Host            string        `yaml:"host"`              // Redis 主机地址
	Port            int           `yaml:"port"`              // Redis 端口
	Password        string        `yaml:"password"`          // Redis 密码
	DB              int           `yaml:"db"`                // Redis 数据库编号
	MaxIdleConns    int           `yaml:"max_idle_conns"`    // 最大空闲连接数
	MaxOpenConns    int           `yaml:"max_open_conns"`    // 最大打开连接数
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"` // 连接最大存活时间
}

// MemoryLimiterConfig 内存限流配置
type MemoryLimiterConfig struct {
	MaxBackups int `yaml:"max_backups"` // 最大备份项数（清理策略相关）
}

// Validate 验证配置的有效性
// 检查驱动类型是否有效、对应的配置是否完整
func (c *LimiterConfig) Validate() error {
	if c.Driver == "" {
		return fmt.Errorf("driver is required")
	}

	// 标准化驱动名称
	c.Driver = strings.ToLower(strings.TrimSpace(c.Driver))

	// 验证驱动类型
	switch c.Driver {
	case "redis", "memory":
		// 有效驱动
	default:
		return fmt.Errorf("unsupported driver: %s (must be redis or memory)", c.Driver)
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

// ParseLimiterConfigFromMap 从 ConfigMap 解析限流配置
// 支持从 map[string]any 格式的配置中解析限流管理器配置
func ParseLimiterConfigFromMap(cfg map[string]any) (*LimiterConfig, error) {
	config := &LimiterConfig{
		Driver: "memory", // 默认使用 memory 驱动
		RedisConfig: &RedisLimiterConfig{
			Host:            DefaultRedisHost,
			Port:            DefaultRedisPort,
			DB:              DefaultRedisDB,
			MaxIdleConns:    DefaultRedisMaxIdleConns,
			MaxOpenConns:    DefaultRedisMaxOpenConns,
			ConnMaxLifetime: DefaultRedisConnMaxLifetime,
		},
		MemoryConfig: &MemoryLimiterConfig{
			MaxBackups: DefaultMemoryMaxBackups,
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
		redisConfig, err := parseRedisLimiterConfig(redisConfigMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse redis_config: %w", err)
		}
		config.RedisConfig = redisConfig
	}

	// 解析 memory_config
	if memoryConfigMap, ok := cfg["memory_config"].(map[string]any); ok {
		memoryConfig, err := parseMemoryLimiterConfig(memoryConfigMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse memory_config: %w", err)
		}
		config.MemoryConfig = memoryConfig
	}

	return config, nil
}

// parseRedisLimiterConfig 解析 Redis 限流配置
// 从 map 中解析 Redis 连接相关的配置项
func parseRedisLimiterConfig(cfg map[string]any) (*RedisLimiterConfig, error) {
	config := &RedisLimiterConfig{
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

	// 解析 port（资源寻址参数：非法即错，不再静默回落默认端口）
	if port, ok := cfg["port"]; ok {
		p, err := parsePortField(port)
		if err != nil {
			return nil, err
		}
		config.Port = p
	}

	// 解析 password
	if password, ok := cfg["password"].(string); ok {
		config.Password = password
	}

	// 解析 db（资源寻址参数：非负整数透传、不设 15 上限——为实例 databases>16 扩容预留；负数或不可解析即错）
	if db, ok := cfg["db"]; ok {
		d, err := parseDBField(db)
		if err != nil {
			return nil, err
		}
		config.DB = d
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

// parseMemoryLimiterConfig 解析内存限流配置
// 从 map 中解析内存限流相关的配置项
func parseMemoryLimiterConfig(cfg map[string]any) (*MemoryLimiterConfig, error) {
	config := &MemoryLimiterConfig{
		MaxBackups: DefaultMemoryMaxBackups,
	}

	// 解析 max_backups
	if maxBackups, ok := cfg["max_backups"]; ok {
		if v, ok := toInt(maxBackups); ok && v >= 0 {
			config.MaxBackups = v
		}
	}

	return config, nil
}

// toInt 将任意类型转换为 int
// 支持 int、int64、float64 类型的转换
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

// parseDuration 解析时间长度
// 支持 int/int64/float64（秒）、字符串（支持 Go duration 格式或纯数字秒数）
func parseDuration(v any) (time.Duration, error) {
	switch val := v.(type) {
	case int:
		return time.Duration(val) * time.Second, nil
	case int64:
		return time.Duration(val) * time.Second, nil
	case float64:
		return time.Duration(val) * time.Second, nil
	case string:
		s := strings.TrimSpace(val)
		if s == "" {
			return 0, fmt.Errorf("empty duration string")
		}

		if duration, err := time.ParseDuration(s); err == nil {
			return duration, nil
		}

		if num, err := parseDurationSeconds(s); err == nil {
			return time.Duration(num) * time.Second, nil
		}

		return 0, fmt.Errorf("invalid duration format: %s", val)
	default:
		return 0, fmt.Errorf("unsupported duration type: %T", v)
	}
}

// parseDurationSeconds 解析秒数形式的字符串
// 将纯数字字符串转换为整数秒数
func parseDurationSeconds(s string) (int, error) {
	s = strings.TrimSpace(s)
	var num int
	if _, err := fmt.Sscanf(s, "%d", &num); err != nil {
		return 0, fmt.Errorf("invalid duration: %s", s)
	}
	return num, nil
}

// toIntStrict 将配置值严格解析为整数：支持 int/int64/整数值 float64/纯数字字符串；其余一律不认
func toIntStrict(v any) (int, bool) {
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
	case string:
		s := strings.TrimSpace(val)
		var n int
		if _, err := fmt.Sscanf(s, "%d", &n); err == nil && strconv.Itoa(n) == s {
			return n, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// parsePortField 解析资源寻址参数 port：越界或不可解析即错（不再静默回落默认端口，
// 否则错误配置会被指向错误地址）。limiter.redis_config 虽不构建客户端，解析口径必须与 cache 一致，
// 消除「配置说谎」温床。错误信息含配置键路径与原值，便于第一现场定位。
func parsePortField(v any) (int, error) {
	n, ok := toIntStrict(v)
	if !ok {
		return 0, fmt.Errorf("limiter.redis_config.port: %v（类型 %T）不可解析为端口（须为 1~65535 的整数）；"+
			"资源寻址参数非法即拒绝启动，不再静默回落默认值（修复指引：修正配置文件中的 limiter.redis_config.port）", v, v)
	}
	if n <= 0 || n > 65535 {
		return 0, fmt.Errorf("limiter.redis_config.port: %v 超出端口合法范围（1~65535）；"+
			"资源寻址参数非法即拒绝启动，不再静默回落默认值（修复指引：修正配置文件中的 limiter.redis_config.port）", v)
	}
	return n, nil
}

// parseDBField 解析资源寻址参数 db：非负整数透传、不设 15 上限（实例 databases>16 扩容无需改代码，
// 真越界由构造期 Ping 与 go-redis 连接期 SELECT 天然 fail-fast）；负数或不可解析即错。
func parseDBField(v any) (int, error) {
	n, ok := toIntStrict(v)
	if !ok {
		return 0, fmt.Errorf("limiter.redis_config.db: %v（类型 %T）不可解析为 Redis 库号（须为非负整数）；"+
			"资源寻址参数非法即拒绝启动，不再静默回落默认值（修复指引：合法取值以 docs/ops/middlewares/middleware-redis.md 分配表为准）", v, v)
	}
	if n < 0 {
		return 0, fmt.Errorf("limiter.redis_config.db: %v 为负数，不是合法的 Redis 库号；"+
			"资源寻址参数非法即拒绝启动，不再静默回落默认值（修复指引：合法取值以 docs/ops/middlewares/middleware-redis.md 分配表为准）", v)
	}
	return n, nil
}
