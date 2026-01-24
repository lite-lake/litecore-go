# CLI 工具使用示例

## 概述

`litecore-cli` 是 litecore 框架提供的代码生成器命令行工具，用于自动生成依赖注入容器代码。

工具会自动扫描项目中的以下目录并生成对应的容器代码：
- `internal/entities` - 实体层，生成 `entity_container.go`
- `internal/repositories` - 仓储层，生成 `repository_container.go`
- `internal/services` - 服务层，生成 `service_container.go`
- `internal/controllers` - 控制器层，生成 `controller_container.go`
- `internal/middlewares` - 中间件层，生成 `middleware_container.go`
- `internal/listeners` - 监听器层，生成 `listener_container.go`
- `internal/schedulers` - 定时任务层，生成 `scheduler_container.go`
- `engine.go` - 引擎初始化代码

## 安装

```bash
# 从源码运行
go run ./cli/main.go [参数]

# 编译为可执行文件
go build -o litecore-generate ./cli/main.go

# 安装到 $GOPATH/bin
go install ./cli/main.go
```

## 命令行参数

| 参数 | 简写 | 说明 | 默认值 |
|------|------|------|--------|
| `version` | `v` | 显示版本信息 | - |
| `project` | `p` | 项目路径 | `.` |
| `output` | `o` | 输出目录 | `internal/application` |
| `package` | `pkg` | 包名 | `application` |
| `configmgr` | `c` | 配置文件路径 | `configs/config.yaml` |

## 项目初始化

### 1. 创建项目结构

```bash
mkdir myapp && cd myapp
go mod init github.com/yourname/myapp

# 创建目录结构
mkdir -p internal/entities
mkdir -p internal/repositories
mkdir -p internal/services
mkdir -p internal/controllers
mkdir -p internal/middlewares
mkdir -p internal/listeners
mkdir -p internal/schedulers
mkdir -p internal/application
mkdir -p configs
mkdir -p cmd/generate
mkdir -p cmd/server
```

### 2. 创建配置文件

创建 `configs/config.yaml`：

```yaml
app:
  name: "myapp"
  version: "1.0.0"

server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"

database:
  driver: "sqlite"
  auto_migrate: true
  sqlite_config:
    dsn: "./data/myapp.db"

logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "info"
      format: "gin"

cache:
  driver: "memory"

limiter:
  driver: "memory"

lock:
  driver: "memory"

mq:
  driver: "memory"

telemetry:
  driver: "none"

scheduler:
  driver: "cron"
  cron_config:
    validate_on_startup: true
```

## 基本使用

### 查看版本

```bash
go run ./cli/main.go -version
# 或
go run ./cli/main.go -v
```

### 使用默认配置生成

```bash
# 在当前项目目录执行
go run ./cli/main.go
```

### 自定义输出目录

```bash
go run ./cli/main.go -o internal/app
```

### 自定义包名

```bash
go run ./cli/main.go -pkg myapp
```

### 自定义配置文件路径

```bash
go run ./cli/main.go -c config/dev.yaml
```

### 完整自定义

```bash
go run ./cli/main.go \
  -p /path/to/project \
  -o internal/application \
  -pkg application \
  -c configs/config.yaml
```

## 代码生成规则

### 扫描规则

生成器会扫描 `internal` 目录下的各层代码，识别组件的规则如下：

| 层级 | 目录 | 识别规则 | 工厂函数 |
|------|------|---------|---------|
| 实体 | `entities/` | 导出的 `struct` 类型 | - |
| 仓储 | `repositories/` | 以 `I` 开头的接口 | `NewXxx()` |
| 服务 | `services/` | 以 `I` 开头的接口 | `NewXxx()` |
| 控制器 | `controllers/` | 以 `I` 开头的接口 | `NewXxx()` |
| 中间件 | `middlewares/` | 以 `I` 开头的接口 | `NewXxx()` |
| 监听器 | `listeners/` | 以 `I` 开头的接口 | `NewXxx()` |
| 调度器 | `schedulers/` | 以 `I` 开头的接口 | `NewXxx()` |

### 命名规范

#### 实体
- 类型名：PascalCase，如 `User`、`Message`
- 文件名：`xxx_entity.go`
- 实现接口：`common.IBaseEntity`

#### 仓储
- 接口名：`I` + PascalCase，如 `IUserRepository`
- 实现类型：小写 + `Impl`，如 `userRepositoryImpl`
- 工厂函数：`NewXxx()`，如 `NewUserRepository()`
- 文件名：`xxx_repository.go`
- 实现接口：`common.IBaseRepository`

#### 服务
- 接口名：`I` + PascalCase，如 `IUserService`
- 实现类型：小写 + `ServiceImpl`，如 `userServiceImpl`
- 工厂函数：`NewXxxService()`，如 `NewUserService()`
- 文件名：`xxx_service.go`
- 实现接口：`common.IBaseService`

#### 控制器
- 接口名：`I` + PascalCase，如 `IUserController`
- 实现类型：小写 + `ControllerImpl`，如 `userControllerImpl`
- 工厂函数：`NewXxxController()`，如 `NewUserController()`
- 文件名：`xxx_controller.go`
- 实现接口：`common.IBaseController`

#### 中间件
- 接口名：`I` + PascalCase，如 `IAuthMiddleware`
- 实现类型：小写 + `MiddlewareImpl`，如 `authMiddlewareImpl`
- 工厂函数：`NewXxxMiddleware()`，如 `NewAuthMiddleware()`
- 文件名：`xxx_middleware.go`
- 实现接口：`common.IBaseMiddleware`

#### 监听器
- 接口名：`I` + PascalCase，如 `IMessageListener`
- 实现类型：小写 + `ListenerImpl`，如 `messageListenerImpl`
- 工厂函数：`NewXxxListener()`，如 `NewMessageListener()`
- 文件名：`xxx_listener.go`
- 实现接口：`common.IBaseListener`

#### 调度器
- 接口名：`I` + PascalCase，如 `ICleanupScheduler`
- 实现类型：小写 + `SchedulerImpl`，如 `cleanupSchedulerImpl`
- 工厂函数：`NewXxxScheduler()`，如 `NewCleanupScheduler()`
- 文件名：`xxx_scheduler.go`
- 实现接口：`common.IBaseScheduler`

### 依赖注入标签

在需要注入依赖的字段上使用 `inject:""` 标签：

```go
type userServiceImpl struct {
    Config    configmgr.IConfigManager    `inject:""`
    DBManager databasemgr.IDatabaseManager `inject:""`
    LoggerMgr loggermgr.ILoggerManager   `inject:""`
    Repo      repository.IUserRepository  `inject:""`
}
```

## 组件开发示例

### 1. 实体（Entity）

创建 `internal/entities/user_entity.go`：

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
	Email     string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) EntityName() string {
	return "User"
}

func (User) TableName() string {
	return "users"
}

func (u *User) GetId() string {
	return fmt.Sprintf("%d", u.ID)
}

var _ common.IBaseEntity = (*User)(nil)
```

### 2. 仓储（Repository）

创建 `internal/repositories/user_repository.go`：

```go
package repositories

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/configmgr"
	"github.com/lite-lake/litecore-go/manager/databasemgr"
	"github.com/lite-lake/litecore-go/internal/entities"
)

type IUserRepository interface {
	common.IBaseRepository
	Create(user *entities.User) error
	GetByID(id uint) (*entities.User, error)
	GetByEmail(email string) (*entities.User, error)
}

type userRepositoryImpl struct {
	Config  configmgr.IConfigManager     `inject:""`
	Manager databasemgr.IDatabaseManager `inject:""`
}

func NewUserRepository() IUserRepository {
	return &userRepositoryImpl{}
}

func (r *userRepositoryImpl) RepositoryName() string {
	return "UserRepository"
}

func (r *userRepositoryImpl) OnStart() error {
	return nil
}

func (r *userRepositoryImpl) OnStop() error {
	return nil
}

func (r *userRepositoryImpl) Create(user *entities.User) error {
	db := r.Manager.DB()
	return db.Create(user).Error
}

func (r *userRepositoryImpl) GetByID(id uint) (*entities.User, error) {
	db := r.Manager.DB()
	var user entities.User
	err := db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) GetByEmail(email string) (*entities.User, error) {
	db := r.Manager.DB()
	var user entities.User
	err := db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

var _ IUserRepository = (*userRepositoryImpl)(nil)
```

### 3. 服务（Service）

创建 `internal/services/user_service.go`：

```go
package services

import (
	"errors"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/configmgr"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/internal/entities"
	"github.com/lite-lake/litecore-go/internal/repositories"
)

type IUserService interface {
	common.IBaseService
	CreateUser(name, email string) (*entities.User, error)
	GetUser(id uint) (*entities.User, error)
}

type userServiceImpl struct {
	Config     configmgr.IConfigManager     `inject:""`
	Repository repositories.IUserRepository `inject:""`
	LoggerMgr  loggermgr.ILoggerManager    `inject:""`
}

func NewUserService() IUserService {
	return &userServiceImpl{}
}

func (s *userServiceImpl) ServiceName() string {
	return "UserService"
}

func (s *userServiceImpl) OnStart() error {
	return nil
}

func (s *userServiceImpl) OnStop() error {
	return nil
}

func (s *userServiceImpl) CreateUser(name, email string) (*entities.User, error) {
	if name == "" || email == "" {
		return nil, errors.New("name and email are required")
	}

	user := &entities.User{
		Name:  name,
		Email: email,
	}

	if err := s.Repository.Create(user); err != nil {
		return nil, err
	}

	s.LoggerMgr.Ins().Info("User created", "id", user.ID, "name", name)
	return user, nil
}

func (s *userServiceImpl) GetUser(id uint) (*entities.User, error) {
	return s.Repository.GetByID(id)
}

var _ IUserService = (*userServiceImpl)(nil)
```

### 4. 控制器（Controller）

创建 `internal/controllers/user_controller.go`：

```go
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/internal/entities"
	"github.com/lite-lake/litecore-go/internal/services"
)

type IUserController interface {
	common.IBaseController
}

type userControllerImpl struct {
	UserService services.IUserService `inject:""`
	LoggerMgr   loggermgr.ILoggerManager `inject:""`
}

func NewUserController() IUserController {
	return &userControllerImpl{}
}

func (c *userControllerImpl) ControllerName() string {
	return "UserController"
}

func (c *userControllerImpl) GetRouter() string {
	return "/api/users [POST],/api/users/:id [GET]"
}

func (c *userControllerImpl) Handle(ctx *gin.Context) {
	method := ctx.Request.Method

	if method == "POST" {
		c.handleCreate(ctx)
	} else if method == "GET" {
		c.handleGet(ctx)
	}
}

func (c *userControllerImpl) handleCreate(ctx *gin.Context) {
	var req struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required,email"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.UserService.CreateUser(req.Name, req.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (c *userControllerImpl) handleGet(ctx *gin.Context) {
	id := ctx.Param("id")
	var userId uint
	if _, err := fmt.Sscanf(id, "%d", &userId); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := c.UserService.GetUser(userId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

var _ IUserController = (*userControllerImpl)(nil)
```

### 5. 中间件（Middleware）

创建 `internal/middlewares/logging_middleware.go`：

```go
package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
)

type ILoggingMiddleware interface {
	common.IBaseMiddleware
}

type loggingMiddlewareImpl struct {
	LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func NewLoggingMiddleware() ILoggingMiddleware {
	return &loggingMiddlewareImpl{}
}

func (m *loggingMiddlewareImpl) MiddlewareName() string {
	return "LoggingMiddleware"
}

func (m *loggingMiddlewareImpl) Order() int {
	return 100
}

func (m *loggingMiddlewareImpl) Wrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.LoggerMgr.Ins().Info("Request received",
			"method", c.Request.Method,
			"path", c.Request.URL.Path)
		c.Next()
	}
}

func (m *loggingMiddlewareImpl) OnStart() error {
	return nil
}

func (m *loggingMiddlewareImpl) OnStop() error {
	return nil
}

var _ ILoggingMiddleware = (*loggingMiddlewareImpl)(nil)
```

### 6. 监听器（Listener）

创建 `internal/listeners/email_listener.go`：

```go
package listeners

import (
	"context"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
)

type IEmailListener interface {
	common.IBaseListener
}

type emailListenerImpl struct {
	LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func NewEmailListener() IEmailListener {
	return &emailListenerImpl{}
}

func (l *emailListenerImpl) ListenerName() string {
	return "EmailListener"
}

func (l *emailListenerImpl) GetQueue() string {
	return "email.send"
}

func (l *emailListenerImpl) GetSubscribeOptions() []common.ISubscribeOption {
	return []common.ISubscribeOption{}
}

func (l *emailListenerImpl) OnStart() error {
	l.LoggerMgr.Ins().Info("Email listener started")
	return nil
}

func (l *emailListenerImpl) OnStop() error {
	l.LoggerMgr.Ins().Info("Email listener stopped")
	return nil
}

func (l *emailListenerImpl) Handle(ctx context.Context, msg common.IMessageListener) error {
	l.LoggerMgr.Ins().Info("Email message received",
		"message_id", msg.ID(),
		"body", string(msg.Body()))
	return nil
}

var _ IEmailListener = (*emailListenerImpl)(nil)
var _ common.IBaseListener = (*emailListenerImpl)(nil)
```

### 7. 调度器（Scheduler）

创建 `internal/schedulers/cleanup_scheduler.go`：

```go
package schedulers

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/logger"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/internal/services"
)

type ICleanupScheduler interface {
	common.IBaseScheduler
}

type cleanupSchedulerImpl struct {
	UserService services.IUserService `inject:""`
	LoggerMgr   loggermgr.ILoggerManager `inject:""`
	logger      logger.ILogger
}

func NewCleanupScheduler() ICleanupScheduler {
	return &cleanupSchedulerImpl{}
}

func (s *cleanupSchedulerImpl) SchedulerName() string {
	return "cleanupScheduler"
}

func (s *cleanupSchedulerImpl) GetRule() string {
	return "0 0 3 * * *"  // 每天凌晨3点执行
}

func (s *cleanupSchedulerImpl) GetTimezone() string {
	return "Asia/Shanghai"
}

func (s *cleanupSchedulerImpl) OnTick(tickID int64) error {
	s.initLogger()
	s.logger.Info("Cleanup task started", "tick_id", tickID)
	// 执行清理逻辑
	s.logger.Info("Cleanup task completed", "tick_id", tickID)
	return nil
}

func (s *cleanupSchedulerImpl) OnStart() error {
	s.initLogger()
	s.logger.Info("Cleanup scheduler started")
	return nil
}

func (s *cleanupSchedulerImpl) OnStop() error {
	s.LoggerMgr.Ins().Info("Cleanup scheduler stopped")
	return nil
}

var _ ICleanupScheduler = (*cleanupSchedulerImpl)(nil)
var _ common.IBaseScheduler = (*cleanupSchedulerImpl)(nil)
```

## 项目集成与应用启动

### 项目集成方式

#### 方式一：命令行工具调用

在项目根目录执行：

```bash
# 生成默认配置
litecore-generate

# 或使用完整路径
go run github.com/lite-lake/litecore-go/cli/main.go \
  -o internal/application \
  -pkg application \
  -c configs/config.yaml
```

#### 方式二：嵌入到项目中

创建 `cmd/generate/main.go`:

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

运行：

```bash
go run ./cmd/generate -o internal/app
```

#### 方式三：最简实现

创建 `cmd/generate/main.go`:

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

### 应用启动

创建应用入口 `cmd/server/main.go`：

```go
package main

import (
	"fmt"
	"os"

	myapp "github.com/yourname/myapp/internal/application"
)

func main() {
	engine, err := myapp.NewEngine()
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

### 运行应用

```bash
# 1. 生成容器代码
go run ./cmd/generate

# 2. 启动应用
go run ./cmd/server
```

### 开发流程

1. 创建实体：在 `internal/entities` 目录下定义数据模型
2. 创建仓储：在 `internal/repositories` 目录下定义数据访问层
3. 创建服务：在 `internal/services` 目录下定义业务逻辑层
4. 创建控制器：在 `internal/controllers` 目录下定义 HTTP 处理器
5. 创建中间件：在 `internal/middlewares` 目录下定义 HTTP 中间件
6. 创建监听器：在 `internal/listeners` 目录下定义消息监听器
7. 创建调度器：在 `internal/schedulers` 目录下定义定时任务
8. 生成代码：运行 `go run ./cmd/generate` 生成容器代码
9. 启动应用：运行 `go run ./cmd/server` 启动应用

## 实际项目示例：messageboard

messageboard 示例展示了完整的 CLI 工具使用方式。

### 1. 目录结构

```
samples/messageboard/
├── cmd/
│   ├── generate/          # 代码生成器入口
│   │   └── main.go
│   ├── genpasswd/         # 密码生成工具
│   │   └── main.go
│   └── server/            # 应用启动入口
│       └── main.go
├── configs/
│   └── config.yaml        # 应用配置文件
├── internal/
│   ├── application/       # 生成的容器代码
│   │   ├── entity_container.go
│   │   ├── repository_container.go
│   │   ├── service_container.go
│   │   ├── controller_container.go
│   │   ├── middleware_container.go
│   │   ├── listener_container.go
│   │   ├── scheduler_container.go
│   │   └── engine.go
│   ├── entities/          # 实体层
│   │   └── message_entity.go
│   ├── repositories/      # 仓储层
│   │   └── message_repository.go
│   ├── services/          # 服务层
│   │   ├── message_service.go
│   │   ├── auth_service.go
│   │   └── session_service.go
│   ├── controllers/       # 控制器层
│   │   ├── msg_create_controller.go
│   │   ├── msg_list_controller.go
│   │   ├── msg_delete_controller.go
│   │   ├── msg_status_controller.go
│   │   ├── msg_all_controller.go
│   │   ├── page_home_controller.go
│   │   ├── page_admin_controller.go
│   │   └── admin_auth_controller.go
│   ├── middlewares/       # 中间件层
│   │   ├── auth_middleware.go
│   │   ├── rate_limiter_middleware.go
│   │   ├── cors_middleware.go
│   │   ├── recovery_middleware.go
│   │   ├── request_logger_middleware.go
│   │   ├── security_headers_middleware.go
│   │   └── telemetry_middleware.go
│   ├── listeners/         # 监听器层
│   │   ├── message_created_listener.go
│   │   └── message_audit_listener.go
│   ├── schedulers/        # 调度器层
│   │   ├── cleanup_scheduler.go
│   │   └── statistics_scheduler.go
│   └── dtos/              # 数据传输对象
│       ├── message_dto.go
│       ├── session_dto.go
│       └── response_dto.go
└── go.mod
```

### 2. 生成容器代码

在 `samples/messageboard` 目录下执行：

```bash
go run ./cmd/generate
```

这将：
- 扫描 `internal/entities` 目录生成实体容器
- 扫描 `internal/repositories` 目录生成仓储容器
- 扫描 `internal/services` 目录生成服务容器
- 扫描 `internal/controllers` 目录生成控制器容器
- 扫描 `internal/middlewares` 目录生成中间件容器
- 扫描 `internal/listeners` 目录生成监听器容器
- 扫描 `internal/schedulers` 目录生成调度器容器
- 生成 `engine.go` 引擎初始化代码

### 3. 使用生成的代码

在 `cmd/server/main.go` 中使用：

```go
package main

import (
	"fmt"
	"os"

	messageboardapp "github.com/lite-lake/litecore-go/samples/messageboard/internal/application"
)

func main() {
	// 创建应用引擎
	engine, err := messageboardapp.NewEngine()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create engine: %v\n", err)
		os.Exit(1)
	}

	// 初始化引擎（注册路由、依赖注入等）
	if err := engine.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize engine: %v\n", err)
		os.Exit(1)
	}

	// 启动引擎（启动所有 Manager 和 HTTP 服务器）
	if err := engine.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start engine: %v\n", err)
		os.Exit(1)
	}

	// 等待关闭信号
	engine.WaitForShutdown()
}
```

### 4. 配置文件

生成的容器代码会读取 `configs/config.yaml` 来初始化内置 Manager 组件。主要配置项包括：

- **logger**: 日志配置（zap、控制台、文件）
- **database**: 数据库配置（MySQL、PostgreSQL、SQLite）
- **cache**: 缓存配置（Redis、Memory）
- **limiter**: 限流配置
- **lock**: 分布式锁配置
- **mq**: 消息队列配置（RabbitMQ、Memory）
- **telemetry**: 遥测配置
- **scheduler**: 定时任务配置

## 生成的代码说明

生成器会在指定的输出目录创建以下文件：

- `entity_container.go` - 实体容器初始化
- `repository_container.go` - 仓储容器初始化
- `service_container.go` - 服务容器初始化
- `controller_container.go` - 控制器容器初始化
- `middleware_container.go` - 中间件容器初始化
- `listener_container.go` - 监听器容器初始化
- `scheduler_container.go` - 调度器容器初始化
- `engine.go` - 引擎创建函数

### 实体容器示例

```go
// Code generated by litecore/cli. DO NOT EDIT.
package application

import (
	"github.com/lite-lake/litecore-go/container"
	"github.com/lite-lake/litecore-go/common"
	entities "github.com/yourname/myapp/internal/entities"
)

// InitEntityContainer Initialize entity container
func InitEntityContainer() *container.EntityContainer {
	entityContainer := container.NewEntityContainer()

	container.RegisterEntity[common.IBaseEntity](entityContainer, &entities.User{})

	return entityContainer
}
```

### 仓储容器示例

```go
// Code generated by litecore/cli. DO NOT EDIT.
package application

import (
	"github.com/lite-lake/litecore-go/container"
	repositories "github.com/yourname/myapp/internal/repositories"
)

// InitRepositoryContainer Initialize repository container
func InitRepositoryContainer(entityContainer *container.EntityContainer) *container.RepositoryContainer {
	repositoryContainer := container.NewRepositoryContainer(entityContainer)

	container.RegisterRepository[repositories.IUserRepository](repositoryContainer, repositories.NewUserRepository())

	return repositoryContainer
}
```

### 引擎示例

```go
// Code generated by litecore/cli. DO NOT EDIT.
package application

import (
	"github.com/lite-lake/litecore-go/server"
)

// NewEngine Create application engine
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

### 生成的 Engine 接口

```go
// NewEngine 创建应用引擎
func NewEngine() (*server.Engine, error)
```

返回的 `server.Engine` 提供以下方法：

```go
// Initialize 初始化引擎
func (e *Engine) Initialize() error

// Start 启动引擎
func (e *Engine) Start() error

// Stop 停止引擎
func (e *Engine) Stop() error

// WaitForShutdown 等待关闭信号
func (e *Engine) WaitForShutdown()
```

## API 参考

### Config 配置结构

```go
type Config struct {
    ProjectPath string  // 项目路径（默认: "."）
    OutputDir   string  // 输出目录（默认: "internal/application"）
    PackageName string  // 包名（默认: "application"）
    ConfigPath  string  // 配置文件路径（默认: "configs/config.yaml"）
}
```

### 可用函数

```go
// DefaultConfig 返回默认配置
func DefaultConfig() *Config

// Run 运行生成器
func Run(cfg *Config) error

// MustRun 运行生成器，失败时 panic
func MustRun(cfg *Config)
```

## 注意事项

1. **代码覆盖**：生成的代码会覆盖同名文件，请确保不要手动修改生成的文件
2. **路径规范**：配置文件路径必须是相对于项目根目录的路径
3. **模块依赖**：确保项目已正确初始化 `go mod`，生成器会读取 `go.mod` 文件
4. **目录结构**：确保项目包含标准的分层目录结构（entities, repositories, services, controllers, middlewares, listeners, schedulers）
5. **依赖注入标签**：在需要依赖注入的字段上使用 `inject:""` 标签
6. **接口定义**：接口名称必须以 `I` 开头（除了实体），实现结构体必须为小写（私有）
7. **工厂函数**：必须提供 `NewXxx()` 工厂函数，返回接口类型
8. **接口实现**：所有实现类型必须在文件末尾使用 `var _ InterfaceType = (*ImplType)(nil)` 确保编译时检查

## 常见问题

### Q: 如何生成代码后查看生成的文件？

A: 生成的代码位于 `internal/application` 目录下，可以直接查看和编辑（但不建议手动修改）。

### Q: 工厂函数可以带参数吗？

A: 工厂函数不能带参数。如果需要参数，请使用依赖注入通过 `inject:""` 标签注入。

### Q: 如何处理循环依赖？

A: 如果存在循环依赖，可以：
1. 重构代码，将共同依赖抽取到单独的服务
2. 使用事件监听器（Listener）来解耦

### Q: 控制器如何路由多个路径？

A: 在 `GetRouter()` 方法中返回逗号分隔的路由字符串，如：
```go
func (c *userControllerImpl) GetRouter() string {
    return "/api/users [POST],/api/users/:id [GET]"
}
```

然后在 `Handle()` 方法中根据 `ctx.Request.Method` 处理不同的请求。

### Q: 中间件的 Order 是什么？

A: Order 是中间件的执行顺序，数值越小越先执行。通常：
- Recovery 中间件：Order = 1
- 日志中间件：Order = 10
- 认证中间件：Order = 100
- 业务中间件：Order = 200+

### Q: 如何禁用某个组件的自动注册？

A: 生成器只识别以 `I` 开头的接口和 `New` 开头的工厂函数。如果不想让某个组件被自动注册：
1. 不要让接口以 `I` 开头
2. 或者删除对应的工厂函数

### Q: 如何在服务中使用日志？

A: 在服务结构体中注入 `loggermgr.ILoggerManager`，直接使用：
```go
type userServiceImpl struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}
```

## 进阶用法

### 多环境配置

```bash
# 开发环境
go run ./cmd/generate -c configs/config.dev.yaml

# 生产环境
go run ./cmd/generate -c configs/config.prod.yaml

# 测试环境
go run ./cmd/generate -c configs/config.test.yaml
```

### 自定义输出目录

```bash
# 将生成的代码放到不同目录
go run ./cmd/generate -o internal/container
```

### 与 CI/CD 集成

```yaml
# .github/workflows/build.yml
name: Build

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.25
      - name: Generate code
        run: go run ./cmd/generate
      - name: Build
        run: go build -o app ./cmd/server
      - name: Test
        run: go test ./...
```
