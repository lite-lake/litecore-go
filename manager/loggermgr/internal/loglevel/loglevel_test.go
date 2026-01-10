package loglevel

import (
	"encoding/json"
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DebugLevel, "debug"},
		{InfoLevel, "info"},
		{WarnLevel, "warn"},
		{ErrorLevel, "error"},
		{FatalLevel, "fatal"},
		{LogLevel(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", DebugLevel},
		{"DEBUG", DebugLevel},
		{"  debug  ", DebugLevel},
		{"info", InfoLevel},
		{"warn", WarnLevel},
		{"warning", WarnLevel},
		{"error", ErrorLevel},
		{"fatal", FatalLevel},
		{"invalid", InfoLevel}, // 默认返回 InfoLevel
		{"", InfoLevel},        // 默认返回 InfoLevel
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ParseLogLevel(tt.input); got != tt.expected {
				t.Errorf("ParseLogLevel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsValidLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"debug", true},
		{"info", true},
		{"warn", true},
		{"warning", true},
		{"error", true},
		{"fatal", true},
		{"", true},        // 空字符串视为有效（默认值）
		{"invalid", false},
		{"ERROR", true},   // 大小写不敏感
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsValidLogLevel(tt.input); got != tt.expected {
				t.Errorf("IsValidLogLevel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestLogLevel_Validate(t *testing.T) {
	tests := []struct {
		level    LogLevel
		wantErr  bool
	}{
		{DebugLevel, false},
		{InfoLevel, false},
		{WarnLevel, false},
		{ErrorLevel, false},
		{FatalLevel, false},
		{LogLevel(-1), true},
		{LogLevel(999), true},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			err := tt.level.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("LogLevel.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLogLevel_Int(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected int
	}{
		{DebugLevel, 0},
		{InfoLevel, 1},
		{WarnLevel, 2},
		{ErrorLevel, 3},
		{FatalLevel, 4},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			if got := tt.level.Int(); got != tt.expected {
				t.Errorf("LogLevel.Int() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLogLevel_MarshalText(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
		wantErr  bool
	}{
		{DebugLevel, "debug", false},
		{InfoLevel, "info", false},
		{WarnLevel, "warn", false},
		{ErrorLevel, "error", false},
		{FatalLevel, "fatal", false},
		{LogLevel(999), "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			got, err := tt.level.MarshalText()
			if (err != nil) != tt.wantErr {
				t.Errorf("LogLevel.MarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.expected {
				t.Errorf("LogLevel.MarshalText() = %v, want %v", string(got), tt.expected)
			}
		})
	}
}

func TestLogLevel_UnmarshalText(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		expected LogLevel
		wantErr  bool
	}{
		{"debug", "debug", DebugLevel, false},
		{"info", "info", InfoLevel, false},
		{"warn", "warn", WarnLevel, false},
		{"error", "error", ErrorLevel, false},
		{"fatal", "fatal", FatalLevel, false},
		{"unknown defaults to info", "unknown", InfoLevel, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var level LogLevel
			err := level.UnmarshalText([]byte(tt.data))
			if (err != nil) != tt.wantErr {
				t.Errorf("LogLevel.UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if level != tt.expected {
				t.Errorf("LogLevel.UnmarshalText() = %v, want %v", level, tt.expected)
			}
		})
	}
}

func TestLogLevelToZap(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected zapcore.Level
	}{
		{"DebugLevel", DebugLevel, zapcore.DebugLevel},
		{"InfoLevel", InfoLevel, zapcore.InfoLevel},
		{"WarnLevel", WarnLevel, zapcore.WarnLevel},
		{"ErrorLevel", ErrorLevel, zapcore.ErrorLevel},
		{"FatalLevel", FatalLevel, zapcore.FatalLevel},
		{"InvalidLevel defaults to info", LogLevel(999), zapcore.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LogLevelToZap(tt.level); got != tt.expected {
				t.Errorf("LogLevelToZap() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestZapToLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    zapcore.Level
		expected LogLevel
	}{
		{"DebugLevel", zapcore.DebugLevel, DebugLevel},
		{"InfoLevel", zapcore.InfoLevel, InfoLevel},
		{"WarnLevel", zapcore.WarnLevel, WarnLevel},
		{"ErrorLevel", zapcore.ErrorLevel, ErrorLevel},
		{"FatalLevel", zapcore.FatalLevel, FatalLevel},
		{"PanicLevel", zapcore.PanicLevel, FatalLevel},
		{"DPanicLevel", zapcore.DPanicLevel, FatalLevel},
		{"InvalidLevel defaults to info", zapcore.Level(100), InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ZapToLogLevel(tt.level); got != tt.expected {
				t.Errorf("ZapToLogLevel() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLogLevel_JSONRoundTrip(t *testing.T) {
	tests := []LogLevel{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel}

	for _, original := range tests {
		t.Run(original.String(), func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(original)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			// Unmarshal
			var decoded LogLevel
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			// Verify
			if decoded != original {
				t.Errorf("Round trip failed: got %v, want %v", decoded, original)
			}
		})
	}
}
