package drivers

import (
	"context"

	"com.litelake.litecore/common"
	"com.litelake.litecore/manager/loggermgr/internal/loglevel"
)

// NoneLogger 空日志输出器
// 在不需要日志功能时使用，提供空实现以避免条件判断
type NoneLogger struct{}

// NewNoneLogger 创建空日志输出器
func NewNoneLogger() *NoneLogger {
	return &NoneLogger{}
}

// Debug 空实现
func (l *NoneLogger) Debug(msg string, args ...any) {}

// Info 空实现
func (l *NoneLogger) Info(msg string, args ...any) {}

// Warn 空实现
func (l *NoneLogger) Warn(msg string, args ...any) {}

// Error 空实现
func (l *NoneLogger) Error(msg string, args ...any) {}

// Fatal 空实现（不退出程序，统一 Fatal 语义）
func (l *NoneLogger) Fatal(msg string, args ...any) {
	// NoneLogger 的 Fatal 不退出程序，保持静默
	// 这样可以避免在测试或不需要日志的场景中程序意外退出
}

// With 返回自身
// 实现 Logger 接口
func (l *NoneLogger) With(args ...any) Logger {
	return l
}

// SetLevel 空实现
func (l *NoneLogger) SetLevel(level loglevel.LogLevel) {}

// NoneLoggerManager 空日志管理器
type NoneLoggerManager struct {
	*BaseManager
}

// NewNoneLoggerManager 创建空日志管理器
func NewNoneLoggerManager() *NoneLoggerManager {
	return &NoneLoggerManager{
		BaseManager: NewBaseManager("none-logger"),
	}
}

// GetLogger 返回空日志输出器（内部方法）
func (m *NoneLoggerManager) GetLogger(name string) *NoneLogger {
	return NewNoneLogger()
}

// SetGlobalLevel 空实现
func (m *NoneLoggerManager) SetGlobalLevel(level loglevel.LogLevel) {}

// Logger 返回空日志输出器（实现 LoggerManager 接口）
// 返回一个适配器，符合 Logger 接口
func (m *NoneLoggerManager) Logger(name string) *NoneLogger {
	return m.GetLogger(name)
}

// Shutdown 空实现，无需清理资源
func (m *NoneLoggerManager) Shutdown(ctx context.Context) error {
	return nil
}

// ensure NoneLoggerManager implements common.BaseManager interface
var _ common.BaseManager = (*NoneLoggerManager)(nil)
