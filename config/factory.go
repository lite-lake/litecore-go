package config

import (
	"errors"

	"github.com/lite-lake/litecore-go/common"
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
