package lockmgr

import (
	"context"
	"errors"
	"github.com/lite-lake/litecore-go/manager/cachemgr"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lite-lake/litecore-go/logger"
)

type mockCacheManager struct {
	mu      sync.Mutex
	data    map[string]any
	expires map[string]time.Time
	closed  bool
}

func newMockCacheManager() *mockCacheManager {
	return &mockCacheManager{
		data:    make(map[string]any),
		expires: make(map[string]time.Time),
	}
}

func (m *mockCacheManager) ManagerName() string {
	return "mockCacheManager"
}

func (m *mockCacheManager) Health() error {
	if m.closed {
		return errors.New("cache manager closed")
	}
	return nil
}

func (m *mockCacheManager) OnStart() error {
	return nil
}

func (m *mockCacheManager) OnStop() error {
	m.closed = true
	return nil
}

func (m *mockCacheManager) Get(ctx context.Context, key string, dest any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("cache manager closed")
	}

	if exp, exists := m.expires[key]; exists && time.Now().After(exp) {
		delete(m.data, key)
		delete(m.expires, key)
		return errors.New("key not found")
	}

	_, exists := m.data[key]
	if !exists {
		return errors.New("key not found")
	}

	return nil
}

func (m *mockCacheManager) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("cache manager closed")
	}

	m.data[key] = value
	if expiration > 0 {
		m.expires[key] = time.Now().Add(expiration)
	}
	return nil
}

func (m *mockCacheManager) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return false, errors.New("cache manager closed")
	}

	if exp, exists := m.expires[key]; exists && time.Now().After(exp) {
		delete(m.data, key)
		delete(m.expires, key)
	}

	if _, exists := m.data[key]; exists {
		return false, nil
	}

	m.data[key] = value
	if expiration > 0 {
		m.expires[key] = time.Now().Add(expiration)
	}
	return true, nil
}

func (m *mockCacheManager) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("cache manager closed")
	}

	delete(m.data, key)
	delete(m.expires, key)
	return nil
}

func (m *mockCacheManager) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return false, errors.New("cache manager closed")
	}

	if exp, exists := m.expires[key]; exists && time.Now().After(exp) {
		delete(m.data, key)
		delete(m.expires, key)
		return false, nil
	}

	_, exists := m.data[key]
	return exists, nil
}

func (m *mockCacheManager) Expire(ctx context.Context, key string, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("cache manager closed")
	}

	if _, exists := m.data[key]; exists {
		if expiration > 0 {
			m.expires[key] = time.Now().Add(expiration)
		} else {
			delete(m.expires, key)
		}
	}
	return nil
}

func (m *mockCacheManager) TTL(ctx context.Context, key string) (time.Duration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return 0, errors.New("cache manager closed")
	}

	if exp, exists := m.expires[key]; exists {
		ttl := time.Until(exp)
		if ttl <= 0 {
			delete(m.data, key)
			delete(m.expires, key)
			return 0, nil
		}
		return ttl, nil
	}

	if _, exists := m.data[key]; exists {
		return -1, nil
	}

	return 0, nil
}

func (m *mockCacheManager) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("cache manager closed")
	}

	m.data = make(map[string]any)
	m.expires = make(map[string]time.Time)
	return nil
}

func (m *mockCacheManager) GetMultiple(ctx context.Context, keys []string) (map[string]any, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make(map[string]any)
	for _, key := range keys {
		if exp, exists := m.expires[key]; exists && time.Now().After(exp) {
			continue
		}
		if val, exists := m.data[key]; exists {
			result[key] = val
		}
	}
	return result, nil
}

func (m *mockCacheManager) SetMultiple(ctx context.Context, items map[string]any, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for key, value := range items {
		m.data[key] = value
		if expiration > 0 {
			m.expires[key] = time.Now().Add(expiration)
		}
	}
	return nil
}

func (m *mockCacheManager) DeleteMultiple(ctx context.Context, keys []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, key := range keys {
		delete(m.data, key)
		delete(m.expires, key)
	}
	return nil
}

func (m *mockCacheManager) Increment(ctx context.Context, key string, value int64) (int64, error) {
	return 0, errors.New("not implemented")
}

func (m *mockCacheManager) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	return 0, errors.New("not implemented")
}

func (m *mockCacheManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	return nil
}

type mockLogger struct {
	logger.ILogger
}

func (l *mockLogger) Debug(msg string, keysAndValues ...interface{}) {}

func (l *mockLogger) Info(msg string, keysAndValues ...interface{}) {}

func (l *mockLogger) Warn(msg string, keysAndValues ...interface{}) {}

func (l *mockLogger) Error(msg string, keysAndValues ...interface{}) {}

func (l *mockLogger) Fatal(msg string, keysAndValues ...interface{}) {}

func TestLockManagerRedisImpl_TryLock(t *testing.T) {
	t.Run("TryLock 成功获取锁", func(t *testing.T) {
		mockCache := newMockCacheManager()
		mgr := NewLockManagerRedisImpl(nil, nil, nil, &RedisLockConfig{})
		mgr.(*lockManagerRedisImpl).cacheMgr = mockCache

		ctx := context.Background()
		key := "test-trylock-key1"

		success, err := mgr.TryLock(ctx, key, time.Second*5)
		assert.NoError(t, err)
		assert.True(t, success)

		err = mgr.Unlock(ctx, key)
		assert.NoError(t, err)
	})

	t.Run("TryLock 失败获取锁", func(t *testing.T) {
		mockCache := newMockCacheManager()
		mgr := NewLockManagerRedisImpl(nil, nil, nil, &RedisLockConfig{})
		mgr.(*lockManagerRedisImpl).cacheMgr = mockCache

		ctx := context.Background()
		key := "test-trylock-key2"

		success, err := mgr.TryLock(ctx, key, time.Second*5)
		assert.NoError(t, err)
		assert.True(t, success)

		success2, err := mgr.TryLock(ctx, key, time.Second*5)
		assert.NoError(t, err)
		assert.False(t, success2)

		err = mgr.Unlock(ctx, key)
		assert.NoError(t, err)
	})

	t.Run("TryLock 释放后可再次获取", func(t *testing.T) {
		mockCache := newMockCacheManager()
		mgr := NewLockManagerRedisImpl(nil, nil, nil, &RedisLockConfig{})
		mgr.(*lockManagerRedisImpl).cacheMgr = mockCache

		ctx := context.Background()
		key := "test-trylock-key3"

		success, err := mgr.TryLock(ctx, key, time.Second*5)
		assert.NoError(t, err)
		assert.True(t, success)

		err = mgr.Unlock(ctx, key)
		assert.NoError(t, err)

		success2, err := mgr.TryLock(ctx, key, time.Second*5)
		assert.NoError(t, err)
		assert.True(t, success2)

		err = mgr.Unlock(ctx, key)
		assert.NoError(t, err)
	})
}

func TestLockManagerRedisImpl_LockUnlock(t *testing.T) {
	t.Run("基础锁获取和释放", func(t *testing.T) {
		mockCache := newMockCacheManager()
		mgr := NewLockManagerRedisImpl(nil, nil, nil, &RedisLockConfig{})
		mgr.(*lockManagerRedisImpl).cacheMgr = mockCache

		ctx := context.Background()
		key := "test-key"

		err := mgr.Lock(ctx, key, time.Second*5)
		assert.NoError(t, err)

		err = mgr.Unlock(ctx, key)
		assert.NoError(t, err)
	})

	t.Run("重复释放锁", func(t *testing.T) {
		mockCache := newMockCacheManager()
		mgr := NewLockManagerRedisImpl(nil, nil, nil, &RedisLockConfig{})
		mgr.(*lockManagerRedisImpl).cacheMgr = mockCache

		ctx := context.Background()
		key := "test-key2"

		err := mgr.Lock(ctx, key, time.Second*5)
		assert.NoError(t, err)

		err = mgr.Unlock(ctx, key)
		assert.NoError(t, err)

		err = mgr.Unlock(ctx, key)
		assert.NoError(t, err)
	})
}

func TestLockManagerRedisImpl_LockTimeout(t *testing.T) {
	t.Run("Lock 超时返回错误", func(t *testing.T) {
		mockCache := newMockCacheManager()
		mgr := NewLockManagerRedisImpl(nil, nil, nil, &RedisLockConfig{})
		mgr.(*lockManagerRedisImpl).cacheMgr = mockCache

		ctx := context.Background()
		key := "test-timeout-key"

		success, err := mgr.TryLock(ctx, key, time.Second*5)
		assert.NoError(t, err)
		assert.True(t, success)

		timeoutCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		err = mgr.Lock(timeoutCtx, key, time.Second*5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "lock acquisition canceled")

		err = mgr.Unlock(ctx, key)
		assert.NoError(t, err)
	})
}

func TestLockManagerRedisImpl_ManagerName(t *testing.T) {
	mgr := NewLockManagerRedisImpl(nil, nil, nil, &RedisLockConfig{})
	assert.Equal(t, "lockManagerRedisImpl", mgr.ManagerName())
}

func TestLockManagerRedisImpl_Health(t *testing.T) {
	t.Run("Health 正常", func(t *testing.T) {
		mockCache := newMockCacheManager()
		mgr := NewLockManagerRedisImpl(nil, nil, nil, &RedisLockConfig{})
		mgr.(*lockManagerRedisImpl).cacheMgr = mockCache

		err := mgr.Health()
		assert.NoError(t, err)
	})

	t.Run("Health 无 cache manager", func(t *testing.T) {
		mgr := NewLockManagerRedisImpl(nil, nil, nil, &RedisLockConfig{})

		err := mgr.Health()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache manager not injected")
	})
}

func TestLockManagerRedisImpl_OnStart(t *testing.T) {
	t.Run("OnStart 正常", func(t *testing.T) {
		mockCache := newMockCacheManager()
		mgr := NewLockManagerRedisImpl(nil, nil, nil, &RedisLockConfig{})
		mgr.(*lockManagerRedisImpl).cacheMgr = mockCache

		err := mgr.OnStart()
		assert.NoError(t, err)
	})

	t.Run("OnStart 无 cache manager", func(t *testing.T) {
		mgr := NewLockManagerRedisImpl(nil, nil, nil, &RedisLockConfig{})

		err := mgr.OnStart()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache manager not injected")
	})
}

func TestLockManagerRedisImpl_OnStop(t *testing.T) {
	t.Run("OnStop 正常", func(t *testing.T) {
		mockCache := newMockCacheManager()
		mgr := NewLockManagerRedisImpl(nil, nil, nil, &RedisLockConfig{})
		mgr.(*lockManagerRedisImpl).cacheMgr = mockCache

		err := mgr.OnStop()
		assert.NoError(t, err)
	})

	t.Run("OnStop 无 cache manager", func(t *testing.T) {
		mgr := NewLockManagerRedisImpl(nil, nil, nil, &RedisLockConfig{})

		err := mgr.OnStop()
		assert.NoError(t, err)
	})
}

func TestLockManagerRedisImpl_Validation(t *testing.T) {
	mockCache := newMockCacheManager()
	mgr := NewLockManagerRedisImpl(nil, nil, nil, &RedisLockConfig{})
	mgr.(*lockManagerRedisImpl).cacheMgr = mockCache

	t.Run("空上下文错误", func(t *testing.T) {
		err := mgr.Lock(nil, "test", time.Second*5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")
	})

	t.Run("空键错误", func(t *testing.T) {
		ctx := context.Background()
		err := mgr.Lock(ctx, "", time.Second*5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "lock key cannot be empty")
	})

	t.Run("TryLock 空上下文错误", func(t *testing.T) {
		_, err := mgr.TryLock(nil, "test", time.Second*5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")
	})

	t.Run("TryLock 空键错误", func(t *testing.T) {
		ctx := context.Background()
		_, err := mgr.TryLock(ctx, "", time.Second*5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "lock key cannot be empty")
	})

	t.Run("Unlock 空上下文错误", func(t *testing.T) {
		err := mgr.Unlock(nil, "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")
	})

	t.Run("Unlock 空键错误", func(t *testing.T) {
		ctx := context.Background()
		err := mgr.Unlock(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "lock key cannot be empty")
	})
}

func TestLockManagerRedisImpl_ZeroTTL(t *testing.T) {
	mockCache := newMockCacheManager()
	mgr := NewLockManagerRedisImpl(nil, nil, nil, &RedisLockConfig{})
	mgr.(*lockManagerRedisImpl).cacheMgr = mockCache

	ctx := context.Background()
	key := "zero-ttl-key"

	err := mgr.Lock(ctx, key, 0)
	assert.NoError(t, err)

	err = mgr.Unlock(ctx, key)
	assert.NoError(t, err)

	success, err := mgr.TryLock(ctx, key, 0)
	assert.NoError(t, err)
	assert.True(t, success)

	err = mgr.Unlock(ctx, key)
	assert.NoError(t, err)
}

func TestLockManagerRedisImpl_NoCacheManager(t *testing.T) {
	mgr := NewLockManagerRedisImpl(nil, nil, nil, &RedisLockConfig{})
	ctx := context.Background()
	key := "test-key"

	t.Run("Lock 无 cache manager", func(t *testing.T) {
		err := mgr.Lock(ctx, key, time.Second*5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache manager not injected")
	})

	t.Run("TryLock 无 cache manager", func(t *testing.T) {
		_, err := mgr.TryLock(ctx, key, time.Second*5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache manager not injected")
	})

	t.Run("Unlock 无 cache manager", func(t *testing.T) {
		err := mgr.Unlock(ctx, key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache manager not injected")
	})
}

var _ cachemgr.ICacheManager = (*mockCacheManager)(nil)
