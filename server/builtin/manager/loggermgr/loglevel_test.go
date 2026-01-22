package loggermgr

import (
	"encoding/json"
	"strings"
	"testing"

	"go.uber.org/zap/zapcore"
)

// TestLogLevelString 测试日志级别字符串表示
func TestLogLevelString(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected string
	}{
		{"Debug level", DebugLevel, "debug"},
		{"Info level", InfoLevel, "info"},
		{"Warn level", WarnLevel, "warn"},
		{"Error level", ErrorLevel, "error"},
		{"Fatal level", FatalLevel, "fatal"},
		{"Invalid level", LogLevel(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestParseLogLevel 测试解析日志级别
func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected LogLevel
	}{
		{"Debug lower", "debug", DebugLevel},
		{"Debug upper", "DEBUG", DebugLevel},
		{"Debug mixed", "DeBuG", DebugLevel},
		{"Info lower", "info", InfoLevel},
		{"Warn lower", "warn", WarnLevel},
		{"Warning full", "warning", WarnLevel},
		{"Error lower", "error", ErrorLevel},
		{"Fatal lower", "fatal", FatalLevel},
		{"With spaces", "  info  ", InfoLevel},
		{"Invalid default", "invalid", InfoLevel},
		{"Empty string", "", InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseLogLevel(tt.input); got != tt.expected {
				t.Errorf("ParseLogLevel() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestIsValidLogLevel 测试日志级别验证
func TestIsValidLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected bool
	}{
		{"Valid debug", "debug", true},
		{"Valid info", "info", true},
		{"Valid warn", "warn", true},
		{"Valid warning", "warning", true},
		{"Valid error", "error", true},
		{"Valid fatal", "fatal", true},
		{"Valid empty", "", true},
		{"Invalid level", "invalid", false},
		{"Invalid partial", "inf", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidLogLevel(tt.level); got != tt.expected {
				t.Errorf("IsValidLogLevel() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestLogLevelValidate 测试日志级别验证方法
func TestLogLevelValidate(t *testing.T) {
	tests := []struct {
		name      string
		level     LogLevel
		wantErr   bool
		errString string
	}{
		{"Valid debug", DebugLevel, false, ""},
		{"Valid info", InfoLevel, false, ""},
		{"Valid warn", WarnLevel, false, ""},
		{"Valid error", ErrorLevel, false, ""},
		{"Valid fatal", FatalLevel, false, ""},
		{"Invalid negative", LogLevel(-1), true, "invalid log level: -1"},
		{"Invalid too large", LogLevel(999), true, "invalid log level: 999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.level.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("LogLevel.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errString) {
				t.Errorf("LogLevel.Validate() error = %v, want contain %v", err, tt.errString)
			}
		})
	}
}

// TestLogLevelMarshalText 测试日志级别序列化
func TestLogLevelMarshalText(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected []byte
	}{
		{"Debug", DebugLevel, []byte("debug")},
		{"Info", InfoLevel, []byte("info")},
		{"Warn", WarnLevel, []byte("warn")},
		{"Error", ErrorLevel, []byte("error")},
		{"Fatal", FatalLevel, []byte("fatal")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.level.MarshalText()
			if err != nil {
				t.Errorf("LogLevel.MarshalText() error = %v", err)
				return
			}
			if string(got) != string(tt.expected) {
				t.Errorf("LogLevel.MarshalText() = %v, want %v", string(got), string(tt.expected))
			}
		})
	}
}

// TestLogLevelUnmarshalText 测试日志级别反序列化
func TestLogLevelUnmarshalText(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected LogLevel
	}{
		{"Debug", []byte("debug"), DebugLevel},
		{"Info", []byte("info"), InfoLevel},
		{"Warn", []byte("warn"), WarnLevel},
		{"Error", []byte("error"), ErrorLevel},
		{"Fatal", []byte("fatal"), FatalLevel},
		{"Invalid", []byte("invalid"), InfoLevel}, // 默认为 InfoLevel
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var level LogLevel
			err := level.UnmarshalText(tt.data)
			if err != nil {
				t.Errorf("LogLevel.UnmarshalText() error = %v", err)
				return
			}
			if level != tt.expected {
				t.Errorf("LogLevel.UnmarshalText() = %v, want %v", level, tt.expected)
			}
		})
	}
}

// TestLogLevelJSONSerialization 测试 JSON 序列化/反序列化
func TestLogLevelJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected string
	}{
		{"Debug", DebugLevel, `"debug"`},
		{"Info", InfoLevel, `"info"`},
		{"Warn", WarnLevel, `"warn"`},
		{"Error", ErrorLevel, `"error"`},
		{"Fatal", FatalLevel, `"fatal"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试序列化
			data, err := json.Marshal(tt.level)
			if err != nil {
				t.Errorf("json.Marshal() error = %v", err)
				return
			}
			if string(data) != tt.expected {
				t.Errorf("json.Marshal() = %v, want %v", string(data), tt.expected)
			}

			// 测试反序列化
			var level LogLevel
			err = json.Unmarshal(data, &level)
			if err != nil {
				t.Errorf("json.Unmarshal() error = %v", err)
				return
			}
			if level != tt.level {
				t.Errorf("json.Unmarshal() = %v, want %v", level, tt.level)
			}
		})
	}
}

// TestLogLevelInt 测试日志级别整数值
func TestLogLevelInt(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected int
	}{
		{"Debug", DebugLevel, 0},
		{"Info", InfoLevel, 1},
		{"Warn", WarnLevel, 2},
		{"Error", ErrorLevel, 3},
		{"Fatal", FatalLevel, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.Int(); got != tt.expected {
				t.Errorf("LogLevel.Int() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestLogLevelToZap 测试 LogLevel 转换为 zapcore.Level
func TestLogLevelToZap(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected zapcore.Level
	}{
		{"Debug", DebugLevel, zapcore.DebugLevel},
		{"Info", InfoLevel, zapcore.InfoLevel},
		{"Warn", WarnLevel, zapcore.WarnLevel},
		{"Error", ErrorLevel, zapcore.ErrorLevel},
		{"Fatal", FatalLevel, zapcore.FatalLevel},
		{"Invalid", LogLevel(127), zapcore.InfoLevel}, // 默认为 InfoLevel
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LogLevelToZap(tt.level); got != tt.expected {
				t.Errorf("LogLevelToZap() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestZapToLogLevel 测试 zapcore.Level 转换为 LogLevel
func TestZapToLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    zapcore.Level
		expected LogLevel
	}{
		{"Debug", zapcore.DebugLevel, DebugLevel},
		{"Info", zapcore.InfoLevel, InfoLevel},
		{"Warn", zapcore.WarnLevel, WarnLevel},
		{"Error", zapcore.ErrorLevel, ErrorLevel},
		{"Fatal", zapcore.FatalLevel, FatalLevel},
		{"Panic", zapcore.PanicLevel, FatalLevel},
		{"DPanic", zapcore.DPanicLevel, FatalLevel},
		{"Invalid", zapcore.Level(127), InfoLevel}, // 默认为 InfoLevel
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ZapToLogLevel(tt.level); got != tt.expected {
				t.Errorf("ZapToLogLevel() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestLogLevelRoundTrip 测试双向转换的一致性
func TestLogLevelRoundTrip(t *testing.T) {
	levels := []LogLevel{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel}

	for _, level := range levels {
		t.Run(level.String(), func(t *testing.T) {
			// LogLevel -> zapcore.Level -> LogLevel
			zapLevel := LogLevelToZap(level)
			backToLogLevel := ZapToLogLevel(zapLevel)
			if backToLogLevel != level {
				t.Errorf("Round trip failed: %v -> %v -> %v", level, zapLevel, backToLogLevel)
			}
		})
	}
}

// TestLogLevelComparison 测试日志级别比较
func TestLogLevelComparison(t *testing.T) {
	tests := []struct {
		name     string
		level1   LogLevel
		level2   LogLevel
		expected int // -1: level1 < level2, 0: equal, 1: level1 > level2
	}{
		{"Debug < Info", DebugLevel, InfoLevel, -1},
		{"Info < Warn", InfoLevel, WarnLevel, -1},
		{"Warn < Error", WarnLevel, ErrorLevel, -1},
		{"Error < Fatal", ErrorLevel, FatalLevel, -1},
		{"Info == Info", InfoLevel, InfoLevel, 0},
		{"Fatal > Error", FatalLevel, ErrorLevel, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result int
			if tt.level1 < tt.level2 {
				result = -1
			} else if tt.level1 > tt.level2 {
				result = 1
			} else {
				result = 0
			}

			if result != tt.expected {
				t.Errorf("Comparison failed: %v vs %v = %d, want %d", tt.level1, tt.level2, result, tt.expected)
			}
		})
	}
}
