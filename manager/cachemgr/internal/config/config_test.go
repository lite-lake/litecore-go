package config

import (
	"testing"
	"time"
)

func TestParseCacheConfigFromMap(t *testing.T) {
	tests := []struct {
		name    string
		cfg     map[string]any
		want    *CacheConfig
		wantErr bool
	}{
		{
			name: "nil config",
			cfg:  nil,
			want: &CacheConfig{
				Driver: "none",
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
			},
			wantErr: false,
		},
		{
			name: "redis driver",
			cfg: map[string]any{
				"driver": "redis",
				"redis_config": map[string]any{
					"host":     "localhost",
					"port":     6379,
					"password": "secret",
					"db":       0,
				},
			},
			want: &CacheConfig{
				Driver: "redis",
				RedisConfig: &RedisConfig{
					Host:            "localhost",
					Port:            6379,
					Password:        "secret",
					DB:              0,
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
			},
			wantErr: false,
		},
		{
			name: "memory driver",
			cfg: map[string]any{
				"driver": "memory",
				"memory_config": map[string]any{
					"max_size":    200,
					"max_age":     "24h",
					"max_backups": 500,
					"compress":    true,
				},
			},
			want: &CacheConfig{
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
					MaxSize:    200,
					MaxAge:     24 * time.Hour,
					MaxBackups: 500,
					Compress:   true,
				},
			},
			wantErr: false,
		},
		{
			name: "none driver",
			cfg: map[string]any{
				"driver": "none",
			},
			want: &CacheConfig{
				Driver: "none",
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
			if !tt.wantErr && got != nil {
				if got.Driver != tt.want.Driver {
					t.Errorf("Driver = %v, want %v", got.Driver, tt.want.Driver)
				}
				if got.RedisConfig.Host != tt.want.RedisConfig.Host {
					t.Errorf("RedisConfig.Host = %v, want %v", got.RedisConfig.Host, tt.want.RedisConfig.Host)
				}
				if got.RedisConfig.Port != tt.want.RedisConfig.Port {
					t.Errorf("RedisConfig.Port = %v, want %v", got.RedisConfig.Port, tt.want.RedisConfig.Port)
				}
				if got.MemoryConfig.MaxSize != tt.want.MemoryConfig.MaxSize {
					t.Errorf("MemoryConfig.MaxSize = %v, want %v", got.MemoryConfig.MaxSize, tt.want.MemoryConfig.MaxSize)
				}
			}
		})
	}
}

func TestCacheConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *CacheConfig
		wantErr bool
	}{
		{
			name: "valid redis config",
			cfg: &CacheConfig{
				Driver: "redis",
				RedisConfig: &RedisConfig{
					Host: "localhost",
					Port: 6379,
					DB:   0,
				},
			},
			wantErr: false,
		},
		{
			name: "valid memory config",
			cfg: &CacheConfig{
				Driver: "memory",
				MemoryConfig: &MemoryConfig{
					MaxSize: 100,
				},
			},
			wantErr: false,
		},
		{
			name: "valid none config",
			cfg: &CacheConfig{
				Driver: "none",
			},
			wantErr: false,
		},
		{
			name: "empty driver",
			cfg: &CacheConfig{
				Driver: "",
			},
			wantErr: true,
		},
		{
			name: "redis without config",
			cfg: &CacheConfig{
				Driver:      "redis",
				RedisConfig: nil,
			},
			wantErr: true,
		},
		{
			name: "memory without config",
			cfg: &CacheConfig{
				Driver:       "memory",
				MemoryConfig: nil,
			},
			wantErr: true,
		},
		{
			name: "unsupported driver",
			cfg: &CacheConfig{
				Driver: "unsupported",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cfg.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    time.Duration
		wantErr bool
	}{
		{"int seconds", 30, 30 * time.Second, false},
		{"int64 seconds", int64(60), 60 * time.Second, false},
		{"float64 seconds", 90.0, 90 * time.Second, false},
		{"string seconds", "120", 120 * time.Second, false},
		{"string duration", "5m", 5 * time.Minute, false},
		{"string duration", "1h", 1 * time.Hour, false},
		{"empty string", "", 0, true},
		{"invalid type", []int{}, 0, true},
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

func TestParseRedisConfig_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		cfg     map[string]any
		want    *RedisConfig
		wantErr bool
	}{
		{
			name: "port as int64",
			cfg: map[string]any{
				"port": int64(6379),
			},
			want: &RedisConfig{
				Host:            DefaultRedisHost,
				Port:            6379,
				DB:              DefaultRedisDB,
				MaxIdleConns:    DefaultRedisMaxIdleConns,
				MaxOpenConns:    DefaultRedisMaxOpenConns,
				ConnMaxLifetime: DefaultRedisConnMaxLifetime,
			},
			wantErr: false,
		},
		{
			name: "port as float64",
			cfg: map[string]any{
				"port": float64(6379),
			},
			want: &RedisConfig{
				Host:            DefaultRedisHost,
				Port:            6379,
				DB:              DefaultRedisDB,
				MaxIdleConns:    DefaultRedisMaxIdleConns,
				MaxOpenConns:    DefaultRedisMaxOpenConns,
				ConnMaxLifetime: DefaultRedisConnMaxLifetime,
			},
			wantErr: false,
		},
		{
			name: "port as string",
			cfg: map[string]any{
				"port": "6379",
			},
			want: &RedisConfig{
				Host:            DefaultRedisHost,
				Port:            6379,
				DB:              DefaultRedisDB,
				MaxIdleConns:    DefaultRedisMaxIdleConns,
				MaxOpenConns:    DefaultRedisMaxOpenConns,
				ConnMaxLifetime: DefaultRedisConnMaxLifetime,
			},
			wantErr: false,
		},
		{
			name: "invalid port float64",
			cfg: map[string]any{
				"port": 6379.5,
			},
			want: &RedisConfig{
				Host:            DefaultRedisHost,
				Port:            DefaultRedisPort,
				DB:              DefaultRedisDB,
				MaxIdleConns:    DefaultRedisMaxIdleConns,
				MaxOpenConns:    DefaultRedisMaxOpenConns,
				ConnMaxLifetime: DefaultRedisConnMaxLifetime,
			},
			wantErr: false,
		},
		{
			name: "db out of range",
			cfg: map[string]any{
				"db": 20,
			},
			want: &RedisConfig{
				Host:            DefaultRedisHost,
				Port:            DefaultRedisPort,
				DB:              DefaultRedisDB,
				MaxIdleConns:    DefaultRedisMaxIdleConns,
				MaxOpenConns:    DefaultRedisMaxOpenConns,
				ConnMaxLifetime: DefaultRedisConnMaxLifetime,
			},
			wantErr: false,
		},
		{
			name: "negative max_idle_conns",
			cfg: map[string]any{
				"max_idle_conns": -1,
			},
			want: &RedisConfig{
				Host:            DefaultRedisHost,
				Port:            DefaultRedisPort,
				DB:              DefaultRedisDB,
				MaxIdleConns:    DefaultRedisMaxIdleConns,
				MaxOpenConns:    DefaultRedisMaxOpenConns,
				ConnMaxLifetime: DefaultRedisConnMaxLifetime,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRedisConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRedisConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Host != tt.want.Host {
					t.Errorf("Host = %v, want %v", got.Host, tt.want.Host)
				}
				if got.Port != tt.want.Port {
					t.Errorf("Port = %v, want %v", got.Port, tt.want.Port)
				}
				if got.DB != tt.want.DB {
					t.Errorf("DB = %v, want %v", got.DB, tt.want.DB)
				}
			}
		})
	}
}

func TestParseMemoryConfig_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		cfg     map[string]any
		want    *MemoryConfig
		wantErr bool
	}{
		{
			name: "zero max_size",
			cfg: map[string]any{
				"max_size": 0,
			},
			want: &MemoryConfig{
				MaxSize:    DefaultMemoryMaxSize,
				MaxAge:     DefaultMemoryMaxAge,
				MaxBackups: DefaultMemoryMaxBackups,
				Compress:   DefaultMemoryCompress,
			},
			wantErr: false,
		},
		{
			name: "negative max_backups",
			cfg: map[string]any{
				"max_backups": -1,
			},
			want: &MemoryConfig{
				MaxSize:    DefaultMemoryMaxSize,
				MaxAge:     DefaultMemoryMaxAge,
				MaxBackups: DefaultMemoryMaxBackups,
				Compress:   DefaultMemoryCompress,
			},
			wantErr: false,
		},
		{
			name: "invalid max_age",
			cfg: map[string]any{
				"max_age": "invalid",
			},
			want: &MemoryConfig{
				MaxSize:    DefaultMemoryMaxSize,
				MaxAge:     DefaultMemoryMaxAge,
				MaxBackups: DefaultMemoryMaxBackups,
				Compress:   DefaultMemoryCompress,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMemoryConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMemoryConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != nil {
				if got.MaxSize != tt.want.MaxSize {
					t.Errorf("MaxSize = %v, want %v", got.MaxSize, tt.want.MaxSize)
				}
				if got.MaxBackups != tt.want.MaxBackups {
					t.Errorf("MaxBackups = %v, want %v", got.MaxBackups, tt.want.MaxBackups)
				}
			}
		})
	}
}

func TestCacheConfig_DriverNormalization(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      string
		wantValid bool
		hasConfig bool
	}{
		{"lowercase", "redis", "redis", true, true},
		{"uppercase", "REDIS", "redis", true, true},
		{"mixed case", "Redis", "redis", true, true},
		{"with spaces", "  redis  ", "redis", true, true},
		{"memory", "MEMORY", "memory", true, true},
		{"none", "  None  ", "none", true, false},
		{"invalid", "invalid", "invalid", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &CacheConfig{Driver: tt.input}
			// 为需要配置的驱动添加配置
			if tt.hasConfig {
				if tt.want == "redis" {
					cfg.RedisConfig = &RedisConfig{
						Host: DefaultRedisHost,
						Port: DefaultRedisPort,
						DB:   DefaultRedisDB,
					}
				} else if tt.want == "memory" {
					cfg.MemoryConfig = &MemoryConfig{
						MaxSize: DefaultMemoryMaxSize,
					}
				}
			}
			err := cfg.Validate()

			if tt.wantValid {
				if err != nil {
					t.Errorf("Validate() error = %v", err)
				}
				if cfg.Driver != tt.want {
					t.Errorf("Driver = %v, want %v", cfg.Driver, tt.want)
				}
			} else {
				if err == nil {
					t.Error("Validate() should return error for invalid driver")
				}
			}
		})
	}
}
