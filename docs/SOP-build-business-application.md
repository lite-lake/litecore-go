# 基于 LiteCore 快速实现业务服务

## 引用私有仓库的 LiteCore

LiteCore 托管在私有 Git 仓库中，有两种方式在业务项目中使用：

### 方式一：配置 GOPRIVATE（推荐）

适用于生产环境和团队协作：

```bash
# 1. 设置私有模块前缀
export GOPRIVATE=github.com/lite-lake/litecore-go

# 2. 在新项目中引用指定版本
go mod init com.litelake.myapp
go get github.com/lite-lake/litecore-go@v0.0.1

# 3. 或使用最新版本
go get github.com/lite-lake/litecore-go@latest
```

### 方式二：使用 replace 指令

适用于本地开发和调试：

```bash
# 1. 初始化项目
go mod init com.litelake.myapp

# 2. 在 go.mod 中添加 replace 指令
# replace github.com/lite-lake/litecore-go => /Users/kentzhu/Projects/lite-lake/litecore-go

# 3. 执行依赖整理
go mod tidy

# 4. 运行应用
go run ./cmd/server
```

### 版本管理

```bash
# 查看可用版本
git -C /path/to/litecore-go tag

# 切换到特定版本
go get github.com/lite-lake/litecore-go@v0.0.1

# 更新到最新版本
go get github.com/lite-lake/litecore-go@latest
```

## 快速开始

### 1. 初始化项目

```bash
mkdir myapp && cd myapp
go mod init github.com/lite-lake/litecore-go/samples/myapp
go get github.com/lite-lake/litecore-go@latest
```

### 2. 项目结构

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
│   │   └── engine.go
│   ├── entities/                # 实体层（无依赖）
│   ├── repositories/            # 仓储层（依赖 Manager）
│   ├── services/                # 服务层（依赖 Repository）
│   ├── controllers/             # 控制器层（依赖 Service）
│   ├── middlewares/             # 中间件层（依赖 Service）
│   ├── dtos/                    # 数据传输对象
│   └── infras/                  # 基础设施（Manager 封装）
│       ├── configproviders/     # 配置提供者
│       │   └── config_provider.go
│       └── managers/            # 管理器封装
│           ├── database_manager.go
│           ├── cache_manager.go
│           └── logger_manager.go
└── go.mod
```

### 3. 配置文件（configs/config.yaml）

```yaml
app:
  name: "myapp"
  version: "1.0.0"

server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"

database:
  driver: "sqlite"              # mysql, postgresql, sqlite, none
  sqlite_config:
    dsn: "./data/myapp.db"

cache:
  driver: "memory"              # redis, memory, none

logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "info"

lock:
  driver: "memory"              # redis, memory
  redis_config:
    host: "localhost"
    port: 6379
    password: ""
    db: 0
  memory_config:
    max_backups: 1000

limiter:
  driver: "memory"              # redis, memory
  redis_config:
    host: "localhost"
    port: 6379
    password: ""
    db: 0
  memory_config:
    max_backups: 1000
```

### 4. 创建应用入口（cmd/server/main.go）

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

**注意**: `NewEngine()` 函数由代码生成器自动生成，位于 `internal/application/engine.go`。

### 5. 配置代码生成器（cmd/generate/main.go）

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
    cfg.OutputDir = "internal/application"
    cfg.PackageName = "application"
    cfg.ConfigPath = "configs/config.yaml"

    if err := generator.Run(cfg); err != nil {
        fmt.Fprintf(os.Stderr, "错误: %v\n", err)
        os.Exit(1)
    }
}
```

### 6. 初始化应用

```bash
# 创建配置目录和文件
mkdir -p configs data
touch configs/config.yaml

# 首次生成容器代码
go run ./cmd/generate

# 运行应用
go run ./cmd/server/main.go
```

## 5层架构使用规范

### 层级关系与依赖注入

依赖注入规则（由框架自动管理）：
- Entity: 无依赖
- Repository: Entity + Config + Manager（内置）
- Service: Repository + Config + Manager（内置） + Service
- Controller: Service + Config + Manager（内置）
- Middleware: Service + Config + Manager（内置）

### 1. Entity 层（实体）

**位置**: `internal/entities/`

**规范**:
- 实现 `common.IBaseEntity` 接口
- 使用 GORM 标签定义表结构
- 无外部依赖

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

### 2. Repository 层（仓储）

**位置**: `internal/repositories/`

**规范**:
- 实现 `common.IBaseRepository` 接口
- 使用 `inject:""` 注入依赖
- 在 `OnStart()` 中自动迁移表结构

```go
package repositories

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/entities"
    "github.com/lite-lake/litecore-go/server/builtin/manager/configmgr"
    "github.com/lite-lake/litecore-go/server/builtin/manager/databasemgr"
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

### 3. Service 层（服务）

**位置**: `internal/services/`

**规范**:
- 实现 `common.IBaseService` 接口
- 业务逻辑实现
- 注入 Repository 和 Manager

```go
package services

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/entities"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/repositories"
    "github.com/lite-lake/litecore-go/server/builtin/manager/configmgr"
)

type IUserService interface {
    common.IBaseService
    CreateUser(name string) (*entities.User, error)
    GetUser(id uint) (*entities.User, error)
}

type userService struct {
    Config     configmgr.IConfigManager    `inject:""`
    Repository repositories.IUserRepository `inject:""`
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
        return nil, err
    }
    return user, nil
}

func (s *userService) GetUser(id uint) (*entities.User, error) {
    return s.Repository.GetByID(id)
}

var _ IUserService = (*userService)(nil)
```

### 4. Controller 层（控制器）

**位置**: `internal/controllers/`

**规范**:
- 实现 `common.IBaseController` 接口
- 使用 `GetRouter()` 定义路由
- 使用 `Handle()` 处理请求

```go
package controllers

import (
    "github.com/gin-gonic/gin"
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/services"
    "github.com/lite-lake/litecore-go/server/builtin/manager/loggermgr"
)

// IUserController 用户控制器接口
type IUserController interface {
    common.IBaseController
}

type userController struct {
    UserService services.IUserService `inject:""`
    LoggerMgr   loggermgr.ILoggerManager `inject:""`
}

// NewUserController 创建控制器实例
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
        if c.LoggerMgr != nil {
            c.LoggerMgr.Ins().Error("创建用户失败：参数绑定失败", "error", err)
        }
        ctx.JSON(common.HTTPStatusBadRequest, gin.H{
            "code":    common.HTTPStatusBadRequest,
            "message": err.Error(),
        })
        return
    }

    user, err := c.UserService.CreateUser(req.Name)
    if err != nil {
        if c.LoggerMgr != nil {
            c.LoggerMgr.Ins().Error("创建用户失败", "error", err)
        }
        ctx.JSON(common.HTTPStatusInternalServerError, gin.H{
            "code":    common.HTTPStatusInternalServerError,
            "message": err.Error(),
        })
        return
    }

    if c.LoggerMgr != nil {
        c.LoggerMgr.Ins().Info("创建用户成功", "id", user.ID, "name", user.Name)
    }

    ctx.JSON(common.HTTPStatusOK, user)
}

var _ IUserController = (*userController)(nil)
```

### 5. Middleware 层（中间件）

**位置**: `internal/middlewares/`

**规范**:
- 实现 `common.IBaseMiddleware` 接口
- 使用 `Order()` 定义执行顺序
- 使用 `Wrapper()` 返回 gin.HandlerFunc

#### 自定义中间件

```go
package middlewares

import (
    "github.com/gin-gonic/gin"
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/services"
)

// IAuthMiddleware 认证中间件接口
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

#### 封装框架中间件

对于 `component/litemiddleware` 中的中间件，使用简洁封装方式：

```go
package middlewares

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/component/litemiddleware"
)

// ICorsMiddleware CORS 跨域中间件接口
type ICorsMiddleware interface {
    common.IBaseMiddleware
}

// NewCorsMiddleware 使用默认配置创建 CORS 中间件
func NewCorsMiddleware() ICorsMiddleware {
    return litemiddleware.NewCorsMiddlewareWithDefaults()
}

// IRequestLoggerMiddleware 请求日志中间件接口
type IRequestLoggerMiddleware interface {
    common.IBaseMiddleware
}

// NewRequestLoggerMiddleware 使用默认配置创建请求日志中间件
func NewRequestLoggerMiddleware() IRequestLoggerMiddleware {
    return litemiddleware.NewRequestLoggerMiddlewareWithDefaults()
}

// IRecoveryMiddleware panic 恢复中间件接口
type IRecoveryMiddleware interface {
    common.IBaseMiddleware
}

// NewRecoveryMiddleware 使用默认配置创建 panic 恢复中间件
func NewRecoveryMiddleware() IRecoveryMiddleware {
    return litemiddleware.NewRecoveryMiddlewareWithDefaults()
}
```

## 内置组件

### Config（配置）

Config 作为服务器内置组件，由引擎自动初始化。在创建引擎时通过 `builtin.Config` 指定配置文件路径，无需手动创建。

### Manager（管理器）

Manager 组件也作为服务器内置组件，由引擎自动初始化。无需手动创建，只需在代码中通过依赖注入使用即可。

框架自动初始化的 Manager：
- `configmgr.IConfigManager` - 配置管理
- `databasemgr.IDatabaseManager` - 数据库管理
- `cachemgr.ICacheManager` - 缓存管理
- `loggermgr.ILoggerManager` - 日志管理
- `telemetrymgr.ITelemetryManager` - 遥测管理

### 可用的内置 Manager

所有 Manager 都通过依赖注入自动初始化，在代码中通过 `inject:""` 标签使用：

- `configmgr.IConfigManager`: 配置管理
- `databasemgr.IDatabaseManager`: 数据库（MySQL/PostgreSQL/SQLite）
- `cachemgr.ICacheManager`: 缓存（Redis/Memory）
- `loggermgr.ILoggerManager`: 日志（Zap）
- `telemetrymgr.ITelemetryManager`: 遥测（OpenTelemetry）
- `lockmgr.ILockManager`: 分布式锁（Redis/Memory）
- `limitermgr.ILimiterManager`: 限流器（Redis/Memory）

### LockMgr（锁管理器）

LockMgr 提供分布式锁功能，支持 Redis 和 Memory 两种实现：

**在 Service 层使用锁：**

```go
package services

import (
    "context"
    "fmt"
    "time"

    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/server/builtin/manager/lockmgr"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/repositories"
)

type IUserService interface {
    common.IBaseService
    CreateUser(name string) error
}

type userService struct {
    Repository repositories.IUserRepository `inject:""`
    LockMgr    lockmgr.ILockManager        `inject:""`
}

func NewUserService() IUserService {
    return &userService{}
}

func (s *userService) ServiceName() string { return "UserService" }
func (s *userService) OnStart() error        { return nil }
func (s *userService) OnStop() error         { return nil }

func (s *userService) CreateUser(name string) error {
    ctx := context.Background()
    lockKey := fmt.Sprintf("user:create:%s", name)

    if err := s.LockMgr.Lock(ctx, lockKey, 10*time.Second); err != nil {
        return fmt.Errorf("获取锁失败: %w", err)
    }
    defer s.LockMgr.Unlock(ctx, lockKey)

    return nil
}

var _ IUserService = (*userService)(nil)
```

### LimiterMgr（限流管理器）

LimiterMgr 提供限流功能，支持 Redis 和 Memory 两种实现：

**在 Service 层使用限流：**

```go
package services

import (
    "context"
    "fmt"
    "time"

    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/server/builtin/manager/limitermgr"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/repositories"
)

type IUserService interface {
    common.IBaseService
    QueryData() error
}

type userService struct {
    Repository  repositories.IUserRepository `inject:""`
    LimiterMgr  limitermgr.ILimiterManager   `inject:""`
}

func NewUserService() IUserService {
    return &userService{}
}

func (s *userService) ServiceName() string { return "UserService" }
func (s *userService) OnStart() error        { return nil }
func (s *userService) OnStop() error         { return nil }

func (s *userService) QueryData() error {
    ctx := context.Background()
    key := "user:query:data"

    allowed, err := s.LimiterMgr.Allow(ctx, key, 100, time.Minute)
    if err != nil {
        return fmt.Errorf("限流检查失败: %w", err)
    }
    if !allowed {
        return fmt.Errorf("请求过于频繁，请稍后再试")
    }

    remaining, _ := s.LimiterMgr.GetRemaining(ctx, key, 100, time.Minute)
    fmt.Printf("剩余可用次数: %d\n", remaining)

    return nil
}

var _ IUserService = (*userService)(nil)
```

## 代码生成器使用

### 基本命令

```bash
# 生成容器代码
go run ./cmd/generate

# 或使用自定义参数
go run ./cmd/generate -o internal/application -pkg application -c configs/config.yaml
```

### 生成时机

- **首次创建项目**: 初始化容器代码
- **新增组件**: 添加 Entity/Repository/Service/Controller/Middleware 后
- **修改依赖**: 修改组件的 `inject` 标签后

### 生成的文件

代码生成器会自动扫描并生成以下文件：
- `entity_container.go`: 实体容器
- `repository_container.go`: 仓储容器
- `service_container.go`: 服务容器
- `controller_container.go`: 控制器容器
- `middleware_container.go`: 中间件容器
- `engine.go`: 引擎创建函数

**重要**: 生成的文件头部标记 `// Code generated by litecore/cli. DO NOT EDIT.`，请勿手动修改。

## 依赖注入规则

### 注入语法

```go
type myService struct {
    // 内置 Manager 组件（引擎自动注入）
    ConfigMgr   configmgr.IConfigManager      `inject:""`
    DBMgr       databasemgr.IDatabaseManager `inject:""`
    CacheMgr    cachemgr.ICacheManager       `inject:""`
    LoggerMgr   loggermgr.ILoggerManager     `inject:""`

    // 业务依赖
    Repository   repositories.IUserRepository  `inject:""`
    OtherService services.IOtherService      `inject:""`
}
```

### 依赖规则

| 层级 | 可注入的依赖 |
|------|-------------|
| Entity | 无 |
| Repository | Config + Manager（内置） |
| Service | Repository + Config + Manager（内置） + Service |
| Controller | Service + Config + Manager（内置） |
| Middleware | Service + Config + Manager（内置） |

### 注意事项

1. **不要跨层注入**: 例如 Controller 不能直接注入 Repository
2. **接口注入**: 优先注入接口，而非具体实现
3. **空标签**: `inject:""` 表示自动注入，无需指定名称
4. **内置组件**: 所有 Manager 组件由引擎自动初始化和注入，无需手动创建
5. **空值检查**: 使用 Manager 前应检查是否为 nil，避免 panic

```go
if m.LoggerMgr != nil {
    m.LoggerMgr.Ins().Info("处理请求")
}
```

## 最佳实践

### 1. 目录组织

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
│   │   └── engine.go
│   ├── entities/                # 实体层（无依赖）
│   ├── repositories/            # 仓储层（依赖 Manager）
│   ├── services/                # 服务层（依赖 Repository）
│   ├── controllers/             # 控制器层（依赖 Service）
│   ├── middlewares/             # 中间件层（依赖 Service）
│   └── dtos/                    # 数据传输对象
└── go.mod
```

### 2. 错误处理

```go
// 在 Service 层包装错误
return nil, fmt.Errorf("failed to create user: %w", err)

// 在 Controller 层返回 HTTP 响应
ctx.JSON(500, gin.H{"error": err.Error()})
```

### 3. 配置管理

```go
// 使用类型安全的配置获取
name, err := config.Get[string](configProvider, "app.name")
timeout, err := config.Get[int](configProvider, "app.timeout")
```

### 4. 日志记录

在业务层组件中通过依赖注入使用日志：

```go
type MyService struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
    logger     loggermgr.ILogger
}

func (s *MyService) initLogger() {
    if s.LoggerMgr != nil {
        s.logger = s.LoggerMgr.Ins()
    }
}

func (s *MyService) SomeMethod() {
    s.initLogger()
    s.logger.Info("操作完成")
}

// 注意：main函数中不要使用logger，因为LoggerMgr需要通过引擎初始化后才能使用
// 直接使用fmt和os处理错误即可
```

### 5. 数据库事务

```go
db := r.Manager.DB()
err := db.Transaction(func(tx *gorm.DB) error {
    // 在事务中执行操作
    if err := tx.Create(user).Error; err != nil {
        return err
    }
    return nil
})
```

### 6. 缓存使用

```go
// 在 Service 层使用缓存
ctx := context.Background()
cacheKey := fmt.Sprintf("user:%d", id)

var user entities.User
if err := s.CacheMgr.Get(ctx, cacheKey, &user); err == nil {
    return &user, nil
}

// 从数据库查询
user, err := s.Repository.GetByID(id)
if err != nil {
    return nil, err
}

// 写入缓存
s.CacheMgr.Set(ctx, cacheKey, user, time.Hour)
```

### 7. 中间件顺序

```go
// 0-49: 系统级（Recovery, CORS）
// 50-99: 认证授权（Auth, RBAC）
// 100-199: 日志监控（Logger, Metrics）
// 200+: 业务级（限流、自定义）

func (m *RecoveryMiddleware) Order() int { return 10 }
func (m *AuthMiddleware) Order() int    { return 100 }
func (m *LoggerMiddleware) Order() int  { return 200 }
```

### 8. 限流器中间件集成

框架提供了限流器中间件，支持基于 IP、路径、用户 ID 等多种限流策略：

**创建限流中间件：**

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
    return litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
        Limit:     100,
        Window:    time.Minute,
        KeyPrefix: "ip",
    })
}
```

**自定义限流策略：**

```go
// 自定义 KeyFunc
func NewRateLimiterByUserID() IRateLimiterMiddleware {
    return litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
        Limit:     50,
        Window:    time.Minute,
        KeyPrefix: "user",
        KeyFunc: func(c *gin.Context) string {
            // 从上下文获取用户ID（需要认证中间件先执行）
            if session, exists := c.Get("session"); exists {
                if s, ok := session.(*Session); ok {
                    return s.UserID
                }
            }
            return c.ClientIP() // 回退到IP
        },
    })
}

// 跳过某些路径
func NewRateLimiterWithSkip() IRateLimiterMiddleware {
    return litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
        Limit:     100,
        Window:    time.Minute,
        KeyPrefix: "ip",
        SkipFunc: func(c *gin.Context) bool {
            return c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics"
        },
    })
}
```

**限流中间件 Order 建议值：** 200（OrderRateLimiter）

框架提供了限流器中间件，支持基于 IP、路径、用户 ID 等多种限流策略：

**创建限流中间件：**

```go
package middlewares

import (
    "time"

    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/component/middleware"
)

type IRateLimiterMiddleware interface {
    common.IBaseMiddleware
}

type rateLimiterMiddleware struct {
    inner common.IBaseMiddleware
}

func NewRateLimiterMiddleware() IRateLimiterMiddleware {
    return &rateLimiterMiddleware{
        inner: middleware.NewRateLimiterByIP(100, time.Minute),
    }
}

func (m *rateLimiterMiddleware) MiddlewareName() string { return "RateLimiterMiddleware" }
func (m *rateLimiterMiddleware) Order() int              { return m.inner.Order() }
func (m *rateLimiterMiddleware) Wrapper() func(c *gin.Context) {
    return m.inner.Wrapper()
}
func (m *rateLimiterMiddleware) OnStart() error { return nil }
func (m *rateLimiterMiddleware) OnStop() error  { return nil }

var _ IRateLimiterMiddleware = (*rateLimiterMiddleware)(nil)
```

**其他限流策略：**

```go
// 按路径限流
middleware.NewRateLimiterByPath(1000, time.Minute)

// 按用户ID限流（需在认证中间件后执行）
middleware.NewRateLimiterByUserID(50, time.Minute)

// 按Header限流
middleware.NewRateLimiterByHeader(100, time.Minute, "X-API-Key")

// 自定义配置
middleware.NewRateLimiter(&middleware.RateLimiterConfig{
    Limit:     100,
    Window:    time.Minute,
    KeyPrefix: "custom",
    KeyFunc: func(c *gin.Context) string {
        return c.GetHeader("X-Custom-Header")
    },
    SkipFunc: func(c *gin.Context) bool {
        return c.Request.URL.Path == "/health"
    },
})
```

**限流中间件 Order 建议值：** 90（在认证之前执行，避免认证后的请求过多）

### 9. 测试建议

```go
// 单元测试：使用 Mock 依赖
type mockUserRepository struct {
    repositories.IUserRepository
    // mock 方法
}

// 集成测试：使用 SQLite 内存数据库
sqlite_config:
  dsn: ":memory:"
```

### 10. 锁和限流的使用建议

**锁的使用场景：**

1. **防止重复操作**：如用户创建时防止同名用户重复创建
2. **资源竞争保护**：如库存扣减、订单生成等需要保证原子性的场景
3. **分布式环境**：多实例部署时使用 Redis 锁，单实例可使用内存锁

**限流的使用场景：**

1. **API 保护**：防止恶意请求或突发流量导致服务崩溃
2. **资源配额**：按用户、IP 等维度限制访问频率
3. **服务降级**：在高负载时拒绝部分请求，保护核心功能

**配置建议：**

```yaml
# 生产环境建议使用 Redis 实现锁和限流
lock:
  driver: "redis"
  redis_config:
    host: "redis.example.com"
    port: 6379

limiter:
  driver: "redis"
  redis_config:
    host: "redis.example.com"
    port: 6379
```

**性能建议：**

1. 锁的 TTL 应合理设置，避免死锁
2. 限流窗口不宜过小，避免性能损耗
3. 高频接口建议使用限流器中间件，而非手动检查
4. 热点数据使用 Redis 实现，避免内存限制

## 常见问题

### 1. 如何添加新的 Manager？

所有 Manager 由框架提供，无需手动创建。在代码中通过 `inject:""` 标签直接使用即可。

### 2. 如何自定义路由？

Controller 的 `GetRouter()` 支持完整的路由语法：
```go
return "/api/users/:id [GET]"
return "/api/users [POST]"
return "/api/files/*filepath [GET]"
```

### 3. 如何使用框架提供的中间件？

框架中间件已封装在 `component/litemiddleware` 中，直接使用 `WithDefaults` 函数：

```go
package middlewares

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/component/litemiddleware"
)

// IRecoveryMiddleware panic 恢复中间件接口
type IRecoveryMiddleware interface {
    common.IBaseMiddleware
}

// NewRecoveryMiddleware 使用默认配置创建 panic 恢复中间件
func NewRecoveryMiddleware() IRecoveryMiddleware {
    return litemiddleware.NewRecoveryMiddlewareWithDefaults()
}
```

### 4. 如何支持多种数据库？

在 `configs/config.yaml` 中切换 `database.driver`，无需修改代码：
```yaml
database:
  driver: "mysql"  # 或 "postgresql", "sqlite"
  mysql_config:
    dsn: "user:pass@tcp(localhost:3306)/dbname"
```

### 5. 如何调试依赖注入？

在 `engine.go` 中查看注入过程，或在组件的构造函数中添加日志：
```go
func NewUserService() IUserService {
    fmt.Println("[DEBUG] NewUserService called")
    return &userService{}
}
```

### 6. 如何处理循环依赖？

LiteCore 的依赖注入不支持循环依赖。解决方法：
- 重构代码，消除循环依赖
- 使用事件驱动架构
- 将共享逻辑提取到独立的服务

### 7. 如何热重载开发？

```bash
# 安装 air
go install github.com/cosmtrek/air@latest

# 初始化配置
air init

# 运行
air
```

### 8. 如何部署应用？

```bash
# 构建二进制文件
go build -o myapp ./cmd/server/main.go

# 运行
./myapp
```
