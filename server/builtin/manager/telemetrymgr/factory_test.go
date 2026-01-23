package telemetrymgr

import (
	"context"
	"errors"
	"github.com/lite-lake/litecore-go/server/builtin/manager/configmgr"
	"testing"
)

// mockConfigProvider 模拟配置提供者
type mockConfigProvider struct {
	data   map[string]any
	getErr map[string]error
}

func newMockConfigProvider(data map[string]any) *mockConfigProvider {
	return &mockConfigProvider{
		data:   data,
		getErr: make(map[string]error),
	}
}

func (m *mockConfigProvider) ManagerName() string {
	return "mock-configmgr-provider"
}

func (m *mockConfigProvider) Health() error {
	return nil
}

func (m *mockConfigProvider) OnStart() error {
	return nil
}

func (m *mockConfigProvider) OnStop() error {
	return nil
}

func (m *mockConfigProvider) Get(key string) (any, error) {
	if err, ok := m.getErr[key]; ok {
		return nil, err
	}
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return nil, errors.New("key not found")
}

func (m *mockConfigProvider) Has(key string) bool {
	_, ok := m.data[key]
	return ok
}

var _ configmgr.IConfigManager = (*mockConfigProvider)(nil)

// TestBuild 测试 Build 函数
func TestBuild(t *testing.T) {
	tests := []struct {
		name         string
		driverType   string
		driverConfig map[string]any
		wantErr      bool
		errMsg       string
		verify       func(*testing.T, ITelemetryManager)
	}{
		{
			name:       "none driver",
			driverType: "none",
			wantErr:    false,
			verify: func(t *testing.T, mgr ITelemetryManager) {
				if mgr.ManagerName() != "none-telemetry" {
					t.Errorf("expected manager name 'none-telemetry', got '%s'", mgr.ManagerName())
				}
			},
		},
		{
			name:       "none driver with configmgr ignored",
			driverType: "none",
			driverConfig: map[string]any{
				"endpoint": "should-be-ignored",
			},
			wantErr: false,
			verify: func(t *testing.T, mgr ITelemetryManager) {
				if mgr.ManagerName() != "none-telemetry" {
					t.Errorf("expected manager name 'none-telemetry', got '%s'", mgr.ManagerName())
				}
			},
		},
		{
			name:       "otel driver with minimal valid configmgr",
			driverType: "otel",
			driverConfig: map[string]any{
				"endpoint": "localhost:4317",
			},
			wantErr: false,
			verify: func(t *testing.T, mgr ITelemetryManager) {
				if mgr.ManagerName() != "otel-telemetry" {
					t.Errorf("expected manager name 'otel-telemetry', got '%s'", mgr.ManagerName())
				}
				// 验证健康检查
				if err := mgr.Health(); err != nil {
					t.Errorf("expected healthy manager, got error: %v", err)
				}
				// 验证 shutdown
				ctx := context.Background()
				if err := mgr.Shutdown(ctx); err != nil {
					t.Errorf("expected clean shutdown, got error: %v", err)
				}
			},
		},
		{
			name:       "otel driver with full configmgr",
			driverType: "otel",
			driverConfig: map[string]any{
				"endpoint": "otel-collector:4317",
				"insecure": true,
				"headers": map[string]any{
					"authorization": "Bearer token",
				},
				"resource_attributes": []any{
					map[string]any{
						"key":   "service.name",
						"value": "test-service",
					},
				},
				"traces": map[string]any{
					"enabled": true,
				},
			},
			wantErr: false,
			verify: func(t *testing.T, mgr ITelemetryManager) {
				if mgr.ManagerName() != "otel-telemetry" {
					t.Errorf("expected manager name 'otel-telemetry', got '%s'", mgr.ManagerName())
				}
				// 获取 tracer 验证工作正常
				tracer := mgr.Tracer("test")
				if tracer == nil {
					t.Error("expected non-nil tracer")
				}
			},
		},
		{
			name:         "otel driver without endpoint uses default",
			driverType:   "otel",
			driverConfig: map[string]any{},
			wantErr:      false,
			verify: func(t *testing.T, mgr ITelemetryManager) {
				// 默认 endpoint 应该被使用
				if mgr == nil {
					t.Error("expected non-nil manager")
				}
			},
		},
		{
			name:       "otel driver with empty endpoint",
			driverType: "otel",
			driverConfig: map[string]any{
				"endpoint": "",
			},
			wantErr: true,
			errMsg:  "invalid configmgr: otel endpoint is required",
		},
		{
			name:       "unsupported driver type",
			driverType: "invalid",
			wantErr:    true,
			errMsg:     "unsupported driver type",
		},
		{
			name:       "otel driver ignores invalid endpoint type and uses default",
			driverType: "otel",
			driverConfig: map[string]any{
				"endpoint": 123, // 无效类型，会被忽略，使用默认值
			},
			wantErr: false,
			verify: func(t *testing.T, mgr ITelemetryManager) {
				if mgr == nil {
					t.Error("expected non-nil manager")
				}
			},
		},
		{
			name:       "driver type case insensitive - NONE",
			driverType: "NONE",
			wantErr:    false,
		},
		{
			name:       "driver type case insensitive - OTEL",
			driverType: "OTEL",
			driverConfig: map[string]any{
				"endpoint": "localhost:4317",
			},
			wantErr: false,
		},
		{
			name:       "driver type with spaces trimmed",
			driverType: "  otel  ",
			driverConfig: map[string]any{
				"endpoint": "localhost:4317",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := Build(tt.driverType, tt.driverConfig)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errMsg)
				}
				if tt.errMsg != "" && err != nil {
					if len(err.Error()) < len(tt.errMsg) || err.Error()[:len(tt.errMsg)] != tt.errMsg {
						t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
					}
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error, got %v", err)
				return
			}

			if mgr == nil {
				t.Fatal("expected non-nil manager")
			}

			if tt.verify != nil {
				tt.verify(t, mgr)
			}
		})
	}
}

// TestBuildWithConfigProvider 测试 BuildWithConfigProvider 函数
func TestBuildWithConfigProvider(t *testing.T) {
	tests := []struct {
		name           string
		configProvider *mockConfigProvider
		wantErr        bool
		errMsg         string
		verify         func(*testing.T, ITelemetryManager)
	}{
		{
			name: "none driver from configmgr provider",
			configProvider: newMockConfigProvider(map[string]any{
				"telemetry.driver": "none",
			}),
			wantErr: false,
			verify: func(t *testing.T, mgr ITelemetryManager) {
				if mgr.ManagerName() != "none-telemetry" {
					t.Errorf("expected manager name 'none-telemetry', got '%s'", mgr.ManagerName())
				}
			},
		},
		{
			name: "otel driver from configmgr provider",
			configProvider: newMockConfigProvider(map[string]any{
				"telemetry.driver": "otel",
				"telemetry.otel_config": map[string]any{
					"endpoint": "otel:4317",
					"insecure": true,
				},
			}),
			wantErr: false,
			verify: func(t *testing.T, mgr ITelemetryManager) {
				if mgr.ManagerName() != "otel-telemetry" {
					t.Errorf("expected manager name 'otel-telemetry', got '%s'", mgr.ManagerName())
				}
			},
		},
		{
			name:           "nil configmgr provider",
			configProvider: nil,
			wantErr:        true,
			errMsg:         "configProvider cannot be nil",
		},
		{
			name:           "missing telemetry.driver",
			configProvider: newMockConfigProvider(map[string]any{}),
			wantErr:        true,
			errMsg:         "failed to get telemetry.driver",
		},
		{
			name: "telemetry.driver is not a string",
			configProvider: newMockConfigProvider(map[string]any{
				"telemetry.driver": 123,
			}),
			wantErr: true,
			errMsg:  "telemetry.driver: expected string",
		},
		{
			name: "otel driver without otel_config",
			configProvider: newMockConfigProvider(map[string]any{
				"telemetry.driver": "otel",
			}),
			wantErr: true,
			errMsg:  "failed to get telemetry.otel_config",
		},
		{
			name: "otel_config is not a map",
			configProvider: newMockConfigProvider(map[string]any{
				"telemetry.driver":      "otel",
				"telemetry.otel_config": "invalid",
			}),
			wantErr: true,
			errMsg:  "telemetry.otel_config: expected map[string]any",
		},
		{
			name: "unsupported driver type",
			configProvider: newMockConfigProvider(map[string]any{
				"telemetry.driver": "invalid",
			}),
			wantErr: true,
			errMsg:  "unsupported driver type",
		},
		{
			name: "driver name case insensitive",
			configProvider: newMockConfigProvider(map[string]any{
				"telemetry.driver": "OTEL",
				"telemetry.otel_config": map[string]any{
					"endpoint": "otel:4317",
				},
			}),
			wantErr: false,
			verify: func(t *testing.T, mgr ITelemetryManager) {
				if mgr.ManagerName() != "otel-telemetry" {
					t.Errorf("expected manager name 'otel-telemetry', got '%s'", mgr.ManagerName())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var provider configmgr.IConfigManager
			if tt.configProvider == nil {
				provider = nil // 确保是接口类型的 nil
			} else {
				provider = tt.configProvider
			}
			mgr, err := BuildWithConfigProvider(provider)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errMsg)
				}
				if tt.errMsg != "" && err != nil {
					if len(err.Error()) < len(tt.errMsg) || err.Error()[:len(tt.errMsg)] != tt.errMsg {
						t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
					}
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error, got %v", err)
				return
			}

			if mgr == nil {
				t.Fatal("expected non-nil manager")
			}

			if tt.verify != nil {
				tt.verify(t, mgr)
			}
		})
	}
}

// TestBuildWithConfigProvider_GetError 测试配置提供者返回错误的情况
func TestBuildWithConfigProvider_GetError(t *testing.T) {
	tests := []struct {
		name          string
		getErrorKey   string
		getError      error
		expectedError string
	}{
		{
			name:          "error getting driver",
			getErrorKey:   "telemetry.driver",
			getError:      errors.New("connection error"),
			expectedError: "failed to get telemetry.driver",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := newMockConfigProvider(map[string]any{})
			provider.getErr[tt.getErrorKey] = tt.getError

			_, err := BuildWithConfigProvider(provider)
			if err == nil {
				t.Errorf("expected error, got nil")
			}
			if len(err.Error()) < len(tt.expectedError) || err.Error()[:len(tt.expectedError)] != tt.expectedError {
				t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, err.Error())
			}
		})
	}
}

// TestBuild_Integration 测试完整集成场景
func TestBuild_Integration(t *testing.T) {
	t.Run("build and shutdown otel manager", func(t *testing.T) {
		mgr, err := Build("otel", map[string]any{
			"endpoint": "localhost:4317",
			"traces": map[string]any{
				"enabled": false,
			},
		})
		if err != nil {
			t.Fatalf("failed to build manager: %v", err)
		}

		// 测试生命周期方法
		if err := mgr.OnStart(); err != nil {
			t.Errorf("OnStart failed: %v", err)
		}

		if err := mgr.Health(); err != nil {
			t.Errorf("Health check failed: %v", err)
		}

		if err := mgr.OnStop(); err != nil {
			t.Errorf("OnStop failed: %v", err)
		}
	})

	t.Run("multiple shutdown calls should be safe", func(t *testing.T) {
		mgr, err := Build("otel", map[string]any{
			"endpoint": "localhost:4317",
		})
		if err != nil {
			t.Fatalf("failed to build manager: %v", err)
		}

		ctx := context.Background()

		// 第一次 shutdown
		if err := mgr.Shutdown(ctx); err != nil {
			t.Errorf("first shutdown failed: %v", err)
		}

		// 第二次 shutdown 应该也是安全的（使用 sync.Once）
		if err := mgr.Shutdown(ctx); err != nil {
			t.Errorf("second shutdown failed: %v", err)
		}
	})
}

// TestBuild_OtelFeatures 测试 OTel 特性配置
func TestBuild_OtelFeatures(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]any
		enabled map[string]bool // traces, metrics, logs
	}{
		{
			name: "all features enabled",
			config: map[string]any{
				"endpoint": "localhost:4317",
				"traces":   map[string]any{"enabled": true},
				"metrics":  map[string]any{"enabled": true},
				"logs":     map[string]any{"enabled": true},
			},
			enabled: map[string]bool{"traces": true, "metrics": true, "logs": true},
		},
		{
			name: "all features disabled",
			config: map[string]any{
				"endpoint": "localhost:4317",
				"traces":   map[string]any{"enabled": false},
				"metrics":  map[string]any{"enabled": false},
				"logs":     map[string]any{"enabled": false},
			},
			enabled: map[string]bool{"traces": false, "metrics": false, "logs": false},
		},
		{
			name: "only traces enabled",
			config: map[string]any{
				"endpoint": "localhost:4317",
				"traces":   map[string]any{"enabled": true},
			},
			enabled: map[string]bool{"traces": true, "metrics": false, "logs": false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := Build("otel", tt.config)
			if err != nil {
				t.Fatalf("failed to build manager: %v", err)
			}
			defer mgr.Shutdown(context.Background())

			// 验证 provider 可以获取（即使特性未启用）
			if mgr.TracerProvider() == nil {
				t.Error("expected non-nil TracerProvider")
			}
			if mgr.MeterProvider() == nil {
				t.Error("expected non-nil MeterProvider")
			}
			if mgr.LoggerProvider() == nil {
				t.Error("expected non-nil LoggerProvider")
			}
		})
	}
}
