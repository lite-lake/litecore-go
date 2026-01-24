package container

import (
	"reflect"
	"sort"
)

// InjectableLayerContainer 可注入层容器基类
// 为 Controller、Middleware 和 Listener 容器提供公共实现
type InjectableLayerContainer[T any] struct {
	base             *injectableContainer[T]
	managerContainer *ManagerContainer
	nameFunc         func(T) string
}

// NewInjectableLayerContainer 创建新的可注入层容器
func NewInjectableLayerContainer[T any](nameFunc func(T) string) *InjectableLayerContainer[T] {
	return &InjectableLayerContainer[T]{
		base: &injectableContainer[T]{
			container: NewTypedContainer(nameFunc),
		},
		nameFunc: nameFunc,
	}
}

// SetManagerContainer 设置管理器容器
func (c *InjectableLayerContainer[T]) SetManagerContainer(container *ManagerContainer) {
	c.managerContainer = container
}

// GetAll 获取所有实例
func (c *InjectableLayerContainer[T]) GetAll() []T {
	return c.base.container.GetAll()
}

// GetAllSorted 获取所有实例并按名称排序
func (c *InjectableLayerContainer[T]) GetAllSorted() []T {
	items := c.GetAll()
	sort.Slice(items, func(i, j int) bool {
		return c.nameFunc(items[i]) < c.nameFunc(items[j])
	})
	return items
}

// GetByType 按类型获取实例
func (c *InjectableLayerContainer[T]) GetByType(ifaceType reflect.Type) T {
	return c.base.container.GetByType(ifaceType)
}

// Count 获取实例数量
func (c *InjectableLayerContainer[T]) Count() int {
	return c.base.container.Count()
}

// RegisterByType 按类型注册实例
func (c *InjectableLayerContainer[T]) RegisterByType(ifaceType reflect.Type, impl T) error {
	return c.base.container.Register(ifaceType, impl)
}

// checkManagerContainer 检查 ManagerContainer 是否已设置
func (c *InjectableLayerContainer[T]) checkManagerContainer(layerName string) {
	if c.managerContainer == nil {
		panic(&ManagerContainerNotSetError{Layer: layerName})
	}
}
