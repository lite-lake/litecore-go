package loggermgr

import (
	"errors"
	"strings"
	"testing"

	"com.litelake.litecore/common"
)

// mockConfigProvider 模拟配置提供者
type mockConfigProvider struct {
	config map[string]any
}

func newMockConfigProvider(config map[string]any) *mockConfigProvider {
	return &mockConfigProvider{config: config}
}

func (m *mockConfigProvider) ConfigProviderName() string {
	return "mock"
}

func (m *mockConfigProvider) Get(key string) (any, error) {
	if m == nil || m.config == nil {
		return nil, errors.New("key not found")
	}
	if val, ok := m.config[key]; ok {
		return val, nil
	}
	return nil, errors.New("key not found")
}

func (m *mockConfigProvider) Has(key string) bool {
	if m == nil || m.config == nil {
		return false
	}
	_, ok := m.config[key]
	return ok
}

// TestBuild 测试 Build 函数
func TestBuild(t *testing.T) {
	tests := []struct {
		name         string
		driverType   string
		driverConfig map[string]any
		wantErr      bool
		errString    string
		validate     func(*testing.T, LoggerManager)
	}{
		{
			name:       "Valid none driver",
			driverType: "none",
			wantErr:    false,
			validate: func(t *testing.T, mgr LoggerManager) {
				if mgr.ManagerName() != "none-logger" {
					t.Errorf("Expected manager name 'none-logger', got '%s'", mgr.ManagerName())
				}
			},
		},
		{
			name:       "Valid zap driver with console",
			driverType: "zap",
			driverConfig: map[string]any{
				"zap_config": map[string]any{
					"console_enabled": true,
					"console_config": map[string]any{
						"level": "info",
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, mgr LoggerManager) {
				if mgr.ManagerName() != "zap-logger" {
					t.Errorf("Expected manager name 'zap-logger', got '%s'", mgr.ManagerName())
				}
			},
		},
		{
			name:       "Valid zap driver with file",
			driverType: "zap",
			driverConfig: map[string]any{
				"zap_config": map[string]any{
					"console_enabled": true,
					"file_enabled":    true,
					"file_config": map[string]any{
						"level": "info",
						"path":  "/tmp/test.log",
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, mgr LoggerManager) {
				if mgr.ManagerName() != "zap-logger" {
					t.Errorf("Expected manager name 'zap-logger', got '%s'", mgr.ManagerName())
				}
			},
		},
		{
			name:       "Valid zap driver with all outputs",
			driverType: "zap",
			driverConfig: map[string]any{
				"zap_config": map[string]any{
					"telemetry_enabled": true,
					"telemetry_config": map[string]any{
						"level": "debug",
					},
					"console_enabled": true,
					"console_config": map[string]any{
						"level": "info",
					},
					"file_enabled": true,
					"file_config": map[string]any{
						"level": "info",
						"path":  "/tmp/test.log",
					},
				},
			},
			wantErr: false,
		},
		{
			name:       "Invalid driver type",
			driverType: "invalid",
			wantErr:    true,
			errString:  "unsupported driver type",
		},
		{
			name:         "Zap driver with no output - has default console enabled",
			driverType:   "zap",
			driverConfig: map[string]any{},
			wantErr:      false, // DefaultZapConfig has console enabled
		},
		{
			name:       "Zap driver with file but no path",
			driverType: "zap",
			driverConfig: map[string]any{
				"zap_config": map[string]any{
					"console_enabled": false, // Disable console so we get the file error
					"file_enabled":    true,
					"file_config": map[string]any{
						"level": "info",
					},
				},
			},
			wantErr:   true,
			errString: "file log path is required",
		},
		{
			name:       "Driver with spaces",
			driverType: "  zap  ",
			driverConfig: map[string]any{
				"zap_config": map[string]any{
					"console_enabled": true,
				},
			},
			wantErr: false,
		},
		{
			name:       "Driver uppercase",
			driverType: "ZAP",
			driverConfig: map[string]any{
				"zap_config": map[string]any{
					"console_enabled": true,
				},
			},
			wantErr: false,
		},
		{
			name:         "Empty driver config",
			driverType:   "zap",
			driverConfig: nil,
			wantErr:      false, // DefaultZapConfig has console enabled
		},
		{
			name:       "Invalid telemetry level",
			driverType: "zap",
			driverConfig: map[string]any{
				"zap_config": map[string]any{
					"console_enabled":   false, // Disable console to get the telemetry error
					"telemetry_enabled": true,
					"telemetry_config": map[string]any{
						"level": "invalid",
					},
				},
			},
			wantErr:   true,
			errString: "invalid telemetry log level",
		},
		{
			name:       "Invalid console level",
			driverType: "zap",
			driverConfig: map[string]any{
				"zap_config": map[string]any{
					"console_enabled": true,
					"console_config": map[string]any{
						"level": "invalid",
					},
				},
			},
			wantErr:   true,
			errString: "invalid console log level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := Build(tt.driverType, tt.driverConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errString) {
				t.Errorf("Build() error = %v, want contain %v", err, tt.errString)
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, mgr)
			}
			if !tt.wantErr && mgr != nil {
				// 清理资源
				mgr.Shutdown(nil)
			}
		})
	}
}

// TestBuildWithConfigProvider 测试 BuildWithConfigProvider 函数
func TestBuildWithConfigProvider(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func() *mockConfigProvider
		wantErr     bool
		errString   string
		validate    func(*testing.T, LoggerManager)
	}{
		{
			name: "Valid none driver",
			setupConfig: func() *mockConfigProvider {
				return newMockConfigProvider(map[string]any{
					"logger.driver": "none",
				})
			},
			wantErr: false,
			validate: func(t *testing.T, mgr LoggerManager) {
				if mgr.ManagerName() != "none-logger" {
					t.Errorf("Expected manager name 'none-logger', got '%s'", mgr.ManagerName())
				}
			},
		},
		{
			name: "Valid zap driver with console",
			setupConfig: func() *mockConfigProvider {
				return newMockConfigProvider(map[string]any{
					"logger.driver": "zap",
					"logger.zap_config": map[string]any{
						"console_enabled": true,
						"console_config": map[string]any{
							"level": "info",
						},
					},
				})
			},
			wantErr: false,
			validate: func(t *testing.T, mgr LoggerManager) {
				if mgr.ManagerName() != "zap-logger" {
					t.Errorf("Expected manager name 'zap-logger', got '%s'", mgr.ManagerName())
				}
			},
		},
		{
			name: "Valid zap driver with file",
			setupConfig: func() *mockConfigProvider {
				return newMockConfigProvider(map[string]any{
					"logger.driver": "zap",
					"logger.zap_config": map[string]any{
						"console_enabled": true,
						"file_enabled":    true,
						"file_config": map[string]any{
							"level": "info",
							"path":  "/tmp/test.log",
							"rotation": map[string]any{
								"max_size":    100,
								"max_age":     30,
								"max_backups": 10,
								"compress":    true,
							},
						},
					},
				})
			},
			wantErr: false,
		},
		{
			name: "Nil config provider - Get returns error",
			setupConfig: func() *mockConfigProvider {
				// 返回 nil provider，实际会触发 "key not found" 错误
				// 因为 mock 的 Get 方法对 nil receiver 返回错误
				return nil
			},
			wantErr:   true,
			errString: "failed to get logger.driver",
		},
		{
			name: "Missing driver",
			setupConfig: func() *mockConfigProvider {
				return newMockConfigProvider(map[string]any{
					"logger.other": "value",
				})
			},
			wantErr:   true,
			errString: "failed to get logger.driver",
		},
		{
			name: "Invalid driver type",
			setupConfig: func() *mockConfigProvider {
				return newMockConfigProvider(map[string]any{
					"logger.driver": "invalid",
				})
			},
			wantErr:   true,
			errString: "unsupported driver type",
		},
		{
			name: "Driver not string",
			setupConfig: func() *mockConfigProvider {
				return newMockConfigProvider(map[string]any{
					"logger.driver": 123,
				})
			},
			wantErr:   true,
			errString: "must be a string",
		},
		{
			name: "Zap driver without zap_config",
			setupConfig: func() *mockConfigProvider {
				return newMockConfigProvider(map[string]any{
					"logger.driver": "zap",
				})
			},
			wantErr:   true,
			errString: "failed to get logger.zap_config",
		},
		{
			name: "Zap config not map",
			setupConfig: func() *mockConfigProvider {
				return newMockConfigProvider(map[string]any{
					"logger.driver":     "zap",
					"logger.zap_config": "invalid",
				})
			},
			wantErr:   true,
			errString: "must be a map",
		},
		{
			name: "Driver with spaces",
			setupConfig: func() *mockConfigProvider {
				return newMockConfigProvider(map[string]any{
					"logger.driver": "  ZAP  ",
					"logger.zap_config": map[string]any{
						"console_enabled": true,
					},
				})
			},
			wantErr: false,
		},
		{
			name: "None driver without zap_config",
			setupConfig: func() *mockConfigProvider {
				return newMockConfigProvider(map[string]any{
					"logger.driver": "none",
				})
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := tt.setupConfig()
			mgr, err := BuildWithConfigProvider(provider)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildWithConfigProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errString) {
				t.Errorf("BuildWithConfigProvider() error = %v, want contain %v", err, tt.errString)
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, mgr)
			}
			if !tt.wantErr && mgr != nil {
				// 清理资源
				mgr.Shutdown(nil)
			}
		})
	}
}

// TestBuildWithConfigProviderNilConfig 测试配置提供者为 nil 的情况
func TestBuildWithConfigProviderNilConfig(t *testing.T) {
	var provider common.BaseConfigProvider = nil
	_, err := BuildWithConfigProvider(provider)
	if err == nil {
		t.Error("Expected error when provider is nil")
		return
	}
	if !strings.Contains(err.Error(), "configProvider cannot be nil") {
		t.Errorf("Expected error about nil provider, got: %v", err)
	}
}

// TestBuildDriverTypeNormalization 测试驱动类型标准化
func TestBuildDriverTypeNormalization(t *testing.T) {
	driverTypes := []string{
		"zap", "ZAP", "Zap", "  zap  ", "\tzap\n",
	}

	for _, driverType := range driverTypes {
		t.Run(driverType, func(t *testing.T) {
			mgr, err := Build(driverType, map[string]any{
				"console_enabled": true,
			})
			if err != nil {
				t.Errorf("Build() with driver '%s' failed: %v", driverType, err)
				return
			}
			if mgr.ManagerName() != "zap-logger" {
				t.Errorf("Expected manager name 'zap-logger', got '%s'", mgr.ManagerName())
			}
			mgr.Shutdown(nil)
		})
	}
}

// TestBuildNoneDriverVariousConfigs 测试 none 驱动忽略配置
func TestBuildNoneDriverVariousConfigs(t *testing.T) {
	configs := []map[string]any{
		nil,
		{},
		{"some": "config"},
		{"console_enabled": true},
	}

	for i, config := range configs {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			mgr, err := Build("none", config)
			if err != nil {
				t.Errorf("Build() with none driver failed: %v", err)
				return
			}
			if mgr.ManagerName() != "none-logger" {
				t.Errorf("Expected manager name 'none-logger', got '%s'", mgr.ManagerName())
			}
			mgr.Shutdown(nil)
		})
	}
}
