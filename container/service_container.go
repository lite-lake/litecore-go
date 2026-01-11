package container

import (
	"fmt"
	"reflect"
	"sync"

	"com.litelake.litecore/common"
)

// ServiceContainer 服务层容器
// InjectAll 行为：
// 1. 注入 BaseConfigProvider（从 ConfigContainer 获取）
// 2. 注入 BaseManager（从 ManagerContainer 获取）
// 3. 注入 BaseRepository（从 RepositoryContainer 获取）
// 4. 注入其他 BaseService（支持同层依赖，按拓扑顺序注入）
type ServiceContainer struct {
	mu                  sync.RWMutex
	items               map[string]common.BaseService
	configContainer     *ConfigContainer
	managerContainer    *ManagerContainer
	repositoryContainer *RepositoryContainer
	injected            bool
}

// NewServiceContainer 创建新的服务容器
func NewServiceContainer(
	config *ConfigContainer,
	manager *ManagerContainer,
	repository *RepositoryContainer,
) *ServiceContainer {
	return &ServiceContainer{
		items:               make(map[string]common.BaseService),
		configContainer:     config,
		managerContainer:    manager,
		repositoryContainer: repository,
	}
}

// Register 注册服务实例
func (s *ServiceContainer) Register(ins common.BaseService) error {
	if ins == nil {
		return &DuplicateRegistrationError{Name: "nil"}
	}

	name := ins.ServiceName()

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.items[name]; exists {
		return &DuplicateRegistrationError{
			Name:     name,
			Existing: s.items[name],
			New:      ins,
		}
	}

	s.items[name] = ins
	return nil
}

// InjectAll 注入所有依赖
// 1. 构建依赖图（同层 Service 依赖）
// 2. 拓扑排序确定注入顺序
// 3. 按顺序注入跨层依赖和同层依赖
func (s *ServiceContainer) InjectAll() error {
	s.mu.Lock()

	if s.injected {
		s.mu.Unlock()
		return nil // 已注入，跳过
	}

	// 1. 构建依赖图
	graph, err := s.buildDependencyGraph()
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("build dependency graph failed: %w", err)
	}

	// 2. 拓扑排序
	order, err := topologicalSort(graph)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("topological sort failed: %w", err)
	}

	s.mu.Unlock()

	// 3. 按顺序注入（在锁外执行，避免死锁）
	for _, name := range order {
		s.mu.RLock()
		svc := s.items[name]
		s.mu.RUnlock()

		fmt.Printf("[DEBUG] Injecting dependencies for %s...\n", name)
		resolver := &serviceDependencyResolver{
			container: s,
		}
		if err := injectDependencies(svc, resolver); err != nil {
			fmt.Printf("[DEBUG] Failed to inject %s: %v\n", name, err)
			return fmt.Errorf("inject %s failed: %w", name, err)
		}
		fmt.Printf("[DEBUG] Successfully injected %s\n", name)
	}

	s.mu.Lock()
	s.injected = true
	s.mu.Unlock()

	return nil
}

// buildDependencyGraph 构建服务之间的依赖图
func (s *ServiceContainer) buildDependencyGraph() (map[string][]string, error) {
	graph := make(map[string][]string)

	for name, item := range s.items {
		deps, err := s.getSameLayerDependencies(item)
		if err != nil {
			return nil, fmt.Errorf("build graph for %s failed: %w", name, err)
		}
		graph[name] = deps
	}

	return graph, nil
}

// getSameLayerDependencies 获取服务的同层依赖（其他 Service）
func (s *ServiceContainer) getSameLayerDependencies(instance interface{}) ([]string, error) {
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

		// 判断是否为 BaseService 类型（同层依赖）
		if s.isBaseServiceType(fieldType) {
			// 查找该字段依赖的 Service 实例名称
			depName, err := s.findDependencyByName(fieldType)
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

// isBaseServiceType 判断类型是否为 BaseService 或其子接口
func (s *ServiceContainer) isBaseServiceType(typ reflect.Type) bool {
	baseServiceType := reflect.TypeOf((*common.BaseService)(nil)).Elem()
	return typ.Implements(baseServiceType)
}

// findDependencyByName 根据类型查找服务实例名称
func (s *ServiceContainer) findDependencyByName(fieldType reflect.Type) (string, error) {
	for name, item := range s.items {
		itemType := reflect.TypeOf(item)
		if itemType == fieldType {
			return name, nil
		}
		if itemType.Implements(fieldType) {
			return name, nil
		}
	}
	return "", nil
}

// GetAll 获取所有已注册的服务
func (s *ServiceContainer) GetAll() []common.BaseService {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]common.BaseService, 0, len(s.items))
	for _, item := range s.items {
		result = append(result, item)
	}
	return result
}

// GetByName 根据名称获取服务
func (s *ServiceContainer) GetByName(name string) (common.BaseService, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if ins, exists := s.items[name]; exists {
		return ins, nil
	}
	return nil, &InstanceNotFoundError{Name: name, Layer: "Service"}
}

// GetByType 根据类型获取服务
// 返回所有实现了该类型的服务列表
func (s *ServiceContainer) GetByType(typ reflect.Type) ([]common.BaseService, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []common.BaseService
	for _, item := range s.items {
		itemType := reflect.TypeOf(item)
		if typeMatches(itemType, typ) {
			result = append(result, item)
		}
	}
	return result, nil
}

// Count 返回已注册的服务数量
func (s *ServiceContainer) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

// serviceDependencyResolver 服务依赖解析器
type serviceDependencyResolver struct {
	container *ServiceContainer
}

// ResolveDependency 解析字段类型对应的依赖实例
func (r *serviceDependencyResolver) ResolveDependency(fieldType reflect.Type) (interface{}, error) {
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

	// 3. 尝试从 RepositoryContainer 获取 BaseRepository
	baseRepositoryType := reflect.TypeOf((*common.BaseRepository)(nil)).Elem()
	if fieldType.Implements(baseRepositoryType) {
		items, err := r.container.repositoryContainer.GetByType(fieldType)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Repository",
			}
		}
		if len(items) > 1 {
			var names []string
			for _, item := range items {
				names = append(names, item.RepositoryName())
			}
			return nil, &AmbiguousMatchError{
				FieldType:  fieldType,
				Candidates: names,
			}
		}
		return items[0], nil
	}

	// 4. 尝试从当前 ServiceContainer 获取 BaseService（同层依赖）
	baseServiceType := reflect.TypeOf((*common.BaseService)(nil)).Elem()
	if fieldType.Implements(baseServiceType) {
		items, err := r.container.GetByType(fieldType)
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
