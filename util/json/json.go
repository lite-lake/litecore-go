package json

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ILiteUtilJSON JSON 工具接口
type ILiteUtilJSON interface {
	// 基础验证和格式化
	IsValid(jsonStr string) bool
	PrettyPrint(jsonStr string, indent string) (string, error)
	PrettyPrintWithIndent(jsonStr string) (string, error)
	Compact(jsonStr string) (string, error)
	Escape(str string) string
	Unescape(str string) (string, error)

	// 数据转换
	ToMap(jsonStr string) (map[string]any, error)
	ToMapStrict(jsonStr string) (map[string]any, error)
	ToStruct(jsonStr string, target any) error
	FromMap(data map[string]any) (string, error)
	FromStruct(data any) (string, error)
	FromStructWithIndent(data any, indent string) (string, error)

	// 路径操作
	GetValue(jsonStr string, path string) (any, error)
	GetString(jsonStr string, path string) (string, error)
	GetFloat64(jsonStr string, path string) (float64, error)
	GetBool(jsonStr string, path string) (bool, error)

	// 高级操作
	Merge(jsonStr1 string, jsonStr2 string) (string, error)
	Diff(jsonStr1 string, jsonStr2 string) (bool, error)

	// 实用工具
	IsObject(jsonStr string) bool
	IsArray(jsonStr string) bool
	GetType(jsonStr string, path string) (string, error)
	GetKeys(jsonStr string, path string) ([]string, error)
	GetSize(jsonStr string, path string) (int, error)
	Contains(jsonStr string, path string, key string) (bool, error)
}

// =========================================

// jsonEngine JSON 操作工具类
type jsonEngine struct{}

// 默认 JSON 操作实例（单例模式）
var (
	JSON ILiteUtilJSON
	// removed
)

// Default 返回默认的 JSON 操作实例（单例模式）
// Deprecated: 请使用 liteutil.LiteUtil().Json() 来获取 JSON 工具实例
func Default() ILiteUtilJSON {
	return JSON
}

// New 创建新的 JSON 操作工具实例
// Deprecated: 请使用 liteutil.LiteUtil().NewJsonOperation() 来创建新的 JSON 工具实例
func newJSONEngine() ILiteUtilJSON {
	return &jsonEngine{}
}

// =========================================
// 基础验证和格式化操作
// =========================================

// IsValid 验证 JSON 字符串是否有效
func (j *jsonEngine) IsValid(jsonStr string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(jsonStr), &js) == nil
}

// PrettyPrint 格式化 JSON 字符串，增加缩进
func (j *jsonEngine) PrettyPrint(jsonStr string, indent string) (string, error) {
	var result any
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	formatted, err := json.MarshalIndent(result, "", indent)
	if err != nil {
		return "", fmt.Errorf("format JSON failed: %w", err)
	}

	return string(formatted), nil
}

// PrettyPrintWithIndent 使用默认缩进（2个空格）格式化 JSON
func (j *jsonEngine) PrettyPrintWithIndent(jsonStr string) (string, error) {
	return j.PrettyPrint(jsonStr, "  ")
}

// Compact 压缩 JSON 字符串，移除所有空白字符
func (j *jsonEngine) Compact(jsonStr string) (string, error) {
	var result any
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	compacted, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("compact JSON failed: %w", err)
	}

	return string(compacted), nil
}

// Escape 转义 JSON 字符串中的特殊字符
func (j *jsonEngine) Escape(str string) string {
	result, _ := json.Marshal(str)
	return string(result[1 : len(result)-1]) // 移除外层引号
}

// Unescape 反转义 JSON 字符串
func (j *jsonEngine) Unescape(str string) (string, error) {
	var result string
	if err := json.Unmarshal([]byte("\""+str+"\""), &result); err != nil {
		return "", fmt.Errorf("unescape JSON failed: %w", err)
	}
	return result, nil
}

// =========================================
// 数据转换操作
// =========================================

// ToMap 将 JSON 字符串转换为 map[string]any
func (j *jsonEngine) ToMap(jsonStr string) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("convert JSON to map failed: %w", err)
	}
	return result, nil
}

// ToMapStrict 严格模式转换，要求 JSON 必须是对象类型
func (j *jsonEngine) ToMapStrict(jsonStr string) (map[string]any, error) {
	result, err := j.ToMap(jsonStr)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("JSON is not an object")
	}
	return result, nil
}

// ToStruct 传统方法，将 JSON 字符串转换为结构体
func (j *jsonEngine) ToStruct(jsonStr string, target any) error {
	if err := json.Unmarshal([]byte(jsonStr), target); err != nil {
		return fmt.Errorf("convert JSON to struct failed: %w", err)
	}
	return nil
}

// FromMap 将 map[string]any 转换为 JSON 字符串
func (j *jsonEngine) FromMap(data map[string]any) (string, error) {
	result, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("convert map to JSON failed: %w", err)
	}
	return string(result), nil
}

// FromStruct 将结构体转换为 JSON 字符串
func (j *jsonEngine) FromStruct(data any) (string, error) {
	result, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("convert struct to JSON failed: %w", err)
	}
	return string(result), nil
}

// FromStructWithIndent 使用缩进格式化结构体为 JSON 字符串
func (j *jsonEngine) FromStructWithIndent(data any, indent string) (string, error) {
	result, err := json.MarshalIndent(data, "", indent)
	if err != nil {
		return "", fmt.Errorf("convert struct to indented JSON failed: %w", err)
	}
	return string(result), nil
}

// =========================================
// 路径操作
// =========================================

// GetValue 根据 JSONPath 获取值（支持简单的 . 语法）
func (j *jsonEngine) GetValue(jsonStr string, path string) (any, error) {
	var data any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if path == "" || path == "." {
		return data, nil
	}

	keys := strings.Split(path, ".")
	current := data

	for _, key := range keys {
		if key == "" {
			continue
		}

		switch v := current.(type) {
		case map[string]any:
			if val, exists := v[key]; exists {
				current = val
			} else {
				return nil, fmt.Errorf("path '%s' not found", path)
			}
		case []any:
			index, err := strconv.Atoi(key)
			if err != nil {
				return nil, fmt.Errorf("invalid array index '%s' in path '%s'", key, path)
			}
			if index < 0 || index >= len(v) {
				return nil, fmt.Errorf("array index '%d' out of bounds in path '%s'", index, path)
			}
			current = v[index]
		default:
			return nil, fmt.Errorf("cannot access key '%s' in non-object/non-array value", key)
		}
	}

	return current, nil
}

// GetString 根据 JSONPath 获取字符串值
func (j *jsonEngine) GetString(jsonStr string, path string) (string, error) {
	val, err := j.GetValue(jsonStr, path)
	if err != nil {
		return "", err
	}

	if str, ok := val.(string); ok {
		return str, nil
	}

	// 尝试转换为字符串
	if val == nil {
		return "", fmt.Errorf("value at path '%s' is null", path)
	}

	return fmt.Sprintf("%v", val), nil
}

// GetFloat64 根据 JSONPath 获取浮点数值
func (j *jsonEngine) GetFloat64(jsonStr string, path string) (float64, error) {
	val, err := j.GetValue(jsonStr, path)
	if err != nil {
		return 0, err
	}

	switch v := val.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("value at path '%s' is not a number", path)
	}
}

// GetBool 根据 JSONPath 获取布尔值
func (j *jsonEngine) GetBool(jsonStr string, path string) (bool, error) {
	val, err := j.GetValue(jsonStr, path)
	if err != nil {
		return false, err
	}

	if b, ok := val.(bool); ok {
		return b, nil
	}

	if str, ok := val.(string); ok {
		return strconv.ParseBool(str)
	}

	return false, fmt.Errorf("value at path '%s' is not a boolean", path)
}

// =========================================
// 高级操作
// =========================================

// Merge 合并两个 JSON 对象字符串
func (j *jsonEngine) Merge(jsonStr1 string, jsonStr2 string) (string, error) {
	obj1, err := j.ToMap(jsonStr1)
	if err != nil {
		return "", fmt.Errorf("parse first JSON failed: %w", err)
	}

	obj2, err := j.ToMap(jsonStr2)
	if err != nil {
		return "", fmt.Errorf("parse second JSON failed: %w", err)
	}

	merged := j.mergeMaps(obj1, obj2)

	result, err := json.Marshal(merged)
	if err != nil {
		return "", fmt.Errorf("marshal merged JSON failed: %w", err)
	}

	return string(result), nil
}

// mergeMaps 递归合并两个 map
func (j *jsonEngine) mergeMaps(map1 map[string]any, map2 map[string]any) map[string]any {
	result := make(map[string]any)

	// 复制第一个 map 的所有键
	for k, v := range map1 {
		result[k] = v
	}

	// 合并第二个 map 的键
	for k, v := range map2 {
		if val1, exists := result[k]; exists {
			// 如果两个值都是 map，递归合并
			if map1Val, ok1 := val1.(map[string]any); ok1 {
				if map2Val, ok2 := v.(map[string]any); ok2 {
					result[k] = j.mergeMaps(map1Val, map2Val)
					continue
				}
			}
		}
		result[k] = v
	}

	return result
}

// Diff 比较两个 JSON 字符串的差异
func (j *jsonEngine) Diff(jsonStr1 string, jsonStr2 string) (bool, error) {
	var obj1, obj2 any

	if err := json.Unmarshal([]byte(jsonStr1), &obj1); err != nil {
		return false, fmt.Errorf("parse first JSON failed: %w", err)
	}

	if err := json.Unmarshal([]byte(jsonStr2), &obj2); err != nil {
		return false, fmt.Errorf("parse second JSON failed: %w", err)
	}

	return !reflect.DeepEqual(obj1, obj2), nil
}

// =========================================
// 实用工具函数
// =========================================

// IsObject 检查 JSON 字符串是否表示对象
func (j *jsonEngine) IsObject(jsonStr string) bool {
	trimmed := strings.TrimSpace(jsonStr)
	return strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")
}

// IsArray 检查 JSON 字符串是否表示数组
func (j *jsonEngine) IsArray(jsonStr string) bool {
	trimmed := strings.TrimSpace(jsonStr)
	return strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")
}

// GetType 获取 JSON 值的类型
func (j *jsonEngine) GetType(jsonStr string, path string) (string, error) {
	val, err := j.GetValue(jsonStr, path)
	if err != nil {
		return "", err
	}

	switch val.(type) {
	case nil:
		return "null", nil
	case bool:
		return "boolean", nil
	case float64:
		return "number", nil
	case string:
		return "string", nil
	case []any:
		return "array", nil
	case map[string]any:
		return "object", nil
	default:
		return "unknown", nil
	}
}

// GetKeys 获取 JSON 对象的所有键
func (j *jsonEngine) GetKeys(jsonStr string, path string) ([]string, error) {
	val, err := j.GetValue(jsonStr, path)
	if err != nil {
		return nil, err
	}

	obj, ok := val.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("value at path '%s' is not an object", path)
	}

	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}

	return keys, nil
}

// GetSize 获取 JSON 数组或对象的长度
func (j *jsonEngine) GetSize(jsonStr string, path string) (int, error) {
	val, err := j.GetValue(jsonStr, path)
	if err != nil {
		return 0, err
	}

	switch v := val.(type) {
	case []any:
		return len(v), nil
	case map[string]any:
		return len(v), nil
	default:
		return 0, fmt.Errorf("value at path '%s' is not an array or object", path)
	}
}

// Contains 检查 JSON 对象是否包含指定键
func (j *jsonEngine) Contains(jsonStr string, path string, key string) (bool, error) {
	keys, err := j.GetKeys(jsonStr, path)
	if err != nil {
		return false, err
	}

	for _, k := range keys {
		if k == key {
			return true, nil
		}
	}

	return false, nil
}
