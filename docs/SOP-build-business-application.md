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
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/services"
    "github.com/lite-lake/litecore-go/server/builtin/manager/configmgr"
    "github.com/gin-gonic/gin"
)

type IUserController interface {
    common.IBaseController
}

type userController struct {
    Config      configmgr.IConfigManager `inject:""`
    UserService services.IUserService    `inject:""`
}

func NewUserController() IUserController {
    return &userController{}
}

func (c *userController) ControllerName() string { return "userController" }

func (c *userController) GetRouter() string {
    return "/api/users [POST]"
}

func (c *userController) Handle(ctx *gin.Context) {
    var req struct {
        Name string `json:"name" binding:"required"`
    }
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(400, gin.H{"error": err.Error()})
        return
    }

    user, err := c.UserService.CreateUser(req.Name)
    if err != nil {
        ctx.JSON(500, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(200, user)
}

var _ IUserController = (*userController)(nil)
```

### 5. Middleware 层（中间件）

**位置**: `internal/middlewares/`

**规范**:
- 实现 `common.IBaseMiddleware` 接口
- 使用 `Order()` 定义执行顺序
- 使用 `Wrapper()` 返回 gin.HandlerFunc

```go
package middlewares

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/server/builtin/manager/configmgr"
    "github.com/gin-gonic/gin"
)

type ILoggerMiddleware interface {
    common.IBaseMiddleware
}

type loggerMiddleware struct {
    order int
}

func NewLoggerMiddleware() ILoggerMiddleware {
    return &loggerMiddleware{order: 100}
}

func (m *loggerMiddleware) MiddlewareName() string { return "LoggerMiddleware" }
func (m *loggerMiddleware) Order() int              { return m.order }

func (m *loggerMiddleware) Wrapper() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 前置处理
        c.Next()
        // 后置处理
    }
}

func (m *loggerMiddleware) OnStart() error { return nil }
func (m *loggerMiddleware) OnStop() error  { return nil }

var _ ILoggerMiddleware = (*loggerMiddleware)(nil)
```

## 内置组件

### Config（配置）

Config 作为服务器内置组件，由引擎自动初始化。在创建引擎时通过 `builtin.Config` 指定配置文件，无需手动创建。

### Manager（管理器）

Manager 组件也作为服务器内置组件，由引擎自动初始化。无需手动创建，只需在代码中通过依赖注入使用即可。

框架自动初始化的 Manager：
- `configmgr.IConfigManager` - 配置管理
- `databasemgr.IDatabaseManager` - 数据库管理
- `cachemgr.ICacheManager` - 缓存管理
- `loggermgr.ILoggerManager` - 日志管理
- `telemetrymgr.ITelemetryManager` - 遥测管理

### 可用的内置 Manager

- `databasemgr.DatabaseManager`: 数据库（MySQL/PostgreSQL/SQLite）
- `cachemgr.CacheManager`: 缓存（Redis/Memory）
- `loggermgr.LoggerManager`: 日志（Zap）
- `telemetrymgr.TelemetryManager`: 遥测（OpenTelemetry）

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

**重要**: 生成的文件头部标记 `// Code generated by litecore/cli. DO NOT EDIT.`，请勿手动修改。

## 依赖注入规则

### 注入语法

```go
import "github.com/lite-lake/litecore-go/server/builtin/manager/configmgr"

type myService struct {
    // 内置组件（引擎自动注入）
    Config         configmgr.IConfigManager      `inject:""`
    DBMgr          databasemgr.IDatabaseManager `inject:""`
    CacheMgr       cachemgr.ICacheManager       `inject:""`

    // 业务依赖
    Repository     repositories.IUserRepository  `inject:""`
    OtherService   services.IOtherService       `inject:""`
}
```

### 依赖规则

| 层级 | 可注入的依赖 |
|------|-------------|
| Repository | Entity, Config, Manager（内置） |
| Service | Repository, Config, Manager（内置）, Service |
| Controller | Service, Config, Manager（内置） |
| Middleware | Service, Config, Manager（内置） |

### 注意事项

1. **不要跨层注入**: 例如 Controller 不能直接注入 Repository
2. **接口注入**: 优先注入接口，而非具体实现
3. **空标签**: `inject:""` 表示自动注入，无需指定名称
4. **内置组件**: Config 和 Manager 由引擎自动初始化和注入，无需手动创建

## 最佳实践

### 1. 目录组织

```
internal/
├── application/         # 自动生成，不要手动修改
├── entities/           # 纯数据实体，无业务逻辑
├── repositories/       # 数据访问层，仅 CRUD
├── services/           # 业务逻辑层，验证、事务、业务规则
├── controllers/        # HTTP 层，仅请求响应处理
├── middlewares/        # 中间件，横切关注点
├── dtos/               # 请求/响应对象
└── infras/             # 基础设施，封装框架组件
    ├── configproviders/ # 配置提供者
    └── managers/        # 管理器封装
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

### 8. 测试建议

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

## 常见问题

### 1. 如何添加新的 Manager？

在 `internal/infras/managers/` 创建新的 Manager 文件，然后运行 `go run ./cmd/generate`。

### 2. 如何自定义路由？

Controller 的 `GetRouter()` 支持完整的路由语法：
```go
return "/api/users/:id [GET]"
return "/api/users [POST]"
return "/api/files/*filepath [GET]"
```

### 3. 如何使用框架提供的中间件？

```go
import "github.com/lite-lake/litecore-go/component/middleware"

type recoveryMiddleware struct {
    inner common.IBaseMiddleware
}

func NewRecoveryMiddleware() IRecoveryMiddleware {
    return &recoveryMiddleware{
        inner: middleware.NewRecoveryMiddleware(),
    }
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
