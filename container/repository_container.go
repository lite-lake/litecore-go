package container

import (
	"fmt"
	"reflect"
	"sort"
	"sync"

	"com.litelake.litecore/common"
)

// RepositoryContainer 存储库层容器
// InjectAll 行为：
// 1. 注入 BaseConfigProvider（从 ConfigContainer 获取）
// 2. 注入 BaseManager（从 ManagerContainer 获取）
// 3. 注入 BaseEntity（从 EntityContainer 获取）
type RepositoryContainer struct {
	mu               sync.RWMutex
	items            map[reflect.Type]common.BaseRepository
	configContainer  *ConfigContainer
	managerContainer *ManagerContainer
	entityContainer  *EntityContainer
	injected         bool
}

// NewRepositoryContainer 创建新的存储库容器
func NewRepositoryContainer(
	config *ConfigContainer,
	manager *ManagerContainer,
	entity *EntityContainer,
) *RepositoryContainer {
	return &RepositoryContainer{
		items:            make(map[reflect.Type]common.BaseRepository),
		configContainer:  config,
		managerContainer: manager,
		entityContainer:  entity,
	}
}

// RegisterRepository 泛型注册函数，按接口类型注册
func RegisterRepository[T common.BaseRepository](r *RepositoryContainer, impl T) error {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	return r.RegisterByType(ifaceType, impl)
}

// RegisterByType 按接口类型注册
func (r *RepositoryContainer) RegisterByType(ifaceType reflect.Type, impl common.BaseRepository) error {
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

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.items[ifaceType]; exists {
		return &InterfaceAlreadyRegisteredError{
			InterfaceType: ifaceType,
			ExistingImpl:  r.items[ifaceType],
			NewImpl:       impl,
		}
	}

	r.items[ifaceType] = impl
	return nil
}

// InjectAll 注入所有依赖
// 1. 注入 BaseConfigProvider
// 2. 注入 BaseManager
// 3. 注入 BaseEntity
func (r *RepositoryContainer) InjectAll() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.injected {
		return nil
	}

	for ifaceType, repo := range r.items {
		resolver := &repositoryDependencyResolver{container: r}
		if err := injectDependencies(repo, resolver); err != nil {
			return fmt.Errorf("inject %v failed: %w", ifaceType, err)
		}
	}

	r.injected = true
	return nil
}

// GetAll 获取所有已注册的存储库
func (r *RepositoryContainer) GetAll() []common.BaseRepository {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]common.BaseRepository, 0, len(r.items))
	for _, item := range r.items {
		result = append(result, item)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].RepositoryName() < result[j].RepositoryName()
	})

	return result
}

// GetByType 按接口类型获取（返回单例）
func (r *RepositoryContainer) GetByType(ifaceType reflect.Type) common.BaseRepository {
	r.mu.RLock()
	defer r.mu.RUnlock()

	impl, exists := r.items[ifaceType]
	if !exists {
		return nil
	}
	return impl
}

// Count 返回已注册的存储库数量
func (r *RepositoryContainer) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.items)
}

// repositoryDependencyResolver 存储库依赖解析器
type repositoryDependencyResolver struct {
	container *RepositoryContainer
}

// ResolveDependency 解析字段类型对应的依赖实例
func (r *repositoryDependencyResolver) ResolveDependency(fieldType reflect.Type) (interface{}, error) {
	baseConfigType := reflect.TypeOf((*common.BaseConfigProvider)(nil)).Elem()
	if fieldType.Implements(baseConfigType) {
		impl := r.container.configContainer.GetByType(fieldType)
		if impl == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Config",
			}
		}
		return impl, nil
	}

	baseManagerType := reflect.TypeOf((*common.BaseManager)(nil)).Elem()
	if fieldType.Implements(baseManagerType) {
		impl := r.container.managerContainer.GetByType(fieldType)
		if impl == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Manager",
			}
		}
		return impl, nil
	}

	baseEntityType := reflect.TypeOf((*common.BaseEntity)(nil)).Elem()
	if fieldType == baseEntityType || fieldType.Implements(baseEntityType) {
		items, err := r.container.entityContainer.GetByType(fieldType)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Entity",
			}
		}
		if len(items) > 1 {
			var names []string
			for _, item := range items {
				names = append(names, item.EntityName())
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
