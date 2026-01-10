package loggermgr

import (
	"context"

	"com.litelake.litecore/manager/loggermgr/internal/drivers"
	"com.litelake.litecore/manager/loggermgr/internal/loglevel"
)

// NoneLoggerManagerAdapter 适配器，将 drivers.NoneLoggerManager 适配到 LoggerManager 接口
type NoneLoggerManagerAdapter struct {
	driver *drivers.NoneLoggerManager
}

// NewNoneLoggerManagerAdapter 创建空日志管理器适配器
func NewNoneLoggerManagerAdapter(driver *drivers.NoneLoggerManager) *NoneLoggerManagerAdapter {
	return &NoneLoggerManagerAdapter{
		driver: driver,
	}
}

// Logger 获取指定名称的 Logger 实例
func (a *NoneLoggerManagerAdapter) Logger(name string) Logger {
	driverLogger := a.driver.GetLogger(name)
	return &NoneLoggerAdapter{driver: driverLogger}
}

// SetGlobalLevel 设置全局日志级别
func (a *NoneLoggerManagerAdapter) SetGlobalLevel(level LogLevel) {
	a.driver.SetGlobalLevel(loglevel.LogLevel(level))
}

// Shutdown 关闭日志管理器，刷新所有待处理的日志
func (a *NoneLoggerManagerAdapter) Shutdown(ctx context.Context) error {
	return a.driver.Shutdown(ctx)
}

// ManagerName 返回管理器名称
func (a *NoneLoggerManagerAdapter) ManagerName() string {
	return a.driver.ManagerName()
}

// Health 检查管理器健康状态
func (a *NoneLoggerManagerAdapter) Health() error {
	return a.driver.Health()
}

// OnStart 在服务器启动时触发
func (a *NoneLoggerManagerAdapter) OnStart() error {
	return a.driver.OnStart()
}

// OnStop 在服务器停止时触发
func (a *NoneLoggerManagerAdapter) OnStop() error {
	return a.driver.OnStop()
}

// NoneLoggerAdapter 适配器，将 drivers.NoneLogger 适配到 Logger 接口
type NoneLoggerAdapter struct {
	driver *drivers.NoneLogger
}

// Debug 空实现
func (a *NoneLoggerAdapter) Debug(msg string, args ...any) {}

// Info 空实现
func (a *NoneLoggerAdapter) Info(msg string, args ...any) {}

// Warn 空实现
func (a *NoneLoggerAdapter) Warn(msg string, args ...any) {}

// Error 空实现
func (a *NoneLoggerAdapter) Error(msg string, args ...any) {}

// Fatal 空实现（不退出程序，统一 Fatal 语义）
func (a *NoneLoggerAdapter) Fatal(msg string, args ...any) {
	// NoneLogger 的 Fatal 不退出程序，保持静默
	// 这样可以避免在测试或不需要日志的场景中程序意外退出
}

// With 返回自身
func (a *NoneLoggerAdapter) With(args ...any) Logger {
	return a
}

// SetLevel 空实现
func (a *NoneLoggerAdapter) SetLevel(level LogLevel) {}

// ensure NoneLoggerManagerAdapter implements LoggerManager interface
var _ LoggerManager = (*NoneLoggerManagerAdapter)(nil)

// ensure NoneLoggerAdapter implements Logger interface
var _ Logger = (*NoneLoggerAdapter)(nil)
