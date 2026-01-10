package common

// ConfigProvider 配置提供者
type ConfigProvider interface {
	// Get 获取配置项 （key 支持 aaa.bbb.ccc 路径查询)
	Get(key string) (any, error)
	// Has 检查配置项是否存在
	Has(key string) bool
}
