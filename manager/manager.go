package manager

// ManagerFactory 管理器工厂接口定义
type ManagerFactory interface {
	// Build 创建管理器实例
	Build(driver string, config ConfigMap) Manager
}

// ConfigMap 配置数据类型
type ConfigMap map[string]any

// Manager 管理器接口定义
type Manager interface {
	// ManagerName 返回管理器名称
	ManagerName() string
	// Health 检查管理器健康状态
	Health() error
	// OnStart 在服务器启动时触发
	OnStart() error
	// OnStop 在服务器停止时触发
	OnStop() error
}
