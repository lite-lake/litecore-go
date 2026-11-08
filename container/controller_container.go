package container

import (
	"reflect"

	"github.com/lite-lake/litecore-go/common"
)

// ControllerContainer 控制器层容器
type ControllerContainer struct {
	*InjectableLayerContainer[common.IBaseController]
	serviceContainer *ServiceContainer
}

// NewControllerContainer 创建新的控制器容器
func NewControllerContainer(service *ServiceContainer) *ControllerContainer {
	return &ControllerContainer{
		InjectableLayerContainer: NewInjectableLayerContainer(func(ctrl common.IBaseController) string {
			return ctrl.ControllerName()
		}),
		serviceContainer: service,
	}
}

// RegisterController 泛型注册函数，按接口类型注册
func RegisterController[T common.IBaseController](c *ControllerContainer, impl T) error {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	return c.RegisterByType(ifaceType, impl)
}

// GetController 按接口类型获取
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

// InjectAll 执行依赖注入
func (c *ControllerContainer) InjectAll() error {
	c.checkManagerContainer("Controller")

	if c.InjectableLayerContainer.base.container.IsInjected() {
		return nil
	}

	c.InjectableLayerContainer.base.sources = c.InjectableLayerContainer.base.buildSources(c, c.managerContainer, c.serviceContainer)
	return c.InjectableLayerContainer.base.injectAll(c)
}

// GetDependency 根据类型获取依赖实例（实现ContainerSource接口）
func (c *ControllerContainer) GetDependency(fieldType reflect.Type) (interface{}, error) {
	baseRepositoryType := reflect.TypeOf((*common.IBaseRepository)(nil)).Elem()
	if fieldType == baseRepositoryType || fieldType.Implements(baseRepositoryType) {
		return nil, &DependencyNotFoundError{
			FieldType:     fieldType,
			ContainerType: "Repository",
			Message:       "Controller cannot directly inject Repository, must access data through Service",
		}
	}

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
