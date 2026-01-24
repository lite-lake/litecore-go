# LiteCore-Go

基于 Gin + GORM + Zap 的企业级 Go Web 开发框架，采用 5 层分层架构和声明式依赖注入。

## 核心特性

- **5 层分层架构** - Entity → Repository → Service → Controller/Middleware，清晰的责任分离
- **内置 Manager 组件** - Config、Logger、Database、Cache、Telemetry、Lock、Limiter、MQ 等管理器，自动初始化和注入
- **声明式依赖注入** - 使用 `inject:""` 标签自动注入依赖，支持同层依赖和循环依赖检测
- **统一配置管理** - 支持 YAML/JSON 配置文件，类型安全的配置读取
- **多驱动支持** - 数据库（MySQL/PostgreSQL/SQLite）、缓存（Redis/Memory）、限流（Redis/Memory）、锁（Redis/Memory）、MQ（RabbitMQ/Memory）
- **内置中间件** - Recovery、日志、CORS、安全头、限流、遥测等开箱即用
- **可观测性** - 集成 OpenTelemetry、结构化日志、健康检查、指标采集
- **Gin 风格日志** - 支持可配置日志格式（gin/json/default），彩色输出，结构化记录
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
mkdir -p internal/{entities,repositories,services,controllers,middlewares}
mkdir -p cmd/{server,generate}
mkdir -p {configs,data}

# 3. 创建配置文件
cat > configs/config.yaml <<EOF
server:
  port: 8080
  mode: "debug"

database:
  driver: "sqlite"
  sqlite_config:
    dsn: "./data/myapp.db"

logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "info"
      format: "gin"
      color: true
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

# 5. 运行生成器
go run ./cmd/generate

# 6. 创建应用入口
cat > cmd/server/main.go <<'SERVEREOF'
package main

import (
    "os"

    "github.com/lite-lake/litecore-go/server"
    builtin "github.com/lite-lake/litecore-go/server/builtin"
    "github.com/lite-lake/litecore-go/container"
)

func main() {
    entityContainer := container.NewEntityContainer()
    repositoryContainer := container.NewRepositoryContainer(entityContainer)
    serviceContainer := container.NewServiceContainer(repositoryContainer)
    controllerContainer := container.NewControllerContainer(serviceContainer)
    middlewareContainer := container.NewMiddlewareContainer(serviceContainer)

    engine := server.NewEngine(
        &builtin.Config{
            Driver:   "yaml",
            FilePath: "configs/config.yaml",
        },
        entityContainer,
        repositoryContainer,
        serviceContainer,
        controllerContainer,
        middlewareContainer,
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
// 1. 创建实体 (internal/entities/user.go)
type User struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `json:"name"`
}

func (u *User) EntityName() string { return "User" }
func (u *User) TableName() string  { return "users" }
func (u *User) GetId() string     { return fmt.Sprintf("%d", u.ID) }

// 2. 创建仓储 (internal/repositories/user_repository.go)
type IUserRepository interface {
    common.IBaseRepository
    List() ([]*User, error)
}

type UserRepository struct {
    Config  configmgr.IConfigManager    `inject:""`
    DBMgr   databasemgr.IDatabaseManager `inject:""`
}

func (r *UserRepository) List() ([]*User, error) {
    var users []*User
    err := r.DBMgr.DB().Find(&users).Error
    return users, err
}

// 3. 创建服务 (internal/services/user_service.go)
type IUserService interface {
    common.IBaseService
    List() ([]*User, error)
}

type UserService struct {
    Config configmgr.IConfigManager `inject:""`
    Repo   IUserRepository          `inject:""`
}

func (s *UserService) List() ([]*User, error) {
    return s.Repo.List()
}

// 4. 创建控制器 (internal/controllers/user_controller.go)
type UserController struct {
    Config configmgr.IConfigManager `inject:""`
    Svc    IUserService             `inject:""`
}

func (ctrl *UserController) GetRouter() string {
    return "/api/users [GET]"
}

func (ctrl *UserController) Handle(c *gin.Context) {
    users, err := ctrl.Svc.List()
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"data": users})
}

// 5. 重新生成容器代码
go run ./cmd/generate

# 6. 启动应用
go run ./cmd/server
```

访问 `http://localhost:8080/api/users` 查看接口

## 架构设计

### 5 层分层架构

```
┌─────────────────────────────────────────────────────┐
│  Controller / Middleware         (控制器/中间件层)   │
│  - 处理 HTTP 请求和响应                               │
│  - 参数验证和转换                                    │
│  - 调用 Service 层业务逻辑                           │
└─────────────────────────────────────────────────────┘
                             ↓ 依赖
┌─────────────────────────────────────────────────────┐
│  Service                           (服务层)         │
│  - 编排业务逻辑                                          │
│  - 事务管理                                             │
│  - 调用 Repository 和其他 Service                      │
└─────────────────────────────────────────────────────┘
                             ↓ 依赖
┌─────────────────────────────────────────────────────┐
│  Repository                       (仓储层)           │
│  - 数据访问抽象                                        │
│  - 与 Manager 和 Entity 交互                          │
└─────────────────────────────────────────────────────┘
               ↓ 依赖              ↑ 使用
┌─────────────────────────┐    ┌──────────────────────┐
│  Manager  (管理器层)     │    │  Entity    (实体层)   │
│  - Config/Logger/DB     │    │  - 数据模型定义        │
│  - Cache/Telemetry      │    │  - 表映射和验证规则    │
│  - Lock/Limiter/MQ      │    │  - 无依赖              │
└─────────────────────────┘    └──────────────────────┘
```

### 依赖注入

使用 `inject:""` 标签声明依赖，Manager 组件由引擎自动注入：

```go
type UserServiceImpl struct {
    // 内置组件（由引擎自动注入）
    Config    configmgr.IConfigManager    `inject:""`
    DBMgr     databasemgr.IDatabaseManager `inject:""`
    CacheMgr  cachemgr.ICacheManager      `inject:""`

    // 业务依赖
    UserRepo  IUserRepository             `inject:""`

    // 同层依赖
    OrderSvc  IOrderService               `inject:""`
}
```

## 内置组件

### 配置管理 (configmgr)

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
    DBMgr databasemgr.IDatabaseManager `inject:""`
}

func (r *MyRepository) FindUser(id uint) (*User, error) {
    var user User
    err := r.DBMgr.DB().First(&user, id).Error
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
    return val, err
}
```

### 日志管理 (loggermgr)

基于 Zap 的结构化日志，支持 Gin/JSON/Default 三种格式：

```yaml
logger:
  driver: "zap"
  zap_config:
    console_config:
      level: "info"      # debug, info, warn, error, fatal
      format: "gin"       # gin | json | default
      color: true
```

**Gin 格式输出示例**：
```
2026-01-24 15:04:05.123 | INFO  | 开始依赖注入 | count=23
2026-01-24 15:04:05.456 | WARN  | 慢查询检测 | duration=1.2s
2026-01-24 15:04:05.789 | ERROR | 数据库连接失败 | error="connection refused"
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

## 内置中间件

框架提供以下内置中间件，开箱即用：

| 中间件 | 说明 |
|--------|------|
| RecoveryMiddleware | panic 恢复，防止服务崩溃 |
| RequestLoggerMiddleware | 记录每个请求的详细信息 |
| CORSMiddleware | 处理跨域请求 |
| SecurityHeadersMiddleware | 添加安全相关的 HTTP 头 |
| TelemetryMiddleware | 集成 OpenTelemetry 追踪 |
| RateLimiterMiddleware | 基于 IP、Header 或用户 ID 的限流 |

```go
import "github.com/lite-lake/litecore-go/component/litemiddleware"

// 限流中间件（100次/分钟）
rateLimiter := litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
    Limit:  100,
    Window: time.Minute,
    KeyPrefix: &[]string{"ip"}[0],
})
middlewareContainer.Register(rateLimiter)
```

## CLI 工具

自动生成容器初始化代码，简化项目搭建：

```bash
# 构建工具
go build -o litecore-generate ./cli

# 使用默认配置生成
./litecore-generate

# 自定义参数
./litecore-generate -project . -output internal/application -package application -config configs/config.yaml
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

token, err := jwt.GenerateHS256Token(jwt.StandardClaims{
    UserID:    "123",
    Username:  "admin",
    ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
}, "secret")

claims, err := jwt.VerifyHS256Token(token, "secret")
```

### Hash 工具

```go
import "github.com/lite-lake/litecore-go/util/hash"

md5 := hash.MD5("hello")
sha256 := hash.SHA256("hello")
hashed, err := hash.BcryptHash("password", hash.DefaultBcryptCost)
err = hash.BcryptVerify("password", hashed)
```

### ID 生成器

```go
import "github.com/lite-lake/litecore-go/util/id"

uuid := id.UUID()
snowflake := id.Snowflake()
nanoID := id.NanoID()
```

## 示例项目

查看 `samples/messageboard` 目录，获取完整的使用示例：

```bash
cd samples/messageboard
go run ./cmd/server
```

示例包含：
- 完整的 5 层架构实现
- 内置 Manager 组件自动初始化
- 用户认证和会话管理
- 留言审核流程
- 数据库迁移
- 中间件集成
- Gin 风格日志输出
- 前端界面

详细文档请参考 [samples/messageboard/README.md](samples/messageboard/README.md)

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
- [Samples](samples/messageboard/README.md) - 示例项目文档

**AGENTS.md 与本 README 的区别**：
- 本 README 面向框架用户，提供快速入门、功能特性和使用指南
- AGENTS.md 面向 AI 编码助手和开发者，提供代码规范、架构细节和贡献指南

## 许可证

BSD 2-Clause License
