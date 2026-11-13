package loggermgr

import "github.com/lite-lake/litecore-go/logger"

type driverDefaultLoggerManager struct {
	ins logger.ILogger
}

// NewDriverDefaultLoggerManager 创建默认日志管理器
func NewDriverDefaultLoggerManager() ILoggerManager {
	return &driverDefaultLoggerManager{
		ins: logger.NewDefaultLogger("DefaultLogger"),
	}
}

func (d *driverDefaultLoggerManager) ManagerName() string {
	return "LoggerDefaultManager"
}

func (d *driverDefaultLoggerManager) Health() error {
	return nil
}

func (d *driverDefaultLoggerManager) OnStart() error {
	return nil
}

func (d *driverDefaultLoggerManager) OnStop() error {
	return nil
}

func (d *driverDefaultLoggerManager) Ins() logger.ILogger {
	return d.ins
}

var _ ILoggerManager = (*driverDefaultLoggerManager)(nil)
