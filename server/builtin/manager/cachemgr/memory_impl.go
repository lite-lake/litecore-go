package cachemgr

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

// cacheManagerMemoryImpl 内存缓存实现
type cacheManagerMemoryImpl struct {
	*cacheManagerBaseImpl
	cache *cache.Cache
	name  string
	mu    sync.RWMutex
}

// NewCacheManagerMemoryImpl 创建内存缓存实现
func NewCacheManagerMemoryImpl(defaultExpiration, cleanupInterval time.Duration) ICacheManager {
	impl := &cacheManagerMemoryImpl{
		cacheManagerBaseImpl: newICacheManagerBaseImpl(),
		cache:                cache.New(defaultExpiration, cleanupInterval),
		name:                 "cacheManagerMemoryImpl",
	}
	impl.initObservability()
	return impl
}

// ManagerName 返回管理器名称
func (m *cacheManagerMemoryImpl) ManagerName() string {
	return m.name
}

// Health 检查管理器健康状态
func (m *cacheManagerMemoryImpl) Health() error {
	return nil
}

// OnStart 在服务器启动时触发
func (m *cacheManagerMemoryImpl) OnStart() error {
	return nil
}

// OnStop 在服务器停止时触发
func (m *cacheManagerMemoryImpl) OnStop() error {
	return nil
}

// Get 获取缓存值
func (m *cacheManagerMemoryImpl) Get(ctx context.Context, key string, dest any) error {
	return m.recordOperation(ctx, m.name, "get", key, func() error {
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

		// 使用反射来支持任意类型的赋值
		destValue := reflect.ValueOf(dest)
		if destValue.Kind() != reflect.Ptr {
			return fmt.Errorf("dest must be a pointer")
		}

		valueValue := reflect.ValueOf(value)
		if !valueValue.IsValid() {
			return fmt.Errorf("cached value is invalid")
		}

		// 如果 value 是指针，获取其指向的值
		if valueValue.Kind() == reflect.Ptr {
			if valueValue.IsNil() {
				return fmt.Errorf("cached value is nil")
			}
			valueValue = valueValue.Elem()
		}

		// 获取 dest 指向的元素
		destElem := destValue.Elem()

		// 检查类型是否匹配
		if !valueValue.Type().AssignableTo(destElem.Type()) {
			return fmt.Errorf("type mismatch: cannot assign %v to %v", valueValue.Type(), destElem.Type())
		}

		// 赋值
		destElem.Set(valueValue)

		return nil
	})
}

// Set 设置缓存值
func (m *cacheManagerMemoryImpl) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return m.recordOperation(ctx, m.name, "set", key, func() error {
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
	})
}

// SetNX 仅当键不存在时才设置值
func (m *cacheManagerMemoryImpl) SetNX(ctx context.Context, key string, value any,
	expiration time.Duration) (bool, error) {
	var result bool

	err := m.recordOperation(ctx, m.name, "setnx", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}

		m.mu.Lock()
		defer m.mu.Unlock()

		// 检查键是否存在
		if _, found := m.cache.Get(key); found {
			result = false
			return nil
		}

		m.cache.Set(key, value, expiration)
		result = true
		return nil
	})

	return result, err
}

// Delete 删除缓存值
func (m *cacheManagerMemoryImpl) Delete(ctx context.Context, key string) error {
	return m.recordOperation(ctx, m.name, "delete", key, func() error {
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
	})
}

// Exists 检查键是否存在
func (m *cacheManagerMemoryImpl) Exists(ctx context.Context, key string) (bool, error) {
	var result bool

	err := m.recordOperation(ctx, m.name, "exists", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}

		m.mu.RLock()
		defer m.mu.RUnlock()

		_, found := m.cache.Get(key)
		result = found
		return nil
	})

	return result, err
}

// Expire 设置过期时间
func (m *cacheManagerMemoryImpl) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return m.recordOperation(ctx, m.name, "expire", key, func() error {
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
	})
}

// TTL 获取剩余过期时间
func (m *cacheManagerMemoryImpl) TTL(ctx context.Context, key string) (time.Duration, error) {
	var result time.Duration

	err := m.recordOperation(ctx, m.name, "ttl", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}

		m.mu.RLock()
		defer m.mu.RUnlock()

		// go-cache 不直接提供 TTL 查询
		// 返回 0 表示未知或永不过期
		// 实际应用中可能需要维护额外的过期时间映射
		result = 0
		return nil
	})

	return result, err
}

// Clear 清空所有缓存
func (m *cacheManagerMemoryImpl) Clear(ctx context.Context) error {
	return m.recordOperation(ctx, m.name, "clear", "", func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}

		m.mu.Lock()
		defer m.mu.Unlock()

		m.cache.Flush()
		return nil
	})
}

// GetMultiple 批量获取
func (m *cacheManagerMemoryImpl) GetMultiple(ctx context.Context, keys []string) (map[string]any, error) {
	var result map[string]any

	key := "batch"
	if len(keys) > 0 {
		key = keys[0]
	}

	err := m.recordOperation(ctx, m.name, "getmultiple", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}

		result = make(map[string]any)

		m.mu.RLock()
		defer m.mu.RUnlock()

		for _, key := range keys {
			if value, found := m.cache.Get(key); found {
				result[key] = value
			}
		}

		return nil
	})

	return result, err
}

// SetMultiple 批量设置
func (m *cacheManagerMemoryImpl) SetMultiple(ctx context.Context, items map[string]any,
	expiration time.Duration) error {
	key := "batch"
	for k := range items {
		key = k
		break
	}

	return m.recordOperation(ctx, m.name, "setmultiple", key, func() error {
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
	})
}

// DeleteMultiple 批量删除
func (m *cacheManagerMemoryImpl) DeleteMultiple(ctx context.Context, keys []string) error {
	key := "batch"
	if len(keys) > 0 {
		key = keys[0]
	}

	return m.recordOperation(ctx, m.name, "deletemultiple", key, func() error {
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
	})
}

// Increment 自增
func (m *cacheManagerMemoryImpl) Increment(ctx context.Context, key string, value int64) (int64, error) {
	var result int64

	err := m.recordOperation(ctx, m.name, "increment", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}

		m.mu.Lock()
		defer m.mu.Unlock()

		// 获取当前值
		var currentValue int64 = 0
		if val, found := m.cache.Get(key); found {
			if num, ok := val.(int64); ok {
				currentValue = num
			} else {
				return fmt.Errorf("value is not an int64")
			}
		}

		// 自增
		newValue := currentValue + value
		m.cache.Set(key, newValue, cache.DefaultExpiration)

		result = newValue
		return nil
	})

	return result, err
}

// Decrement 自减
func (m *cacheManagerMemoryImpl) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	var result int64

	err := m.recordOperation(ctx, m.name, "decrement", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}

		m.mu.Lock()
		defer m.mu.Unlock()

		// 获取当前值
		var currentValue int64 = 0
		if val, found := m.cache.Get(key); found {
			if num, ok := val.(int64); ok {
				currentValue = num
			} else {
				return fmt.Errorf("value is not an int64")
			}
		}

		// 自减
		newValue := currentValue - value
		m.cache.Set(key, newValue, cache.DefaultExpiration)

		result = newValue
		return nil
	})

	return result, err
}

// Close 关闭内存缓存
func (m *cacheManagerMemoryImpl) Close() error {
	return nil
}

// ItemCount 返回缓存项数量
func (m *cacheManagerMemoryImpl) ItemCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cache.ItemCount()
}

// 确保 cacheManagerMemoryImpl 实现 ICacheManager 接口
var _ ICacheManager = (*cacheManagerMemoryImpl)(nil)
