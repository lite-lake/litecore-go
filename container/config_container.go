package container

import (
	"reflect"
	"sync"

	"com.litelake.litecore/common"
)

// ConfigContainer 配置层容器
// Config 层无依赖，InjectAll 为空操作
type ConfigContainer struct {
	mu    sync.RWMutex
	items map[string]common.BaseConfigProvider
}

// NewConfigContainer 创建新的配置容器
func NewConfigContainer() *ConfigContainer {
	return &ConfigContainer{
		items: make(map[string]common.BaseConfigProvider),
	}
}

// Register 注册配置提供者实例
func (c *ConfigContainer) Register(ins common.BaseConfigProvider) error {
	if ins == nil {
		return &DuplicateRegistrationError{Name: "nil"}
	}

	name := ins.ConfigProviderName()

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[name]; exists {
		return &DuplicateRegistrationError{
			Name:     name,
			Existing: c.items[name],
			New:      ins,
		}
	}

	c.items[name] = ins
	return nil
}

// InjectAll 注入所有依赖
// Config 层无依赖，此方法为空操作
func (c *ConfigContainer) InjectAll() error {
	// Config 层无依赖，无需注入
	return nil
}

// GetAll 获取所有已注册的配置提供者
func (c *ConfigContainer) GetAll() []common.BaseConfigProvider {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]common.BaseConfigProvider, 0, len(c.items))
	for _, item := range c.items {
		result = append(result, item)
	}
	return result
}

// GetByName 根据名称获取配置提供者
func (c *ConfigContainer) GetByName(name string) (common.BaseConfigProvider, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if ins, exists := c.items[name]; exists {
		return ins, nil
	}
	return nil, &InstanceNotFoundError{Name: name, Layer: "Config"}
}

// GetByType 根据类型获取配置提供者
// 返回所有实现了该类型的配置提供者列表
func (c *ConfigContainer) GetByType(typ reflect.Type) ([]common.BaseConfigProvider, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []common.BaseConfigProvider
	for _, item := range c.items {
		itemType := reflect.TypeOf(item)
		if typeMatches(itemType, typ) {
			result = append(result, item)
		}
	}
	return result, nil
}

// Count 返回已注册的配置提供者数量
func (c *ConfigContainer) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}
