package loggermgr

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap/zapcore"
)

// loggerManagerBaseImpl 提供基础实现
type loggerManagerBaseImpl struct {
	name         string
	globalLevel  zapcore.Level
	mu           sync.RWMutex
	shutdownOnce sync.Once
}

// newLoggerManagerBaseImpl 创建基类
func newLoggerManagerBaseImpl(name string) *loggerManagerBaseImpl {
	return &loggerManagerBaseImpl{
		name:        name,
		globalLevel: zapcore.InfoLevel,
	}
}

// ManagerName 返回管理器名称
func (b *loggerManagerBaseImpl) ManagerName() string {
	return b.name
}

// Health 检查管理器健康状态
func (b *loggerManagerBaseImpl) Health() error {
	return nil
}

// OnStart 在服务器启动时触发
func (b *loggerManagerBaseImpl) OnStart() error {
	return nil
}

// OnStop 在服务器停止时触发
func (b *loggerManagerBaseImpl) OnStop() error {
	return nil
}

// SetGlobalLevel 设置全局日志级别
func (b *loggerManagerBaseImpl) setGlobalLevel(level zapcore.Level) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.globalLevel = level
}

// getGlobalLevel 获取全局日志级别
func (b *loggerManagerBaseImpl) getGlobalLevel() zapcore.Level {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.globalLevel
}

// ValidateContext 验证上下文是否有效
func ValidateContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	return nil
}

// CustomTimeEncoder 自定义时间编码器
func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// clampSize 限制大小在安全范围内（导出供测试使用）
func ClampSize(size int) int {
	return clampSize(size)
}

// clampAge 限制时间在安全范围内（导出供测试使用）
func ClampAge(age int) int {
	return clampAge(age)
}

// clampBackups 限制备份数在安全范围内（导出供测试使用）
func ClampBackups(backups int) int {
	return clampBackups(backups)
}
