package logger

import (
	"context"
	"github.com/lite-lake/litecore-go/common"
	"sync"
)

type LoggerRegistry struct {
	mu       sync.RWMutex
	handlers map[string]*LoggerBridge
	mgr      ILoggerManager
}

func NewLoggerRegistry() *LoggerRegistry {
	return &LoggerRegistry{
		handlers: make(map[string]*LoggerBridge),
	}
}

func (r *LoggerRegistry) GetLogger(name string) common.ILogger {
	r.mu.Lock()
	defer r.mu.Unlock()

	if h, ok := r.handlers[name]; ok {
		return h
	}

	h := NewLoggerBridge(name)
	r.handlers[name] = h

	if r.mgr != nil {
		h.SetLoggerManager(r.mgr)
	}

	return h
}

func (r *LoggerRegistry) SetLoggerManager(mgr ILoggerManager) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.mgr = mgr

	for _, h := range r.handlers {
		h.SetLoggerManager(mgr)
	}
}

func (r *LoggerRegistry) ManagerName() string {
	return "LoggerRegistry"
}

func (r *LoggerRegistry) Health() error {
	if r.mgr != nil {
		return r.mgr.Health()
	}
	return nil
}

func (r *LoggerRegistry) OnStart() error {
	if r.mgr != nil {
		return r.mgr.OnStart()
	}
	return nil
}

func (r *LoggerRegistry) OnStop() error {
	if r.mgr != nil {
		return r.mgr.OnStop()
	}
	return nil
}

func (r *LoggerRegistry) Logger(name string) common.ILogger {
	return r.GetLogger(name)
}

func (r *LoggerRegistry) SetGlobalLevel(level common.LogLevel) {
	if r.mgr != nil {
		r.mgr.SetGlobalLevel(level)
	}
}

func (r *LoggerRegistry) Shutdown(ctx context.Context) error {
	if r.mgr != nil {
		return r.mgr.Shutdown(ctx)
	}
	return nil
}
