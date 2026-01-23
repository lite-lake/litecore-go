# LockMgr - 锁管理器

LockMgr 提供统一的锁管理功能，支持 Redis 分布式锁和本地内存锁。

## 特性

- **多驱动支持**：支持 Redis（分布式锁）和 Memory（本地内存锁）
- **统一接口**：提供统一的 ILockManager 接口，便于切换锁实现
- **自动过期**：支持锁的自动过期机制，防止死锁
- **非阻塞模式**：TryLock 支持非阻塞获取锁
- **可观测性**：内置日志、指标和链路追踪支持

## 快速开始

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/lite-lake/litecore-go/manager/lockmgr"
)

func main() {
    ctx := context.Background()

    // 创建内存锁管理器
    lockMgr, err := lockmgr.Build("memory", map[string]any{
        "max_backups": 1000,
    })
    if err != nil {
        panic(err)
    }
    defer lockMgr.Close()

    // 获取锁（阻塞）
    err = lockMgr.Lock(ctx, "resource:123", 10*time.Second)
    if err != nil {
        panic(err)
    }
    defer lockMgr.Unlock(ctx, "resource:123")

    // 执行需要加锁的操作
    fmt.Println("操作执行中...")
}
```

## Lock - 阻塞获取锁

Lock 方法会阻塞直到成功获取锁或超时。

```go
ctx := context.Background()

// 获取锁，最多等待 10 秒
err := lockMgr.Lock(ctx, "resource:123", 10*time.Second)
if err != nil {
    // 处理错误
    return
}
defer lockMgr.Unlock(ctx, "resource:123")

// 执行需要加锁的操作
```

## Unlock - 释放锁

Unlock 方法释放已持有的锁。

```go
// 释放锁
err := lockMgr.Unlock(ctx, "resource:123")
if err != nil {
    // 处理错误
    return
}
```

建议使用 defer 确保锁一定会被释放：

```go
lockMgr.Lock(ctx, "resource:123", 10*time.Second)
defer lockMgr.Unlock(ctx, "resource:123")
```

## TryLock - 非阻塞获取锁

TryLock 方法尝试获取锁，立即返回结果。

```go
ctx := context.Background()

// 尝试获取锁（非阻塞）
locked, err := lockMgr.TryLock(ctx, "resource:123", 10*time.Second)
if err != nil {
    // 处理错误
    return
}

if locked {
    // 成功获取锁，执行操作
    defer lockMgr.Unlock(ctx, "resource:123")
    // ...
} else {
    // 锁已被占用，执行其他逻辑
}
```

## 工厂函数

### Build - 手动创建

```go
// 创建内存锁
mgr, err := lockmgr.Build("memory", map[string]any{
    "max_backups": 1000,
})
if err != nil {
    panic(err)
}
defer mgr.Close()
```

### BuildWithConfigProvider - 从配置创建

```go
// 从配置提供者创建
mgr, err := lockmgr.BuildWithConfigProvider(configProvider)
if err != nil {
    panic(err)
}
defer mgr.Close()
```

## API

### 锁操作接口

```go
type ILockManager interface {
    // Lock 获取锁（阻塞直到成功或超时）
    Lock(ctx context.Context, key string, ttl time.Duration) error

    // Unlock 释放锁
    Unlock(ctx context.Context, key string) error

    // TryLock 尝试获取锁（非阻塞）
    TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
}
```

### 工厂函数

```go
// Build 创建锁管理器实例
func Build(driverType string, driverConfig map[string]any) (ILockManager, error)

// BuildWithConfigProvider 从配置提供者创建锁管理器实例
func BuildWithConfigProvider(configProvider configmgr.IConfigManager) (ILockManager, error)
```

### 配置类型

```go
type LockConfig struct {
    Driver       string            // 驱动类型: redis, memory
    RedisConfig  *RedisLockConfig  // Redis 配置
    MemoryConfig *MemoryLockConfig // Memory 配置
}

type RedisLockConfig struct {
    Host            string        // Redis 主机地址
    Port            int           // Redis 端口
    Password        string        // Redis 密码
    DB              int           // Redis 数据库编号
    MaxIdleConns    int           // 最大空闲连接数
    MaxOpenConns    int           // 最大打开连接数
    ConnMaxLifetime time.Duration // 连接最大存活时间
}

type MemoryLockConfig struct {
    MaxBackups int // 最大备份项数
}
```

## 配置示例

### Redis 驱动

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

### Memory 驱动

```yaml
lock:
  driver: memory
  memory_config:
    max_backups: 1000
```

## 错误处理

所有锁操作都可能返回错误，建议进行错误处理：

```go
err := lockMgr.Lock(ctx, "resource:123", 10*time.Second)
if err != nil {
    // 处理错误（如超时、网络问题等）
    return err
}
```

常见错误：
- 超时：Lock 在 TTL 时间内未获取到锁
- 连接失败：Redis 驱动无法连接到 Redis 服务器
- 参数错误：key 为空或 TTL 无效

## 性能与线程安全

- **Memory 驱动**：基于 sync.Mutex，线程安全，仅适用于单机环境
- **Redis 驱动**：基于 Redis SET NX EX 命令实现，适用于分布式环境
- **自动过期**：所有锁都支持 TTL，自动过期防止死锁
- **重试机制**：Lock 方法内部会自动重试获取锁直到超时

## 可观测性

LockMgr 内置了完整的可观测性支持：

- **日志**：记录锁操作的成功和失败事件
- **指标**：
  - `lock.acquire`：锁获取次数
  - `lock.release`：锁释放次数
  - `lock.acquire_failed`：锁获取失败次数
  - `lock.operation.duration`：锁操作耗时
- **链路追踪**：支持 OpenTelemetry 链路追踪

## 最佳实践

1. **使用 defer 确保锁释放**：在获取锁后立即使用 defer 释放锁
2. **设置合理的 TTL**：根据业务场景设置合适的 TTL，防止死锁
3. **选择合适的驱动**：分布式环境使用 Redis 驱动，单机环境使用 Memory 驱动
4. **错误处理**：所有锁操作都应进行错误处理
5. **锁粒度**：锁的粒度应尽可能小，减少锁的持有时间
