package drivers

import (
	"context"

	"com.litelake.litecore/manager/loggermgr/internal/loglevel"
)

// NoneDriver 空日志驱动
// 实现 Driver 接口，提供空实现以避免条件判断
type NoneDriver struct{}

// NewNoneDriver 创建空日志驱动
func NewNoneDriver() *NoneDriver {
	return &NoneDriver{}
}

// Start 空实现
func (d *NoneDriver) Start() error {
	return nil
}

// Shutdown 空实现
func (d *NoneDriver) Shutdown(ctx context.Context) error {
	return nil
}

// Health 空实现
func (d *NoneDriver) Health() error {
	return nil
}

// GetLogger 返回空日志记录器
func (d *NoneDriver) GetLogger(name string) Logger {
	return NewNoneLogger()
}

// SetLevel 空实现
func (d *NoneDriver) SetLevel(level loglevel.LogLevel) {
	// 空实现，不设置任何级别
}

// 确保 NoneDriver 实现 Driver 接口
var _ Driver = (*NoneDriver)(nil)
