package container

import (
	"reflect"
	"sort"

	"github.com/lite-lake/litecore-go/common"
)

// RepositoryContainer 仓储层容器
type RepositoryContainer struct {
	base             *injectableContainer[common.IBaseRepository]
	managerContainer *ManagerContainer
	entityContainer  *EntityContainer
}

// NewRepositoryContainer 创建新的仓储容器
func NewRepositoryContainer(entity *EntityContainer) *RepositoryContainer {
	return &RepositoryContainer{
		base: &injectableContainer[common.IBaseRepository]{
			container: NewTypedContainer(func(repo common.IBaseRepository) string {
				return repo.RepositoryName()
			}),
		},
		entityContainer: entity,
	}
}

// RegisterRepository 泛型注册函数，按接口类型注册
func RegisterRepository[T common.IBaseRepository](r *RepositoryContainer, impl T) error {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	return r.RegisterByType(ifaceType, impl)
}

// GetRepository 按接口类型获取
func GetRepository[T common.IBaseRepository](r *RepositoryContainer) (T, error) {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	impl := r.GetByType(ifaceType)
	if impl == nil {
		var zero T
		return zero, &InstanceNotFoundError{
			Name:  ifaceType.Name(),
			Layer: "Repository",
		}
	}
	return impl.(T), nil
}

// RegisterByType 按接口类型注册
func (r *RepositoryContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseRepository) error {
	return r.base.container.Register(ifaceType, impl)
}

// InjectAll 执行依赖注入
func (r *RepositoryContainer) InjectAll() error {
	if r.managerContainer == nil {
		panic(&ManagerContainerNotSetError{Layer: "Repository"})
	}

	if r.base.container.IsInjected() {
		return nil
	}

	r.base.sources = r.base.buildSources(r, r.managerContainer, r.entityContainer)
	return r.base.injectAll(r)
}

// GetAll 获取所有已注册的仓储
func (r *RepositoryContainer) GetAll() []common.IBaseRepository {
	return r.base.container.GetAll()
}

// GetAllSorted 获取所有已注册的仓储（按名称排序）
func (r *RepositoryContainer) GetAllSorted() []common.IBaseRepository {
	items := r.GetAll()
	sort.Slice(items, func(i, j int) bool {
		return items[i].RepositoryName() < items[j].RepositoryName()
	})
	return items
}

// GetByType 按接口类型获取
func (r *RepositoryContainer) GetByType(ifaceType reflect.Type) common.IBaseRepository {
	return r.base.container.GetByType(ifaceType)
}

// Count 返回已注册的仓储数量
func (r *RepositoryContainer) Count() int {
	return r.base.container.Count()
}

// SetManagerContainer 设置管理器容器
func (r *RepositoryContainer) SetManagerContainer(container *ManagerContainer) {
	r.managerContainer = container
}

// GetDependency 根据类型获取依赖实例（实现ContainerSource接口）
func (r *RepositoryContainer) GetDependency(fieldType reflect.Type) (interface{}, error) {
	if dep, err := resolveDependencyFromManager(fieldType, r.managerContainer); dep != nil || err != nil {
		return dep, err
	}

	if dep, err := resolveDependencyFromEntity(fieldType, r.entityContainer); dep != nil || err != nil {
		return dep, err
	}

	return nil, nil
}
