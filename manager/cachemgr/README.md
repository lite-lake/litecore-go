# Cache Manager

缓存管理器，提供统一的缓存操作接口，支持 Redis、内存和空缓存三种驱动。

## 特性

- **多驱动支持** - 支持 Redis（分布式）、Memory（高性能内存）、None（降级）三种缓存驱动
- **高性能内存缓存** - 内存驱动基于 Ristretto 库实现，具有极高的性能和内存效率
- **统一接口** - 提供统一的 ICacheManager 接口，便于切换缓存实现
- **可观测性** - 内置日志、指标和链路追踪支持
- **连接池管理** - Redis 驱动支持连接池配置和自动管理
- **批量操作** - 支持批量获取、设置和删除操作
- **原子操作** - 支持 SetNX、Increment、Decrement 等原子操作

## 快速开始

### 使用内存缓存

```go
import (
    "context"
    "time"
    "github.com/lite-lake/litecore-go/manager/cachemgr"
)

// 创建内存缓存管理器
mgr := cachemgr.NewCacheManagerMemoryImpl(10*time.Minute, 5*time.Minute)
defer mgr.Close()

ctx := context.Background()

// 设置缓存
err := mgr.Set(ctx, "user:123", userData, 5*time.Minute)
if err != nil {
    log.Fatal(err)
}

// 获取缓存
var data User
err = mgr.Get(ctx, "user:123", &data)
if err != nil {
    log.Fatal(err)
}
```

### 使用 Redis 缓存

```go
// 创建 Redis 配置
cfg := &cachemgr.RedisConfig{
    Host:            "localhost",
    Port:            6379,
    Password:        "",
    DB:              0,
    MaxIdleConns:    10,
    MaxOpenConns:    100,
    ConnMaxLifetime: 30 * time.Second,
}

// 创建 Redis 缓存管理器
mgr, err := cachemgr.NewCacheManagerRedisImpl(cfg)
if err != nil {
    log.Fatal(err)
}
defer mgr.Close()

// 使用方式与内存缓存相同
err = mgr.Set(ctx, "key", "value", 5*time.Minute)
```

### 使用配置创建

```go
// 通过 Build 函数创建
mgr, err := cachemgr.Build("memory", map[string]any{
    "max_age": "1h",
})

// 通过配置提供者创建
mgr, err := cachemgr.BuildWithConfigProvider(configProvider)
```

## 核心功能

### 基本操作

```go
// 设置缓存
err := mgr.Set(ctx, "key", "value", 5*time.Minute)

// 获取缓存
var result string
err := mgr.Get(ctx, "key", &result)

// 删除缓存
err := mgr.Delete(ctx, "key")

// 检查键是否存在
exists, err := mgr.Exists(ctx, "key")
```

### 过期时间管理

```go
// 设置过期时间
err := mgr.Expire(ctx, "key", 10*time.Minute)

// 查看剩余时间
ttl, err := mgr.TTL(ctx, "key")
```

### 批量操作

```go
// 批量获取
keys := []string{"key1", "key2", "key3"}
results, err := mgr.GetMultiple(ctx, keys)

// 批量设置
items := map[string]any{
    "key1": "value1",
    "key2": "value2",
    "key3": "value3",
}
err := mgr.SetMultiple(ctx, items, 5*time.Minute)

// 批量删除
err := mgr.DeleteMultiple(ctx, keys)
```

### 原子操作

```go
// SetNX - 仅当键不存在时设置（用于分布式锁）
locked, err := mgr.SetNX(ctx, "lock:resource", "owner", 10*time.Second)
if locked {
    // 执行需要加锁的操作
    // ...

    // 释放锁
    mgr.Delete(ctx, "lock:resource")
}

// Increment - 自增
counter, err := mgr.Increment(ctx, "counter", 1)

// Decrement - 自减
counter, err := mgr.Decrement(ctx, "counter", 1)
```

### 清空缓存

```go
// 清空所有缓存（慎用）
err := mgr.Clear(ctx)
```

## API

### 接口

#### ICacheManager

缓存管理器接口，所有驱动实现都必须实现该接口。

```go
type ICacheManager interface {
    // 基本操作
    Get(ctx context.Context, key string, dest any) error
    Set(ctx context.Context, key string, value any, expiration time.Duration) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)

    // 过期时间
    Expire(ctx context.Context, key string, expiration time.Duration) error
    TTL(ctx context.Context, key string) (time.Duration, error)

    // 原子操作
    SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error)
    Increment(ctx context.Context, key string, value int64) (int64, error)
    Decrement(ctx context.Context, key string, value int64) (int64, error)

    // 批量操作
    GetMultiple(ctx context.Context, keys []string) (map[string]any, error)
    SetMultiple(ctx context.Context, items map[string]any, expiration time.Duration) error
    DeleteMultiple(ctx context.Context, keys []string) error

    // 清空
    Clear(ctx context.Context) error

    // 生命周期
    Close() error
}
```

### 配置类型

#### RedisConfig

Redis 驱动配置。

```go
type RedisConfig struct {
    Host            string        // Redis 主机地址
    Port            int           // Redis 端口
    Password        string        // Redis 密码
    DB              int           // Redis 数据库编号
    MaxIdleConns    int           // 最大空闲连接数
    MaxOpenConns    int           // 最大打开连接数
    ConnMaxLifetime time.Duration // 连接最大存活时间
}
```

#### MemoryConfig

内存驱动配置。

```go
type MemoryConfig struct {
    MaxSize    int           // 最大缓存大小（MB）
    MaxAge     time.Duration // 最大缓存时间
    MaxBackups int           // 最大备份项数
    Compress   bool          // 是否压缩
}
```

**注意**：内存驱动基于 Ristretto 库实现，内部配置了以下参数：
- `NumCounters`: 1e6（统计计数器数量）
- `MaxCost`: 1e8（最大缓存成本）
- `BufferItems`: 64（缓冲区大小）
- `TtlTickerDurationInSec`: 1（TTL 检查间隔，秒）

Ristretto 使用基于 TinyLFU 的淘汰策略，具有极高的命中率和并发性能。

### 工厂函数

#### Build

根据驱动类型创建缓存管理器。

```go
func Build(driverType string, driverConfig map[string]any) (ICacheManager, error)
```

参数：
- `driverType`: 驱动类型（"redis", "memory", "none"）
- `driverConfig`: 驱动配置

#### BuildWithConfigProvider

从配置提供者创建缓存管理器。

```go
func BuildWithConfigProvider(configProvider configmgr.IConfigManager) (ICacheManager, error)
```

参数：
- `configProvider`: 配置提供者

配置路径：
- `cache.driver`: 驱动类型
- `cache.redis_config`: Redis 配置
- `cache.memory_config`: 内存配置

### 构造函数

#### NewCacheManagerRedisImpl

创建 Redis 缓存管理器。

```go
func NewCacheManagerRedisImpl(cfg *RedisConfig) (ICacheManager, error)
```

#### NewCacheManagerMemoryImpl

创建内存缓存管理器。

```go
func NewCacheManagerMemoryImpl(defaultExpiration, cleanupInterval time.Duration) ICacheManager
```

#### NewCacheManagerNoneImpl

创建空缓存管理器。

```go
func NewCacheManagerNoneImpl() ICacheManager
```

## 性能考虑

### 内存缓存

- 基于 Ristretto 库，高性能、高并发
- 使用 TinyLFU 淘汰策略，命中率高
- 支持自动清理过期项
- 适合单机应用或开发测试环境

### Redis 缓存

- 分布式缓存，支持多实例共享
- 使用连接池管理连接
- 支持管道（Pipeline）批量操作
- 使用 Gob 序列化复杂数据类型

### 空缓存

- 降级模式，不存储任何数据
- 所有操作都是空操作或返回默认值
- 适合测试或临时禁用缓存的场景

## 错误处理

缓存操作可能返回以下错误：

- 上下文无效：`context cannot be nil`
- 键无效：`cache key cannot be empty`
- 键不存在：`key not found: xxx`
- 连接失败：`failed to connect to redis: xxx`
- 类型不匹配：`type mismatch: cannot assign xxx to xxx`

建议在业务代码中处理这些错误，根据需要重试或降级处理。

## 最佳实践

1. **使用上下文超时**：为缓存操作设置合理的超时时间，避免长时间阻塞
2. **合理设置过期时间**：根据数据更新频率设置合适的过期时间
3. **使用批量操作**：对于多个键的操作，优先使用批量 API
4. **处理连接错误**：Redis 连接失败时，考虑降级到内存缓存或空缓存
5. **监控缓存命中率**：利用内置指标监控缓存命中率，优化缓存策略
6. **避免存储大对象**：大对象会占用较多内存和网络带宽，建议压缩或分片存储
