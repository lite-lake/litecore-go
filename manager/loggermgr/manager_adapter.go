package loggermgr

import (
	"context"

	"com.litelake.litecore/manager/loggermgr/internal/drivers"
	"com.litelake.litecore/manager/loggermgr/internal/loglevel"
)

// LoggerManagerAdapter 适配器，将 drivers.ZapLoggerManager 适配到 LoggerManager 接口
type LoggerManagerAdapter struct {
	driver *drivers.ZapLoggerManager
}

// NewLoggerManagerAdapter 创建日志管理器适配器
func NewLoggerManagerAdapter(driver *drivers.ZapLoggerManager) *LoggerManagerAdapter {
	return &LoggerManagerAdapter{
		driver: driver,
	}
}

// Logger 获取指定名称的 Logger 实例
func (a *LoggerManagerAdapter) Logger(name string) Logger {
	driverLogger := a.driver.GetLogger(name)
	return NewLoggerAdapter(driverLogger)
}

// SetGlobalLevel 设置全局日志级别
func (a *LoggerManagerAdapter) SetGlobalLevel(level LogLevel) {
	a.driver.SetGlobalLevel(loglevel.LogLevelToZap(loglevel.LogLevel(level)))
}

// Shutdown 关闭日志管理器，刷新所有待处理的日志
func (a *LoggerManagerAdapter) Shutdown(ctx context.Context) error {
	return a.driver.Shutdown(ctx)
}

// ManagerName 返回管理器名称
func (a *LoggerManagerAdapter) ManagerName() string {
	return a.driver.ManagerName()
}

// Health 检查管理器健康状态
func (a *LoggerManagerAdapter) Health() error {
	return a.driver.Health()
}

// OnStart 在服务器启动时触发
func (a *LoggerManagerAdapter) OnStart() error {
	return a.driver.OnStart()
}

// OnStop 在服务器停止时触发
func (a *LoggerManagerAdapter) OnStop() error {
	return a.driver.OnStop()
}

// ensure LoggerManagerAdapter implements LoggerManager interface
var _ LoggerManager = (*LoggerManagerAdapter)(nil)
