package drivers

import (
	"context"
	"fmt"

	"com.litelake.litecore/common"
)

// BaseManager 基础管理器
// 提供缓存管理器的公共实现
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

// Close 关闭管理器
func (m *BaseManager) Close() error {
	return nil
}

// Ensure BaseManager implements common.BaseManager interface
var _ common.BaseManager = (*BaseManager)(nil)

// ValidateContext 验证上下文是否有效
func ValidateContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	return nil
}

// ValidateKey 验证缓存键是否有效
func ValidateKey(key string) error {
	if key == "" {
		return fmt.Errorf("cache key cannot be empty")
	}
	return nil
}
