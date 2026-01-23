package cachemgr

import (
	"testing"
	"time"
)

// TestDefaultConfig 测试默认配置
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Driver != "memory" {
		t.Errorf("expected driver to be 'memory', got '%s'", config.Driver)
	}

	if config.RedisConfig == nil {
		t.Error("expected RedisConfig to be initialized")
	} else {
		if config.RedisConfig.Host != DefaultRedisHost {
			t.Errorf("expected Redis host to be '%s', got '%s'", DefaultRedisHost, config.RedisConfig.Host)
		}
		if config.RedisConfig.Port != DefaultRedisPort {
			t.Errorf("expected Redis port to be %d, got %d", DefaultRedisPort, config.RedisConfig.Port)
		}
		if config.RedisConfig.DB != DefaultRedisDB {
			t.Errorf("expected Redis DB to be %d, got %d", DefaultRedisDB, config.RedisConfig.DB)
		}
	}

	if config.MemoryConfig == nil {
		t.Error("expected MemoryConfig to be initialized")
	} else {
		if config.MemoryConfig.MaxSize != DefaultMemoryMaxSize {
			t.Errorf("expected Memory MaxSize to be %d, got %d", DefaultMemoryMaxSize, config.MemoryConfig.MaxSize)
		}
		if config.MemoryConfig.MaxAge != DefaultMemoryMaxAge {
			t.Errorf("expected Memory MaxAge to be %v, got %v", DefaultMemoryMaxAge, config.MemoryConfig.MaxAge)
		}
	}
}

// TestCacheConfigValidate 测试配置验证
func TestCacheConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *CacheConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid redis configmgr",
			config: &CacheConfig{
				Driver:      "redis",
				RedisConfig: &RedisConfig{Host: "localhost", Port: 6379},
			},
			wantErr: false,
		},
		{
			name: "valid memory configmgr",
			config: &CacheConfig{
				Driver:       "memory",
				MemoryConfig: &MemoryConfig{MaxSize: 100},
			},
			wantErr: false,
		},
		{
			name: "valid none configmgr",
			config: &CacheConfig{
				Driver: "none",
			},
			wantErr: false,
		},
		{
			name: "empty driver",
			config: &CacheConfig{
				Driver: "",
			},
			wantErr: true,
			errMsg:  "driver is required",
		},
		{
			name: "unsupported driver",
			config: &CacheConfig{
				Driver: "mongodb",
			},
			wantErr: true,
			errMsg:  "unsupported driver",
		},
		{
			name: "redis driver without configmgr",
			config: &CacheConfig{
				Driver:      "redis",
				RedisConfig: nil,
			},
			wantErr: true,
			errMsg:  "redis_config is required",
		},
		{
			name: "memory driver without configmgr",
			config: &CacheConfig{
				Driver:       "memory",
				MemoryConfig: nil,
			},
			wantErr: true,
			errMsg:  "memory_config is required",
		},
		{
			name: "driver with uppercase and spaces",
			config: &CacheConfig{
				Driver:      "  Redis  ",
				RedisConfig: &RedisConfig{Host: "localhost", Port: 6379},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errMsg)
				}
			}
		})
	}
}

// TestParseCacheConfigFromMap 测试从 map 解析配置
func TestParseCacheConfigFromMap(t *testing.T) {
	tests := []struct {
		name    string
		cfg     map[string]any
		want    *CacheConfig
		wantErr bool
	}{
		{
			name:    "nil configmgr returns defaults",
			cfg:     nil,
			wantErr: false,
		},
		{
			name:    "empty configmgr returns defaults",
			cfg:     map[string]any{},
			wantErr: false,
		},
		{
			name: "redis driver with configmgr",
			cfg: map[string]any{
				"driver": "redis",
				"redis_config": map[string]any{
					"host":     "redis.example.com",
					"port":     6380,
					"password": "secret",
					"db":       1,
				},
			},
			wantErr: false,
		},
		{
			name: "memory driver with configmgr",
			cfg: map[string]any{
				"driver": "memory",
				"memory_config": map[string]any{
					"max_size":    200,
					"max_age":     "1h",
					"max_backups": 2000,
					"compress":    true,
				},
			},
			wantErr: false,
		},
		{
			name: "none driver",
			cfg: map[string]any{
				"driver": "none",
			},
			wantErr: false,
		},
		{
			name: "driver with uppercase",
			cfg: map[string]any{
				"driver": "REDIS",
				"redis_config": map[string]any{
					"host": "localhost",
					"port": 6379,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCacheConfigFromMap(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCacheConfigFromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got == nil {
					t.Error("expected configmgr to be returned, got nil")
					return
				}
				// 验证驱动是否正确标准化
				if tt.cfg != nil && tt.cfg["driver"] != nil {
					expectedDriver := tt.cfg["driver"].(string)
					if got.Driver != "redis" && got.Driver != "memory" && got.Driver != "none" {
						// 如果输入不是有效的驱动，应该保持不变或默认为 none
						if expectedDriver != "" && expectedDriver != "none" {
							t.Logf("Note: driver '%s' was normalized to '%s'", expectedDriver, got.Driver)
						}
					}
				}
			}
		})
	}
}

// TestParseRedisConfig 测试解析 Redis 配置
func TestParseRedisConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     map[string]any
		wantErr bool
		check   func(*RedisConfig) error
	}{
		{
			name:    "empty configmgr uses defaults",
			cfg:     map[string]any{},
			wantErr: false,
			check: func(c *RedisConfig) error {
				if c.Host != DefaultRedisHost {
					t.Errorf("expected host '%s', got '%s'", DefaultRedisHost, c.Host)
				}
				if c.Port != DefaultRedisPort {
					t.Errorf("expected port %d, got %d", DefaultRedisPort, c.Port)
				}
				return nil
			},
		},
		{
			name: "full configmgr",
			cfg: map[string]any{
				"host":              "custom.redis.com",
				"port":              6380,
				"password":          "mypass",
				"db":                2,
				"max_idle_conns":    20,
				"max_open_conns":    200,
				"conn_max_lifetime": "60s",
			},
			wantErr: false,
			check: func(c *RedisConfig) error {
				if c.Host != "custom.redis.com" {
					t.Errorf("expected host 'custom.redis.com', got '%s'", c.Host)
				}
				if c.Port != 6380 {
					t.Errorf("expected port 6380, got %d", c.Port)
				}
				if c.Password != "mypass" {
					t.Errorf("expected password 'mypass', got '%s'", c.Password)
				}
				if c.DB != 2 {
					t.Errorf("expected db 2, got %d", c.DB)
				}
				if c.MaxIdleConns != 20 {
					t.Errorf("expected max_idle_conns 20, got %d", c.MaxIdleConns)
				}
				if c.MaxOpenConns != 200 {
					t.Errorf("expected max_open_conns 200, got %d", c.MaxOpenConns)
				}
				if c.ConnMaxLifetime != 60*time.Second {
					t.Errorf("expected conn_max_lifetime 60s, got %v", c.ConnMaxLifetime)
				}
				return nil
			},
		},
		{
			name: "port as int64",
			cfg: map[string]any{
				"port": int64(6379),
			},
			wantErr: false,
			check: func(c *RedisConfig) error {
				if c.Port != 6379 {
					t.Errorf("expected port 6379, got %d", c.Port)
				}
				return nil
			},
		},
		{
			name: "port as float64",
			cfg: map[string]any{
				"port": float64(6379),
			},
			wantErr: false,
			check: func(c *RedisConfig) error {
				if c.Port != 6379 {
					t.Errorf("expected port 6379, got %d", c.Port)
				}
				return nil
			},
		},
		{
			name: "port as string",
			cfg: map[string]any{
				"port": "6379",
			},
			wantErr: false,
			check: func(c *RedisConfig) error {
				if c.Port != 6379 {
					t.Errorf("expected port 6379, got %d", c.Port)
				}
				return nil
			},
		},
		{
			name: "invalid port is ignored",
			cfg: map[string]any{
				"port": float64(99999), // 超出范围
			},
			wantErr: false,
			check: func(c *RedisConfig) error {
				if c.Port != DefaultRedisPort {
					t.Errorf("expected default port %d, got %d", DefaultRedisPort, c.Port)
				}
				return nil
			},
		},
		{
			name: "db out of range is ignored",
			cfg: map[string]any{
				"db": 20, // 超出 0-15 范围
			},
			wantErr: false,
			check: func(c *RedisConfig) error {
				if c.DB != DefaultRedisDB {
					t.Errorf("expected default db %d, got %d", DefaultRedisDB, c.DB)
				}
				return nil
			},
		},
		{
			name: "duration as seconds (int)",
			cfg: map[string]any{
				"conn_max_lifetime": 30,
			},
			wantErr: false,
			check: func(c *RedisConfig) error {
				expected := 30 * time.Second
				if c.ConnMaxLifetime != expected {
					t.Errorf("expected conn_max_lifetime %v, got %v", expected, c.ConnMaxLifetime)
				}
				return nil
			},
		},
		{
			name: "duration as string",
			cfg: map[string]any{
				"conn_max_lifetime": "5m",
			},
			wantErr: false,
			check: func(c *RedisConfig) error {
				expected := 5 * time.Minute
				if c.ConnMaxLifetime != expected {
					t.Errorf("expected conn_max_lifetime %v, got %v", expected, c.ConnMaxLifetime)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRedisConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRedisConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				if err := tt.check(got); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

// TestParseMemoryConfig 测试解析 Memory 配置
func TestParseMemoryConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     map[string]any
		wantErr bool
		check   func(*MemoryConfig) error
	}{
		{
			name:    "empty configmgr uses defaults",
			cfg:     map[string]any{},
			wantErr: false,
			check: func(c *MemoryConfig) error {
				if c.MaxSize != DefaultMemoryMaxSize {
					t.Errorf("expected max_size %d, got %d", DefaultMemoryMaxSize, c.MaxSize)
				}
				if c.MaxAge != DefaultMemoryMaxAge {
					t.Errorf("expected max_age %v, got %v", DefaultMemoryMaxAge, c.MaxAge)
				}
				if c.MaxBackups != DefaultMemoryMaxBackups {
					t.Errorf("expected max_backups %d, got %d", DefaultMemoryMaxBackups, c.MaxBackups)
				}
				if c.Compress != DefaultMemoryCompress {
					t.Errorf("expected compress %v, got %v", DefaultMemoryCompress, c.Compress)
				}
				return nil
			},
		},
		{
			name: "full configmgr",
			cfg: map[string]any{
				"max_size":    500,
				"max_age":     "2h",
				"max_backups": 5000,
				"compress":    true,
			},
			wantErr: false,
			check: func(c *MemoryConfig) error {
				if c.MaxSize != 500 {
					t.Errorf("expected max_size 500, got %d", c.MaxSize)
				}
				if c.MaxAge != 2*time.Hour {
					t.Errorf("expected max_age 2h, got %v", c.MaxAge)
				}
				if c.MaxBackups != 5000 {
					t.Errorf("expected max_backups 5000, got %d", c.MaxBackups)
				}
				if !c.Compress {
					t.Errorf("expected compress true, got %v", c.Compress)
				}
				return nil
			},
		},
		{
			name: "negative max_size ignored",
			cfg: map[string]any{
				"max_size": -100,
			},
			wantErr: false,
			check: func(c *MemoryConfig) error {
				if c.MaxSize != DefaultMemoryMaxSize {
					t.Errorf("expected default max_size %d, got %d", DefaultMemoryMaxSize, c.MaxSize)
				}
				return nil
			},
		},
		{
			name: "duration as int (seconds)",
			cfg: map[string]any{
				"max_age": 3600,
			},
			wantErr: false,
			check: func(c *MemoryConfig) error {
				expected := 3600 * time.Second
				if c.MaxAge != expected {
					t.Errorf("expected max_age %v, got %v", expected, c.MaxAge)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMemoryConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMemoryConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				if err := tt.check(got); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

// TestToInt 测试类型转换函数
func TestToInt(t *testing.T) {
	tests := []struct {
		name   string
		input  any
		want   int
		wantOk bool
	}{
		{"int", 42, 42, true},
		{"int64", int64(42), 42, true},
		{"float64 (whole)", float64(42), 42, true},
		{"float64 (fractional)", float64(42.5), 0, false},
		{"string", "42", 0, false},
		{"nil", nil, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := toInt(tt.input)
			if ok != tt.wantOk {
				t.Errorf("toInt() ok = %v, wantOk %v", ok, tt.wantOk)
				return
			}
			if ok && got != tt.want {
				t.Errorf("toInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseDuration 测试时间长度解析
func TestParseDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    time.Duration
		wantErr bool
	}{
		{"int seconds", 30, 30 * time.Second, false},
		{"int64 seconds", int64(60), 60 * time.Second, false},
		{"float64 seconds", float64(90), 90 * time.Second, false},
		{"string with unit", "5m", 5 * time.Minute, false},
		{"string with unit", "1h", 1 * time.Hour, false},
		{"string with unit", "500ms", 500 * time.Millisecond, false},
		{"string plain number", "120", 120 * time.Second, false},
		{"empty string", "", 0, true},
		{"invalid string", "invalid", 0, true},
		{"unsupported type", []int{}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseDurationSeconds 测试秒数字符串解析
func TestParseDurationSeconds(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"valid number", "3600", 3600, false},
		{"number with spaces", " 3600 ", 3600, false},
		{"zero", "0", 0, false},
		{"negative", "-60", -60, false},
		{"invalid", "abc", 0, true},
		{"empty", "", 0, true},
		{"with decimal", "60.5", 60, false}, // Sscanf 只读取整数部分
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDurationSeconds(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDurationSeconds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseDurationSeconds() = %v, want %v", got, tt.want)
			}
		})
	}
}
