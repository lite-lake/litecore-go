package limitermgr

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/lite-lake/litecore-go/manager/cachemgr"
	"github.com/stretchr/testify/assert"
)

// newTestLimiterManagerRedis 创建带内存缓存的 Redis 限流管理器测试实例
func newTestLimiterManagerRedis() *limiterManagerRedisImpl {
	cacheMgr := cachemgr.NewCacheManagerMemoryImpl(time.Minute, time.Minute, nil, nil)
	return NewLimiterManagerRedisImpl(nil, nil, cacheMgr).(*limiterManagerRedisImpl)
}

func TestNewLimiterManagerRedisImpl(t *testing.T) {
	t.Run("创建Redis限流管理器", func(t *testing.T) {
		mgr := NewLimiterManagerRedisImpl(nil, nil, nil)
		assert.NotNil(t, mgr)
		assert.Equal(t, "limiterManagerRedisImpl", mgr.ManagerName())
	})
}

func TestLimiterManagerRedisImpl_Health(t *testing.T) {
	t.Run("未初始化cache manager时健康检查失败", func(t *testing.T) {
		mgr := NewLimiterManagerRedisImpl(nil, nil, nil)
		err := mgr.Health()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache manager is not initialized")
	})

	t.Run("初始化cache manager后健康检查通过", func(t *testing.T) {
		mgr := newTestLimiterManagerRedis()

		err := mgr.Health()
		assert.NoError(t, err)
	})
}

func TestLimiterManagerRedisImpl_Allow(t *testing.T) {
	ctx := context.Background()
	key := "test_key"
	limit := 5
	window := time.Second

	t.Run("限流前允许访问", func(t *testing.T) {
		mgr := newTestLimiterManagerRedis()

		for i := 0; i < limit; i++ {
			allowed, err := mgr.Allow(ctx, key, limit, window)
			assert.NoError(t, err)
			assert.True(t, allowed, "第%d次应该允许", i+1)
		}
	})

	t.Run("超过限制后拒绝访问", func(t *testing.T) {
		mgr := newTestLimiterManagerRedis()
		key := "test_key_2"

		for i := 0; i < limit; i++ {
			mgr.Allow(ctx, key, limit, window)
		}

		allowed, err := mgr.Allow(ctx, key, limit, window)
		assert.NoError(t, err)
		assert.False(t, allowed, "超过限制后应该拒绝")
	})

	t.Run("不同key独立限流", func(t *testing.T) {
		mgr := newTestLimiterManagerRedis()

		for i := 0; i < limit; i++ {
			mgr.Allow(ctx, "key1", limit, window)
		}

		allowed, err := mgr.Allow(ctx, "key2", limit, window)
		assert.NoError(t, err)
		assert.True(t, allowed, "不同key应该独立限流")
	})

	t.Run("未初始化cache manager时返回错误", func(t *testing.T) {
		mgr := NewLimiterManagerRedisImpl(nil, nil, nil)

		_, err := mgr.Allow(ctx, key, limit, window)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache manager is not initialized")
	})
}

func TestLimiterManagerRedisImpl_GetRemaining(t *testing.T) {
	ctx := context.Background()
	key := "test_remaining"
	limit := 5
	window := time.Second

	t.Run("获取初始剩余次数", func(t *testing.T) {
		mgr := newTestLimiterManagerRedis()

		remaining, err := mgr.GetRemaining(ctx, key, limit, window)
		assert.NoError(t, err)
		assert.Equal(t, limit, remaining)
	})

	t.Run("获取剩余次数递减", func(t *testing.T) {
		mgr := newTestLimiterManagerRedis()
		key := "test_remaining_2"

		for i := 0; i < 3; i++ {
			mgr.Allow(ctx, key, limit, window)
		}

		remaining, err := mgr.GetRemaining(ctx, key, limit, window)
		assert.NoError(t, err)
		assert.Equal(t, limit-3, remaining)
	})

	t.Run("剩余次数为0", func(t *testing.T) {
		mgr := newTestLimiterManagerRedis()
		key := "test_remaining_3"

		for i := 0; i < limit; i++ {
			mgr.Allow(ctx, key, limit, window)
		}

		remaining, err := mgr.GetRemaining(ctx, key, limit, window)
		assert.NoError(t, err)
		assert.Equal(t, 0, remaining)
	})

	t.Run("未初始化cache manager时返回错误", func(t *testing.T) {
		mgr := NewLimiterManagerRedisImpl(nil, nil, nil)

		_, err := mgr.GetRemaining(ctx, key, limit, window)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache manager is not initialized")
	})
}

func TestLimiterManagerRedisImpl_SlidingWindow(t *testing.T) {
	ctx := context.Background()
	key := "test_sliding"
	limit := 5
	window := 2 * time.Second

	t.Run("时间窗口滑动后恢复", func(t *testing.T) {
		t.Skip("内存缓存 TTL 行为与 Redis 不同，需要在真实 Redis 环境中测试")
		mgr := newTestLimiterManagerRedis()

		for i := 0; i < limit; i++ {
			allowed, _ := mgr.Allow(ctx, key, limit, window)
			assert.True(t, allowed)
		}

		allowed, _ := mgr.Allow(ctx, key, limit, window)
		assert.False(t, allowed)

		time.Sleep(window + 100*time.Millisecond)

		allowed, err := mgr.Allow(ctx, key, limit, window)
		assert.NoError(t, err)
		assert.True(t, allowed, "窗口过期后应该恢复")
	})
}

func TestLimiterManagerRedisImpl_Concurrent(t *testing.T) {
	ctx := context.Background()
	key := "test_concurrent"
	limit := 100
	window := time.Second

	t.Run("并发安全测试", func(t *testing.T) {
		t.Skip("内存缓存的 Increment 非原子，需要在真实 Redis 环境中测试")
		mgr := newTestLimiterManagerRedis()
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
		assert.GreaterOrEqual(t, limit, successCount, "并发请求不应超过limit个")
		assert.LessOrEqual(t, successCount, limit+5, "由于竞态条件，允许轻微超出limit")
	})
}

func TestLimiterManagerRedisImpl_Validation(t *testing.T) {
	mgr := newTestLimiterManagerRedis()
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

func TestLimiterManagerRedisImpl_EdgeConditions(t *testing.T) {
	ctx := context.Background()
	limit := 1
	window := 10 * time.Millisecond

	t.Run("limit为1", func(t *testing.T) {
		mgr := newTestLimiterManagerRedis()
		key := "edge1"

		allowed, err := mgr.Allow(ctx, key, limit, window)
		assert.NoError(t, err)
		assert.True(t, allowed)

		allowed, err = mgr.Allow(ctx, key, limit, window)
		assert.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("非常短的窗口", func(t *testing.T) {
		mgr := newTestLimiterManagerRedis()
		key := "edge2"

		allowed, _ := mgr.Allow(ctx, key, limit, 1*time.Millisecond)
		assert.True(t, allowed)

		time.Sleep(2 * time.Millisecond)

		allowed, err := mgr.Allow(ctx, key, limit, 1*time.Millisecond)
		assert.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("非常长的窗口", func(t *testing.T) {
		mgr := newTestLimiterManagerRedis()
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
