package container

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/lite-lake/litecore-go/common"
)

// UninjectedFieldError 未注入字段错误
type UninjectedFieldError struct {
	InstanceName string
	FieldName    string
	FieldType    reflect.Type
}

func (e *UninjectedFieldError) Error() string {
	return fmt.Sprintf("field %s.%s (type %s) marked with inject:\"\" is still nil after injection",
		e.InstanceName, e.FieldName, e.FieldType)
}

// verifyInjectTags 验证所有 inject:"" 标签的字段是否已被注入
func verifyInjectTags(instance interface{}) {
	val := reflect.ValueOf(instance)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return
	}

	typ := val.Type()
	instanceName := extractNameFromType(typ)

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		tagValue, ok := field.Tag.Lookup("inject")
		if !ok {
			continue
		}

		if tagValue == "optional" {
			continue
		}

		if !fieldVal.CanInterface() || fieldVal.IsZero() || fieldVal.IsNil() {
			panic(&UninjectedFieldError{
				InstanceName: instanceName,
				FieldName:    field.Name,
				FieldType:    field.Type,
			})
		}
	}
}

// IDependencyResolver 依赖解析器接口
// 各容器通过实现此接口提供自己的依赖解析逻辑
type IDependencyResolver interface {
	// ResolveDependency 解析字段类型对应的依赖实例
	ResolveDependency(fieldType reflect.Type, structType reflect.Type, fieldName string) (interface{}, error)
}

// ContainerSource 容器依赖源接口
// 容器实现此接口后，可以通过通用依赖解析器解析依赖
type ContainerSource interface {
	// GetDependency 根据类型获取依赖实例
	// 返回nil表示类型不匹配，返回error表示类型匹配但查找失败
	GetDependency(fieldType reflect.Type) (interface{}, error)
}

// injectDependencies 向实例注入依赖
// 使用反射解析实例字段，根据 inject 标签查找并注入依赖
func injectDependencies(instance interface{}, resolver IDependencyResolver) error {
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
		// 注意：`inject:""` 会被 Tag.Ins 返回为 ""，但我们需要区分"没有标签"和"空值标签"
		tagValue, ok := field.Tag.Lookup("inject")
		if !ok {
			// 没有 inject 标签，跳过
			continue
		}

		// 检查是否为可选依赖
		if tagValue == "optional" {
			// 可选依赖，如果找不到也不报错
			if dep, err := resolver.ResolveDependency(field.Type, typ, field.Name); err == nil && dep != nil {
				if fieldVal.CanSet() {
					fieldVal.Set(reflect.ValueOf(dep))
				}
			}
			continue
		}

		// 必需依赖（包括 tagValue == "" 的情况）
		dependency, err := resolver.ResolveDependency(field.Type, typ, field.Name)
		if err != nil {
			return err
		}

		if fieldVal.CanSet() {
			fieldVal.Set(reflect.ValueOf(dependency))
		} else {
			fieldPtr := unsafe.Pointer(fieldVal.UnsafeAddr())
			reflect.NewAt(field.Type, fieldPtr).Elem().Set(reflect.ValueOf(dependency))
		}
	}

	return nil
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

// GenericDependencyResolver 通用依赖解析器
// 支持按优先级顺序从多个容器源解析依赖
type GenericDependencyResolver struct {
	sources []ContainerSource
}

// NewGenericDependencyResolver 创建通用依赖解析器
func NewGenericDependencyResolver(
	sources ...ContainerSource,
) *GenericDependencyResolver {
	return &GenericDependencyResolver{
		sources: sources,
	}
}

// ResolveDependency 解析字段类型对应的依赖实例
// 按照sources的顺序依次尝试解析，找到第一个匹配的依赖
func (r *GenericDependencyResolver) ResolveDependency(fieldType reflect.Type, structType reflect.Type, fieldName string) (interface{}, error) {

	for _, source := range r.sources {
		dep, err := source.GetDependency(fieldType)
		if dep != nil {
			return dep, nil
		}
		if err != nil {
			return nil, err
		}
	}

	return nil, &DependencyNotFoundError{
		FieldType:     fieldType,
		ContainerType: "Unknown",
	}
}

// extractLoggerName 从结构体类型推断 logger 名称
func (r *GenericDependencyResolver) extractLoggerName(structType reflect.Type) string {
	name := extractNameFromType(structType)
	return name
}

// resolveDependencyFromManager 从管理器容器解析依赖
func resolveDependencyFromManager(
	fieldType reflect.Type,
	managerContainer *ManagerContainer,
) (interface{}, error) {
	if managerContainer == nil {
		return nil, nil
	}

	baseManagerType := reflect.TypeOf((*common.IBaseManager)(nil)).Elem()
	if fieldType == baseManagerType || fieldType.Implements(baseManagerType) {
		impl := managerContainer.GetByType(fieldType)
		if impl == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Manager",
			}
		}
		return impl, nil
	}
	return nil, nil
}

// resolveDependencyFromEntity 从实体容器解析依赖
func resolveDependencyFromEntity(
	fieldType reflect.Type,
	entityContainer *EntityContainer,
) (interface{}, error) {
	if entityContainer == nil {
		return nil, nil
	}

	baseEntityType := reflect.TypeOf((*common.IBaseEntity)(nil)).Elem()
	if fieldType == baseEntityType || fieldType.Implements(baseEntityType) {
		items, err := entityContainer.GetByType(fieldType)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Entity",
			}
		}
		if len(items) > 1 {
			var names []string
			for _, item := range items {
				names = append(names, item.EntityName())
			}
			return nil, &AmbiguousMatchError{
				FieldType:  fieldType,
				Candidates: names,
			}
		}
		return items[0], nil
	}
	return nil, nil
}
