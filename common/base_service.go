package common

// BaseService 服务基类接口
// 所有 Service 类必须继承此接口并实现 GetServiceName 方法
// 系统通过此接口判断是否符合标准服务定义
type BaseService interface {
	// GetServiceName 返回当前服务实现的类名
	// 用于标识和调试服务实例
	GetServiceName() string
}
