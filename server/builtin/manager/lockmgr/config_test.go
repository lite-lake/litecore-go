package lockmgr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	t.Run("默认驱动为 memory", func(t *testing.T) {
		config := DefaultConfig()
		assert.Equal(t, "memory", config.Driver)
	})

	t.Run("Redis 默认配置值", func(t *testing.T) {
		config := DefaultConfig()
		assert.NotNil(t, config.RedisConfig)
		assert.Equal(t, DefaultRedisHost, config.RedisConfig.Host)
		assert.Equal(t, DefaultRedisPort, config.RedisConfig.Port)
		assert.Equal(t, DefaultRedisDB, config.RedisConfig.DB)
		assert.Equal(t, DefaultRedisMaxIdleConns, config.RedisConfig.MaxIdleConns)
		assert.Equal(t, DefaultRedisMaxOpenConns, config.RedisConfig.MaxOpenConns)
		assert.Equal(t, DefaultRedisConnMaxLifetime, config.RedisConfig.ConnMaxLifetime)
	})

	t.Run("Memory 默认配置值", func(t *testing.T) {
		config := DefaultConfig()
		assert.NotNil(t, config.MemoryConfig)
		assert.Equal(t, DefaultMemoryMaxBackups, config.MemoryConfig.MaxBackups)
	})
}

func TestLockConfig_Validate(t *testing.T) {
	t.Run("有效驱动 redis", func(t *testing.T) {
		config := &LockConfig{
			Driver:      "redis",
			RedisConfig: &RedisLockConfig{},
		}
		err := config.Validate()
		assert.NoError(t, err)
		assert.Equal(t, "redis", config.Driver)
	})

	t.Run("有效驱动 memory", func(t *testing.T) {
		config := &LockConfig{
			Driver:       "memory",
			MemoryConfig: &MemoryLockConfig{},
		}
		err := config.Validate()
		assert.NoError(t, err)
		assert.Equal(t, "memory", config.Driver)
	})

	t.Run("空驱动错误", func(t *testing.T) {
		config := &LockConfig{}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "driver is required")
	})

	t.Run("不支持的驱动错误", func(t *testing.T) {
		config := &LockConfig{
			Driver: "invalid",
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported driver")
		assert.Contains(t, err.Error(), "invalid")
	})

	t.Run("Redis 驱动缺少配置错误", func(t *testing.T) {
		config := &LockConfig{
			Driver: "redis",
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "redis_config is required")
	})

	t.Run("Memory 驱动缺少配置错误", func(t *testing.T) {
		config := &LockConfig{
			Driver: "memory",
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "memory_config is required")
	})

	t.Run("驱动名称被标准化为小写", func(t *testing.T) {
		config := &LockConfig{
			Driver:      "REDIS",
			RedisConfig: &RedisLockConfig{},
		}
		err := config.Validate()
		assert.NoError(t, err)
		assert.Equal(t, "redis", config.Driver)
	})

	t.Run("驱动名称带空格会被去除", func(t *testing.T) {
		config := &LockConfig{
			Driver:       "  memory  ",
			MemoryConfig: &MemoryLockConfig{},
		}
		err := config.Validate()
		assert.NoError(t, err)
		assert.Equal(t, "memory", config.Driver)
	})
}

func TestParseLockConfigFromMap(t *testing.T) {
	t.Run("空配置返回默认值", func(t *testing.T) {
		config, err := ParseLockConfigFromMap(nil)
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "memory", config.Driver)
		assert.NotNil(t, config.RedisConfig)
		assert.NotNil(t, config.MemoryConfig)
	})

	t.Run("空 map 返回默认值", func(t *testing.T) {
		config, err := ParseLockConfigFromMap(map[string]any{})
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "memory", config.Driver)
	})

	t.Run("成功解析完整配置", func(t *testing.T) {
		cfgMap := map[string]any{
			"driver": "redis",
			"redis_config": map[string]any{
				"host":     "127.0.0.1",
				"port":     6380,
				"password": "testpass",
				"db":       1,
			},
			"memory_config": map[string]any{
				"max_backups": 2000,
			},
		}
		config, err := ParseLockConfigFromMap(cfgMap)
		assert.NoError(t, err)
		assert.Equal(t, "redis", config.Driver)
		assert.Equal(t, "127.0.0.1", config.RedisConfig.Host)
		assert.Equal(t, 6380, config.RedisConfig.Port)
		assert.Equal(t, "testpass", config.RedisConfig.Password)
		assert.Equal(t, 1, config.RedisConfig.DB)
		assert.Equal(t, 2000, config.MemoryConfig.MaxBackups)
	})

	t.Run("解析 redis 驱动类型", func(t *testing.T) {
		cfgMap := map[string]any{
			"driver":       "redis",
			"redis_config": map[string]any{},
		}
		config, err := ParseLockConfigFromMap(cfgMap)
		assert.NoError(t, err)
		assert.Equal(t, "redis", config.Driver)
	})

	t.Run("解析 memory 驱动类型", func(t *testing.T) {
		cfgMap := map[string]any{
			"driver":        "memory",
			"memory_config": map[string]any{},
		}
		config, err := ParseLockConfigFromMap(cfgMap)
		assert.NoError(t, err)
		assert.Equal(t, "memory", config.Driver)
	})

	t.Run("解析 Redis 配置", func(t *testing.T) {
		cfgMap := map[string]any{
			"redis_config": map[string]any{
				"host":              "redis.example.com",
				"port":              6379,
				"password":          "secret",
				"db":                5,
				"max_idle_conns":    20,
				"max_open_conns":    200,
				"conn_max_lifetime": 60,
			},
		}
		config, err := ParseLockConfigFromMap(cfgMap)
		assert.NoError(t, err)
		assert.Equal(t, "redis.example.com", config.RedisConfig.Host)
		assert.Equal(t, 6379, config.RedisConfig.Port)
		assert.Equal(t, "secret", config.RedisConfig.Password)
		assert.Equal(t, 5, config.RedisConfig.DB)
		assert.Equal(t, 20, config.RedisConfig.MaxIdleConns)
		assert.Equal(t, 200, config.RedisConfig.MaxOpenConns)
		assert.Equal(t, 60*time.Second, config.RedisConfig.ConnMaxLifetime)
	})

	t.Run("解析 Memory 配置", func(t *testing.T) {
		cfgMap := map[string]any{
			"memory_config": map[string]any{
				"max_backups": 5000,
			},
		}
		config, err := ParseLockConfigFromMap(cfgMap)
		assert.NoError(t, err)
		assert.Equal(t, 5000, config.MemoryConfig.MaxBackups)
	})
}

func TestParseRedisLockConfig(t *testing.T) {
	t.Run("默认值", func(t *testing.T) {
		config, err := parseRedisLockConfig(map[string]any{})
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, DefaultRedisHost, config.Host)
		assert.Equal(t, DefaultRedisPort, config.Port)
		assert.Equal(t, DefaultRedisDB, config.DB)
		assert.Equal(t, DefaultRedisMaxIdleConns, config.MaxIdleConns)
		assert.Equal(t, DefaultRedisMaxOpenConns, config.MaxOpenConns)
		assert.Equal(t, DefaultRedisConnMaxLifetime, config.ConnMaxLifetime)
	})

	t.Run("解析 int 类型 port", func(t *testing.T) {
		cfg := map[string]any{
			"port": 6380,
		}
		config, err := parseRedisLockConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, 6380, config.Port)
	})

	t.Run("解析 int64 类型 port", func(t *testing.T) {
		cfg := map[string]any{
			"port": int64(6381),
		}
		config, err := parseRedisLockConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, 6381, config.Port)
	})

	t.Run("解析 float64 类型 port", func(t *testing.T) {
		cfg := map[string]any{
			"port": float64(6382),
		}
		config, err := parseRedisLockConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, 6382, config.Port)
	})

	t.Run("解析 string 类型 port", func(t *testing.T) {
		cfg := map[string]any{
			"port": "6383",
		}
		config, err := parseRedisLockConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, 6383, config.Port)
	})

	t.Run("端口号边界测试_最大值", func(t *testing.T) {
		cfg := map[string]any{
			"port": 65535,
		}
		config, err := parseRedisLockConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, 65535, config.Port)
	})

	t.Run("端口号边界测试_超出最大值", func(t *testing.T) {
		cfg := map[string]any{
			"port": 65536,
		}
		config, err := parseRedisLockConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, DefaultRedisPort, config.Port)
	})

	t.Run("端口号边界测试_负数保持默认", func(t *testing.T) {
		cfg := map[string]any{
			"port": -1,
		}
		config, err := parseRedisLockConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, DefaultRedisPort, config.Port)
	})

	t.Run("DB 编号范围测试_有效值 0", func(t *testing.T) {
		cfg := map[string]any{
			"db": 0,
		}
		config, err := parseRedisLockConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, 0, config.DB)
	})

	t.Run("DB 编号范围测试_有效值 15", func(t *testing.T) {
		cfg := map[string]any{
			"db": 15,
		}
		config, err := parseRedisLockConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, 15, config.DB)
	})

	t.Run("DB 编号范围测试_超出最大值保持默认", func(t *testing.T) {
		cfg := map[string]any{
			"db": 16,
		}
		config, err := parseRedisLockConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, DefaultRedisDB, config.DB)
	})

	t.Run("DB 编号范围测试_负数保持默认", func(t *testing.T) {
		cfg := map[string]any{
			"db": -1,
		}
		config, err := parseRedisLockConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, DefaultRedisDB, config.DB)
	})

	t.Run("max_idle_conns 参数解析", func(t *testing.T) {
		cfg := map[string]any{
			"max_idle_conns": 50,
		}
		config, err := parseRedisLockConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, 50, config.MaxIdleConns)
	})

	t.Run("max_idle_conns 为 0 保持默认", func(t *testing.T) {
		cfg := map[string]any{
			"max_idle_conns": 0,
		}
		config, err := parseRedisLockConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, DefaultRedisMaxIdleConns, config.MaxIdleConns)
	})

	t.Run("max_idle_conns 为负数保持默认", func(t *testing.T) {
		cfg := map[string]any{
			"max_idle_conns": -10,
		}
		config, err := parseRedisLockConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, DefaultRedisMaxIdleConns, config.MaxIdleConns)
	})

	t.Run("max_open_conns 参数解析", func(t *testing.T) {
		cfg := map[string]any{
			"max_open_conns": 200,
		}
		config, err := parseRedisLockConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, 200, config.MaxOpenConns)
	})

	t.Run("conn_max_lifetime 时间解析", func(t *testing.T) {
		cfg := map[string]any{
			"conn_max_lifetime": 120,
		}
		config, err := parseRedisLockConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, 120*time.Second, config.ConnMaxLifetime)
	})

	t.Run("host 参数解析", func(t *testing.T) {
		cfg := map[string]any{
			"host": "  redis.example.com  ",
		}
		config, err := parseRedisLockConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, "redis.example.com", config.Host)
	})

	t.Run("password 参数解析", func(t *testing.T) {
		cfg := map[string]any{
			"password": "mypassword",
		}
		config, err := parseRedisLockConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, "mypassword", config.Password)
	})
}

func TestParseMemoryLockConfig(t *testing.T) {
	t.Run("默认值", func(t *testing.T) {
		config, err := parseMemoryLockConfig(map[string]any{})
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, DefaultMemoryMaxBackups, config.MaxBackups)
	})

	t.Run("max_backups 参数解析", func(t *testing.T) {
		cfg := map[string]any{
			"max_backups": 3000,
		}
		config, err := parseMemoryLockConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, 3000, config.MaxBackups)
	})

	t.Run("max_backups 为 0", func(t *testing.T) {
		cfg := map[string]any{
			"max_backups": 0,
		}
		config, err := parseMemoryLockConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, 0, config.MaxBackups)
	})

	t.Run("max_backups 为负数保持默认", func(t *testing.T) {
		cfg := map[string]any{
			"max_backups": -100,
		}
		config, err := parseMemoryLockConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, DefaultMemoryMaxBackups, config.MaxBackups)
	})
}

func TestToInt(t *testing.T) {
	t.Run("int 类型转换", func(t *testing.T) {
		v, ok := toInt(42)
		assert.True(t, ok)
		assert.Equal(t, 42, v)
	})

	t.Run("int64 类型转换", func(t *testing.T) {
		v, ok := toInt(int64(100))
		assert.True(t, ok)
		assert.Equal(t, 100, v)
	})

	t.Run("float64 类型转换_整数", func(t *testing.T) {
		v, ok := toInt(float64(50))
		assert.True(t, ok)
		assert.Equal(t, 50, v)
	})

	t.Run("float64 类型转换_浮点数", func(t *testing.T) {
		v, ok := toInt(3.14)
		assert.False(t, ok)
		assert.Equal(t, 0, v)
	})

	t.Run("不支持的类型", func(t *testing.T) {
		v, ok := toInt("string")
		assert.False(t, ok)
		assert.Equal(t, 0, v)

		v, ok = toInt(nil)
		assert.False(t, ok)
		assert.Equal(t, 0, v)
	})
}

func TestParseDuration(t *testing.T) {
	t.Run("int 类型转换为秒", func(t *testing.T) {
		duration, err := parseDuration(60)
		assert.NoError(t, err)
		assert.Equal(t, 60*time.Second, duration)
	})

	t.Run("int64 类型转换为秒", func(t *testing.T) {
		duration, err := parseDuration(int64(120))
		assert.NoError(t, err)
		assert.Equal(t, 120*time.Second, duration)
	})

	t.Run("float64 类型转换为秒", func(t *testing.T) {
		duration, err := parseDuration(30.5)
		assert.NoError(t, err)
		assert.Equal(t, 30*time.Second, duration)
	})

	t.Run("string 类型支持 time.ParseDuration 格式", func(t *testing.T) {
		duration, err := parseDuration("1h30m")
		assert.NoError(t, err)
		assert.Equal(t, 90*time.Minute, duration)

		duration, err = parseDuration("30s")
		assert.NoError(t, err)
		assert.Equal(t, 30*time.Second, duration)

		duration, err = parseDuration("500ms")
		assert.NoError(t, err)
		assert.Equal(t, 500*time.Millisecond, duration)
	})

	t.Run("string 类型支持秒数格式", func(t *testing.T) {
		duration, err := parseDuration("60")
		assert.NoError(t, err)
		assert.Equal(t, 60*time.Second, duration)

		duration, err = parseDuration("  120  ")
		assert.NoError(t, err)
		assert.Equal(t, 120*time.Second, duration)
	})

	t.Run("空字符串错误", func(t *testing.T) {
		duration, err := parseDuration("")
		assert.Error(t, err)
		assert.Equal(t, time.Duration(0), duration)
		assert.Contains(t, err.Error(), "empty duration")
	})

	t.Run("空格字符串错误", func(t *testing.T) {
		duration, err := parseDuration("   ")
		assert.Error(t, err)
		assert.Equal(t, time.Duration(0), duration)
	})

	t.Run("无效格式错误", func(t *testing.T) {
		duration, err := parseDuration("invalid")
		assert.Error(t, err)
		assert.Equal(t, time.Duration(0), duration)
		assert.Contains(t, err.Error(), "invalid duration")
	})

	t.Run("不支持的类型", func(t *testing.T) {
		duration, err := parseDuration(nil)
		assert.Error(t, err)
		assert.Equal(t, time.Duration(0), duration)
		assert.Contains(t, err.Error(), "unsupported duration type")
	})
}
