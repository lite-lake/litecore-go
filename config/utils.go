package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lite-lake/litecore-go/common"
	"github.com/duke-git/lancet/v2/convertor"
)

// 定义特定错误
var (
	ErrKeyNotFound  = errors.New("config key not found")
	ErrTypeMismatch = errors.New("type mismatch")
	ErrInvalidValue = errors.New("invalid value")
)

// IsConfigKeyNotFound 判断是否为键不存在错误
func IsConfigKeyNotFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrKeyNotFound) || strings.Contains(err.Error(), "not found")
}

// Get 获取配置项并进行类型转换
// 支持数字类型的智能转换：JSON 中的 float64 可转换为 int/int64（如果是整数）
func Get[T any](p common.IBaseConfigProvider, key string) (T, error) {
	var zero T // 零值

	val, err := p.Get(key)
	if err != nil {
		if IsConfigKeyNotFound(err) {
			return zero, fmt.Errorf("%w: config key '%s'", ErrKeyNotFound, key)
		}
		return zero, err
	}

	// 尝试直接类型断言
	if typed, ok := val.(T); ok {
		return typed, nil
	}

	// 使用 lancet 进行智能类型转换
	converted, err := convertType[T](val)
	if err != nil {
		actualType := fmt.Sprintf("%T", val)
		expectedType := fmt.Sprintf("%T", zero)
		return zero, fmt.Errorf("%w: config key '%s' - expected %s, got %s", ErrTypeMismatch, key, expectedType, actualType)
	}

	return converted, nil
}

// GetWithDefault 获取配置项，如果不存在则返回默认值
func GetWithDefault[T any](p common.IBaseConfigProvider, key string, defaultValue T) T {
	val, err := Get[T](p, key)
	if err != nil {
		return defaultValue
	}
	return val
}

// convertType 使用 lancet 进行类型转换
func convertType[T any](val any) (T, error) {
	var zero T

	switch any(zero).(type) {
	case int:
		i, err := convertor.ToInt(val)
		if err != nil {
			return zero, err
		}
		return any(int(i)).(T), nil
	case int64:
		i, err := convertor.ToInt(val)
		if err != nil {
			return zero, err
		}
		return any(i).(T), nil
	case int32:
		i, err := convertor.ToInt(val)
		if err != nil {
			return zero, err
		}
		return any(int32(i)).(T), nil
	case float64:
		f, err := convertor.ToFloat(val)
		if err != nil {
			return zero, err
		}
		return any(f).(T), nil
	case string:
		return any(convertor.ToString(val)).(T), nil
	case bool:
		if s, ok := val.(string); ok {
			b, err := convertor.ToBool(s)
			if err != nil {
				return zero, err
			}
			return any(b).(T), nil
		}
		// 直接类型断言
		if b, ok := val.(bool); ok {
			return any(b).(T), nil
		}
		return zero, fmt.Errorf("cannot convert %T to bool", val)
	default:
		return zero, fmt.Errorf("unsupported type conversion to %T", zero)
	}
}
