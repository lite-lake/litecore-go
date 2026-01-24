package server

import (
	"sync"

	"github.com/lite-lake/litecore-go/logger"
)

// StartupEvent 启动日志事件
type StartupEvent struct {
	Phase   StartupPhase
	Message string
	Fields  []logger.Field
}

// AsyncStartupLogger 异步启动日志记录器
type AsyncStartupLogger struct {
	logger   logger.ILogger
	buffer   chan *StartupEvent
	wg       sync.WaitGroup
	closeCh  chan struct{}
	isClosed bool
	mu       sync.Mutex
}

// NewAsyncStartupLogger 创建异步启动日志记录器
func NewAsyncStartupLogger(baseLogger logger.ILogger, bufferSize int) *AsyncStartupLogger {
	l := &AsyncStartupLogger{
		logger:  baseLogger,
		buffer:  make(chan *StartupEvent, bufferSize),
		closeCh: make(chan struct{}),
	}
	l.wg.Add(1)
	go l.flushLoop()
	return l
}

// Log 记录启动日志
func (l *AsyncStartupLogger) Log(phase StartupPhase, msg string, fields ...logger.Field) {
	l.mu.Lock()
	if l.isClosed {
		l.mu.Unlock()
		return
	}
	l.mu.Unlock()

	event := &StartupEvent{
		Phase:   phase,
		Message: msg,
		Fields:  fields,
	}

	select {
	case l.buffer <- event:
	default:
		l.logger.Warn("Startup log buffer is full, discarding log", "msg", msg, "phase", phase.String())
	}
}

// flushLoop 日志刷新循环（后台 goroutine）
func (l *AsyncStartupLogger) flushLoop() {
	defer l.wg.Done()
	for {
		select {
		case event := <-l.buffer:
			l.logger.Info(event.Message, event.Fields...)
		case <-l.closeCh:
			for len(l.buffer) > 0 {
				event := <-l.buffer
				l.logger.Info(event.Message, event.Fields...)
			}
			return
		}
	}
}

// Stop 停止日志记录器
func (l *AsyncStartupLogger) Stop() {
	l.mu.Lock()
	if !l.isClosed {
		l.isClosed = true
		close(l.closeCh)
	}
	l.mu.Unlock()
	l.wg.Wait()
}
