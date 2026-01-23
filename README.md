# LiteCore-Go

基于 Gin + GORM + Zap 的 5 层分层架构企业级 Go Web 开发框架

## 特性

- **5 层分层架构** - Entity → Repository → Service → Controller/Middleware，清晰的责任分离
- **内置组件** - Config 和 Manager 作为服务器内置组件，自动初始化和注入
- **依赖注入容器** - 自动注入依赖，支持同层依赖、可选依赖、循环依赖检测
- **统一配置管理** - 支持 YAML/JSON 配置文件，类型安全的配置读取
- **多数据库支持** - 基于 GORM，支持 MySQL、PostgreSQL、SQLite
- **内置中间件** - 日志、CORS、安全头、认证、Recovery 等开箱即用
- **可观测性** - 集成 OpenTelemetry、结构化日志、健康检查、指标采集
- **CLI 代码生成** - 自动生成容器初始化代码，简化项目搭建
- **生命周期管理** - 统一的启动/停止机制，优雅关闭
- **限流器** - 支持 Redis 和内存两种驱动的分布式限流

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
mkdir -p internal/{entities,repositories,services,controllers,middlewares,infras,dtos}
mkdir -p cmd/{server,generate}
mkdir -p {configs,templates,static}

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
    app "com.litelake.myapp/internal/application"
    loggermgr "github.com/lite-lake/litecore-go/component/manager/loggermgr"
)

func main() {
    loggerMgr := loggermgr.GetLoggerManager()
    logger := loggerMgr.Logger("main")
    
    engine, err := app.NewEngine()
    if err != nil {
        logger.Fatal("Failed to create engine", "error", err)
    }

    if err := engine.Run(); err != nil {
        logger.Fatal("Engine run failed", "error", err)
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
    Manager databasemgr.IDatabaseManager `inject:""`
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
│  - 封装数据查询逻辑                                    │
└─────────────────────────────────────────────────────┘
             ↓ 依赖              ↑ 使用
┌─────────────────────────┐    ┌──────────────────────┐
│  Manager  (内置组件)     │    │  Entity    (实体层)   │
│  - server/builtin/      │    │  - 数据模型定义        │
│    manager/             │    │  - 表映射和验证规则    │
│  - configmgr            │    │  - 无依赖              │
│  - loggermgr            │    │                      │
│  - databasemgr          │    │                      │
│  - cachemgr             │    │                      │
│  - telemetrymgr         │    │                      │
│  - lockmgr              │    │                      │
│  - limitermgr           │    │                      │
│  - 由引擎自动初始化       │    │                      │
└─────────────────────────┘    └──────────────────────┘
```

### 依赖规则

- **向下依赖**：上层只能依赖下层
- **单向依赖**：Entity → Manager → Repository → Service → Controller/Middleware
- **同层依赖**：Service 支持同层依赖，通过拓扑排序解决循环依赖
- **内置组件**：Manager 作为服务器内置组件（位于 server/builtin/manager/），由引擎自动初始化和注入

### 依赖注入

使用 `inject:""` 标签声明依赖，Manager 由引擎自动注入：

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

    // 可选依赖（不存在时不会报错）
    OptionalService IOtherService         `inject:"optional"`
}
```

## 核心组件

### 1. 配置管理 (configmgr)

支持 YAML/JSON 配置文件，内置组件自动加载配置：

```go
// 配置由引擎自动初始化，通过依赖注入使用
type MyService struct {
    Config configmgr.IConfigManager `inject:""`
}

func (s *MyService) OnStart() error {
    // 获取配置值
    port := s.Config.GetInt("server.port", 8080)
    mode := s.Config.GetString("server.mode", "debug")

    // 获取结构化配置
    var dbConfig databasemgr.MySQLConfig
    if err := s.Config.Unmarshal("database.mysql_config", &dbConfig); err != nil {
        return err
    }
    return nil
}
```

### 2. 数据库管理 (databasemgr)

基于 GORM 的多数据库支持：

```go
// 由引擎自动初始化，通过依赖注入使用
type MyRepository struct {
    Manager databasemgr.IDatabaseManager `inject:""`
}

func (r *MyRepository) OnStart() error {
    // 自动迁移
    return r.Manager.AutoMigrate(&User{})
}

func (r *MyRepository) FindUser(id uint) (*User, error) {
    // 数据库操作
    var user User
    err := r.Manager.DB().First(&user, id).Error
    return &user, err
}

// 健康检查
err := r.Manager.Health()

// 连接池统计
stats := r.Manager.Stats()
```

// 自动迁移
dbMgr.AutoMigrate(&User{})

// 数据库操作
var user User
dbMgr.DB().First(&user, 1)

// 健康检查
err := dbMgr.Health()

// 连接池统计
stats := dbMgr.Stats()
```

支持的数据库：
- MySQL
- PostgreSQL
- SQLite
- None (空实现，用于测试)

### 3. 缓存管理 (cachemgr)

统一缓存接口，由引擎自动初始化，通过依赖注入使用：

```go
// 由引擎自动初始化，通过依赖注入使用
type MyService struct {
    CacheMgr cachemgr.ICacheManager `inject:""`
}

func (s *MyService) GetData(ctx context.Context, key string) (string, error) {
    // 获取值
    val, err := s.CacheMgr.Get(ctx, key)
    return val, err
}

func (s *MyService) SetData(ctx context.Context, key, value string) error {
    // 设置值
    return s.CacheMgr.Set(ctx, key, value, time.Hour)
}

func (s *MyService) DeleteData(ctx context.Context, key string) error {
    // 删除值
    return s.CacheMgr.Delete(ctx, key)
}
```

### 4. 日志管理 (loggermgr)

基于 Zap 的结构化日志，由引擎自动初始化，通过依赖注入使用：

```go
// 由引擎自动初始化，通过依赖注入使用
type MyService struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func (s *MyService) DoSomething() {
    logger := s.LoggerMgr.Ins()
    // 记录日志
    logger.Info("user login", "user_id", "123", "ip", "127.0.0.1")
    logger.Error("database error", "error", err, "query", sql)
}
```

### 5. 遥测管理 (telemetrymgr)

OpenTelemetry 集成，由引擎自动初始化，通过依赖注入使用：

```go
// 由引擎自动初始化，通过依赖注入使用
type MyService struct {
    TelemetryMgr telemetrymgr.ITelemetryManager `inject:""`
}

func (s *MyService) DoSomething(ctx context.Context) error {
    // 创建 span
    ctx, span := s.TelemetryMgr.Tracer("MyService").Start(ctx, "operation-name")
    defer span.End()

    // 记录属性
    span.SetAttributes(attribute.String("key", "value"))

    // 记录事件
    span.AddEvent("event-name")

    return nil
}
```

### 6. 锁管理 (lockmgr)

支持 Redis 和 Memory 两种驱动的分布式锁，由引擎自动初始化，通过依赖注入使用：

```go
// 由引擎自动初始化，通过依赖注入使用
type MyService struct {
    LockMgr lockmgr.ILockManager `inject:""`
}

func (s *MyService) ProcessResource(ctx context.Context, resourceID string) error {
    // 获取锁（阻塞直到成功或超时）
    err := s.LockMgr.Lock(ctx, "resource:"+resourceID, 10*time.Second)
    if err != nil {
        return err
    }
    defer s.LockMgr.Unlock(ctx, "resource:"+resourceID)

    // 执行需要加锁的操作
    // ...

    return nil
}

func (s *MyService) TryProcessResource(ctx context.Context, resourceID string) error {
    // 尝试获取锁（非阻塞）
    locked, err := s.LockMgr.TryLock(ctx, "resource:"+resourceID, 10*time.Second)
    if err != nil {
        return err
    }
    if !locked {
        return fmt.Errorf("资源已被占用")
    }
    defer s.LockMgr.Unlock(ctx, "resource:"+resourceID)

    // 执行需要加锁的操作
    // ...

    return nil
}
```

使用场景：
- 分布式环境下的资源互斥访问
- 并发控制，防止重复操作
- 任务队列的任务消费
- 限流控制

支持的驱动：
- Redis（分布式锁）
- Memory（本地内存锁）
- None（空实现，用于测试）

### 7. 限流管理 (limitermgr)

支持 Redis、Memory 和 None 三种驱动的限流器，由引擎自动初始化，通过依赖注入使用：

```go
// 由引擎自动初始化，通过依赖注入使用
type MyService struct {
    LimiterMgr limitermgr.ILimiterManager `inject:""`
}

func (s *MyService) CheckUserLimit(ctx context.Context, userID string) (bool, error) {
    // 检查是否允许通过（100次/分钟）
    allowed, err := s.LimiterMgr.Allow(ctx, "user:"+userID, 100, time.Minute)
    if err != nil {
        return false, err
    }
    return allowed, nil
}

func (s *MyService) GetUserRemaining(ctx context.Context, userID string) (int, error) {
    // 获取剩余可访问次数
    remaining, err := s.LimiterMgr.GetRemaining(ctx, "user:"+userID, 100, time.Minute)
    if err != nil {
        return 0, err
    }
    return remaining, nil
}
```

使用场景：
- API 接口请求频率限制
- 用户行为频率控制（如点赞、评论）
- 防止恶意请求和爬虫
- 资源使用配额管理

支持的驱动：
- Redis（分布式限流）
- Memory（本地内存限流）
- None（无限流，用于测试）

### 8. HTTP 服务引擎 (server)

统一的服务启动和生命周期管理：

```go
// 一键启动
engine, err := app.NewEngine()
engine.Run()

// 分步启动
engine.Initialize()
engine.Start()
engine.WaitForShutdown()
```

## 内置中间件

### 1. Recovery 中间件

自动恢复 panic，防止服务崩溃：

```go
import "github.com/lite-lake/litecore-go/component/litemiddleware"

// 在中间件容器中注册
recovery := litemiddleware.NewRecoveryMiddleware()
middlewareContainer.Register(recovery)
```

### 2. 日志中间件

记录每个请求的详细信息：

```go
// 在中间件容器中注册
requestLogger := middleware.NewRequestLoggerMiddleware()
middlewareContainer.Register(requestLogger)
```

### 3. CORS 中间件

处理跨域请求：

```go
cors := middleware.NewCorsMiddleware(&middleware.CorsConfig{
    AllowOrigins:     []string{"*"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Origin", "Content-Type"},
    ExposeHeaders:    []string{"Content-Length"},
    AllowCredentials: false,
    MaxAge:           12 * time.Hour,
})
middlewareContainer.Register(cors)
```

### 4. 安全头中间件

添加安全相关的 HTTP 头：

```go
securityHeaders := middleware.NewSecurityHeadersMiddleware()
middlewareContainer.Register(securityHeaders)
```

### 5. 遥测中间件

集成 OpenTelemetry 追踪：

```go
telemetry := middleware.NewTelemetryMiddleware()
middlewareContainer.Register(telemetry)
```

### 6. 限流中间件

基于 IP、路径、Header 或用户 ID 的限流：

```go
// 基于 IP 限流（100次/分钟）
rateLimiter := middleware.NewRateLimiterByIP(100, time.Minute)
middlewareContainer.Register(rateLimiter)

// 基于路径限流（1000次/分钟）
rateLimiter := middleware.NewRateLimiterByPath(1000, time.Minute)
middlewareContainer.Register(rateLimiter)

// 基于 Header 限流（500次/分钟）
rateLimiter := middleware.NewRateLimiterByHeader(500, time.Minute, "X-User-ID")
middlewareContainer.Register(rateLimiter)

// 基于用户 ID 限流（200次/分钟）
rateLimiter := middleware.NewRateLimiterByUserID(200, time.Minute)
middlewareContainer.Register(rateLimiter)

// 自定义配置
rateLimiter := middleware.NewRateLimiter(&middleware.RateLimiterConfig{
    Limit:     100,
    Window:    time.Minute,
    KeyPrefix: "custom",
    KeyFunc: func(c *gin.Context) string {
        // 自定义 key 生成逻辑
        return c.GetHeader("X-API-Key")
    },
    SkipFunc: func(c *gin.Context) bool {
        // 跳过限流的条件
        return c.Request.URL.Path == "/health"
    },
})
middlewareContainer.Register(rateLimiter)
```

### 中间件排序

可以通过实现 `Order()` 方法控制中间件执行顺序：

```go
const (
    OrderRecovery   = 10  // panic 恢复
    OrderLogger     = 20  // 日志记录
    OrderCors       = 30  // 跨域处理
    OrderTelemetry  = 40  // 遥测监控
    OrderSecurity   = 50  // 安全头
    OrderRateLimit  = 90  // 限流
    OrderAuth       = 100 // 认证
)
```

## CLI 工具

### 安装

```bash
go build -o litecore-generate ./cli
```

### 使用

```bash
# 在项目中生成容器代码
./litecore-generate

# 或指定参数
./litecore-generate -project . -output internal/application -package application -configmgr configs/config.yaml
```

### 在业务项目中使用

创建 `cmd/generate/main.go`：

```go
package main

import "github.com/lite-lake/litecore-go/cli/generator"

func main() {
    generator.MustRun(generator.DefaultConfig())
}
```

运行：

```bash
go run ./cmd/generate
```

## 实用工具

### 1. JWT 工具

```go
import "github.com/lite-lake/litecore-go/util/jwt"

// 生成 Token
token, err := jwt.GenerateHS256Token(jwt.StandardClaims{
    UserID:   "123",
    Username: "admin",
    ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
}, "secret")

// 验证 Token
claims, err := jwt.VerifyHS256Token(token, "secret")
```

### 2. Hash 工具

```go
import "github.com/lite-lake/litecore-go/util/hash"

// MD5
md5 := hash.MD5("hello")

// SHA256
sha256 := hash.SHA256("hello")

// bcrypt（密码哈希）
hashed, err := hash.BcryptHash("password", hash.DefaultBcryptCost)
err = hash.BcryptVerify("password", hashed)
```

### 3. 验证器

```go
import "github.com/lite-lake/litecore-go/util/validator"

type User struct {
    Name  string `validate:"required,min=3,max=50"`
    Email string `validate:"required,email"`
    Age   int    `validate:"gte=0,lte=130"`
}

v := validator.New()
err := v.Struct(&User{Name: "abc", Email: "test@example.com", Age: 25})
```

### 4. ID 生成器

```go
import "github.com/lite-lake/litecore-go/util/id"

// UUID
uuid := id.UUID()

// Snowflake ID
snowflake := id.Snowflake()

// Nano ID
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
- 内置组件自动初始化
- 用户认证和会话管理
- 留言审核流程
- 数据库迁移
- 中间件集成
- 前端界面

## 目录结构规范

```
myapp/
├── cmd/
│   ├── server/            # 应用入口
│   │   └── main.go
│   └── generate/          # 代码生成器入口
│       └── main.go
├── internal/
 │   ├── application/       # 容器初始化代码（自动生成）
 │   │   ├── entity_container.go
 │   │   ├── repository_container.go
 │   │   ├── service_container.go
 │   │   ├── controller_container.go
 │   │   ├── middleware_container.go
 │   │   └── engine.go
│   ├── entities/          # 实体层
│   ├── repositories/      # 仓储层
│   ├── services/          # 服务层
│   ├── controllers/       # 控制器层
│   ├── middlewares/       # 中间件层
│   ├── dtos/              # 数据传输对象
│   └── infras/            # 基础设施（Manager 实现）
├── configs/               # 配置文件
│   └── config.yaml
├── templates/             # HTML 模板
├── static/                # 静态资源
├── data/                  # 数据目录
├── go.mod
└── go.sum
```

## 命名规范

### 接口命名

使用 `I` 前缀：

```go
type IUserService interface {
    common.IBaseService
    Get(id string) (*User, error)
}
```

### 实现命名

使用 `Impl` 后缀或无后缀：

```go
type UserServiceImpl struct {
    common.BaseService
    Repo IUserRepository `inject:""`
}

// 或
type UserService struct {
    common.BaseService
    Repo IUserRepository `inject:""`
}
```

### 导出函数

使用 PascalCase：

```go
func (s *UserService) GetUser(id string) (*User, error) {
    return s.Repo.Get(id)
}
```

### 私有函数

使用 camelCase：

```go
func (s *UserService) validateUser(user *User) error {
    return nil
}
```

## 测试

```bash
# 运行所有测试
go test ./...

# 带覆盖率
go test -cover ./...

# 运行指定包测试
go test ./util/jwt

# 运行指定测试
go test ./util/jwt -run TestGenerateHS256Token

# 性能测试
go test -bench=. ./util/hash

# 详细输出
go test -v ./util/jwt
```

## 代码规范

### 导入顺序

```go
import (
    "crypto"       // 标准库
    "errors"
    "time"

    "github.com/gin-gonic/gin"  // 第三方库
    "github.com/stretchr/testify/assert"

    "github.com/lite-lake/litecore-go/common"  // 本地模块
)
```

### 错误处理

```go
if err != nil {
    return "", fmt.Errorf("operation failed: %w", err)
}
```

### 格式化

```bash
go fmt ./...
go vet ./...
```

## 最佳实践

### 1. 中间件排序

```go
const (
    OrderRecovery   = 10  // panic 恢复
    OrderLogger     = 20  // 日志记录
    OrderCors       = 30  // 跨域处理
    OrderTelemetry  = 40  // 遥测监控
    OrderSecurity   = 50  // 安全头
    OrderRateLimit  = 90  // 限流
    OrderAuth       = 100 // 认证
)
```

### 2. 路由定义

使用 OpenAPI 风格：`/path [METHOD]`

```go
func (ctrl *UserController) GetRouter() string {
    return "/api/users [GET]"
}
```

### 3. 配置管理

使用统一的配置管理器：

```go
type MyService struct {
    Config configmgr.IConfigManager `inject:""`
}

func (s *MyService) OnStart() error {
    port := s.Config.GetInt("server.port", 8080)
    return nil
}
```

### 4. 事务处理

在 Service 层处理事务：

```go
func (s *UserService) CreateUser(user *User) error {
    return s.DBMgr.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(user).Error; err != nil {
            return err
        }
        // 其他数据库操作
        return nil
    })
}
```

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交变更
4. 推送到分支
5. 创建 Pull Request

## 许可证

BSD 2-Clause License

## 文档

- [AGENTS.md](AGENTS.md) - AI 编码助手指南
- [CLI README](cli/README.md) - CLI 工具文档
- [Server README](server/README.md) - 服务引擎文档
- [Samples](samples/messageboard/README.md) - 示例项目文档

## 联系方式

- GitHub: https://github.com/litelake/litecore-go
