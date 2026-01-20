package loggermgr

import (
	"context"
)

// noneLoggerImpl 空日志输出器
// 在不需要日志功能时使用，提供空实现以避免条件判断
type noneLoggerImpl struct{}

// newNoneLoggerImpl 创建空日志输出器
func newNoneLoggerImpl() *noneLoggerImpl {
	return &noneLoggerImpl{}
}

// Debug 空实现
func (l *noneLoggerImpl) Debug(msg string, args ...any) {}

// Info 空实现
func (l *noneLoggerImpl) Info(msg string, args ...any) {}

// Warn 空实现
func (l *noneLoggerImpl) Warn(msg string, args ...any) {}

// Error 空实现
func (l *noneLoggerImpl) Error(msg string, args ...any) {}

// Fatal 空实现（不退出程序，统一 Fatal 语义）
func (l *noneLoggerImpl) Fatal(msg string, args ...any) {
	// NoneLogger 的 Fatal 不退出程序，保持静默
	// 这样可以避免在测试或不需要日志的场景中程序意外退出
}

// With 返回自身
func (l *noneLoggerImpl) With(args ...any) ILogger {
	return l
}

// SetLevel 空实现
func (l *noneLoggerImpl) SetLevel(level LogLevel) {}

// noneLoggerManagerImpl 空日志管理器
type noneLoggerManagerImpl struct {
	*loggerManagerBaseImpl
}

// NewLoggerManagerNoneImpl 创建空日志管理器实现
func NewLoggerManagerNoneImpl() ILoggerManager {
	return &noneLoggerManagerImpl{
		loggerManagerBaseImpl: newLoggerManagerBaseImpl("none-logger"),
	}
}

// Logger 返回空日志输出器
func (m *noneLoggerManagerImpl) Logger(name string) ILogger {
	return newNoneLoggerImpl()
}

// SetGlobalLevel 空实现
func (m *noneLoggerManagerImpl) SetGlobalLevel(level LogLevel) {}

// Shutdown 空实现，无需清理资源
func (m *noneLoggerManagerImpl) Shutdown(ctx context.Context) error {
	return nil
}
