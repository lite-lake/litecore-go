package container

import (
	"fmt"
	"reflect"
	"sort"
	"sync"

	"github.com/lite-lake/litecore-go/common"
)

// UninjectedFieldError 未注入字段错误
type UninjectedFieldError struct {
	InstanceName string
	FieldName    string
	FieldType    reflect.Type
}

func (e *UninjectedFieldError) Error() string {
	return fmt.Sprintf("field %s.%s (type %s) marked with inject:\"\" is still nil after injection",
		e.InstanceName, e.FieldName, e.FieldType)
}

// verifyInjectTags 验证所有 inject:"" 标签的字段是否已被注入
func verifyInjectTags(instance interface{}) {
	val := reflect.ValueOf(instance)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return
	}

	typ := val.Type()
	instanceName := extractNameFromType(typ)

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		tagValue, ok := field.Tag.Lookup("inject")
		if !ok {
			continue
		}

		if tagValue == "optional" {
			continue
		}

		if !fieldVal.CanInterface() || fieldVal.IsZero() || fieldVal.IsNil() {
			panic(&UninjectedFieldError{
				InstanceName: instanceName,
				FieldName:    field.Name,
				FieldType:    field.Type,
			})
		}
	}
}

// InjectableContainer 可注入容器接口
type InjectableContainer interface {
	InjectAll() error
}

// TypedContainer 类型化容器
type TypedContainer[T any] struct {
	mu       sync.RWMutex
	items    map[reflect.Type]T
	nameFunc func(T) string
	injected bool
}

func NewTypedContainer[T any](nameFunc func(T) string) *TypedContainer[T] {
	return &TypedContainer[T]{
		items:    make(map[reflect.Type]T),
		nameFunc: nameFunc,
	}
}

func (c *TypedContainer[T]) Register(ifaceType reflect.Type, impl T) error {
	implVal := reflect.ValueOf(impl)

	if !implVal.IsValid() || (implVal.Kind() == reflect.Ptr && implVal.IsNil()) {
		return &DuplicateRegistrationError{Name: "nil"}
	}

	implType := implVal.Type()

	if !implType.Implements(ifaceType) {
		return &ImplementationDoesNotImplementInterfaceError{
			InterfaceType:  ifaceType,
			Implementation: impl,
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[ifaceType]; exists {
		return &InterfaceAlreadyRegisteredError{
			InterfaceType: ifaceType,
			ExistingImpl:  c.items[ifaceType],
			NewImpl:       impl,
		}
	}

	c.items[ifaceType] = impl
	return nil
}

func (c *TypedContainer[T]) GetByType(ifaceType reflect.Type) T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	impl, exists := c.items[ifaceType]
	if !exists {
		var zero T
		return zero
	}
	return impl
}

func (c *TypedContainer[T]) GetAll() []T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]T, 0, len(c.items))
	for _, item := range c.items {
		result = append(result, item)
	}
	return result
}

func (c *TypedContainer[T]) GetNames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]string, 0, len(c.items))
	for _, item := range c.items {
		result = append(result, c.nameFunc(item))
	}
	return result
}

func (c *TypedContainer[T]) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func (c *TypedContainer[T]) IsInjected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.injected
}

func (c *TypedContainer[T]) setInjected(injected bool) {
	c.mu.Lock()
	c.injected = injected
	c.mu.Unlock()
}

func (c *TypedContainer[T]) RangeItems(fn func(reflect.Type, T) bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for ifaceType, item := range c.items {
		if !fn(ifaceType, item) {
			break
		}
	}
}

// NamedContainer 命名容器
type NamedContainer[T any] struct {
	mu       sync.RWMutex
	items    map[string]T
	nameFunc func(T) string
}

func NewNamedContainer[T any](nameFunc func(T) string) *NamedContainer[T] {
	return &NamedContainer[T]{
		items:    make(map[string]T),
		nameFunc: nameFunc,
	}
}

func (c *NamedContainer[T]) Register(impl T) error {
	name := c.nameFunc(impl)

	if name == "" {
		return &DuplicateRegistrationError{Name: "empty name"}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[name]; exists {
		return &DuplicateRegistrationError{
			Name:     name,
			Existing: c.items[name],
			New:      impl,
		}
	}

	c.items[name] = impl
	return nil
}

func (c *NamedContainer[T]) GetByName(name string) (T, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	impl, exists := c.items[name]
	if !exists {
		var zero T
		return zero, &InstanceNotFoundError{Name: name, Layer: "Named"}
	}
	return impl, nil
}

func (c *NamedContainer[T]) GetAll() []T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]T, 0, len(c.items))
	for _, item := range c.items {
		result = append(result, item)
	}
	return result
}

func (c *NamedContainer[T]) GetNames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]string, 0, len(c.items))
	for _, item := range c.items {
		result = append(result, c.nameFunc(item))
	}
	return result
}

func (c *NamedContainer[T]) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// injectableContainer 可注入容器的基础实现
type injectableContainer[T any] struct {
	container *TypedContainer[T]
	sources   []ContainerSource
}

func (ic *injectableContainer[T]) buildSources(self ContainerSource, sources ...ContainerSource) []ContainerSource {
	result := []ContainerSource{self}
	result = append(result, sources...)
	return result
}

func (ic *injectableContainer[T]) injectAll(self ContainerSource) error {
	if ic.container.IsInjected() {
		return nil
	}

	resolver := NewGenericDependencyResolver(ic.sources...)

	items := ic.container.GetAll()
	for _, item := range items {
		if err := injectDependencies(item, resolver); err != nil {
			return err
		}
		verifyInjectTags(item)
	}

	ic.container.setInjected(true)
	return nil
}

// getAllSorted 获取所有实例并按名称排序
func getAllSorted[T any](items []T, getName func(T) string) []T {
	result := make([]T, len(items))
	copy(result, items)

	sort.Slice(result, func(i, j int) bool {
		return getName(result[i]) < getName(result[j])
	})

	return result
}

// resolveDependency 从指定容器解析依赖
func resolveDependency[T any](
	fieldType reflect.Type,
	baseType reflect.Type,
	container *TypedContainer[T],
	layerName string,
) (interface{}, error) {
	if fieldType == baseType || fieldType.Implements(baseType) {
		impl := container.GetByType(fieldType)
		if reflect.ValueOf(impl).IsNil() {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: layerName,
			}
		}
		return impl, nil
	}
	return nil, nil
}

// resolveDependencyFromManager 从管理器容器解析依赖
func resolveDependencyFromManager(
	fieldType reflect.Type,
	managerContainer *ManagerContainer,
) (interface{}, error) {
	if managerContainer == nil {
		return nil, nil
	}

	baseManagerType := reflect.TypeOf((*common.IBaseManager)(nil)).Elem()
	if fieldType == baseManagerType || fieldType.Implements(baseManagerType) {
		impl := managerContainer.GetByType(fieldType)
		if impl == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Manager",
			}
		}
		return impl, nil
	}
	return nil, nil
}

// resolveDependencyFromEntity 从实体容器解析依赖
func resolveDependencyFromEntity(
	fieldType reflect.Type,
	entityContainer *EntityContainer,
) (interface{}, error) {
	if entityContainer == nil {
		return nil, nil
	}

	baseEntityType := reflect.TypeOf((*common.IBaseEntity)(nil)).Elem()
	if fieldType == baseEntityType || fieldType.Implements(baseEntityType) {
		items, err := entityContainer.GetByType(fieldType)
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
