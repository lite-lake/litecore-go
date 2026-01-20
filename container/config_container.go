package container

import (
	"reflect"
	"sort"
	"sync"

	"com.litelake.litecore/common"
)

// ConfigContainer 配置层容器
// Config 层无依赖，InjectAll 为空操作
type ConfigContainer struct {
	mu    sync.RWMutex
	items map[reflect.Type]common.IBaseConfigProvider
}

// NewConfigContainer 创建新的配置容器
func NewConfigContainer() *ConfigContainer {
	return &ConfigContainer{
		items: make(map[reflect.Type]common.IBaseConfigProvider),
	}
}

// RegisterConfig 泛型注册函数，按接口类型注册
func RegisterConfig[T common.IBaseConfigProvider](c *ConfigContainer, impl T) error {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	return c.RegisterByType(ifaceType, impl)
}

// RegisterByType 按接口类型注册
func (c *ConfigContainer) RegisterByType(ifaceType reflect.Type, impl common.IBaseConfigProvider) error {
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

// InjectAll 注入所有依赖
// Config 层无依赖，此方法为空操作
func (c *ConfigContainer) InjectAll() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for ifaceType, impl := range c.items {
		if impl == nil {
			return &ImplementationDoesNotImplementInterfaceError{
				InterfaceType:  ifaceType,
				Implementation: nil,
			}
		}
	}

	return nil
}

// GetAll 获取所有已注册的配置提供者
func (c *ConfigContainer) GetAll() []common.IBaseConfigProvider {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]common.IBaseConfigProvider, 0, len(c.items))
	for _, item := range c.items {
		result = append(result, item)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ConfigProviderName() < result[j].ConfigProviderName()
	})

	return result
}

// GetByType 按接口类型获取（返回单例）
func (c *ConfigContainer) GetByType(ifaceType reflect.Type) common.IBaseConfigProvider {
	c.mu.RLock()
	defer c.mu.RUnlock()

	impl, exists := c.items[ifaceType]
	if !exists {
		return nil
	}
	return impl
}

// Count 返回已注册的配置提供者数量
func (c *ConfigContainer) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// GetDependency 根据类型获取依赖实例（实现ContainerSource接口）
func (c *ConfigContainer) GetDependency(fieldType reflect.Type) (interface{}, error) {
	baseConfigType := reflect.TypeOf((*common.IBaseConfigProvider)(nil)).Elem()
	if fieldType == baseConfigType || fieldType.Implements(baseConfigType) {
		impl := c.GetByType(fieldType)
		if impl == nil {
			return nil, &DependencyNotFoundError{
				FieldType:     fieldType,
				ContainerType: "Config",
			}
		}
		return impl, nil
	}
	return nil, nil
}
