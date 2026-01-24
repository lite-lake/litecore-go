# Server

提供统一的 HTTP 服务引擎，支持自动依赖注入、生命周期管理和中间件集成。

## 特性

- **内置组件自动初始化** - 自动初始化 9 个内置 Manager（Config、Telemetry、Logger、Database、Cache、Lock、Limiter、MQ、Scheduler）
- **5 层依赖注入架构** - Entity → Repository → Service → Controller → Middleware，支持层级依赖和同层依赖
- **生命周期管理** - 统一管理各层组件的启动和停止，启动和停止顺序可预测
- **中间件集成** - 自动排序并注册全局中间件，支持通过配置自定义名称和执行顺序
- **路由管理** - 自动注册控制器路由，支持 OpenAPI 风格的路由定义 `/path [METHOD]`
- **优雅关闭** - 支持信号处理，超时控制的安全关闭机制
- **启动日志** - 支持异步启动日志，记录各阶段启动状态和耗时
- **定时任务支持** - 集成 SchedulerManager，支持 Cron 表达式的定时任务
- **消息队列监听器** - 集成 Listener 层，支持消息队列消费者
- **自动数据库迁移** - 支持配置驱动的自动数据库表结构迁移

## 快速开始

### 方式一：使用 CLI 生成的应用引擎（推荐）

```go
package main

import (
    "log"

    "github.com/lite-lake/litecore-go/samples/messageboard/internal/application"
)

func main() {
    // 创建应用引擎（由 CLI 工具自动生成）
    engine, err := application.NewEngine()
    if err != nil {
        log.Fatalf("Failed to create engine: %v", err)
    }

    // 一键启动
    if err := engine.Run(); err != nil {
        log.Fatalf("Engine run failed: %v", err)
    }
}
```

### 方式二：手动创建引擎

```go
package main

import (
    "github.com/lite-lake/litecore-go/server"
    "github.com/lite-lake/litecore-go/container"
)

func main() {
    // 创建容器
    entityContainer := container.NewEntityContainer()
    repositoryContainer := container.NewRepositoryContainer(entityContainer)
    serviceContainer := container.NewServiceContainer(repositoryContainer)
    controllerContainer := container.NewControllerContainer(serviceContainer)
    middlewareContainer := container.NewMiddlewareContainer(serviceContainer)
    listenerContainer := container.NewListenerContainer(serviceContainer)
    schedulerContainer := container.NewSchedulerContainer(serviceContainer)

    // 注册其他组件（实体、仓储、服务、控制器、中间件）...

    // Manager 由引擎自动初始化和注入（位于 manager/ 目录）

    // 创建并启动引擎
    engine := server.NewEngine(
        &server.BuiltinConfig{
            Driver:   "yaml",
            FilePath: "config.yaml",
        },
        entityContainer,
        repositoryContainer,
        serviceContainer,
        controllerContainer,
        middlewareContainer,
        listenerContainer,
        schedulerContainer,
    )
    if err := engine.Run(); err != nil {
        panic(err)
    }
}
```

## 核心功能

### 创建引擎

使用 CLI 生成的 `NewEngine()` 创建服务引擎（推荐）：

```go
// 由 CLI 工具自动生成
engine, err := application.NewEngine()
if err != nil {
    log.Fatalf("Failed to create engine: %v", err)
}
```

或手动创建（传入配置和所有容器实例）：

```go
engine := server.NewEngine(
    &server.BuiltinConfig{
        Driver:   "yaml",
        FilePath: "config.yaml",
    },
    entityContainer,
    repositoryContainer,
    serviceContainer,
    controllerContainer,
    middlewareContainer,
    listenerContainer,
    schedulerContainer,
)
```

### BuiltinConfig 配置

`BuiltinConfig` 用于配置内置组件的初始化参数：

```go
type BuiltinConfig struct {
    Driver   string // 配置驱动类型（支持：yaml、json 等）
    FilePath string // 配置文件路径
}
```

**配置示例：**

```go
// YAML 配置文件（推荐）
builtinConfig := &server.BuiltinConfig{
    Driver:   "yaml",
    FilePath: "configs/config.yaml",
}

// JSON 配置文件
builtinConfig := &server.BuiltinConfig{
    Driver:   "json",
    FilePath: "configs/config.json",
}
```

配置文件结构示例（`configs/config.yaml`）：

```yaml
# 应用配置
app:
  name: "my-app"
  version: "1.0.0"

# 服务器配置
server:
  host: "0.0.0.0"              # 监听地址
  port: 8080                   # 监听端口
  mode: "release"              # 运行模式：debug/release/test
  read_timeout: "10s"          # 读取超时
  write_timeout: "10s"         # 写入超时
  idle_timeout: "60s"          # 空闲超时
  shutdown_timeout: "30s"      # 关闭超时
  startup_log:                 # 启动日志配置
    enabled: true              # 是否启用启动日志
    async: true                # 是否异步输出
    buffer: 100                # 缓冲区大小

# 数据库配置（支持自动迁移）
database:
  driver: "sqlite"
  auto_migrate: true           # 是否自动迁移数据库表结构
  sqlite_config:
    dsn: "./data/app.db"

# 日志配置
logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "info"
      format: "gin"            # 格式：gin | json | default

# 其他 Manager 配置（cache、lock、limiter、mq、scheduler 等）
...
```

### 启动服务

提供两种启动方式：

**方式一：一步启动（推荐）**

```go
// Run() = Initialize() + Start() + WaitForShutdown()
if err := engine.Run(); err != nil {
    log.Fatalf("Engine run failed: %v", err)
}
```

**方式二：分步启动（需要自定义初始化时）**

```go
// 1. 初始化（初始化内置组件、依赖注入、创建 Gin 引擎、注册中间件和路由）
if err := engine.Initialize(); err != nil {
    log.Fatalf("Failed to initialize engine: %v", err)
}

// 2. 启动服务（启动各层组件和 HTTP 服务器）
if err := engine.Start(); err != nil {
    log.Fatalf("Failed to start engine: %v", err)
}

// 3. 等待关闭信号
engine.WaitForShutdown()
```

### 启动日志

Engine 支持启动日志功能，记录各阶段的启动状态和耗时：

| 启动阶段 | 说明 |
|---------|------|
| 配置加载 | 加载配置文件 |
| 管理器初始化 | 初始化所有 Manager |
| 配置验证 | 验证 Scheduler crontab 配置 |
| 依赖注入 | 执行各层组件的依赖注入 |
| 路由注册 | 注册中间件和控制器路由 |
| 组件启动 | 启动各层组件 |
| 运行中 | HTTP 服务器运行 |
| 关闭中 | 优雅关闭各层组件 |

日志格式（Gin 风格）：
```
2026-01-24 15:04:05.123 | INFO  | 开始初始化内置组件
2026-01-24 15:04:05.456 | INFO  | 初始化完成: ConfigManager
2026-01-24 15:04:05.789 | INFO  | 管理器初始化完成 | count=9 | duration=1.2s
2026-01-24 15:04:06.123 | INFO  | 开始依赖注入
2026-01-24 15:04:06.456 | INFO  | [Repository 层] MessageRepository: 注入完成
2026-01-24 15:04:06.789 | INFO  | 依赖注入完成 | count=23 | duration=0.5s
2026-01-24 15:04:07.123 | INFO  | HTTP 服务器启动成功 | address=0.0.0.0:8080 | total_duration=2.0s
```

**启动日志配置：**

```yaml
server:
  startup_log:
    enabled: true   # 是否启用启动日志
    async: true     # 是否异步输出（默认 true）
    buffer: 100     # 缓冲区大小（默认 100）
```

### 生命周期管理

Engine 按以下顺序管理组件生命周期：

**Initialize() 初始化顺序：**
1. 管理器初始化（按顺序初始化 9 个内置 Manager）
   - ConfigManager（必须最先初始化）
   - TelemetryManager（依赖 ConfigManager）
   - LoggerManager（依赖 ConfigManager、TelemetryManager）
   - DatabaseManager（依赖 ConfigManager）
   - CacheManager（依赖 ConfigManager）
   - LockManager（依赖 ConfigManager）
   - LimiterManager（依赖 ConfigManager）
   - MQManager（依赖 ConfigManager）
   - SchedulerManager（依赖 ConfigManager、LoggerManager）
2. 配置验证（验证 Scheduler crontab 规则）
3. 依赖注入（按层顺序）
   - Repository 层（依赖 Manager、Entity）
   - Service 层（依赖 Manager、Repository 和同层 Service）
   - Controller 层（依赖 Manager、Service）
   - Middleware 层（依赖 Manager、Service）
   - Listener 层（依赖 Manager）
   - Scheduler 层（依赖 Manager）
4. 创建 Gin 引擎
5. 注册中间件和路由

**Start() 启动顺序：**
1. Manager 层（按注册顺序）
2. 自动迁移数据库（如果启用）
3. Repository 层（按注册顺序）
4. Service 层（按注册顺序）
5. Middleware 层（按注册顺序）
6. Scheduler 层（注册到 SchedulerManager 并启动）
7. Listener 层（注册到 MQManager 并启动）
8. HTTP 服务器

**Stop() 停止顺序（反转启动顺序）：**
1. HTTP 服务器（优雅关闭）
2. Listener 层（反转注册顺序）
3. Scheduler 层（反转注册顺序）
4. Middleware 层（反转注册顺序）
5. Service 层（反转注册顺序）
6. Repository 层（反转注册顺序）
7. Manager 层（反转注册顺序）

```go
// 手动停止
if err := engine.Stop(); err != nil {
    log.Printf("Stop error: %v", err)
}
```

### 自动数据库迁移

Engine 支持配置驱动的自动数据库迁移功能：

**配置方式：**

```yaml
database:
  driver: "sqlite"
  auto_migrate: true           # 启用自动迁移
  sqlite_config:
    dsn: "./data/app.db"
```

**迁移逻辑：**
1. 在 Start() 阶段，如果 `auto_migrate` 为 `true`，Engine 会自动执行数据库迁移
2. 迁移所有已注册的 Entity（通过 EntityContainer.GetAll() 获取）
3. 使用 DatabaseManager 的 AutoMigrate() 方法执行迁移

**注意事项：**
- 仅在开发环境使用，生产环境建议使用专门的数据库迁移工具
- 确保 Entity 定义与预期数据库表结构一致
- 迁移失败会中断启动流程

### 依赖注入

Engine 在初始化时自动按以下顺序执行依赖注入：

1. **Manager 层**（由 `server.Initialize()` 自动初始化并注入）
2. **Entity 层**（无依赖）
3. **Repository 层**（依赖 Manager、Entity）
4. **Service 层**（依赖 Manager、Repository 和同层 Service）
5. **Controller 层**（依赖 Manager、Service）
6. **Middleware 层**（依赖 Manager、Service）
7. **Listener 层**（依赖 Manager）
8. **Scheduler 层**（依赖 Manager）

各层组件通过 `inject:""` 标签声明依赖，Manager 由引擎自动注入：

```go
type UserServiceImpl struct {
    // 内置组件（引擎自动注入，来自 manager 包）
    Config     configmgr.IConfigManager      `inject:""`
    DBManager  databasemgr.IDatabaseManager  `inject:""`
    LoggerMgr  loggermgr.ILoggerManager     `inject:""`
    LockMgr    lockmgr.ILockManager         `inject:""`
    LimiterMgr limitermgr.ILimiterManager    `inject:""`
    CacheMgr   cachemgr.ICacheManager        `inject:""`

    // 业务依赖
    UserRepo   repository.IUserRepository   `inject:""`
}
```

### 中间件管理

中间件按 `Order()` 排序后自动注册到 Gin 引擎。内置中间件支持通过配置自定义名称和执行顺序：

```go
type AuthMiddleware struct {
    Name  string
    Order int // 越小越先执行
}

func (m *AuthMiddleware) MiddlewareName() string {
    return m.Name
}

func (m *AuthMiddleware) Order() int {
    return m.Order
}

func (m *AuthMiddleware) Wrapper() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        // 中间件逻辑
        c.Next()
    })
}
```

#### 预定义的中间件 Order

| 中间件 | Order | 说明 |
|--------|-------|------|
| Recovery | 0 | panic 恢复（最先执行） |
| RequestLogger | 50 | 请求日志 |
| CORS | 100 | 跨域处理 |
| SecurityHeaders | 150 | 安全头 |
| RateLimiter | 200 | 限流（认证前执行） |
| Telemetry | 250 | 遥测 |
| Auth | 300 | 认证（预留） |

业务自定义中间件建议从 Order 350 开始。

### 路由管理

**通过控制器定义路由**

所有路由都必须通过控制器（Controller）的 `GetRouter()` 方法定义，格式为：`/path [METHOD]`

```go
type UserController struct {
    Config    configmgr.IConfigManager  `inject:""`
    UserService service.IUserService `inject:""`
}

func (ctrl *UserController) GetRouter() string {
    return "/users [GET]"
}

func (ctrl *UserController) Handle(c *gin.Context) {
    users, err := ctrl.UserService.List()
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, users)
}
```

**支持的路由格式：**

```go
// 正确格式（OpenAPI 风格）
"/users [GET]"
"/users/:id [GET]"
"/users [POST]"
"/users/:id [PUT]"
"/users/:id [DELETE]"
"/static/*filepath [GET]"

// 错误格式（会被忽略）
""                // 空字符串
"users"           // 缺少方法
"/users[GET]"     // 缺少空格
```

### 定时任务（Scheduler）

Engine 支持集成 SchedulerManager，管理定时任务的注册和启动：

**Scheduler 实现示例：**

```go
type StatisticsScheduler struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
    Service   service.IStatisticsService `inject:""`
}

func (s *StatisticsScheduler) SchedulerName() string {
    return "StatisticsScheduler"
}

func (s *StatisticsScheduler) GetRule() string {
    return "0 0 * * *"  // 每天 00:00 执行
}

func (s *StatisticsScheduler) GetTimezone() string {
    return "Asia/Shanghai"
}

func (s *StatisticsScheduler) OnStart() error {
    s.initLogger()
    s.logger.Info("Statistics scheduler started")
    return nil
}

func (s *StatisticsScheduler) OnStop() error {
    s.logger.Info("Statistics scheduler stopped")
    return nil
}

func (s *StatisticsScheduler) Execute(ctx context.Context) error {
    return s.Service.GenerateDailyReport(ctx)
}
```

**配置方式：**

```yaml
scheduler:
  driver: "cron"
  cron_config:
    validate_on_startup: true  # 启动时验证所有 crontab 规则
```

### 消息队列监听器（Listener）

Engine 支持集成 Listener 层，订阅并处理消息队列消息：

**Listener 实现示例：**

```go
type MessageCreatedListener struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
    Service   service.INotificationService `inject:""`
}

func (l *MessageCreatedListener) ListenerName() string {
    return "MessageCreatedListener"
}

func (l *MessageCreatedListener) GetQueue() string {
    return "message.created"
}

func (l *MessageCreatedListener) GetSubscribeOptions() []any {
    return []any{
        mqmgr.WithAck(),
    }
}

func (l *MessageCreatedListener) OnStart() error {
    l.initLogger()
    l.logger.Info("MessageCreatedListener started")
    return nil
}

func (l *MessageCreatedListener) OnStop() error {
    l.logger.Info("MessageCreatedListener stopped")
    return nil
}

func (l *MessageCreatedListener) Handle(ctx context.Context, msg mqmgr.Message) error {
    return l.Service.SendNotification(ctx, msg)
}
```

## API

### Engine

服务引擎主结构，提供完整的服务管理功能。

#### 构造函数

```go
func NewEngine(
    builtinConfig *BuiltinConfig,
    entity *container.EntityContainer,
    repository *container.RepositoryContainer,
    service *container.ServiceContainer,
    controller *container.ControllerContainer,
    middleware *container.MiddlewareContainer,
    listener *container.ListenerContainer,
    scheduler *container.SchedulerContainer,
) *Engine
```

#### 生命周期方法

| 方法 | 说明 |
|------|------|
| `Initialize() error` | 初始化引擎（初始化内置组件、依赖注入、创建 Gin 引擎、注册中间件和路由） |
| `Start() error` | 启动引擎（启动各层组件和 HTTP 服务器） |
| `Stop() error` | 停止引擎（优雅关闭） |
| `Run() error` | 一键启动（Initialize + Start + WaitForShutdown） |
| `WaitForShutdown()` | 等待关闭信号 |

### BuiltinConfig

内置组件配置结构。

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| Driver | string | 配置驱动类型 | `"yaml"`, `"json"` |
| FilePath | string | 配置文件路径 | `"configs/config.yaml"` |

| 方法 | 说明 |
|------|------|
| `Validate() error` | 验证配置参数是否有效 |

### StartupLogConfig

启动日志配置结构。

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| Enabled | bool | 是否启用启动日志 | `true` |
| Async | bool | 是否异步输出 | `true` |
| Buffer | int | 缓冲区大小 | `100` |

### StartupPhase

启动阶段枚举。

| 常量 | 值 | 说明 |
|------|-----|------|
| `PhaseConfig` | 0 | 配置加载 |
| `PhaseManagers` | 1 | 管理器初始化 |
| `PhaseValidation` | 2 | 配置验证 |
| `PhaseInjection` | 3 | 依赖注入 |
| `PhaseRouter` | 4 | 路由注册 |
| `PhaseStartup` | 5 | 组件启动 |
| `PhaseRunning` | 6 | 运行中 |
| `PhaseShutdown` | 7 | 关闭中 |

## 最佳实践

### 1. 使用 CLI 工具生成应用引擎

推荐使用 CLI 工具生成容器初始化代码，确保容器注册顺序和依赖注入正确：

```go
// 由 CLI 工具自动生成 application/engine.go
func NewEngine() (*server.Engine, error) {
    entityContainer := InitEntityContainer()
    repositoryContainer := InitRepositoryContainer(entityContainer)
    serviceContainer := InitServiceContainer(repositoryContainer)
    controllerContainer := InitControllerContainer(serviceContainer)
    middlewareContainer := InitMiddlewareContainer(serviceContainer)
    listenerContainer := InitListenerContainer(serviceContainer)
    schedulerContainer := InitSchedulerContainer(serviceContainer)

    return server.NewEngine(
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
    ), nil
}
```

### 2. 中间件排序

为中间件设置合理的 Order 值，控制执行顺序：

```go
const (
    OrderRecovery        = 0   // panic 恢复
    OrderRequestLogger   = 50  // 日志记录
    OrderCORS            = 100 // 跨域处理
    OrderSecurityHeaders = 150 // 安全头
    OrderRateLimiter     = 200 // 限流
    OrderTelemetry       = 250 // 遥测监控
    OrderAuth            = 300 // 认证
)
```

### 3. 路由命名规范

控制器路由使用 OpenAPI 风格：`/path [METHOD]`

```go
// 正确
"/messages [GET]"
"/messages/:id [GET]"
"/messages [POST]"
"/messages/:id [PUT]"
"/messages/:id [DELETE]"
```

### 4. 配置管理

使用统一的配置文件管理所有组件：

```yaml
# configs/config.yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"

database:
  driver: "sqlite"
  auto_migrate: true
  sqlite_config:
    dsn: "./data/app.db"

logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "info"
      format: "gin"

scheduler:
  driver: "cron"
  cron_config:
    validate_on_startup: true
```

### 5. 错误处理

启动和停止时的错误应妥善处理：

```go
if err := engine.Run(); err != nil {
    log.Fatalf("Engine run failed: %v", err)
}
```

### 6. 自动数据库迁移

仅在开发环境启用自动迁移：

```yaml
# 开发环境
database:
  auto_migrate: true

# 生产环境
database:
  auto_migrate: false
```

## 信号处理

Engine 自动处理以下信号，触发优雅关闭：

- `SIGINT`（Ctrl+C）
- `SIGTERM`
- `SIGQUIT`

关闭流程：

1. 捕获信号
2. HTTP 服务器优雅关闭（等待现有请求完成）
3. 停止 Listener 层（反转注册顺序）
4. 停止 Scheduler 层（反转注册顺序）
5. 停止 Middleware 层（反转注册顺序）
6. 停止 Service 层（反转注册顺序）
7. 停止 Repository 层（反转注册顺序）
8. 停止 Manager 层（反转注册顺序）

## 注意事项

1. **依赖注入**：确保组件使用 `inject:""` 标签声明依赖
2. **Manager 引用**：Manager 组件位于 `manager/` 目录，导入路径为 `github.com/lite-lake/litecore-go/manager/xxxmgr`
3. **线程安全**：Engine 使用读写锁保护内部状态
4. **重入保护**：Start 方法已实现重入保护，重复调用返回错误
5. **路由定义**：所有路由必须通过 BaseController 的 `GetRouter()` 方法定义
6. **中间件顺序**：中间件按 Order 升序排序，越小的值越先执行
7. **中间件配置**：内置中间件支持通过配置自定义名称和执行顺序
8. **Scheduler 配置**：启用 `validate_on_startup` 可在启动时验证 crontab 规则
9. **数据库迁移**：生产环境建议使用专门的迁移工具，而非自动迁移
10. **启动日志**：异步日志可以提高启动性能，但可能丢失部分日志（缓冲区满时）
