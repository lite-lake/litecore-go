package container

import (
	"reflect"

	"github.com/lite-lake/litecore-go/common"
)

type MiddlewareContainer struct {
	base             *injectableContainer[common.IBaseMiddleware]
	managerContainer *ManagerContainer
	serviceContainer *ServiceContainer
}

func NewMiddlewareContainer(service *ServiceContainer) *MiddlewareContainer {
	return &MiddlewareContainer{
		base: &injectableContainer[common.IBaseMiddleware]{
			container: NewTypedContainer(func(m common.IBaseMiddleware) string {
				return m.MiddlewareName()
			}),
		},
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

func (m *MiddlewareContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseMiddleware) error {
	return m.base.container.Register(ifaceType, impl)
}

func (m *MiddlewareContainer) InjectAll() error {
	if m.base.container.IsInjected() {
		return nil
	}

	m.base.sources = m.base.buildSources(m, m.managerContainer, m.serviceContainer)
	return m.base.injectAll(m)
}

func (m *MiddlewareContainer) GetAll() []common.IBaseMiddleware {
	return m.base.container.GetAll()
}

func (m *MiddlewareContainer) GetAllSorted() []common.IBaseMiddleware {
	return getAllSorted(m.GetAll(), func(m common.IBaseMiddleware) string {
		return m.MiddlewareName()
	})
}

func (m *MiddlewareContainer) GetByType(ifaceType reflect.Type) common.IBaseMiddleware {
	return m.base.container.GetByType(ifaceType)
}

func (m *MiddlewareContainer) Count() int {
	return m.base.container.Count()
}

func (m *MiddlewareContainer) SetManagerContainer(container *ManagerContainer) {
	m.managerContainer = container
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
