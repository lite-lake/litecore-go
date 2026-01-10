package common

// ConfigProvider 配置提供者
type ConfigProvider interface {
	// GetConfig 获取配置项 （key 支持 aaa.bbb.ccc 路径查询)
	GetConfig(key string) (any, error)
}
