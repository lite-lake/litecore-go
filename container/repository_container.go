package container

import (
	"fmt"
	"reflect"
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
	items            map[string]common.BaseRepository
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
		items:            make(map[string]common.BaseRepository),
		configContainer:  config,
		managerContainer: manager,
		entityContainer:  entity,
	}
}

// Register 注册存储库实例
func (r *RepositoryContainer) Register(ins common.BaseRepository) error {
	if ins == nil {
		return &DuplicateRegistrationError{Name: "nil"}
	}

	name := ins.RepositoryName()

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.items[name]; exists {
		return &DuplicateRegistrationError{
			Name:     name,
			Existing: r.items[name],
			New:      ins,
		}
	}

	r.items[name] = ins
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
		return nil // 已注入，跳过
	}

	// 按任意顺序注入（Repository 之间无同层依赖）
	for name, repo := range r.items {
		resolver := &repositoryDependencyResolver{
			container: r,
		}
		if err := injectDependencies(repo, resolver); err != nil {
			return fmt.Errorf("inject %s failed: %w", name, err)
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
	return result
}

// GetByName 根据名称获取存储库
func (r *RepositoryContainer) GetByName(name string) (common.BaseRepository, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if ins, exists := r.items[name]; exists {
		return ins, nil
	}
	return nil, &InstanceNotFoundError{Name: name, Layer: "Repository"}
}

// GetByType 根据类型获取存储库
// 返回所有实现了该类型的存储库列表
func (r *RepositoryContainer) GetByType(typ reflect.Type) ([]common.BaseRepository, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []common.BaseRepository
	for _, item := range r.items {
		itemType := reflect.TypeOf(item)
		if typeMatches(itemType, typ) {
			result = append(result, item)
		}
	}
	return result, nil
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

	// 3. 尝试从 EntityContainer 获取 BaseEntity
	baseEntityType := reflect.TypeOf((*common.BaseEntity)(nil)).Elem()
	if fieldType.Implements(baseEntityType) {
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
