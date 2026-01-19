package common

// IBaseConfigProvider 配置提供者基础接口
type IBaseConfigProvider interface {
	// ConfigProviderName 返回当前配置提供者的类名
	ConfigProviderName() string
	// Get 获取配置项 （key 支持 aaa.bbb.ccc 路径查询)
	Get(key string) (any, error)
	// Has 检查配置项是否存在
	Has(key string) bool
}
