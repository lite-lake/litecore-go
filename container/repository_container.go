package container

import (
	"reflect"
	"sort"

	"github.com/lite-lake/litecore-go/common"
)

type RepositoryContainer struct {
	base             *injectableContainer[common.IBaseRepository]
	managerContainer *ManagerContainer
	entityContainer  *EntityContainer
}

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

func RegisterRepository[T common.IBaseRepository](r *RepositoryContainer, impl T) error {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	return r.RegisterByType(ifaceType, impl)
}

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

func (r *RepositoryContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseRepository) error {
	return r.base.container.Register(ifaceType, impl)
}

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

func (r *RepositoryContainer) GetAll() []common.IBaseRepository {
	return r.base.container.GetAll()
}

func (r *RepositoryContainer) GetAllSorted() []common.IBaseRepository {
	items := r.GetAll()
	sort.Slice(items, func(i, j int) bool {
		return items[i].RepositoryName() < items[j].RepositoryName()
	})
	return items
}

func (r *RepositoryContainer) GetByType(ifaceType reflect.Type) common.IBaseRepository {
	return r.base.container.GetByType(ifaceType)
}

func (r *RepositoryContainer) Count() int {
	return r.base.container.Count()
}

func (r *RepositoryContainer) SetManagerContainer(container *ManagerContainer) {
	r.managerContainer = container
}

func (r *RepositoryContainer) GetDependency(fieldType reflect.Type) (interface{}, error) {
	if dep, err := resolveDependencyFromManager(fieldType, r.managerContainer); dep != nil || err != nil {
		return dep, err
	}

	if dep, err := resolveDependencyFromEntity(fieldType, r.entityContainer); dep != nil || err != nil {
		return dep, err
	}

	return nil, nil
}
