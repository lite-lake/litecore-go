package lockmgr

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type lockManagerRedisImpl struct {
	*lockManagerBaseImpl
	config *RedisLockConfig
	name   string
}

func NewLockManagerRedisImpl(config *RedisLockConfig) ILockManager {
	impl := &lockManagerRedisImpl{
		lockManagerBaseImpl: newLockManagerBaseImpl(),
		config:              config,
		name:                "lockManagerRedisImpl",
	}
	impl.initObservability()
	return impl
}

func (r *lockManagerRedisImpl) ManagerName() string {
	return r.name
}

func (r *lockManagerRedisImpl) Health() error {
	if r.cacheMgr == nil {
		return fmt.Errorf("cache manager not injected")
	}
	return r.cacheMgr.Health()
}

func (r *lockManagerRedisImpl) OnStart() error {
	if r.cacheMgr == nil {
		return fmt.Errorf("cache manager not injected")
	}
	return r.cacheMgr.OnStart()
}

func (r *lockManagerRedisImpl) OnStop() error {
	if r.cacheMgr == nil {
		return nil
	}
	return r.cacheMgr.OnStop()
}

func (r *lockManagerRedisImpl) Lock(ctx context.Context, key string, ttl time.Duration) error {
	if err := ValidateContext(ctx); err != nil {
		return err
	}
	if err := ValidateKey(key); err != nil {
		return err
	}
	if r.cacheMgr == nil {
		return fmt.Errorf("cache manager not injected")
	}

	lockKey := fmt.Sprintf("lock:%s", key)
	lockValue := uuid.New().String()

	const retryInterval = 50 * time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("lock acquisition canceled: %w", ctx.Err())
		default:
		}

		acquired, err := r.cacheMgr.SetNX(ctx, lockKey, lockValue, ttl)
		if err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}

		if acquired {
			r.recordLockAcquire(ctx, "redis", true)
			return nil
		}

		r.recordLockAcquire(ctx, "redis", false)

		select {
		case <-ctx.Done():
			return fmt.Errorf("lock acquisition canceled: %w", ctx.Err())
		case <-time.After(retryInterval):
			continue
		}
	}
}

func (r *lockManagerRedisImpl) Unlock(ctx context.Context, key string) error {
	return r.recordOperation(ctx, "redis", "unlock", key, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateKey(key); err != nil {
			return err
		}
		if r.cacheMgr == nil {
			return fmt.Errorf("cache manager not injected")
		}

		lockKey := fmt.Sprintf("lock:%s", key)

		err := r.cacheMgr.Delete(ctx, lockKey)
		if err != nil {
			return fmt.Errorf("failed to release lock: %w", err)
		}

		r.recordLockRelease(ctx, "redis")
		return nil
	})
}

func (r *lockManagerRedisImpl) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if err := ValidateContext(ctx); err != nil {
		return false, err
	}
	if err := ValidateKey(key); err != nil {
		return false, err
	}
	if r.cacheMgr == nil {
		return false, fmt.Errorf("cache manager not injected")
	}

	lockKey := fmt.Sprintf("lock:%s", key)
	lockValue := uuid.New().String()

	acquired, err := r.cacheMgr.SetNX(ctx, lockKey, lockValue, ttl)
	if err != nil {
		r.recordLockAcquire(ctx, "redis", false)
		return false, fmt.Errorf("failed to try acquire lock: %w", err)
	}

	if acquired {
		r.recordLockAcquire(ctx, "redis", true)
	} else {
		r.recordLockAcquire(ctx, "redis", false)
	}

	return acquired, nil
}

var _ ILockManager = (*lockManagerRedisImpl)(nil)
