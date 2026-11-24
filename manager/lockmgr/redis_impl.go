package lockmgr

import (
	"context"
	"fmt"
	"time"

	"github.com/lite-lake/litecore-go/manager/cachemgr"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/manager/telemetrymgr"

	"github.com/google/uuid"
)

// lockManagerRedisImpl Redis锁管理器实现
type lockManagerRedisImpl struct {
	*lockManagerBaseImpl                  // 基础锁管理器实现
	config               *RedisLockConfig // Redis锁配置
	name                 string           // 管理器名称
}

// NewLockManagerRedisImpl 创建Redis锁管理器实例
// 参数：
//   - loggerMgr: 日志管理器
//   - telemetryMgr: 遥测管理器
//   - cacheMgr: 缓存管理器
//   - config: Redis锁配置
//
// @return ILockManager 锁管理器接口
func NewLockManagerRedisImpl(
	loggerMgr loggermgr.ILoggerManager,
	telemetryMgr telemetrymgr.ITelemetryManager,
	cacheMgr cachemgr.ICacheManager,
	config *RedisLockConfig,
) ILockManager {
	impl := &lockManagerRedisImpl{
		lockManagerBaseImpl: newLockManagerBaseImpl(loggerMgr, telemetryMgr, cacheMgr),
		config:              config,
		name:                "lockManagerRedisImpl",
	}
	impl.initObservability()
	return impl
}

// ManagerName 返回管理器名称
// @return string 管理器名称
func (r *lockManagerRedisImpl) ManagerName() string {
	return r.name
}

// Health 健康检查
// @return error 健康状态，nil表示正常
func (r *lockManagerRedisImpl) Health() error {
	if r.cacheMgr == nil {
		return fmt.Errorf("cache manager not injected")
	}
	return r.cacheMgr.Health()
}

// OnStart 启动时的初始化操作
// @return error 初始化错误，nil表示成功
func (r *lockManagerRedisImpl) OnStart() error {
	if r.cacheMgr == nil {
		return fmt.Errorf("cache manager not injected")
	}
	return r.cacheMgr.OnStart()
}

// OnStop 停止时的清理操作
// @return error 清理错误，nil表示成功
func (r *lockManagerRedisImpl) OnStop() error {
	if r.cacheMgr == nil {
		return nil
	}
	return r.cacheMgr.OnStop()
}

// Lock 获取锁，阻塞直到成功或上下文取消
// @param ctx 上下文
// @param key 锁的键
// @param ttl 锁的存活时间，0表示不过期
// @return error 错误信息
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

// Unlock 释放锁
// @param ctx 上下文
// @param key 锁的键
// @return error 错误信息
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

// TryLock 尝试获取锁，不阻塞
// @param ctx 上下文
// @param key 锁的键
// @param ttl 锁的存活时间，0表示不过期
// @return bool 是否获取成功
// @return error 错误信息
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
