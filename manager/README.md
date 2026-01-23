# Manager 组件

Manager 组件是 litecore-go 的基础能力层，提供可观测、日志、缓存、数据库、锁、限流器、队列等核心功能。每个 Manager 组件都是独立的模块，通过依赖注入的方式集成到应用中。

## 目录结构

```
manager/
├── README.md          # 本文档
├── cachemgr/          # 缓存管理器
├── configmgr/         # 配置管理器
├── databasemgr/       # 数据库管理器
├── limitermgr/        # 限流管理器
├── lockmgr/           # 锁管理器
├── loggermgr/         # 日志管理器
├── mqmgr/             # 消息队列管理器
└── telemetrymgr/      # 可观测性管理器
```

## Manager 组件概览

### 1. cachemgr - 缓存管理器

提供统一的缓存操作接口，支持多种缓存驱动。

**支持的驱动：**
- `memory` - 基于 Ristretto 的高性能内存缓存
- `redis` - Redis 分布式缓存
- `none` - 禁用缓存

**核心功能：**
- 基本 CRUD 操作（Get、Set、Delete）
- 批量操作（GetMultiple、SetMultiple、DeleteMultiple）
- TTL 管理（Expire、TTL）
- 原子操作（SetNX、Increment、Decrement）
- 键存在性检查（Exists）

**接口：** `ICacheManager`

---

### 2. configmgr - 配置管理器

提供统一的配置加载和访问接口，支持路径查询（如 `aaa.bbb.ccc`）。

**支持的格式：**
- YAML
- JSON

**核心功能：**
- 配置加载和解析
- 路径查询支持
- 配置项存在性检查

**接口：** `IConfigManager`

---

### 3. databasemgr - 数据库管理器

基于 GORM 的数据库管理器，支持多种数据库驱动。

**支持的驱动：**
- MySQL
- PostgreSQL
- SQLite
- `none` - 禁用数据库

**核心功能：**
- GORM 完整支持（DB、Model、Table、WithContext）
- 事务管理（Transaction、Begin）
- 自动迁移（AutoMigrate、Migrator）
- 连接池管理（Ping、Stats）
- 原生 SQL 执行（Exec、Raw）

**接口：** `IDatabaseManager`

---

### 4. limitermgr - 限流管理器

提供限流功能，支持基于时间窗口的请求频率控制。

**支持的驱动：**
- `memory` - 内存限流
- `redis` - Redis 分布式限流

**核心功能：**
- 请求限流检查（Allow）
- 剩余配额查询（GetRemaining）

**使用场景：**
- API 限流
- 防止暴力破解
- 资源访问控制

**接口：** `ILimiterManager`

---

### 5. lockmgr - 锁管理器

提供分布式锁功能，支持阻塞和非阻塞锁获取。

**支持的驱动：**
- `memory` - 内存锁
- `redis` - Redis 分布式锁

**核心功能：**
- 阻塞锁获取（Lock）
- 非阻塞锁尝试（TryLock）
- 锁释放（Unlock）
- TTL 支持（自动过期）

**使用场景：**
- 分布式任务调度
- 资源互斥访问
- 幂等性控制

**接口：** `ILockManager`

---

### 6. loggermgr - 日志管理器

提供结构化日志功能，支持多种格式和输出目标。

**支持的驱动：**
- `zap` - Uber Zap 日志驱动
- `default` - 默认 ConsoleEncoder
- `none` - 禁用日志

**支持的格式：**
- `gin` - Gin 风格格式（竖线分隔符，适合控制台输出）
- `json` - JSON 格式（适合日志分析和监控）
- `default` - 默认 ConsoleEncoder 格式

**核心功能：**
- 多级别日志（Debug、Info、Warn、Error、Fatal）
- 结构化日志字段
- 控制台输出（支持颜色）
- 文件输出（支持日志轮转）
- 敏感信息脱敏

**接口：** `ILoggerManager`

**Gin 格式特点：**
- 统一格式：`{时间} | {级别} | {消息} | {字段1}={值1} {字段2}={值2} ...`
- 时间固定宽度 23 字符
- 级别固定宽度 5 字符，右对齐，带颜色
- 字段格式：`key=value`

---

### 7. mqmgr - 消息队列管理器

提供消息队列功能，支持异步消息处理。

**支持的驱动：**
- `rabbitmq` - RabbitMQ
- `memory` - 内存队列（用于开发和测试）

**核心功能：**
- 消息发布（Publish）
- 消息订阅（Subscribe、SubscribeWithCallback）
- 消息确认（Ack、Nack）
- 队列管理（QueueLength、Purge）

**接口：** `IMQManager`

---

### 8. telemetrymgr - 可观测性管理器

统一提供 Traces、Metrics、Logs 三大观测能力。

**核心功能：**
- Tracing - 分布式追踪（Tracer、TracerProvider）
- Metrics - 指标收集（Meter、MeterProvider）
- Logging - 结构化日志（Logger、LoggerProvider）

**接口：** `ITelemetryManager`

---

## 依赖关系

```
configmgr
    ↓
    ├── loggermgr
    ├── cachemgr
    │       └── configmgr
    ├── databasemgr
    │       └── configmgr
    ├── limitermgr
    │       ├── configmgr
    │       └── cachemgr (可选)
    ├── lockmgr
    │       ├── configmgr
    │       └── cachemgr (可选)
    ├── mqmgr
    │       └── configmgr
    ├── telemetrymgr
    │       └── configmgr
    └── loggermgr
            └── configmgr
```

## 配置示例

```yaml
# config.yaml
cache:
  driver: "memory"
  memory_config:
    max_size: 100
    max_age: 720h  # 30 days

database:
  driver: "mysql"
  mysql_config:
    host: "localhost"
    port: 3306
    database: "mydb"
    username: "root"
    password: "password"

limiter:
  driver: "memory"
  memory_config:
    max_backups: 1000

lock:
  driver: "redis"
  redis_config:
    host: "localhost"
    port: 6379
    db: 0

logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "info"
      format: "gin"      # gin | json | default
      color: true
      time_format: "2006-01-24 15:04:05.000"
    file_enabled: true
    file_config:
      level: "info"
      path: "./logs/app.log"
      rotation:
        max_size: 100
        max_age: 30
        max_backups: 10
        compress: true

mq:
  driver: "rabbitmq"
  rabbitmq_config:
    url: "amqp://guest:guest@localhost:5672/"
    durable: true

telemetry:
  driver: "otel"
  otel_config:
    endpoint: "http://localhost:4318"
    enabled_traces: true
    enabled_metrics: true
    enabled_logs: true
```

## 使用方式

### 依赖注入

```go
type MyService struct {
    CacheMgr   cachemgr.ICacheManager   `inject:""`
    DBMgr      databasemgr.IDatabaseManager `inject:""`
    LoggerMgr  loggermgr.ILoggerManager  `inject:""`
}
```

### 初始化日志记录器

```go
func (s *MyService) initLogger() {
    if s.LoggerMgr != nil {
        s.logger = s.LoggerMgr.Ins()
    }
}
```

### 日志记录

```go
s.logger.Info("操作开始", "param", value)
s.logger.Warn("慢查询检测", "duration", 1.2*time.Second)
s.logger.Error("数据库连接失败", "error", err)
```

### 缓存操作

```go
// 设置缓存
err := s.CacheMgr.Set(ctx, "user:123", user, 10*time.Minute)

// 获取缓存
err := s.CacheMgr.Get(ctx, "user:123", &user)
```

### 数据库操作

```go
// 查询
result := s.DBMgr.WithContext(ctx).Where("id = ?", id).First(&user)

// 事务
err := s.DBMgr.Transaction(func(tx *gorm.DB) error {
    return tx.Create(&user).Error
})
```

### 限流

```go
allowed, err := s.LimiterMgr.Allow(ctx, "user:123", 100, time.Minute)
if !allowed {
    return errors.New("请求过于频繁")
}
```

### 锁

```go
// 尝试获取锁
ok, err := s.LockMgr.TryLock(ctx, "resource:123", 10*time.Second)
if !ok {
    return errors.New("资源已被占用")
}
defer s.LockMgr.Unlock(ctx, "resource:123")

// 执行业务逻辑
```

## 日志级别

| 级别 | 说明 |
|------|------|
| Debug | 开发调试信息 |
| Info | 正常业务流程（请求开始/完成、资源创建） |
| Warn | 降级处理、慢查询、重试 |
| Error | 业务错误、操作失败（需人工关注） |
| Fatal | 致命错误，需要立即终止 |

## 日志格式

### Gin 格式（推荐用于控制台）

```
2026-01-24 15:04:05.123 | INFO  | 开始依赖注入 | count=23
2026-01-24 15:04:05.456 | WARN  | 慢查询检测 | duration=1.2s
2026-01-24 15:04:05.789 | ERROR | 数据库连接失败 | error="connection refused"
```

### JSON 格式（推荐用于日志分析）

```json
{"time":"2026-01-24T15:04:05.123Z","level":"INFO","msg":"开始依赖注入","count":23}
```

## 敏感信息处理

日志管理器会自动过滤和脱敏敏感信息：
- 密码、token、密钥等必须脱敏
- 支持内置过滤规则或自定义脱敏函数
