package container

import (
	"fmt"
	"reflect"
	"sort"
	"sync"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/util/logger"
)

// ControllerContainer 控制器层容器
// InjectAll 行为：
// 1. 注入 BuiltinProvider（ConfigProvider、Managers）
// 2. 注入 BaseService（从 ServiceContainer 获取）
type ControllerContainer struct {
	mu               sync.RWMutex
	items            map[reflect.Type]common.IBaseController
	serviceContainer *ServiceContainer
	builtinProvider  BuiltinProvider
	loggerRegistry   *logger.LoggerRegistry
	injected         bool
}

// NewControllerContainer 创建新的控制器容器
func NewControllerContainer(
	service *ServiceContainer,
) *ControllerContainer {
	return &ControllerContainer{
		items:            make(map[reflect.Type]common.IBaseController),
		serviceContainer: service,
	}
}

// SetBuiltinProvider 设置内置组件提供者
func (c *ControllerContainer) SetBuiltinProvider(provider BuiltinProvider) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.builtinProvider = provider
	c.loggerRegistry = nil
}

// SetLoggerRegistry 设置 LoggerRegistry
func (c *ControllerContainer) SetLoggerRegistry(registry *logger.LoggerRegistry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loggerRegistry = registry
}

// RegisterController 泛型注册函数，按接口类型注册
func RegisterController[T common.IBaseController](c *ControllerContainer, impl T) error {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	return c.RegisterByType(ifaceType, impl)
}

// RegisterByType 按接口类型注册
func (c *ControllerContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseController) error {
	implType := reflect.TypeOf(impl)

	if impl == nil {
		return &DuplicateRegistrationError{Name: "nil"}
	}

	if !implType.Implements(ifaceType) {
		return &ImplementationDoesNotImplementInterfaceError{
			InterfaceType:  ifaceType,
			Implementation: impl,
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[ifaceType]; exists {
		return &InterfaceAlreadyRegisteredError{
			InterfaceType: ifaceType,
			ExistingImpl:  c.items[ifaceType],
			NewImpl:       impl,
		}
	}

	c.items[ifaceType] = impl
	return nil
}

// InjectAll 注入所有依赖
// 1. 注入 BuiltinProvider（ConfigProvider、Managers）
// 2. 注入 BaseService
func (c *ControllerContainer) InjectAll() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.injected {
		return nil
	}

	resolver := NewGenericDependencyResolver(c.loggerRegistry, c.serviceContainer, c)
	for ifaceType, ctrl := range c.items {
		if err := injectDependencies(ctrl, resolver); err != nil {
			return fmt.Errorf("inject %v failed: %w", ifaceType, err)
		}
	}

	c.injected = true
	return nil
}

// GetAll 获取所有已注册的控制器
func (c *ControllerContainer) GetAll() []common.IBaseController {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]common.IBaseController, 0, len(c.items))
	for _, item := range c.items {
		result = append(result, item)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ControllerName() < result[j].ControllerName()
	})

	return result
}

// GetByType 按接口类型获取（返回单例）
func (c *ControllerContainer) GetByType(ifaceType reflect.Type) common.IBaseController {
	c.mu.RLock()
	defer c.mu.RUnlock()

	impl, exists := c.items[ifaceType]
	if !exists {
		return nil
	}
	return impl
}

// Count 返回已注册的控制器数量
func (c *ControllerContainer) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// GetDependency 根据类型获取依赖实例（实现ContainerSource接口）
func (c *ControllerContainer) GetDependency(fieldType reflect.Type) (interface{}, error) {
	baseConfigType := reflect.TypeOf((*common.IBaseConfigProvider)(nil)).Elem()
	if fieldType == baseConfigType || fieldType.Implements(baseConfigType) {
		if c.builtinProvider == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Builtin",
			}
		}
		impl := c.builtinProvider.GetConfigProvider()
		if impl == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Builtin",
			}
		}
		return impl, nil
	}

	baseManagerType := reflect.TypeOf((*common.IBaseManager)(nil)).Elem()
	if fieldType.Implements(baseManagerType) {
		if c.builtinProvider == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Builtin",
			}
		}

		managers := c.builtinProvider.GetManagers()
		for _, impl := range managers {
			if impl == nil {
				continue
			}
			implType := reflect.TypeOf(impl)
			if implType == fieldType || implType.Implements(fieldType) {
				return impl, nil
			}
		}

		return nil, &DependencyNotFoundError{
			FieldType:     fieldType,
			ContainerType: "Builtin",
		}
	}

	baseServiceType := reflect.TypeOf((*common.IBaseService)(nil)).Elem()
	if fieldType == baseServiceType || fieldType.Implements(baseServiceType) {
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
