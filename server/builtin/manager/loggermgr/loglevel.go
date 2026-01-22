package loggermgr

import (
	"github.com/lite-lake/litecore-go/common"
	"go.uber.org/zap/zapcore"
)

// 日志级别常量
const (
	// DebugLevel 调试级别
	DebugLevel LogLevel = common.DebugLevel
	// InfoLevel 信息级别
	InfoLevel LogLevel = common.InfoLevel
	// WarnLevel 警告级别
	WarnLevel LogLevel = common.WarnLevel
	// ErrorLevel 错误级别
	ErrorLevel LogLevel = common.ErrorLevel
	// FatalLevel 致命错误级别
	FatalLevel LogLevel = common.FatalLevel
)

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
