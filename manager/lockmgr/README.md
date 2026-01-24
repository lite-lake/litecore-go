# LockMgr - 锁管理器

LockMgr 提供统一的锁管理功能，支持 Redis 分布式锁和 Memory 本地内存锁。

## 特性

- **多驱动支持**：支持 Redis（分布式锁）和 Memory（本地内存锁）
- **统一接口**：提供统一的 ILockManager 接口，便于切换实现
- **可观测性**：内置日志、指标和链路追踪支持
- **自动过期**：支持锁的 TTL 自动过期，防止死锁
- **非阻塞模式**：TryLock 支持非阻塞获取锁
- **依赖注入**：支持通过 `inject:""` 标签注入 Manager

## 快速开始

```go
package main

import (
	"context"
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

	// 获取锁
	err = lockMgr.Lock(ctx, "resource:123", 10*time.Second)
	if err != nil {
		panic(err)
	}
	defer lockMgr.Unlock(ctx, "resource:123")

	// 执行需要加锁的操作
	// ...
}
```

## Lock - 阻塞获取锁

Lock 方法会阻塞直到成功获取锁或上下文取消。

```go
ctx := context.Background()

// 获取锁，自动重试直到成功或上下文取消
err := lockMgr.Lock(ctx, "resource:123", 10*time.Second)
if err != nil {
	return err
}
defer lockMgr.Unlock(ctx, "resource:123")

// 执行需要加锁的操作
```

## Unlock - 释放锁

Unlock 方法释放已持有的锁。

```go
err := lockMgr.Unlock(ctx, "resource:123")
if err != nil {
	return err
}
```

建议使用 defer 确保锁一定会被释放：

```go
lockMgr.Lock(ctx, "resource:123", 10*time.Second)
defer lockMgr.Unlock(ctx, "resource:123")
```

## TryLock - 非阻塞获取锁

TryLock 方法尝试获取锁，立即返回结果，不阻塞。

```go
ctx := context.Background()

locked, err := lockMgr.TryLock(ctx, "resource:123", 10*time.Second)
if err != nil {
	return err
}

if locked {
	defer lockMgr.Unlock(ctx, "resource:123")
	// 执行需要加锁的操作
} else {
	// 锁已被占用
}
```

## 工厂函数

### Build - 手动创建

```go
// 创建内存锁
mgr, err := lockmgr.Build("memory", map[string]any{
	"max_backups": 1000,
})

// 创建 Redis 锁
mgr, err := lockmgr.Build("redis", map[string]any{
	"host":              "localhost",
	"port":              6379,
	"password":          "",
	"db":                0,
	"max_idle_conns":    10,
	"max_open_conns":    100,
	"conn_max_lifetime": "30s",
})
```

### BuildWithConfigProvider - 从配置创建

```go
mgr, err := lockmgr.BuildWithConfigProvider(configProvider)
```

## 驱动实现

### Memory 驱动

基于 sync.Mutex 实现的本地内存锁，适用于单机环境。

```go
cfg := &lockmgr.MemoryLockConfig{
	MaxBackups: 1000,
}
mgr := lockmgr.NewLockManagerMemoryImpl(cfg)
```

### Redis 驱动

基于 Redis SET NX EX 命令实现的分布式锁，适用于分布式环境。

```go
cfg := &lockmgr.RedisLockConfig{
	Host:            "localhost",
	Port:            6379,
	Password:        "",
	DB:              0,
	MaxIdleConns:    10,
	MaxOpenConns:    100,
	ConnMaxLifetime: 30 * time.Second,
}
mgr := lockmgr.NewLockManagerRedisImpl(cfg)
```

Redis 驱动需要依赖 cachemgr.ICacheManager，通过依赖注入自动注入。

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

## API

### ILockManager 接口

```go
type ILockManager interface {
	// Lock 获取锁（阻塞直到成功或上下文取消）
	Lock(ctx context.Context, key string, ttl time.Duration) error

	// Unlock 释放锁
	Unlock(ctx context.Context, key string) error

	// TryLock 尝试获取锁（非阻塞）
	TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
}
```

### 配置类型

```go
type LockConfig struct {
	Driver       string            `yaml:"driver"`        // 驱动类型: redis, memory
	RedisConfig  *RedisLockConfig  `yaml:"redis_config"`  // Redis 配置
	MemoryConfig *MemoryLockConfig `yaml:"memory_config"` // Memory 配置
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

## 错误处理

所有锁操作都可能返回错误，常见错误包括：

- **上下文取消**：Lock 操作在获取锁前上下文被取消
- **连接失败**：Redis 驱动无法连接到 Redis 服务器
- **参数错误**：key 为空或上下文为 nil

```go
err := lockMgr.Lock(ctx, "resource:123", 10*time.Second)
if err != nil {
	return err
}
```

## 可观测性

LockMgr 内置完整的可观测性支持：

- **日志**：记录锁操作的成功和失败事件
- **指标**：
  - `lock.acquire`：锁获取次数
  - `lock.release`：锁释放次数
  - `lock.acquire_failed`：锁获取失败次数
  - `lock.operation.duration`：锁操作耗时
- **链路追踪**：支持 OpenTelemetry 链路追踪

## 性能与线程安全

- **Memory 驱动**：基于 sync.Mutex，线程安全，仅适用于单机环境
- **Redis 驱动**：基于 Redis SET NX EX 命令，适用于分布式环境
- **自动过期**：所有锁都支持 TTL，自动过期防止死锁
- **重试机制**：Lock 方法内部会自动重试获取锁直到上下文取消

## 使用场景

- 分布式环境下的资源互斥访问
- 并发控制，防止重复操作
- 任务队列的任务消费
- 限流控制
- 数据库更新防并发

## 最佳实践

1. **使用 defer 确保锁释放**：在获取锁后立即使用 defer 释放锁
2. **设置合理的 TTL**：根据业务场景设置合适的 TTL，防止死锁
3. **选择合适的驱动**：分布式环境使用 Redis 驱动，单机环境使用 Memory 驱动
4. **锁粒度控制**：锁的粒度应尽可能小，减少锁的持有时间
5. **错误处理**：所有锁操作都应进行错误处理
