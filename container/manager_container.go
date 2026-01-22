package container

import (
	"reflect"
	"sort"

	"github.com/lite-lake/litecore-go/common"
)

// ManagerContainer 管理器层容器
type ManagerContainer struct {
	container *TypedContainer[common.IBaseManager]
}

// NewManagerContainer 创建新的管理器容器
func NewManagerContainer() *ManagerContainer {
	return &ManagerContainer{
		container: NewTypedContainer(func(manager common.IBaseManager) string {
			return manager.ManagerName()
		}),
	}
}

// RegisterManager 泛型注册函数，按接口类型注册
func RegisterManager[T common.IBaseManager](m *ManagerContainer, impl T) error {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	return m.RegisterByType(ifaceType, impl)
}

// GetManager 按接口类型获取
func GetManager[T common.IBaseManager](m *ManagerContainer) (T, error) {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	impl := m.GetByType(ifaceType)
	if impl == nil {
		var zero T
		return zero, &InstanceNotFoundError{
			Name:  ifaceType.Name(),
			Layer: "Manager",
		}
	}
	return impl.(T), nil
}

// RegisterByType 按接口类型注册
func (m *ManagerContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseManager) error {
	return m.container.Register(ifaceType, impl)
}

// GetByType 按接口类型获取（返回单例）
func (m *ManagerContainer) GetByType(ifaceType reflect.Type) common.IBaseManager {
	return m.container.GetByType(ifaceType)
}

// GetAll 获取所有已注册的管理器
func (m *ManagerContainer) GetAll() []common.IBaseManager {
	return m.container.GetAll()
}

// GetNames 获取所有管理器的名称
func (m *ManagerContainer) GetNames() []string {
	return m.container.GetNames()
}

// Count 返回已注册的管理器数量
func (m *ManagerContainer) Count() int {
	return m.container.Count()
}

// GetAllSorted 获取所有已注册的管理器（按名称排序）
func (m *ManagerContainer) GetAllSorted() []common.IBaseManager {
	result := m.GetAll()
	sort.Slice(result, func(i, j int) bool {
		return result[i].ManagerName() < result[j].ManagerName()
	})
	return result
}

// GetDependency 根据类型获取依赖实例（实现ContainerSource接口）
func (m *ManagerContainer) GetDependency(fieldType reflect.Type) (interface{}, error) {
	baseManagerType := reflect.TypeOf((*common.IBaseManager)(nil)).Elem()
	if fieldType == baseManagerType || fieldType.Implements(baseManagerType) {
		impl := m.GetByType(fieldType)
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
