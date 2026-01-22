package configmgr

import "github.com/lite-lake/litecore-go/common"

// IConfigManager 配置管理器基础接口
type IConfigManager interface {
	common.IBaseManager

	// Get 获取配置项 （key 支持 aaa.bbb.ccc 路径查询)
	Get(key string) (any, error)
	// Has 检查配置项是否存在
	Has(key string) bool
}

type IConfigLoadHandler func() (map[string]any, error)
