package configproviders

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/config"
)

func NewConfigProvider() (common.IBaseConfigProvider, error) {
	return config.NewConfigProvider("yaml", "configs/config.yaml")
}
