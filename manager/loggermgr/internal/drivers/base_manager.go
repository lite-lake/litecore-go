package drivers

import (
	"context"
	"fmt"
	"sync"

	"com.litelake.litecore/common"
	"com.litelake.litecore/manager/loggermgr/internal/loglevel"
)

// BaseManager 基础管理器
// 提供管理器的公共实现
type BaseManager struct {
	name string
}

// NewBaseManager 创建基础管理器
func NewBaseManager(name string) *BaseManager {
	return &BaseManager{
		name: name,
	}
}

// ManagerName 返回管理器名称
func (m *BaseManager) ManagerName() string {
	return m.name
}

// Health 检查管理器健康状态
func (m *BaseManager) Health() error {
	return nil
}

// OnStart 在服务器启动时触发
func (m *BaseManager) OnStart() error {
	return nil
}

// OnStop 在服务器停止时触发
func (m *BaseManager) OnStop() error {
	return nil
}

// Shutdown 关闭管理器
func (m *BaseManager) Shutdown(ctx context.Context) error {
	return nil
}

// ensure BaseManager implements common.Manager interface
var _ common.Manager = (*BaseManager)(nil)

// ValidateContext 验证上下文是否有效
func ValidateContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	return nil
}

// BaseLogger 基础 Logger 实现
type BaseLogger struct {
	name  string
	level loglevel.LogLevel
	mu    sync.RWMutex
}

// NewBaseLogger 创建基础 Logger
func NewBaseLogger(name string, level loglevel.LogLevel) *BaseLogger {
	return &BaseLogger{
		name:  name,
		level: level,
	}
}

// LoggerName 返回 Logger 名称
func (l *BaseLogger) LoggerName() string {
	return l.name
}

// SetLevel 设置日志级别
func (l *BaseLogger) SetLevel(level loglevel.LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel 获取当前日志级别
func (l *BaseLogger) GetLevel() loglevel.LogLevel {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// IsEnabled 检查是否启用指定级别的日志
func (l *BaseLogger) IsEnabled(level loglevel.LogLevel) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return level >= l.level
}
