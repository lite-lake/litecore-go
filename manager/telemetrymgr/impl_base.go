package telemetrymgr

import (
	"context"
	"fmt"
)

// telemetryManagerBaseImpl 提供基础实现和工具函数
type telemetryManagerBaseImpl struct {
	name string
}

// newTelemetryManagerBaseImpl 创建基类
func newTelemetryManagerBaseImpl(name string) *telemetryManagerBaseImpl {
	return &telemetryManagerBaseImpl{
		name: name,
	}
}

// ManagerName 返回管理器名称
func (b *telemetryManagerBaseImpl) ManagerName() string {
	return b.name
}

// Health 检查管理器健康状态（默认实现）
func (b *telemetryManagerBaseImpl) Health() error {
	return nil
}

// OnStart 在服务器启动时触发（默认实现）
func (b *telemetryManagerBaseImpl) OnStart() error {
	return nil
}

// OnStop 在服务器停止时触发（默认实现）
func (b *telemetryManagerBaseImpl) OnStop() error {
	return nil
}

// ValidateContext 验证上下文是否有效
func ValidateContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	return nil
}
