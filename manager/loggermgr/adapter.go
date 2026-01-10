package loggermgr

import (
	"com.litelake.litecore/manager/loggermgr/internal/drivers"
	"com.litelake.litecore/manager/loggermgr/internal/loglevel"
)

// LoggerAdapter 适配器，将 drivers.ZapLogger 适配到 Logger 接口
type LoggerAdapter struct {
	driver *drivers.ZapLogger
}

// NewLoggerAdapter 创建日志适配器
func NewLoggerAdapter(driver *drivers.ZapLogger) *LoggerAdapter {
	return &LoggerAdapter{
		driver: driver,
	}
}

// Debug 记录调试级别日志
func (a *LoggerAdapter) Debug(msg string, args ...any) {
	a.driver.Debug(msg, args...)
}

// Info 记录信息级别日志
func (a *LoggerAdapter) Info(msg string, args ...any) {
	a.driver.Info(msg, args...)
}

// Warn 记录警告级别日志
func (a *LoggerAdapter) Warn(msg string, args ...any) {
	a.driver.Warn(msg, args...)
}

// Error 记录错误级别日志
func (a *LoggerAdapter) Error(msg string, args ...any) {
	a.driver.Error(msg, args...)
}

// Fatal 记录致命错误级别日志，然后退出程序
func (a *LoggerAdapter) Fatal(msg string, args ...any) {
	a.driver.Fatal(msg, args...)
}

// With 返回一个带有额外字段的新 Logger
func (a *LoggerAdapter) With(args ...any) Logger {
	// 调用底层驱动的 With 方法，返回新的适配器
	newDriver := a.driver.With(args...)
	return &LoggerAdapter{driver: newDriver}
}

// SetLevel 设置日志级别
func (a *LoggerAdapter) SetLevel(level LogLevel) {
	a.driver.SetLevel(loglevel.LogLevelToZap(loglevel.LogLevel(level)))
}

// ensure LoggerAdapter implements Logger interface
var _ Logger = (*LoggerAdapter)(nil)
