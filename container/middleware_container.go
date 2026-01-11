package container

import (
	"fmt"
	"reflect"
	"sync"

	"com.litelake.litecore/common"
)

// MiddlewareContainer 中间件层容器
// InjectAll 行为：
// 1. 注入 BaseConfigProvider（从 ConfigContainer 获取）
// 2. 注入 BaseManager（从 ManagerContainer 获取）
// 3. 注入 BaseService（从 ServiceContainer 获取）
type MiddlewareContainer struct {
	mu                sync.RWMutex
	items             map[string]common.BaseMiddleware
	configContainer   *ConfigContainer
	managerContainer  *ManagerContainer
	serviceContainer  *ServiceContainer
	injected          bool
}

// NewMiddlewareContainer 创建新的中间件容器
func NewMiddlewareContainer(
	config *ConfigContainer,
	manager *ManagerContainer,
	service *ServiceContainer,
) *MiddlewareContainer {
	return &MiddlewareContainer{
		items:            make(map[string]common.BaseMiddleware),
		configContainer:  config,
		managerContainer: manager,
		serviceContainer: service,
	}
}

// Register 注册中间件实例
func (m *MiddlewareContainer) Register(ins common.BaseMiddleware) error {
	if ins == nil {
		return &DuplicateRegistrationError{Name: "nil"}
	}

	name := ins.MiddlewareName()

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.items[name]; exists {
		return &DuplicateRegistrationError{
			Name:     name,
			Existing: m.items[name],
			New:      ins,
		}
	}

	m.items[name] = ins
	return nil
}

// InjectAll 注入所有依赖
// 1. 注入 BaseConfigProvider
// 2. 注入 BaseManager
// 3. 注入 BaseService
func (m *MiddlewareContainer) InjectAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.injected {
		return nil // 已注入，跳过
	}

	// 按任意顺序注入（Middleware 之间无同层依赖）
	for name, mw := range m.items {
		resolver := &middlewareDependencyResolver{
			container: m,
		}
		if err := injectDependencies(mw, resolver); err != nil {
			return fmt.Errorf("inject %s failed: %w", name, err)
		}
	}

	m.injected = true
	return nil
}

// GetAll 获取所有已注册的中间件
func (m *MiddlewareContainer) GetAll() []common.BaseMiddleware {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]common.BaseMiddleware, 0, len(m.items))
	for _, item := range m.items {
		result = append(result, item)
	}
	return result
}

// GetByName 根据名称获取中间件
func (m *MiddlewareContainer) GetByName(name string) (common.BaseMiddleware, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if ins, exists := m.items[name]; exists {
		return ins, nil
	}
	return nil, &InstanceNotFoundError{Name: name, Layer: "Middleware"}
}

// GetByType 根据类型获取中间件
// 返回所有实现了该类型的中间件列表
func (m *MiddlewareContainer) GetByType(typ reflect.Type) ([]common.BaseMiddleware, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []common.BaseMiddleware
	for _, item := range m.items {
		itemType := reflect.TypeOf(item)
		if typeMatches(itemType, typ) {
			result = append(result, item)
		}
	}
	return result, nil
}

// Count 返回已注册的中间件数量
func (m *MiddlewareContainer) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.items)
}

// middlewareDependencyResolver 中间件依赖解析器
type middlewareDependencyResolver struct {
	container *MiddlewareContainer
}

// ResolveDependency 解析字段类型对应的依赖实例
func (r *middlewareDependencyResolver) ResolveDependency(fieldType reflect.Type) (interface{}, error) {
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
				FieldType:   fieldType,
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
				FieldType:   fieldType,
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
				FieldType:   fieldType,
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
