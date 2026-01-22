package common

// ILogger 日志接口
type ILogger interface {
	// Debug 记录调试级别日志
	Debug(msg string, args ...any)

	// Info 记录信息级别日志
	Info(msg string, args ...any)

	// Warn 记录警告级别日志
	Warn(msg string, args ...any)

	// Error 记录错误级别日志
	Error(msg string, args ...any)

	// Fatal 记录致命错误级别日志，然后退出程序
	Fatal(msg string, args ...any)

	// With 返回一个带有额外字段的新 Logger
	With(args ...any) ILogger

	// SetLevel 设置日志级别
	SetLevel(level LogLevel)
}
