package container

import (
	"reflect"
	"strings"
)

// DependencyResolver 依赖解析器接口
// 各容器通过实现此接口提供自己的依赖解析逻辑
type DependencyResolver interface {
	// ResolveDependency 解析字段类型对应的依赖实例
	ResolveDependency(fieldType reflect.Type) (interface{}, error)
}

// injectDependencies 向实例注入依赖
// 使用反射解析实例字段，根据 inject 标签查找并注入依赖
func injectDependencies(instance interface{}, resolver DependencyResolver) error {
	val := reflect.ValueOf(instance)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 如果不是结构体，无法注入
	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// 检查是否有 inject 标签
		// 注意：`inject:""` 会被 Tag.Get 返回为 ""，但我们需要区分"没有标签"和"空值标签"
		tagValue, ok := field.Tag.Lookup("inject")
		if !ok {
			// 没有 inject 标签，跳过
			continue
		}

		// 检查是否为可选依赖
		if tagValue == "optional" {
			// 可选依赖，如果找不到也不报错
			if dep, err := resolver.ResolveDependency(field.Type); err == nil && dep != nil {
				if fieldVal.CanSet() {
					fieldVal.Set(reflect.ValueOf(dep))
				}
			}
			continue
		}

		// 必需依赖（包括 tagValue == "" 的情况）
		dependency, err := resolver.ResolveDependency(field.Type)
		if err != nil {
			return err
		}

		if fieldVal.CanSet() {
			fieldVal.Set(reflect.ValueOf(dependency))
		}
	}

	return nil
}

// typeMatches 检查 itemType 是否匹配 targetType
// 支持：
// 1. 精确类型匹配
// 2. 接口实现检查（targetType 是接口）
// 3. 指针类型的元素匹配
func typeMatches(itemType, targetType reflect.Type) bool {
	// 精确匹配
	if itemType == targetType {
		return true
	}

	// 如果 targetType 是接口类型，检查 item 是否实现该接口
	if targetType.Kind() == reflect.Interface && itemType.Implements(targetType) {
		return true
	}

	// 如果 item 是指针类型，检查其元素类型
	if itemType.Kind() == reflect.Ptr {
		elemType := itemType.Elem()
		// 检查元素类型是否精确匹配
		if elemType == targetType {
			return true
		}
		// 检查元素类型是否实现接口（如果 targetType 是接口）
		if targetType.Kind() == reflect.Interface && elemType.Implements(targetType) {
			return true
		}
	}

	return false
}

// extractNameFromType 从类型名称中提取简单名称
// 例如：*UserServiceImpl -> UserServiceImpl
func extractNameFromType(typ reflect.Type) string {
	switch typ.Kind() {
	case reflect.Ptr:
		return extractNameFromType(typ.Elem())
	case reflect.Interface:
		return typ.Name()
	case reflect.Struct:
		return typ.Name()
	default:
		return typ.String()
	}
}

// toLowerCamelCase 将字符串转换为小驼峰命名
// 例如：UserService -> userService
func toLowerCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
