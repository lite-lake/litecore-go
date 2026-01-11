package hooks

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
)

// Manager 钩子管理器
type Manager struct {
	mu        sync.RWMutex
	callbacks map[string][]interface{}
}

// NewManager 创建钩子管理器
func NewManager() *Manager {
	return &Manager{
		callbacks: make(map[string][]interface{}),
	}
}

// Register 注册钩子
// name: 钩子名称 (beforeCreate, afterCreate, beforeUpdate, afterUpdate, beforeDelete, afterDelete, afterFind)
// hook: 钩子函数，签名参考 GORM 钩子
func (m *Manager) Register(name string, hook interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 验证钩子名称
	validHooks := map[string]bool{
		"beforeCreate": true,
		"afterCreate":  true,
		"beforeUpdate": true,
		"afterUpdate":  true,
		"beforeDelete": true,
		"afterDelete":  true,
		"afterFind":    true,
	}

	if !validHooks[name] {
		return fmt.Errorf("invalid hook name: %s", name)
	}

	if m.callbacks[name] == nil {
		m.callbacks[name] = []interface{}{}
	}

	m.callbacks[name] = append(m.callbacks[name], hook)
	return nil
}

// Get 获取指定名称的所有钩子
func (m *Manager) Get(name string) []interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if hooks, ok := m.callbacks[name]; ok {
		return hooks
	}

	return []interface{}{}
}

// ApplyTo 应用钩子到 GORM
func (m *Manager) ApplyTo(db *gorm.DB) {
	// 应用钩子逻辑
	// 注意：这个方法需要在模型级别实现钩子
	// 这里只是管理钩子的注册，实际应用需要在模型的钩子方法中调用
}

// Clear 清除所有钩子
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callbacks = make(map[string][]interface{})
}

// Remove 移除指定名称的所有钩子
func (m *Manager) Remove(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.callbacks, name)
}
