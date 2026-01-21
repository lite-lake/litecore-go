package container

import (
	"fmt"
	"reflect"
	"sort"
	"sync"

	"github.com/lite-lake/litecore-go/common"
)

type BuiltinProvider interface {
	GetConfigProvider() common.IBaseConfigProvider
	GetManagers() []interface{}
}

// RepositoryContainer 存储库层容器
// InjectAll 行为：
// 1. 注入 BuiltinProvider（ConfigProvider、Managers）
// 2. 注入 BaseEntity（从 EntityContainer 获取）
type RepositoryContainer struct {
	mu              sync.RWMutex
	items           map[reflect.Type]common.IBaseRepository
	entityContainer *EntityContainer
	builtinProvider BuiltinProvider
	injected        bool
}

// NewRepositoryContainer 创建新的存储库容器
func NewRepositoryContainer(
	entity *EntityContainer,
) *RepositoryContainer {
	return &RepositoryContainer{
		items:           make(map[reflect.Type]common.IBaseRepository),
		entityContainer: entity,
	}
}

// SetBuiltinProvider 设置内置组件提供者
func (r *RepositoryContainer) SetBuiltinProvider(provider BuiltinProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.builtinProvider = provider
}

// RegisterRepository 泛型注册函数，按接口类型注册
func RegisterRepository[T common.IBaseRepository](r *RepositoryContainer, impl T) error {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	return r.RegisterByType(ifaceType, impl)
}

// RegisterByType 按接口类型注册
func (r *RepositoryContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseRepository) error {
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
func (r *RepositoryContainer) InjectAll() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.injected {
		return nil
	}

	resolver := NewGenericDependencyResolver(r.entityContainer, r)
	for ifaceType, repo := range r.items {
		if err := injectDependencies(repo, resolver); err != nil {
			return fmt.Errorf("inject %v failed: %w", ifaceType, err)
		}
	}

	r.injected = true
	return nil
}

// GetAll 获取所有已注册的存储库
func (r *RepositoryContainer) GetAll() []common.IBaseRepository {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]common.IBaseRepository, 0, len(r.items))
	for _, item := range r.items {
		result = append(result, item)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].RepositoryName() < result[j].RepositoryName()
	})

	return result
}

// GetByType 按接口类型获取（返回单例）
func (r *RepositoryContainer) GetByType(ifaceType reflect.Type) common.IBaseRepository {
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

// GetDependency 根据类型获取依赖实例（实现ContainerSource接口）
func (r *RepositoryContainer) GetDependency(fieldType reflect.Type) (interface{}, error) {
	baseConfigType := reflect.TypeOf((*common.IBaseConfigProvider)(nil)).Elem()
	if fieldType == baseConfigType || fieldType.Implements(baseConfigType) {
		if r.builtinProvider == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Builtin",
			}
		}
		impl := r.builtinProvider.GetConfigProvider()
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
		if r.builtinProvider == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Builtin",
			}
		}

		impl, err := r.getManagerByType(fieldType)
		if err != nil {
			return nil, err
		}
		return impl, nil
	}

	baseEntityType := reflect.TypeOf((*common.IBaseEntity)(nil)).Elem()
	if fieldType == baseEntityType || fieldType.Implements(baseEntityType) {
		items, err := r.entityContainer.GetByType(fieldType)
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

	return nil, nil
}

// getManagerByType 根据类型获取内置管理器
func (r *RepositoryContainer) getManagerByType(fieldType reflect.Type) (interface{}, error) {
	if r.builtinProvider == nil {
		return nil, &DependencyNotFoundError{
			FieldType:     fieldType,
			ContainerType: "Builtin",
		}
	}

	managers := r.builtinProvider.GetManagers()
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
