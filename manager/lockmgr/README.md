# LockMgr - 锁管理器

LockMgr 提供统一的锁管理功能，支持 Redis 分布式锁和 Memory 本地内存锁，内置完整的可观测性支持。

## 特性

- **多驱动支持**：支持 Redis（分布式锁）和 Memory（本地内存锁）两种驱动
- **统一接口**：提供统一的 ILockManager 接口，便于切换实现
- **可观测性**：内置日志、指标和链路追踪支持
- **自动过期**：支持锁的 TTL 自动过期，防止死锁
- **非阻塞模式**：TryLock 支持非阻塞获取锁
- **依赖注入**：支持通过 `inject:""` 标签注入 Manager
- **健康检查**：内置健康检查接口，便于监控

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
	mgr := lockmgr.NewLockManagerMemoryImpl(nil, nil, &lockmgr.MemoryLockConfig{
		MaxBackups: 1000,
	})
	defer mgr.Close()

	// 获取锁
	err := mgr.Lock(ctx, "resource:123", 10*time.Second)
	if err != nil {
		panic(err)
	}
	defer mgr.Unlock(ctx, "resource:123")

	// 执行需要加锁的操作
}
```

## 支持的锁驱动

### Memory 驱动

基于 sync.Mutex 实现的本地内存锁，适用于单机环境。

**特点**：
- 高性能，无网络开销
- 仅适用于单机进程
- 锁在内存中维护，进程重启后丢失

**适用场景**：
- 单机应用
- 单元测试
- 不需要分布式协调的场景

### Redis 驱动

基于 Redis SET NX EX 命令实现的分布式锁，适用于分布式环境。

**特点**：
- 支持分布式环境
- 依赖 cachemgr.ICacheManager
- 锁键格式：`lock:{key}`
- Lock 方法内部自动重试（50ms 间隔）

**适用场景**：
- 微服务架构
- 多实例部署
- 需要跨进程锁的场景

## API 说明

### ILockManager 接口

```go
type ILockManager interface {
	common.IBaseManager

	// Lock 获取锁（阻塞直到成功或上下文取消）
	// ctx: 上下文
	// key: 锁的键
	// ttl: 锁的过期时间，0表示不过期
	Lock(ctx context.Context, key string, ttl time.Duration) error

	// Unlock 释放锁
	// ctx: 上下文
	// key: 锁的键
	Unlock(ctx context.Context, key string) error

	// TryLock 尝试获取锁（非阻塞）
	// ctx: 上下文
	// key: 锁的键
	// ttl: 锁的过期时间，0表示不过期
	// 返回: 成功返回 true，失败返回 false
	TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
}
```

### Lock - 阻塞获取锁

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

**注意事项**：
- Memory 驱动使用 sync.Mutex，会阻塞当前 goroutine
- Redis 驱动每 50ms 重试一次
- 建议使用 defer 确保锁一定会被释放

### Unlock - 释放锁

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

### TryLock - 非阻塞获取锁

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
import "github.com/lite-lake/litecore-go/manager/lockmgr"

// 创建内存锁
mgr, err := lockmgr.Build("memory", map[string]any{
	"max_backups": 1000,
}, loggerMgr, telemetryMgr, nil)

// 创建 Redis 锁
mgr, err := lockmgr.Build("redis", map[string]any{
	"host":              "localhost",
	"port":              6379,
	"password":          "",
	"db":                0,
	"max_idle_conns":    10,
	"max_open_conns":    100,
	"conn_max_lifetime": "30s",
}, loggerMgr, telemetryMgr, cacheMgr)
```

### BuildWithConfigProvider - 从配置创建

```go
mgr, err := lockmgr.BuildWithConfigProvider(
	configProvider,
	loggerMgr,
	telemetryMgr,
	cacheMgr,
)
```

## 配置

### Redis 驱动配置

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

**配置参数**：
| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| host | string | localhost | Redis 主机地址 |
| port | int | 6379 | Redis 端口 |
| password | string | "" | Redis 密码 |
| db | int | 0 | Redis 数据库编号 |
| max_idle_conns | int | 10 | 最大空闲连接数 |
| max_open_conns | int | 100 | 最大打开连接数 |
| conn_max_lifetime | duration | 30s | 连接最大存活时间 |

### Memory 驱动配置

```yaml
lock:
  driver: memory
  memory_config:
    max_backups: 1000
```

**配置参数**：
| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| max_backups | int | 1000 | 最大备份项数 |

## 依赖注入

```go
type MyService struct {
	LockMgr lockmgr.ILockManager `inject:""`
	logger   loggermgr.ILogger
}

func (s *MyService) ProcessResource(ctx context.Context, resourceID string) error {
	if s.logger == nil && s.LockMgr != nil {
		s.logger = s.LockMgr.Ins()
	}

	// 获取锁
	key := fmt.Sprintf("resource:%s", resourceID)
	err := s.LockMgr.Lock(ctx, key, 10*time.Second)
	if err != nil {
		s.logger.Error("获取锁失败", "resource_id", resourceID, "error", err)
		return err
	}
	defer s.LockMgr.Unlock(ctx, key)

	// 执行业务逻辑
	return nil
}
```

## 可观测性

LockMgr 内置完整的可观测性支持：

### 指标

| 指标名称 | 类型 | 说明 |
|----------|------|------|
| lock.acquire | Counter | 锁获取次数 |
| lock.release | Counter | 锁释放次数 |
| lock.acquire_failed | Counter | 锁获取失败次数 |
| lock.operation.duration | Histogram | 锁操作耗时（秒） |

**指标标签**：
- `lock.driver`: 驱动类型（redis/memory）
- `operation`: 操作类型（lock/unlock/trylock）
- `status`: 状态（success/error）

### 日志

记录锁操作的成功和失败事件，包含操作类型、锁键（已脱敏）、耗时等信息。

### 链路追踪

支持 OpenTelemetry 链路追踪，每个锁操作都会创建一个 span。

## 错误处理

所有锁操作都可能返回错误，常见错误包括：

| 错误类型 | 说明 | 处理建议 |
|----------|------|----------|
| context canceled | 上下文被取消 | 检查上下文超时设置 |
| context deadline exceeded | 操作超时 | 增加 TTL 或优化业务逻辑 |
| cache manager not injected | Redis 驱动未注入缓存管理器 | 注入 cachemgr.ICacheManager |
| lock key cannot be empty | 锁键为空 | 检查 key 参数 |
| context cannot be nil | 上下文为 nil | 传递有效的 context.Context |

## 使用场景

1. **分布式环境下的资源互斥访问**：防止多个进程同时修改同一资源
2. **并发控制**：防止重复操作，如防止重复提交表单
3. **任务队列**：确保任务只能被一个消费者处理
4. **限流控制**：配合令牌桶等算法实现限流
5. **数据库更新防并发**：防止并发更新导致数据不一致

## 最佳实践

1. **使用 defer 确保锁释放**：在获取锁后立即使用 defer 释放锁
2. **设置合理的 TTL**：根据业务场景设置合适的 TTL，防止死锁
3. **选择合适的驱动**：分布式环境使用 Redis 驱动，单机环境使用 Memory 驱动
4. **锁粒度控制**：锁的粒度应尽可能小，减少锁的持有时间
5. **错误处理**：所有锁操作都应进行错误处理
6. **上下文使用**：始终传递有效的 context.Context，支持超时和取消
7. **锁命名规范**：使用清晰的锁键命名，如 `resource:{id}`、`task:{name}`
8. **避免长时间持有锁**：锁的持有时间应尽可能短，避免阻塞其他操作
