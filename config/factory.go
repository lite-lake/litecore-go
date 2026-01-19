package config

import (
	"errors"

	"com.litelake.litecore/common"
)

func NewConfigProvider(driver string, filePath string) (common.IBaseConfigProvider, error) {
	switch driver {
	case "yaml":
		return NewYamlConfigProvider(filePath)
	case "json":
		return NewJsonConfigProvider(filePath)
	default:
		return nil, errors.New("unsupported driver")
	}
}
