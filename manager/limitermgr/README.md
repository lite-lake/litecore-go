# Limiter Manager (限流管理器)

提供统一的限流管理功能，支持 Redis 和 Memory 两种限流驱动。

## 特性

- **多驱动支持** - Redis（分布式限流）和 Memory（本地内存限流）
- **统一接口** - ILimiterManager 接口，便于切换实现
- **可观测性** - 内置日志、指标和链路追踪
- **滑动窗口** - Memory 实现使用滑动窗口算法
- **固定窗口** - Redis 实现使用固定窗口算法
- **剩余查询** - 支持查询剩余可访问次数

## 快速开始

```go
import (
    "context"
    "time"

    "github.com/lite-lake/litecore-go/manager/limitermgr"
)

// 使用内存限流
mgr := limitermgr.NewLimiterManagerMemoryImpl()
defer mgr.Close()

ctx := context.Background()

// 检查是否允许通过（100次/分钟）
allowed, err := mgr.Allow(ctx, "user:123", 100, time.Minute)
if err != nil {
    log.Fatal(err)
}
if !allowed {
    log.Println("请求被限流")
}

// 获取剩余次数
remaining, err := mgr.GetRemaining(ctx, "user:123", 100, time.Minute)
if err != nil {
    log.Fatal(err)
}
log.Printf("剩余可访问次数: %d", remaining)
```

## Memory 驱动

Memory 驱动使用滑动窗口算法，适用于单实例场景。

### 配置

```go
type MemoryLimiterConfig struct {
    MaxBackups int // 最大备份项数（清理策略相关）
}
```

### 创建实例

```go
mgr := limitermgr.NewLimiterManagerMemoryImpl()
```

### 使用场景

- 单机应用限流
- 测试环境
- 无需 Redis 的场景

## Redis 驱动

Redis 驱动使用固定窗口算法，适用于分布式场景。

### 配置

```go
type RedisLimiterConfig struct {
    Host            string        // Redis 主机地址（默认 localhost）
    Port            int           // Redis 端口（默认 6379）
    Password        string        // Redis 密码
    DB              int           // Redis 数据库编号（默认 0）
    MaxIdleConns    int           // 最大空闲连接数（默认 10）
    MaxOpenConns    int           // 最大打开连接数（默认 100）
    ConnMaxLifetime time.Duration // 连接最大存活时间（默认 30秒）
}
```

### 创建实例

```go
mgr := limitermgr.NewLimiterManagerRedisImpl()
```

### 依赖

Redis 驱动依赖 `cachemgr.ICacheManager`，通过依赖注入自动初始化。

### 使用场景

- 分布式应用限流
- 多实例共享限流状态
- 生产环境推荐

## 配置文件

```yaml
limiter:
  driver: "redis"  # 驱动类型: redis, memory

  # Redis 配置（当 driver=redis 时使用）
  redis_config:
    host: "localhost"
    port: 6379
    password: ""
    db: 0
    max_idle_conns: 10
    max_open_conns: 100
    conn_max_lifetime: 30s

  # Memory 配置（当 driver=memory 时使用）
  memory_config:
    max_backups: 1000
```

## 工厂方法

### Build

根据驱动类型和配置创建实例。

```go
mgr, err := limitermgr.Build("memory", map[string]any{
    "max_backups": 1000,
})

mgr, err := limitermgr.Build("redis", map[string]any{
    "host":     "localhost",
    "port":     6379,
    "password": "",
    "db":       0,
})
```

### BuildWithConfigProvider

从配置提供者创建实例，自动读取 `limiter.driver` 和对应驱动配置。

```go
mgr, err := limitermgr.BuildWithConfigProvider(configProvider)
```

## API

### ILimiterManager 接口

```go
type ILimiterManager interface {
    common.IBaseManager

    // Allow 检查是否允许通过限流
    Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)

    // GetRemaining 获取剩余可访问次数
    GetRemaining(ctx context.Context, key string, limit int, window time.Duration) (int, error)
}
```

### Allow 方法

检查是否允许通过限流。

参数：
- `ctx`: 上下文
- `key`: 限流键（如用户ID、IP等）
- `limit`: 时间窗口内的最大请求数
- `window`: 时间窗口大小

返回：
- `bool`: 允许返回 true，否则返回 false
- `error`: 错误信息

### GetRemaining 方法

获取剩余可访问次数。

参数：
- `ctx`: 上下文
- `key`: 限流键
- `limit`: 时间窗口内的最大请求数
- `window`: 时间窗口大小

返回：
- `int`: 剩余次数
- `error`: 错误信息

## 使用场景

### API 接口限流

```go
// 每 IP 每分钟最多 100 次请求
allowed, err := mgr.Allow(ctx, "ip:"+clientIP, 100, time.Minute)
if !allowed {
    return "请求过于频繁，请稍后再试", 429
}
```

### 用户行为限流

```go
// 每用户每小时最多 10 次点赞
allowed, err := mgr.Allow(ctx, "user_like:"+userID, 10, time.Hour)
if !allowed {
    return "点赞过于频繁，请稍后再试", 429
}
```

### 分布式限流

```go
// 使用 Redis 实现分布式限流
mgr := limitermgr.NewLimiterManagerRedisImpl()

// 每个节点共享限流状态
allowed, err := mgr.Allow(ctx, "api_endpoint:/api/message", 1000, time.Minute)
```

## 中间件集成

limitermgr 与 `litemiddleware.RateLimiterMiddleware` 集成，提供开箱即用的限流中间件。

```go
import (
    "time"
    "github.com/lite-lake/litecore-go/component/litemiddleware"
)

// 创建限流中间件
limit := 100
window := time.Minute
keyPrefix := "ip"
middleware := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
    Limit:     &limit,
    Window:    &window,
    KeyPrefix: &keyPrefix,
})

// 注册中间件
middlewareContainer.RegisterMiddleware(middleware)
```

参考 messageboard 示例中的使用：`samples/messageboard/internal/middlewares/rate_limiter_middleware.go`

## 算法对比

| 驱动 | 算法 | 场景 | 分布式 |
|------|------|------|--------|
| Memory | 滑动窗口 | 单机应用 | 否 |
| Redis | 固定窗口 | 分布式应用 | 是 |

## 依赖注入

limitermgr 支持依赖注入，自动注入以下组件：

- `loggermgr.ILoggerManager` - 日志管理器
- `telemetrymgr.ITelemetryManager` - 可观测性管理器
- `cachemgr.ICacheManager` - 缓存管理器（仅 Redis 驱动需要）

示例：

```go
type MyService struct {
    LimiterMgr limitermgr.ILimiterManager `inject:""`
    logger     logger.ILogger
}

func (s *MyService) initLogger() {
    // limitermgr 内部已实现日志初始化
}
```

## 错误处理

```go
allowed, err := mgr.Allow(ctx, "key", 10, time.Minute)
if err != nil {
    // 参数验证失败或 Redis 连接错误
    return fmt.Errorf("限流检查失败: %w", err)
}

if !allowed {
    // 请求被限流
    return errors.New("请求过于频繁")
}
```

## 可观测性

### 日志

limitermgr 自动记录限流操作日志：

- Debug 级别：操作成功
- Error 级别：操作失败

### 指标

limitermgr 收集以下指标：

- `limiter.allowed`: 限流允许通过次数
- `limiter.rejected`: 限流拒绝次数
- `limiter.operation.duration`: 限流操作耗时（秒）

### 链路追踪

limitermgr 自动记录操作链路，包含以下属性：

- `limiter.key`: 限流键（脱敏）
- `limiter.driver`: 限流驱动类型

## 最佳实践

1. **选择合适的驱动**
   - 单机应用使用 Memory
   - 分布式应用使用 Redis

2. **限流键设计**
   - IP 限流：`"ip:" + clientIP`
   - 用户限流：`"user:" + userID`
   - 接口限流：`"api:" + path`

3. **时间窗口选择**
   - 分钟级：`time.Minute`
   - 小时级：`time.Hour`
   - 天级：`time.Hour * 24`

4. **限流阈值设置**
   - 根据业务需求设置合理阈值
   - 避免设置过严导致用户体验差
   - 建议从松到紧逐步调整

5. **降级处理**
   - 限流器故障时建议降级放行
   - 记录降级日志便于排查
