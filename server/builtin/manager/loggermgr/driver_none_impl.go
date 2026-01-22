package loggermgr

import "github.com/lite-lake/litecore-go/logger"

type driverNoneLoggerManager struct {
	ins logger.ILogger
}

func NewDriverNoneLoggerManager() ILoggerManager {
	return &driverNoneLoggerManager{
		ins: &noneLogger{},
	}
}

func (d *driverNoneLoggerManager) ManagerName() string {
	return "LoggerNoneManager"
}

func (d *driverNoneLoggerManager) Health() error {
	return nil
}

func (d *driverNoneLoggerManager) OnStart() error {
	return nil
}

func (d *driverNoneLoggerManager) OnStop() error {
	return nil
}

func (d *driverNoneLoggerManager) Ins() logger.ILogger {
	return d.ins
}

// ---

type noneLogger struct {
}

func (n *noneLogger) Debug(msg string, args ...any) {
}

func (n *noneLogger) Info(msg string, args ...any) {
}

func (n *noneLogger) Warn(msg string, args ...any) {
}

func (n *noneLogger) Error(msg string, args ...any) {
}

func (n *noneLogger) Fatal(msg string, args ...any) {
}

func (n *noneLogger) With(args ...any) logger.ILogger {
	return n
}

func (n *noneLogger) SetLevel(level logger.LogLevel) {
}

var _ logger.ILogger = (*noneLogger)(nil)
