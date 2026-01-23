package logger

import (
	"log"
)

// DefaultLogger 默认日志记录器实现
type DefaultLogger struct {
	prefix    string   // 日志前缀
	level     LogLevel // 当前日志级别
	extraArgs []any    // 额外的字段
}

// NewDefaultLogger 创建默认日志记录器
func NewDefaultLogger(name string) *DefaultLogger {
	return &DefaultLogger{
		prefix:    "[" + name + "] ",
		level:     InfoLevel,
		extraArgs: []any{},
	}
}

// Debug 记录调试级别日志
func (l *DefaultLogger) Debug(msg string, args ...any) {
	if l.level > DebugLevel {
		return
	}
	allArgs := append(l.extraArgs, args...)
	log.Printf(l.prefix+"DEBUG: %s %v", msg, allArgs)
}

// Info 记录信息级别日志
func (l *DefaultLogger) Info(msg string, args ...any) {
	if l.level > InfoLevel {
		return
	}
	allArgs := append(l.extraArgs, args...)
	log.Printf(l.prefix+"INFO: %s %v", msg, allArgs)
}

// Warn 记录警告级别日志
func (l *DefaultLogger) Warn(msg string, args ...any) {
	if l.level > WarnLevel {
		return
	}
	allArgs := append(l.extraArgs, args...)
	log.Printf(l.prefix+"WARN: %s %v", msg, allArgs)
}

// Error 记录错误级别日志
func (l *DefaultLogger) Error(msg string, args ...any) {
	if l.level > ErrorLevel {
		return
	}
	allArgs := append(l.extraArgs, args...)
	log.Printf(l.prefix+"ERROR: %s %v", msg, allArgs)
}

// Fatal 记录致命错误级别日志，然后退出程序
func (l *DefaultLogger) Fatal(msg string, args ...any) {
	allArgs := append(l.extraArgs, args...)
	log.Printf(l.prefix+"FATAL: %s %v", msg, allArgs)
	args = append([]any{l.prefix + "FATAL: " + msg}, args...)
	log.Fatal(args...)
}

// With 返回一个带有额外字段的新 Logger
func (l *DefaultLogger) With(args ...any) ILogger {
	newLogger := &DefaultLogger{
		prefix:    l.prefix,
		level:     l.level,
		extraArgs: append(append([]any{}, l.extraArgs...), args...),
	}
	return newLogger
}

// SetLevel 设置日志级别
func (l *DefaultLogger) SetLevel(level LogLevel) {
	l.level = level
}

// ensure DefaultLogger implements ILogger
var _ ILogger = (*DefaultLogger)(nil)
