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
			name:     "未知级别",
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
			name:     "小写 debug",
			input:    "debug",
			expected: DebugLevel,
		},
		{
			name:     "大写 DEBUG",
			input:    "DEBUG",
			expected: DebugLevel,
		},
		{
			name:     "混合大小写 DeBuG",
			input:    "DeBuG",
			expected: DebugLevel,
		},
		{
			name:     "带空格的 debug",
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
			name:     "无效级别返回默认 InfoLevel",
			input:    "invalid",
			expected: InfoLevel,
		},
		{
			name:     "空字符串返回默认 InfoLevel",
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
			name:     "有效 debug",
			level:    "debug",
			expected: true,
		},
		{
			name:     "有效 info",
			level:    "info",
			expected: true,
		},
		{
			name:     "有效 warn",
			level:    "warn",
			expected: true,
		},
		{
			name:     "有效 warning",
			level:    "warning",
			expected: true,
		},
		{
			name:     "有效 error",
			level:    "error",
			expected: true,
		},
		{
			name:     "有效 fatal",
			level:    "fatal",
			expected: true,
		},
		{
			name:     "空字符串有效",
			level:    "",
			expected: true,
		},
		{
			name:     "无效级别",
			level:    "invalid",
			expected: false,
		},
		{
			name:     "部分匹配",
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
			name:    "DebugLevel 有效",
			level:   DebugLevel,
			wantErr: false,
		},
		{
			name:    "InfoLevel 有效",
			level:   InfoLevel,
			wantErr: false,
		},
		{
			name:    "WarnLevel 有效",
			level:   WarnLevel,
			wantErr: false,
		},
		{
			name:    "ErrorLevel 有效",
			level:   ErrorLevel,
			wantErr: false,
		},
		{
			name:    "FatalLevel 有效",
			level:   FatalLevel,
			wantErr: false,
		},
		{
			name:    "负数级别无效",
			level:   LogLevel(-1),
			wantErr: true,
		},
		{
			name:    "超出范围的级别无效",
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
			name:     "无效字符串",
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
			name:     "无效级别",
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
			name:  "普通名称",
			input: "TestLogger",
		},
		{
			name:  "空名称",
			input: "",
		},
		{
			name:  "带特殊字符",
			input: "Logger-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewDefaultLogger(tt.input)
			if logger == nil {
				t.Fatal("NewDefaultLogger() 返回 nil")
			}
			if logger.level != InfoLevel {
				t.Errorf("默认级别应该是 InfoLevel, got %v", logger.level)
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
			name:  "设置 DebugLevel",
			level: DebugLevel,
		},
		{
			name:  "设置 InfoLevel",
			level: InfoLevel,
		},
		{
			name:  "设置 WarnLevel",
			level: WarnLevel,
		},
		{
			name:  "设置 ErrorLevel",
			level: ErrorLevel,
		},
		{
			name:  "设置 FatalLevel",
			level: FatalLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.SetLevel(tt.level)
			if logger.level != tt.level {
				t.Errorf("设置级别后应该为 %v, got %v", tt.level, logger.level)
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
			name: "添加单个字段",
			args: []any{"key", "value"},
		},
		{
			name: "添加多个字段",
			args: []any{"key1", "value1", "key2", "value2"},
		},
		{
			name: "添加空字段",
			args: []any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newLogger := logger.With(tt.args...)
			if newLogger == nil {
				t.Fatal("With() 返回 nil")
			}
			if newLogger == logger {
				t.Error("With() 应该返回新的 logger 实例")
			}
		})
	}
}

func TestDefaultLogger_WithPreservesLevel(t *testing.T) {
	logger := NewDefaultLogger("TestLogger")
	logger.SetLevel(DebugLevel)

	newLogger := logger.With("key", "value")
	if newLogger.(*DefaultLogger).level != DebugLevel {
		t.Errorf("With() 应该保持原级别, got %v", newLogger.(*DefaultLogger).level)
	}
}

func TestDefaultLogger_WithAccumulatesFields(t *testing.T) {
	logger := NewDefaultLogger("TestLogger")
	logger1 := logger.With("key1", "value1")
	logger2 := logger1.With("key2", "value2")

	if len(logger2.(*DefaultLogger).extraArgs) != 4 {
		t.Errorf("应该累积所有字段, got %v", len(logger2.(*DefaultLogger).extraArgs))
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
			name:           "DebugLevel 记录所有",
			level:          DebugLevel,
			shouldLogDebug: true,
			shouldLogInfo:  true,
			shouldLogWarn:  true,
			shouldLogError: true,
		},
		{
			name:           "InfoLevel 跳过 Debug",
			level:          InfoLevel,
			shouldLogDebug: false,
			shouldLogInfo:  true,
			shouldLogWarn:  true,
			shouldLogError: true,
		},
		{
			name:           "WarnLevel 跳过 Debug 和 Info",
			level:          WarnLevel,
			shouldLogDebug: false,
			shouldLogInfo:  false,
			shouldLogWarn:  true,
			shouldLogError: true,
		},
		{
			name:           "ErrorLevel 只记录 Error 和 Fatal",
			level:          ErrorLevel,
			shouldLogDebug: false,
			shouldLogInfo:  false,
			shouldLogWarn:  false,
			shouldLogError: true,
		},
		{
			name:           "FatalLevel 只记录 Fatal",
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
				t.Errorf("Debug 日志记录结果: %v, 期望: %v", debugLogged, tt.shouldLogDebug)
			}
			if infoLogged != tt.shouldLogInfo {
				t.Errorf("Info 日志记录结果: %v, 期望: %v", infoLogged, tt.shouldLogInfo)
			}
			if warnLogged != tt.shouldLogWarn {
				t.Errorf("Warn 日志记录结果: %v, 期望: %v", warnLogged, tt.shouldLogWarn)
			}
			if errorLogged != tt.shouldLogError {
				t.Errorf("Error 日志记录结果: %v, 期望: %v", errorLogged, tt.shouldLogError)
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
		t.Error("日志输出应该包含前缀")
	}
	if !strings.Contains(output, "INFO:") {
		t.Error("日志输出应该包含级别")
	}
	if !strings.Contains(output, "test message") {
		t.Error("日志输出应该包含消息")
	}
	if !strings.Contains(output, "key") {
		t.Error("日志输出应该包含字段键")
	}
	if !strings.Contains(output, "value") {
		t.Error("日志输出应该包含字段值")
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
		t.Error("日志输出应该包含 With 添加的字段")
	}
	if !strings.Contains(output, "value1") {
		t.Error("日志输出应该包含 With 添加的字段值")
	}
	if !strings.Contains(output, "key2") {
		t.Error("日志输出应该包含直接传递的字段")
	}
	if !strings.Contains(output, "value2") {
		t.Error("日志输出应该包含直接传递的字段值")
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
