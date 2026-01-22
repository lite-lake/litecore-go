package container

import (
	"reflect"

	"github.com/lite-lake/litecore-go/common"
)

type MiddlewareContainer struct {
	*InjectableLayerContainer[common.IBaseMiddleware]
	serviceContainer *ServiceContainer
}

func NewMiddlewareContainer(service *ServiceContainer) *MiddlewareContainer {
	return &MiddlewareContainer{
		InjectableLayerContainer: NewInjectableLayerContainer(func(m common.IBaseMiddleware) string {
			return m.MiddlewareName()
		}),
		serviceContainer: service,
	}
}

func RegisterMiddleware[T common.IBaseMiddleware](m *MiddlewareContainer, impl T) error {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	return m.RegisterByType(ifaceType, impl)
}

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

func (m *MiddlewareContainer) InjectAll() error {
	m.checkManagerContainer("Middleware")

	if m.InjectableLayerContainer.base.container.IsInjected() {
		return nil
	}

	m.InjectableLayerContainer.base.sources = m.InjectableLayerContainer.base.buildSources(m, m.managerContainer, m.serviceContainer)
	return m.InjectableLayerContainer.base.injectAll(m)
}

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
