package common

import (
	"fmt"
	"strings"

	"go.uber.org/zap/zapcore"
)

// LogLevel 日志级别类型
type LogLevel int

const (
	// DebugLevel 调试级别
	DebugLevel LogLevel = iota
	// InfoLevel 信息级别
	InfoLevel
	// WarnLevel 警告级别
	WarnLevel
	// ErrorLevel 错误级别
	ErrorLevel
	// FatalLevel 致命错误级别
	FatalLevel
)

// String 返回日志级别的字符串表示
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	default:
		return "unknown"
	}
}

// ParseLogLevel 从字符串解析日志级别
func ParseLogLevel(level string) LogLevel {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}

// IsValidLogLevel 检查日志级别字符串是否有效
func IsValidLogLevel(level string) bool {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug", "info", "warn", "warning", "error", "fatal", "":
		return true
	default:
		return false
	}
}

// Validate 验证日志级别并返回错误（如果无效）
func (l LogLevel) Validate() error {
	if l < DebugLevel || l > FatalLevel {
		return fmt.Errorf("invalid log level: %d", l)
	}
	return nil
}

// MarshalText 实现 encoding.TextMarshaler 接口，用于 YAML/JSON 序列化
func (l LogLevel) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

// UnmarshalText 实现 encoding.TextUnmarshaler 接口，用于 YAML/JSON 反序列化
func (l *LogLevel) UnmarshalText(data []byte) error {
	level := ParseLogLevel(string(data))
	*l = level
	return nil
}

// Int 返回日志级别的整数值
func (l LogLevel) Int() int {
	return int(l)
}

// LogLevelToZap 转换 LogLevel 到 zapcore.Level
func LogLevelToZap(level LogLevel) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// ZapToLogLevel 转换 zapcore.Level 到 LogLevel
func ZapToLogLevel(level zapcore.Level) LogLevel {
	switch level {
	case zapcore.DebugLevel:
		return DebugLevel
	case zapcore.InfoLevel:
		return InfoLevel
	case zapcore.WarnLevel:
		return WarnLevel
	case zapcore.ErrorLevel:
		return ErrorLevel
	case zapcore.FatalLevel, zapcore.PanicLevel, zapcore.DPanicLevel:
		return FatalLevel
	default:
		return InfoLevel
	}
}
