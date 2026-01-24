# 基于 LiteCore 快速实现业务服务

## 概述

 LiteCore 是一个基于 Go 1.25+ 的轻量级 Web 框架，采用 5 层分层架构（内置管理器层 → Entity → Repository → Service → 交互层）和依赖注入设计，内置了 Gin、GORM、 Zap 等常用组件，帮助开发者快速构建企业级应用。

 ### 核心特性

 - **5 层分层架构**：清晰的依赖关系，遵循单向依赖原则
- **依赖注入**：自动管理组件依赖，通过 `inject:""` 标签注入
- **内置组件**：Manager 组件自动初始化并注入，开箱即用
- **代码生成**：自动生成容器代码，减少重复工作
- **多种数据库支持**：MySQL、PostgreSQL、SQLite 无缝切换
- **可观测性**：日志、指标、链路追踪支持
- **限流与锁**：基于 Redis/Memory 的分布式限流和锁
- **消息监听**：支持 RabbitMQ/Memory 消息队列监听
- **定时任务**：基于 Cron 表达式的定时任务调度

### 项目结构

 ```
 myapp/
 ├── cmd/
 │   ├── server/main.go          # 应用入口
 │   └── generate/main.go         # 代码生成器
 ├── configs/config.yaml          # 配置文件
 ├── internal/
 │   ├── application/             # 自动生成的容器（DO NOT EDIT）
 │   │   ├── entity_container.go
 │   │   ├── repository_container.go
 │   │   ├── service_container.go
 │   │   ├── controller_container.go
 │   │   ├── middleware_container.go
 │   │   ├── listener_container.go
 │   │   ├── scheduler_container.go
 │   │   └── engine.go
 │   ├── entities/                # 实体层（无依赖）
 │   ├── repositories/            # 仓储层（依赖 Manager）
 │   ├── services/                # 服务层（依赖 Repository）
 │   ├── controllers/             # 交互层 - 控制器（依赖 Service）
 │   ├── middlewares/             # 交互层 - 中间件（依赖 Service）
 │   ├── listeners/               # 交互层 - 监听器（依赖 Service）
 │   ├── schedulers/              # 交互层 - 定时器（依赖 Service）
 │   └── dtos/                    # 数据传输对象
 └── go.mod
 ```

## 规范

### 1. 引用 LiteCore

#### 方式一：配置 GOPRIVATE（推荐）

```bash
# 设置私有模块前缀
export GOPRIVATE=github.com/lite-lake/litecore-go

# 引用指定版本
go get github.com/lite-lake/litecore-go@v0.0.1
```

#### 方式二：使用 replace 指令

```go
// go.mod
module com.litelake.myapp

go 1.25

replace github.com/lite-lake/litecore-go => /Users/kentzhu/Projects/lite-lake/litecore-go

require github.com/lite-lake/litecore-go v0.0.0
```

### 2. 配置文件规范

配置文件采用 YAML 格式，位于 `configs/config.yaml`：

```yaml
# 应用配置
app:
  name: "myapp"
  version: "1.0.0"

# 服务器配置
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"
  startup_log:
    enabled: true
    async: true
    buffer: 100

# 数据库配置
database:
  driver: "sqlite"              # mysql, postgresql, sqlite, none
  sqlite_config:
    dsn: "./data/myapp.db"
  observability_config:
    slow_query_threshold: "1s"
    log_sql: false
    sample_rate: 1.0

# 缓存配置
cache:
  driver: "memory"              # redis, memory, none
  memory_config:
    max_size: 100
    max_age: "720h"

# 日志配置
logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "info"
      format: "gin"             # gin | json | default
      color: true
      time_format: "2006-01-02 15:04:05.000"

# 遥测配置
telemetry:
  driver: "none"               # none, otel

# 锁配置
lock:
  driver: "memory"              # redis, memory

# 限流配置
limiter:
  driver: "memory"              # redis, memory

# 消息队列配置
mq:
  driver: "memory"              # rabbitmq, memory

# 定时任务配置
scheduler:
  driver: "cron"                # cron
  cron_config:
    validate_on_startup: true   # 启动时是否检查所有 Scheduler 配置
```

### 3. 应用入口规范

```go
package main

import (
    "fmt"
    "os"
    app "github.com/lite-lake/litecore-go/samples/myapp/internal/application"
)

func main() {
    engine, err := app.NewEngine()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create engine: %v\n", err)
        os.Exit(1)
    }

    if err := engine.Initialize(); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to initialize engine: %v\n", err)
        os.Exit(1)
    }

    if err := engine.Start(); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to start engine: %v\n", err)
        os.Exit(1)
    }

    engine.WaitForShutdown()
}
```

### 4. 代码生成器规范

```go
package main

import (
    "flag"
    "fmt"
    "os"

    "github.com/lite-lake/litecore-go/cli/generator"
)

func main() {
    cfg := generator.DefaultConfig()

    outputDir := flag.String("o", cfg.OutputDir, "输出目录")
    packageName := flag.String("pkg", cfg.PackageName, "包名")
    configPath := flag.String("c", cfg.ConfigPath, "配置文件路径")

    flag.Parse()

    if outputDir != nil {
        cfg.OutputDir = *outputDir
    }
    if packageName != nil {
        cfg.PackageName = *packageName
    }
    if configPath != nil {
        cfg.ConfigPath = *configPath
    }

    if err := generator.Run(cfg); err != nil {
        fmt.Fprintf(os.Stderr, "错误: %v\n", err)
        os.Exit(1)
    }
}
```

### 5. 7 层架构使用规范

#### 依赖注入规则

| 层级 | 可注入的依赖 |
|------|-------------|
| Entity | 无 |
| Repository | Config + Manager（内置） |
| Service | Repository + Config + Manager（内置） + Service |
| Controller | Service + Config + Manager（内置） |
| Middleware | Service + Config + Manager（内置） |
| Listener | Service + Config + Manager（内置） |
| Scheduler | Service + Config + Manager（内置） |

#### Entity 层规范

**位置**: `internal/entities/`

```go
package entities

import (
    "fmt"
    "time"
    "github.com/lite-lake/litecore-go/common"
)

type User struct {
    ID        uint      `gorm:"primarykey" json:"id"`
    Name      string    `gorm:"type:varchar(50);not null" json:"name"`
    CreatedAt time.Time `json:"created_at"`
}

func (u *User) EntityName() string { return "User" }
func (u *User) TableName() string  { return "users" }
func (u *User) GetId() string       { return fmt.Sprintf("%d", u.ID) }

var _ common.IBaseEntity = (*User)(nil)
```

#### Repository 层规范

**位置**: `internal/repositories/`

```go
package repositories

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/configmgr"
    "github.com/lite-lake/litecore-go/manager/databasemgr"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/entities"
)

type IUserRepository interface {
    common.IBaseRepository
    Create(user *entities.User) error
    GetByID(id uint) (*entities.User, error)
}

type userRepository struct {
    Config  configmgr.IConfigManager     `inject:""`
    Manager databasemgr.IDatabaseManager `inject:""`
}

func NewUserRepository() IUserRepository {
    return &userRepository{}
}

func (r *userRepository) RepositoryName() string { return "UserRepository" }

func (r *userRepository) OnStart() error {
    return r.Manager.AutoMigrate(&entities.User{})
}

func (r *userRepository) OnStop() error { return nil }

func (r *userRepository) Create(user *entities.User) error {
    return r.Manager.DB().Create(user).Error
}

func (r *userRepository) GetByID(id uint) (*entities.User, error) {
    var user entities.User
    err := r.Manager.DB().First(&user, id).Error
    return &user, err
}

var _ IUserRepository = (*userRepository)(nil)
```

#### Service 层规范

**位置**: `internal/services/`

```go
package services

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/configmgr"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/entities"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/repositories"
)

type IUserService interface {
    common.IBaseService
    CreateUser(name string) (*entities.User, error)
    GetUser(id uint) (*entities.User, error)
}

type userService struct {
    Config     configmgr.IConfigManager    `inject:""`
    Repository repositories.IUserRepository `inject:""`
    LoggerMgr  loggermgr.ILoggerManager    `inject:""`
}

func NewUserService() IUserService {
    return &userService{}
}

func (s *userService) ServiceName() string { return "UserService" }
func (s *userService) OnStart() error        { return nil }
func (s *userService) OnStop() error         { return nil }

func (s *userService) CreateUser(name string) (*entities.User, error) {
    user := &entities.User{Name: name}
    if err := s.Repository.Create(user); err != nil {
        s.LoggerMgr.Ins().Error("创建用户失败", "name", name, "error", err)
        return nil, err
    }
    s.LoggerMgr.Ins().Info("创建用户成功", "id", user.ID, "name", user.Name)
    return user, nil
}

func (s *userService) GetUser(id uint) (*entities.User, error) {
    return s.Repository.GetByID(id)
}

var _ IUserService = (*userService)(nil)
```

#### Controller 层规范

**位置**: `internal/controllers/`

```go
package controllers

import (
    "github.com/gin-gonic/gin"
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/dtos"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/services"
)

type IUserController interface {
    common.IBaseController
}

type userController struct {
    UserService services.IUserService `inject:""`
    LoggerMgr  loggermgr.ILoggerManager `inject:""`
}

func NewUserController() IUserController {
    return &userController{}
}

func (c *userController) ControllerName() string {
    return "userController"
}

func (c *userController) GetRouter() string {
    return "/api/users [POST]"
}

func (c *userController) Handle(ctx *gin.Context) {
    var req struct {
        Name string `json:"name" binding:"required"`
    }
    if err := ctx.ShouldBindJSON(&req); err != nil {
        c.LoggerMgr.Ins().Error("创建用户失败：参数绑定失败", "error", err)
        ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
        return
    }

    user, err := c.UserService.CreateUser(req.Name)
    if err != nil {
        c.LoggerMgr.Ins().Error("创建用户失败", "name", req.Name, "error", err)
        ctx.JSON(common.HTTPStatusInternalServerError, dtos.ErrorResponse(common.HTTPStatusInternalServerError, err.Error()))
        return
    }

    c.LoggerMgr.Ins().Info("创建用户成功", "id", user.ID, "name", user.Name)

    ctx.JSON(common.HTTPStatusOK, dtos.SuccessResponse("创建成功", gin.H{"id": user.ID}))
}

var _ IUserController = (*userController)(nil)
```

#### Middleware 层规范

**位置**: `internal/middlewares/`

##### 自定义中间件

```go
package middlewares

import (
    "github.com/gin-gonic/gin"
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/services"
)

type IAuthMiddleware interface {
    common.IBaseMiddleware
}

type authMiddleware struct {
    AuthService services.IAuthService `inject:""`
}

func NewAuthMiddleware() IAuthMiddleware {
    return &authMiddleware{}
}

func (m *authMiddleware) MiddlewareName() string {
    return "AuthMiddleware"
}

func (m *authMiddleware) Order() int {
    return 100
}

func (m *authMiddleware) Wrapper() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(common.HTTPStatusUnauthorized, gin.H{
                "code":    common.HTTPStatusUnauthorized,
                "message": "未提供认证令牌",
            })
            c.Abort()
            return
        }

        session, err := m.AuthService.ValidateToken(authHeader)
        if err != nil {
            c.JSON(common.HTTPStatusUnauthorized, gin.H{
                "code":    common.HTTPStatusUnauthorized,
                "message": "认证令牌无效或已过期",
            })
            c.Abort()
            return
        }

        c.Set("session", session)
        c.Next()
    }
}

func (m *authMiddleware) OnStart() error { return nil }
func (m *authMiddleware) OnStop() error  { return nil }

var _ IAuthMiddleware = (*authMiddleware)(nil)
```

##### 封装框架中间件

```go
package middlewares

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/component/litemiddleware"
)

type ICorsMiddleware interface {
    common.IBaseMiddleware
}

func NewCorsMiddleware() ICorsMiddleware {
    return litemiddleware.NewCorsMiddlewareWithDefaults()
}

type IRequestLoggerMiddleware interface {
    common.IBaseMiddleware
}

func NewRequestLoggerMiddleware() IRequestLoggerMiddleware {
    return litemiddleware.NewRequestLoggerMiddlewareWithDefaults()
}

type IRecoveryMiddleware interface {
    common.IBaseMiddleware
}

func NewRecoveryMiddleware() IRecoveryMiddleware {
    return litemiddleware.NewRecoveryMiddlewareWithDefaults()
}
```

##### 限流中间件

```go
package middlewares

import (
    "time"

    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/component/litemiddleware"
)

type IRateLimiterMiddleware interface {
    common.IBaseMiddleware
}

func NewRateLimiterMiddleware() IRateLimiterMiddleware {
    limit := 100
    window := time.Minute
    keyPrefix := "ip"
    return litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
        Limit:     &limit,
        Window:    &window,
        KeyPrefix: &keyPrefix,
    })
}
```

#### Listener 层规范

**位置**: `internal/listeners/`

```go
package listeners

import (
    "context"

    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
)

type IMessageCreatedListener interface {
    common.IBaseListener
}

type messageCreatedListenerImpl struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func NewMessageCreatedListener() IMessageCreatedListener {
    return &messageCreatedListenerImpl{}
}

func (l *messageCreatedListenerImpl) ListenerName() string {
    return "MessageCreatedListener"
}

func (l *messageCreatedListenerImpl) GetQueue() string {
    return "message.created"
}

func (l *messageCreatedListenerImpl) GetSubscribeOptions() []common.ISubscribeOption {
    return []common.ISubscribeOption{}
}

func (l *messageCreatedListenerImpl) OnStart() error {
    l.LoggerMgr.Ins().Info("Message created listener started")
    return nil
}

func (l *messageCreatedListenerImpl) OnStop() error {
    l.LoggerMgr.Ins().Info("Message created listener stopped")
    return nil
}

func (l *messageCreatedListenerImpl) Handle(ctx context.Context, msg common.IMessageListener) error {
    l.LoggerMgr.Ins().Info("Received message created event",
        "message_id", msg.ID(),
        "body", string(msg.Body()),
        "headers", msg.Headers())
    return nil
}

var _ IMessageCreatedListener = (*messageCreatedListenerImpl)(nil)
var _ common.IBaseListener = (*messageCreatedListenerImpl)(nil)
```

#### Scheduler 层规范

**位置**: `internal/schedulers/`

```go
package schedulers

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/logger"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
    "github.com/lite-lake/litecore-go/myapp/internal/services"
)

type IStatisticsScheduler interface {
    common.IBaseScheduler
}

type statisticsSchedulerImpl struct {
    MessageService services.IMessageService `inject:""`
    LoggerMgr      loggermgr.ILoggerManager `inject:""`
    logger         logger.ILogger
}

func NewStatisticsScheduler() IStatisticsScheduler {
    return &statisticsSchedulerImpl{}
}

func (s *statisticsSchedulerImpl) SchedulerName() string {
    return "statisticsScheduler"
}

func (s *statisticsSchedulerImpl) GetRule() string {
    return "0 0 * * * *"
}

func (s *statisticsSchedulerImpl) GetTimezone() string {
    return "Asia/Shanghai"
}

func (s *statisticsSchedulerImpl) OnTick(tickID int64) error {
    s.initLogger()
    s.logger.Info("Starting statistics task", "tick_id", tickID)

    stats, err := s.MessageService.GetStatistics()
    if err != nil {
        s.logger.Error("Failed to get statistics", "error", err)
        return err
    }

    s.logger.Info("Statistics task completed", "tick_id", tickID, "total", stats["total"])
    return nil
}

func (s *statisticsSchedulerImpl) OnStart() error {
    s.initLogger()
    s.logger.Info("Statistics scheduler started")
    return nil
}

func (s *statisticsSchedulerImpl) OnStop() error {
    s.LoggerMgr.Ins().Info("Statistics scheduler stopped")
    return nil
}

var _ IStatisticsScheduler = (*statisticsSchedulerImpl)(nil)
var _ common.IBaseScheduler = (*statisticsSchedulerImpl)(nil)
```

### 6. 内置组件使用规范

#### 可用的内置 Manager

所有 Manager 都通过依赖注入自动初始化，在代码中通过 `inject:""` 标签使用：

- `configmgr.IConfigManager`: 配置管理
- `databasemgr.IDatabaseManager`: 数据库（MySQL/PostgreSQL/SQLite）
- `cachemgr.ICacheManager`: 缓存（Redis/Memory）
- `loggermgr.ILoggerManager`: 日志（Zap）
- `telemetrymgr.ITelemetryManager`: 遥测（OpenTelemetry）
- `lockmgr.ILockManager`: 分布式锁（Redis/Memory）
- `limitermgr.ILimiterManager`: 限流器（Redis/Memory）
- `mqmgr.IMQManager`: 消息队列（RabbitMQ/Memory）
- `schedulermgr.ISchedulerManager`: 定时任务调度器（Cron）

#### 日志使用

```go
type MyService struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func (s *MyService) SomeMethod() {
    s.LoggerMgr.Ins().Info("操作完成", "key", "value")
    s.LoggerMgr.Ins().Error("操作失败", "error", err)
}
```

#### 缓存使用

```go
ctx := context.Background()
cacheKey := fmt.Sprintf("user:%d", id)

var user entities.User
if err := s.CacheMgr.Get(ctx, cacheKey, &user); err == nil {
    return &user, nil
}

user, err := s.Repository.GetByID(id)
if err != nil {
    return nil, err
}

s.CacheMgr.Set(ctx, cacheKey, user, time.Hour)
```

#### 锁使用

```go
ctx := context.Background()
lockKey := fmt.Sprintf("user:create:%s", name)

if err := s.LockMgr.Lock(ctx, lockKey, 10*time.Second); err != nil {
    return fmt.Errorf("获取锁失败: %w", err)
}
defer s.LockMgr.Unlock(ctx, lockKey)
```

#### 限流使用

```go
ctx := context.Background()
key := "user:query:data"

allowed, err := s.LimiterMgr.Allow(ctx, key, 100, time.Minute)
if err != nil {
    return fmt.Errorf("限流检查失败: %w", err)
}
if !allowed {
    return fmt.Errorf("请求过于频繁，请稍后再试")
}
```

#### 消息队列使用

消息队列用于发布订阅模式的消息处理。Listener 组件自动订阅指定队列并处理消息。

**发布消息（Service 层）**:

```go
type MyService struct {
    MQMgr mqmgr.IMQManager `inject:""`
}

func (s *MyService) CreateMessage(message string) error {
    ctx := context.Background()

    err := s.MQMgr.Publish(ctx, "message.created", []byte(message))
    if err != nil {
        return fmt.Errorf("发布消息失败: %w", err)
    }

    return nil
}
```

**监听消息（Listener 层）**:

见上文 Listener 层规范部分，Listener 会自动订阅 `GetQueue()` 返回的队列名称，并通过 `Handle()` 方法处理消息。

## 注意点

### 1. 依赖注入

- 不要跨层注入：例如 Controller 不能直接注入 Repository
- 接口注入：优先注入接口，而非具体实现
- 空标签：`inject:""` 表示自动注入，无需指定名称
- 直接使用：Manager 一定会被注入，无需 nil 检查

### 2. 代码生成

- 生成时机：首次创建项目、新增组件、修改依赖后
- 生成的文件头部标记 `// Code generated by litecore/cli. DO NOT EDIT.`，请勿手动修改
- 运行 `go run ./cmd/generate` 生成容器代码

### 3. 日志使用

- main 函数中不要使用 logger，使用 fmt 和 os 处理错误
- 业务层组件通过依赖注入使用日志
- 敏感信息（密码、token、密钥等）必须脱敏

### 4. 数据库事务

```go
db := r.Manager.DB()
err := db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(user).Error; err != nil {
        return err
    }
    return nil
})
```

### 5. 中间件顺序

```go
// 0-49: 系统级（Recovery, CORS）
// 50-99: 认证授权（Auth）
// 100-199: 日志监控（Logger, Metrics）
// 200-299: 限流、安全（RateLimiter, Security）
// 300+: 业务级（自定义）
```

### 6. 配置和构建命令

```bash
go build -o myapp ./...
go test ./...
go test ./util/jwt
go test -v ./util/jwt -run TestGenerateHS256Token
go test -bench=. ./util/hash
go fmt ./...
go vet ./...
go mod tidy
```

### 7. 中间件配置支持

所有中间件支持通过配置自定义 Name 和 Order：

```go
func NewCustomRateLimiterMiddleware() IRateLimiterMiddleware {
    name := "CustomRateLimiter"
    order := 250
    limit := 50
    window := time.Minute
    keyPrefix := "user"

    return litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
        Name:      &name,
        Order:     &order,
        Limit:     &limit,
        Window:    &window,
        KeyPrefix: &keyPrefix,
    })
}
```

## 完整示例

### Messageboard 留言板项目

框架提供了一个完整的留言板示例项目 `samples/messageboard`，展示了如何使用所有核心功能：

#### 功能特性

- 用户留言（支持昵称和内容验证）
- 留言审核（pending/approved/rejected）
- 管理员登录（基于 session）
- 限流中间件（按 IP 限流）
- 数据库自动迁移
- 日志记录（Gin 格式）
- 静态资源服务
- 消息监听器（留言创建、审核事件）
- 定时任务（统计任务、清理任务）

#### 运行示例项目

```bash
cd samples/messageboard

# 创建数据目录
mkdir -p data

# 运行应用
go run ./cmd/server/main.go
```

访问 http://localhost:8080 查看留言板。

#### 示例项目结构

```
samples/messageboard/
├── cmd/
│   ├── server/main.go          # 应用入口
│   ├── generate/main.go         # 代码生成器
│   └── genpasswd/main.go       # 密码生成工具
├── configs/config.yaml          # 配置文件
├── internal/
│   ├── application/             # 自动生成的容器
│   ├── entities/                # 实体层
│   │   └── message_entity.go
│   ├── repositories/            # 仓储层
│   │   └── message_repository.go
│   ├── services/                # 服务层
│   │   ├── message_service.go
│   │   ├── auth_service.go
│   │   └── session_service.go
│   ├── controllers/             # 控制器层
│   │   ├── msg_create_controller.go
│   │   ├── msg_list_controller.go
│   │   └── ...
│   ├── middlewares/             # 中间件层
│   │   ├── auth_middleware.go
│   │   ├── rate_limiter_middleware.go
│   │   └── ...
│   ├── listeners/               # 消息监听层
│   │   ├── message_created_listener.go
│   │   └── message_audit_listener.go
│   ├── schedulers/              # 定时任务层
│   │   ├── statistics_scheduler.go
│   │   └── cleanup_scheduler.go
│   └── dtos/                    # 数据传输对象
└── go.mod
```

通过学习示例项目，可以更好地理解如何构建基于 LiteCore 的业务应用。
