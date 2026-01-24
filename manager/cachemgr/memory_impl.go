package cachemgr

import (
	"context"
	"fmt"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/manager/telemetrymgr"
)

// cacheManagerMemoryImpl 内存缓存实现
// 基于 Ristretto 库实现的高性能内存缓存
type cacheManagerMemoryImpl struct {
	*cacheManagerBaseImpl
	// cache Ristretto 缓存实例
	cache *ristretto.Cache[string, any]
	// name 管理器名称
	name string
	// itemCount 缓存项数量计数器（原子操作）
	itemCount atomic.Int64
}

// NewCacheManagerMemoryImpl 创建内存缓存实现
// 参数：
//   - defaultExpiration: 默认过期时间（仅用于配置参考）
//   - cleanupInterval: 清理间隔（仅用于配置参考）
//   - loggerMgr: 日志管理器（可选）
//   - telemetryMgr: 遥测管理器（可选）
//
// 返回 ICacheManager 接口实例
func NewCacheManagerMemoryImpl(
	defaultExpiration, cleanupInterval time.Duration,
	loggerMgr loggermgr.ILoggerManager,
	telemetryMgr telemetrymgr.ITelemetryManager,
) ICacheManager {
	// 配置 Ristretto 缓存参数
	numCounters := int64(1e6) // 统计计数器数量
	maxCost := int64(1e8)     // 最大缓存成本
	bufferItems := int64(64)  // 缓冲区大小

	// 创建 Ristretto 缓存实例
	cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters:            numCounters,
		MaxCost:                maxCost,
		BufferItems:            bufferItems,
		TtlTickerDurationInSec: 1, // TTL 检查间隔（秒）
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create ristretto cache: %v", err))
	}

	impl := &cacheManagerMemoryImpl{
		cacheManagerBaseImpl: newICacheManagerBaseImpl(loggerMgr, telemetryMgr),
		cache:                cache,
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
// 内存缓存总是健康的
func (m *cacheManagerMemoryImpl) Health() error {
	return nil
}

// OnStart 在服务器启动时触发
// 内存缓存无需额外初始化
func (m *cacheManagerMemoryImpl) OnStart() error {
	return nil
}

// OnStop 在服务器停止时触发
// 内存缓存无需额外清理
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

		// 检查键是否已存在
		_, existed := m.cache.Get(key)

		success := m.cache.SetWithTTL(key, value, 1, expiration)
		if success {
			m.cache.Wait()
			// 只有键不存在时才增加计数
			if !existed {
				m.itemCount.Add(1)
			}
		}
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

		// 检查键是否存在
		if _, found := m.cache.Get(key); found {
			result = false
			return nil
		}

		success := m.cache.SetWithTTL(key, value, 1, expiration)
		if success {
			m.cache.Wait()
			result = true
			m.itemCount.Add(1)
		} else {
			result = false
		}
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

		// 检查键是否存在
		if _, found := m.cache.Get(key); found {
			m.cache.Del(key)
			m.itemCount.Add(-1)
		}
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

		// Ristretto 需要先获取值，再重新设置
		value, found := m.cache.Get(key)
		if !found {
			return fmt.Errorf("key not found: %s", key)
		}

		m.cache.SetWithTTL(key, value, 1, expiration)
		m.cache.Wait()
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

		ttl, found := m.cache.GetTTL(key)
		if !found {
			return fmt.Errorf("key not found: %s", key)
		}
		result = ttl
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

		m.cache.Clear()
		m.itemCount.Store(0)
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

		for key, value := range items {
			m.cache.SetWithTTL(key, value, 1, expiration)
		}
		m.cache.Wait()

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

		for _, key := range keys {
			if _, found := m.cache.Get(key); found {
				m.cache.Del(key)
				m.itemCount.Add(-1)
			}
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
		m.cache.SetWithTTL(key, newValue, 1, 0)
		m.cache.Wait()

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
		m.cache.SetWithTTL(key, newValue, 1, 0)
		m.cache.Wait()

		result = newValue
		return nil
	})

	return result, err
}

// Close 关闭内存缓存
// 释放 Ristretto 缓存资源
func (m *cacheManagerMemoryImpl) Close() error {
	m.cache.Close()
	return nil
}

// ItemCount 返回缓存项数量
// 使用原子操作确保并发安全
func (m *cacheManagerMemoryImpl) ItemCount() int {
	return int(m.itemCount.Load())
}

// 确保 cacheManagerMemoryImpl 实现 ICacheManager 接口
var _ ICacheManager = (*cacheManagerMemoryImpl)(nil)
