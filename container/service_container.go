package container

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/lite-lake/litecore-go/common"
)

// ServiceContainer 服务层容器
type ServiceContainer struct {
	base                *injectableContainer[common.IBaseService]
	managerContainer    *ManagerContainer
	repositoryContainer *RepositoryContainer
}

// NewServiceContainer 创建新的服务容器
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

// RegisterService 泛型注册函数，按接口类型注册
func RegisterService[T common.IBaseService](s *ServiceContainer, impl T) error {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	return s.RegisterByType(ifaceType, impl)
}

// GetService 按接口类型获取
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

// RegisterByType 按接口类型注册
func (s *ServiceContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseService) error {
	return s.base.container.Register(ifaceType, impl)
}

// InjectAll 执行依赖注入
func (s *ServiceContainer) InjectAll() error {
	if s.managerContainer == nil {
		panic(&ManagerContainerNotSetError{Layer: "Service"})
	}

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

// buildDependencyGraph 构建服务依赖图
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
			_, ok := field.Tag.Lookup("inject")
			if !ok {
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

// isBaseServiceType 检查类型是否为服务类型
func (s *ServiceContainer) isBaseServiceType(typ reflect.Type) bool {
	baseServiceType := reflect.TypeOf((*common.IBaseService)(nil)).Elem()
	return typ.Implements(baseServiceType)
}

// GetAll 获取所有已注册的服务
func (s *ServiceContainer) GetAll() []common.IBaseService {
	return s.base.container.GetAll()
}

// GetAllSorted 获取所有已注册的服务（按名称排序）
func (s *ServiceContainer) GetAllSorted() []common.IBaseService {
	items := s.GetAll()
	sort.Slice(items, func(i, j int) bool {
		return items[i].ServiceName() < items[j].ServiceName()
	})
	return items
}

// GetByType 按接口类型获取
func (s *ServiceContainer) GetByType(ifaceType reflect.Type) common.IBaseService {
	return s.base.container.GetByType(ifaceType)
}

// Count 返回已注册的服务数量
func (s *ServiceContainer) Count() int {
	return s.base.container.Count()
}

// SetManagerContainer 设置管理器容器
func (s *ServiceContainer) SetManagerContainer(container *ManagerContainer) {
	s.managerContainer = container
}

// GetDependency 根据类型获取依赖实例（实现ContainerSource接口）
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
