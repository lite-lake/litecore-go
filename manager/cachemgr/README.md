# Cache Manager

缓存管理器，提供统一的缓存操作接口，支持 Redis、内存和空缓存三种驱动。

## 特性

- **多驱动支持** - 支持 Redis（分布式）、Memory（高性能内存）、None（降级）三种缓存驱动
- **高性能内存缓存** - 内存驱动基于 Ristretto 库实现，具有极高的性能和内存效率
- **统一接口** - 提供统一的 ICacheManager 接口，便于切换缓存实现
- **可观测性** - 内置日志、指标和链路追踪支持，支持缓存命中率、操作耗时监控
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
mgr := cachemgr.NewCacheManagerMemoryImpl(
    10*time.Minute,  // 默认过期时间（配置参考）
    5*time.Minute,   // 清理间隔（配置参考）
    nil,             // 日志管理器（可选）
    nil,             // 遥测管理器（可选）
)
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
mgr, err := cachemgr.NewCacheManagerRedisImpl(cfg, nil, nil)
if err != nil {
    log.Fatal(err)
}
defer mgr.Close()

// 使用方式与内存缓存相同
err = mgr.Set(ctx, "key", "value", 5*time.Minute)
```

### 使用配置文件创建

```yaml
cache:
  driver: "memory"  # 驱动类型: redis, memory, none
  memory_config:
    max_size: 100    # 最大缓存大小（MB）
    max_age: "720h"  # 最大缓存时间（30天）
    max_backups: 1000
    compress: false
```

```go
// 通过 Build 函数创建
mgr, err := cachemgr.Build("memory", map[string]any{
    "max_size": 100,
    "max_age": "720h",
}, nil, nil)

// 通过配置提供者创建
mgr, err := cachemgr.BuildWithConfigProvider(configProvider, nil, nil)
```

### 依赖注入使用

```go
type SessionService struct {
    CacheMgr cachemgr.ICacheManager `inject:""`
}

func (s *SessionService) CreateSession() (string, error) {
    token := uuid.New().String()
    sessionKey := fmt.Sprintf("session:%s", token)
    session := &Session{Token: token, ExpiresAt: time.Now().Add(time.Hour)}

    ctx := context.Background()
    if err := s.CacheMgr.Set(ctx, sessionKey, session, time.Hour); err != nil {
        return "", err
    }

    return token, nil
}
```

## 支持的缓存驱动

### Memory（内存缓存）

基于 Ristretto 库实现的高性能内存缓存，适合单机应用或开发测试环境。

**Ristretto 优势**：
- 高性能：基于 TinyLFU 淘汰策略，命中率极高
- 高并发：完全无锁设计，支持高并发访问
- 自适应：自动调整淘汰策略，适应不同的访问模式
- 低开销：内存占用低，GC 压力小

**Ristretto 配置参数**：
- `NumCounters`: 1e6（统计计数器数量，用于频率统计）
- `MaxCost`: 1e8（最大缓存成本，约 100MB）
- `BufferItems`: 64（缓冲区大小，影响写入性能）
- `TtlTickerDurationInSec`: 1（TTL 检查间隔，秒）

### Redis（分布式缓存）

基于 Redis 实现的分布式缓存，适合多实例共享缓存的场景。

**Redis 特性**：
- 分布式缓存：支持多实例共享
- 持久化：支持数据持久化到磁盘
- 数据结构：支持丰富的数据类型
- Pipeline：支持批量操作，提高性能

### None（空缓存）

降级模式，不存储任何数据，适合测试或临时禁用缓存的场景。

**None 特性**：
- 所有操作都是空操作或返回默认值
- Get 操作返回错误
- Set/SetNX/Delete 等操作返回成功但不存储数据

## 配置说明

### 缓存配置（CacheConfig）

```yaml
cache:
  driver: "memory"  # 驱动类型: redis, memory, none
```

### Redis 配置（RedisConfig）

```yaml
redis_config:
  host: "localhost"         # Redis 主机地址
  port: 6379                # Redis 端口
  password: ""              # Redis 密码
  db: 0                     # Redis 数据库编号（0-15）
  max_idle_conns: 10        # 最大空闲连接数
  max_open_conns: 100       # 最大打开连接数
  conn_max_lifetime: "30s"  # 连接最大存活时间
```

### Memory 配置（MemoryConfig）

```yaml
memory_config:
  max_size: 100    # 最大缓存大小（MB）
  max_age: "720h"  # 最大缓存时间（支持: "30s", "5m", "1h", "30"）
  max_backups: 1000
  compress: false  # 是否压缩
```

## API 说明

### 基本操作

#### Get

获取缓存值。

```go
var data User
err := mgr.Get(ctx, "user:123", &data)
```

#### Set

设置缓存值。

```go
err := mgr.Set(ctx, "key", "value", 5*time.Minute)
```

#### Delete

删除缓存值。

```go
err := mgr.Delete(ctx, "key")
```

#### Exists

检查键是否存在。

```go
exists, err := mgr.Exists(ctx, "key")
```

### 过期时间管理

#### Expire

设置过期时间。

```go
err := mgr.Expire(ctx, "key", 10*time.Minute)
```

#### TTL

获取剩余过期时间。

```go
ttl, err := mgr.TTL(ctx, "key")
```

### 批量操作

#### GetMultiple

批量获取缓存值。

```go
keys := []string{"key1", "key2", "key3"}
results, err := mgr.GetMultiple(ctx, keys)
```

#### SetMultiple

批量设置缓存值。

```go
items := map[string]any{
    "key1": "value1",
    "key2": "value2",
    "key3": "value3",
}
err := mgr.SetMultiple(ctx, items, 5*time.Minute)
```

#### DeleteMultiple

批量删除缓存值。

```go
err := mgr.DeleteMultiple(ctx, keys)
```

### 原子操作

#### SetNX

仅当键不存在时才设置值（用于分布式锁）。

```go
locked, err := mgr.SetNX(ctx, "lock:resource", "owner", 10*time.Second)
if locked {
    defer mgr.Delete(ctx, "lock:resource")
}
```

#### Increment

自增操作。

```go
counter, err := mgr.Increment(ctx, "counter", 1)
```

#### Decrement

自减操作。

```go
counter, err := mgr.Decrement(ctx, "counter", 1)
```

### 清空操作

#### Clear

清空所有缓存（慎用）。

```go
err := mgr.Clear(ctx)
```

## 工厂函数

### Build

根据驱动类型创建缓存管理器。

```go
func Build(
    driverType string,
    driverConfig map[string]any,
    loggerMgr loggermgr.ILoggerManager,
    telemetryMgr telemetrymgr.ITelemetryManager,
) (ICacheManager, error)
```

**参数**：
- `driverType`: 驱动类型（"redis", "memory", "none"）
- `driverConfig`: 驱动配置（根据驱动类型不同而不同）
- `loggerMgr`: 日志管理器（可选）
- `telemetryMgr`: 遥测管理器（可选）

### BuildWithConfigProvider

从配置提供者创建缓存管理器。

```go
func BuildWithConfigProvider(
    configProvider configmgr.IConfigManager,
    loggerMgr loggermgr.ILoggerManager,
    telemetryMgr telemetrymgr.ITelemetryManager,
) (ICacheManager, error)
```

**配置路径**：
- `cache.driver`: 驱动类型
- `cache.redis_config`: Redis 配置（当 driver=redis 时使用）
- `cache.memory_config`: Memory 配置（当 driver=memory 时使用）

## 可观测性

### 日志

缓存管理器内置日志支持，记录所有操作：

- 操作类型（get, set, delete 等）
- 缓存键（脱敏处理，只保留前 5 个字符）
- 操作状态（成功/失败）
- 操作耗时

### 指标

内置 Prometheus 指标支持：

- `cache.hit`: 缓存命中次数
- `cache.miss`: 缓存未命中次数
- `cache.operation.duration`: 缓存操作耗时（秒）

### 链路追踪

内置 OpenTelemetry 链路追踪支持：

- 记录缓存操作的完整调用链
- 添加缓存键、驱动类型等属性
- 记录操作错误状态

## 错误处理

缓存操作可能返回以下错误：

- 上下文无效：`context cannot be nil`
- 键无效：`cache key cannot be empty`
- 键不存在：`key not found: xxx`
- 连接失败：`failed to connect to redis: xxx`
- 类型不匹配：`type mismatch: cannot assign xxx to xxx`
- 缓存不可用：`cache not available`（None 驱动）

## 最佳实践

1. **使用上下文超时**：为缓存操作设置合理的超时时间，避免长时间阻塞

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
err := mgr.Get(ctx, "key", &data)
```

2. **合理设置过期时间**：根据数据更新频率设置合适的过期时间

```go
// 热点数据：较短的过期时间
mgr.Set(ctx, "hot_data", data, 5*time.Minute)

// 静态数据：较长的过期时间
mgr.Set(ctx, "static_config", config, 24*time.Hour)
```

3. **使用批量操作**：对于多个键的操作，优先使用批量 API

```go
// 不推荐
for _, key := range keys {
    mgr.Get(ctx, key, &data)
}

// 推荐
results, err := mgr.GetMultiple(ctx, keys)
```

4. **处理连接错误**：Redis 连接失败时，考虑降级到内存缓存或空缓存

```go
err := mgr.Get(ctx, "key", &data)
if err != nil && strings.Contains(err.Error(), "failed to connect") {
    log.Warn("Redis 连接失败，降级到内存缓存")
    fallbackMgr.Get(ctx, "key", &data)
}
```

5. **监控缓存命中率**：利用内置指标监控缓存命中率，优化缓存策略

```go
// 通过 Prometheus 查询缓存命中率
rate(cache_hit_total[5m]) / (rate(cache_hit_total[5m]) + rate(cache_miss_total[5m]))
```

6. **避免存储大对象**：大对象会占用较多内存和网络带宽，建议压缩或分片存储

```go
// 不推荐
mgr.Set(ctx, "large_data", hugeObject, time.Hour)

// 推荐：压缩或拆分
compressed := compress(data)
mgr.Set(ctx, "data:compressed", compressed, time.Hour)
```

7. **使用分布式锁**：在并发场景下使用 SetNX 实现分布式锁

```go
locked, err := mgr.SetNX(ctx, "lock:resource", owner, 10*time.Second)
if err != nil || !locked {
    return fmt.Errorf("获取锁失败")
}
defer mgr.Delete(ctx, "lock:resource")

// 执行需要加锁的操作
```
