package container

import (
	"fmt"
	"reflect"
	"sort"
	"sync"

	"com.litelake.litecore/common"
)

// ManagerContainer 管理器层容器
// InjectAll 行为：
// 1. 注入 BaseConfigProvider（从 ConfigContainer 获取）
// 2. 注入其他 BaseManager（支持同层依赖，按拓扑顺序注入）
type ManagerContainer struct {
	mu              sync.RWMutex
	items           map[reflect.Type]common.IBaseManager
	configContainer *ConfigContainer
	injected        bool
}

// NewManagerContainer 创建新的管理器容器
func NewManagerContainer(config *ConfigContainer) *ManagerContainer {
	return &ManagerContainer{
		items:           make(map[reflect.Type]common.IBaseManager),
		configContainer: config,
	}
}

// RegisterManager 泛型注册函数，按接口类型注册
func RegisterManager[T common.IBaseManager](m *ManagerContainer, impl T) error {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	return m.RegisterByType(ifaceType, impl)
}

// RegisterByType 按接口类型注册
func (m *ManagerContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseManager) error {
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
// 1. 构建依赖图（同层 Manager 依赖）
// 2. 拓扑排序确定注入顺序
// 3. 按顺序注入跨层依赖（Config）和同层依赖（Manager）
func (m *ManagerContainer) InjectAll() error {
	m.mu.Lock()

	if m.injected {
		m.mu.Unlock()
		return nil
	}

	graph, err := m.buildDependencyGraph()
	if err != nil {
		m.mu.Unlock()
		return fmt.Errorf("build dependency graph failed: %w", err)
	}

	order, err := topologicalSortByInterfaceType(graph)
	if err != nil {
		m.mu.Unlock()
		return fmt.Errorf("topological sort failed: %w", err)
	}

	m.mu.Unlock()

	for _, ifaceType := range order {
		m.mu.RLock()
		mgr := m.items[ifaceType]
		m.mu.RUnlock()

		resolver := NewGenericDependencyResolver(m.configContainer, m)
		if err := injectDependencies(mgr, resolver); err != nil {
			return fmt.Errorf("inject %v failed: %w", ifaceType, err)
		}
	}

	m.mu.Lock()
	m.injected = true
	m.mu.Unlock()

	return nil
}

// buildDependencyGraph 构建管理器之间的依赖图（按接口类型）
func (m *ManagerContainer) buildDependencyGraph() (map[reflect.Type][]reflect.Type, error) {
	graph := make(map[reflect.Type][]reflect.Type)

	for ifaceType, item := range m.items {
		deps, err := m.getSameLayerDependencies(item)
		if err != nil {
			return nil, fmt.Errorf("build graph for %v failed: %w", ifaceType, err)
		}
		graph[ifaceType] = deps
	}

	return graph, nil
}

// getSameLayerDependencies 获取管理器的同层依赖（其他 Manager）
func (m *ManagerContainer) getSameLayerDependencies(instance interface{}) ([]reflect.Type, error) {
	val := reflect.ValueOf(instance)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, nil
	}

	typ := val.Type()
	var deps []reflect.Type

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)

		if field.Tag.Get("inject") == "" {
			continue
		}

		fieldType := field.Type

		if m.isBaseManagerType(fieldType) {
			if _, exists := m.items[fieldType]; !exists {
				return nil, &DependencyNotFoundError{
					FieldType:     fieldType,
					ContainerType: "Manager",
				}
			}
			deps = append(deps, fieldType)
		}
	}

	return deps, nil
}

// isBaseManagerType 判断类型是否为 BaseManager 或其子接口
func (m *ManagerContainer) isBaseManagerType(typ reflect.Type) bool {
	baseManagerType := reflect.TypeOf((*common.IBaseManager)(nil)).Elem()
	return typ.Implements(baseManagerType)
}

// GetAll 获取所有已注册的管理器
func (m *ManagerContainer) GetAll() []common.IBaseManager {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]common.IBaseManager, 0, len(m.items))
	for _, item := range m.items {
		result = append(result, item)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ManagerName() < result[j].ManagerName()
	})

	return result
}

// GetByType 按接口类型获取（返回单例）
func (m *ManagerContainer) GetByType(ifaceType reflect.Type) common.IBaseManager {
	m.mu.RLock()
	defer m.mu.RUnlock()

	impl, exists := m.items[ifaceType]
	if !exists {
		return nil
	}
	return impl
}

// Count 返回已注册的管理器数量
func (m *ManagerContainer) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.items)
}

// GetDependency 根据类型获取依赖实例（实现ContainerSource接口）
func (m *ManagerContainer) GetDependency(fieldType reflect.Type) (interface{}, error) {
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
		impl := m.GetByType(fieldType)
		if impl == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Manager",
			}
		}
		return impl, nil
	}

	return nil, nil
}
