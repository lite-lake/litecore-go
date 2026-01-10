package loggermgr

import (
	"context"

	"com.litelake.litecore/manager/loggermgr/internal/loglevel"
)

// LogLevel 日志级别类型（对外暴露的类型别名）
type LogLevel = loglevel.LogLevel

const (
	// DebugLevel 调试级别
	DebugLevel = loglevel.DebugLevel
	// InfoLevel 信息级别
	InfoLevel = loglevel.InfoLevel
	// WarnLevel 警告级别
	WarnLevel = loglevel.WarnLevel
	// ErrorLevel 错误级别
	ErrorLevel = loglevel.ErrorLevel
	// FatalLevel 致命错误级别
	FatalLevel = loglevel.FatalLevel
)

// ParseLogLevel 从字符串解析日志级别
var ParseLogLevel = loglevel.ParseLogLevel

// Logger 日志接口
type Logger interface {
	// Debug 记录调试级别日志
	Debug(msg string, args ...any)

	// Info 记录信息级别日志
	Info(msg string, args ...any)

	// Warn 记录警告级别日志
	Warn(msg string, args ...any)

	// Error 记录错误级别日志
	Error(msg string, args ...any)

	// Fatal 记录致命错误级别日志，然后退出程序
	Fatal(msg string, args ...any)

	// With 返回一个带有额外字段的新 Logger
	With(args ...any) Logger

	// SetLevel 设置日志级别
	SetLevel(level LogLevel)
}

// LoggerManager 日志管理器接口
type LoggerManager interface {
	// Logger 获取指定名称的 Logger 实例
	Logger(name string) Logger

	// SetGlobalLevel 设置全局日志级别
	SetGlobalLevel(level LogLevel)

	// Shutdown 关闭日志管理器，刷新所有待处理的日志
	Shutdown(ctx context.Context) error
}
