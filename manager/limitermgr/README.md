# Limiter Manager (限流管理器)

提供统一的限流管理功能，支持 Redis 和 Memory 两种限流驱动。

## 特性

- **多驱动支持** - Redis（分布式限流）、Memory（本地内存限流）
- **统一接口** - ILimiterManager 接口，便于切换限流实现
- **可观测性** - 内置日志、指标和链路追踪支持
- **滑动窗口** - Memory 实现使用滑动窗口算法，精确统计时间窗口内的请求
- **固定窗口** - Redis 实现使用固定窗口算法，适用于分布式场景
- **剩余查询** - 支持查询剩余可访问次数

## 快速开始

```go
import (
    "context"
    "time"

    "github.com/lite-lake/litecore-go/manager/limitermgr"
)

// 使用内存限流（单机场景）
mgr := limitermgr.NewLimiterManagerMemoryImpl(nil, nil)
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

### 算法原理

滑动窗口算法维护一个时间窗口内的所有请求时间戳，每次请求时：
1. 清理时间窗口外的过期请求
2. 统计当前窗口内的请求数
3. 未达到阈值则允许通过，否则拒绝

### 配置

```go
type MemoryLimiterConfig struct {
    MaxBackups int // 最大备份项数（清理策略相关）
}
```

### 创建实例

```go
import (
    "github.com/lite-lake/litecore-go/manager/loggermgr"
    "github.com/lite-lake/litecore-go/manager/telemetrymgr"
)

// 创建内存限流管理器
mgr := limitermgr.NewLimiterManagerMemoryImpl(loggerMgr, telemetryMgr)
defer mgr.Close()
```

### 使用场景

- 单机应用限流
- 测试环境
- 无需 Redis 的场景
- 需要精确限流的场景

## Redis 驱动

Redis 驱动使用固定窗口算法，适用于分布式场景。

### 算法原理

固定窗口算法使用 Redis 计数器实现：
1. 每次请求时增加计数器（使用 Redis INCR 命令）
2. 第一次增加时设置过期时间
3. 判断计数器是否超过阈值

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
import (
    "github.com/lite-lake/litecore-go/manager/cachemgr"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
    "github.com/lite-lake/litecore-go/manager/telemetrymgr"
)

// 创建 Redis 限流管理器
mgr := limitermgr.NewLimiterManagerRedisImpl(loggerMgr, telemetryMgr, cacheMgr)
defer mgr.Close()
```

### 依赖

Redis 驱动依赖 `cachemgr.ICacheManager`，通过依赖注入自动初始化。

### 使用场景

- 分布式应用限流
- 多实例共享限流状态
- 生产环境推荐
- 对限流精度要求不高的场景

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
// 使用配置映射创建
mgr, err := limitermgr.Build("memory", map[string]any{
    "max_backups": 1000,
}, loggerMgr, telemetryMgr, nil)

mgr, err := limitermgr.Build("redis", map[string]any{
    "host":     "localhost",
    "port":     6379,
    "password": "",
    "db":       0,
}, loggerMgr, telemetryMgr, cacheMgr)
```

### BuildWithConfigProvider

从配置提供者创建实例，自动读取 `limiter.driver` 和对应驱动配置。

```go
mgr, err := limitermgr.BuildWithConfigProvider(configProvider, loggerMgr, telemetryMgr, cacheMgr)
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

**参数：**
- `ctx`: 上下文
- `key`: 限流键（如用户ID、IP等）
- `limit`: 时间窗口内的最大请求数
- `window`: 时间窗口大小

**返回：**
- `bool`: 允许返回 true，否则返回 false
- `error`: 错误信息

**示例：**
```go
allowed, err := mgr.Allow(ctx, "user:123", 100, time.Minute)
if err != nil {
    return fmt.Errorf("限流检查失败: %w", err)
}
if !allowed {
    return errors.New("请求过于频繁，请稍后再试")
}
```

### GetRemaining 方法

获取剩余可访问次数。

**参数：**
- `ctx`: 上下文
- `key`: 限流键
- `limit`: 时间窗口内的最大请求数
- `window`: 时间窗口大小

**返回：**
- `int`: 剩余次数
- `error`: 错误信息

**示例：**
```go
remaining, err := mgr.GetRemaining(ctx, "user:123", 100, time.Minute)
if err != nil {
    return fmt.Errorf("获取剩余次数失败: %w", err)
}
fmt.Printf("剩余可访问次数: %d\n", remaining)
```

## 限流策略

### 固定窗口算法（Redis）

**原理：**
- 将时间划分为固定大小的窗口
- 每个窗口内维护一个计数器
- 每次请求时增加计数器，超过阈值则拒绝

**优点：**
- 实现简单，性能高
- 内存占用少
- 适合分布式场景

**缺点：**
- 窗口边界处可能出现流量突刺
- 限流精度相对较低

**适用场景：**
- API 接口限流
- 防止恶意请求
- 对限流精度要求不高的场景

### 滑动窗口算法（Memory）

**原理：**
- 维护一个时间窗口内的所有请求时间戳
- 每次请求时清理过期请求，统计有效请求数
- 动态滑动窗口，精确限流

**优点：**
- 限流精度高
- 避免边界问题
- 平滑流量

**缺点：**
- 内存占用较大
- 性能相对较低
- 不适合分布式场景

**适用场景：**
- 单机应用限流
- 用户行为限流（如点赞、评论）
- 需要精确限流的场景

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

### 接口级别限流

```go
// 每个 API 端点每秒最多 100 次请求
allowed, err := mgr.Allow(ctx, "api:/api/message", 100, time.Second)
if !allowed {
    return "系统繁忙，请稍后再试", 503
}
```

### 分布式限流

```go
// 使用 Redis 实现分布式限流
mgr := limitermgr.NewLimiterManagerRedisImpl(loggerMgr, telemetryMgr, cacheMgr)

// 每个节点共享限流状态
allowed, err := mgr.Allow(ctx, "global:api_endpoint", 1000, time.Minute)
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

| 驱动 | 算法 | 精度 | 性能 | 分布式 | 适用场景 |
|------|------|------|------|--------|----------|
| Memory | 滑动窗口 | 高 | 中 | 否 | 单机应用、精确限流 |
| Redis | 固定窗口 | 中 | 高 | 是 | 分布式应用、API 限流 |

## 依赖注入

limitermgr 支持依赖注入，自动注入以下组件：

- `loggermgr.ILoggerManager` - 日志管理器
- `telemetrymgr.ITelemetryManager` - 可观测性管理器
- `cachemgr.ICacheManager` - 缓存管理器（仅 Redis 驱动需要）

**示例：**

```go
type MyService struct {
    LimiterMgr limitermgr.ILimiterManager `inject:""`
    logger     logger.ILogger
}

func (s *MyService) initLogger() {
    if s.LimiterMgr != nil && s.LimiterMgr.GetLoggerMgr() != nil {
        s.logger = s.LimiterMgr.GetLoggerMgr().Ins()
    }
}

func (s *MyService) SomeMethod() {
    s.initLogger()
    allowed, err := s.LimiterMgr.Allow(context.Background(), "user:123", 100, time.Minute)
    if err != nil {
        s.logger.Error("限流检查失败", "error", err)
        return
    }
    if !allowed {
        s.logger.Warn("请求被限流")
        return
    }
    // 处理业务逻辑
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

**常见错误：**
- `context cannot be nil` - 上下文为空
- `limiter key cannot be empty` - 限流键为空
- `limit must be greater than 0` - 限流阈值无效
- `window must be greater than 0` - 时间窗口无效
- `cache manager is not initialized` - Redis 驱动未初始化缓存管理器

## 可观测性

### 日志

limitermgr 自动记录限流操作日志：

- Debug 级别：操作成功
- Error 级别：操作失败
- Warn 级别：中间件中记录限流拒绝

**日志示例：**
```go
// 操作成功
2026-01-24 15:04:05.123 | DEBUG | limiter operation success | operation=allow key="user:***" duration=0.001

// 操作失败
2026-01-24 15:04:05.456 | ERROR | limiter operation failed | operation=allow key="user:***" error="context canceled" duration=0.002
```

### 指标

limitermgr 收集以下指标：

- `limiter.allowed` - 限流允许通过次数
- `limiter.rejected` - 限流拒绝次数
- `limiter.operation.duration` - 限流操作耗时（秒）

**指标属性：**
- `limiter.driver` - 限流驱动类型
- `operation` - 操作名称
- `status` - 操作状态

### 链路追踪

limitermgr 自动记录操作链路，包含以下属性：

- `limiter.key` - 限流键（脱敏，只显示前5个字符）
- `limiter.driver` - 限流驱动类型

**Span 名称：**
- `limiter.allow` - Allow 操作
- `limiter.get_remaining` - GetRemaining 操作

## 最佳实践

### 选择合适的驱动

1. **单机应用使用 Memory**
   - 无需外部依赖
   - 限流精度高
   - 性能良好

2. **分布式应用使用 Redis**
   - 多实例共享限流状态
   - 高性能
   - 适合生产环境

### 限流键设计

1. **IP 限流：** `"ip:" + clientIP`
2. **用户限流：** `"user:" + userID`
3. **接口限流：** `"api:" + path`
4. **组合限流：** `"user:" + userID + ":" + action`

### 时间窗口选择

1. **分钟级：** `time.Minute` - API 限流
2. **小时级：** `time.Hour` - 用户行为限流
3. **天级：** `time.Hour * 24` - 资源配额管理

### 限流阈值设置

1. 根据业务需求设置合理阈值
2. 避免设置过严导致用户体验差
3. 建议从松到紧逐步调整
4. 通过监控指标动态调整

### 降级处理

1. 限流器故障时建议降级放行
2. 记录降级日志便于排查
3. 避免因限流器故障导致服务不可用

### 中间件配置

```go
// 配置跳过限流的条件（如白名单 IP）
middleware := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
    Limit:  &limit,
    Window: &window,
    SkipFunc: func(c *gin.Context) bool {
        return c.ClientIP() == "127.0.0.1"
    },
})

// 自定义 key 生成函数（如按用户限流）
middleware := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
    Limit:  &limit,
    Window: &window,
    KeyFunc: func(c *gin.Context) string {
        return c.GetString("userID")
    },
})
```
