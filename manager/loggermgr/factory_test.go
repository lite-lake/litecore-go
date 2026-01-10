package loggermgr

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"com.litelake.litecore/manager/loggermgr/internal/config"
)

func TestBuild(t *testing.T) {
	tests := []struct {
		name           string
		config         map[string]any
		expectedType   interface{}
	}{
		{
			name: "valid config with console",
			config: map[string]any{
				"console_enabled": true,
				"console_config": map[string]any{
					"level": "info",
				},
			},
			expectedType: &LoggerManagerAdapter{},
		},
		{
			name: "valid config with file",
			config: map[string]any{
				"file_enabled": true,
				"file_config": map[string]any{
					"level": "info",
					"path":  filepath.Join(os.TempDir(), "test.log"),
				},
			},
			expectedType: &LoggerManagerAdapter{},
		},
		{
			name: "invalid config - no outputs",
			config: map[string]any{
				"console_enabled": false,
				"file_enabled":    false,
			},
			expectedType: &NoneLoggerManagerAdapter{},
		},
		{
			name: "invalid config - invalid log level",
			config: map[string]any{
				"console_enabled": true,
				"console_config": map[string]any{
					"level": "invalid",
				},
			},
			expectedType: &NoneLoggerManagerAdapter{},
		},
		{
			name: "nil config - returns default",
			config: nil,
			expectedType: &NoneLoggerManagerAdapter{},
		},
		{
			name: "empty config - returns default",
			config: map[string]any{},
			expectedType: &NoneLoggerManagerAdapter{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := Build(tt.config, nil)

			if mgr == nil {
				t.Fatal("Build() returned nil")
			}

			// Check the type
			switch tt.expectedType.(type) {
			case *LoggerManagerAdapter:
				if _, ok := mgr.(*LoggerManagerAdapter); !ok {
					t.Errorf("Build() should return LoggerManagerAdapter, got %T", mgr)
				}
				// Cleanup
				if loggerMgr, ok := mgr.(LoggerManager); ok {
					_ = loggerMgr.Shutdown(context.Background())
				}
			case *NoneLoggerManagerAdapter:
				if _, ok := mgr.(*NoneLoggerManagerAdapter); !ok {
					t.Errorf("Build() should return NoneLoggerManagerAdapter, got %T", mgr)
				}
			}
		})
	}
}

func TestBuild_WithAllOutputs(t *testing.T) {
	cfg := map[string]any{
		"telemetry_enabled": true,
		"telemetry_config": map[string]any{
			"level": "info",
		},
		"console_enabled": true,
		"console_config": map[string]any{
			"level": "debug",
		},
		"file_enabled": true,
		"file_config": map[string]any{
			"level": "warn",
			"path":  filepath.Join(os.TempDir(), "factory-test.log"),
			"rotation": map[string]any{
				"max_size":    100,
				"max_age":     30,
				"max_backups": 10,
				"compress":    true,
			},
		},
	}

	mgr := Build(cfg, nil)

	if mgr == nil {
		t.Fatal("Build() returned nil")
	}

	// Should return LoggerManagerAdapter
	if _, ok := mgr.(*LoggerManagerAdapter); !ok {
		t.Errorf("Build() should return LoggerManagerAdapter, got %T", mgr)
	}

	// Verify the manager works
	logger := mgr.(LoggerManager).Logger("test")
	if logger == nil {
		t.Error("Logger() returned nil")
	}

	// Cleanup
	if loggerMgr, ok := mgr.(LoggerManager); ok {
		_ = loggerMgr.Shutdown(context.Background())
	}
}

func TestBuildWithConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.LoggerConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &config.LoggerConfig{
				ConsoleEnabled: true,
				ConsoleConfig:  &config.LogLevelConfig{Level: "info"},
			},
			wantErr: false,
		},
		{
			name: "valid config with file",
			config: &config.LoggerConfig{
				FileEnabled: true,
				FileConfig: &config.FileLogConfig{
					Level: "info",
					Path:  filepath.Join(os.TempDir(), "test.log"),
				},
			},
			wantErr: false,
		},
		{
			name: "invalid config - no outputs",
			config: &config.LoggerConfig{
				ConsoleEnabled: false,
				FileEnabled:    false,
			},
			wantErr: true,
		},
		{
			name: "invalid config - invalid log level",
			config: &config.LoggerConfig{
				ConsoleEnabled: true,
				ConsoleConfig:  &config.LogLevelConfig{Level: "invalid"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := BuildWithConfig(tt.config, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("BuildWithConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if mgr == nil {
					t.Error("BuildWithConfig() returned nil manager")
				}
				// Cleanup
				if loggerMgr, ok := mgr.(LoggerManager); ok {
					_ = loggerMgr.Shutdown(context.Background())
				}
			}
		})
	}
}

func TestBuild_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name         string
		config       map[string]any
		description  string
	}{
		{
			name: "missing file path",
			config: map[string]any{
				"file_enabled": true,
				"file_config": map[string]any{
					"level": "info",
				},
			},
			description: "should return NoneLoggerManagerAdapter when file path is missing",
		},
		{
			name: "invalid file config type",
			config: map[string]any{
				"file_enabled": true,
				"file_config": "invalid",
			},
			description: "should return NoneLoggerManagerAdapter when file config is invalid type",
		},
		{
			name: "invalid console config type",
			config: map[string]any{
				"console_enabled": true,
				"console_config":  "invalid",
			},
			description: "should return LoggerManagerAdapter when console config is invalid type (uses defaults)",
		},
		{
			name: "invalid telemetry config type",
			config: map[string]any{
				"telemetry_enabled": true,
				"telemetry_config":  "invalid",
			},
			description: "should return LoggerManagerAdapter when telemetry config is invalid type (uses defaults)",
		},
		{
			name: "invalid max_size type",
			config: map[string]any{
				"file_enabled": true,
				"file_config": map[string]any{
					"level": "info",
					"path":  filepath.Join(os.TempDir(), "test.log"),
					"rotation": map[string]any{
						"max_size": "invalid",
					},
				},
			},
			description: "should return NoneLoggerManagerAdapter when max_size is invalid",
		},
		{
			name: "invalid max_age type",
			config: map[string]any{
				"file_enabled": true,
				"file_config": map[string]any{
					"level": "info",
					"path":  filepath.Join(os.TempDir(), "test.log"),
					"rotation": map[string]any{
						"max_age": "invalid",
					},
				},
			},
			description: "should return NoneLoggerManagerAdapter when max_age is invalid",
		},
		{
			name: "invalid max_backups type",
			config: map[string]any{
				"file_enabled": true,
				"file_config": map[string]any{
					"level": "info",
					"path":  filepath.Join(os.TempDir(), "test.log"),
					"rotation": map[string]any{
						"max_backups": "invalid",
					},
				},
			},
			description: "should return NoneLoggerManagerAdapter when max_backups is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := Build(tt.config, nil)

			if mgr == nil {
				t.Fatal("Build() returned nil")
			}

			// For invalid config type scenarios, the parser ignores the invalid type and uses defaults
			// So it should return LoggerManagerAdapter, not NoneLoggerManagerAdapter
			if tt.name == "invalid console config type" || tt.name == "invalid telemetry config type" {
				if _, ok := mgr.(*LoggerManagerAdapter); !ok {
					t.Errorf("Build() should return LoggerManagerAdapter for %s (uses defaults), got %T", tt.name, mgr)
				}
			} else {
				// For other error scenarios, should return NoneLoggerManagerAdapter as fallback
				if _, ok := mgr.(*NoneLoggerManagerAdapter); !ok {
					t.Errorf("Build() should return NoneLoggerManagerAdapter for %s, got %T", tt.name, mgr)
				}
			}
		})
	}
}

func TestBuild_DifferentRotationFormats(t *testing.T) {
	tests := []struct {
		name    string
		maxSize any
		maxAge  any
	}{
		{"string size and age", "100MB", "30d"},
		{"int size and age", 100, 30},
		{"GB size", "1GB", "7d"},
		{"hours age", "100MB", "48h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := map[string]any{
				"file_enabled": true,
				"file_config": map[string]any{
					"level": "info",
					"path":  filepath.Join(os.TempDir(), "test.log"),
					"rotation": map[string]any{
						"max_size": tt.maxSize,
						"max_age":  tt.maxAge,
					},
				},
			}

			mgr := Build(cfg, nil)

			// Should return LoggerManagerAdapter
			if _, ok := mgr.(*LoggerManagerAdapter); !ok {
				t.Errorf("Build() should return LoggerManagerAdapter, got %T", mgr)
			}

			// Cleanup
			if loggerMgr, ok := mgr.(LoggerManager); ok {
				_ = loggerMgr.Shutdown(context.Background())
			}
		})
	}
}

func TestBuild_FallbackToNoneLogger(t *testing.T) {
	// Test that config parsing failures result in NoneLoggerManagerAdapter
	errorConfigs := []map[string]any{
		{
			"console_enabled": false,
			"file_enabled":    false,
		},
		{
			"console_enabled": true,
			"console_config": map[string]any{
				"level": "invalid-level",
			},
		},
	}

	for i, cfg := range errorConfigs {
		t.Run("error_config_"+string(rune(i)), func(t *testing.T) {
			mgr := Build(cfg, nil)

			if _, ok := mgr.(*NoneLoggerManagerAdapter); !ok {
				t.Errorf("Build() should fallback to NoneLoggerManagerAdapter, got %T", mgr)
			}
		})
	}
}

func TestBuild_WithTelemetryManager(t *testing.T) {
	cfg := map[string]any{
		"telemetry_enabled": true,
		"telemetry_config": map[string]any{
			"level": "info",
		},
		"console_enabled": true,
		"console_config": map[string]any{
			"level": "debug",
		},
	}

	// Build with nil telemetry manager (should still work)
	mgr := Build(cfg, nil)
	if mgr == nil {
		t.Fatal("Build() returned nil")
	}
	if _, ok := mgr.(*LoggerManagerAdapter); !ok {
		t.Errorf("Build() should return LoggerManagerAdapter, got %T", mgr)
	}
	if loggerMgr, ok := mgr.(LoggerManager); ok {
		_ = loggerMgr.Shutdown(context.Background())
	}

	// Test that the config with telemetry enabled is parsed correctly
	cfg2 := map[string]any{
		"telemetry_enabled": false,
		"console_enabled": true,
		"console_config": map[string]any{
			"level": "info",
		},
	}

	mgr2 := Build(cfg2, nil)
	if mgr2 == nil {
		t.Fatal("Build() returned nil")
	}
	if _, ok := mgr2.(*LoggerManagerAdapter); !ok {
		t.Errorf("Build() should return LoggerManagerAdapter, got %T", mgr2)
	}
	if loggerMgr, ok := mgr2.(LoggerManager); ok {
		_ = loggerMgr.Shutdown(context.Background())
	}
}

func TestBuildWithConfig_Telemetry(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.LoggerConfig
		wantErr bool
	}{
		{
			name: "config with telemetry enabled",
			config: &config.LoggerConfig{
				TelemetryEnabled: true,
				TelemetryConfig:  &config.LogLevelConfig{Level: "info"},
				ConsoleEnabled:  true,
				ConsoleConfig:   &config.LogLevelConfig{Level: "debug"},
			},
			wantErr: false,
		},
		{
			name: "config with telemetry disabled",
			config: &config.LoggerConfig{
				TelemetryEnabled: false,
				ConsoleEnabled:   true,
				ConsoleConfig:    &config.LogLevelConfig{Level: "info"},
			},
			wantErr: false,
		},
		{
			name: "config with invalid telemetry level",
			config: &config.LoggerConfig{
				TelemetryEnabled: true,
				TelemetryConfig:  &config.LogLevelConfig{Level: "invalid"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := BuildWithConfig(tt.config, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("BuildWithConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if mgr == nil {
					t.Error("BuildWithConfig() returned nil manager")
				}
				// Cleanup
				if loggerMgr, ok := mgr.(LoggerManager); ok {
					_ = loggerMgr.Shutdown(context.Background())
				}
			}
		})
	}
}

func TestBuild_InvalidCompressType(t *testing.T) {
	tests := []struct {
		name         string
		compress     any
		description  string
	}{
		{
			name:        "compress is string",
			compress:    "true",
			description: "should return LoggerManagerAdapter when compress is string (uses default)",
		},
		{
			name:        "compress is int",
			compress:    1,
			description: "should return LoggerManagerAdapter when compress is int (uses default)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := map[string]any{
				"file_enabled": true,
				"file_config": map[string]any{
					"level": "info",
					"path":  filepath.Join(os.TempDir(), "test.log"),
					"rotation": map[string]any{
						"compress": tt.compress,
					},
				},
			}

			mgr := Build(cfg, nil)

			// Invalid compress type uses default value, so it should still return LoggerManagerAdapter
			if _, ok := mgr.(*LoggerManagerAdapter); !ok {
				t.Errorf("Build() should return LoggerManagerAdapter for %s, got %T", tt.name, mgr)
			}

			// Cleanup
			if loggerMgr, ok := mgr.(LoggerManager); ok {
				_ = loggerMgr.Shutdown(context.Background())
			}
		})
	}
}
