package cachemgr

import (
	"context"
	"time"

	"com.litelake.litecore/manager/cachemgr/internal/drivers"
)

// CacheManagerAdapter 缓存管理器适配器
// 将内部驱动适配到 CacheManager 接口
type CacheManagerAdapter struct {
	driver CacheDriver
}

// CacheDriver 缓存驱动接口
// 内部驱动需要实现此接口
type CacheDriver interface {
	// Manager 接口方法
	ManagerName() string
	Health() error
	OnStart() error
	OnStop() error
	Close() error

	// 缓存操作方法
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	Clear(ctx context.Context) error
	GetMultiple(ctx context.Context, keys []string) (map[string]any, error)
	SetMultiple(ctx context.Context, items map[string]any, expiration time.Duration) error
	DeleteMultiple(ctx context.Context, keys []string) error
	Increment(ctx context.Context, key string, value int64) (int64, error)
	Decrement(ctx context.Context, key string, value int64) (int64, error)
}

// NewCacheManagerAdapter 创建缓存管理器适配器
func NewCacheManagerAdapter(driver CacheDriver) *CacheManagerAdapter {
	return &CacheManagerAdapter{
		driver: driver,
	}
}

// ManagerName 返回管理器名称
func (a *CacheManagerAdapter) ManagerName() string {
	return a.driver.ManagerName()
}

// Health 检查管理器健康状态
func (a *CacheManagerAdapter) Health() error {
	return a.driver.Health()
}

// OnStart 在服务器启动时触发
func (a *CacheManagerAdapter) OnStart() error {
	return a.driver.OnStart()
}

// OnStop 在服务器停止时触发
func (a *CacheManagerAdapter) OnStop() error {
	return a.driver.OnStop()
}

// Get 获取缓存值
func (a *CacheManagerAdapter) Get(ctx context.Context, key string, dest any) error {
	return a.driver.Get(ctx, key, dest)
}

// Set 设置缓存值
func (a *CacheManagerAdapter) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return a.driver.Set(ctx, key, value, expiration)
}

// SetNX 仅当键不存在时才设置值
func (a *CacheManagerAdapter) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	return a.driver.SetNX(ctx, key, value, expiration)
}

// Delete 删除缓存值
func (a *CacheManagerAdapter) Delete(ctx context.Context, key string) error {
	return a.driver.Delete(ctx, key)
}

// Exists 检查键是否存在
func (a *CacheManagerAdapter) Exists(ctx context.Context, key string) (bool, error) {
	return a.driver.Exists(ctx, key)
}

// Expire 设置过期时间
func (a *CacheManagerAdapter) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return a.driver.Expire(ctx, key, expiration)
}

// TTL 获取剩余过期时间
func (a *CacheManagerAdapter) TTL(ctx context.Context, key string) (time.Duration, error) {
	return a.driver.TTL(ctx, key)
}

// Clear 清空所有缓存
func (a *CacheManagerAdapter) Clear(ctx context.Context) error {
	return a.driver.Clear(ctx)
}

// GetMultiple 批量获取
func (a *CacheManagerAdapter) GetMultiple(ctx context.Context, keys []string) (map[string]any, error) {
	return a.driver.GetMultiple(ctx, keys)
}

// SetMultiple 批量设置
func (a *CacheManagerAdapter) SetMultiple(ctx context.Context, items map[string]any, expiration time.Duration) error {
	return a.driver.SetMultiple(ctx, items, expiration)
}

// DeleteMultiple 批量删除
func (a *CacheManagerAdapter) DeleteMultiple(ctx context.Context, keys []string) error {
	return a.driver.DeleteMultiple(ctx, keys)
}

// Increment 自增
func (a *CacheManagerAdapter) Increment(ctx context.Context, key string, value int64) (int64, error) {
	return a.driver.Increment(ctx, key, value)
}

// Decrement 自减
func (a *CacheManagerAdapter) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	return a.driver.Decrement(ctx, key, value)
}

// Close 关闭缓存连接
func (a *CacheManagerAdapter) Close() error {
	return a.driver.Close()
}

// Ensure CacheManagerAdapter implements CacheManager interface
var _ CacheManager = (*CacheManagerAdapter)(nil)

// 确保 RedisManager 实现 CacheDriver 接口
var _ CacheDriver = (*drivers.RedisManager)(nil)

// 确保 MemoryManager 实现 CacheDriver 接口
var _ CacheDriver = (*drivers.MemoryManager)(nil)

// 确保 NoneManager 实现 CacheDriver 接口
var _ CacheDriver = (*drivers.NoneManager)(nil)
