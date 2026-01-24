package container

import (
	"fmt"
	"reflect"
	"time"

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

// ValidateAll 验证所有定时器配置
// 在程序加载时调用，配置错误直接 panic
// 注意：此方法只做基础验证，Crontab 表达式解析由 SchedulerManager 完成
func (c *SchedulerContainer) ValidateAll() {
	schedulers := c.GetAll()
	if len(schedulers) == 0 {
		return
	}

	for _, scheduler := range schedulers {
		if err := c.validateScheduler(scheduler); err != nil {
			panic(fmt.Sprintf("scheduler %s validation failed: %v", scheduler.SchedulerName(), err))
		}
	}
}

// validateScheduler 验证单个定时器（基础验证）
// Crontab 表达式的完整解析由 SchedulerManager.ValidateScheduler() 完成
func (c *SchedulerContainer) validateScheduler(scheduler common.IBaseScheduler) error {
	// 验证 Crontab 表达式非空
	rule := scheduler.GetRule()
	if rule == "" {
		return fmt.Errorf("rule cannot be empty")
	}

	// 验证时区格式（如果不为空）
	timezone := scheduler.GetTimezone()
	if timezone != "" {
		_, err := time.LoadLocation(timezone)
		if err != nil {
			return fmt.Errorf("invalid timezone: %w", err)
		}
	}

	// 注意：不在这里解析 Crontab 表达式，避免容器层依赖管理器实现
	// Crontab 表达式的解析和验证由 SchedulerManager.ValidateScheduler() 完成
	return nil
}
