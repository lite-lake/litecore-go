package limitermgr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	t.Run("返回默认配置", func(t *testing.T) {
		config := DefaultConfig()

		assert.NotNil(t, config)
		assert.Equal(t, "memory", config.Driver)
		assert.NotNil(t, config.RedisConfig)
		assert.NotNil(t, config.MemoryConfig)
		assert.Equal(t, DefaultRedisHost, config.RedisConfig.Host)
		assert.Equal(t, DefaultRedisPort, config.RedisConfig.Port)
		assert.Equal(t, DefaultRedisDB, config.RedisConfig.DB)
		assert.Equal(t, DefaultRedisMaxIdleConns, config.RedisConfig.MaxIdleConns)
		assert.Equal(t, DefaultRedisMaxOpenConns, config.RedisConfig.MaxOpenConns)
		assert.Equal(t, DefaultRedisConnMaxLifetime, config.RedisConfig.ConnMaxLifetime)
		assert.Equal(t, DefaultMemoryMaxBackups, config.MemoryConfig.MaxBackups)
	})
}

func TestLimiterConfig_Validate(t *testing.T) {
	t.Run("driver为空", func(t *testing.T) {
		config := &LimiterConfig{
			Driver: "",
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "driver is required")
	})

	t.Run("driver不支持", func(t *testing.T) {
		config := &LimiterConfig{
			Driver: "mysql",
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported driver")
	})

	t.Run("driver为redis但redis_config为nil", func(t *testing.T) {
		config := &LimiterConfig{
			Driver:      "redis",
			RedisConfig: nil,
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "redis_config is required")
	})

	t.Run("driver为memory但memory_config为nil", func(t *testing.T) {
		config := &LimiterConfig{
			Driver:       "memory",
			MemoryConfig: nil,
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "memory_config is required")
	})

	t.Run("driver为redis配置正确", func(t *testing.T) {
		config := &LimiterConfig{
			Driver: "redis",
			RedisConfig: &RedisLimiterConfig{
				Host:            "localhost",
				Port:            6379,
				Password:        "",
				DB:              0,
				MaxIdleConns:    10,
				MaxOpenConns:    100,
				ConnMaxLifetime: 30 * time.Second,
			},
		}
		err := config.Validate()
		assert.NoError(t, err)
		assert.Equal(t, "redis", config.Driver)
	})

	t.Run("driver为memory配置正确", func(t *testing.T) {
		config := &LimiterConfig{
			Driver: "memory",
			MemoryConfig: &MemoryLimiterConfig{
				MaxBackups: 1000,
			},
		}
		err := config.Validate()
		assert.NoError(t, err)
		assert.Equal(t, "memory", config.Driver)
	})

	t.Run("driver名称标准化", func(t *testing.T) {
		config := &LimiterConfig{
			Driver: "  REDIS  ",
			RedisConfig: &RedisLimiterConfig{
				Host: "localhost",
			},
		}
		err := config.Validate()
		assert.NoError(t, err)
		assert.Equal(t, "redis", config.Driver)
	})
}

func TestParseLimiterConfigFromMap(t *testing.T) {
	t.Run("nil输入", func(t *testing.T) {
		config, err := ParseLimiterConfigFromMap(nil)
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "memory", config.Driver)
		assert.NotNil(t, config.RedisConfig)
		assert.NotNil(t, config.MemoryConfig)
	})

	t.Run("空map", func(t *testing.T) {
		config, err := ParseLimiterConfigFromMap(map[string]any{})
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "memory", config.Driver)
	})

	t.Run("只有driver", func(t *testing.T) {
		config, err := ParseLimiterConfigFromMap(map[string]any{
			"driver": "redis",
		})
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "redis", config.Driver)
	})

	t.Run("driver为redis包含redis_config", func(t *testing.T) {
		config, err := ParseLimiterConfigFromMap(map[string]any{
			"driver": "redis",
			"redis_config": map[string]any{
				"host": "127.0.0.1",
				"port": 6380,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, "redis", config.Driver)
		assert.NotNil(t, config.RedisConfig)
		assert.Equal(t, "127.0.0.1", config.RedisConfig.Host)
		assert.Equal(t, 6380, config.RedisConfig.Port)
	})

	t.Run("driver为memory包含memory_config", func(t *testing.T) {
		config, err := ParseLimiterConfigFromMap(map[string]any{
			"driver": "memory",
			"memory_config": map[string]any{
				"max_backups": 500,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, "memory", config.Driver)
		assert.NotNil(t, config.MemoryConfig)
		assert.Equal(t, 500, config.MemoryConfig.MaxBackups)
	})

	t.Run("完整配置", func(t *testing.T) {
		config, err := ParseLimiterConfigFromMap(map[string]any{
			"driver": "redis",
			"redis_config": map[string]any{
				"host":              "192.168.1.100",
				"port":              6379,
				"password":          "secret",
				"db":                1,
				"max_idle_conns":    20,
				"max_open_conns":    200,
				"conn_max_lifetime": 60,
			},
			"memory_config": map[string]any{
				"max_backups": 2000,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, "redis", config.Driver)
		assert.Equal(t, "192.168.1.100", config.RedisConfig.Host)
		assert.Equal(t, 6379, config.RedisConfig.Port)
		assert.Equal(t, "secret", config.RedisConfig.Password)
		assert.Equal(t, 1, config.RedisConfig.DB)
		assert.Equal(t, 20, config.RedisConfig.MaxIdleConns)
		assert.Equal(t, 200, config.RedisConfig.MaxOpenConns)
		assert.Equal(t, 60*time.Second, config.RedisConfig.ConnMaxLifetime)
		assert.Equal(t, 2000, config.MemoryConfig.MaxBackups)
	})

	t.Run("driver名称标准化", func(t *testing.T) {
		config, err := ParseLimiterConfigFromMap(map[string]any{
			"driver": "  MEMORY  ",
		})
		assert.NoError(t, err)
		assert.Equal(t, "memory", config.Driver)
	})
}

func TestParseRedisLimiterConfig(t *testing.T) {
	t.Run("默认值", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{})
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, DefaultRedisHost, config.Host)
		assert.Equal(t, DefaultRedisPort, config.Port)
		assert.Equal(t, DefaultRedisDB, config.DB)
		assert.Equal(t, DefaultRedisMaxIdleConns, config.MaxIdleConns)
		assert.Equal(t, DefaultRedisMaxOpenConns, config.MaxOpenConns)
		assert.Equal(t, DefaultRedisConnMaxLifetime, config.ConnMaxLifetime)
	})

	t.Run("解析host", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"host": "example.com",
		})
		assert.NoError(t, err)
		assert.Equal(t, "example.com", config.Host)
	})

	t.Run("解析port为int", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"port": 6380,
		})
		assert.NoError(t, err)
		assert.Equal(t, 6380, config.Port)
	})

	t.Run("解析port为int64", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"port": int64(6380),
		})
		assert.NoError(t, err)
		assert.Equal(t, 6380, config.Port)
	})

	t.Run("解析port为float64整数", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"port": float64(6380),
		})
		assert.NoError(t, err)
		assert.Equal(t, 6380, config.Port)
	})

	t.Run("解析port为string", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"port": "6380",
		})
		assert.NoError(t, err)
		assert.Equal(t, 6380, config.Port)
	})

	t.Run("port边界值1", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"port": 1,
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, config.Port)
	})

	t.Run("port边界值65535", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"port": 65535,
		})
		assert.NoError(t, err)
		assert.Equal(t, 65535, config.Port)
	})

	t.Run("port为int超出范围不检查", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"port": 65536,
		})
		assert.NoError(t, err)
		assert.Equal(t, 65536, config.Port)
	})

	t.Run("port为int负数不检查", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"port": -1,
		})
		assert.NoError(t, err)
		assert.Equal(t, -1, config.Port)
	})

	t.Run("port为int64超出范围使用默认值", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"port": int64(65536),
		})
		assert.NoError(t, err)
		assert.Equal(t, DefaultRedisPort, config.Port)
	})

	t.Run("port为int64负数使用默认值", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"port": int64(-1),
		})
		assert.NoError(t, err)
		assert.Equal(t, DefaultRedisPort, config.Port)
	})

	t.Run("port为float64小数使用默认值", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"port": 6380.5,
		})
		assert.NoError(t, err)
		assert.Equal(t, DefaultRedisPort, config.Port)
	})

	t.Run("解析password", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"password": "mypassword",
		})
		assert.NoError(t, err)
		assert.Equal(t, "mypassword", config.Password)
	})

	t.Run("解析db为int", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"db": 5,
		})
		assert.NoError(t, err)
		assert.Equal(t, 5, config.DB)
	})

	t.Run("解析db为int64", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"db": int64(5),
		})
		assert.NoError(t, err)
		assert.Equal(t, 5, config.DB)
	})

	t.Run("解析db为float64整数", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"db": float64(5),
		})
		assert.NoError(t, err)
		assert.Equal(t, 5, config.DB)
	})

	t.Run("db边界值0", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"db": 0,
		})
		assert.NoError(t, err)
		assert.Equal(t, 0, config.DB)
	})

	t.Run("db边界值15", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"db": 15,
		})
		assert.NoError(t, err)
		assert.Equal(t, 15, config.DB)
	})

	t.Run("db超出范围使用默认值", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"db": 16,
		})
		assert.NoError(t, err)
		assert.Equal(t, DefaultRedisDB, config.DB)
	})

	t.Run("db为负数使用默认值", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"db": -1,
		})
		assert.NoError(t, err)
		assert.Equal(t, DefaultRedisDB, config.DB)
	})

	t.Run("解析max_idle_conns", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"max_idle_conns": 20,
		})
		assert.NoError(t, err)
		assert.Equal(t, 20, config.MaxIdleConns)
	})

	t.Run("max_idle_conns为0保持默认值", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"max_idle_conns": 0,
		})
		assert.NoError(t, err)
		assert.Equal(t, DefaultRedisMaxIdleConns, config.MaxIdleConns)
	})

	t.Run("max_idle_conns为负数保持默认值", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"max_idle_conns": -1,
		})
		assert.NoError(t, err)
		assert.Equal(t, DefaultRedisMaxIdleConns, config.MaxIdleConns)
	})

	t.Run("解析max_open_conns", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"max_open_conns": 200,
		})
		assert.NoError(t, err)
		assert.Equal(t, 200, config.MaxOpenConns)
	})

	t.Run("max_open_conns为0保持默认值", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"max_open_conns": 0,
		})
		assert.NoError(t, err)
		assert.Equal(t, DefaultRedisMaxOpenConns, config.MaxOpenConns)
	})

	t.Run("解析conn_max_lifetime为int", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"conn_max_lifetime": 60,
		})
		assert.NoError(t, err)
		assert.Equal(t, 60*time.Second, config.ConnMaxLifetime)
	})

	t.Run("解析conn_max_lifetime为string", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"conn_max_lifetime": "2m",
		})
		assert.NoError(t, err)
		assert.Equal(t, 2*time.Minute, config.ConnMaxLifetime)
	})

	t.Run("完整Redis配置", func(t *testing.T) {
		config, err := parseRedisLimiterConfig(map[string]any{
			"host":              "redis.example.com",
			"port":              6379,
			"password":          "pass123",
			"db":                2,
			"max_idle_conns":    30,
			"max_open_conns":    300,
			"conn_max_lifetime": "5m",
		})
		assert.NoError(t, err)
		assert.Equal(t, "redis.example.com", config.Host)
		assert.Equal(t, 6379, config.Port)
		assert.Equal(t, "pass123", config.Password)
		assert.Equal(t, 2, config.DB)
		assert.Equal(t, 30, config.MaxIdleConns)
		assert.Equal(t, 300, config.MaxOpenConns)
		assert.Equal(t, 5*time.Minute, config.ConnMaxLifetime)
	})
}

func TestParseMemoryLimiterConfig(t *testing.T) {
	t.Run("默认值", func(t *testing.T) {
		config, err := parseMemoryLimiterConfig(map[string]any{})
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, DefaultMemoryMaxBackups, config.MaxBackups)
	})

	t.Run("解析max_backups", func(t *testing.T) {
		config, err := parseMemoryLimiterConfig(map[string]any{
			"max_backups": 500,
		})
		assert.NoError(t, err)
		assert.Equal(t, 500, config.MaxBackups)
	})

	t.Run("max_backups为0", func(t *testing.T) {
		config, err := parseMemoryLimiterConfig(map[string]any{
			"max_backups": 0,
		})
		assert.NoError(t, err)
		assert.Equal(t, 0, config.MaxBackups)
	})

	t.Run("max_backups为负数保持默认值", func(t *testing.T) {
		config, err := parseMemoryLimiterConfig(map[string]any{
			"max_backups": -1,
		})
		assert.NoError(t, err)
		assert.Equal(t, DefaultMemoryMaxBackups, config.MaxBackups)
	})
}

func TestToInt(t *testing.T) {
	t.Run("int类型", func(t *testing.T) {
		result, ok := toInt(42)
		assert.True(t, ok)
		assert.Equal(t, 42, result)
	})

	t.Run("int64类型", func(t *testing.T) {
		result, ok := toInt(int64(42))
		assert.True(t, ok)
		assert.Equal(t, 42, result)
	})

	t.Run("float64整数", func(t *testing.T) {
		result, ok := toInt(float64(42))
		assert.True(t, ok)
		assert.Equal(t, 42, result)
	})

	t.Run("float64小数失败", func(t *testing.T) {
		result, ok := toInt(float64(42.5))
		assert.False(t, ok)
		assert.Equal(t, 0, result)
	})

	t.Run("字符串类型失败", func(t *testing.T) {
		result, ok := toInt("42")
		assert.False(t, ok)
		assert.Equal(t, 0, result)
	})

	t.Run("nil类型失败", func(t *testing.T) {
		result, ok := toInt(nil)
		assert.False(t, ok)
		assert.Equal(t, 0, result)
	})

	t.Run("bool类型失败", func(t *testing.T) {
		result, ok := toInt(true)
		assert.False(t, ok)
		assert.Equal(t, 0, result)
	})

	t.Run("float64大整数", func(t *testing.T) {
		result, ok := toInt(float64(1234567890))
		assert.True(t, ok)
		assert.Equal(t, 1234567890, result)
	})

	t.Run("负数", func(t *testing.T) {
		result, ok := toInt(-42)
		assert.True(t, ok)
		assert.Equal(t, -42, result)
	})
}

func TestParseDuration(t *testing.T) {
	t.Run("int类型", func(t *testing.T) {
		result, err := parseDuration(60)
		assert.NoError(t, err)
		assert.Equal(t, 60*time.Second, result)
	})

	t.Run("int64类型", func(t *testing.T) {
		result, err := parseDuration(int64(60))
		assert.NoError(t, err)
		assert.Equal(t, 60*time.Second, result)
	})

	t.Run("float64类型", func(t *testing.T) {
		result, err := parseDuration(float64(60))
		assert.NoError(t, err)
		assert.Equal(t, 60*time.Second, result)
	})

	t.Run("float64小数", func(t *testing.T) {
		result, err := parseDuration(float64(60.5))
		assert.NoError(t, err)
		assert.Equal(t, 60*time.Second, result)
	})

	t.Run("Go duration格式字符串", func(t *testing.T) {
		result, err := parseDuration("2m30s")
		assert.NoError(t, err)
		assert.Equal(t, 2*time.Minute+30*time.Second, result)
	})

	t.Run("纯数字秒数字符串", func(t *testing.T) {
		result, err := parseDuration("120")
		assert.NoError(t, err)
		assert.Equal(t, 120*time.Second, result)
	})

	t.Run("空字符串", func(t *testing.T) {
		result, err := parseDuration("")
		assert.Error(t, err)
		assert.Equal(t, time.Duration(0), result)
	})

	t.Run("空格字符串", func(t *testing.T) {
		result, err := parseDuration("   ")
		assert.Error(t, err)
		assert.Equal(t, time.Duration(0), result)
	})

	t.Run("无效格式字符串", func(t *testing.T) {
		result, err := parseDuration("invalid")
		assert.Error(t, err)
		assert.Equal(t, time.Duration(0), result)
	})

	t.Run("无效类型", func(t *testing.T) {
		result, err := parseDuration(true)
		assert.Error(t, err)
		assert.Equal(t, time.Duration(0), result)
	})

	t.Run("零值", func(t *testing.T) {
		result, err := parseDuration(0)
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(0), result)
	})

	t.Run("负数", func(t *testing.T) {
		result, err := parseDuration(-60)
		assert.NoError(t, err)
		assert.Equal(t, -60*time.Second, result)
	})

	t.Run("duration格式ms", func(t *testing.T) {
		result, err := parseDuration("500ms")
		assert.NoError(t, err)
		assert.Equal(t, 500*time.Millisecond, result)
	})

	t.Run("duration格式h", func(t *testing.T) {
		result, err := parseDuration("2h")
		assert.NoError(t, err)
		assert.Equal(t, 2*time.Hour, result)
	})
}

func TestParseDurationSeconds(t *testing.T) {
	t.Run("有效数字", func(t *testing.T) {
		result, err := parseDurationSeconds("120")
		assert.NoError(t, err)
		assert.Equal(t, 120, result)
	})

	t.Run("零值", func(t *testing.T) {
		result, err := parseDurationSeconds("0")
		assert.NoError(t, err)
		assert.Equal(t, 0, result)
	})

	t.Run("负数", func(t *testing.T) {
		result, err := parseDurationSeconds("-60")
		assert.NoError(t, err)
		assert.Equal(t, -60, result)
	})

	t.Run("大数字", func(t *testing.T) {
		result, err := parseDurationSeconds("86400")
		assert.NoError(t, err)
		assert.Equal(t, 86400, result)
	})

	t.Run("空格包裹", func(t *testing.T) {
		result, err := parseDurationSeconds("  120  ")
		assert.NoError(t, err)
		assert.Equal(t, 120, result)
	})

	t.Run("无效格式", func(t *testing.T) {
		result, err := parseDurationSeconds("abc")
		assert.Error(t, err)
		assert.Equal(t, 0, result)
	})

	t.Run("包含小数点", func(t *testing.T) {
		result, err := parseDurationSeconds("120.5")
		assert.NoError(t, err)
		assert.Equal(t, 120, result)
	})

	t.Run("空字符串", func(t *testing.T) {
		result, err := parseDurationSeconds("")
		assert.Error(t, err)
		assert.Equal(t, 0, result)
	})

	t.Run("空格字符串", func(t *testing.T) {
		result, err := parseDurationSeconds("   ")
		assert.Error(t, err)
		assert.Equal(t, 0, result)
	})

	t.Run("包含非数字字符", func(t *testing.T) {
		result, err := parseDurationSeconds("120s")
		assert.NoError(t, err)
		assert.Equal(t, 120, result)
	})
}
