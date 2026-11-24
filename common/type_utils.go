package common

import (
	"fmt"
)

// GetString 从 any 类型中安全获取字符串值
func GetString(value any) (string, error) {
	if value == nil {
		return "", fmt.Errorf("value is nil")
	}
	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("expected string, got %T", value)
	}
	return str, nil
}

// GetStringOrDefault 从 any 类型中安全获取字符串值，失败时返回默认值
func GetStringOrDefault(value any, defaultValue string) string {
	str, err := GetString(value)
	if err != nil {
		return defaultValue
	}
	return str
}

// GetMap 从 any 类型中安全获取 map[string]any 值
func GetMap(value any) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("value is nil")
	}
	m, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected map[string]any, got %T", value)
	}
	return m, nil
}

// GetMapOrDefault 从 any 类型中安全获取 map[string]any 值，失败时返回默认值
func GetMapOrDefault(value any, defaultValue map[string]any) map[string]any {
	m, err := GetMap(value)
	if err != nil {
		return defaultValue
	}
	return m
}
