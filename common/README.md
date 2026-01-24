# Common - 公共基础接口

定义 5 层依赖注入架构的基础接口，规范各层的行为契约和生命周期管理，并提供类型转换工具函数。

## 概述

common 包是框架的核心基础模块，定义了标准化的分层架构接口，所有业务组件都必须实现对应的基础接口。通过接口约束，确保架构的一致性和可维护性。

## 架构层次

```
┌──────────────────────────────────────────────────────────┐
│                   Controller/Middleware                   │
│              (HTTP 请求处理和请求拦截)                       │
├──────────────────────────────────────────────────────────┤
│                      Service                             │
│                   (业务逻辑层)                             │
├──────────────────────────────────────────────────────────┤
│                    Repository                             │
│                   (数据访问层)                             │
├──────────────────────────────────────────────────────────┤
│                      Entity                               │
│                   (数据实体层)                             │
└──────────────────────────────────────────────────────────┘
            ↑                                              ↑
            └───────────────── Manager Layer ───────────────┘
          (configmgr、databasemgr、loggermgr、cachemgr、
           lockmgr、limitermgr、mqmgr、telemetrymgr)
```

## 特性

- **标准接口定义** - 定义 Entity、Manager、Repository、Service、Controller、Middleware 的标准接口
- **生命周期管理** - 提供统一的 OnStart 和 OnStop 钩子方法
- **命名规范** - 每层接口要求实现对应的名称方法，便于调试和日志
- **依赖注入支持** - 为容器提供标准接口类型，支持类型安全的依赖注入
- **类型转换工具** - 提供安全的类型转换函数，避免 panic 并支持默认值
- **HTTP 状态码常量** - 定义完整的 HTTP 状态码常量，便于统一使用

## 快速开始

```go
import "github.com/lite-lake/litecore-go/common"

// 定义实体，实现 IBaseEntity 接口
type User struct {
    ID   string `gorm:"primaryKey"`
    Name string
}

func (u *User) EntityName() string {
    return "User"
}

func (u *User) TableName() string {
    return "users"
}

func (u *User) GetId() string {
    return u.ID
}

// 定义服务，实现 IBaseService 接口
type UserService struct {
    Config    configmgr.IConfigManager    `inject:""`
    LoggerMgr loggermgr.ILoggerManager   `inject:""`
}

func (s *UserService) ServiceName() string {
    return "UserService"
}

func (s *UserService) OnStart() error {
    return nil
}

func (s *UserService) OnStop() error {
    return nil
}
```

## 核心接口

### IBaseEntity - 实体层接口

定义数据实体的标准接口，所有实体必须实现：

```go
type IBaseEntity interface {
    EntityName() string  // 返回实体名称，用于标识和调试
    TableName() string   // 返回数据库表名
    GetId() string       // 返回实体唯一标识
}
```

### IBaseManager - 管理器层接口

定义资源管理器的标准接口，提供健康检查和生命周期管理：

```go
type IBaseManager interface {
    ManagerName() string  // 返回管理器名称
    Health() error        // 检查健康状态
    OnStart() error       // 启动时触发
    OnStop() error        // 停止时触发
}
```

**注意**：Manager 组件位于 `manager/` 目录下，由引擎自动初始化和注入。常见管理器包括：
- `configmgr` - 配置管理
- `databasemgr` - 数据库管理
- `loggermgr` - 日志管理
- `cachemgr` - 缓存管理
- `lockmgr` - 锁管理
- `limitermgr` - 限流管理
- `mqmgr` - 消息队列管理
- `telemetrymgr` - 可观测性管理

### IBaseRepository - 存储库层接口

定义数据访问层的标准接口：

```go
type IBaseRepository interface {
    RepositoryName() string  // 返回存储库名称
    OnStart() error         // 启动时触发
    OnStop() error          // 停止时触发
}
```

### IBaseService - 服务层接口

定义业务逻辑层的标准接口：

```go
type IBaseService interface {
    ServiceName() string  // 返回服务名称
    OnStart() error      // 启动时触发
    OnStop() error       // 停止时触发
}
```

### IBaseController - 控制器层接口

定义 HTTP 处理层的标准接口：

```go
type IBaseController interface {
    ControllerName() string              // 返回控制器名称
    GetRouter() string                   // 返回路由定义（OpenAPI @Router 规范）
    Handle(ctx *gin.Context)             // 处理请求
}
```

**路由格式**：`/path [METHOD]`，例如 `/api/messages [GET]`、`/api/messages [POST]`

### IBaseMiddleware - 中间件层接口

定义中间件的标准接口：

```go
type IBaseMiddleware interface {
    MiddlewareName() string        // 返回中间件名称
    Order() int                   // 返回执行顺序，数值越小越先执行
    Wrapper() gin.HandlerFunc     // 返回 Gin 中间件函数
    OnStart() error               // 启动时触发
    OnStop() error                // 停止时触发
}
```

## 依赖规则

各层之间有明确的依赖关系：

- **Entity 层**：无依赖
- **Repository 层**：可依赖 Entity、Manager
- **Service 层**：可依赖 Repository、Entity、Manager、其他 Service
- **Controller 层**：可依赖 Service、Manager
- **Middleware 层**：可依赖 Service、Manager

**原则**：上层可以依赖下层，下层不能依赖上层。

## 依赖注入

所有基础接口都支持依赖注入，使用 `inject:""` 标签：

```go
type UserServiceImpl struct {
    // 内置管理器（由引擎自动注入）
    Config     configmgr.IConfigManager    `inject:""`
    LoggerMgr  loggermgr.ILoggerManager   `inject:""`
    DBManager  databasemgr.IDatabaseManager `inject:""`

    // 业务依赖
    UserRepo   IUserRepository            `inject:""`
    CacheMgr   cachemgr.ICacheManager     `inject:""`
}
```

## 生命周期管理

实现 IBaseManager、IBaseRepository、IBaseService、IBaseMiddleware 接口的组件，会在以下时机调用生命周期方法：

1. **OnStart** - 服务器启动时调用，用于初始化资源（如连接数据库、加载缓存等）
2. **OnStop** - 服务器停止时调用，用于清理资源（如关闭连接、刷新缓存等）

生命周期方法返回 error，如果初始化失败会阻止服务器启动。

## 类型转换工具

提供安全的类型转换函数，用于从 `any` 类型中获取特定类型的值：

```go
// GetString 从 any 类型中安全获取字符串值
func GetString(value any) (string, error)

// GetStringOrDefault 从 any 类型中安全获取字符串值，失败时返回默认值
func GetStringOrDefault(value any, defaultValue string) string

// GetMap 从 any 类型中安全获取 map[string]any 值
func GetMap(value any) (map[string]any, error)

// GetMapOrDefault 从 any 类型中安全获取 map[string]any 值，失败时返回默认值
func GetMapOrDefault(value any, defaultValue map[string]any) map[string]any
```

使用示例：

```go
// 从配置中获取字符串值
name, err := common.GetString(config["name"])
if err != nil {
    log.Error("无效的名称配置")
}

// 带默认值的字符串获取
timeout := common.GetStringOrDefault(config["timeout"], "30s")

// 从配置中获取 map 值
settings, err := common.GetMap(config["settings"])
if err != nil {
    log.Error("无效的设置配置")
}
```

## HTTP 状态码常量

定义完整的 HTTP 状态码常量，便于统一使用：

```go
const (
    HTTPStatusContinue                    = 100
    HTTPStatusOK                          = 200
    HTTPStatusCreated                     = 201
    HTTPStatusNoContent                   = 204
    HTTPStatusMovedPermanently            = 301
    HTTPStatusBadRequest                  = 400
    HTTPStatusUnauthorized                = 401
    HTTPStatusForbidden                   = 403
    HTTPStatusNotFound                    = 404
    HTTPStatusInternalServerError         = 500
    HTTPStatusServiceUnavailable          = 503
    // ... 更多状态码
)
```

使用示例：

```go
ctx.JSON(common.HTTPStatusOK, gin.H{"message": "success"})
ctx.JSON(common.HTTPStatusNotFound, gin.H{"error": "not found"})
ctx.JSON(common.HTTPStatusInternalServerError, gin.H{"error": "internal error"})
```

## 最佳实践

1. **接口实现** - 确保所有组件实现对应的基础接口（以 `I` 开头）
2. **命名规范** - 使用结构体类型名作为名称方法返回值
3. **生命周期** - 在 OnStart 中初始化资源，在 OnStop 中清理资源
4. **依赖关系** - 严格遵循分层架构的依赖规则
5. **错误处理** - 生命周期方法中的错误应该被正确处理和传播
6. **类型转换** - 使用 common 包提供的类型转换工具函数，避免直接类型断言导致的 panic
7. **HTTP 状态码** - 使用 common 包定义的 HTTP 状态码常量，保持代码一致性

## 实际应用示例

参考 `samples/messageboard` 目录下的完整示例：

```go
// entities/message_entity.go
type Message struct {
    ID        uint      `gorm:"primarykey"`
    Nickname  string    `gorm:"type:varchar(20);not null"`
    Content   string    `gorm:"type:varchar(500);not null"`
    Status    string    `gorm:"type:varchar(20);default:'pending'"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

func (m *Message) EntityName() string { return "Message" }
func (m *Message) TableName() string { return "messages" }
func (m *Message) GetId() string { return fmt.Sprintf("%d", m.ID) }

// repositories/message_repository.go
type messageRepository struct {
    Config  configmgr.IConfigManager    `inject:""`
    Manager databasemgr.IDatabaseManager `inject:""`
}

func (r *messageRepository) RepositoryName() string { return "MessageRepository" }
func (r *messageRepository) OnStart() error { return r.Manager.AutoMigrate(&Message{}) }
func (r *messageRepository) OnStop() error { return nil }

// services/message_service.go
type messageService struct {
    Config     configmgr.IConfigManager     `inject:""`
    Repository IMessageRepository           `inject:""`
    LoggerMgr  loggermgr.ILoggerManager    `inject:""`
}

func (s *messageService) ServiceName() string { return "MessageService" }
func (s *messageService) OnStart() error { return nil }
func (s *messageService) OnStop() error { return nil }

// controllers/msg_create_controller.go
type msgCreateControllerImpl struct {
    MessageService IMessageService         `inject:""`
    LoggerMgr      loggermgr.ILoggerManager `inject:""`
}

func (c *msgCreateControllerImpl) ControllerName() string { return "msgCreateControllerImpl" }
func (c *msgCreateControllerImpl) GetRouter() string { return "/api/messages [POST]" }
func (c *msgCreateControllerImpl) Handle(ctx *gin.Context) { /* ... */ }

// middlewares/auth_middleware.go
type authMiddleware struct {
    AuthService IAuthService `inject:""`
}

func (m *authMiddleware) MiddlewareName() string { return "AuthMiddleware" }
func (m *authMiddleware) Order() int { return 100 }
func (m *authMiddleware) Wrapper() gin.HandlerFunc { /* ... */ }
func (m *authMiddleware) OnStart() error { return nil }
func (m *authMiddleware) OnStop() error { return nil }
```

## 与其他包的关系

- **manager/** - Manager 组件实现 IBaseManager 接口，作为基础设施层提供各种能力
- **container/** - 依赖注入容器使用 common 包定义的接口类型进行类型安全的依赖注入
- **component/** - 业务组件实现 IBaseEntity、IBaseRepository、IBaseService、IBaseController、IBaseMiddleware 接口
- **util/** - 工具函数包提供特定功能的工具函数，common 包提供通用的类型转换函数
