# Common - 公共基础接口

定义七层架构的基础接口，规范各层的行为契约和生命周期管理，并提供类型转换工具函数。

## 职责和作用

common 包是框架的核心基础模块，主要职责包括：

1. **架构接口定义** - 定义七层架构的标准接口（Entity、Manager、Repository、Service、Controller、Middleware、ConfigMgr）
2. **生命周期管理** - 提供统一的 OnStart 和 OnStop 钩子方法，管理组件启动和停止
3. **类型安全** - 为依赖注入容器提供标准接口类型，支持类型安全的依赖注入
4. **命名规范** - 每层接口要求实现对应的名称方法，便于调试、日志和监控
5. **行为契约** - 通过接口定义各层的核心行为，确保系统分层架构的一致性
6. **工具函数** - 提供类型转换工具函数，支持安全的类型断言和默认值处理
7. **常量定义** - 定义 HTTP 状态码等常用常量

## 特性

- **七层架构基础接口** - 定义 Entity、Manager、Repository、Service、Controller、Middleware、ConfigMgr 的标准接口
- **生命周期管理** - 提供统一的 OnStart 和 OnStop 钩子方法，管理组件启动和停止
- **命名规范** - 每层接口要求实现对应的名称方法，便于调试、日志和监控
- **行为契约** - 通过接口定义各层的核心行为，确保系统分层架构的一致性
- **依赖注入支持** - 为依赖注入容器提供标准接口类型，支持类型安全的依赖注入
- **类型转换工具** - 提供安全的类型转换函数，避免 panic 并支持默认值
- **HTTP 状态码常量** - 定义完整的 HTTP 状态码常量，便于统一使用

## 快速开始

```go
package main

import "github.com/lite-lake/litecore-go/common"

// 定义实体，实现 IBaseEntity 接口
type User struct {
	ID   string `gorm:"primaryKey"`
	Name string
	Age  int
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
	// 注入依赖的存储库
	repository common.IBaseRepository
}

func (s *UserService) ServiceName() string {
	return "UserService"
}

func (s *UserService) OnStart() error {
	// 服务启动时初始化资源
	return nil
}

func (s *UserService) OnStop() error {
	// 服务停止时清理资源
	return nil
}

// 定义控制器，实现 IBaseController 接口
type UserController struct {
	// 注入依赖的服务
	service common.IBaseService
}

func (c *UserController) ControllerName() string {
	return "UserController"
}

func (c *UserController) GetRouter() string {
	return "/users [GET]"
}

func (c *UserController) Handle(ctx *gin.Context) {
	// 处理用户列表请求
	ctx.JSON(200, gin.H{"message": "user list"})
}
```

## 核心功能

### IBaseEntity - 实体基类

定义数据实体的标准接口，所有实体必须实现以下方法：

```go
type User struct {
    ID   string `gorm:"primaryKey"`
    Name string
}

// 返回实体类名，用于标识和调试
func (u *User) EntityName() string {
    return "User"
}

// 返回数据库表名
func (u *User) TableName() string {
    return "users"
}

// 返回实体唯一标识
func (u *User) GetId() string {
    return u.ID
}
```

### IBaseManager - 管理器基类

定义资源管理器的标准接口，提供健康检查和生命周期管理：

```go
type DatabaseManager struct{}

func (m *DatabaseManager) ManagerName() string {
    return "DatabaseManager"
}

func (m *DatabaseManager) Health() error {
    // 检查数据库连接健康状态
    return nil
}

func (m *DatabaseManager) OnStart() error {
    // 初始化数据库连接
    return nil
}

func (m *DatabaseManager) OnStop() error {
    // 关闭数据库连接
    return nil
}
```

### IBaseRepository - 存储库基类

定义数据访问层的标准接口，提供数据持久化和生命周期管理：

```go
type UserRepository struct {
    db *gorm.DB
}

func (r *UserRepository) RepositoryName() string {
    return "UserRepository"
}

func (r *UserRepository) OnStart() error {
    // 初始化存储资源
    return nil
}

func (r *UserRepository) OnStop() error {
    // 清理存储资源
    return nil
}
```

### IBaseService - 服务基类

定义业务逻辑层的标准接口，提供服务实现和生命周期管理：

```go
type UserService struct {
    repository common.IBaseRepository
}

func (s *UserService) ServiceName() string {
    return "UserService"
}

func (s *UserService) OnStart() error {
    // 加载缓存、连接外部服务等
    return nil
}

func (s *UserService) OnStop() error {
    // 刷新缓存、关闭连接等
    return nil
}
```

### IBaseController - 控制器基类

定义 HTTP 处理层的标准接口，提供路由定义和请求处理：

```go
type UserController struct {
    service *UserService
}

func (c *UserController) ControllerName() string {
    return "UserController"
}

// 返回路由定义，格式同 OpenAPI @Router 规范
func (c *UserController) GetRouter() string {
    return "/users [GET]"
}

func (c *UserController) Handle(ctx *gin.Context) {
    // 处理请求逻辑
    ctx.JSON(200, gin.H{"data": "users"})
}
```

### IBaseMiddleware - 中间件基类

定义中间件的标准接口，提供请求拦截和生命周期管理：

```go
type AuthMiddleware struct {
    config configmgr.IConfigManager
}

func (m *AuthMiddleware) MiddlewareName() string {
    return "AuthMiddleware"
}

// 定义中间件执行顺序
func (m *AuthMiddleware) Order() int {
    return 100
}

// 返回 Gin 中间件函数
func (m *AuthMiddleware) Wrapper() gin.HandlerFunc {
    return func(ctx *gin.Context) {
        // 验证身份
        ctx.Next()
    }
}

func (m *AuthMiddleware) OnStart() error {
    return nil
}

func (m *AuthMiddleware) OnStop() error {
    return nil
}
```

## 公共工具和函数

### 类型转换工具函数

提供安全的类型转换函数，用于从 `any` 类型中获取特定类型的值，支持错误处理和默认值：

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
import "github.com/lite-lake/litecore-go/common"

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

// 带默认值的 map 获取
defaultSettings := map[string]any{"enabled": true}
settings := common.GetMapOrDefault(config["settings"], defaultSettings)
```

### HTTP 状态码常量

定义完整的 HTTP 状态码常量，便于统一使用：

```go
const (
    HTTPStatusContinue           = 100
    HTTPStatusOK                   = 200
    HTTPStatusCreated              = 201
    HTTPStatusAccepted             = 202
    HTTPStatusNoContent            = 204
    HTTPStatusMovedPermanently  = 301
    HTTPStatusFound             = 302
    HTTPStatusBadRequest                  = 400
    HTTPStatusUnauthorized                = 401
    HTTPStatusForbidden                   = 403
    HTTPStatusNotFound                    = 404
    HTTPStatusInternalServerError           = 500
    HTTPStatusServiceUnavailable            = 503
    // ... 更多状态码
)
```

使用示例：

```go
import "github.com/lite-lake/litecore-go/common"

ctx.JSON(common.HTTPStatusOK, gin.H{"message": "success"})
ctx.JSON(common.HTTPStatusNotFound, gin.H{"error": "not found"})
ctx.JSON(common.HTTPStatusInternalServerError, gin.H{"error": "internal error"})
```

## API

### 实体层

```go
type IBaseEntity interface {
    EntityName() string
    TableName() string
    GetId() string
}
```

### 管理器层

```go
type IBaseManager interface {
    ManagerName() string
    Health() error
    OnStart() error
    OnStop() error
}
```

### 存储库层

```go
type IBaseRepository interface {
    RepositoryName() string
    OnStart() error
    OnStop() error
}
```

### 服务层

```go
type IBaseService interface {
    ServiceName() string
    OnStart() error
    OnStop() error
}
```

### 控制器层

```go
type IBaseController interface {
    ControllerName() string
    GetRouter() string
    Handle(ctx *gin.Context)
}
```

### 中间件层

```go
type IBaseMiddleware interface {
    MiddlewareName() string
    Order() int
    Wrapper() gin.HandlerFunc
    OnStart() error
    OnStop() error
}
```

## 架构层次

各层之间有明确的依赖关系，从低到高依次为：

```
ConfigMgr (配置管理器层)
      ↓
Entity (实体层)
      ↓
Manager (管理器层)
      ↓
Repository (存储库层)
      ↓
Service (服务层)
      ↓
Controller (控制器层) / Middleware (中间件层)
```

- 上层可以依赖下层
- 下层不能依赖上层
- 同层之间可以相互依赖（Service 层支持同层依赖）
- ConfigMgr 和 Manager 作为服务器内置组件，由引擎自动初始化和注入
- Manager 组件位于 `manager/` 目录下，包括：configmgr、databasemgr、loggermgr、cachemgr、lockmgr、limitermgr、mqmgr、telemetrymgr

## 生命周期管理

实现了 IBaseManager、IBaseRepository、IBaseService、IBaseMiddleware 接口的组件，会在以下时机调用生命周期方法：

1. **OnStart** - 服务器启动时调用，用于初始化资源
2. **OnStop** - 服务器停止时调用，用于清理资源

生命周期方法返回 error，如果初始化失败会阻止服务器启动。

## 依赖注入

所有基础接口都支持依赖注入，推荐使用 `inject:""` 标签：

```go
import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/configmgr"
)

type UserServiceImpl struct {
    // 内置组件（由引擎自动注入）
    Config    configmgr.IConfigManager `inject:""`
    DBManager common.IBaseManager      `inject:""`

    // 业务依赖
    UserRepo  common.IBaseRepository   `inject:""`
}
```

## 最佳实践

1. **接口实现** - 确保所有组件实现对应的基础接口（以 `I` 开头）
2. **命名规范** - 使用结构体类型名作为名称方法返回值
3. **生命周期** - 在 OnStart 中初始化资源，在 OnStop 中清理资源
4. **依赖关系** - 严格遵循分层架构的依赖规则
5. **错误处理** - 生命周期方法中的错误应该被正确处理和传播
6. **类型转换** - 使用 common 包提供的类型转换工具函数，避免直接类型断言导致的 panic
7. **HTTP 状态码** - 使用 common 包定义的 HTTP 状态码常量，保持代码一致性

## 与其他包的关系

- **manager/** - Manager 组件实现 IBaseManager 接口，包括 configmgr、databasemgr、loggermgr、cachemgr、lockmgr、limitermgr、mqmgr、telemetrymgr
- **component/** - 业务组件实现 IBaseEntity、IBaseRepository、IBaseService、IBaseController、IBaseMiddleware 接口
- **container/** - 依赖注入容器使用 common 包定义的接口类型进行类型安全的依赖注入
- **util/** - 提供 util 包无法覆盖的通用类型转换工具函数
