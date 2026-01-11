package loggermgr

import (
	"strings"
	"testing"
)

// TestDefaultConfig 测试默认配置
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Driver != "none" {
		t.Errorf("DefaultConfig().Driver = %v, want 'none'", config.Driver)
	}

	if config.ZapConfig == nil {
		t.Error("DefaultConfig().ZapConfig should not be nil")
	}
}

// TestDefaultZapConfig 测试默认 Zap 配置
func TestDefaultZapConfig(t *testing.T) {
	config := DefaultZapConfig()

	if config.TelemetryEnabled {
		t.Error("DefaultZapConfig().TelemetryEnabled should be false")
	}

	if !config.ConsoleEnabled {
		t.Error("DefaultZapConfig().ConsoleEnabled should be true")
	}

	if config.FileEnabled {
		t.Error("DefaultZapConfig().FileEnabled should be false")
	}

	if config.TelemetryConfig == nil {
		t.Error("DefaultZapConfig().TelemetryConfig should not be nil")
	}

	if config.ConsoleConfig == nil {
		t.Error("DefaultZapConfig().ConsoleConfig should not be nil")
	}

	if config.FileConfig == nil {
		t.Error("DefaultZapConfig().FileConfig should not be nil")
	}

	if config.FileConfig.Rotation == nil {
		t.Error("DefaultZapConfig().FileConfig.Rotation should not be nil")
	}

	// 验证默认轮转值
	if config.FileConfig.Rotation.MaxSize != DefaultMaxSize {
		t.Errorf("DefaultZapConfig().FileConfig.Rotation.MaxSize = %v, want %v", config.FileConfig.Rotation.MaxSize, DefaultMaxSize)
	}

	if config.FileConfig.Rotation.MaxAge != DefaultMaxAge {
		t.Errorf("DefaultZapConfig().FileConfig.Rotation.MaxAge = %v, want %v", config.FileConfig.Rotation.MaxAge, DefaultMaxAge)
	}

	if config.FileConfig.Rotation.MaxBackups != DefaultMaxBackups {
		t.Errorf("DefaultZapConfig().FileConfig.Rotation.MaxBackups = %v, want %v", config.FileConfig.Rotation.MaxBackups, DefaultMaxBackups)
	}

	if !config.FileConfig.Rotation.Compress {
		t.Error("DefaultZapConfig().FileConfig.Rotation.Compress should be true")
	}
}

// TestLoggerConfigValidate 测试配置验证
func TestLoggerConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *LoggerConfig
		wantErr   bool
		errString string
	}{
		{
			name: "Valid none driver",
			config: &LoggerConfig{
				Driver: "none",
			},
			wantErr: false,
		},
		{
			name: "Valid zap driver",
			config: &LoggerConfig{
				Driver:    "zap",
				ZapConfig: DefaultZapConfig(),
			},
			wantErr: false,
		},
		{
			name: "Missing driver",
			config: &LoggerConfig{
				Driver: "",
			},
			wantErr:   true,
			errString: "driver is required",
		},
		{
			name: "Unsupported driver",
			config: &LoggerConfig{
				Driver: "unsupported",
			},
			wantErr:   true,
			errString: "unsupported driver",
		},
		{
			name: "Zap driver without zap config",
			config: &LoggerConfig{
				Driver:    "zap",
				ZapConfig: nil,
			},
			wantErr:   true,
			errString: "zap_config is required",
		},
		{
			name: "Driver with spaces",
			config: &LoggerConfig{
				Driver: "  zap  ",
			},
			wantErr:   true,
			errString: "zap_config is required",
		},
		{
			name: "Driver uppercase",
			config: &LoggerConfig{
				Driver:    "ZAP",
				ZapConfig: DefaultZapConfig(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoggerConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errString) {
				t.Errorf("LoggerConfig.Validate() error = %v, want contain %v", err, tt.errString)
			}
		})
	}
}

// TestZapConfigValidate 测试 Zap 配置验证
func TestZapConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *ZapConfig
		wantErr   bool
		errString string
	}{
		{
			name: "Valid console only",
			config: &ZapConfig{
				ConsoleEnabled: true,
				ConsoleConfig:  &LogLevelConfig{Level: "info"},
			},
			wantErr: false,
		},
		{
			name: "Valid file only",
			config: &ZapConfig{
				FileEnabled: true,
				FileConfig: &FileLogConfig{
					Level: "info",
					Path:  "./logs/app.log",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid telemetry only",
			config: &ZapConfig{
				TelemetryEnabled: true,
				TelemetryConfig:  &LogLevelConfig{Level: "info"},
			},
			wantErr: false,
		},
		{
			name: "Valid all enabled",
			config: &ZapConfig{
				TelemetryEnabled: true,
				TelemetryConfig:  &LogLevelConfig{Level: "debug"},
				ConsoleEnabled:   true,
				ConsoleConfig:    &LogLevelConfig{Level: "info"},
				FileEnabled:      true,
				FileConfig: &FileLogConfig{
					Level: "info",
					Path:  "./logs/app.log",
				},
			},
			wantErr: false,
		},
		{
			name:      "No output enabled",
			config:    &ZapConfig{},
			wantErr:   true,
			errString: "at least one logger output",
		},
		{
			name: "Invalid telemetry level",
			config: &ZapConfig{
				TelemetryEnabled: true,
				TelemetryConfig:  &LogLevelConfig{Level: "invalid"},
			},
			wantErr:   true,
			errString: "invalid telemetry log level",
		},
		{
			name: "Invalid console level",
			config: &ZapConfig{
				ConsoleEnabled: true,
				ConsoleConfig:  &LogLevelConfig{Level: "invalid"},
			},
			wantErr:   true,
			errString: "invalid console log level",
		},
		{
			name: "Invalid file level",
			config: &ZapConfig{
				FileEnabled: true,
				FileConfig: &FileLogConfig{
					Level: "invalid",
					Path:  "./logs/app.log",
				},
			},
			wantErr:   true,
			errString: "invalid file log level",
		},
		{
			name: "Missing file path",
			config: &ZapConfig{
				FileEnabled: true,
				FileConfig: &FileLogConfig{
					Level: "info",
					Path:  "",
				},
			},
			wantErr:   true,
			errString: "file log path is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ZapConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errString) {
				t.Errorf("ZapConfig.Validate() error = %v, want contain %v", err, tt.errString)
			}
		})
	}
}

// TestParseLoggerConfigFromMap 测试从 map 解析配置
func TestParseLoggerConfigFromMap(t *testing.T) {
	tests := []struct {
		name      string
		input     map[string]any
		wantErr   bool
		errString string
		validate  func(*testing.T, *LoggerConfig)
	}{
		{
			name:    "Nil config",
			input:   nil,
			wantErr: false,
			validate: func(t *testing.T, cfg *LoggerConfig) {
				if cfg.Driver != "none" {
					t.Errorf("Expected driver 'none', got '%s'", cfg.Driver)
				}
			},
		},
		{
			name: "Valid none driver",
			input: map[string]any{
				"driver": "none",
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *LoggerConfig) {
				if cfg.Driver != "none" {
					t.Errorf("Expected driver 'none', got '%s'", cfg.Driver)
				}
			},
		},
		{
			name: "Valid zap driver with minimal config",
			input: map[string]any{
				"driver": "zap",
				"zap_config": map[string]any{
					"console_enabled": true,
					"console_config": map[string]any{
						"level": "info",
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *LoggerConfig) {
				if cfg.Driver != "zap" {
					t.Errorf("Expected driver 'zap', got '%s'", cfg.Driver)
				}
				if !cfg.ZapConfig.ConsoleEnabled {
					t.Error("Expected console_enabled to be true")
				}
			},
		},
		{
			name: "Valid zap driver with file config",
			input: map[string]any{
				"driver": "zap",
				"zap_config": map[string]any{
					"console_enabled": true,
					"file_enabled":    true,
					"file_config": map[string]any{
						"level": "info",
						"path":  "./logs/app.log",
						"rotation": map[string]any{
							"max_size":    100,
							"max_age":     30,
							"max_backups": 10,
							"compress":    true,
						},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *LoggerConfig) {
				if !cfg.ZapConfig.FileEnabled {
					t.Error("Expected file_enabled to be true")
				}
				if cfg.ZapConfig.FileConfig.Path != "./logs/app.log" {
					t.Errorf("Expected path './logs/app.log', got '%s'", cfg.ZapConfig.FileConfig.Path)
				}
			},
		},
		{
			name: "Driver with spaces",
			input: map[string]any{
				"driver": "  ZAP  ",
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *LoggerConfig) {
				if cfg.Driver != "zap" {
					t.Errorf("Expected driver 'zap', got '%s'", cfg.Driver)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := ParseLoggerConfigFromMap(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLoggerConfigFromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errString) {
				t.Errorf("ParseLoggerConfigFromMap() error = %v, want contain %v", err, tt.errString)
			}
			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

// TestParseSizeValue 测试大小值解析
func TestParseSizeValue(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		expected  int
		wantErr   bool
		errString string
	}{
		{"Int valid", int(100), 100, false, ""},
		{"Int clamped", int(20000), MaxSafeSize, false, ""},
		{"Int zero", int(0), DefaultMaxSize, false, ""},
		{"Int negative", int(-1), DefaultMaxSize, false, ""},
		{"Int64 valid", int64(100), 100, false, ""},
		{"Int64 within range", int64(10000), 10000, false, ""},
		{"Int64 overflow", int64(9999999999), DefaultMaxSize, true, "out of safe range"},
		{"Float64 valid int", float64(100), 100, false, ""},
		{"Float64 not int", float64(100.5), DefaultMaxSize, true, "not an integer"},
		{"Float64 overflow", float64(9999999999), DefaultMaxSize, true, "out of safe range"},
		{"String MB", "100MB", 100, false, ""},
		{"String GB", "1GB", 1024, false, ""},
		{"String KB", "1024KB", 1, false, ""},
		{"String number", "100", 100, false, ""},
		{"String lowercase", "100mb", 100, false, ""},
		{"String mixed", "100Mb", 100, false, ""},
		{"String invalid", "invalid", DefaultMaxSize, true, "invalid size format"},
		{"String invalid unit", "100XB", DefaultMaxSize, true, "unsupported size unit"},
		{"Unsupported type", true, DefaultMaxSize, true, "unsupported type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSizeValue(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSizeValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errString) {
				t.Errorf("parseSizeValue() error = %v, want contain %v", err, tt.errString)
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("parseSizeValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestParseAgeValue 测试时间值解析
func TestParseAgeValue(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		expected  int
		wantErr   bool
		errString string
	}{
		{"Int valid", int(30), 30, false, ""},
		{"Int clamped", int(4000), MaxSafeAge, false, ""},
		{"Int zero", int(0), DefaultMaxAge, false, ""},
		{"Int negative", int(-1), DefaultMaxAge, false, ""},
		{"Int64 valid", int64(30), 30, false, ""},
		{"Int64 overflow", int64(9999999999), DefaultMaxAge, true, "out of safe range"},
		{"Float64 valid int", float64(30), 30, false, ""},
		{"Float64 not int", float64(30.5), DefaultMaxAge, true, "not an integer"},
		{"Float64 overflow", float64(9999999999), DefaultMaxAge, true, "out of safe range"},
		{"String days", "30d", 30, false, ""},
		{"String hours", "48h", 2, false, ""},
		{"String number", "30", 30, false, ""},
		{"String uppercase", "30D", 30, false, ""},
		{"String invalid", "invalid", DefaultMaxAge, true, "invalid age format"},
		{"String invalid unit", "30m", DefaultMaxAge, true, "unsupported age unit"},
		{"Unsupported type", true, DefaultMaxAge, true, "unsupported type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseAgeValue(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAgeValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errString) {
				t.Errorf("parseAgeValue() error = %v, want contain %v", err, tt.errString)
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("parseAgeValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestParseBackupsValue 测试备份数值解析
func TestParseBackupsValue(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		expected  int
		wantErr   bool
		errString string
	}{
		{"Int valid", int(10), 10, false, ""},
		{"Int clamped", int(2000), MaxSafeBackups, false, ""},
		{"Int zero", int(0), 0, false, ""},
		{"Int negative", int(-1), DefaultMaxBackups, false, ""},
		{"Int64 valid", int64(10), 10, false, ""},
		{"Int64 overflow", int64(9999999999), DefaultMaxBackups, true, "out of safe range"},
		{"Float64 valid int", float64(10), 10, false, ""},
		{"Float64 not int", float64(10.5), DefaultMaxBackups, true, "not an integer"},
		{"Float64 overflow", float64(9999999999), DefaultMaxBackups, true, "out of safe range"},
		{"String valid", "10", 10, false, ""},
		{"String with spaces", "  10  ", 10, false, ""},
		{"String invalid", "invalid", DefaultMaxBackups, true, "invalid backups value"},
		{"Unsupported type", true, DefaultMaxBackups, true, "unsupported type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseBackupsValue(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseBackupsValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errString) {
				t.Errorf("parseBackupsValue() error = %v, want contain %v", err, tt.errString)
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("parseBackupsValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestClampSize 测试大小限制函数
func TestClampSize(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"Normal value", 100, 100},
		{"Zero", 0, DefaultMaxSize},
		{"Negative", -1, DefaultMaxSize},
		{"Max safe", MaxSafeSize, MaxSafeSize},
		{"Over max safe", MaxSafeSize + 1, MaxSafeSize},
		{"Very large", 999999, MaxSafeSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClampSize(tt.input)
			if result != tt.expected {
				t.Errorf("ClampSize() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestClampAge 测试时间限制函数
func TestClampAge(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"Normal value", 30, 30},
		{"Zero", 0, DefaultMaxAge},
		{"Negative", -1, DefaultMaxAge},
		{"Max safe", MaxSafeAge, MaxSafeAge},
		{"Over max safe", MaxSafeAge + 1, MaxSafeAge},
		{"Very large", 99999, MaxSafeAge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClampAge(tt.input)
			if result != tt.expected {
				t.Errorf("ClampAge() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestClampBackups 测试备份数限制函数
func TestClampBackups(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"Normal value", 10, 10},
		{"Zero", 0, 0},
		{"Negative", -1, DefaultMaxBackups},
		{"Max safe", MaxSafeBackups, MaxSafeBackups},
		{"Over max safe", MaxSafeBackups + 1, MaxSafeBackups},
		{"Very large", 99999, MaxSafeBackups},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClampBackups(tt.input)
			if result != tt.expected {
				t.Errorf("ClampBackups() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestParseRotationConfig 测试轮转配置解析
func TestParseRotationConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		validate func(*testing.T, *RotationConfig)
		wantErr  bool
	}{
		{
			name:  "Nil config",
			input: nil,
			validate: func(t *testing.T, cfg *RotationConfig) {
				if cfg.MaxSize != DefaultMaxSize {
					t.Errorf("Expected MaxSize %d, got %d", DefaultMaxSize, cfg.MaxSize)
				}
			},
			wantErr: false,
		},
		{
			name: "Full config",
			input: map[string]any{
				"max_size":    200,
				"max_age":     60,
				"max_backups": 20,
				"compress":    false,
			},
			validate: func(t *testing.T, cfg *RotationConfig) {
				if cfg.MaxSize != 200 {
					t.Errorf("Expected MaxSize 200, got %d", cfg.MaxSize)
				}
				if cfg.MaxAge != 60 {
					t.Errorf("Expected MaxAge 60, got %d", cfg.MaxAge)
				}
				if cfg.MaxBackups != 20 {
					t.Errorf("Expected MaxBackups 20, got %d", cfg.MaxBackups)
				}
				if cfg.Compress {
					t.Error("Expected Compress to be false")
				}
			},
			wantErr: false,
		},
		{
			name: "String size values",
			input: map[string]any{
				"max_size":    "100MB",
				"max_age":     "30d",
				"max_backups": "10",
			},
			validate: func(t *testing.T, cfg *RotationConfig) {
				if cfg.MaxSize != 100 {
					t.Errorf("Expected MaxSize 100, got %d", cfg.MaxSize)
				}
				if cfg.MaxAge != 30 {
					t.Errorf("Expected MaxAge 30, got %d", cfg.MaxAge)
				}
				if cfg.MaxBackups != 10 {
					t.Errorf("Expected MaxBackups 10, got %d", cfg.MaxBackups)
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseRotationConfig(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRotationConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}
