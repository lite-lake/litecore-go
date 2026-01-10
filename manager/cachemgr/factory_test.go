package cachemgr

import (
	"testing"
	"time"

	"com.litelake.litecore/manager/cachemgr/internal/config"
)

func TestBuild(t *testing.T) {
	tests := []struct {
		name    string
		cfg     map[string]any
		wantMgr string
	}{
		{
			name: "none driver with nil config",
			cfg:  nil,
			wantMgr: "none-cache",
		},
		{
			name: "none driver",
			cfg: map[string]any{
				"driver": "none",
			},
			wantMgr: "none-cache",
		},
		{
			name: "memory driver",
			cfg: map[string]any{
				"driver": "memory",
				"memory_config": map[string]any{
					"max_size": 100,
					"max_age":  "30d",
				},
			},
			wantMgr: "memory-cache",
		},
		{
			name: "redis driver without connection",
			cfg: map[string]any{
				"driver": "redis",
				"redis_config": map[string]any{
					"host": "localhost",
					"port": 6379,
				},
			},
			wantMgr: "none-cache", // 因为无法连接，降级到 none
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := Build(tt.cfg, nil, nil)
			if mgr.ManagerName() != tt.wantMgr {
				t.Errorf("Build() ManagerName() = %v, want %v", mgr.ManagerName(), tt.wantMgr)
			}
		})
	}
}

func TestBuildWithConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.CacheConfig
		wantErr bool
		wantMgr string
	}{
		{
			name: "valid none config",
			cfg: &config.CacheConfig{
				Driver: "none",
			},
			wantErr: false,
			wantMgr: "none-cache",
		},
		{
			name: "valid memory config",
			cfg: &config.CacheConfig{
				Driver: "memory",
				MemoryConfig: &config.MemoryConfig{
					MaxSize: 100,
					MaxAge:  30 * 24 * time.Hour,
				},
			},
			wantErr: false,
			wantMgr: "memory-cache",
		},
		{
			name: "invalid config - empty driver",
			cfg: &config.CacheConfig{
				Driver: "",
			},
			wantErr: true,
		},
		{
			name: "invalid config - redis without config",
			cfg: &config.CacheConfig{
				Driver:      "redis",
				RedisConfig: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := BuildWithConfig(tt.cfg, nil, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildWithConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && mgr.ManagerName() != tt.wantMgr {
				t.Errorf("BuildWithConfig() ManagerName() = %v, want %v", mgr.ManagerName(), tt.wantMgr)
			}
		})
	}
}

func TestBuildMemory(t *testing.T) {
	mgr := BuildMemory(5*time.Minute, 10*time.Minute, nil, nil)
	if mgr.ManagerName() != "memory-cache" {
		t.Errorf("BuildMemory() ManagerName() = %v, want %v", mgr.ManagerName(), "memory-cache")
	}
}

func TestBuildNone(t *testing.T) {
	mgr := BuildNone(nil, nil)
	if mgr.ManagerName() != "none-cache" {
		t.Errorf("BuildNone(nil, nil) ManagerName() = %v, want %v", mgr.ManagerName(), "none-cache")
	}
}

func TestBuild_DegradationScenarios(t *testing.T) {
	tests := []struct {
		name        string
		cfg         map[string]any
		expectMgr   string
		description string
	}{
		{
			name:        "invalid driver type",
			cfg: map[string]any{
				"driver": "invalid_driver",
			},
			expectMgr:   "none-cache",
			description: "should degrade to none when driver is invalid",
		},
		{
			name:        "missing redis config",
			cfg: map[string]any{
				"driver": "redis",
			},
			expectMgr:   "none-cache",
			description: "should degrade to none when redis config is missing",
		},
		{
			name:        "missing memory config",
			cfg: map[string]any{
				"driver": "memory",
			},
			expectMgr:   "memory-cache",
			description: "memory driver uses default config when config is missing",
		},
		{
			name: "malformed redis config",
			cfg: map[string]any{
				"driver": "redis",
				"redis_config": map[string]any{
					"host": "invalid-host-that-does-not-exist",
					"port": 9999,
				},
			},
			expectMgr:   "none-cache",
			description: "should degrade to none when redis connection fails",
		},
		{
			name:        "empty driver",
			cfg: map[string]any{
				"driver": "",
			},
			expectMgr:   "none-cache",
			description: "should degrade to none when driver is empty",
		},
		{
			name:        "nil config",
			cfg:         nil,
			expectMgr:   "none-cache",
			description: "should degrade to none when config is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := Build(tt.cfg, nil, nil)
			if mgr.ManagerName() != tt.expectMgr {
				t.Errorf("%s: ManagerName() = %v, want %v", tt.description, mgr.ManagerName(), tt.expectMgr)
			}
		})
	}
}

func TestBuildWithConfig_Errors(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.CacheConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "unsupported driver",
			cfg: &config.CacheConfig{
				Driver: "unsupported",
			},
			wantErr: true,
			errMsg:  "unsupported driver",
		},
		{
			name: "redis without config",
			cfg: &config.CacheConfig{
				Driver:      "redis",
				RedisConfig: nil,
			},
			wantErr: true,
		},
		{
			name: "memory without config",
			cfg: &config.CacheConfig{
				Driver:       "memory",
				MemoryConfig: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildWithConfig(tt.cfg, nil, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildWithConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message should contain %q, got %q", tt.errMsg, err.Error())
				}
			}
		})
	}
}

func TestBuildRedis_ConvenienceMethod(t *testing.T) {
	// 使用无效的主机名测试降级
	mgr := BuildRedis("invalid-host-that-does-not-exist", 9999, "", 0, nil, nil)
	if mgr.ManagerName() != "none-cache" {
		t.Errorf("BuildRedis() with invalid host should return none-cache, got %v", mgr.ManagerName())
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInMiddle(s, substr)))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
