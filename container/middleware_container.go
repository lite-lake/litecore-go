package container

import (
	"fmt"
	"reflect"
	"sort"
	"sync"

	"github.com/lite-lake/litecore-go/common"
)

// MiddlewareContainer 中间件层容器
// InjectAll 行为：
// 1. 注入 BaseConfigProvider（从 ConfigContainer 获取）
// 2. 注入 BaseManager（从 ManagerContainer 获取）
// 3. 注入 BaseService（从 ServiceContainer 获取）
type MiddlewareContainer struct {
	mu               sync.RWMutex
	items            map[reflect.Type]common.IBaseMiddleware
	configContainer  *ConfigContainer
	managerContainer *ManagerContainer
	serviceContainer *ServiceContainer
	injected         bool
}

// NewMiddlewareContainer 创建新的中间件容器
func NewMiddlewareContainer(
	config *ConfigContainer,
	manager *ManagerContainer,
	service *ServiceContainer,
) *MiddlewareContainer {
	return &MiddlewareContainer{
		items:            make(map[reflect.Type]common.IBaseMiddleware),
		configContainer:  config,
		managerContainer: manager,
		serviceContainer: service,
	}
}

// RegisterMiddleware 泛型注册函数，按接口类型注册
func RegisterMiddleware[T common.IBaseMiddleware](m *MiddlewareContainer, impl T) error {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	return m.RegisterByType(ifaceType, impl)
}

// RegisterByType 按接口类型注册
func (m *MiddlewareContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseMiddleware) error {
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

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.items[ifaceType]; exists {
		return &InterfaceAlreadyRegisteredError{
			InterfaceType: ifaceType,
			ExistingImpl:  m.items[ifaceType],
			NewImpl:       impl,
		}
	}

	m.items[ifaceType] = impl
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
		return nil
	}

	for ifaceType, mw := range m.items {
		resolver := NewGenericDependencyResolver(m.configContainer, m.managerContainer, m.serviceContainer, m)
		if err := injectDependencies(mw, resolver); err != nil {
			return fmt.Errorf("inject %v failed: %w", ifaceType, err)
		}
	}

	m.injected = true
	return nil
}

// GetAll 获取所有已注册的中间件
func (m *MiddlewareContainer) GetAll() []common.IBaseMiddleware {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]common.IBaseMiddleware, 0, len(m.items))
	for _, item := range m.items {
		result = append(result, item)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].MiddlewareName() < result[j].MiddlewareName()
	})

	return result
}

// GetByType 按接口类型获取（返回单例）
func (m *MiddlewareContainer) GetByType(ifaceType reflect.Type) common.IBaseMiddleware {
	m.mu.RLock()
	defer m.mu.RUnlock()

	impl, exists := m.items[ifaceType]
	if !exists {
		return nil
	}
	return impl
}

// Count 返回已注册的中间件数量
func (m *MiddlewareContainer) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.items)
}

// GetDependency 根据类型获取依赖实例（实现ContainerSource接口）
func (m *MiddlewareContainer) GetDependency(fieldType reflect.Type) (interface{}, error) {
	baseConfigType := reflect.TypeOf((*common.IBaseConfigProvider)(nil)).Elem()
	if fieldType == baseConfigType || fieldType.Implements(baseConfigType) {
		impl := m.configContainer.GetByType(fieldType)
		if impl == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Config",
			}
		}
		return impl, nil
	}

	baseManagerType := reflect.TypeOf((*common.IBaseManager)(nil)).Elem()
	if fieldType == baseManagerType || fieldType.Implements(baseManagerType) {
		impl := m.managerContainer.GetByType(fieldType)
		if impl == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Manager",
			}
		}
		return impl, nil
	}

	baseServiceType := reflect.TypeOf((*common.IBaseService)(nil)).Elem()
	if fieldType == baseServiceType || fieldType.Implements(baseServiceType) {
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
