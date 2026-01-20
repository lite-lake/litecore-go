package container

import (
	"fmt"
	"reflect"
	"sort"
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
	items               map[reflect.Type]common.IBaseService
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
		items:               make(map[reflect.Type]common.IBaseService),
		configContainer:     config,
		managerContainer:    manager,
		repositoryContainer: repository,
	}
}

// RegisterService 泛型注册函数，按接口类型注册
func RegisterService[T common.IBaseService](s *ServiceContainer, impl T) error {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	return s.RegisterByType(ifaceType, impl)
}

// RegisterByType 按接口类型注册
func (s *ServiceContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseService) error {
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

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.items[ifaceType]; exists {
		return &InterfaceAlreadyRegisteredError{
			InterfaceType: ifaceType,
			ExistingImpl:  s.items[ifaceType],
			NewImpl:       impl,
		}
	}

	s.items[ifaceType] = impl
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
		return nil
	}

	graph, err := s.buildDependencyGraph()
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("build dependency graph failed: %w", err)
	}

	order, err := topologicalSortByInterfaceType(graph)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("topological sort failed: %w", err)
	}

	s.mu.Unlock()

	for _, ifaceType := range order {
		s.mu.RLock()
		svc := s.items[ifaceType]
		s.mu.RUnlock()

		resolver := NewGenericDependencyResolver(s.configContainer, s.managerContainer, s.repositoryContainer, s)
		if err := injectDependencies(svc, resolver); err != nil {
			return fmt.Errorf("inject %v failed: %w", ifaceType, err)
		}
	}

	s.mu.Lock()
	s.injected = true
	s.mu.Unlock()

	return nil
}

// buildDependencyGraph 构建服务之间的依赖图（按接口类型）
func (s *ServiceContainer) buildDependencyGraph() (map[reflect.Type][]reflect.Type, error) {
	graph := make(map[reflect.Type][]reflect.Type)

	for ifaceType, item := range s.items {
		deps, err := s.getSameLayerDependencies(item)
		if err != nil {
			return nil, fmt.Errorf("build graph for %v failed: %w", ifaceType, err)
		}
		graph[ifaceType] = deps
	}

	return graph, nil
}

// getSameLayerDependencies 获取服务的同层依赖（其他 Service）
func (s *ServiceContainer) getSameLayerDependencies(instance interface{}) ([]reflect.Type, error) {
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

		if s.isBaseServiceType(fieldType) {
			if _, exists := s.items[fieldType]; !exists {
				return nil, &DependencyNotFoundError{
					FieldType:     fieldType,
					ContainerType: "Service",
				}
			}
			deps = append(deps, fieldType)
		}
	}

	return deps, nil
}

// isBaseServiceType 判断类型是否为 BaseService 或其子接口
func (s *ServiceContainer) isBaseServiceType(typ reflect.Type) bool {
	baseServiceType := reflect.TypeOf((*common.IBaseService)(nil)).Elem()
	return typ.Implements(baseServiceType)
}

// GetAll 获取所有已注册的服务
func (s *ServiceContainer) GetAll() []common.IBaseService {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]common.IBaseService, 0, len(s.items))
	for _, item := range s.items {
		result = append(result, item)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ServiceName() < result[j].ServiceName()
	})

	return result
}

// GetByType 按接口类型获取（返回单例）
func (s *ServiceContainer) GetByType(ifaceType reflect.Type) common.IBaseService {
	s.mu.RLock()
	defer s.mu.RUnlock()

	impl, exists := s.items[ifaceType]
	if !exists {
		return nil
	}
	return impl
}

// Count 返回已注册的服务数量
func (s *ServiceContainer) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

// GetDependency 根据类型获取依赖实例（实现ContainerSource接口）
func (s *ServiceContainer) GetDependency(fieldType reflect.Type) (interface{}, error) {
	baseConfigType := reflect.TypeOf((*common.IBaseConfigProvider)(nil)).Elem()
	if fieldType.Implements(baseConfigType) {
		impl := s.configContainer.GetByType(fieldType)
		if impl == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Config",
			}
		}
		return impl, nil
	}

	baseManagerType := reflect.TypeOf((*common.IBaseManager)(nil)).Elem()
	if fieldType.Implements(baseManagerType) {
		impl := s.managerContainer.GetByType(fieldType)
		if impl == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Manager",
			}
		}
		return impl, nil
	}

	baseRepositoryType := reflect.TypeOf((*common.IBaseRepository)(nil)).Elem()
	if fieldType == baseRepositoryType || fieldType.Implements(baseRepositoryType) {
		impl := s.repositoryContainer.GetByType(fieldType)
		if impl == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Repository",
			}
		}
		return impl, nil
	}

	baseServiceType := reflect.TypeOf((*common.IBaseService)(nil)).Elem()
	if fieldType == baseServiceType || fieldType.Implements(baseServiceType) {
		impl := s.GetByType(fieldType)
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
