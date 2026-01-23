package limitermgr

import (
	"context"
	"fmt"
	"time"

	"github.com/lite-lake/litecore-go/server/builtin/manager/cachemgr"
)

type limiterManagerRedisImpl struct {
	*limiterManagerBaseImpl
	CacheMgr cachemgr.ICacheManager `inject:""`
	name     string
}

const limiterKeyPrefix = "limiter:"

func NewLimiterManagerRedisImpl() ILimiterManager {
	impl := &limiterManagerRedisImpl{
		limiterManagerBaseImpl: newILimiterManagerBaseImpl(),
		name:                   "limiterManagerRedisImpl",
	}
	impl.initObservability()
	return impl
}

func (r *limiterManagerRedisImpl) ManagerName() string {
	return r.name
}

func (r *limiterManagerRedisImpl) Health() error {
	if r.CacheMgr == nil {
		return fmt.Errorf("cache manager is not initialized")
	}
	return r.CacheMgr.Health()
}

func (r *limiterManagerRedisImpl) OnStart() error {
	return nil
}

func (r *limiterManagerRedisImpl) OnStop() error {
	return nil
}

func (r *limiterManagerRedisImpl) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	var result bool

	err := r.recordOperation(ctx, "redis", "allow", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}
		if err := ValidateLimit(limit); err != nil {
			return err
		}
		if err := ValidateWindow(window); err != nil {
			return err
		}

		if r.CacheMgr == nil {
			return fmt.Errorf("cache manager is not initialized")
		}

		cacheKey := limiterKeyPrefix + key

		count, err := r.CacheMgr.Increment(ctx, cacheKey, 1)
		if err != nil {
			return fmt.Errorf("failed to increment counter: %w", err)
		}

		if count == 1 {
			if err := r.CacheMgr.Expire(ctx, cacheKey, window); err != nil {
				return fmt.Errorf("failed to set expiration: %w", err)
			}
		}

		allowed := int(count) <= limit
		result = allowed
		r.recordAllowance(ctx, "redis", allowed)

		return nil
	})

	return result, err
}

func (r *limiterManagerRedisImpl) GetRemaining(ctx context.Context, key string, limit int, window time.Duration) (int, error) {
	var result int

	err := r.recordOperation(ctx, "redis", "get_remaining", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}
		if err := ValidateLimit(limit); err != nil {
			return err
		}
		if err := ValidateWindow(window); err != nil {
			return err
		}

		if r.CacheMgr == nil {
			return fmt.Errorf("cache manager is not initialized")
		}

		cacheKey := limiterKeyPrefix + key

		var count int64
		err := r.CacheMgr.Get(ctx, cacheKey, &count)
		if err != nil {
			result = limit
			return nil
		}

		remaining := limit - int(count)
		if remaining < 0 {
			remaining = 0
		}
		result = remaining

		return nil
	})

	return result, err
}

var _ ILimiterManager = (*limiterManagerRedisImpl)(nil)
