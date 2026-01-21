package common

import "github.com/lite-lake/litecore-go/util/logger"

// IBaseService 服务基类接口
// 所有 Service 类必须继承此接口并实现 GetServiceName 方法
// 系统通过此接口判断是否符合标准服务定义
type IBaseService interface {
	// ServiceName 返回当前服务实现的类名
	// 用于标识和调试服务实例
	ServiceName() string

	// OnStart 在服务器启动时触发
	OnStart() error
	// OnStop 在服务器停止时触发
	OnStop() error

	// Logger 获取日志实例
	Logger() logger.ILogger

	// SetLoggerManager 设置日志管理器
	SetLoggerManager(mgr logger.ILoggerManager)
}
