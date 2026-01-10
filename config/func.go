package config

import (
	"fmt"

	"com.litelake.litecore/config/common"
)

func Get[T any](p common.ConfigProvider, key string) (T, error) {
	val, err := p.GetConfig(key)
	if err != nil {
		var zero T
		return zero, err
	}
	typed, ok := val.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("config value for %s is not of expected type", key)
	}
	return typed, nil
}
