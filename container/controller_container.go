package container

import (
	"reflect"

	"github.com/lite-lake/litecore-go/common"
)

type ControllerContainer struct {
	*InjectableLayerContainer[common.IBaseController]
	serviceContainer *ServiceContainer
}

func NewControllerContainer(service *ServiceContainer) *ControllerContainer {
	return &ControllerContainer{
		InjectableLayerContainer: NewInjectableLayerContainer(func(ctrl common.IBaseController) string {
			return ctrl.ControllerName()
		}),
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

func (c *ControllerContainer) InjectAll() error {
	c.checkManagerContainer("Controller")

	if c.InjectableLayerContainer.base.container.IsInjected() {
		return nil
	}

	c.InjectableLayerContainer.base.sources = c.InjectableLayerContainer.base.buildSources(c, c.managerContainer, c.serviceContainer)
	return c.InjectableLayerContainer.base.injectAll(c)
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
