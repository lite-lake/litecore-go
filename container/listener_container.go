package container

import (
	"reflect"

	"github.com/lite-lake/litecore-go/common"
)

// ListenerContainer 监听器层容器
type ListenerContainer struct {
	*InjectableLayerContainer[common.IBaseListener]
	serviceContainer *ServiceContainer
}

// NewListenerContainer 创建新的监听器容器
func NewListenerContainer(service *ServiceContainer) *ListenerContainer {
	return &ListenerContainer{
		InjectableLayerContainer: NewInjectableLayerContainer(func(l common.IBaseListener) string {
			return l.ListenerName()
		}),
		serviceContainer: service,
	}
}

// SetManagerContainer 设置管理器容器
func (l *ListenerContainer) SetManagerContainer(container *ManagerContainer) {
	l.InjectableLayerContainer.SetManagerContainer(container)
}

// RegisterListener 泛型注册函数，按接口类型注册
func RegisterListener[T common.IBaseListener](l *ListenerContainer, impl T) error {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	return l.RegisterByType(ifaceType, impl)
}

// GetListener 按接口类型获取
func GetListener[T common.IBaseListener](l *ListenerContainer) (T, error) {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	impl := l.GetByType(ifaceType)
	if impl == nil {
		var zero T
		return zero, &InstanceNotFoundError{
			Name:  ifaceType.Name(),
			Layer: "Listener",
		}
	}
	return impl.(T), nil
}

// InjectAll 执行依赖注入
func (l *ListenerContainer) InjectAll() error {
	l.checkManagerContainer("Listener")

	if l.InjectableLayerContainer.base.container.IsInjected() {
		return nil
	}

	l.InjectableLayerContainer.base.sources = l.InjectableLayerContainer.base.buildSources(l, l.managerContainer, l.serviceContainer)
	return l.InjectableLayerContainer.base.injectAll(l)
}

// GetDependency 根据类型获取依赖实例（实现ContainerSource接口）
func (l *ListenerContainer) GetDependency(fieldType reflect.Type) (interface{}, error) {
	baseRepositoryType := reflect.TypeOf((*common.IBaseRepository)(nil)).Elem()
	if fieldType == baseRepositoryType || fieldType.Implements(baseRepositoryType) {
		return nil, &DependencyNotFoundError{
			FieldType:     fieldType,
			ContainerType: "Repository",
			Message:       "Listener cannot directly inject Repository, must access data through Service",
		}
	}

	if dep, err := resolveDependencyFromManager(fieldType, l.managerContainer); dep != nil || err != nil {
		return dep, err
	}

	baseServiceType := reflect.TypeOf((*common.IBaseService)(nil)).Elem()
	if fieldType == baseServiceType || fieldType.Implements(baseServiceType) {
		if l.serviceContainer == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Service",
			}
		}
		impl := l.serviceContainer.GetByType(fieldType)
		if impl == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Service",
			}
		}
		return impl, nil
	}

	return nil, nil
}
