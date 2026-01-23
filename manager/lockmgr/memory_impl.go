package lockmgr

import (
	"context"
	"sync"
	"time"
)

// lockEntry 内存锁条目
type lockEntry struct {
	mu         sync.Mutex // 互斥锁
	holder     string     // 锁持有者标识
	expireTime time.Time  // 锁过期时间
}

// lockManagerMemoryImpl 内存锁管理器实现
type lockManagerMemoryImpl struct {
	*lockManagerBaseImpl          // 基础锁管理器实现
	locks                sync.Map // 锁存储映射
	name                 string   // 管理器名称
}

// NewLockManagerMemoryImpl 创建内存锁管理器实例
// @param config 内存锁配置
// @return ILockManager 锁管理器接口
func NewLockManagerMemoryImpl(config *MemoryLockConfig) ILockManager {
	impl := &lockManagerMemoryImpl{
		lockManagerBaseImpl: newLockManagerBaseImpl(),
		locks:               sync.Map{},
		name:                "lockManagerMemoryImpl",
	}
	impl.initObservability()
	return impl
}

// ManagerName 返回管理器名称
// @return string 管理器名称
func (m *lockManagerMemoryImpl) ManagerName() string {
	return m.name
}

// Health 健康检查
// @return error 健康状态，nil表示正常
func (m *lockManagerMemoryImpl) Health() error {
	return nil
}

// OnStart 启动时的初始化操作
// @return error 初始化错误，nil表示成功
func (m *lockManagerMemoryImpl) OnStart() error {
	return nil
}

// OnStop 停止时的清理操作
// @return error 清理错误，nil表示成功
func (m *lockManagerMemoryImpl) OnStop() error {
	return nil
}

// Lock 获取锁，阻塞直到成功或上下文取消
// @param ctx 上下文
// @param key 锁的键
// @param ttl 锁的存活时间，0表示不过期
// @return error 错误信息
func (m *lockManagerMemoryImpl) Lock(ctx context.Context, key string, ttl time.Duration) error {
	return m.recordOperation(ctx, "memory", "lock", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}

		value, _ := m.locks.LoadOrStore(key, &lockEntry{})
		entry := value.(*lockEntry)

		entry.mu.Lock()
		now := time.Now()
		if ttl > 0 {
			entry.expireTime = now.Add(ttl)
		} else {
			entry.expireTime = time.Time{}
		}

		m.recordLockAcquire(ctx, "memory", true)
		return nil
	})
}

// Unlock 释放锁
// @param ctx 上下文
// @param key 锁的键
// @return error 错误信息
func (m *lockManagerMemoryImpl) Unlock(ctx context.Context, key string) error {
	return m.recordOperation(ctx, "memory", "unlock", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}

		value, ok := m.locks.Load(key)
		if !ok {
			return nil
		}

		entry := value.(*lockEntry)
		if entry.mu.TryLock() {
			entry.mu.Unlock()
		} else {
			entry.mu.Unlock()
		}
		m.recordLockRelease(ctx, "memory")
		return nil
	})
}

// TryLock 尝试获取锁，不阻塞
// @param ctx 上下文
// @param key 锁的键
// @param ttl 锁的存活时间，0表示不过期
// @return bool 是否获取成功
// @return error 错误信息
func (m *lockManagerMemoryImpl) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	var result bool
	err := m.recordOperation(ctx, "memory", "trylock", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}

		value, _ := m.locks.LoadOrStore(key, &lockEntry{})
		entry := value.(*lockEntry)

		if entry.mu.TryLock() {
			now := time.Now()
			if ttl > 0 {
				entry.expireTime = now.Add(ttl)
			} else {
				entry.expireTime = time.Time{}
			}
			result = true
			m.recordLockAcquire(ctx, "memory", true)
		} else {
			result = false
			m.recordLockAcquire(ctx, "memory", false)
		}

		return nil
	})

	return result, err
}

var _ ILockManager = (*lockManagerMemoryImpl)(nil)
