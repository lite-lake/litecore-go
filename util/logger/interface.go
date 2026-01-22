package logger

import (
	"context"
	"github.com/lite-lake/litecore-go/common"
)

// ILoggerManager 日志管理器接口
type ILoggerManager interface {
	// ========== 生命周期管理（符合 BaseManager 接口） ==========
	// ManagerName 返回管理器名称
	ManagerName() string

	// Health 检查管理器健康状态
	Health() error

	// OnStart 在服务器启动时触发
	OnStart() error

	// OnStop 在服务器停止时触发
	OnStop() error

	// ========== 日志管理 ==========
	// Logger 获取指定名称的 Logger 实例
	Logger(name string) common.ILogger

	// SetGlobalLevel 设置全局日志级别
	SetGlobalLevel(level common.LogLevel)

	// Shutdown 关闭日志管理器，刷新所有待处理的日志
	Shutdown(ctx context.Context) error
}
