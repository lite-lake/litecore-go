package container

import (
	"reflect"

	"github.com/lite-lake/litecore-go/common"
)

// SchedulerContainer 定时器层容器
type SchedulerContainer struct {
	*InjectableLayerContainer[common.IBaseScheduler]
	serviceContainer *ServiceContainer
}

// NewSchedulerContainer 创建新的定时器容器
func NewSchedulerContainer(service *ServiceContainer) *SchedulerContainer {
	return &SchedulerContainer{
		InjectableLayerContainer: NewInjectableLayerContainer(func(s common.IBaseScheduler) string {
			return s.SchedulerName()
		}),
		serviceContainer: service,
	}
}

// SetManagerContainer 设置管理器容器
func (c *SchedulerContainer) SetManagerContainer(container *ManagerContainer) {
	c.InjectableLayerContainer.SetManagerContainer(container)
}

// RegisterScheduler 泛型注册函数，按接口类型注册
func RegisterScheduler[T common.IBaseScheduler](c *SchedulerContainer, impl T) error {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	return c.RegisterByType(ifaceType, impl)
}

// GetScheduler 按接口类型获取
func GetScheduler[T common.IBaseScheduler](c *SchedulerContainer) (T, error) {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	impl := c.GetByType(ifaceType)
	if impl == nil {
		var zero T
		return zero, &InstanceNotFoundError{
			Name:  ifaceType.Name(),
			Layer: "Scheduler",
		}
	}
	return impl.(T), nil
}

// InjectAll 执行依赖注入
func (c *SchedulerContainer) InjectAll() error {
	c.checkManagerContainer("Scheduler")

	if c.InjectableLayerContainer.base.container.IsInjected() {
		return nil
	}

	c.InjectableLayerContainer.base.sources = c.InjectableLayerContainer.base.buildSources(c, c.managerContainer, c.serviceContainer)
	return c.InjectableLayerContainer.base.injectAll(c)
}

// GetDependency 根据类型获取依赖实例（实现ContainerSource接口）
func (c *SchedulerContainer) GetDependency(fieldType reflect.Type) (interface{}, error) {
	baseRepositoryType := reflect.TypeOf((*common.IBaseRepository)(nil)).Elem()
	if fieldType == baseRepositoryType || fieldType.Implements(baseRepositoryType) {
		return nil, &DependencyNotFoundError{
			FieldType:     fieldType,
			ContainerType: "Repository",
			Message:       "Scheduler cannot directly inject Repository, must access data through Service",
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
