package lockmgr

import (
	"context"
	"time"

	"github.com/lite-lake/litecore-go/common"
)

// ILockManager 锁管理器接口
type ILockManager interface {
	common.IBaseManager

	// Lock 获取锁（阻塞直到成功或超时）
	// ctx: 上下文
	// key: 锁的键
	// ttl: 锁的过期时间
	// 返回: 成功获取锁返回 nil，否则返回错误
	Lock(ctx context.Context, key string, ttl time.Duration) error

	// Unlock 释放锁
	// ctx: 上下文
	// key: 锁的键
	Unlock(ctx context.Context, key string) error

	// TryLock 尝试获取锁（非阻塞）
	// ctx: 上下文
	// key: 锁的键
	// ttl: 锁的过期时间
	// 返回: 成功返回 true，失败返回 false
	TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
}
