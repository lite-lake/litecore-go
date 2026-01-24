# Manager 组件

Manager 组件是 litecore-go 的基础能力层，提供配置管理、日志、缓存、数据库、锁、限流器、消息队列、可观测性、定时任务等核心功能。所有 Manager 组件都继承自 `common.IBaseManager` 接口，通过依赖注入的方式集成到应用中。

## 目录结构

```
manager/
├── README.md           # 本文档
├── configmgr/          # 配置管理器
├── telemetrymgr/       # 可观测性管理器
├── loggermgr/          # 日志管理器
├── databasemgr/        # 数据库管理器
├── cachemgr/           # 缓存管理器
├── lockmgr/            # 锁管理器
├── limitermgr/         # 限流管理器
├── mqmgr/              # 消息队列管理器
└── schedulermgr/       # 定时任务管理器
```

## 统一接口规范

### IBaseManager 基础接口

所有 Manager 组件都实现 `common.IBaseManager` 接口：

```go
type IBaseManager interface {
    ManagerName() string    // 返回管理器名称
    Health() error          // 检查健康状态
    OnStart() error         // 服务器启动时触发
    OnStop() error          // 服务器停止时触发
}
```

### 可观测性支持

大部分 Manager 组件都提供可观测性支持（日志、指标、链路追踪）：

```go
type BaseManagerImpl struct {
    LoggerMgr    loggermgr.ILoggerManager    // 日志管理器
    TelemetryMgr telemetrymgr.ITelemetryManager // 遥测管理器
    Tracer       trace.Tracer                 // 链路追踪器
    Meter        metric.Meter                 // 指标记录器
}
```

## 可用 Manager 列表

| Manager | 包路径 | 接口 | 功能 | 支持驱动 |
|---------|--------|------|------|----------|
| ConfigManager | `manager/configmgr` | `IConfigManager` | 配置加载和路径查询 | yaml、json |
| TelemetryManager | `manager/telemetrymgr` | `ITelemetryManager` | Traces、Metrics、Logs | otel、none |
| LoggerManager | `manager/loggermgr` | `ILoggerManager` | 结构化日志 | zap、default、none |
| DatabaseManager | `manager/databasemgr` | `IDatabaseManager` | GORM 数据库操作 | mysql、postgresql、sqlite、none |
| CacheManager | `manager/cachemgr` | `ICacheManager` | 缓存操作 | redis、memory、none |
| LockManager | `manager/lockmgr` | `ILockManager` | 分布式锁 | redis、memory |
| LimiterManager | `manager/limitermgr` | `ILimiterManager` | 请求限流 | redis、memory |
| MQManager | `manager/mqmgr` | `IMQManager` | 消息队列 | rabbitmq、memory |
| SchedulerManager | `manager/schedulermgr` | `ISchedulerManager` | 定时任务管理 | cron |

## 初始化顺序

Manager 组件按以下顺序初始化，确保依赖关系正确：

```
1. ConfigManager        （必须最先初始化）
    ↓
2. TelemetryManager     （依赖 ConfigManager）
    ↓
3. LoggerManager        （依赖 ConfigManager + TelemetryManager）
    ↓
4. DatabaseManager      （依赖 ConfigManager + LoggerManager + TelemetryManager）
    ↓
5. CacheManager         （依赖 ConfigManager + LoggerManager + TelemetryManager）
    ↓
6. LockManager          （依赖 ConfigManager + LoggerManager + TelemetryManager + CacheManager）
    ↓
7. LimiterManager       （依赖 ConfigManager + LoggerManager + TelemetryManager + CacheManager）
    ↓
8. MQManager            （依赖 ConfigManager + LoggerManager + TelemetryManager）
    ↓
9. SchedulerManager     （依赖 ConfigManager + LoggerManager）
```

## 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 包名 | `<功能名>mgr` | `cachemgr`、`configmgr` |
| 接口 | `I<功能名>Manager` | `ICacheManager`、`IDatabaseManager` |
| 实现类 | `<功能名>Manager<驱动名>Impl` | `cacheManagerRedisImpl` |
| 构造函数 | `New<功能名>Manager<驱动名>Impl` | `NewCacheManagerRedisImpl` |
| 配置结构 | `<功能名>Config`、`<驱动名>Config` | `CacheConfig`、`RedisConfig` |

## 依赖注入方式

### 在 Repository/Service/Controller/Middleware 层注入

使用 `inject:""` 标签注入 Manager：

```go
type MessageService struct {
    ConfigMgr   configmgr.IConfigManager    `inject:""`
    LoggerMgr   loggermgr.ILoggerManager   `inject:""`
    CacheMgr    cachemgr.ICacheManager     `inject:""`
    DBMgr       databasemgr.IDatabaseManager `inject:""`
}
```

### 日志记录模式

LoggerMgr 需要使用 `Ins()` 获取实际的日志实例：

```go
type MessageService struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
    logger    logger.ILogger            // 实际日志实例
}

func (s *MessageService) initLogger() {
    if s.LoggerMgr != nil {
        s.logger = s.LoggerMgr.Ins()
    }
}

func (s *MessageService) SomeMethod() error {
    s.initLogger()
    s.logger.Info("操作开始", "param", value)
    return nil
}
```

## 工厂模式

### Build 函数

直接创建 Manager 实例：

```go
// 创建缓存管理器
cacheMgr, err := cachemgr.Build(
    "redis",
    map[string]any{
        "host": "localhost",
        "port": 6379,
    },
    loggerMgr,
    telemetryMgr,
)

// 创建数据库管理器
dbMgr, err := databasemgr.Build(
    "mysql",
    &databasemgr.MySQLConfig{
        DSN: "root:password@tcp(localhost:3306)/mydb",
    },
    loggerMgr,
    telemetryMgr,
)
```

### BuildWithConfigProvider 函数

从配置提供者创建 Manager 实例：

```go
// 从配置文件创建所有 Manager
configMgr := configmgr.Build("yaml", "configs/config.yaml")
telemetryMgr := telemetrymgr.BuildWithConfigProvider(configMgr)
loggerMgr := loggermgr.BuildWithConfigProvider(configMgr, telemetryMgr)
dbMgr := databasemgr.BuildWithConfigProvider(configMgr, loggerMgr, telemetryMgr)
cacheMgr := cachemgr.BuildWithConfigProvider(configMgr, loggerMgr, telemetryMgr)
```

## 自动初始化

Manager 组件由 Engine 自动初始化和注入：

```go
func main() {
    entityContainer := container.NewEntityContainer()
    repositoryContainer := container.NewRepositoryContainer(entityContainer)
    serviceContainer := container.NewServiceContainer(repositoryContainer)
    controllerContainer := container.NewControllerContainer(serviceContainer)
    middlewareContainer := container.NewMiddlewareContainer(serviceContainer)
    listenerContainer := container.NewListenerContainer()
    schedulerContainer := container.NewSchedulerContainer()

    engine := server.NewEngine(
        &server.BuiltinConfig{
            Driver:   "yaml",
            FilePath: "configs/config.yaml",
        },
        entityContainer,
        repositoryContainer,
        serviceContainer,
        controllerContainer,
        middlewareContainer,
        listenerContainer,
        schedulerContainer,
    )

    engine.Run()
}
```

## 配置示例

```yaml
# 配置管理器
config:
  driver: "yaml"        # 支持 yaml、json

# 遥测管理器
telemetry:
  driver: "otel"        # 支持 otel、none
  otel_config:
    endpoint: "http://localhost:4318"
    service_name: "litecore-app"

# 日志管理器
logger:
  driver: "zap"         # 支持 zap、default、none
  zap_config:
    console_enabled: true
    console_config:
      level: "info"
      format: "gin"      # gin | json | default
      color: true
    file_enabled: true
    file_config:
      filename: "logs/app.log"
      max_size: 100      # MB
      max_age: 30        # days
      max_backups: 10
      compress: true

# 数据库管理器
database:
  driver: "mysql"       # 支持 mysql、postgresql、sqlite、none
  auto_migrate: true
  mysql_config:
    dsn: "root:password@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local"
  observability_config:
    slow_query_threshold: "1s"
    log_sql: true
    sample_rate: 1.0

# 缓存管理器
cache:
  driver: "redis"       # 支持 redis、memory、none
  redis_config:
    host: "localhost"
    port: 6379
    password: ""
    db: 0
  memory_config:
    max_age: "720h"     # 30 days

# 锁管理器
lock:
  driver: "redis"       # 支持 redis、memory
  redis_config:
    host: "localhost"
    port: 6379

# 限流管理器
limiter:
  driver: "memory"      # 支持 redis、memory
  memory_config:
    max_backups: 1000

# 消息队列管理器
mq:
  driver: "rabbitmq"    # 支持 rabbitmq、memory
  rabbitmq_config:
    url: "amqp://guest:guest@localhost:5672/"

# 定时任务管理器
scheduler:
  driver: "cron"        # 支持 cron
```

## 组件概览

### ConfigManager

配置管理器提供统一的配置加载和访问接口，支持路径查询（如 `aaa.bbb.ccc`）。

**核心功能：**
- 配置加载和解析（YAML、JSON）
- 路径查询支持
- 配置项存在性检查

**接口：** `IConfigManager`

```go
type IConfigManager interface {
    common.IBaseManager
    Get(key string) (any, error)    // 获取配置项
    Has(key string) bool             // 检查配置项是否存在
}
```

---

### TelemetryManager

可观测性管理器统一提供 Traces、Metrics、Logs 三大观测能力。

**核心功能：**
- Tracing - 分布式追踪
- Metrics - 指标收集
- Logging - 结构化日志

**接口：** `ITelemetryManager`

```go
type ITelemetryManager interface {
    common.IBaseManager
    Tracer(name string) trace.Tracer
    TracerProvider() *sdktrace.TracerProvider
    Meter(name string) metric.Meter
    MeterProvider() *sdkmetric.MeterProvider
    Logger(name string) log.Logger
    LoggerProvider() *sdklog.LoggerProvider
    Shutdown(ctx context.Context) error
}
```

---

### LoggerManager

日志管理器提供结构化日志功能，支持多种格式和输出目标。

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

```go
type ILoggerManager interface {
    common.IBaseManager
    Ins() logger.ILogger
}
```

**日志格式示例：**

```
# Gin 格式
2026-01-24 15:04:05.123 | INFO  | 开始依赖注入 | count=23
2026-01-24 15:04:05.456 | WARN  | 慢查询检测 | duration=1.2s
2026-01-24 15:04:05.789 | ERROR | 数据库连接失败 | error="connection refused"

# JSON 格式
{"time":"2026-01-24T15:04:05.123Z","level":"INFO","msg":"开始依赖注入","count":23}
```

---

### DatabaseManager

数据库管理器基于 GORM，支持多种数据库驱动。

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
- 可观测性插件（慢查询、SQL 日志、指标）

**接口：** `IDatabaseManager`

```go
type IDatabaseManager interface {
    common.IBaseManager
    // GORM 核心
    DB() *gorm.DB
    Model(value any) *gorm.DB
    Table(name string) *gorm.DB
    WithContext(ctx context.Context) *gorm.DB

    // 事务管理
    Transaction(fn func(*gorm.DB) error, opts ...*sql.TxOptions) error
    Begin(opts ...*sql.TxOptions) *gorm.DB

    // 迁移管理
    AutoMigrate(models ...any) error
    Migrator() gorm.Migrator

    // 连接管理
    Driver() string
    Ping(ctx context.Context) error
    Stats() sql.DBStats
    Close() error

    // 原生 SQL
    Exec(sql string, values ...any) *gorm.DB
    Raw(sql string, values ...any) *gorm.DB
}
```

---

### CacheManager

缓存管理器提供统一的缓存操作接口，支持多种缓存驱动。

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

```go
type ICacheManager interface {
    common.IBaseManager
    Get(ctx context.Context, key string, dest any) error
    Set(ctx context.Context, key string, value any, expiration time.Duration) error
    SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error)
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    Expire(ctx context.Context, key string, expiration time.Duration) error
    TTL(ctx context.Context, key string) (time.Duration, error)
    Clear(ctx context.Context) error
    GetMultiple(ctx context.Context, keys []string) (map[string]any, error)
    SetMultiple(ctx context.Context, items map[string]any, expiration time.Duration) error
    DeleteMultiple(ctx context.Context, keys []string) error
    Increment(ctx context.Context, key string, value int64) (int64, error)
    Decrement(ctx context.Context, key string, value int64) (int64, error)
    Close() error
}
```

---

### LockManager

锁管理器提供分布式锁功能，支持阻塞和非阻塞锁获取。

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

```go
type ILockManager interface {
    common.IBaseManager
    Lock(ctx context.Context, key string, ttl time.Duration) error
    Unlock(ctx context.Context, key string) error
    TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
}
```

---

### LimiterManager

限流管理器提供限流功能，支持基于时间窗口的请求频率控制。

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

```go
type ILimiterManager interface {
    common.IBaseManager
    Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
    GetRemaining(ctx context.Context, key string, limit int, window time.Duration) (int, error)
}
```

---

### MQManager

消息队列管理器提供消息队列功能，支持异步消息处理。

**支持的驱动：**
- `rabbitmq` - RabbitMQ
- `memory` - 内存队列（用于开发和测试）

**核心功能：**
- 消息发布（Publish）
- 消息订阅（Subscribe、SubscribeWithCallback）
- 消息确认（Ack、Nack）
- 队列管理（QueueLength、Purge）

**接口：** `IMQManager`

```go
type IMQManager interface {
    common.IBaseManager
    Publish(ctx context.Context, queue string, message []byte, options ...PublishOption) error
    Subscribe(ctx context.Context, queue string, options ...SubscribeOption) (<-chan Message, error)
    SubscribeWithCallback(ctx context.Context, queue string, handler MessageHandler, options ...SubscribeOption) error
    Ack(ctx context.Context, message Message) error
    Nack(ctx context.Context, message Message, requeue bool) error
    QueueLength(ctx context.Context, queue string) (int64, error)
    Purge(ctx context.Context, queue string) error
    Close() error
}
```

---

### SchedulerManager

定时任务管理器提供定时任务管理功能，支持 Cron 表达式。

**支持的驱动：**
- `cron` - Cron 定时任务

**核心功能：**
- 定时任务验证（ValidateScheduler）
- 定时任务注册（RegisterScheduler）
- 定时任务注销（UnregisterScheduler）

**接口：** `ISchedulerManager`

```go
type ISchedulerManager interface {
    common.IBaseManager
    ValidateScheduler(scheduler common.IBaseScheduler) error
    RegisterScheduler(scheduler common.IBaseScheduler) error
    UnregisterScheduler(scheduler common.IBaseScheduler) error
}
```

## 依赖关系图

```
ConfigManager (无依赖)
    ↓
    ├── TelemetryManager
    │       └── ConfigManager
    │
    ├── LoggerManager
    │       ├── ConfigManager
    │       └── TelemetryManager
    │
    ├── DatabaseManager
    │       ├── ConfigManager
    │       ├── LoggerManager
    │       └── TelemetryManager
    │
    ├── CacheManager
    │       ├── ConfigManager
    │       ├── LoggerManager
    │       └── TelemetryManager
    │
    ├── LockManager
    │       ├── ConfigManager
    │       ├── LoggerManager
    │       ├── TelemetryManager
    │       └── CacheManager (可选)
    │
    ├── LimiterManager
    │       ├── ConfigManager
    │       ├── LoggerManager
    │       ├── TelemetryManager
    │       └── CacheManager (可选)
    │
    ├── MQManager
    │       ├── ConfigManager
    │       ├── LoggerManager
    │       └── TelemetryManager
    │
    └── SchedulerManager
            ├── ConfigManager
            └── LoggerManager
```

## 日志级别

| 级别 | 说明 | 使用场景 |
|------|------|----------|
| Debug | 开发调试信息 | 详细的操作日志 |
| Info | 正常业务流程 | 请求开始/完成、资源创建 |
| Warn | 降级处理、慢查询、重试 | 需要注意但不影响功能 |
| Error | 业务错误、操作失败 | 需要人工关注的错误 |
| Fatal | 致命错误 | 需要立即终止 |

## 敏感信息处理

日志管理器和数据库管理器会自动过滤和脱敏敏感信息：

- 密码、token、密钥等必须脱敏
- 支持内置过滤规则或自定义脱敏函数
- SQL 语句中的敏感参数会被脱敏

## 使用示例

### Repository 层使用 DatabaseManager

```go
type messageRepository struct {
    Manager databasemgr.IDatabaseManager `inject:""`
}

func (r *messageRepository) Create(message *entities.Message) error {
    db := r.Manager.DB()
    return db.Create(message).Error
}

func (r *messageRepository) FindByID(id uint) (*entities.Message, error) {
    var message entities.Message
    err := r.Manager.DB().First(&message, id).Error
    return &message, err
}
```

### Service 层使用多个 Manager

```go
type messageService struct {
    LoggerMgr  loggermgr.ILoggerManager   `inject:""`
    CacheMgr   cachemgr.ICacheManager     `inject:""`
    LockMgr    lockmgr.ILockManager      `inject:""`
    LimiterMgr limitermgr.ILimiterManager `inject:""`
}

func (s *messageService) CreateMessage(nickname, content string) (*entities.Message, error) {
    s.initLogger()
    s.logger.Info("创建留言", "nickname", nickname)

    // 限流检查
    allowed, err := s.LimiterMgr.Allow(context.Background(), "create_message", 10, time.Minute)
    if err != nil {
        return nil, err
    }
    if !allowed {
        return nil, errors.New("请求过于频繁")
    }

    // 获取分布式锁
    lockKey := "lock:create_message:" + nickname
    if err := s.LockMgr.Lock(context.Background(), lockKey, 5*time.Second); err != nil {
        return nil, err
    }
    defer s.LockMgr.Unlock(context.Background(), lockKey)

    // 业务逻辑...

    return message, nil
}
```

### Middleware 层使用 LoggerManager

```go
type RateLimiterMiddleware struct {
    Name      string
    Order     int
    Limiter   limitermgr.ILimiterManager `inject:""`
    LoggerMgr loggermgr.ILoggerManager   `inject:""`
    logger    logger.ILogger
}

func (m *RateLimiterMiddleware) initLogger() {
    if m.LoggerMgr != nil {
        m.logger = m.LoggerMgr.Ins()
    }
}

func (m *RateLimiterMiddleware) Handle(c *gin.Context) {
    m.initLogger()

    // 限流检查
    allowed, err := m.Limiter.Allow(c, c.ClientIP(), 100, time.Minute)
    if err != nil {
        m.logger.Error("限流检查失败", "error", err)
        c.JSON(500, gin.H{"error": "Internal Server Error"})
        c.Abort()
        return
    }

    if !allowed {
        m.logger.Warn("请求被限流", "ip", c.ClientIP())
        c.JSON(429, gin.H{"error": "Too Many Requests"})
        c.Abort()
        return
    }

    c.Next()
}
```

## 相关文档

- [SOP-package-document.md](../docs/SOP-package-document.md) - 功能包文档撰写规范
- [AGENTS.md](../AGENTS.md) - 项目开发指南
- [README.md](../README.md) - 项目主文档
- [container/README.md](../container/README.md) - 依赖注入容器说明
- [common/README.md](../common/README.md) - 公共接口说明
