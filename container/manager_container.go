package container

import (
	"fmt"
	"reflect"
	"sync"

	"com.litelake.litecore/common"
)

// ManagerContainer 管理器层容器
// InjectAll 行为：
// 1. 注入 BaseConfigProvider（从 ConfigContainer 获取）
// 2. 注入其他 BaseManager（支持同层依赖，按拓扑顺序注入）
type ManagerContainer struct {
	mu              sync.RWMutex
	items           map[string]common.BaseManager
	configContainer *ConfigContainer
	injected        bool // 标记是否已执行注入
}

// NewManagerContainer 创建新的管理器容器
func NewManagerContainer(config *ConfigContainer) *ManagerContainer {
	return &ManagerContainer{
		items:           make(map[string]common.BaseManager),
		configContainer: config,
	}
}

// Register 注册管理器实例
func (m *ManagerContainer) Register(ins common.BaseManager) error {
	if ins == nil {
		return &DuplicateRegistrationError{Name: "nil"}
	}

	name := ins.ManagerName()

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
// 1. 构建依赖图（同层 Manager 依赖）
// 2. 拓扑排序确定注入顺序
// 3. 按顺序注入跨层依赖（Config）和同层依赖（Manager）
func (m *ManagerContainer) InjectAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.injected {
		return nil // 已注入，跳过
	}

	// 1. 构建依赖图
	graph, err := m.buildDependencyGraph()
	if err != nil {
		return fmt.Errorf("build dependency graph failed: %w", err)
	}

	// 2. 拓扑排序
	order, err := topologicalSort(graph)
	if err != nil {
		return fmt.Errorf("topological sort failed: %w", err)
	}

	// 3. 按顺序注入
	for _, name := range order {
		mgr := m.items[name]
		resolver := &managerDependencyResolver{
			container: m,
		}
		if err := injectDependencies(mgr, resolver); err != nil {
			return fmt.Errorf("inject %s failed: %w", name, err)
		}
	}

	m.injected = true
	return nil
}

// buildDependencyGraph 构建管理器之间的依赖图
func (m *ManagerContainer) buildDependencyGraph() (map[string][]string, error) {
	graph := make(map[string][]string)

	for name, item := range m.items {
		deps, err := m.getSameLayerDependencies(item)
		if err != nil {
			return nil, fmt.Errorf("build graph for %s failed: %w", name, err)
		}
		graph[name] = deps
	}

	return graph, nil
}

// getSameLayerDependencies 获取管理器的同层依赖（其他 Manager）
func (m *ManagerContainer) getSameLayerDependencies(instance interface{}) ([]string, error) {
	val := reflect.ValueOf(instance)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, nil
	}

	typ := val.Type()
	var deps []string

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)

		// 只处理标记了 inject 的字段
		if field.Tag.Get("inject") == "" {
			continue
		}

		fieldType := field.Type

		// 判断是否为 BaseManager 类型（同层依赖）
		if m.isBaseManagerType(fieldType) {
			// 查找该字段依赖的 Manager 实例名称
			depName, err := m.findDependencyByName(fieldType)
			if err != nil {
				return nil, err
			}
			if depName != "" {
				deps = append(deps, depName)
			}
		}
	}

	return deps, nil
}

// isBaseManagerType 判断类型是否为 BaseManager 或其子接口
func (m *ManagerContainer) isBaseManagerType(typ reflect.Type) bool {
	// 检查是否实现了 BaseManager 接口
	baseManagerType := reflect.TypeOf((*common.BaseManager)(nil)).Elem()
	return typ.Implements(baseManagerType)
}

// findDependencyByName 根据类型查找管理器实例名称
func (m *ManagerContainer) findDependencyByName(fieldType reflect.Type) (string, error) {
	for name, item := range m.items {
		itemType := reflect.TypeOf(item)
		// 精确匹配
		if itemType == fieldType {
			return name, nil
		}
		// 接口匹配
		if itemType.Implements(fieldType) {
			return name, nil
		}
	}
	return "", nil // 未找到，返回空（不是错误）
}

// GetAll 获取所有已注册的管理器
func (m *ManagerContainer) GetAll() []common.BaseManager {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]common.BaseManager, 0, len(m.items))
	for _, item := range m.items {
		result = append(result, item)
	}
	return result
}

// GetByName 根据名称获取管理器
func (m *ManagerContainer) GetByName(name string) (common.BaseManager, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if ins, exists := m.items[name]; exists {
		return ins, nil
	}
	return nil, &InstanceNotFoundError{Name: name, Layer: "Manager"}
}

// GetByType 根据类型获取管理器
// 返回所有实现了该类型的管理器列表
func (m *ManagerContainer) GetByType(typ reflect.Type) ([]common.BaseManager, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []common.BaseManager
	for _, item := range m.items {
		itemType := reflect.TypeOf(item)
		if typeMatches(itemType, typ) {
			result = append(result, item)
		}
	}
	return result, nil
}

// Count 返回已注册的管理器数量
func (m *ManagerContainer) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.items)
}

// managerDependencyResolver 管理器依赖解析器
type managerDependencyResolver struct {
	container *ManagerContainer
}

// ResolveDependency 解析字段类型对应的依赖实例
func (r *managerDependencyResolver) ResolveDependency(fieldType reflect.Type) (interface{}, error) {
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

	// 2. 尝试从当前 ManagerContainer 获取 BaseManager
	baseManagerType := reflect.TypeOf((*common.BaseManager)(nil)).Elem()
	if fieldType.Implements(baseManagerType) {
		items, err := r.container.GetByType(fieldType)
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

	return nil, &DependencyNotFoundError{
		FieldType:     fieldType,
		ContainerType: "Unknown",
	}
}
