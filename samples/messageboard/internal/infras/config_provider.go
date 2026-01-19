package infras

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/config"
)

// NewConfigProvider 创建配置提供者
func NewConfigProvider() (common.BaseConfigProvider, error) {
	return config.NewConfigProvider("yaml", "configs/config.yaml")
}
