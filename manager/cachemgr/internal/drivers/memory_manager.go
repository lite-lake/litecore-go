package drivers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"com.litelake.litecore/common"

	"github.com/patrickmn/go-cache"
)

// MemoryManager 内存缓存管理器
// 使用 go-cache 库实现本地内存缓存
type MemoryManager struct {
	*BaseManager
	cache *cache.Cache
	mu    sync.RWMutex
}

// NewMemoryManager 创建内存缓存管理器
func NewMemoryManager(defaultExpiration, cleanupInterval time.Duration) *MemoryManager {
	return &MemoryManager{
		BaseManager: NewBaseManager("memory-cache"),
		cache:       cache.New(defaultExpiration, cleanupInterval),
	}
}

// Get 获取缓存值
func (m *MemoryManager) Get(ctx context.Context, key string, dest any) error {
	if err := ValidateContext(ctx); err != nil {
		return err
	}
	if err := ValidateKey(key); err != nil {
		return err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	value, found := m.cache.Get(key)
	if !found {
		return fmt.Errorf("key not found: %s", key)
	}

	// 类型断言和赋值
	switch d := dest.(type) {
	case *any:
		*d = value
	case *string:
		if str, ok := value.(string); ok {
			*d = str
		} else {
			return fmt.Errorf("value is not a string")
		}
	case *int, *int64, *float64, *bool:
		// 基本类型需要通过反射或类型断言处理
		return fmt.Errorf("unsupported destination type %T, use *any instead", dest)
	default:
		// 尝试直接赋值（用于指针类型）
		// 这里简化处理，实际可能需要更复杂的类型转换
		return fmt.Errorf("unsupported destination type %T", dest)
	}

	return nil
}

// GetAny 获取缓存值（返回 any 类型）
func (m *MemoryManager) GetAny(ctx context.Context, key string) (any, bool) {
	if err := ValidateContext(ctx); err != nil {
		return nil, false
	}
	if err := ValidateKey(key); err != nil {
		return nil, false
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.cache.Get(key)
}

// Set 设置缓存值
func (m *MemoryManager) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	if err := ValidateContext(ctx); err != nil {
		return err
	}
	if err := ValidateKey(key); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache.Set(key, value, expiration)
	return nil
}

// SetNX 仅当键不存在时才设置值
func (m *MemoryManager) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	if err := ValidateContext(ctx); err != nil {
		return false, err
	}
	if err := ValidateKey(key); err != nil {
		return false, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查键是否存在
	if _, found := m.cache.Get(key); found {
		return false, nil
	}

	m.cache.Set(key, value, expiration)
	return true, nil
}

// Delete 删除缓存值
func (m *MemoryManager) Delete(ctx context.Context, key string) error {
	if err := ValidateContext(ctx); err != nil {
		return err
	}
	if err := ValidateKey(key); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache.Delete(key)
	return nil
}

// Exists 检查键是否存在
func (m *MemoryManager) Exists(ctx context.Context, key string) (bool, error) {
	if err := ValidateContext(ctx); err != nil {
		return false, err
	}
	if err := ValidateKey(key); err != nil {
		return false, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	_, found := m.cache.Get(key)
	return found, nil
}

// Expire 设置过期时间
func (m *MemoryManager) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if err := ValidateContext(ctx); err != nil {
		return err
	}
	if err := ValidateKey(key); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// go-cache 不直接支持修改过期时间
	// 需要先获取值，再重新设置
	value, found := m.cache.Get(key)
	if !found {
		return fmt.Errorf("key not found: %s", key)
	}

	m.cache.Set(key, value, expiration)
	return nil
}

// TTL 获取剩余过期时间
func (m *MemoryManager) TTL(ctx context.Context, key string) (time.Duration, error) {
	if err := ValidateContext(ctx); err != nil {
		return 0, err
	}
	if err := ValidateKey(key); err != nil {
		return 0, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// go-cache 不直接提供 TTL 查询
	// 返回 0 表示未知或永不过期
	// 实际应用中可能需要维护额外的过期时间映射
	return 0, nil
}

// Clear 清空所有缓存
func (m *MemoryManager) Clear(ctx context.Context) error {
	if err := ValidateContext(ctx); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache.Flush()
	return nil
}

// GetMultiple 批量获取
func (m *MemoryManager) GetMultiple(ctx context.Context, keys []string) (map[string]any, error) {
	if err := ValidateContext(ctx); err != nil {
		return nil, err
	}

	result := make(map[string]any)

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, key := range keys {
		if value, found := m.cache.Get(key); found {
			result[key] = value
		}
	}

	return result, nil
}

// SetMultiple 批量设置
func (m *MemoryManager) SetMultiple(ctx context.Context, items map[string]any, expiration time.Duration) error {
	if err := ValidateContext(ctx); err != nil {
		return err
	}

	if len(items) == 0 {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for key, value := range items {
		m.cache.Set(key, value, expiration)
	}

	return nil
}

// DeleteMultiple 批量删除
func (m *MemoryManager) DeleteMultiple(ctx context.Context, keys []string) error {
	if err := ValidateContext(ctx); err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, key := range keys {
		m.cache.Delete(key)
	}

	return nil
}

// Increment 自增
func (m *MemoryManager) Increment(ctx context.Context, key string, value int64) (int64, error) {
	if err := ValidateContext(ctx); err != nil {
		return 0, err
	}
	if err := ValidateKey(key); err != nil {
		return 0, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 获取当前值
	var currentValue int64 = 0
	if val, found := m.cache.Get(key); found {
		if num, ok := val.(int64); ok {
			currentValue = num
		} else {
			return 0, fmt.Errorf("value is not an int64")
		}
	}

	// 自增
	newValue := currentValue + value
	m.cache.Set(key, newValue, cache.DefaultExpiration)

	return newValue, nil
}

// Decrement 自减
func (m *MemoryManager) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	if err := ValidateContext(ctx); err != nil {
		return 0, err
	}
	if err := ValidateKey(key); err != nil {
		return 0, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 获取当前值
	var currentValue int64 = 0
	if val, found := m.cache.Get(key); found {
		if num, ok := val.(int64); ok {
			currentValue = num
		} else {
			return 0, fmt.Errorf("value is not an int64")
		}
	}

	// 自减
	newValue := currentValue - value
	m.cache.Set(key, newValue, cache.DefaultExpiration)

	return newValue, nil
}

// ItemCount 返回缓存项数量
func (m *MemoryManager) ItemCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cache.ItemCount()
}

// Ensure MemoryManager implements common.BaseManager interface
var _ common.BaseManager = (*MemoryManager)(nil)
