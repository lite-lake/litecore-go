package container

import (
	"fmt"
	"reflect"
	"sync"

	"com.litelake.litecore/common"
)

// ControllerContainer 控制器层容器
// InjectAll 行为：
// 1. 注入 BaseConfigProvider（从 ConfigContainer 获取）
// 2. 注入 BaseManager（从 ManagerContainer 获取）
// 3. 注入 BaseService（从 ServiceContainer 获取）
type ControllerContainer struct {
	mu               sync.RWMutex
	items            map[string]common.BaseController
	configContainer  *ConfigContainer
	managerContainer *ManagerContainer
	serviceContainer *ServiceContainer
	injected         bool
}

// NewControllerContainer 创建新的控制器容器
func NewControllerContainer(
	config *ConfigContainer,
	manager *ManagerContainer,
	service *ServiceContainer,
) *ControllerContainer {
	return &ControllerContainer{
		items:            make(map[string]common.BaseController),
		configContainer:  config,
		managerContainer: manager,
		serviceContainer: service,
	}
}

// Register 注册控制器实例
func (c *ControllerContainer) Register(ins common.BaseController) error {
	if ins == nil {
		return &DuplicateRegistrationError{Name: "nil"}
	}

	name := ins.ControllerName()

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[name]; exists {
		return &DuplicateRegistrationError{
			Name:     name,
			Existing: c.items[name],
			New:      ins,
		}
	}

	c.items[name] = ins
	return nil
}

// InjectAll 注入所有依赖
// 1. 注入 BaseConfigProvider
// 2. 注入 BaseManager
// 3. 注入 BaseService
func (c *ControllerContainer) InjectAll() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.injected {
		return nil // 已注入，跳过
	}

	// 按任意顺序注入（Controller 之间无同层依赖）
	for name, ctrl := range c.items {
		resolver := &controllerDependencyResolver{
			container: c,
		}
		if err := injectDependencies(ctrl, resolver); err != nil {
			return fmt.Errorf("inject %s failed: %w", name, err)
		}
	}

	c.injected = true
	return nil
}

// GetAll 获取所有已注册的控制器
func (c *ControllerContainer) GetAll() []common.BaseController {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]common.BaseController, 0, len(c.items))
	for _, item := range c.items {
		result = append(result, item)
	}
	return result
}

// GetByName 根据名称获取控制器
func (c *ControllerContainer) GetByName(name string) (common.BaseController, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if ins, exists := c.items[name]; exists {
		return ins, nil
	}
	return nil, &InstanceNotFoundError{Name: name, Layer: "Controller"}
}

// GetByType 根据类型获取控制器
// 返回所有实现了该类型的控制器列表
func (c *ControllerContainer) GetByType(typ reflect.Type) ([]common.BaseController, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []common.BaseController
	for _, item := range c.items {
		itemType := reflect.TypeOf(item)
		if typeMatches(itemType, typ) {
			result = append(result, item)
		}
	}
	return result, nil
}

// Count 返回已注册的控制器数量
func (c *ControllerContainer) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// controllerDependencyResolver 控制器依赖解析器
type controllerDependencyResolver struct {
	container *ControllerContainer
}

// ResolveDependency 解析字段类型对应的依赖实例
func (r *controllerDependencyResolver) ResolveDependency(fieldType reflect.Type) (interface{}, error) {
	// 1. 尝试从 ConfigContainer 获取 BaseConfigProvider
	baseConfigType := reflect.TypeOf((*common.BaseConfigProvider)(nil)).Elem()
	if fieldType.Implements(baseConfigType) {
		items, err := r.container.configContainer.GetByType(fieldType)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Config",
			}
		}
		if len(items) > 1 {
			var names []string
			for _, item := range items {
				names = append(names, item.ConfigProviderName())
			}
			return nil, &AmbiguousMatchError{
				FieldType:  fieldType,
				Candidates: names,
			}
		}
		return items[0], nil
	}

	// 2. 尝试从 ManagerContainer 获取 BaseManager
	baseManagerType := reflect.TypeOf((*common.BaseManager)(nil)).Elem()
	if fieldType.Implements(baseManagerType) {
		items, err := r.container.managerContainer.GetByType(fieldType)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Manager",
			}
		}
		if len(items) > 1 {
			var names []string
			for _, item := range items {
				names = append(names, item.ManagerName())
			}
			return nil, &AmbiguousMatchError{
				FieldType:  fieldType,
				Candidates: names,
			}
		}
		return items[0], nil
	}

	// 3. 尝试从 ServiceContainer 获取 BaseService
	baseServiceType := reflect.TypeOf((*common.BaseService)(nil)).Elem()
	if fieldType.Implements(baseServiceType) {
		items, err := r.container.serviceContainer.GetByType(fieldType)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Service",
			}
		}
		if len(items) > 1 {
			var names []string
			for _, item := range items {
				names = append(names, item.ServiceName())
			}
			return nil, &AmbiguousMatchError{
				FieldType:  fieldType,
				Candidates: names,
			}
		}
		return items[0], nil
	}

	return nil, &DependencyNotFoundError{
		FieldType:     fieldType,
		ContainerType: "Unknown",
	}
}
