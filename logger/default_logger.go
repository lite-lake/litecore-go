package logger

import (
	"log"
)

type DefaultLogger struct {
	prefix string
	level  LogLevel
}

func NewDefaultLogger(name string) *DefaultLogger {
	return &DefaultLogger{
		prefix: "[" + name + "] ",
	}
}

func (l *DefaultLogger) Debug(msg string, args ...any) {
	if l.level >= DebugLevel {
		return
	}
	log.Printf(l.prefix+"DEBUG: %s %v", msg, args)
}

func (l *DefaultLogger) Info(msg string, args ...any) {
	if l.level >= InfoLevel {
		return
	}
	log.Printf(l.prefix+"INFO: %s %v", msg, args)
}

func (l *DefaultLogger) Warn(msg string, args ...any) {
	if l.level >= WarnLevel {
		return
	}
	log.Printf(l.prefix+"WARN: %s %v", msg, args)
}

func (l *DefaultLogger) Error(msg string, args ...any) {
	if l.level >= ErrorLevel {
		return
	}
	log.Printf(l.prefix+"ERROR: %s %v", msg, args)
}

func (l *DefaultLogger) Fatal(msg string, args ...any) {
	if l.level >= FatalLevel {
		return
	}
	log.Printf(l.prefix+"FATAL: %s %v", msg, args)
	args = append([]any{l.prefix}, args...)
	log.Fatal(args...)
}

func (l *DefaultLogger) With(args ...any) ILogger {
	return l
}

func (l *DefaultLogger) SetLevel(level LogLevel) {
}

var _ ILogger = (*DefaultLogger)(nil)
