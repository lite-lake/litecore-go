package container

import (
	"reflect"

	"github.com/lite-lake/litecore-go/common"
)

type ControllerContainer struct {
	base             *injectableContainer[common.IBaseController]
	managerContainer *ManagerContainer
	serviceContainer *ServiceContainer
}

func NewControllerContainer(service *ServiceContainer) *ControllerContainer {
	return &ControllerContainer{
		base: &injectableContainer[common.IBaseController]{
			container: NewTypedContainer(func(ctrl common.IBaseController) string {
				return ctrl.ControllerName()
			}),
		},
		serviceContainer: service,
	}
}

func RegisterController[T common.IBaseController](c *ControllerContainer, impl T) error {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	return c.RegisterByType(ifaceType, impl)
}

func GetController[T common.IBaseController](c *ControllerContainer) (T, error) {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	impl := c.GetByType(ifaceType)
	if impl == nil {
		var zero T
		return zero, &InstanceNotFoundError{
			Name:  ifaceType.Name(),
			Layer: "Controller",
		}
	}
	return impl.(T), nil
}

func (c *ControllerContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseController) error {
	return c.base.container.Register(ifaceType, impl)
}

func (c *ControllerContainer) InjectAll() error {
	if c.base.container.IsInjected() {
		return nil
	}

	c.base.sources = c.base.buildSources(c, c.managerContainer, c.serviceContainer)
	return c.base.injectAll(c)
}

func (c *ControllerContainer) GetAll() []common.IBaseController {
	return c.base.container.GetAll()
}

func (c *ControllerContainer) GetAllSorted() []common.IBaseController {
	return getAllSorted(c.GetAll(), func(c common.IBaseController) string {
		return c.ControllerName()
	})
}

func (c *ControllerContainer) GetByType(ifaceType reflect.Type) common.IBaseController {
	return c.base.container.GetByType(ifaceType)
}

func (c *ControllerContainer) Count() int {
	return c.base.container.Count()
}

func (c *ControllerContainer) SetManagerContainer(container *ManagerContainer) {
	c.managerContainer = container
}

func (c *ControllerContainer) GetDependency(fieldType reflect.Type) (interface{}, error) {
	if dep, err := resolveDependencyFromManager(fieldType, c.managerContainer); dep != nil || err != nil {
		return dep, err
	}

	baseServiceType := reflect.TypeOf((*common.IBaseService)(nil)).Elem()
	if fieldType == baseServiceType || fieldType.Implements(baseServiceType) {
		if c.serviceContainer == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Service",
			}
		}
		impl := c.serviceContainer.GetByType(fieldType)
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
