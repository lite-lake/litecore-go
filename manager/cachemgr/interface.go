package cachemgr

import (
	"context"
	"time"

	"github.com/lite-lake/litecore-go/common"
)

// ICacheManager 缓存管理器接口
// 提供统一的缓存操作接口，支持多种缓存驱动
type ICacheManager interface {
	common.IBaseManager

	// Get 获取缓存值
	// ctx: 上下文
	// key: 缓存键
	// dest: 目标变量指针，用于接收缓存值
	Get(ctx context.Context, key string, dest any) error

	// Set 设置缓存值
	// ctx: 上下文
	// key: 缓存键
	// value: 缓存值
	// expiration: 过期时间
	Set(ctx context.Context, key string, value any, expiration time.Duration) error

	// SetNX 仅当键不存在时才设置值（Set if Not eXists）
	// 返回值表示是否设置成功：true 表示设置成功，false 表示键已存在
	// 常用于分布式锁、幂等性控制等场景
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error)

	// Delete 删除缓存值
	Delete(ctx context.Context, key string) error

	// Exists 检查键是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// Expire 设置过期时间
	Expire(ctx context.Context, key string, expiration time.Duration) error

	// TTL 获取剩余过期时间
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Clear 清空所有缓存（慎用）
	Clear(ctx context.Context) error

	// GetMultiple 批量获取
	GetMultiple(ctx context.Context, keys []string) (map[string]any, error)

	// SetMultiple 批量设置
	SetMultiple(ctx context.Context, items map[string]any, expiration time.Duration) error

	// DeleteMultiple 批量删除
	DeleteMultiple(ctx context.Context, keys []string) error

	// Increment 自增
	Increment(ctx context.Context, key string, value int64) (int64, error)

	// Decrement 自减
	Decrement(ctx context.Context, key string, value int64) (int64, error)

	// Close 关闭缓存连接
	Close() error
}
