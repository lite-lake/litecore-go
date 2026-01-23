package lockmgr

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLockManagerMemoryImpl_LockUnlock(t *testing.T) {
	t.Run("基础锁获取和释放", func(t *testing.T) {
		mgr := NewLockManagerMemoryImpl(&MemoryLockConfig{})
		ctx := context.Background()
		key := "test-key"

		err := mgr.Lock(ctx, key, time.Second*5)
		assert.NoError(t, err)

		err = mgr.Unlock(ctx, key)
		assert.NoError(t, err)
	})

	t.Run("重复释放锁", func(t *testing.T) {
		mgr := NewLockManagerMemoryImpl(&MemoryLockConfig{})
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

func TestLockManagerMemoryImpl_TryLock(t *testing.T) {
	t.Run("TryLock 成功获取锁", func(t *testing.T) {
		mgr := NewLockManagerMemoryImpl(&MemoryLockConfig{})
		ctx := context.Background()
		key := "test-trylock-key1"

		success, err := mgr.TryLock(ctx, key, time.Second*5)
		assert.NoError(t, err)
		assert.True(t, success)

		err = mgr.Unlock(ctx, key)
		assert.NoError(t, err)
	})

	t.Run("TryLock 失败获取锁", func(t *testing.T) {
		mgr := NewLockManagerMemoryImpl(&MemoryLockConfig{})
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
		mgr := NewLockManagerMemoryImpl(&MemoryLockConfig{})
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

func TestLockManagerMemoryImpl_Concurrent(t *testing.T) {
	t.Run("并发获取不同键的锁", func(t *testing.T) {
		mgr := NewLockManagerMemoryImpl(&MemoryLockConfig{})
		ctx := context.Background()
		keys := []string{"key1", "key2", "key3", "key4", "key5"}

		var wg sync.WaitGroup
		for _, key := range keys {
			wg.Add(1)
			go func(k string) {
				defer wg.Done()
				err := mgr.Lock(ctx, k, time.Second*5)
				assert.NoError(t, err)
				time.Sleep(time.Millisecond * 100)
				err = mgr.Unlock(ctx, k)
				assert.NoError(t, err)
			}(key)
		}

		wg.Wait()
	})

	t.Run("并发获取同一键的锁互斥", func(t *testing.T) {
		mgr := NewLockManagerMemoryImpl(&MemoryLockConfig{})
		ctx := context.Background()
		key := "concurrent-key"

		counter := 0
		var wg sync.WaitGroup
		mu := sync.Mutex{}

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				err := mgr.Lock(ctx, key, time.Second*5)
				assert.NoError(t, err)

				mu.Lock()
				counter++
				current := counter
				mu.Unlock()

				assert.Equal(t, 1, current, "counter should be 1 during critical section")

				time.Sleep(time.Millisecond * 10)

				mu.Lock()
				counter--
				mu.Unlock()

				err = mgr.Unlock(ctx, key)
				assert.NoError(t, err)
			}(i)
		}

		wg.Wait()
		assert.Equal(t, 0, counter)
	})
}

func TestLockManagerMemoryImpl_ManagerName(t *testing.T) {
	mgr := NewLockManagerMemoryImpl(&MemoryLockConfig{})
	assert.Equal(t, "lockManagerMemoryImpl", mgr.ManagerName())
}

func TestLockManagerMemoryImpl_Health(t *testing.T) {
	mgr := NewLockManagerMemoryImpl(&MemoryLockConfig{})
	assert.NoError(t, mgr.Health())
}

func TestLockManagerMemoryImpl_OnStart(t *testing.T) {
	mgr := NewLockManagerMemoryImpl(&MemoryLockConfig{})
	assert.NoError(t, mgr.OnStart())
}

func TestLockManagerMemoryImpl_OnStop(t *testing.T) {
	mgr := NewLockManagerMemoryImpl(&MemoryLockConfig{})
	assert.NoError(t, mgr.OnStop())
}

func TestLockManagerMemoryImpl_Validation(t *testing.T) {
	mgr := NewLockManagerMemoryImpl(&MemoryLockConfig{})

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

func TestLockManagerMemoryImpl_ZeroTTL(t *testing.T) {
	mgr := NewLockManagerMemoryImpl(&MemoryLockConfig{})
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
