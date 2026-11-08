package container

import (
	"fmt"
	"reflect"
	"sync"
)

// InjectableContainer 可注入容器接口
type InjectableContainer interface {
	InjectAll() error
}

// TypedContainer 类型化容器
// 使用接口类型作为键，存储对应的实现实例
type TypedContainer[T any] struct {
	mu       sync.RWMutex
	items    map[reflect.Type]T
	nameFunc func(T) string
	injected bool
}

// NewTypedContainer 创建新的类型化容器
func NewTypedContainer[T any](nameFunc func(T) string) *TypedContainer[T] {
	return &TypedContainer[T]{
		items:    make(map[reflect.Type]T),
		nameFunc: nameFunc,
	}
}

// Register 按接口类型注册实现实例
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

// GetByType 按接口类型获取实现实例
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

// GetAll 获取所有已注册的实例
func (c *TypedContainer[T]) GetAll() []T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]T, 0, len(c.items))
	for _, item := range c.items {
		result = append(result, item)
	}
	return result
}

// GetNames 获取所有实例的名称
func (c *TypedContainer[T]) GetNames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]string, 0, len(c.items))
	for _, item := range c.items {
		result = append(result, c.nameFunc(item))
	}
	return result
}

// Count 返回已注册的实例数量
func (c *TypedContainer[T]) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// IsInjected 返回是否已完成依赖注入
func (c *TypedContainer[T]) IsInjected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.injected
}

// setInjected 设置注入状态
func (c *TypedContainer[T]) setInjected(injected bool) {
	c.mu.Lock()
	c.injected = injected
	c.mu.Unlock()
}

// RangeItems 遍历所有实例
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
// 使用名称字符串作为键，存储对应的实例
type NamedContainer[T any] struct {
	mu       sync.RWMutex
	items    map[string]T
	nameFunc func(T) string
}

// NewNamedContainer 创建新的命名容器
func NewNamedContainer[T any](nameFunc func(T) string) *NamedContainer[T] {
	return &NamedContainer[T]{
		items:    make(map[string]T),
		nameFunc: nameFunc,
	}
}

// Register 注册实例
func (c *NamedContainer[T]) Register(impl T) error {
	name := c.nameFunc(impl)

	if name == "" {
		return fmt.Errorf("instance name cannot be empty")
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

// GetByName 按名称获取实例
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

// GetAll 获取所有已注册的实例
func (c *NamedContainer[T]) GetAll() []T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]T, 0, len(c.items))
	for _, item := range c.items {
		result = append(result, item)
	}
	return result
}

// GetNames 获取所有实例的名称
func (c *NamedContainer[T]) GetNames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]string, 0, len(c.items))
	for _, item := range c.items {
		result = append(result, c.nameFunc(item))
	}
	return result
}

// Count 返回已注册的实例数量
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

// buildSources 构建依赖源列表
func (ic *injectableContainer[T]) buildSources(self ContainerSource, sources ...ContainerSource) []ContainerSource {
	result := []ContainerSource{self}
	result = append(result, sources...)
	return result
}

// injectAll 执行依赖注入
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
