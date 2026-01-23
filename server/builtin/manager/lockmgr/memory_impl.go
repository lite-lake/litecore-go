package lockmgr

import (
	"context"
	"sync"
	"time"
)

type lockEntry struct {
	mu         sync.Mutex
	holder     string
	expireTime time.Time
}

type lockManagerMemoryImpl struct {
	*lockManagerBaseImpl
	locks sync.Map
	name  string
}

func NewLockManagerMemoryImpl(config *MemoryLockConfig) ILockManager {
	impl := &lockManagerMemoryImpl{
		lockManagerBaseImpl: newLockManagerBaseImpl(),
		locks:               sync.Map{},
		name:                "lockManagerMemoryImpl",
	}
	impl.initObservability()
	return impl
}

func (m *lockManagerMemoryImpl) ManagerName() string {
	return m.name
}

func (m *lockManagerMemoryImpl) Health() error {
	return nil
}

func (m *lockManagerMemoryImpl) OnStart() error {
	return nil
}

func (m *lockManagerMemoryImpl) OnStop() error {
	return nil
}

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
