package drivers

import (
	"context"

	"com.litelake.litecore/manager/loggermgr/internal/loglevel"
)

// Driver 日志驱动接口
// 定义日志驱动需要实现的通用方法
type Driver interface {
	// Start 启动日志驱动
	Start() error

	// Shutdown 关闭日志驱动
	Shutdown(ctx context.Context) error

	// Health 检查驱动健康状态
	Health() error

	// GetLogger 获取指定名称的 Logger 实例
	GetLogger(name string) Logger

	// SetLevel 设置日志级别
	SetLevel(level loglevel.LogLevel)
}

// Logger 日志接口
// 定义日志记录器需要实现的通用方法
type Logger interface {
	// Debug 记录调试级别日志
	Debug(msg string, args ...any)

	// Info 记录信息级别日志
	Info(msg string, args ...any)

	// Warn 记录警告级别日志
	Warn(msg string, args ...any)

	// Error 记录错误级别日志
	Error(msg string, args ...any)

	// Fatal 记录致命错误级别日志
	Fatal(msg string, args ...any)

	// With 返回一个带有额外字段的新 Logger
	With(args ...any) Logger

	// SetLevel 设置日志级别
	SetLevel(level loglevel.LogLevel)
}

// 确保 ZapLogger 实现 Logger 接口
var _ Logger = (*ZapLogger)(nil)

// 确保 NoneLogger 实现 Logger 接口
var _ Logger = (*NoneLogger)(nil)
