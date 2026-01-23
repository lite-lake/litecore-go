# LimiterMgr

限流管理器，提供统一的限流接口，支持多种驱动实现。

## 功能特性

- **多驱动支持**：Redis（分布式限流）、Memory（本地内存限流）
- **统一接口**：提供 `ILimiterManager` 接口，便于切换实现
- **可观测性**：内置日志、指标和链路追踪支持
- **滑动窗口**：支持时间窗口内的请求数限制
- **剩余查询**：支持查询剩余可访问次数

## 驱动类型

### Redis 驱动
- 使用 Redis 实现分布式限流
- 适合多实例部署场景
- 支持高并发请求

### Memory 驱动
- 使用本地内存实现限流
- 高性能，低延迟
- 仅适用于单实例场景

### 默认驱动
- 默认使用 Memory 驱动
- 适用于大多数单实例应用场景

## 配置说明

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

## 使用示例

### 基本用法

```go
// 创建限流管理器
mgr := limitermgr.NewLimiterManagerMemoryImpl(&limitermgr.MemoryLimiterConfig{})
defer mgr.Close()

ctx := context.Background()

// 检查是否允许通过（100次/分钟）
allowed, err := mgr.Allow(ctx, "user:123", 100, time.Minute)
if err != nil {
    log.Fatal(err)
}

if !allowed {
    // 请求被限流
    return
}

// 处理请求
// ...
```

### 查询剩余次数

```go
// 获取剩余可访问次数
remaining, err := mgr.GetRemaining(ctx, "user:123", 100, time.Minute)
if err != nil {
    log.Fatal(err)
}

// 返回剩余次数
log.Printf("剩余可访问次数: %d", remaining)
```

### 中间件集成

```go
func RateLimitMiddleware(limiter limitermgr.ILimiterManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 使用 IP 作为限流键
        key := fmt.Sprintf("ip:%s", c.ClientIP())

        // 检查是否允许通过（100次/分钟）
        allowed, err := limiter.Allow(c.Request.Context(), key, 100, time.Minute)
        if err != nil {
            c.JSON(500, gin.H{"error": "限流服务异常"})
            c.Abort()
            return
        }

        if !allowed {
            c.JSON(429, gin.H{"error": "请求过于频繁，请稍后再试"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

## 限流策略

### 固定窗口
- 时间窗口内请求数超过限制则拒绝
- 窗口结束后计数重置

### 适用场景
- API 接口请求频率限制
- 用户行为频率控制（如点赞、评论）
- 防止恶意请求和爬虫
- 资源使用配额管理

## 可观测性

### 日志
- Debug: 限流操作成功
- Error: 限流操作失败
- Warn: 限流触发（可选）

### 指标
- `limiter.allowed`: 限流允许通过次数
- `limiter.rejected`: 限流拒绝次数
- `limiter.operation.duration`: 限流操作耗时

### 链路追踪
- 记录限流操作的完整链路
- 包含限流键、驱动类型、操作结果等信息
