package container

import (
	"reflect"

	"github.com/lite-lake/litecore-go/common"
)

// EntityContainer 实体层容器
// Entity 层无依赖，无 InjectAll 操作
type EntityContainer struct {
	container *NamedContainer[common.IBaseEntity]
}

// NewEntityContainer 创建新的实体容器
func NewEntityContainer() *EntityContainer {
	return &EntityContainer{
		container: NewNamedContainer(func(entity common.IBaseEntity) string {
			return entity.EntityName()
		}),
	}
}

// RegisterEntity 泛型注册函数，注册实体实例
func RegisterEntity[T common.IBaseEntity](e *EntityContainer, impl T) error {
	return e.Register(impl)
}

// Register 注册实体实例
func (e *EntityContainer) Register(ins common.IBaseEntity) error {
	return e.container.Register(ins)
}

// GetAll 获取所有已注册的实体
func (e *EntityContainer) GetAll() []common.IBaseEntity {
	return e.container.GetAll()
}

// GetByName 根据名称获取实体
func (e *EntityContainer) GetByName(name string) (common.IBaseEntity, error) {
	return e.container.GetByName(name)
}

// GetByType 根据类型获取实体
// 返回所有实现了该类型的实体列表
func (e *EntityContainer) GetByType(typ reflect.Type) ([]common.IBaseEntity, error) {
	entities := e.GetAll()
	var result []common.IBaseEntity

	for _, item := range entities {
		itemType := reflect.TypeOf(item)

		if itemType == typ {
			result = append(result, item)
			continue
		}

		if typ.Kind() == reflect.Interface && itemType.Implements(typ) {
			result = append(result, item)
			continue
		}

		if itemType.Kind() == reflect.Ptr {
			elemType := itemType.Elem()
			if elemType == typ {
				result = append(result, item)
				continue
			}
			if typ.Kind() == reflect.Interface && elemType.Implements(typ) {
				result = append(result, item)
			}
		}
	}
	return result, nil
}

// Count 返回已注册的实体数量
func (e *EntityContainer) Count() int {
	return e.container.Count()
}

// GetDependency 根据类型获取依赖实例（实现ContainerSource接口）
// Entity返回列表，但依赖注入需要单个实例，所以返回第一个匹配项
// 如果有多个匹配项，返回错误
func (e *EntityContainer) GetDependency(fieldType reflect.Type) (interface{}, error) {
	baseEntityType := reflect.TypeOf((*common.IBaseEntity)(nil)).Elem()
	if fieldType == baseEntityType || fieldType.Implements(baseEntityType) {
		items, err := e.GetByType(fieldType)
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
