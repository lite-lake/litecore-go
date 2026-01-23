# LockMgr - 锁管理器

LockMgr 提供统一的锁管理功能，支持多种锁驱动实现。

## 功能特性

- **多驱动支持**：支持 Redis（分布式锁）、Memory（本地内存锁）
- **统一接口**：提供统一的 `ILockManager` 接口，便于切换锁实现
- **可观测性**：内置日志、指标和链路追踪支持
- **自动过期**：支持锁的自动过期机制，防止死锁
- **非阻塞模式**：TryLock 支持非阻塞获取锁

## 驱动类型

### Redis（分布式锁）

基于 Redis 的分布式锁，适用于分布式环境下的资源互斥访问。

配置示例：
```yaml
lock:
  driver: redis
  redis_config:
    host: localhost
    port: 6379
    password: ""
    db: 0
    max_idle_conns: 10
    max_open_conns: 100
    conn_max_lifetime: 30s
```

### Memory（内存锁）

基于本地内存的锁，适用于单机环境下的并发控制。

配置示例：
```yaml
lock:
  driver: memory
  memory_config:
    max_backups: 1000
```

## 基本用法

### 获取和释放锁

```go
// 获取锁（阻塞）
err := lockMgr.Lock(ctx, "resource:123", 10*time.Second)
if err != nil {
    log.Fatal(err)
}

// 执行需要加锁的操作
// ...

// 释放锁
err = lockMgr.Unlock(ctx, "resource:123")
```

### 非阻塞获取锁

```go
// 尝试获取锁（非阻塞）
locked, err := lockMgr.TryLock(ctx, "resource:123", 10*time.Second)
if err != nil {
    log.Fatal(err)
}
if locked {
    // 成功获取锁，执行操作
    defer lockMgr.Unlock(ctx, "resource:123")
    // ...
} else {
    // 锁已被占用，执行其他逻辑
}
```

### 使用工厂函数创建

```go
// 使用 Build 函数创建
mgr, err := lockmgr.Build("memory", map[string]any{
    "max_backups": 1000,
})
if err != nil {
    log.Fatal(err)
}
defer mgr.Close()

// 使用 BuildWithConfigProvider 从配置文件创建
mgr, err := lockmgr.BuildWithConfigProvider(configProvider)
if err != nil {
    log.Fatal(err)
}
defer mgr.Close()
```

## 接口定义

```go
type ILockManager interface {
    common.IBaseManager

    // Lock 获取锁（阻塞直到成功或超时）
    Lock(ctx context.Context, key string, ttl time.Duration) error

    // Unlock 释放锁
    Unlock(ctx context.Context, key string) error

    // TryLock 尝试获取锁（非阻塞）
    TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
}
```

## 配置结构

### LockConfig

```go
type LockConfig struct {
    Driver       string            `yaml:"driver"`        // 驱动类型
    RedisConfig  *RedisLockConfig  `yaml:"redis_config"`  // Redis 配置
    MemoryConfig *MemoryLockConfig `yaml:"memory_config"` // Memory 配置
}
```

### RedisLockConfig

```go
type RedisLockConfig struct {
    Host            string        // Redis 主机地址
    Port            int           // Redis 端口
    Password        string        // Redis 密码
    DB              int           // Redis 数据库编号
    MaxIdleConns    int           // 最大空闲连接数
    MaxOpenConns    int           // 最大打开连接数
    ConnMaxLifetime time.Duration // 连接最大存活时间
}
```

### MemoryLockConfig

```go
type MemoryLockConfig struct {
    MaxBackups int // 最大备份项数（清理策略相关）
}
```

## 可观测性

LockMgr 内置了完整的可观测性支持：

- **日志**：记录锁操作的成功和失败事件
- **指标**：
  - `lock.acquire`：锁获取次数
  - `lock.release`：锁释放次数
  - `lock.acquire_failed`：锁获取失败次数
  - `lock.operation.duration`：锁操作耗时
- **链路追踪**：支持 OpenTelemetry 链路追踪

## 使用场景

- 分布式环境下的资源互斥访问
- 并发控制，防止重复操作
- 任务队列的任务消费
- 限流控制
- 防止缓存击穿

## 注意事项

1. **锁的自动过期**：确保设置合理的 TTL，防止死锁
2. **锁的释放**：使用 defer 确保锁一定能被释放
3. **Redis 驱动依赖**：Redis 驱动需要注入 ICacheManager
4. **分布式环境**：分布式环境请使用 Redis 驱动
