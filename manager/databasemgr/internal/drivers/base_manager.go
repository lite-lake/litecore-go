package drivers

import (
	"context"
	"fmt"

	"com.litelake.litecore/common"
)

// BaseManager 基础数据库管理器
// 提供 common.Manager 接口的公共实现
type BaseManager struct {
	name string
}

// NewBaseManager 创建基础数据库管理器
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

// 编译时检查：确保 BaseManager 实现了 common.Manager 接口
var _ common.Manager = (*BaseManager)(nil)

// ValidateContext 验证上下文是否有效
func ValidateContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	return nil
}