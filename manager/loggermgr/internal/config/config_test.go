package config

import (
	"testing"
)

func TestLoggerConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *LoggerConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with console",
			config: &LoggerConfig{
				ConsoleEnabled: true,
				ConsoleConfig:  &LogLevelConfig{Level: "info"},
			},
			wantErr: false,
		},
		{
			name: "valid config with file",
			config: &LoggerConfig{
				FileEnabled: true,
				FileConfig: &FileLogConfig{
					Level: "info",
					Path:  "/tmp/test.log",
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with telemetry",
			config: &LoggerConfig{
				TelemetryEnabled: true,
				TelemetryConfig:  &LogLevelConfig{Level: "debug"},
			},
			wantErr: false,
		},
		{
			name: "valid config with all enabled",
			config: &LoggerConfig{
				TelemetryEnabled: true,
				TelemetryConfig:  &LogLevelConfig{Level: "info"},
				ConsoleEnabled:   true,
				ConsoleConfig:    &LogLevelConfig{Level: "debug"},
				FileEnabled:      true,
				FileConfig: &FileLogConfig{
					Level: "warn",
					Path:  "/var/log/app.log",
				},
			},
			wantErr: false,
		},
		{
			name:    "no output enabled",
			config:  &LoggerConfig{},
			wantErr: true,
			errMsg:  "at least one logger output",
		},
		{
			name: "invalid telemetry log level",
			config: &LoggerConfig{
				TelemetryEnabled: true,
				TelemetryConfig:  &LogLevelConfig{Level: "invalid"},
			},
			wantErr: true,
			errMsg:  "invalid telemetry log level",
		},
		{
			name: "invalid console log level",
			config: &LoggerConfig{
				ConsoleEnabled: true,
				ConsoleConfig:  &LogLevelConfig{Level: "invalid"},
			},
			wantErr: true,
			errMsg:  "invalid console log level",
		},
		{
			name: "invalid file log level",
			config: &LoggerConfig{
				FileEnabled: true,
				FileConfig: &FileLogConfig{
					Level: "invalid",
					Path:  "/tmp/test.log",
				},
			},
			wantErr: true,
			errMsg:  "invalid file log level",
		},
		{
			name: "missing file path",
			config: &LoggerConfig{
				FileEnabled: true,
				FileConfig: &FileLogConfig{
					Level: "info",
				},
			},
			wantErr: true,
			errMsg:  "file log path is required",
		},
		{
			name: "empty file path",
			config: &LoggerConfig{
				FileEnabled: true,
				FileConfig: &FileLogConfig{
					Level: "info",
					Path:  "",
				},
			},
			wantErr: true,
			errMsg:  "file log path is required",
		},
		{
			name: "nil telemetry config with enabled flag",
			config: &LoggerConfig{
				TelemetryEnabled: true,
			},
			wantErr: false, // nil config is valid when enabled
		},
		{
			name: "nil console config with enabled flag",
			config: &LoggerConfig{
				ConsoleEnabled: true,
			},
			wantErr: false, // nil config is valid when enabled
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoggerConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if got := err.Error(); got[:len(tt.errMsg)] != tt.errMsg {
					t.Errorf("LoggerConfig.Validate() error = %v, want prefix %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestParseLoggerConfigFromMap(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]any
		wantErr bool
		check   func(*testing.T, *LoggerConfig)
	}{
		{
			name:    "nil map returns defaults",
			input:   nil,
			wantErr: false,
			check: func(t *testing.T, cfg *LoggerConfig) {
				if cfg.ConsoleConfig == nil || cfg.ConsoleConfig.Level != "info" {
					t.Errorf("Expected default console level 'info', got %v", cfg.ConsoleConfig)
				}
			},
		},
		{
			name:    "empty map returns defaults",
			input:   map[string]any{},
			wantErr: false,
		},
		{
			name: "parse telemetry_enabled",
			input: map[string]any{
				"telemetry_enabled": true,
			},
			wantErr: false,
			check: func(t *testing.T, cfg *LoggerConfig) {
				if !cfg.TelemetryEnabled {
					t.Error("Expected telemetry_enabled to be true")
				}
			},
		},
		{
			name: "parse console_enabled",
			input: map[string]any{
				"console_enabled": true,
			},
			wantErr: false,
			check: func(t *testing.T, cfg *LoggerConfig) {
				if !cfg.ConsoleEnabled {
					t.Error("Expected console_enabled to be true")
				}
			},
		},
		{
			name: "parse file_enabled",
			input: map[string]any{
				"file_enabled": true,
			},
			wantErr: false,
			check: func(t *testing.T, cfg *LoggerConfig) {
				if !cfg.FileEnabled {
					t.Error("Expected file_enabled to be true")
				}
			},
		},
		{
			name: "parse telemetry_config",
			input: map[string]any{
				"telemetry_enabled": true,
				"telemetry_config": map[string]any{
					"level": "debug",
				},
			},
			wantErr: false,
			check: func(t *testing.T, cfg *LoggerConfig) {
				if cfg.TelemetryConfig == nil || cfg.TelemetryConfig.Level != "debug" {
					t.Errorf("Expected telemetry level 'debug', got %v", cfg.TelemetryConfig)
				}
			},
		},
		{
			name: "parse console_config",
			input: map[string]any{
				"console_enabled": true,
				"console_config": map[string]any{
					"level": "warn",
				},
			},
			wantErr: false,
			check: func(t *testing.T, cfg *LoggerConfig) {
				if cfg.ConsoleConfig == nil || cfg.ConsoleConfig.Level != "warn" {
					t.Errorf("Expected console level 'warn', got %v", cfg.ConsoleConfig)
				}
			},
		},
		{
			name: "parse file_config with path",
			input: map[string]any{
				"file_enabled": true,
				"file_config": map[string]any{
					"level": "error",
					"path":  "/var/log/app.log",
				},
			},
			wantErr: false,
			check: func(t *testing.T, cfg *LoggerConfig) {
				if cfg.FileConfig == nil || cfg.FileConfig.Level != "error" || cfg.FileConfig.Path != "/var/log/app.log" {
					t.Errorf("Expected file level 'error' and path '/var/log/app.log', got %v", cfg.FileConfig)
				}
			},
		},
		{
			name: "parse file_config with rotation",
			input: map[string]any{
				"file_enabled": true,
				"file_config": map[string]any{
					"level": "info",
					"path":  "/tmp/log.log",
					"rotation": map[string]any{
						"max_size":    200,
						"max_age":     60,
						"max_backups": 20,
						"compress":    false,
					},
				},
			},
			wantErr: false,
			check: func(t *testing.T, cfg *LoggerConfig) {
				if cfg.FileConfig == nil || cfg.FileConfig.Rotation == nil {
					t.Error("Expected rotation config")
					return
				}
				if cfg.FileConfig.Rotation.MaxSize != 200 {
					t.Errorf("Expected MaxSize 200, got %d", cfg.FileConfig.Rotation.MaxSize)
				}
				if cfg.FileConfig.Rotation.MaxAge != 60 {
					t.Errorf("Expected MaxAge 60, got %d", cfg.FileConfig.Rotation.MaxAge)
				}
				if cfg.FileConfig.Rotation.MaxBackups != 20 {
					t.Errorf("Expected MaxBackups 20, got %d", cfg.FileConfig.Rotation.MaxBackups)
				}
				if cfg.FileConfig.Rotation.Compress != false {
					t.Errorf("Expected Compress false, got %v", cfg.FileConfig.Rotation.Compress)
				}
			},
		},
		{
			name: "parse full config",
			input: map[string]any{
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
					"path":  "/app/logs/test.log",
					"rotation": map[string]any{
						"max_size":    "100MB",
						"max_age":     "30d",
						"max_backups": 10,
						"compress":    true,
					},
				},
			},
			wantErr: false,
			check: func(t *testing.T, cfg *LoggerConfig) {
				if !cfg.TelemetryEnabled || !cfg.ConsoleEnabled || !cfg.FileEnabled {
					t.Error("Expected all outputs enabled")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLoggerConfigFromMap(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLoggerConfigFromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestParseSizeValue(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected int
		wantErr  bool
	}{
		{"int value", 100, 100, false},
		{"int clamped to max", 20000, MaxSafeSize, false},
		{"int zero", 0, DefaultMaxSize, false},
		{"int negative", -10, DefaultMaxSize, false},
		{"int64 valid", int64(200), 200, false},
		{"int64 out of range", int64(1 << 33), DefaultMaxSize, true},
		{"float64 integer", float64(150), 150, false},
		{"float64 non-integer", float64(150.5), DefaultMaxSize, true},
		{"float64 negative", float64(-10), DefaultMaxSize, true},
		{"string number", "200", 200, false},
		{"string MB", "100MB", 100, false},
		{"string GB", "1GB", 1024, false},
		{"string KB", "2048KB", 2, false},
		{"string mb lowercase", "100mb", 100, false},
		{"string gb lowercase", "1gb", 1024, false},
		{"invalid string", "invalid", DefaultMaxSize, true},
		{"unsupported type", true, DefaultMaxSize, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSizeValue(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSizeValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("parseSizeValue() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseAgeValue(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected int
		wantErr  bool
	}{
		{"int value", 45, 45, false},
		{"int clamped to max", 4000, MaxSafeAge, false},
		{"int zero", 0, DefaultMaxAge, false},
		{"int negative", -10, DefaultMaxAge, false},
		{"int64 valid", int64(60), 60, false},
		{"int64 out of range", int64(1 << 33), DefaultMaxAge, true},
		{"float64 integer", float64(30), 30, false},
		{"float64 non-integer", float64(30.5), DefaultMaxAge, true},
		{"float64 negative", float64(-10), DefaultMaxAge, true},
		{"string number", "30", 30, false},
		{"string days", "30d", 30, false},
		{"string days uppercase", "30D", 30, false},
		{"string hours", "48h", 2, false},
		{"string hours uppercase", "48H", 2, false},
		{"invalid string", "invalid", DefaultMaxAge, true},
		{"unsupported type", true, DefaultMaxAge, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAgeValue(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAgeValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("parseAgeValue() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseBackupsValue(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected int
		wantErr  bool
	}{
		{"int value", 15, 15, false},
		{"int clamped to max", 2000, MaxSafeBackups, false},
		{"int zero", 0, 0, false},
		{"int negative", -10, DefaultMaxBackups, false},
		{"int64 valid", int64(25), 25, false},
		{"int64 out of range", int64(1 << 33), DefaultMaxBackups, true},
		{"float64 integer", float64(20), 20, false},
		{"float64 non-integer", float64(20.5), DefaultMaxBackups, true},
		{"float64 negative", float64(-10), DefaultMaxBackups, true},
		{"string number", "10", 10, false},
		{"string with spaces", " 15 ", 15, false},
		{"invalid string", "invalid", DefaultMaxBackups, true},
		{"unsupported type", true, DefaultMaxBackups, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBackupsValue(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseBackupsValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("parseBackupsValue() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestClampSize(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"normal value", 100, 100},
		{"zero", 0, DefaultMaxSize},
		{"negative", -10, DefaultMaxSize},
		{"max safe", MaxSafeSize, MaxSafeSize},
		{"exceeds max", MaxSafeSize + 1000, MaxSafeSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clampSize(tt.input); got != tt.expected {
				t.Errorf("clampSize() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestClampAge(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"normal value", 30, 30},
		{"zero", 0, DefaultMaxAge},
		{"negative", -10, DefaultMaxAge},
		{"max safe", MaxSafeAge, MaxSafeAge},
		{"exceeds max", MaxSafeAge + 100, MaxSafeAge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clampAge(tt.input); got != tt.expected {
				t.Errorf("clampAge() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestClampBackups(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"normal value", 10, 10},
		{"zero", 0, 0},
		{"negative", -10, DefaultMaxBackups},
		{"max safe", MaxSafeBackups, MaxSafeBackups},
		{"exceeds max", MaxSafeBackups + 100, MaxSafeBackups},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clampBackups(tt.input); got != tt.expected {
				t.Errorf("clampBackups() = %v, want %v", got, tt.expected)
			}
		})
	}
}
