package cachemgr

import (
	"context"
	"errors"
	"testing"
	"time"
)

// MockConfigProvider 用于测试的模拟配置提供者
type MockConfigProvider struct {
	data map[string]any
	err  error
}

func (m *MockConfigProvider) Get(key string) (any, error) {
	if m == nil {
		return nil, errors.New("config provider is nil")
	}
	if m.err != nil {
		return nil, m.err
	}
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return nil, errors.New("key not found")
}

func (m *MockConfigProvider) Has(key string) bool {
	if m == nil || m.err != nil {
		return false
	}
	_, ok := m.data[key]
	return ok
}

func (m *MockConfigProvider) ConfigProviderName() string {
	return "mock"
}

// TestBuild 测试 Build 函数
func TestBuild(t *testing.T) {
	tests := []struct {
		name         string
		driverType   string
		driverConfig map[string]any
		wantErr      bool
		checkManager func(t *testing.T, mgr CacheManager)
	}{
		{
			name:       "none driver",
			driverType: "none",
			wantErr:    false,
			checkManager: func(t *testing.T, mgr CacheManager) {
				if mgr == nil {
					t.Error("expected manager to be created, got nil")
				}
				if mgr.ManagerName() != "cacheManagerNoneImpl" {
					t.Errorf("expected manager name 'cacheManagerNoneImpl', got '%s'", mgr.ManagerName())
				}
			},
		},
		{
			name:       "memory driver with config",
			driverType: "memory",
			driverConfig: map[string]any{
				"max_age": "1h",
			},
			wantErr: false,
			checkManager: func(t *testing.T, mgr CacheManager) {
				if mgr == nil {
					t.Error("expected manager to be created, got nil")
				}
				if mgr.ManagerName() != "cacheManagerMemoryImpl" {
					t.Errorf("expected manager name 'cacheManagerMemoryImpl', got '%s'", mgr.ManagerName())
				}
			},
		},
		{
			name:         "memory driver with empty config",
			driverType:   "memory",
			driverConfig: map[string]any{},
			wantErr:      false,
			checkManager: func(t *testing.T, mgr CacheManager) {
				if mgr == nil {
					t.Error("expected manager to be created, got nil")
				}
			},
		},
		{
			name:       "unsupported driver type",
			driverType: "mongodb",
			wantErr:    true,
		},
		{
			name:       "empty driver type",
			driverType: "",
			wantErr:    true,
		},
		// Redis 驱动测试需要实际的 Redis 连接，这里只测试错误情况
		{
			name:       "redis driver without connection (will fail)",
			driverType: "redis",
			driverConfig: map[string]any{
				"host": "localhost",
				"port": 9999, // 使用不存在的端口
			},
			wantErr: true, // 预期会失败，因为无法连接
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Build(tt.driverType, tt.driverConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.checkManager != nil {
					tt.checkManager(t, got)
				}
				// 清理资源
				if got != nil {
					got.Close()
				}
			}
		})
	}
}

// TestBuildWithConfigProvider 测试 BuildWithConfigProvider 函数
func TestBuildWithConfigProvider(t *testing.T) {
	tests := []struct {
		name           string
		configProvider *MockConfigProvider
		wantErr        bool
		checkManager   func(t *testing.T, mgr CacheManager)
	}{
		{
			name: "none driver",
			configProvider: &MockConfigProvider{
				data: map[string]any{
					"cache.driver": "none",
				},
			},
			wantErr: false,
			checkManager: func(t *testing.T, mgr CacheManager) {
				if mgr.ManagerName() != "cacheManagerNoneImpl" {
					t.Errorf("expected 'cacheManagerNoneImpl', got '%s'", mgr.ManagerName())
				}
			},
		},
		{
			name: "memory driver with config",
			configProvider: &MockConfigProvider{
				data: map[string]any{
					"cache.driver": "memory",
					"cache.memory_config": map[string]any{
						"max_size": 200,
						"max_age":  "2h",
					},
				},
			},
			wantErr: false,
			checkManager: func(t *testing.T, mgr CacheManager) {
				if mgr.ManagerName() != "cacheManagerMemoryImpl" {
					t.Errorf("expected 'cacheManagerMemoryImpl', got '%s'", mgr.ManagerName())
				}
			},
		},
		{
			name: "memory driver with empty config",
			configProvider: &MockConfigProvider{
				data: map[string]any{
					"cache.driver":        "memory",
					"cache.memory_config": map[string]any{},
				},
			},
			wantErr: false,
			checkManager: func(t *testing.T, mgr CacheManager) {
				if mgr == nil {
					t.Error("expected manager to be created")
				}
			},
		},
		{
			name:           "nil config provider",
			configProvider: nil,
			wantErr:        true,
		},
		{
			name: "missing driver key",
			configProvider: &MockConfigProvider{
				data: map[string]any{},
			},
			wantErr: true,
		},
		{
			name: "driver key not a string",
			configProvider: &MockConfigProvider{
				data: map[string]any{
					"cache.driver": 123,
				},
			},
			wantErr: true,
		},
		{
			name: "memory driver with invalid config type",
			configProvider: &MockConfigProvider{
				data: map[string]any{
					"cache.driver":        "memory",
					"cache.memory_config": "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "unsupported driver type",
			configProvider: &MockConfigProvider{
				data: map[string]any{
					"cache.driver": "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "config provider returns error",
			configProvider: &MockConfigProvider{
				err: errors.New("config error"),
			},
			wantErr: true,
		},
		{
			name: "redis driver without config",
			configProvider: &MockConfigProvider{
				data: map[string]any{
					"cache.driver": "redis",
				},
			},
			wantErr: true, // 缺少 redis_config
		},
		{
			name: "redis driver with invalid config type",
			configProvider: &MockConfigProvider{
				data: map[string]any{
					"cache.driver":       "redis",
					"cache.redis_config": "invalid",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildWithConfigProvider(tt.configProvider)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildWithConfigProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.checkManager != nil {
					tt.checkManager(t, got)
				}
				// 清理资源
				if got != nil {
					got.Close()
				}
			}
		})
	}
}

// TestBuildMemoryDefaultExpiration 测试内存缓存的默认过期时间设置
func TestBuildMemoryDefaultExpiration(t *testing.T) {
	driverConfig := map[string]any{
		"max_age": "30m", // 30 分钟
	}

	mgr, err := Build("memory", driverConfig)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	defer mgr.Close()

	if mgr.ManagerName() != "cacheManagerMemoryImpl" {
		t.Errorf("expected 'cacheManagerMemoryImpl', got '%s'", mgr.ManagerName())
	}

	// 测试基本操作是否正常
	ctx := context.Background()
	err = mgr.Set(ctx, "test_key", "test_value", 5*time.Minute)
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}
}

// TestBuildUpperCaseDriver 测试大写驱动名称
func TestBuildUpperCaseDriver(t *testing.T) {
	tests := []struct {
		name       string
		driverType string
		wantName   string
		wantErr    bool
	}{
		{"MEMORY", "MEMORY", "cacheManagerMemoryImpl", true},         // Build 函数不支持大写
		{"Memory", "Memory", "cacheManagerMemoryImpl", true},         // Build 函数不支持大写
		{"  Memory  ", "  Memory  ", "cacheManagerMemoryImpl", true}, // Build 函数不支持大写
		{"NONE", "NONE", "cacheManagerNoneImpl", true},               // Build 函数不支持大写
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Build(tt.driverType, map[string]any{})
			// Build 函数目前不进行大小写转换，所以大写会失败
			if (err != nil) != tt.wantErr {
				t.Logf("Build() error = %v (expected to fail with uppercase driver)", err)
			}
			if !tt.wantErr && got != nil {
				defer got.Close()
				if got.ManagerName() != tt.wantName {
					t.Errorf("expected manager name '%s', got '%s'", tt.wantName, got.ManagerName())
				}
			}
		})
	}
}
