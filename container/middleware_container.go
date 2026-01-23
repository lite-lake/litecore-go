package container

import (
	"reflect"

	"github.com/lite-lake/litecore-go/common"
)

// MiddlewareContainer 中间件层容器
type MiddlewareContainer struct {
	*InjectableLayerContainer[common.IBaseMiddleware]
	serviceContainer *ServiceContainer
}

// NewMiddlewareContainer 创建新的中间件容器
func NewMiddlewareContainer(service *ServiceContainer) *MiddlewareContainer {
	return &MiddlewareContainer{
		InjectableLayerContainer: NewInjectableLayerContainer(func(m common.IBaseMiddleware) string {
			return m.MiddlewareName()
		}),
		serviceContainer: service,
	}
}

// RegisterMiddleware 泛型注册函数，按接口类型注册
func RegisterMiddleware[T common.IBaseMiddleware](m *MiddlewareContainer, impl T) error {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	return m.RegisterByType(ifaceType, impl)
}

// GetMiddleware 按接口类型获取
func GetMiddleware[T common.IBaseMiddleware](m *MiddlewareContainer) (T, error) {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	impl := m.GetByType(ifaceType)
	if impl == nil {
		var zero T
		return zero, &InstanceNotFoundError{
			Name:  ifaceType.Name(),
			Layer: "Middleware",
		}
	}
	return impl.(T), nil
}

// InjectAll 执行依赖注入
func (m *MiddlewareContainer) InjectAll() error {
	m.checkManagerContainer("Middleware")

	if m.InjectableLayerContainer.base.container.IsInjected() {
		return nil
	}

	m.InjectableLayerContainer.base.sources = m.InjectableLayerContainer.base.buildSources(m, m.managerContainer, m.serviceContainer)
	return m.InjectableLayerContainer.base.injectAll(m)
}

// GetDependency 根据类型获取依赖实例（实现ContainerSource接口）
func (m *MiddlewareContainer) GetDependency(fieldType reflect.Type) (interface{}, error) {
	if dep, err := resolveDependencyFromManager(fieldType, m.managerContainer); dep != nil || err != nil {
		return dep, err
	}

	baseServiceType := reflect.TypeOf((*common.IBaseService)(nil)).Elem()
	if fieldType == baseServiceType || fieldType.Implements(baseServiceType) {
		if m.serviceContainer == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Service",
			}
		}
		impl := m.serviceContainer.GetByType(fieldType)
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
