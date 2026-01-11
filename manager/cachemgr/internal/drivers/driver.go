package drivers

import (
	"context"
	"time"

	"com.litelake.litecore/manager/cachemgr/internal/config"
)

// Driver 缓存驱动接口
// 所有缓存驱动（Redis、Memory、None）必须实现此接口
type Driver interface {
	// 生命周期管理
	Name() string
	Start() error
	Stop() error
	Health() error

	// 基本操作
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// 批量操作
	Clear(ctx context.Context) error
	GetMultiple(ctx context.Context, keys []string) (map[string]any, error)
	SetMultiple(ctx context.Context, items map[string]any, expiration time.Duration) error
	DeleteMultiple(ctx context.Context, keys []string) error

	// 计数器操作
	Increment(ctx context.Context, key string, value int64) (int64, error)
	Decrement(ctx context.Context, key string, value int64) (int64, error)
}

// NewRedisDriver 创建 Redis 驱动
func NewRedisDriver(cfg *config.RedisConfig) (Driver, error) {
	manager, err := NewRedisManager(cfg)
	if err != nil {
		return nil, err
	}
	return &redisDriverAdapter{manager: manager}, nil
}

// NewMemoryDriver 创建内存驱动
func NewMemoryDriver(cfg *config.MemoryConfig) (Driver, error) {
	manager := NewMemoryManager(cfg.MaxAge, cfg.MaxAge)
	return &memoryDriverAdapter{manager: manager}, nil
}

// NewNoneDriver 创建空驱动（降级使用）
func NewNoneDriver() Driver {
	manager := NewNoneManager()
	return &noneDriverAdapter{manager: manager}
}

// redisDriverAdapter Redis 驱动适配器
type redisDriverAdapter struct {
	manager *RedisManager
}

func (a *redisDriverAdapter) Name() string {
	return "redis"
}

func (a *redisDriverAdapter) Start() error {
	return nil
}

func (a *redisDriverAdapter) Stop() error {
	return a.manager.Close()
}

func (a *redisDriverAdapter) Health() error {
	return a.manager.Health()
}

func (a *redisDriverAdapter) Get(ctx context.Context, key string, dest any) error {
	return a.manager.Get(ctx, key, dest)
}

func (a *redisDriverAdapter) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return a.manager.Set(ctx, key, value, expiration)
}

func (a *redisDriverAdapter) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	return a.manager.SetNX(ctx, key, value, expiration)
}

func (a *redisDriverAdapter) Delete(ctx context.Context, key string) error {
	return a.manager.Delete(ctx, key)
}

func (a *redisDriverAdapter) Exists(ctx context.Context, key string) (bool, error) {
	return a.manager.Exists(ctx, key)
}

func (a *redisDriverAdapter) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return a.manager.Expire(ctx, key, expiration)
}

func (a *redisDriverAdapter) TTL(ctx context.Context, key string) (time.Duration, error) {
	return a.manager.TTL(ctx, key)
}

func (a *redisDriverAdapter) Clear(ctx context.Context) error {
	return a.manager.Clear(ctx)
}

func (a *redisDriverAdapter) GetMultiple(ctx context.Context, keys []string) (map[string]any, error) {
	return a.manager.GetMultiple(ctx, keys)
}

func (a *redisDriverAdapter) SetMultiple(ctx context.Context, items map[string]any, expiration time.Duration) error {
	return a.manager.SetMultiple(ctx, items, expiration)
}

func (a *redisDriverAdapter) DeleteMultiple(ctx context.Context, keys []string) error {
	return a.manager.DeleteMultiple(ctx, keys)
}

func (a *redisDriverAdapter) Increment(ctx context.Context, key string, value int64) (int64, error) {
	return a.manager.Increment(ctx, key, value)
}

func (a *redisDriverAdapter) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	return a.manager.Decrement(ctx, key, value)
}

// memoryDriverAdapter 内存驱动适配器
type memoryDriverAdapter struct {
	manager *MemoryManager
}

func (a *memoryDriverAdapter) Name() string {
	return "memory"
}

func (a *memoryDriverAdapter) Start() error {
	return nil
}

func (a *memoryDriverAdapter) Stop() error {
	return nil
}

func (a *memoryDriverAdapter) Health() error {
	return nil
}

func (a *memoryDriverAdapter) Get(ctx context.Context, key string, dest any) error {
	return a.manager.Get(ctx, key, dest)
}

func (a *memoryDriverAdapter) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return a.manager.Set(ctx, key, value, expiration)
}

func (a *memoryDriverAdapter) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	return a.manager.SetNX(ctx, key, value, expiration)
}

func (a *memoryDriverAdapter) Delete(ctx context.Context, key string) error {
	return a.manager.Delete(ctx, key)
}

func (a *memoryDriverAdapter) Exists(ctx context.Context, key string) (bool, error) {
	return a.manager.Exists(ctx, key)
}

func (a *memoryDriverAdapter) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return a.manager.Expire(ctx, key, expiration)
}

func (a *memoryDriverAdapter) TTL(ctx context.Context, key string) (time.Duration, error) {
	return a.manager.TTL(ctx, key)
}

func (a *memoryDriverAdapter) Clear(ctx context.Context) error {
	return a.manager.Clear(ctx)
}

func (a *memoryDriverAdapter) GetMultiple(ctx context.Context, keys []string) (map[string]any, error) {
	return a.manager.GetMultiple(ctx, keys)
}

func (a *memoryDriverAdapter) SetMultiple(ctx context.Context, items map[string]any, expiration time.Duration) error {
	return a.manager.SetMultiple(ctx, items, expiration)
}

func (a *memoryDriverAdapter) DeleteMultiple(ctx context.Context, keys []string) error {
	return a.manager.DeleteMultiple(ctx, keys)
}

func (a *memoryDriverAdapter) Increment(ctx context.Context, key string, value int64) (int64, error) {
	return a.manager.Increment(ctx, key, value)
}

func (a *memoryDriverAdapter) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	return a.manager.Decrement(ctx, key, value)
}

// noneDriverAdapter 空驱动适配器
type noneDriverAdapter struct {
	manager *NoneManager
}

func (a *noneDriverAdapter) Name() string {
	return "none"
}

func (a *noneDriverAdapter) Start() error {
	return nil
}

func (a *noneDriverAdapter) Stop() error {
	return nil
}

func (a *noneDriverAdapter) Health() error {
	return nil
}

func (a *noneDriverAdapter) Get(ctx context.Context, key string, dest any) error {
	return a.manager.Get(ctx, key, dest)
}

func (a *noneDriverAdapter) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return a.manager.Set(ctx, key, value, expiration)
}

func (a *noneDriverAdapter) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	return a.manager.SetNX(ctx, key, value, expiration)
}

func (a *noneDriverAdapter) Delete(ctx context.Context, key string) error {
	return a.manager.Delete(ctx, key)
}

func (a *noneDriverAdapter) Exists(ctx context.Context, key string) (bool, error) {
	return a.manager.Exists(ctx, key)
}

func (a *noneDriverAdapter) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return a.manager.Expire(ctx, key, expiration)
}

func (a *noneDriverAdapter) TTL(ctx context.Context, key string) (time.Duration, error) {
	return a.manager.TTL(ctx, key)
}

func (a *noneDriverAdapter) Clear(ctx context.Context) error {
	return a.manager.Clear(ctx)
}

func (a *noneDriverAdapter) GetMultiple(ctx context.Context, keys []string) (map[string]any, error) {
	return a.manager.GetMultiple(ctx, keys)
}

func (a *noneDriverAdapter) SetMultiple(ctx context.Context, items map[string]any, expiration time.Duration) error {
	return a.manager.SetMultiple(ctx, items, expiration)
}

func (a *noneDriverAdapter) DeleteMultiple(ctx context.Context, keys []string) error {
	return a.manager.DeleteMultiple(ctx, keys)
}

func (a *noneDriverAdapter) Increment(ctx context.Context, key string, value int64) (int64, error) {
	return a.manager.Increment(ctx, key, value)
}

func (a *noneDriverAdapter) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	return a.manager.Decrement(ctx, key, value)
}

// Ensure all adapters implement Driver interface
var _ Driver = (*redisDriverAdapter)(nil)
var _ Driver = (*memoryDriverAdapter)(nil)
var _ Driver = (*noneDriverAdapter)(nil)
