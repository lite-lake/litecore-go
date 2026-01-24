# CLAUDE.md

此文件为 Claude Code (claude.ai/code) 在此仓库中工作时提供指导。

## 项目概述

**LiteCore-Go** 是一个 Go 语言的分层应用框架，提供依赖注入容器、生命周期管理和基础设施管理器。基于 Gin、GORM 和 Zap 构建。

**模块**: `github.com/lite-lake/litecore-go`
**Go 版本**: 1.25+
 **架构**: 5 层分层依赖注入（内置管理器层 → Entity → Repository → Service → 交互层）
  - **交互层**: Controller/Middleware/Listener/Scheduler 统一处理 HTTP 请求、MQ 消息和定时任务
**内置组件**:
- **管理器组件**: `manager/*/` (configmgr, loggermgr, databasemgr, cachemgr, telemetrymgr, limitermgr, lockmgr, mqmgr)
- **组件层**: `component/litecontroller`, `component/litemiddleware`, `component/liteservice`

## 基本命令

### 构建、测试和检查

```bash
# 构建所有包
go build -o litecore ./...

# 运行所有测试
go test ./...

# 运行测试并生成覆盖率
go test -cover ./...

# 测试特定包
go test ./util/jwt

# 运行单个测试
go test ./util/jwt -run TestGenerateHS256Token
go test -v ./util/jwt -run TestGenerateHS256Token/valid_StandardClaims

# 运行基准测试
go test -bench=. ./util/hash
go test -bench=BenchmarkMD5 ./util/hash

# 格式化和检查
go fmt ./...
go vet ./...
go mod tidy
```

## 架构概述

  ### 5 层分层架构

  框架强制执行严格的层边界和单向依赖关系：

  ```
  ┌──────────────────────────────────────────────────────────────────┐
  │                交互层 (Interaction Layer)                     │
  │  Controller   - HTTP 请求处理 (component/litecontroller) │
  │  Middleware   - 请求拦截 (component/litemiddleware)  │
  │  Listener     - MQ 消息处理                          │
  │  Scheduler    - 定时任务                               │
  ├──────────────────────────────────────────────────────────────────┤
  │  Service Layer (component/liteservice)                         │
  ├──────────────────────────────────────────────────────────────────┤
  │  Repository Layer (BaseRepository)                             │
  ├──────────────────────────────────────────────────────────────────┤
  │  Entity Layer (BaseEntity)                                    │
  │  Manager Layer (manager/*)                                    │
  │  Managers: configmgr, loggermgr, databasemgr, cachemgr,        │
  │           telemetrymgr, limitermgr, lockmgr, mqmgr, schedulermgr│
  └──────────────────────────────────────────────────────────────────┘
  ```

  **依赖规则**:
  - **Entity** - 无依赖
  - **Manager** → Config, other Managers (同层依赖)
  - **Repository** → Config, Manager, Entity
  - **Service** → Config, Manager, Repository, other Services (同层)
  - **交互层 (Controller/Middleware/Listener/Scheduler)** → Config, Manager, Service

  **管理器组件** (`manager/*/`):
  - `configmgr` - 配置管理 (YAML/JSON)
  - `loggermgr` - 结构化日志 (Zap 支持 Gin/JSON/default 格式)
  - `databasemgr` - 数据库管理 (GORM: MySQL/PostgreSQL/SQLite)
  - `cachemgr` - 缓存管理 (Ristretto/Redis/None)
  - `telemetrymgr` - OpenTelemetry (链路追踪/指标/日志)
  - `limitermgr` - 限流 (内存/Redis)
  - `lockmgr` - 分布式锁 (内存/Redis)
  - `mqmgr` - 消息队列 (RabbitMQ/内存)
  - `schedulermgr` - 定时任务管理 (基于 Cron)

### 依赖注入容器

容器系统 (`container/`) 管理组件生命周期并强制执行层边界。

**关键模式 - 按接口类型注册**:
```go
// 使用 RegisterByType 按接口类型注册实例
serviceContainer.RegisterByType(
    reflect.TypeOf((*UserService)(nil)).Elem(),
    &UserServiceImpl{},
)
```

**两阶段注入**:
 1. **注册阶段** (`RegisterByType`) - 将实例添加到容器，不进行注入
 2. **注入阶段** (`InjectAll`) - 解析并按拓扑排序注入依赖

**依赖声明**:
```go
type UserServiceImpl struct {
    Config    configmgr.IConfigManager    `inject:""`
    DBMgr     databasemgr.IDatabaseManager `inject:""`
    UserRepo  IUserRepository              `inject:""`
    OrderSvc  IOrderService               `inject:""`  // 同层依赖
}
```

### 管理器模式

管理器 (`manager/*/`) 提供基础设施能力（数据库、缓存、日志、遥测、限流、锁、消息）。

 **可用管理器**:
  - `configmgr` - 配置加载器 (YAML/JSON)，支持路径查询
  - `loggermgr` - 结构化日志，支持 Gin/JSON/default 格式和颜色
  - `databasemgr` - GORM 数据库 (MySQL/PostgreSQL/SQLite)
  - `cachemgr` - 高性能缓存 (内存用 Ristretto，分布式用 Redis，或 None)
  - `telemetrymgr` - OpenTelemetry 集成 (链路追踪、指标、日志)
  - `limitermgr` - 限流 (滑动窗口，内存/Redis)
  - `lockmgr` - 分布式锁 (阻塞/非阻塞，内存/Redis)
  - `mqmgr` - 消息队列 (RabbitMQ，内存队列)
  - `schedulermgr` - 支持 Cron 的定时任务管理器

**标准管理器结构**:
- `interface.go` - 核心接口 (扩展 `common.IBaseManager`)
- `config.go` - 配置结构和解析
- `impl_base.go` - 带可观测性的基础实现
- `{driver}_impl.go` - 驱动特定实现
- `factory.go` - DI 工厂函数

**配置路径约定**:
```
{manager}.driver           # 驱动类型 (如 mysql, redis, rabbitmq)
{manager}.{driver}_config  # 驱动配置
```

示例:
- `database.driver` + `database.mysql_config`
- `cache.driver` + `cache.redis_config` / `cache.memory_config`
- `logger.driver` + `logger.zap_config`
- `limiter.driver` + `limiter.redis_config` / `limiter.memory_config`
- `lock.driver` + `lock.redis_config` / `lock.memory_config`
- `mq.driver` + `mq.rabbitmq_config`

### 服务器引擎

`server` 包提供 HTTP 服务器生命周期管理：

 **生命周期流程**:
 1. `Initialize()` - 自动初始化管理器 (config → telemetry → logger → database → cache → lock → limiter → mq → scheduler)，注册到容器
 2. `NewEngine()` - 创建带容器的引擎
 3. `Start()` - 启动管理器、仓库、服务、交互层组件和 HTTP 服务器，输出启动日志
 4. `Stop()` - 优雅关闭

**启动日志**:
框架记录每个启动阶段：
- 配置文件和驱动信息
- 管理器初始化状态
- 组件计数 (controllers, middlewares, services)
- 依赖注入结果

## 代码风格和约定

### 导入顺序
```go
import (
    "crypto"       // 标准库优先
    "errors"
    "time"

    "github.com/gin-gonic/gin"  // 第三方库其次
    "github.com/stretchr/testify/assert"

    "github.com/lite-lake/litecore-go/common"  // 本地模块最后
)
```

### 命名约定
- **接口**: `I*` 前缀 (如 `IConfigManager`, `IDatabaseManager`, `IUserService`)
- **私有结构体**: 小写 (如 `messageService`, `messageRepository`)
- **公共结构体**: 大驼峰 (如 `ServerConfig`, `User`)
- **函数**: 导出用大驼峰，私有用小驼峰
- **工厂函数**: `Build()`, `BuildWithConfigProvider()`, `NewXxx()`

### 注释和文档
- 所有面向用户的文档和代码注释使用**中文**
- 包文档放在 `doc.go` 中
- 函数注释必须说明目的和参数
- 导出函数必须有注释

### 日志最佳实践
- 在业务层**注入 ILoggerManager**: `LoggerMgr loggermgr.ILoggerManager \`inject:""\``
- **初始化日志**: DI 后使用 `s.logger = s.LoggerMgr.Ins()`
- **结构化日志**: `s.logger.Info("msg", "key", value)`
- **上下文感知**: `s.logger.With("user_id", id).Info("...")`
- **日志级别**:
  - Debug: 开发调试信息
  - Info: 正常业务流程 (请求开始/完成、资源创建)
  - Warn: 降级处理、慢查询、重试
  - Error: 业务错误、操作失败 (需人工关注)
  - Fatal: 致命错误，需要立即终止

**日志格式** (通过 `logger.zap_config.console_config.format` 配置):
- `gin` - Gin 风格，竖线分隔符，控制台友好 (默认)
- `json` - JSON 格式，用于日志分析和监控
- `default` - 默认 ConsoleEncoder 格式

**Gin 格式示例**:
```
2026-01-24 15:04:05.123 | INFO  | 开始依赖注入 | count=23
2026-01-24 15:04:05.456 | WARN  | 慢查询检测 | duration=1.2s
2026-01-24 15:04:05.789 | ERROR | 数据库连接失败 | error="connection refused"
```

### 错误处理
```go
if err != nil {
    return "", fmt.Errorf("operation failed: %w", err)
}
```

### 依赖注入标签
```go
type MyService struct {
    Config     configmgr.IConfigManager    `inject:""`
    DBMgr      databasemgr.IDatabaseManager `inject:""`
    CacheMgr   cachemgr.ICacheManager      `inject:""`
    LimiterMgr limitermgr.ILimiterManager `inject:""`
}
```

### 测试模式
- 使用 `t.Run()` 子测试的表驱动测试
- 使用 `testify/assert` 进行断言
- 基准测试函数以 `Benchmark` 为前缀

```go
func TestGenerateToken(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid input", "data", false},
        {"empty input", "", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

## 管理器实现标准流程

创建或修改管理器时，应放置在 `manager/*/`：

 1. **扁平结构** - 管理器包内无子目录
 2. **文件组织**:
    - `interface.go` - 核心接口 (扩展 `common.IBaseManager`)
    - `config.go` - 配置结构和解析
    - `impl_base.go` - 带可观测性的基础实现
    - `{driver}_impl.go` - 驱动特定实现
    - `factory.go` - DI 工厂函数
 3. **DI 标签** - 使用 `inject:""` 声明依赖
 4. **配置路径** - 遵循 `{manager}.driver` 约定
 5. **自动初始化** - 按正确顺序添加到 `server/builtin.go:Initialize()`

 **管理器初始化顺序** (依赖关系很重要):
  1. configmgr (必须第一个)
  2. telemetrymgr
  3. loggermgr
  4. databasemgr
  5. cachemgr
  6. lockmgr
  7. limitermgr
  8. mqmgr
  9. schedulermgr (依赖 loggermgr, configmgr)

## 常见开发模式

 ### 添加新功能
 1. 在 `entities/` 创建 Entity
 2. 在 `repositories/` 创建 Repository 接口和实现
 3. 在 `services/` 创建 Service 接口和实现
 4. 创建交互层组件:
     - Controller 在 `controllers/` (HTTP 请求处理)
     - Middleware 在 `middlewares/` (请求拦截)
     - Listener 在 `listeners/` (MQ 消息处理)
     - Scheduler 在 `schedulers/` (定时任务)
 5. 使用 `RegisterByType()` 在容器中注册所有组件
 6. 在 `InjectAll()` 时自动注入依赖

### 创建管理器
1. 定义扩展 `common.IBaseManager` 的接口
2. 使用 `impl_base.go` 实现可观测性
3. 创建驱动实现 (memory, redis 等)
4. 提供 `Build()` 和 `BuildWithConfigProvider()` 工厂函数
5. 遵循配置路径约定 (`{manager}.driver`, `{manager}.{driver}_config`)
6. 按正确顺序添加到 `server/builtin.go:Initialize()`

### 使用内置组件
Controllers、middlewares 和 services 在 `component/` 中可用:
- `component/litecontroller` - Health, Metrics, Pprof, Resource controllers
- `component/litemiddleware` - CORS, Recovery, RequestLogger, SecurityHeaders, RateLimiter, Telemetry
- `component/liteservice` - HTMLTemplateService

```go
// 使用默认配置注册内置中间件
cors := litemiddleware.NewCorsMiddleware(nil)
recovery := litemiddleware.NewRecoveryMiddleware(nil)
limiter := litemiddleware.NewRateLimiterMiddleware(nil)
middlewareContainer.RegisterMiddleware(cors)
middlewareContainer.RegisterMiddleware(recovery)
middlewareContainer.RegisterMiddleware(limiter)

// 自定义中间件配置
allowOrigins := []string{"https://example.com"}
customCors := litemiddleware.NewCorsMiddleware(&litemiddleware.CorsConfig{
    AllowOrigins: &allowOrigins,
})
```

### 示例应用
参见 `samples/messageboard/` 了解完整工作示例，演示:
- 完整的分层架构
- 容器注册
- 管理器使用 (database, cache, limiter, lock, mq)
- GORM 与 Ristretto 缓存集成
- 内置中间件 (CORS, RateLimiter, Telemetry)
- 自定义路由和中间件

## 配置

所有配置使用 YAML 格式。管理器组件遵循此模式：

```yaml
# 管理器配置遵循模式:
database:
  driver: mysql
  mysql_config:
    host: "localhost"
    port: 3306
    database: "mydb"
    username: "root"
    password: "password"

cache:
  driver: memory  # memory | redis | none
  memory_config:
    max_size: 100        # MB
    max_age: 720h        # 30 days
    compress: false
  # redis_config:
  #   host: localhost
  #   port: 6379

logger:
  driver: zap
  zap_config:
    console_enabled: true
    console_config:
      level: "info"
      format: "gin"      # gin | json | default
      color: true
      time_format: "2006-01-02 15:04:05.000"
    file_enabled: false
    file_config:
      level: "info"
      path: "./logs/app.log"
      rotation:
        max_size: 100
        max_age: 30
        max_backups: 10
        compress: true

limiter:
  driver: memory  # memory | redis
  memory_config:
    max_backups: 1000

lock:
  driver: redis  # memory | redis
  redis_config:
    host: localhost
    port: 6379
    db: 0

mq:
  driver: rabbitmq  # rabbitmq | memory
  rabbitmq_config:
    url: "amqp://guest:guest@localhost:5672/"
    durable: true

 telemetry:
   driver: otel
   otel_config:
     endpoint: "http://localhost:4318"
     enabled_traces: true
     enabled_metrics: true
     enabled_logs: true

 scheduler:
   driver: cron
   cron_config:
     validate_on_startup: true  # 启动时验证 crontab 规则
  ```

 **内置管理器** (`manager/*/`):
  - `configmgr` - 配置管理
  - `loggermgr` - 日志管理，支持 Gin/JSON/default 格式
  - `databasemgr` - 数据库 (MySQL/PostgreSQL/SQLite)
  - `cachemgr` - 缓存 (Ristretto/Redis/None)
  - `telemetrymgr` - OpenTelemetry
  - `limitermgr` - 限流
  - `lockmgr` - 分布式锁
  - `mqmgr` - 消息队列
  - `schedulermgr` - 定时任务管理，支持 Cron

**内置组件** (`component/*/`):
- `litecontroller` - Health, Metrics, Pprof controllers
- `litemiddleware` - CORS, Recovery, RequestLogger, SecurityHeaders, RateLimiter, Telemetry
- `liteservice` - HTMLTemplateService

 ## 重要架构约束

  1. **无循环依赖** - 容器检测并报告循环
  2. **强制层边界** - 上层不能依赖下层
  3. **基于接口的 DI** - 按接口类型注册，而非具体类型
  4. **两阶段注入** - 先注册，后注入
  5. **管理器生命周期** - 所有管理器实现 `OnStart()/OnStop()`
  6. **管理器初始化顺序** - config → telemetry → logger → database → cache → lock → limiter → mq → scheduler
  7. **交互层组件** - Controller/Middleware/Listener/Scheduler 都在同一层，仅依赖 Service 和 Manager
  8. **中间件执行顺序** - Recovery (0) → RequestLogger (50) → CORS (100) → SecurityHeaders (150) → RateLimiter (200) → Telemetry (250)
  9. **组件路径** - 管理器在 `manager/*/`，组件在 `component/litecontroller`, `component/litemiddleware`, `component/liteservice`

## 测试策略

- 单元测试在源文件旁边的 `*_test.go` 文件中
- 使用表驱动测试处理多个场景
- 使用 `testify/mock` 模拟接口
- 集成测试在 samples 中
- 对关键路径进行基准测试

## 相关文档

- **AGENTS.md** - AI 代理开发指南 (编码规范、日志、架构)
- **manager/README.md** - 管理器组件文档 (详细 API 和用法)
- **component/README.md** - 内置组件文档
- **component/litemiddleware/README.md** - 中间件配置指南
