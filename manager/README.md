# Manager 组件

Manager 组件是 litecore-go 的基础能力层，提供配置管理、日志、缓存、数据库、锁、限流器、消息队列、可观测性等核心功能。每个 Manager 组件都是独立的模块，通过依赖注入的方式集成到应用中。

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

## 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 包名 | `<功能名>mgr` | `cachemgr`、`configmgr` |
| 接口 | `I<功能名>Manager` | `ICacheManager`、`IDatabaseManager` |
| 实现类 | `<功能名>Manager<驱动名>Impl` | `cacheManagerRedisImpl` |
| 构造函数 | `New<功能名>Manager<驱动名>Impl` | `NewCacheManagerRedisImpl` |
| 配置结构 | `<功能名>Config`、`<驱动名>Config` | `CacheConfig`、`RedisConfig` |

## 接口设计

所有 Manager 接口都继承自 `common.IBaseManager`，该接口定义了以下方法：

```go
type IBaseManager interface {
    ManagerName() string    // 返回管理器名称
    Health() error          // 检查健康状态
    OnStart() error         // 服务器启动时触发
    OnStop() error          // 服务器停止时触发
}
```

## 组件概览

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

## 架构模式

### 工厂模式

每个 Manager 都提供两种工厂函数：

1. **Build** - 直接创建实例
```go
mgr, err := cachemgr.Build("memory", map[string]any{
    "max_age": "1h",
})
```

2. **BuildWithConfigProvider** - 从配置提供者创建实例
```go
mgr, err := cachemgr.BuildWithConfigProvider(configProvider)
```

### 基类实现

大部分 Manager 都有基类实现（`impl_base.go`），提供：
- 可观测性支持（日志、指标、链路追踪）
- 工具函数（上下文验证、键验证等）

### 配置驱动

支持通过配置文件切换驱动：

```yaml
cache:
  driver: "memory"        # redis, memory, none
  memory_config:
    max_age: "1h"

database:
  driver: "sqlite"         # mysql, postgresql, sqlite, none
  sqlite_config:
    dsn: "./data/app.db"
```

## 依赖注入

Manager 组件通过 `inject:""` 标签注入到各层：

```go
type MyService struct {
    CacheMgr   cachemgr.ICacheManager      `inject:""`
    DBMgr      databasemgr.IDatabaseManager `inject:""`
    LoggerMgr  loggermgr.ILoggerManager    `inject:""`
}
```

**注意：** `LoggerMgr.Ins()` 返回 `logger.ILogger` 实例，用于日志记录。

### 日志记录模式

```go
func (s *MyService) initLogger() {
    if s.LoggerMgr != nil {
        s.logger = s.LoggerMgr.Ins()
    }
}

func (s *MyService) SomeMethod() {
    s.initLogger()
    s.logger.Info("操作开始", "param", value)
}
```

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
# 缓存配置
cache:
  driver: "memory"
  memory_config:
    max_size: 100
    max_age: "720h"  # 30 days

# 数据库配置
database:
  driver: "mysql"
  mysql_config:
    dsn: "root:password@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local"

# 日志配置
logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "info"
      format: "gin"      # gin | json | default
      color: true

# 限流配置
limiter:
  driver: "memory"
  memory_config:
    max_backups: 1000

# 锁配置
lock:
  driver: "redis"
  redis_config:
    host: "localhost"
    port: 6379

# 消息队列配置
mq:
  driver: "rabbitmq"
  rabbitmq_config:
    url: "amqp://guest:guest@localhost:5672/"

# 遥测配置
telemetry:
  driver: "otel"
  otel_config:
    endpoint: "http://localhost:4318"
```

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

## 日志级别

| 级别 | 说明 |
|------|------|
| Debug | 开发调试信息 |
| Info | 正常业务流程（请求开始/完成、资源创建） |
| Warn | 降级处理、慢查询、重试 |
| Error | 业务错误、操作失败（需人工关注） |
| Fatal | 致命错误，需要立即终止 |

## 敏感信息处理

日志管理器会自动过滤和脱敏敏感信息：
- 密码、token、密钥等必须脱敏
- 支持内置过滤规则或自定义脱敏函数

## messageboard 使用示例

### Repository 层

```go
type messageRepository struct {
    Manager databasemgr.IDatabaseManager `inject:""`
}

func (r *messageRepository) Create(message *entities.Message) error {
    db := r.Manager.DB()
    return db.Create(message).Error
}
```

### Service 层

```go
type messageService struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func (s *messageService) CreateMessage(nickname, content string) (*entities.Message, error) {
    s.LoggerMgr.Ins().Info("创建留言成功", "id", message.ID, "nickname", message.Nickname)
    return message, nil
}
```

### Middleware 层

```go
func NewRateLimiterMiddleware() IRateLimiterMiddleware {
    return litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
        Limit:  &limit,
        Window: &window,
    })
}
```
