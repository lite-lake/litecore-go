package configproviders

import (
	"com.litelake.litecore/common"
	"com.litelake.litecore/config"
)

func NewConfigProvider() (common.IBaseConfigProvider, error) {
	return config.NewConfigProvider("yaml", "configs/config.yaml")
}
