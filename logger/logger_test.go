package logger

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected string
	}{
		{
			name:     "DebugLevel",
			level:    DebugLevel,
			expected: "debug",
		},
		{
			name:     "InfoLevel",
			level:    InfoLevel,
			expected: "info",
		},
		{
			name:     "WarnLevel",
			level:    WarnLevel,
			expected: "warn",
		},
		{
			name:     "ErrorLevel",
			level:    ErrorLevel,
			expected: "error",
		},
		{
			name:     "FatalLevel",
			level:    FatalLevel,
			expected: "fatal",
		},
		{
			name:     "unknown level",
			level:    LogLevel(99),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected LogLevel
	}{
		{
			name:     "lowercase debug",
			input:    "debug",
			expected: DebugLevel,
		},
		{
			name:     "uppercase DEBUG",
			input:    "DEBUG",
			expected: DebugLevel,
		},
		{
			name:     "mixed case DeBuG",
			input:    "DeBuG",
			expected: DebugLevel,
		},
		{
			name:     "debug with spaces",
			input:    " debug ",
			expected: DebugLevel,
		},
		{
			name:     "info",
			input:    "info",
			expected: InfoLevel,
		},
		{
			name:     "warn",
			input:    "warn",
			expected: WarnLevel,
		},
		{
			name:     "warning",
			input:    "warning",
			expected: WarnLevel,
		},
		{
			name:     "error",
			input:    "error",
			expected: ErrorLevel,
		},
		{
			name:     "fatal",
			input:    "fatal",
			expected: FatalLevel,
		},
		{
			name:     "invalid level returns default InfoLevel",
			input:    "invalid",
			expected: InfoLevel,
		},
		{
			name:     "empty string returns default InfoLevel",
			input:    "",
			expected: InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseLogLevel(tt.input); got != tt.expected {
				t.Errorf("ParseLogLevel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsValidLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected bool
	}{
		{
			name:     "valid debug",
			level:    "debug",
			expected: true,
		},
		{
			name:     "valid info",
			level:    "info",
			expected: true,
		},
		{
			name:     "valid warn",
			level:    "warn",
			expected: true,
		},
		{
			name:     "valid warning",
			level:    "warning",
			expected: true,
		},
		{
			name:     "valid error",
			level:    "error",
			expected: true,
		},
		{
			name:     "valid fatal",
			level:    "fatal",
			expected: true,
		},
		{
			name:     "empty string is valid",
			level:    "",
			expected: true,
		},
		{
			name:     "invalid level",
			level:    "invalid",
			expected: false,
		},
		{
			name:     "partial match",
			level:    "debugging",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidLogLevel(tt.level); got != tt.expected {
				t.Errorf("IsValidLogLevel(%q) = %v, want %v", tt.level, got, tt.expected)
			}
		})
	}
}

func TestLogLevel_Validate(t *testing.T) {
	tests := []struct {
		name    string
		level   LogLevel
		wantErr bool
	}{
		{
			name:    "DebugLevel is valid",
			level:   DebugLevel,
			wantErr: false,
		},
		{
			name:    "InfoLevel is valid",
			level:   InfoLevel,
			wantErr: false,
		},
		{
			name:    "WarnLevel is valid",
			level:   WarnLevel,
			wantErr: false,
		},
		{
			name:    "ErrorLevel is valid",
			level:   ErrorLevel,
			wantErr: false,
		},
		{
			name:    "FatalLevel is valid",
			level:   FatalLevel,
			wantErr: false,
		},
		{
			name:    "negative level is invalid",
			level:   LogLevel(-1),
			wantErr: true,
		},
		{
			name:    "out of range level is invalid",
			level:   LogLevel(99),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.level.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("LogLevel.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLogLevel_MarshalText(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected []byte
	}{
		{
			name:     "DebugLevel",
			level:    DebugLevel,
			expected: []byte("debug"),
		},
		{
			name:     "InfoLevel",
			level:    InfoLevel,
			expected: []byte("info"),
		},
		{
			name:     "WarnLevel",
			level:    WarnLevel,
			expected: []byte("warn"),
		},
		{
			name:     "ErrorLevel",
			level:    ErrorLevel,
			expected: []byte("error"),
		},
		{
			name:     "FatalLevel",
			level:    FatalLevel,
			expected: []byte("fatal"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.level.MarshalText()
			if err != nil {
				t.Errorf("LogLevel.MarshalText() error = %v", err)
				return
			}
			if string(got) != string(tt.expected) {
				t.Errorf("LogLevel.MarshalText() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLogLevel_UnmarshalText(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected LogLevel
	}{
		{
			name:     "debug",
			data:     []byte("debug"),
			expected: DebugLevel,
		},
		{
			name:     "INFO",
			data:     []byte("INFO"),
			expected: InfoLevel,
		},
		{
			name:     "warning",
			data:     []byte("warning"),
			expected: WarnLevel,
		},
		{
			name:     "ERROR",
			data:     []byte("ERROR"),
			expected: ErrorLevel,
		},
		{
			name:     "fatal",
			data:     []byte("fatal"),
			expected: FatalLevel,
		},
		{
			name:     "invalid string",
			data:     []byte("invalid"),
			expected: InfoLevel,
		},
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

func TestLogLevel_Int(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected int
	}{
		{
			name:     "DebugLevel",
			level:    DebugLevel,
			expected: 0,
		},
		{
			name:     "InfoLevel",
			level:    InfoLevel,
			expected: 1,
		},
		{
			name:     "WarnLevel",
			level:    WarnLevel,
			expected: 2,
		},
		{
			name:     "ErrorLevel",
			level:    ErrorLevel,
			expected: 3,
		},
		{
			name:     "FatalLevel",
			level:    FatalLevel,
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.Int(); got != tt.expected {
				t.Errorf("LogLevel.Int() = %v, want %v", got, tt.expected)
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
		{
			name:     "DebugLevel",
			level:    DebugLevel,
			expected: zapcore.DebugLevel,
		},
		{
			name:     "InfoLevel",
			level:    InfoLevel,
			expected: zapcore.InfoLevel,
		},
		{
			name:     "WarnLevel",
			level:    WarnLevel,
			expected: zapcore.WarnLevel,
		},
		{
			name:     "ErrorLevel",
			level:    ErrorLevel,
			expected: zapcore.ErrorLevel,
		},
		{
			name:     "FatalLevel",
			level:    FatalLevel,
			expected: zapcore.FatalLevel,
		},
		{
			name:     "invalid level",
			level:    LogLevel(99),
			expected: zapcore.InfoLevel,
		},
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
		{
			name:     "DebugLevel",
			level:    zapcore.DebugLevel,
			expected: DebugLevel,
		},
		{
			name:     "InfoLevel",
			level:    zapcore.InfoLevel,
			expected: InfoLevel,
		},
		{
			name:     "WarnLevel",
			level:    zapcore.WarnLevel,
			expected: WarnLevel,
		},
		{
			name:     "ErrorLevel",
			level:    zapcore.ErrorLevel,
			expected: ErrorLevel,
		},
		{
			name:     "FatalLevel",
			level:    zapcore.FatalLevel,
			expected: FatalLevel,
		},
		{
			name:     "PanicLevel",
			level:    zapcore.PanicLevel,
			expected: FatalLevel,
		},
		{
			name:     "DPanicLevel",
			level:    zapcore.DPanicLevel,
			expected: FatalLevel,
		},
		{
			name:     "无效级别",
			level:    zapcore.Level(99),
			expected: InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ZapToLogLevel(tt.level); got != tt.expected {
				t.Errorf("ZapToLogLevel() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewDefaultLogger(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "normal name",
			input: "TestLogger",
		},
		{
			name:  "empty name",
			input: "",
		},
		{
			name:  "name with special chars",
			input: "Logger-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewDefaultLogger(tt.input)
			if logger == nil {
				t.Fatal("NewDefaultLogger() returned nil")
			}
			if logger.level != InfoLevel {
				t.Errorf("default level should be InfoLevel, got %v", logger.level)
			}
		})
	}
}

func TestDefaultLogger_SetLevel(t *testing.T) {
	logger := NewDefaultLogger("TestLogger")

	tests := []struct {
		name  string
		level LogLevel
	}{
		{
			name:  "set DebugLevel",
			level: DebugLevel,
		},
		{
			name:  "set InfoLevel",
			level: InfoLevel,
		},
		{
			name:  "set WarnLevel",
			level: WarnLevel,
		},
		{
			name:  "set ErrorLevel",
			level: ErrorLevel,
		},
		{
			name:  "set FatalLevel",
			level: FatalLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.SetLevel(tt.level)
			if logger.level != tt.level {
				t.Errorf("level should be %v after setting, got %v", tt.level, logger.level)
			}
		})
	}
}

func TestDefaultLogger_With(t *testing.T) {
	logger := NewDefaultLogger("TestLogger")

	tests := []struct {
		name string
		args []any
	}{
		{
			name: "add single field",
			args: []any{"key", "value"},
		},
		{
			name: "add multiple fields",
			args: []any{"key1", "value1", "key2", "value2"},
		},
		{
			name: "add empty fields",
			args: []any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newLogger := logger.With(tt.args...)
			if newLogger == nil {
				t.Fatal("With() returned nil")
			}
			if newLogger == logger {
				t.Error("With() should return new logger instance")
			}
		})
	}
}

func TestDefaultLogger_WithPreservesLevel(t *testing.T) {
	logger := NewDefaultLogger("TestLogger")
	logger.SetLevel(DebugLevel)

	newLogger := logger.With("key", "value")
	if newLogger.(*DefaultLogger).level != DebugLevel {
		t.Errorf("With() should preserve original level, got %v", newLogger.(*DefaultLogger).level)
	}
}

func TestDefaultLogger_WithAccumulatesFields(t *testing.T) {
	logger := NewDefaultLogger("TestLogger")
	logger1 := logger.With("key1", "value1")
	logger2 := logger1.With("key2", "value2")

	if len(logger2.(*DefaultLogger).extraArgs) != 4 {
		t.Errorf("should accumulate all fields, got %v", len(logger2.(*DefaultLogger).extraArgs))
	}
}

func TestDefaultLogger_LevelFiltering(t *testing.T) {
	tests := []struct {
		name           string
		level          LogLevel
		shouldLogDebug bool
		shouldLogInfo  bool
		shouldLogWarn  bool
		shouldLogError bool
	}{
		{
			name:           "DebugLevel logs all",
			level:          DebugLevel,
			shouldLogDebug: true,
			shouldLogInfo:  true,
			shouldLogWarn:  true,
			shouldLogError: true,
		},
		{
			name:           "InfoLevel skips Debug",
			level:          InfoLevel,
			shouldLogDebug: false,
			shouldLogInfo:  true,
			shouldLogWarn:  true,
			shouldLogError: true,
		},
		{
			name:           "WarnLevel skips Debug and Info",
			level:          WarnLevel,
			shouldLogDebug: false,
			shouldLogInfo:  false,
			shouldLogWarn:  true,
			shouldLogError: true,
		},
		{
			name:           "ErrorLevel only logs Error and Fatal",
			level:          ErrorLevel,
			shouldLogDebug: false,
			shouldLogInfo:  false,
			shouldLogWarn:  false,
			shouldLogError: true,
		},
		{
			name:           "FatalLevel only logs Fatal",
			level:          FatalLevel,
			shouldLogDebug: false,
			shouldLogInfo:  false,
			shouldLogWarn:  false,
			shouldLogError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)
			log.SetFlags(0)

			logger := NewDefaultLogger("TestLogger")
			logger.SetLevel(tt.level)

			logger.Debug("debug message")
			debugLogged := strings.Contains(buf.String(), "DEBUG:")

			buf.Reset()
			logger.Info("info message")
			infoLogged := strings.Contains(buf.String(), "INFO:")

			buf.Reset()
			logger.Warn("warn message")
			warnLogged := strings.Contains(buf.String(), "WARN:")

			buf.Reset()
			logger.Error("error message")
			errorLogged := strings.Contains(buf.String(), "ERROR:")

			if debugLogged != tt.shouldLogDebug {
				t.Errorf("Debug log result: %v, expected: %v", debugLogged, tt.shouldLogDebug)
			}
			if infoLogged != tt.shouldLogInfo {
				t.Errorf("Info log result: %v, expected: %v", infoLogged, tt.shouldLogInfo)
			}
			if warnLogged != tt.shouldLogWarn {
				t.Errorf("Warn log result: %v, expected: %v", warnLogged, tt.shouldLogWarn)
			}
			if errorLogged != tt.shouldLogError {
				t.Errorf("Error log result: %v, expected: %v", errorLogged, tt.shouldLogError)
			}

			log.SetOutput(os.Stderr)
		})
	}
}

func TestDefaultLogger_LogOutputFormat(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetFlags(0)
	defer log.SetOutput(os.Stderr)

	logger := NewDefaultLogger("TestLogger")
	logger.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "[TestLogger]") {
		t.Error("log output should contain prefix")
	}
	if !strings.Contains(output, "INFO:") {
		t.Error("log output should contain level")
	}
	if !strings.Contains(output, "test message") {
		t.Error("log output should contain message")
	}
	if !strings.Contains(output, "key") {
		t.Error("log output should contain field key")
	}
	if !strings.Contains(output, "value") {
		t.Error("log output should contain field value")
	}
}

func TestDefaultLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetFlags(0)
	defer log.SetOutput(os.Stderr)

	logger := NewDefaultLogger("TestLogger")
	logger.With("key1", "value1").Info("test message", "key2", "value2")

	output := buf.String()
	if !strings.Contains(output, "key1") {
		t.Error("log output should contain field added by With")
	}
	if !strings.Contains(output, "value1") {
		t.Error("log output should contain field value added by With")
	}
	if !strings.Contains(output, "key2") {
		t.Error("log output should contain directly passed field")
	}
	if !strings.Contains(output, "value2") {
		t.Error("log output should contain directly passed field value")
	}
}

func BenchmarkParseLogLevel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseLogLevel("info")
	}
}

func BenchmarkLogLevel_String(b *testing.B) {
	level := InfoLevel
	for i := 0; i < b.N; i++ {
		_ = level.String()
	}
}

func BenchmarkLogLevelToZap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LogLevelToZap(InfoLevel)
	}
}

func BenchmarkZapToLogLevel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ZapToLogLevel(zapcore.InfoLevel)
	}
}
