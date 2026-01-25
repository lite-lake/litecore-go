# LiteCore-Go

 基于 Gin + GORM + Zap 的 Go Web 开发框架，采用 5 层分层架构（内置管理器层 → Entity → Repository → Service → 交互层）和声明式依赖注入。

## 核心特性

 - **5 层分层架构** - 内置管理器层 → Entity → Repository → Service → 交互层（Controller/Middleware/Listener/Scheduler），清晰的责任分离
 - **实体基类** - 提供 3 种预定义基类，支持 CUID2 ID 自动生成和时间戳自动填充
 - **内置 Manager 组件** - Config、Telemetry、Logger、Database、Cache、Lock、Limiter、MQ、Scheduler 等 9 个管理器，自动初始化和注入
 - **声明式依赖注入** - 使用 `inject:""` 标签自动注入依赖，支持同层依赖和循环依赖检测
 - **统一配置管理** - 支持 YAML/JSON 配置文件，类型安全的配置读取
 - **多驱动支持** - 数据库（MySQL/PostgreSQL/SQLite）、缓存（Redis/Memory）、限流（Redis/Memory）、锁（Redis/Memory）、MQ（RabbitMQ/Memory）
 - **内置中间件** - Recovery、日志、CORS、安全头、限流、遥测等开箱即用
 - **可观测性** - 集成 OpenTelemetry、结构化日志、健康检查、指标采集
 - **Gin 格式日志** - 支持可配置日志格式（gin/json/default），彩色输出，结构化记录
 - **CLI 代码生成** - 自动生成容器初始化代码，简化项目搭建
 - **生命周期管理** - 统一的启动/停止机制，优雅关闭

## 快速开始

### 安装

```bash
go get github.com/lite-lake/litecore-go
```

### 创建项目

```bash
# 1. 初始化项目
mkdir myapp && cd myapp
go mod init com.litelake.myapp
go get github.com/lite-lake/litecore-go

 # 2. 创建目录结构
 mkdir -p internal/{entities,repositories,services,controllers,middlewares,listeners,schedulers}
 mkdir -p cmd/{server,generate}
 mkdir -p {configs,data}

# 3. 创建配置文件
cat > configs/config.yaml <<EOF
# 应用配置
app:
  name: "myapp"
  version: "1.0.0"

# 服务器配置
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"
  read_timeout: "10s"
  write_timeout: "10s"
  idle_timeout: "60s"
  shutdown_timeout: "30s"

# 数据库配置
database:
  driver: "sqlite"
  auto_migrate: true
  sqlite_config:
    dsn: "./data/myapp.db"

# 缓存配置
cache:
  driver: "memory"
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
      format: "gin"
      color: true

# 限流配置
limiter:
  driver: "memory"
  memory_config:
    max_backups: 1000

# 锁配置
lock:
  driver: "memory"
  memory_config:
    max_backups: 1000

# 消息队列配置
mq:
  driver: "memory"
  memory_config:
    max_queue_size: 10000
    channel_buffer: 100

# 定时任务配置
scheduler:
  driver: "cron"
  cron_config:
    validate_on_startup: true
EOF

 # 4. 创建生成器入口
cat > cmd/generate/main.go <<'MAINEOF'
package main

import (
    "github.com/lite-lake/litecore-go/cli/generator"
)

func main() {
    generator.MustRun(generator.DefaultConfig())
}
MAINEOF

# 5. 运行生成器（生成容器初始化代码）
go run ./cmd/generate

# 6. 创建应用入口
cat > cmd/server/main.go <<'SERVEREOF'
package main

import (
    "os"

    "github.com/lite-lake/litecore-go/server"
    "github.com/lite-lake/litecore-go/container"
)

 func main() {
    entityContainer := container.NewEntityContainer()
    repositoryContainer := container.NewRepositoryContainer(entityContainer)
    serviceContainer := container.NewServiceContainer(repositoryContainer)
    controllerContainer := container.NewControllerContainer(serviceContainer)
    middlewareContainer := container.NewMiddlewareContainer(serviceContainer)
    listenerContainer := container.NewListenerContainer(serviceContainer)
    schedulerContainer := container.NewSchedulerContainer(serviceContainer)

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

    if err := engine.Run(); err != nil {
        os.Exit(1)
    }
 }
SERVEREOF

# 7. 运行应用
 go run ./cmd/server
 ```

### 添加第一个接口

```go
// 1. 创建实体 (internal/entities/user.go) - 使用基类
type User struct {
    common.BaseEntityWithTimestamps  // 自动生成 ID、CreatedAt、UpdatedAt
    Name string `gorm:"type:varchar(100);not null" json:"name"`
}

func (u *User) EntityName() string { return "User" }
func (u *User) TableName() string { return "users" }
func (u *User) GetId() string     { return u.ID }

var _ common.IBaseEntity = (*User)(nil)

// 2. 创建仓储 (internal/repositories/user_repository.go)
type IUserRepository interface {
    common.IBaseRepository
    Create(user *User) error
    GetByID(id string) (*User, error)  // ID 类型为 string
    List() ([]*User, error)
}

type UserRepository struct {
    Config    configmgr.IConfigManager     `inject:""`
    DBManager databasemgr.IDatabaseManager `inject:""`
}

func (r *UserRepository) Create(user *User) error {
    return r.DBManager.DB().Create(user).Error  // Hook 自动填充 ID、CreatedAt、UpdatedAt
}

func (r *UserRepository) GetByID(id string) (*User, error) {
    var user User
    err := r.DBManager.DB().Where("id = ?", id).First(&user).Error  // 使用 Where 查询
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) List() ([]*User, error) {
    var users []*User
    err := r.DBManager.DB().Find(&users).Error
    return users, err
}

// 3. 创建服务 (internal/services/user_service.go)
type IUserService interface {
    common.IBaseService
    CreateUser(name string) (*User, error)
    GetUser(id string) (*User, error)
    List() ([]*User, error)
}

type UserService struct {
    Config     configmgr.IConfigManager    `inject:""`
    Repository IUserRepository             `inject:""`
    LoggerMgr  loggermgr.ILoggerManager   `inject:""`
}

func (s *UserService) CreateUser(name string) (*User, error) {
    user := &User{
        Name: name,
        // ID、CreatedAt、UpdatedAt 由 Hook 自动填充，无需手动设置
    }
    
    if err := s.Repository.Create(user); err != nil {
        return nil, err
    }
    
    s.LoggerMgr.Ins().Info("用户创建成功", "id", user.ID, "name", name)
    return user, nil
}

func (s *UserService) GetUser(id string) (*User, error) {
    return s.Repository.GetByID(id)
}

func (s *UserService) List() ([]*User, error) {
    return s.Repository.List()
}

// 4. 创建控制器 (internal/controllers/user_controller.go)
type IUserController interface {
    common.IBaseController
}

type UserController struct {
    Service  IUserService             `inject:""`
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func (c *UserController) GetRouter() string {
    return "/api/users [POST],/api/users/:id [GET]"
}

func (c *UserController) Handle(ctx *gin.Context) {
    method := ctx.Request.Method
    
    if method == "POST" {
        var req struct {
            Name string `json:"name" binding:"required"`
        }
        
        if err := ctx.ShouldBindJSON(&req); err != nil {
            ctx.JSON(400, gin.H{"error": err.Error()})
            return
        }
        
        user, err := c.Service.CreateUser(req.Name)
        if err != nil {
            ctx.JSON(500, gin.H{"error": err.Error()})
            return
        }
        
        ctx.JSON(200, gin.H{"data": user})
    } else if method == "GET" {
        id := ctx.Param("id")  // ID 类型为 string，直接使用
        
        user, err := c.Service.GetUser(id)
        if err != nil {
            ctx.JSON(404, gin.H{"error": "用户不存在"})
            return
        }
        
        ctx.JSON(200, gin.H{"data": user})
    }
}

// 5. 重新生成容器代码（添加新组件后必须重新生成）
go run ./cmd/generate

# 6. 启动应用
go run ./cmd/server

访问 `http://localhost:8080/api/users` 查看接口

**实体基类特性**：
- **CUID2 ID**：25 位字符串，时间有序、高唯一性、分布式安全
- **自动填充**：ID、CreatedAt、UpdatedAt 通过 GORM Hook 自动设置
- **数据库存储**：varchar(32)，预留更多兼容空间
- **类型安全**：ID 类型为 string，Repository/Service/Controller 无需类型转换

 ## 架构设计

 ### 5 层分层架构

 ```
 ┌───────────────────────────────────────────────────────────────────┐
 │  Controller / Middleware / Listener / Scheduler      (交互层)    │
 │  - Controller: 处理 HTTP 请求和响应                              │
 │  - Middleware: 请求预处理和拦截                                  │
 │  - Listener: 处理 MQ 消息队列                                    │
 │  - Scheduler: 执行定时任务                                        │
 └───────────────────────────────────────────────────────────────────┘
                                    ↓ 依赖
 ┌───────────────────────────────────────────────────────────────────┐
 │  Service                                            (服务层)    │
 │  - 编排业务逻辑                                                  │
 │  - 事务管理                                                      │
 │  - 调用 Repository 和其他 Service                                │
 └───────────────────────────────────────────────────────────────────┘
                                    ↓ 依赖
 ┌───────────────────────────────────────────────────────────────────┐
 │  Repository                                          (仓储层)   │
 │  - 数据访问抽象                                                  │
 │  - 与 Manager 和 Entity 交互                                    │
 └───────────────────────────────────────────────────────────────────┘
                ↓ 依赖                              ↑ 使用
 ┌───────────────────────────────┐    ┌─────────────────────────────┐
 │  Manager        (内置管理器层)  │    │  Entity          (实体层)   │
 │  - Config/Logger/Database     │    │  - 数据模型定义              │
 │  - Cache/Telemetry/Lock       │    │  - 表映射和验证规则          │
 │  - Limiter/MQ/Scheduler       │    │  - 无依赖                    │
 └───────────────────────────────┘    └─────────────────────────────┘
 ```

### 依赖注入

使用 `inject:""` 标签声明依赖，Manager 组件由引擎自动注入：

```go
type UserServiceImpl struct {
    // 内置组件（由引擎自动注入）
    Config     configmgr.IConfigManager      `inject:""`
    DBManager  databasemgr.IDatabaseManager  `inject:""`
    CacheMgr   cachemgr.ICacheManager        `inject:""`
    LoggerMgr  loggermgr.ILoggerManager     `inject:""`

    // 业务依赖
    UserRepo   IUserRepository               `inject:""`

    // 同层依赖
    OrderSvc   IOrderService                `inject:""`
}
```

## 内置组件

### 配置管理 (configmgr)

支持 YAML/JSON 配置文件，由引擎自动初始化：

支持 YAML/JSON 配置文件，由引擎自动初始化：

```yaml
server:
  port: 8080
  mode: "debug"

database:
  driver: "sqlite"
  sqlite_config:
    dsn: "./data/myapp.db"

cache:
  driver: "memory"
  memory_config:
    max_size: 100
    max_age: "720h"

logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "info"
      format: "gin"
      color: true
```

### 数据库管理 (databasemgr)

基于 GORM 的多数据库支持：

- MySQL
- PostgreSQL
- SQLite
- None（空实现，用于测试）

```go
type MyRepository struct {
    DBManager databasemgr.IDatabaseManager `inject:""`
}

func (r *MyRepository) FindUser(id uint) (*User, error) {
    var user User
    err := r.DBManager.DB().First(&user, id).Error
    return &user, err
}
```

### 缓存管理 (cachemgr)

统一缓存接口，支持多种驱动：

- Redis（分布式缓存）
- Memory（基于 Ristretto 的高性能内存缓存）
- None（空实现，用于测试）

```go
type MyService struct {
    CacheMgr cachemgr.ICacheManager `inject:""`
}

func (s *MyService) GetData(ctx context.Context, key string) (string, error) {
    var val string
    err := s.CacheMgr.Get(ctx, key, &val)
    if err != nil {
        return "", err
    }
    return val, nil
}
```

### 日志管理 (loggermgr)

基于 Zap 的结构化日志，支持 Gin/JSON/Default 三种格式：

```yaml
logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "info"      # debug, info, warn, error, fatal
      format: "gin"       # gin | json | default
      color: true
      time_format: "2006-01-02 15:04:05.000"
```

**Gin 格式输出示例**：
```
2026-01-24 15:04:05.123 | INFO  | 开始依赖注入 | count=23
2026-01-24 15:04:05.456 | WARN  | 慢查询检测 | duration=1.2s
2026-01-24 15:04:05.789 | ERROR | 数据库连接失败 | error="connection refused"
```

**Gin 格式特点**：
- 统一格式：`{时间} | {级别} | {消息} | {字段1}={值1} {字段2}={值2} ...`
- 时间固定宽度 23 字符：`2006-01-02 15:04:05.000`
- 级别固定宽度 5 字符，右对齐，带颜色
- 字段格式：`key=value`，字符串值用引号包裹

**请求日志示例**：
```
2026-01-24 15:04:05.123 | 200   | 1.234ms | 127.0.0.1 | GET | /api/messages
```

### 遥测管理 (telemetrymgr)

OpenTelemetry 集成，支持分布式追踪和指标采集：

```go
type MyService struct {
    TelemetryMgr telemetrymgr.ITelemetryManager `inject:""`
}

func (s *MyService) DoSomething(ctx context.Context) error {
    ctx, span := s.TelemetryMgr.Tracer("MyService").Start(ctx, "operation-name")
    defer span.End()
    span.SetAttributes(attribute.String("key", "value"))
    return nil
}
```

### 锁管理 (lockmgr)

支持 Redis 和 Memory 两种驱动的分布式锁：

```go
type MyService struct {
    LockMgr lockmgr.ILockManager `inject:""`
}

func (s *MyService) ProcessResource(ctx context.Context, resourceID string) error {
    err := s.LockMgr.Lock(ctx, "resource:"+resourceID, 10*time.Second)
    if err != nil {
        return err
    }
    defer s.LockMgr.Unlock(ctx, "resource:"+resourceID)
    // 执行需要加锁的操作
    return nil
}
```

### 限流管理 (limitermgr)

支持 Redis、Memory 和 None 三种驱动的限流器：

```go
type MyService struct {
    LimiterMgr limitermgr.ILimiterManager `inject:""`
}

func (s *MyService) CheckUserLimit(ctx context.Context, userID string) (bool, error) {
    allowed, err := s.LimiterMgr.Allow(ctx, "user:"+userID, 100, time.Minute)
    return allowed, err
}
```

### 消息队列管理 (mqmgr)

支持 RabbitMQ 和 Memory 两种驱动的消息队列：

```go
type MyService struct {
    MQMgr mqmgr.IMQManager `inject:""`
}

func (s *MyService) SendMessage(ctx context.Context, queue string, data []byte) error {
    return s.MQMgr.Publish(ctx, queue, data)
}

func (s *MyService) ConsumeMessages(ctx context.Context, queue string) error {
    return s.MQMgr.SubscribeWithCallback(ctx, queue, func(ctx context.Context, msg mqmgr.Message) error {
        fmt.Printf("Received: %s\n", string(msg.Body()))
        return s.MQMgr.Ack(ctx, msg)
    })
}
```

### 定时任务管理 (schedulermgr)

基于 Cron 的定时任务调度，支持 Cron 表达式：

```go
type MyScheduler struct {
    SchedulerMgr schedulermgr.ISchedulerManager `inject:""`
}

func (s *MyScheduler) RunTask(ctx context.Context) error {
    return s.SchedulerMgr.Schedule(ctx, "task-name", "0 0 * * *", func(tickID int64) error {
        // 定时任务逻辑
        return nil
    })
}
```

## 内置中间件

框架提供以下内置中间件，开箱即用：

 | 中间件 | 默认 Order | 说明 |
|--------|-----------|------|
| RecoveryMiddleware | 0 | panic 恢复，防止服务崩溃 |
| RequestLoggerMiddleware | 50 | 记录每个请求的详细信息 |
| CORSMiddleware | 100 | 处理跨域请求 |
| SecurityHeadersMiddleware | 150 | 添加安全相关的 HTTP 头 |
| RateLimiterMiddleware | 200 | 基于 IP、Header 或用户 ID 的限流 |
| TelemetryMiddleware | 250 | 集成 OpenTelemetry 追踪 |

```go
import "github.com/lite-lake/litecore-go/component/litemiddleware"

// 限流中间件（使用默认配置）
rateLimiter := litemiddleware.NewRateLimiterMiddlewareWithDefaults()
middlewareContainer.RegisterMiddleware(rateLimiter)

// 或自定义限流规则
limit := 200
window := time.Minute
keyPrefix := "api"
rateLimiter := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
    Limit:     &limit,
    Window:    &window,
    KeyPrefix: &keyPrefix,
})
middlewareContainer.RegisterMiddleware(rateLimiter)
```

## CLI 工具

自动生成容器初始化代码，简化项目搭建：

```bash
# 构建工具
go build -o litecore-cli ./cli

# 使用默认配置生成
./litecore-cli generate

# 自定义参数
./litecore-cli generate --project . --output internal/application --package application --config configs/config.yaml

# 查看帮助
./litecore-cli --help
./litecore-cli generate --help
```

或在业务项目中使用：

```go
// cmd/generate/main.go
package main

import "github.com/lite-lake/litecore-go/cli/generator"

func main() {
    generator.MustRun(generator.DefaultConfig())
}
```

运行：`go run ./cmd/generate`

 生成的文件位于 `internal/application/`：
  - `entity_container.go`
  - `repository_container.go`
  - `service_container.go`
  - `controller_container.go`
  - `middleware_container.go`
  - `listener_container.go`
  - `scheduler_container.go`
  - `init_container.go`
  - `engine.go`

## 目录结构规范

 ```
 myapp/
 ├── cmd/
 │   ├── server/            # 应用入口
 │   └── generate/          # 代码生成器入口
 ├── internal/
 │   ├── application/       # 容器初始化代码（自动生成）
 │   ├── entities/          # 实体层
 │   ├── repositories/      # 仓储层
 │   ├── services/          # 服务层
 │   ├── controllers/       # 控制器层
 │   ├── middlewares/       # 中间件层
 │   ├── listeners/         # 监听器层
 │   ├── schedulers/        # 定时器层
 │   └── dtos/              # 数据传输对象
 ├── configs/               # 配置文件
 │   └── config.yaml
 ├── data/                  # 数据目录
 ├── go.mod
 └── go.sum
 ```

## 命名规范

| 类型 | 命名规则 | 示例 |
|------|----------|------|
| 接口 | `I` 前缀 | `IUserService` |
| 实现结构体 | PascalCase | `UserService` 或 `UserServiceImpl` |
| 导出函数 | PascalCase | `GetUser(id)` |
| 私有函数 | camelCase | `validateUser(user)` |
| 工厂函数 | `New` 前缀 | `NewUserService()` |

## 实用工具

### JWT 工具

```go
import "github.com/lite-lake/litecore-go/util/jwt"

// 生成令牌
claims := &jwt.StandardClaims{
    Issuer:    "myapp",
    Subject:   "user123",
    ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
}
token, err := jwt.GenerateHS256Token(claims, "your-secret-key")

// 解析令牌
parsedClaims, err := jwt.ParseHS256Token(token, "your-secret-key")
```

### Hash 工具

```go
import "github.com/lite-lake/litecore-go/util/hash"

// 计算哈希值
md5Hash := hash.MD5String("hello")
sha256Hash := hash.SHA256String("hello")

// Bcrypt 密码哈希
hashedPassword, err := hash.BcryptHash("password")
isValid := hash.BcryptVerify("password", hashedPassword)

// HMAC 计算
hmacHash := hash.HMACSHA256String("data", "secret-key")
```

### ID 生成器

```go
import "github.com/lite-lite-litecore-go/util/id"

// 生成 CUID2 格式的唯一标识符（25 位字符串）
uniqueID := id.NewCUID2()
// 输出: 2k4d2j3h8f9g3n7p6q5r4s3t (示例)

// 在实体基类中自动使用
// 无需手动调用，基类通过 GORM Hook 自动生成
```

**特性**：
- 时间有序：前缀包含时间戳
- 高唯一性：结合时间戳和加密级随机数
- 分布式安全：无需中央协调
- 可读性：仅包含小写字母和数字

**与实体基类集成**：
- 基类使用 `id.NewCUID2()` 自动生成 ID
- 数据库存储为 varchar(32)，预留更多兼容空间
- 通过 GORM Hook 自动填充，Service 层无需手动设置

## 示例项目

 查看 `samples/messageboard/` 目录，获取完整的使用示例：

```bash
cd samples/messageboard

# 首次使用需要生成管理员密码
go run cmd/genpasswd/main.go

# 将生成的加密密码复制到 configs/config.yaml 的 app.admin.password 字段

# 运行应用
go run cmd/server/main.go
```

访问地址：
- 用户首页: http://localhost:8080/
- 管理页面: http://localhost:8080/admin.html

示例包含：
- 完整的 5 层架构实现
- 实体基类使用（CUID2 ID + 时间戳自动填充）
- 9 个内置 Manager 组件自动初始化
- GORM 与 Ristretto 缓存集成
- 内置中间件（CORS、RateLimiter、Telemetry）
- 自定义路由和中间件

详细文档请参考 [samples/messageboard/README.md](samples/messageboard/README.md)

**实体基类应用**：
示例中的 `Message` 实体使用 `common.BaseEntityWithTimestamps` 基类，自动获得以下特性：
- **CUID2 ID**：25 位字符串 ID，通过 GORM Hook 自动生成
- **时间戳自动填充**：CreatedAt 和 UpdatedAt 自动设置
- **类型安全**：Repository/Service/Controller 层 ID 类型统一为 string
- **数据库存储**：varchar(32)，预留更多兼容空间

## 测试

```bash
go test ./...                           # 运行所有测试
go test -cover ./...                    # 带覆盖率
go test ./util/jwt                      # 指定包
go test -bench=. ./util/hash            # 性能测试
```

## 代码规范

```bash
go fmt ./...                            # 格式化代码
go vet ./...                            # 检查问题
go mod tidy                             # 整理依赖
```

### 导入顺序

```go
import (
    "crypto"                            // 标准库
    "errors"
    "time"

    "github.com/gin-gonic/gin"         // 第三方库
    "github.com/stretchr/testify/assert"

    "github.com/lite-lake/litecore-go/common"  // 本地模块
)
```

## 文档

- [AGENTS.md](AGENTS.md) - AI 编码助手指南（面向开发者）
- [CLI README](cli/README.md) - CLI 工具文档
- [Server README](server/README.md) - 服务引擎文档
- [Component README](component/README.md) - 内置组件文档
- [Samples README](samples/messageboard/README.md) - 示例项目文档

**AGENTS.md 与本 README 的区别**：
- 本 README 面向框架用户，提供快速入门、功能特性和使用指南
- AGENTS.md 面向 AI 编码助手和开发者，提供代码规范、架构细节和贡献指南

## 许可证

BSD 2-Clause License
