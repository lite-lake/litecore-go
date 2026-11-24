package limitermgr

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewLimiterManagerMemoryImpl(t *testing.T) {
	t.Run("创建内存限流管理器", func(t *testing.T) {
		mgr := NewLimiterManagerMemoryImpl(nil, nil)
		assert.NotNil(t, mgr)
		assert.Equal(t, "limiterManagerMemoryImpl", mgr.ManagerName())
		assert.Nil(t, mgr.Health())
	})
}

func TestLimiterManagerMemoryImpl_Allow(t *testing.T) {
	ctx := context.Background()
	key := "test_key"
	limit := 5
	window := time.Second

	t.Run("限流前允许访问", func(t *testing.T) {
		mgr := NewLimiterManagerMemoryImpl(nil, nil)

		for i := 0; i < limit; i++ {
			allowed, err := mgr.Allow(ctx, key, limit, window)
			assert.NoError(t, err)
			assert.True(t, allowed, "第%d次应该允许", i+1)
		}
	})

	t.Run("超过限制后拒绝访问", func(t *testing.T) {
		mgr := NewLimiterManagerMemoryImpl(nil, nil)
		key := "test_key_2"

		for i := 0; i < limit; i++ {
			mgr.Allow(ctx, key, limit, window)
		}

		allowed, err := mgr.Allow(ctx, key, limit, window)
		assert.NoError(t, err)
		assert.False(t, allowed, "超过限制后应该拒绝")
	})

	t.Run("不同key独立限流", func(t *testing.T) {
		mgr := NewLimiterManagerMemoryImpl(nil, nil)

		for i := 0; i < limit; i++ {
			mgr.Allow(ctx, "key1", limit, window)
		}

		allowed, err := mgr.Allow(ctx, "key2", limit, window)
		assert.NoError(t, err)
		assert.True(t, allowed, "不同key应该独立限流")
	})
}

func TestLimiterManagerMemoryImpl_GetRemaining(t *testing.T) {
	ctx := context.Background()
	key := "test_remaining"
	limit := 5
	window := time.Second

	t.Run("获取初始剩余次数", func(t *testing.T) {
		mgr := NewLimiterManagerMemoryImpl(nil, nil)
		remaining, err := mgr.GetRemaining(ctx, key, limit, window)
		assert.NoError(t, err)
		assert.Equal(t, limit, remaining)
	})

	t.Run("获取剩余次数递减", func(t *testing.T) {
		mgr := NewLimiterManagerMemoryImpl(nil, nil)
		key := "test_remaining_2"

		for i := 0; i < 3; i++ {
			mgr.Allow(ctx, key, limit, window)
		}

		remaining, err := mgr.GetRemaining(ctx, key, limit, window)
		assert.NoError(t, err)
		assert.Equal(t, limit-3, remaining)
	})

	t.Run("剩余次数为0", func(t *testing.T) {
		mgr := NewLimiterManagerMemoryImpl(nil, nil)
		key := "test_remaining_3"

		for i := 0; i < limit; i++ {
			mgr.Allow(ctx, key, limit, window)
		}

		remaining, err := mgr.GetRemaining(ctx, key, limit, window)
		assert.NoError(t, err)
		assert.Equal(t, 0, remaining)
	})
}

func TestLimiterManagerMemoryImpl_SlidingWindow(t *testing.T) {
	ctx := context.Background()
	key := "test_sliding"
	limit := 5
	window := 100 * time.Millisecond

	t.Run("时间窗口滑动后恢复", func(t *testing.T) {
		mgr := NewLimiterManagerMemoryImpl(nil, nil)

		for i := 0; i < limit; i++ {
			allowed, _ := mgr.Allow(ctx, key, limit, window)
			assert.True(t, allowed)
		}

		allowed, _ := mgr.Allow(ctx, key, limit, window)
		assert.False(t, allowed)

		time.Sleep(window + 10*time.Millisecond)

		allowed, err := mgr.Allow(ctx, key, limit, window)
		assert.NoError(t, err)
		assert.True(t, allowed, "窗口过期后应该恢复")
	})

	t.Run("部分时间过期", func(t *testing.T) {
		mgr := NewLimiterManagerMemoryImpl(nil, nil)
		key := "test_sliding_2"

		interval := window / time.Duration(limit+1)
		for i := 0; i < limit; i++ {
			mgr.Allow(ctx, key, limit, window)
			time.Sleep(interval)
		}

		allowed, _ := mgr.Allow(ctx, key, limit, window)
		assert.False(t, allowed)

		time.Sleep(window / 2)

		remaining, err := mgr.GetRemaining(ctx, key, limit, window)
		assert.NoError(t, err)
		assert.Greater(t, remaining, 0, "部分时间过期后应该有剩余次数")
	})
}

func TestLimiterManagerMemoryImpl_Concurrent(t *testing.T) {
	ctx := context.Background()
	key := "test_concurrent"
	limit := 100
	window := time.Second

	t.Run("并发安全测试", func(t *testing.T) {
		mgr := NewLimiterManagerMemoryImpl(nil, nil)
		var wg sync.WaitGroup
		successCount := 0
		var mu sync.Mutex

		for i := 0; i < 200; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				allowed, err := mgr.Allow(ctx, key, limit, window)
				assert.NoError(t, err)
				if allowed {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}()
		}

		wg.Wait()
		assert.Equal(t, limit, successCount, "并发请求应该正好允许limit个")
	})
}

func TestLimiterManagerMemoryImpl_Validation(t *testing.T) {
	mgr := NewLimiterManagerMemoryImpl(nil, nil)
	ctx := context.Background()
	key := "test_validation"
	limit := 5
	window := time.Second

	t.Run("空上下文", func(t *testing.T) {
		_, err := mgr.Allow(nil, key, limit, window)
		assert.Error(t, err)
	})

	t.Run("空key", func(t *testing.T) {
		_, err := mgr.Allow(ctx, "", limit, window)
		assert.Error(t, err)
	})

	t.Run("无效limit", func(t *testing.T) {
		_, err := mgr.Allow(ctx, key, 0, window)
		assert.Error(t, err)
	})

	t.Run("无效window", func(t *testing.T) {
		_, err := mgr.Allow(ctx, key, limit, 0)
		assert.Error(t, err)
	})

	t.Run("负数limit", func(t *testing.T) {
		_, err := mgr.Allow(ctx, key, -1, window)
		assert.Error(t, err)
	})

	t.Run("负数window", func(t *testing.T) {
		_, err := mgr.Allow(ctx, key, limit, -1*time.Second)
		assert.Error(t, err)
	})
}

func TestLimiterManagerMemoryImpl_EdgeConditions(t *testing.T) {
	ctx := context.Background()
	limit := 1
	window := 10 * time.Millisecond

	t.Run("limit为1", func(t *testing.T) {
		mgr := NewLimiterManagerMemoryImpl(nil, nil)
		key := "edge1"

		allowed, err := mgr.Allow(ctx, key, limit, window)
		assert.NoError(t, err)
		assert.True(t, allowed)

		allowed, err = mgr.Allow(ctx, key, limit, window)
		assert.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("非常短的窗口", func(t *testing.T) {
		mgr := NewLimiterManagerMemoryImpl(nil, nil)
		key := "edge2"

		allowed, _ := mgr.Allow(ctx, key, limit, 1*time.Millisecond)
		assert.True(t, allowed)

		time.Sleep(2 * time.Millisecond)

		allowed, err := mgr.Allow(ctx, key, limit, 1*time.Millisecond)
		assert.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("非常长的窗口", func(t *testing.T) {
		mgr := NewLimiterManagerMemoryImpl(nil, nil)
		key := "edge3"

		for i := 0; i < limit; i++ {
			allowed, err := mgr.Allow(ctx, key, limit, time.Hour)
			assert.NoError(t, err)
			assert.True(t, allowed)
		}

		allowed, err := mgr.Allow(ctx, key, limit, time.Hour)
		assert.NoError(t, err)
		assert.False(t, allowed)
	})
}
