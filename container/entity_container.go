package container

import (
	"reflect"
	"sync"

	"com.litelake.litecore/common"
)

// EntityContainer 实体层容器
// Entity 层无依赖，InjectAll 为空操作
type EntityContainer struct {
	mu    sync.RWMutex
	items map[string]common.BaseEntity
}

// NewEntityContainer 创建新的实体容器
func NewEntityContainer() *EntityContainer {
	return &EntityContainer{
		items: make(map[string]common.BaseEntity),
	}
}

// RegisterEntity 泛型注册函数，注册实体实例
func RegisterEntity[T common.BaseEntity](e *EntityContainer, impl T) error {
	return e.Register(impl)
}

// Register 注册实体实例
func (e *EntityContainer) Register(ins common.BaseEntity) error {
	if ins == nil {
		return &DuplicateRegistrationError{Name: "nil"}
	}

	name := ins.EntityName()

	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.items[name]; exists {
		return &DuplicateRegistrationError{
			Name:     name,
			Existing: e.items[name],
			New:      ins,
		}
	}

	e.items[name] = ins
	return nil
}

// InjectAll 注入所有依赖
// Entity 层无依赖，此方法为空操作
func (e *EntityContainer) InjectAll() error {
	// Entity 层无依赖，无需注入
	return nil
}

// GetAll 获取所有已注册的实体
func (e *EntityContainer) GetAll() []common.BaseEntity {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]common.BaseEntity, 0, len(e.items))
	for _, item := range e.items {
		result = append(result, item)
	}
	return result
}

// GetByName 根据名称获取实体
func (e *EntityContainer) GetByName(name string) (common.BaseEntity, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if ins, exists := e.items[name]; exists {
		return ins, nil
	}
	return nil, &InstanceNotFoundError{Name: name, Layer: "Entity"}
}

// GetByType 根据类型获取实体
// 返回所有实现了该类型的实体列表
func (e *EntityContainer) GetByType(typ reflect.Type) ([]common.BaseEntity, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []common.BaseEntity
	for _, item := range e.items {
		itemType := reflect.TypeOf(item)

		if itemType == typ {
			result = append(result, item)
			continue
		}

		if typ.Kind() == reflect.Interface && itemType.Implements(typ) {
			result = append(result, item)
			continue
		}

		if itemType.Kind() == reflect.Ptr {
			elemType := itemType.Elem()
			if elemType == typ {
				result = append(result, item)
				continue
			}
			if typ.Kind() == reflect.Interface && elemType.Implements(typ) {
				result = append(result, item)
			}
		}
	}
	return result, nil
}

// Count 返回已注册的实体数量
func (e *EntityContainer) Count() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.items)
}
