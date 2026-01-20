package cachemgr

import (
	"context"
	"fmt"
	"time"
)

// cacheManagerNoneImpl 空缓存实现（降级）
type cacheManagerNoneImpl struct {
	*cacheManagerBaseImpl
	name string
}

// NewCacheManagerNoneImpl 创建空缓存实现
func NewCacheManagerNoneImpl() ICacheManager {
	impl := &cacheManagerNoneImpl{
		cacheManagerBaseImpl: newICacheManagerBaseImpl(),
		name:                 "cacheManagerNoneImpl",
	}
	impl.initObservability()
	return impl
}

// ManagerName 返回管理器名称
func (n *cacheManagerNoneImpl) ManagerName() string {
	return n.name
}

// Health 检查管理器健康状态
func (n *cacheManagerNoneImpl) Health() error {
	return nil
}

// OnStart 在服务器启动时触发
func (n *cacheManagerNoneImpl) OnStart() error {
	return nil
}

// OnStop 在服务器停止时触发
func (n *cacheManagerNoneImpl) OnStop() error {
	return nil
}

// Get 获取缓存值（返回错误）
func (n *cacheManagerNoneImpl) Get(ctx context.Context, key string, dest any) error {
	return n.recordOperation(ctx, n.name, "get", key, func() error {
		return fmt.Errorf("cache not available")
	})
}

// Set 设置缓存值（空操作）
func (n *cacheManagerNoneImpl) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return n.recordOperation(ctx, n.name, "set", key, func() error {
		// 空操作，不返回错误
		return nil
	})
}

// SetNX 仅当键不存在时才设置值（返回 false，表示未设置）
func (n *cacheManagerNoneImpl) SetNX(ctx context.Context, key string, value any,
	expiration time.Duration) (bool, error) {
	var result bool

	err := n.recordOperation(ctx, n.name, "setnx", key, func() error {
		// 返回 false 表示未设置成功
		result = false
		return nil
	})

	return result, err
}

// Delete 删除缓存值（空操作）
func (n *cacheManagerNoneImpl) Delete(ctx context.Context, key string) error {
	return n.recordOperation(ctx, n.name, "delete", key, func() error {
		return nil
	})
}

// Exists 检查键是否存在（返回 false）
func (n *cacheManagerNoneImpl) Exists(ctx context.Context, key string) (bool, error) {
	var result bool

	err := n.recordOperation(ctx, n.name, "exists", key, func() error {
		result = false
		return nil
	})

	return result, err
}

// Expire 设置过期时间（空操作）
func (n *cacheManagerNoneImpl) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return n.recordOperation(ctx, n.name, "expire", key, func() error {
		return nil
	})
}

// TTL 获取剩余过期时间（返回 0）
func (n *cacheManagerNoneImpl) TTL(ctx context.Context, key string) (time.Duration, error) {
	var result time.Duration

	err := n.recordOperation(ctx, n.name, "ttl", key, func() error {
		result = 0
		return nil
	})

	return result, err
}

// Clear 清空所有缓存（空操作）
func (n *cacheManagerNoneImpl) Clear(ctx context.Context) error {
	return n.recordOperation(ctx, n.name, "clear", "", func() error {
		return nil
	})
}

// GetMultiple 批量获取（返回空 map）
func (n *cacheManagerNoneImpl) GetMultiple(ctx context.Context, keys []string) (map[string]any, error) {
	var result map[string]any

	key := "batch"
	if len(keys) > 0 {
		key = keys[0]
	}

	err := n.recordOperation(ctx, n.name, "getmultiple", key, func() error {
		result = make(map[string]any)
		return nil
	})

	return result, err
}

// SetMultiple 批量设置（空操作）
func (n *cacheManagerNoneImpl) SetMultiple(ctx context.Context, items map[string]any, expiration time.Duration) error {
	key := "batch"
	for k := range items {
		key = k
		break
	}

	return n.recordOperation(ctx, n.name, "setmultiple", key, func() error {
		return nil
	})
}

// DeleteMultiple 批量删除（空操作）
func (n *cacheManagerNoneImpl) DeleteMultiple(ctx context.Context, keys []string) error {
	key := "batch"
	if len(keys) > 0 {
		key = keys[0]
	}

	return n.recordOperation(ctx, n.name, "deletemultiple", key, func() error {
		return nil
	})
}

// Increment 自增（返回 0）
func (n *cacheManagerNoneImpl) Increment(ctx context.Context, key string, value int64) (int64, error) {
	var result int64

	err := n.recordOperation(ctx, n.name, "increment", key, func() error {
		result = 0
		return nil
	})

	return result, err
}

// Decrement 自减（返回 0）
func (n *cacheManagerNoneImpl) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	var result int64

	err := n.recordOperation(ctx, n.name, "decrement", key, func() error {
		result = 0
		return nil
	})

	return result, err
}

// Close 关闭空缓存
func (n *cacheManagerNoneImpl) Close() error {
	return nil
}

// 确保 cacheManagerNoneImpl 实现 ICacheManager 接口
var _ ICacheManager = (*cacheManagerNoneImpl)(nil)
