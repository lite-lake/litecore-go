package container

import (
	"fmt"
	"reflect"

	"github.com/lite-lake/litecore-go/common"
)

type ServiceContainer struct {
	base                *injectableContainer[common.IBaseService]
	managerContainer    *ManagerContainer
	repositoryContainer *RepositoryContainer
}

func NewServiceContainer(repository *RepositoryContainer) *ServiceContainer {
	return &ServiceContainer{
		base: &injectableContainer[common.IBaseService]{
			container: NewTypedContainer(func(svc common.IBaseService) string {
				return svc.ServiceName()
			}),
		},
		repositoryContainer: repository,
	}
}

func RegisterService[T common.IBaseService](s *ServiceContainer, impl T) error {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	return s.RegisterByType(ifaceType, impl)
}

func GetService[T common.IBaseService](s *ServiceContainer) (T, error) {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	impl := s.GetByType(ifaceType)
	if impl == nil {
		var zero T
		return zero, &InstanceNotFoundError{
			Name:  ifaceType.Name(),
			Layer: "Service",
		}
	}
	return impl.(T), nil
}

func (s *ServiceContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseService) error {
	return s.base.container.Register(ifaceType, impl)
}

func (s *ServiceContainer) InjectAll() error {
	if s.base.container.IsInjected() {
		return nil
	}

	graph, err := s.buildDependencyGraph()
	if err != nil {
		return fmt.Errorf("build dependency graph failed: %w", err)
	}

	sortedTypes, err := topologicalSortByInterfaceType(graph)
	if err != nil {
		return fmt.Errorf("topological sort failed: %w", err)
	}

	s.base.sources = s.base.buildSources(s, s.managerContainer, s.repositoryContainer)
	resolver := NewGenericDependencyResolver(s.base.sources...)

	for _, ifaceType := range sortedTypes {
		svc := s.GetByType(ifaceType)
		if svc != nil {
			if err := injectDependencies(svc, resolver); err != nil {
				return err
			}
			verifyInjectTags(svc)
		}
	}

	s.base.container.setInjected(true)
	return nil
}

func (s *ServiceContainer) buildDependencyGraph() (map[reflect.Type][]reflect.Type, error) {
	graph := make(map[reflect.Type][]reflect.Type)

	s.base.container.RangeItems(func(ifaceType reflect.Type, svc common.IBaseService) bool {
		val := reflect.ValueOf(svc)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if val.Kind() != reflect.Struct {
			return true
		}

		var deps []reflect.Type

		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			tagValue, ok := field.Tag.Lookup("inject")
			if !ok || tagValue == "optional" {
				continue
			}

			fieldType := field.Type
			if s.isBaseServiceType(fieldType) {
				if s.GetByType(fieldType) == nil {
					panic(&DependencyNotFoundError{
						FieldType:     fieldType,
						ContainerType: "Service",
					})
				}
				deps = append(deps, fieldType)
			}
		}

		graph[ifaceType] = deps
		return true
	})

	return graph, nil
}

func (s *ServiceContainer) isBaseServiceType(typ reflect.Type) bool {
	baseServiceType := reflect.TypeOf((*common.IBaseService)(nil)).Elem()
	return typ.Implements(baseServiceType)
}

func (s *ServiceContainer) GetAll() []common.IBaseService {
	return s.base.container.GetAll()
}

func (s *ServiceContainer) GetAllSorted() []common.IBaseService {
	return getAllSorted(s.GetAll(), func(s common.IBaseService) string {
		return s.ServiceName()
	})
}

func (s *ServiceContainer) GetByType(ifaceType reflect.Type) common.IBaseService {
	return s.base.container.GetByType(ifaceType)
}

func (s *ServiceContainer) Count() int {
	return s.base.container.Count()
}

func (s *ServiceContainer) SetManagerContainer(container *ManagerContainer) {
	s.managerContainer = container
}

func (s *ServiceContainer) GetDependency(fieldType reflect.Type) (interface{}, error) {
	if dep, err := resolveDependencyFromManager(fieldType, s.managerContainer); dep != nil || err != nil {
		return dep, err
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

	baseRepositoryType := reflect.TypeOf((*common.IBaseRepository)(nil)).Elem()
	if fieldType == baseRepositoryType || fieldType.Implements(baseRepositoryType) {
		if s.repositoryContainer == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Repository",
			}
		}
		impl := s.repositoryContainer.GetByType(fieldType)
		if impl == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Repository",
			}
		}
		return impl, nil
	}

	return nil, nil
}
