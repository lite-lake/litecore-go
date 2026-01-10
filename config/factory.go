package config

import (
	"errors"

	"com.litelake.litecore/config/common"
	"com.litelake.litecore/config/drivers"
)

func BuildConfig(driver string, filePath string) (common.ConfigProvider, error) {
	switch driver {
	case "yaml":
		return drivers.NewYamlConfigProvider(filePath)
	case "json":
		return drivers.NewJsonConfigProvider(filePath)
	default:
		return nil, errors.New("unsupported driver")
	}
}
