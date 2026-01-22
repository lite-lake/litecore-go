package logger

import "github.com/lite-lake/litecore-go/common"

type LoggerBridge struct {
	name      string
	loggerMgr ILoggerManager
	logger    common.ILogger
}

func NewLoggerBridge(name string) *LoggerBridge {
	b := &LoggerBridge{
		name: name,
	}
	b.initDefaultLogger()
	return b
}

func (b *LoggerBridge) initDefaultLogger() {
	b.logger = newDefaultLogger(b.name)
}

func (b *LoggerBridge) SetLoggerManager(mgr ILoggerManager) {
	b.loggerMgr = mgr
	if mgr != nil {
		b.logger = mgr.Logger(b.name)
	} else {
		b.initDefaultLogger()
	}
}

func (b *LoggerBridge) Debug(msg string, args ...any) {
	b.logger.Debug(msg, args...)
}

func (b *LoggerBridge) Info(msg string, args ...any) {
	b.logger.Info(msg, args...)
}

func (b *LoggerBridge) Warn(msg string, args ...any) {
	b.logger.Warn(msg, args...)
}

func (b *LoggerBridge) Error(msg string, args ...any) {
	b.logger.Error(msg, args...)
}

func (b *LoggerBridge) Fatal(msg string, args ...any) {
	b.logger.Fatal(msg, args...)
}

func (b *LoggerBridge) With(args ...any) common.ILogger {
	return b.logger.With(args...)
}

func (b *LoggerBridge) SetLevel(level common.LogLevel) {
	b.logger.SetLevel(level)
}
