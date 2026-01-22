package logger

import (
	"github.com/lite-lake/litecore-go/common"
	"log"
)

type defaultLogger struct {
	prefix string
}

func newDefaultLogger(name string) *defaultLogger {
	return &defaultLogger{
		prefix: "[" + name + "] ",
	}
}

func (l *defaultLogger) Debug(msg string, args ...any) {
	log.Printf(l.prefix+"DEBUG: %s %v", msg, args)
}

func (l *defaultLogger) Info(msg string, args ...any) {
	log.Printf(l.prefix+"INFO: %s %v", msg, args)
}

func (l *defaultLogger) Warn(msg string, args ...any) {
	log.Printf(l.prefix+"WARN: %s %v", msg, args)
}

func (l *defaultLogger) Error(msg string, args ...any) {
	log.Printf(l.prefix+"ERROR: %s %v", msg, args)
}

func (l *defaultLogger) Fatal(msg string, args ...any) {
	log.Printf(l.prefix+"FATAL: %s %v", msg, args)
	args = append([]any{l.prefix}, args...)
	log.Fatal(args...)
}

func (l *defaultLogger) With(args ...any) common.ILogger {
	return l
}

func (l *defaultLogger) SetLevel(level common.LogLevel) {
}
