package scaffold

import (
	"bytes"
	"fmt"
	"text/template"
)

const goModTemplate = `module {{.ModulePath}}

go 1.25.0

require (
	github.com/gin-gonic/gin v1.11.0
	github.com/lite-lake/litecore-go ` + LiteCoreGoVersion + `
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.31.1
)
`

const readmeTemplate = `# {{.ProjectName}}

基于 [LiteCore Go](https://github.com/lite-lake/litecore-go) 框架构建的项目。

## 注意事项

由于 LiteCore Go ` + LiteCoreGoVersion + ` 版本可能尚未发布，如果 ` + "`" + `go mod tidy` + "`" + ` 失败，
请在 ` + "`" + `go.mod` + "`" + ` 文件中添加以下 replace 指令：

` + "```go" + `
replace github.com/lite-lake/litecore-go => /path/to/litecore-go
` + "```" + `
  将 ` + "`" + `/path/to/litecore-go` + "`" + ` 替换为实际的本地项目路径。

## 项目结构

` + "```" + `.
├── cmd/
│   ├── generate/        # 代码生成器入口
│   └── server/           # 应用启动入口
├── configs/              # 配置文件
├── internal/
│   ├── application/      # 生成的容器代码
│   ├── entities/         # 实体层
│   ├── repositories/     # 仓储层
│   ├── services/         # 服务层
│   ├── controllers/      # 控制器层
│   ├── middlewares/      # 中间件层
│   ├── listeners/       # 监听器层
│   └── schedulers/      # 调度器层
├── static/               # 静态文件（CSS/JS）
├── templates/            # HTML 模板文件
└── README.md
` + "```" + `

## 快速开始

### 1. 生成容器代码

` + "```bash" + `
go run ./cmd/generate
` + "```" + `

### 2. 启动应用

` + "```bash" + `
go run ./cmd/server
` + "```" + `

### 3. 构建应用

` + "```bash" + `
go build -o bin/{{.ProjectName}} ./cmd/server
` + "```" + `

## 开发指南

### 添加新实体

1. 在 ` + "internal/entities/" + ` 目录创建实体文件
2. 运行 ` + "go run ./cmd/generate" + ` 生成容器代码
3. 在业务逻辑中使用实体

### 添加新仓储

1. 在 ` + "internal/repositories/" + ` 目录创建仓储接口和实现
2. 接口名以 ` + "I" + ` 开头，如 ` + "IUserRepository" + `
3. 工厂函数以 ` + "New" + ` 开头，如 ` + "NewUserRepository()" + `
4. 运行 ` + "go run ./cmd/generate" + ` 生成容器代码

### 添加新服务

1. 在 ` + "internal/services/" + ` 目录创建服务接口和实现
2. 接口名以 ` + "I" + ` 开头，如 ` + "IUserService" + `
3. 工厂函数以 ` + "New" + ` 开头，如 ` + "NewUserService()" + `
4. 在结构体中使用 ` + `inject:""` + ` 标签注入依赖
5. 运行 ` + "go run ./cmd/generate" + ` 生成容器代码

### 添加新控制器

1. 在 ` + "internal/controllers/" + ` 目录创建控制器接口和实现
2. 接口名以 ` + "I" + ` 开头，如 ` + "IUserController" + `
3. 工厂函数以 ` + "New" + ` 开头，如 ` + "NewUserController()" + `
4. 实现 ` + "GetRouter()" + ` 方法定义路由
5. 实现 ` + "Handle()" + ` 方法处理请求
6. 运行 ` + "go run ./cmd/generate" + ` 生成容器代码

## Web 开发

### 静态文件

静态文件位于 ` + "`" + `static/` + "`" + ` 目录，通过 ` + "`" + `/static/*filepath` + "`" + ` 路径访问：
- CSS 文件: ` + "`" + `static/css/` + "`" + `
- JS 文件: ` + "`" + `static/js/` + "`" + `

### HTML 模板

HTML 模板位于 ` + "`" + `templates/` + "`" + ` 目录，使用 Gin 模板引擎渲染。

### 健康检查

应用提供健康检查接口 ` + "`" + `/health` + "`" + `，返回所有 Manager 的状态。

## 更多信息

- [LiteCore Go 文档](https://github.com/lite-lake/litecore-go)
- [CLI 工具文档](../../cli/README.md)
- [示例项目](../../samples/)
`

const configYamlTemplate = `# 应用配置
app:
  name: "{{.ProjectName}}"                       # 应用名称
  version: "1.0.0"                               # 应用版本

# 服务器配置
server:
  host: "0.0.0.0"                               # 监听主机地址
  port: 8080                                    # 监听端口
  mode: "debug"                                 # 运行模式：debug, release, test
  read_timeout: "10s"                            # 读取超时时间
  write_timeout: "10s"                           # 写入超时时间
  idle_timeout: "60s"                           # 空闲超时时间
  enable_recovery: true                         # 是否启用panic恢复
  shutdown_timeout: "30s"                       # 优雅关闭超时时间
  startup_log:                                  # 启动日志配置
    enabled: true                               # 是否启用启动日志
    async: true                                 # 是否异步日志
    buffer: 100                                 # 日志缓冲区大小

# 数据库配置
database:
  driver: "sqlite"                              # 驱动类型：mysql, postgresql, sqlite, none
  auto_migrate: true                            # 是否自动迁移数据库表结构（默认 false）
  # MySQL 配置示例
  # mysql_config:
  #   dsn: "root:password@tcp(localhost:3306)/app?charset=utf8mb4&parseTime=True&loc=Local" # MySQL 数据源名称
  #   pool_config:                             # 连接池配置
  #     max_open_conns: 10                     # 最大打开连接数
  #     max_idle_conns: 5                      # 最大空闲连接数
  #     conn_max_lifetime: "30s"               # 连接最大存活时间
  #     conn_max_idle_time: "5m"               # 连接最大空闲时间
  # PostgreSQL 配置示例
  # postgresql_config:
  #   dsn: "host=localhost port=5432 user=postgres password= dbname=app sslmode=disable" # PostgreSQL 数据源名称
  #   pool_config:                             # 连接池配置
  #     max_open_conns: 10                     # 最大打开连接数
  #     max_idle_conns: 5                      # 最大空闲连接数
  #     conn_max_lifetime: "30s"               # 连接最大存活时间
  #     conn_max_idle_time: "5m"               # 连接最大空闲时间
  sqlite_config:
    dsn: "./data/app.db"                        # SQLite 数据库文件路径
    pool_config:                                # 连接池配置
      max_open_conns: 1                         # 最大打开连接数（SQLite通常设置为1）
      max_idle_conns: 1                         # 最大空闲连接数
      conn_max_lifetime: "30s"                  # 连接最大存活时间
      conn_max_idle_time: "5m"                  # 连接最大空闲时间
  observability_config:                         # 可观测性配置
    slow_query_threshold: "1s"                  # 慢查询阈值
    log_sql: false                              # 是否记录完整SQL（生产环境建议关闭）
    sample_rate: 1.0                            # 采样率（0.0-1.0）

# 缓存配置
cache:
  driver: "memory"                              # 驱动类型：redis, memory, none
  # Redis 配置示例
  # redis_config:
  #   host: "localhost"                         # Redis主机地址
  #   port: 6379                                # Redis端口
  #   password: ""                              # Redis密码
  #   db: 0                                    # Redis数据库编号
  #   max_idle_conns: 10                        # 最大空闲连接数
  #   max_open_conns: 100                       # 最大打开连接数
  #   conn_max_lifetime: "30s"                  # 连接最大存活时间
  memory_config:
    max_size: 100                               # 最大缓存大小（MB）
    max_age: "720h"                             # 最大缓存时间（30天）
    max_backups: 1000                           # 最大备份项数
    compress: false                             # 是否压缩

# 日志配置
logger:
  driver: "zap"                                 # 驱动类型：zap, default, none
  zap_config:
    telemetry_enabled: false                    # 是否启用观测日志
    telemetry_config:                           # 观测日志配置
      level: "info"                             # 日志级别：debug, info, warn, error, fatal
    console_enabled: true                       # 是否启用控制台日志
    console_config:                             # 控制台日志配置
      level: "info"                             # 日志级别：debug, info, warn, error, fatal
      format: "gin"                             # 格式：gin | json | default
      color: true                               # 是否启用颜色
      time_format: "2006-01-02 15:04:05.000"   # 时间格式（默认：2006-01-02 15:04:05.000）
    file_enabled: false                         # 是否启用文件日志
    file_config:                                # 文件日志配置
      level: "info"                             # 日志级别：debug, info, warn, error, fatal
      path: "./logs/app.log"                    # 日志文件路径
      rotation:                                 # 日志轮转配置
        max_size: 100                           # 单个日志文件最大大小（MB）
        max_age: 30                             # 日志文件保留天数
        max_backups: 10                         # 保留的旧日志文件最大数量
        compress: true                          # 是否压缩旧日志文件

# 遥测配置
telemetry:
  driver: "none"                               # 驱动类型：none, otel
  # OTEL 配置示例
  # otel_config:
  #   endpoint: "localhost:4317"               # OTLP端点地址
  #   insecure: false                           # 是否使用不安全连接（默认false，使用TLS）
  #   headers:                                 # 请求头（用于认证）
  #     Authorization: "Bearer your-token"
  #   resource_attributes:                     # 资源属性
  #     - key: "service.name"
  #       value: "app"
  #     - key: "service.version"
  #       value: "1.0.0"
  #   traces:                                  # 链路追踪配置
  #     enabled: false                         # 是否启用链路追踪
  #   metrics:                                 # 指标配置
  #     enabled: false                         # 是否启用指标
  #   logs:                                    # 日志配置
  #     enabled: false                         # 是否启用日志观测

# 限流配置
limiter:
  driver: "memory"                              # 驱动类型：redis, memory
  # Redis 配置示例
  # redis_config:
  #   host: "localhost"                         # Redis主机地址
  #   port: 6379                                # Redis端口
  #   password: ""                              # Redis密码
  #   db: 0                                    # Redis数据库编号
  #   max_idle_conns: 10                        # 最大空闲连接数
  #   max_open_conns: 100                       # 最大打开连接数
  #   conn_max_lifetime: "30s"                  # 连接最大存活时间
  memory_config:
    max_backups: 1000                           # 最大备份项数（清理策略相关）

# 锁配置
lock:
  driver: "memory"                              # 驱动类型：redis, memory
  # Redis 配置示例
  # redis_config:
  #   host: "localhost"                         # Redis主机地址
  #   port: 6379                                # Redis端口
  #   password: ""                              # Redis密码
  #   db: 0                                    # Redis数据库编号
  #   max_idle_conns: 10                        # 最大空闲连接数
  #   max_open_conns: 100                       # 最大打开连接数
  #   conn_max_lifetime: "30s"                  # 连接最大存活时间
  memory_config:
    max_backups: 1000                           # 最大备份项数（清理策略相关）

# 消息队列配置
mq:
  driver: "memory"                              # 驱动类型：rabbitmq, memory
  # RabbitMQ 配置示例
  # rabbitmq_config:
  #   url: "amqp://guest:guest@localhost:5672/" # RabbitMQ连接地址
  #   durable: true                             # 是否持久化队列
  memory_config:
    max_queue_size: 10000                       # 最大队列大小
    channel_buffer: 100                         # 通道缓冲区大小

# 定时任务配置
scheduler:
  driver: "cron"                               # 驱动类型：cron
  cron_config:
    validate_on_startup: true                  # 启动时是否检查所有 Scheduler 配置
`

const gitignoreTemplate = `# Binaries
bin/
*.exe
*.dll
*.so
*.test

# Go workspace
go.work

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Database
data/*.db
data/*.db-shm
data/*.db-wal

# Logs
logs/*.log

# Environment
.env
.env.local

# Coverage
*.out
coverage.html
`

const serverMainTemplate = `package main

import (
	"fmt"
	"os"

	"{{.ModulePath}}/internal/application"
)

func main() {
	engine, err := application.NewEngine()
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建引擎失败: %v\n", err)
		os.Exit(1)
	}

	if err := engine.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "初始化引擎失败: %v\n", err)
		os.Exit(1)
	}

	if err := engine.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "启动引擎失败: %v\n", err)
		os.Exit(1)
	}

	engine.WaitForShutdown()
}
`

const generateMainTemplate = `package main

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
`

const entityTemplate = `package entities

import (
	"fmt"
	"time"

	"github.com/lite-lake/litecore-go/common"
)

type Example struct {
	ID        uint      ` + "`" + `gorm:"primarykey" json:"id"` + "`" + `
	Name      string    ` + "`" + `gorm:"type:varchar(100);not null" json:"name"` + "`" + `
	CreatedAt time.Time ` + "`" + `json:"created_at"` + "`" + `
	UpdatedAt time.Time ` + "`" + `json:"updated_at"` + "`" + `
}

func (e *Example) EntityName() string {
	return "Example"
}

func (Example) TableName() string {
	return "examples"
}

func (e *Example) GetId() string {
	return fmt.Sprintf("%d", e.ID)
}

var _ common.IBaseEntity = (*Example)(nil)
`

const repositoryTemplate = `package repositories

import (
	"{{.ModulePath}}/internal/entities"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/configmgr"
	"github.com/lite-lake/litecore-go/manager/databasemgr"
)

type IExampleRepository interface {
	common.IBaseRepository
	Create(example *entities.Example) error
	GetByID(id uint) (*entities.Example, error)
}

type exampleRepositoryImpl struct {
	Config  configmgr.IConfigManager     ` + "`" + `inject:""` + "`" + `
	Manager databasemgr.IDatabaseManager ` + "`" + `inject:""` + "`" + `
}

func NewExampleRepository() IExampleRepository {
	return &exampleRepositoryImpl{}
}

func (r *exampleRepositoryImpl) RepositoryName() string {
	return "ExampleRepository"
}

func (r *exampleRepositoryImpl) OnStart() error {
	return nil
}

func (r *exampleRepositoryImpl) OnStop() error {
	return nil
}

func (r *exampleRepositoryImpl) Create(example *entities.Example) error {
	db := r.Manager.DB()
	return db.Create(example).Error
}

func (r *exampleRepositoryImpl) GetByID(id uint) (*entities.Example, error) {
	db := r.Manager.DB()
	var example entities.Example
	err := db.First(&example, id).Error
	if err != nil {
		return nil, err
	}
	return &example, nil
}

var _ IExampleRepository = (*exampleRepositoryImpl)(nil)
`

const serviceTemplate = `package services

import (
	"errors"

	"{{.ModulePath}}/internal/entities"
	"{{.ModulePath}}/internal/repositories"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/configmgr"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/logger"
)

type IExampleService interface {
	common.IBaseService
	CreateExample(name string) (*entities.Example, error)
	GetExample(id uint) (*entities.Example, error)
}

type exampleServiceImpl struct {
	Config     configmgr.IConfigManager          ` + "`" + `inject:""` + "`" + `
	Repository repositories.IExampleRepository ` + "`" + `inject:""` + "`" + `
	LoggerMgr  loggermgr.ILoggerManager         ` + "`" + `inject:""` + "`" + `
}

func NewExampleService() IExampleService {
	return &exampleServiceImpl{}
}

func (s *exampleServiceImpl) ServiceName() string {
	return "ExampleService"
}

func (s *exampleServiceImpl) OnStart() error {
	return nil
}

func (s *exampleServiceImpl) OnStop() error {
	return nil
}

func (s *exampleServiceImpl) CreateExample(name string) (*entities.Example, error) {
	if name == "" {
		return nil, errors.New("名称不能为空")
	}

	example := &entities.Example{
		Name: name,
	}

	if err := s.Repository.Create(example); err != nil {
		return nil, err
	}

	s.LoggerMgr.Ins().Info("示例创建成功", "id", example.ID, "name", name)
	return example, nil
}

func (s *exampleServiceImpl) GetExample(id uint) (*entities.Example, error) {
	return s.Repository.GetByID(id)
}

var _ IExampleService = (*exampleServiceImpl)(nil)
`

const controllerTemplate = `package controllers

import (
	"fmt"
	"net/http"

	"{{.ModulePath}}/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
)

type IExampleController interface {
	common.IBaseController
}

type exampleControllerImpl struct {
	ExampleService services.IExampleService ` + "`" + `inject:""` + "`" + `
	LoggerMgr     loggermgr.ILoggerManager ` + "`" + `inject:""` + "`" + `
}

func NewExampleController() IExampleController {
	return &exampleControllerImpl{}
}

func (c *exampleControllerImpl) ControllerName() string {
	return "ExampleController"
}

func (c *exampleControllerImpl) GetRouter() string {
	return "/api/examples [POST],/api/examples/:id [GET]"
}

func (c *exampleControllerImpl) Handle(ctx *gin.Context) {
	method := ctx.Request.Method

	if method == "POST" {
		c.handleCreate(ctx)
	} else if method == "GET" {
		c.handleGet(ctx)
	}
}

func (c *exampleControllerImpl) handleCreate(ctx *gin.Context) {
	var req struct {
		Name string ` + "`" + `json:"name" binding:"required"` + "`" + `
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	example, err := c.ExampleService.CreateExample(req.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, example)
}

func (c *exampleControllerImpl) handleGet(ctx *gin.Context) {
	id := ctx.Param("id")
	var exampleID uint
	if _, err := fmt.Sscanf(id, "%d", &exampleID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	example, err := c.ExampleService.GetExample(exampleID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "示例不存在"})
		return
	}

	ctx.JSON(http.StatusOK, example)
}

var _ IExampleController = (*exampleControllerImpl)(nil)
`

const middlewareTemplate = `package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
)

type IRecoveryMiddleware interface {
	common.IBaseMiddleware
}

type recoveryMiddlewareImpl struct {
	LoggerMgr loggermgr.ILoggerManager ` + "`" + `inject:""` + "`" + `
}

func NewRecoveryMiddleware() IRecoveryMiddleware {
	return &recoveryMiddlewareImpl{}
}

func (m *recoveryMiddlewareImpl) MiddlewareName() string {
	return "RecoveryMiddleware"
}

func (m *recoveryMiddlewareImpl) Order() int {
	return 1
}

func (m *recoveryMiddlewareImpl) Wrapper() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		m.LoggerMgr.Ins().Error("panic recovered",
			"error", recovered,
			"path", c.Request.URL.Path,
			"method", c.Request.Method)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "内部服务器错误",
		})
	})
}

func (m *recoveryMiddlewareImpl) OnStart() error {
	return nil
}

func (m *recoveryMiddlewareImpl) OnStop() error {
	return nil
}

var _ IRecoveryMiddleware = (*recoveryMiddlewareImpl)(nil)
`

const listenerTemplate = `package listeners

import (
	"context"

	"{{.ModulePath}}/internal/services"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/mqmgr"
)

type IExampleListener interface {
	common.IBaseListener
}

type exampleListenerImpl struct {
	ExampleService services.IExampleService ` + "`" + `inject:""` + "`" + `
	MQManager     mqmgr.IMQManager        ` + "`" + `inject:""` + "`" + `
}

func NewExampleListener() IExampleListener {
	return &exampleListenerImpl{}
}

func (l *exampleListenerImpl) ListenerName() string {
	return "ExampleListener"
}

func (l *exampleListenerImpl) Subscribe() mqmgr.SubscribeConfig {
	return mqmgr.SubscribeConfig{
		Channel: "example.events",
		Handler: l.handleMessage,
	}
}

func (l *exampleListenerImpl) handleMessage(ctx context.Context, msg []byte) error {
	l.ExampleService.GetExample(1)
	return nil
}

func (l *exampleListenerImpl) OnStart() error {
	return nil
}

func (l *exampleListenerImpl) OnStop() error {
	return nil
}

var _ IExampleListener = (*exampleListenerImpl)(nil)
`

const schedulerTemplate = `package schedulers

import (
	"{{.ModulePath}}/internal/services"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/schedulermgr"
)

type IExampleScheduler interface {
	common.IBaseScheduler
}

type exampleSchedulerImpl struct {
	ExampleService  services.IExampleService ` + "`" + `inject:""` + "`" + `
	SchedulerManager schedulermgr.ISchedulerManager ` + "`" + `inject:""` + "`" + `
}

func NewExampleScheduler() IExampleScheduler {
	return &exampleSchedulerImpl{}
}

func (s *exampleSchedulerImpl) SchedulerName() string {
	return "ExampleScheduler"
}

func (s *exampleSchedulerImpl) Schedule() schedulermgr.ScheduleConfig {
	return schedulermgr.ScheduleConfig{
		CronExpression: "0 */5 * * * *",
		Handler:        s.handleTask,
	}
}

func (s *exampleSchedulerImpl) handleTask() error {
	s.ExampleService.GetExample(1)
	return nil
}

func (s *exampleSchedulerImpl) OnStart() error {
	return nil
}

func (s *exampleSchedulerImpl) OnStop() error {
	return nil
}

var _ IExampleScheduler = (*exampleSchedulerImpl)(nil)
`

const htmlTemplateServiceTemplate = `package services

import (
	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/liteservice"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
)

type IHTMLTemplateService interface {
	common.IBaseService
	Render(ctx *gin.Context, name string, data interface{})
	SetGinEngine(engine *gin.Engine)
}

type htmlTemplateServiceImpl struct {
	inner     *liteservice.HTMLTemplateService
	LoggerMgr loggermgr.ILoggerManager ` + "`" + `inject:""` + "`" + `
}

func NewHTMLTemplateService() IHTMLTemplateService {
	return &htmlTemplateServiceImpl{
		inner: liteservice.NewHTMLTemplateService("templates/*"),
	}
}

func (s *htmlTemplateServiceImpl) ServiceName() string {
	return "HTMLTemplateService"
}

func (s *htmlTemplateServiceImpl) OnStart() error {
	return s.inner.OnStart()
}

func (s *htmlTemplateServiceImpl) OnStop() error {
	return s.inner.OnStop()
}

func (s *htmlTemplateServiceImpl) Render(ctx *gin.Context, name string, data interface{}) {
	s.inner.Render(ctx, name, data)
}

func (s *htmlTemplateServiceImpl) SetGinEngine(engine *gin.Engine) {
	s.inner.SetGinEngine(engine)
}

var _ IHTMLTemplateService = (*htmlTemplateServiceImpl)(nil)
`

const staticControllerTemplate = `package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/litecontroller"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
)

type IStaticController interface {
	common.IBaseController
}

type staticControllerImpl struct {
	componentController *litecontroller.ResourceStaticController
	LoggerMgr           loggermgr.ILoggerManager ` + "`" + `inject:""` + "`" + `
}

func NewStaticController() IStaticController {
	return &staticControllerImpl{
		componentController: litecontroller.NewResourceStaticController("/static", "./static"),
	}
}

func (c *staticControllerImpl) ControllerName() string {
	return "staticControllerImpl"
}

func (c *staticControllerImpl) GetRouter() string {
	return c.componentController.GetRouter()
}

func (c *staticControllerImpl) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

var _ IStaticController = (*staticControllerImpl)(nil)
`

const healthControllerTemplate = `package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/litecontroller"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
)

type IHealthController interface {
	common.IBaseController
}

type healthControllerImpl struct {
	componentController litecontroller.IHealthController
	LoggerMgr           loggermgr.ILoggerManager ` + "`" + `inject:""` + "`" + `
}

func NewHealthController() IHealthController {
	return &healthControllerImpl{
		componentController: litecontroller.NewHealthController(),
	}
}

func (c *healthControllerImpl) ControllerName() string {
	return "healthControllerImpl"
}

func (c *healthControllerImpl) GetRouter() string {
	return c.componentController.GetRouter()
}

func (c *healthControllerImpl) Handle(ctx *gin.Context) {
	c.componentController.Handle(ctx)
}

var _ IHealthController = (*healthControllerImpl)(nil)
`

const pageControllerTemplate = `package controllers

import (
	"{{.ModulePath}}/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
)

type IPageController interface {
	common.IBaseController
}

type pageControllerImpl struct {
	HTMLTemplateService services.IHTMLTemplateService ` + "`" + `inject:""` + "`" + `
	LoggerMgr           loggermgr.ILoggerManager      ` + "`" + `inject:""` + "`" + `
}

func NewPageController() IPageController {
	return &pageControllerImpl{}
}

func (c *pageControllerImpl) ControllerName() string {
	return "pageControllerImpl"
}

func (c *pageControllerImpl) GetRouter() string {
	return "/ [GET]"
}

func (c *pageControllerImpl) Handle(ctx *gin.Context) {
	c.HTMLTemplateService.Render(ctx, "index.html", gin.H{
		"title": "{{.ProjectName}}",
	})
}

var _ IPageController = (*pageControllerImpl)(nil)
`

const staticCSSTemplate = `* {
	margin: 0;
	padding: 0;
	box-sizing: border-box;
}

body {
	font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
	line-height: 1.6;
	color: #333;
}

.container {
	max-width: 1200px;
	margin: 0 auto;
	padding: 0 20px;
}

header {
	background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
	color: white;
	padding: 60px 0;
	margin-bottom: 40px;
}

header h1 {
	font-size: 3rem;
	font-weight: 300;
	margin-bottom: 10px;
}

header p {
	font-size: 1.2rem;
	opacity: 0.9;
}

.card {
	border: none;
	box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
	margin-bottom: 20px;
	transition: transform 0.3s, box-shadow 0.3s;
}

.card:hover {
	transform: translateY(-5px);
	box-shadow: 0 8px 15px rgba(0, 0, 0, 0.15);
}

.message-list {
	margin-bottom: 30px;
}

.message-item {
	padding: 20px;
	margin-bottom: 15px;
	border-radius: 8px;
	background: white;
	border-left: 4px solid #667eea;
}

.message-item .nickname {
	font-weight: 600;
	color: #667eea;
	margin-bottom: 5px;
}

.message-item .content {
	color: #555;
	white-space: pre-wrap;
}

.message-item .timestamp {
	font-size: 0.875rem;
	color: #999;
	margin-top: 10px;
}

.btn-primary {
	background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
	border: none;
	padding: 12px 30px;
	font-size: 1rem;
	transition: all 0.3s;
}

.btn-primary:hover {
	transform: translateY(-2px);
	box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
}

footer {
	padding: 30px 0;
	color: #999;
	border-top: 1px solid #eee;
}
`

const staticJSTemplate = `$(document).ready(function() {
	// Load messages
	loadMessages();

	// Form validation and submit
	$('#message-form').validate({
		rules: {
			nickname: {
				required: true,
				minlength: 2,
				maxlength: 20
			},
			content: {
				required: true,
				minlength: 5,
				maxlength: 500
			}
		},
		messages: {
			nickname: {
				required: 'Please enter nickname',
				minlength: 'Nickname must be at least 2 characters',
				maxlength: 'Nickname must be no more than 20 characters'
			},
			content: {
				required: 'Please enter message content',
				minlength: 'Content must be at least 5 characters',
				maxlength: 'Content must be no more than 500 characters'
			}
		},
		submitHandler: function(form) {
			submitMessage(form);
		}
	});

	// Load messages
	function loadMessages() {
		$.ajax({
			url: '/api/messages',
			method: 'GET',
			success: function(messages) {
				renderMessages(messages);
			},
			error: function(xhr, status, error) {
				console.error('Failed to load messages:', error);
				$('#message-list').html(
					'<div class="alert alert-warning">' +
						'Failed to load messages, please refresh page' +
					'</div>'
				);
			}
		});
	}

	// Render messages
	function renderMessages(messages) {
		if (!messages || messages.length === 0) {
			$('#message-list').html(
				'<div class="alert alert-info">' +
					'No messages yet, be the first to post!' +
				'</div>'
			);
			return;
		}

		var html = '';
		for (var i = 0; i < messages.length; i++) {
			var msg = messages[i];
			html += '<div class="message-item">' +
					'<div class="nickname">' + escapeHtml(msg.nickname) + '</div>' +
					'<div class="content">' + escapeHtml(msg.content) + '</div>' +
					'<div class="timestamp">' + new Date(msg.created_at).toLocaleString('zh-CN') + '</div>' +
				'</div>';
		}

		$('#message-list').html(html);
	}

	// Submit message
	function submitMessage(form) {
		var formData = {
			nickname: $(form).find('#nickname').val(),
			content: $(form).find('#content').val()
		};

		$.ajax({
			url: '/api/messages',
			method: 'POST',
			contentType: 'application/json',
			data: JSON.stringify(formData),
			success: function(response) {
				$(form)[0].reset();
				loadMessages();
				$('#form-section').before(
					'<div class="alert alert-success" role="alert">' +
						'Message posted successfully!' +
					'</div>'
				);
				setTimeout(function() {
					$('.alert-success').fadeOut();
				}, 3000);
			},
			error: function(xhr, status, error) {
				var errorMsg = xhr.responseJSON && xhr.responseJSON.error ? xhr.responseJSON.error : 'Failed to submit, please try again later';
				$('#form-section').before(
					'<div class="alert alert-danger" role="alert">' +
						errorMsg +
					'</div>'
				);
				setTimeout(function() {
					$('.alert-danger').fadeOut();
				}, 3000);
			}
		});
	}

	// HTML escape
	function escapeHtml(text) {
		if (!text) return '';
		const div = document.createElement('div');
		div.textContent = text;
		return div.innerHTML;
	}
});
`

const indexHTMLTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.title}}</title>
    <!-- Bootstrap 5 CSS -->
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <!-- 自定义样式 -->
    <link href="/static/css/style.css" rel="stylesheet">
</head>
<body>
    <div class="container">
        <header class="text-center py-5">
            <h1 class="display-4 fw-light">{{.title}}</h1>
            <p class="text-muted">欢迎使用 LiteCore 框架</p>
        </header>

        <!-- 留言列表 -->
        <section id="messages-section" class="mb-5">
            <h2 class="h4 mb-4 fw-light">最新留言</h2>
            <div id="message-list" class="message-list">
                <div class="text-center text-muted py-5">
                    <div class="spinner-border spinner-border-sm" role="status"></div>
                    <p class="mt-2">加载中...</p>
                </div>
            </div>
        </section>

        <!-- 提交表单 -->
        <section id="form-section" class="mt-5">
            <h2 class="h4 mb-4 fw-light">发表留言</h2>
            <div class="card border-0 shadow-sm">
                <div class="card-body p-4">
                    <form id="message-form">
                        <div class="mb-3">
                            <label for="nickname" class="form-label">昵称</label>
                            <input type="text" class="form-control" id="nickname" name="nickname"
                                   placeholder="请输入您的昵称（2-20个字符）" required minlength="2" maxlength="20">
                        </div>
                        <div class="mb-3">
                            <label for="content" class="form-label">留言内容</label>
                            <textarea class="form-control" id="content" name="content" rows="5"
                                      placeholder="请输入留言内容（5-500个字符）" required minlength="5" maxlength="500"></textarea>
                        </div>
                        <button type="submit" class="btn btn-primary w-100">提交留言</button>
                    </form>
                </div>
            </div>
        </section>

        <footer class="text-center py-5 mt-5 text-muted">
            <small>&copy; 2025 {{.title}}. All rights reserved.</small>
        </footer>
    </div>

    <!-- Bootstrap 5 JS -->
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
    <!-- jQuery -->
    <script src="https://code.jquery.com/jquery-3.7.0.min.js"></script>
    <!-- jQuery Validation -->
    <script src="https://cdn.jsdelivr.net/npm/jquery-validation@1.21.0/dist/jquery.validate.min.js"></script>
    <script src="/static/js/app.js"></script>
</body>
</html>
`

var (
	goModTmpl      *template.Template
	readmeTmpl     *template.Template
	configYamlTmpl *template.Template
	gitignoreTmpl  *template.Template

	serverMainTmpl   *template.Template
	generateMainTmpl *template.Template

	entityTmpl     *template.Template
	repositoryTmpl *template.Template
	serviceTmpl    *template.Template
	controllerTmpl *template.Template
	middlewareTmpl *template.Template
	listenerTmpl   *template.Template
	schedulerTmpl  *template.Template

	htmlTemplateServiceTmpl *template.Template
	staticControllerTmpl    *template.Template
	healthControllerTmpl    *template.Template
	pageControllerTmpl      *template.Template

	staticCSSTmpl *template.Template
	staticJSTmpl  *template.Template
)

func init() {
	goModTmpl = template.Must(template.New("go.mod").Parse(goModTemplate))
	readmeTmpl = template.Must(template.New("README.md").Parse(readmeTemplate))
	configYamlTmpl = template.Must(template.New("config.yaml").Parse(configYamlTemplate))
	gitignoreTmpl = template.Must(template.New(".gitignore").Parse(gitignoreTemplate))

	serverMainTmpl = template.Must(template.New("server_main.go").Parse(serverMainTemplate))
	generateMainTmpl = template.Must(template.New("generate_main.go").Parse(generateMainTemplate))

	entityTmpl = template.Must(template.New("entity").Parse(entityTemplate))
	repositoryTmpl = template.Must(template.New("repository").Parse(repositoryTemplate))
	serviceTmpl = template.Must(template.New("service").Parse(serviceTemplate))
	controllerTmpl = template.Must(template.New("controller").Parse(controllerTemplate))
	middlewareTmpl = template.Must(template.New("middleware").Parse(middlewareTemplate))
	listenerTmpl = template.Must(template.New("listener").Parse(listenerTemplate))
	schedulerTmpl = template.Must(template.New("scheduler").Parse(schedulerTemplate))

	htmlTemplateServiceTmpl = template.Must(template.New("html_template_service").Parse(htmlTemplateServiceTemplate))
	staticControllerTmpl = template.Must(template.New("static_controller").Parse(staticControllerTemplate))
	healthControllerTmpl = template.Must(template.New("health_controller").Parse(healthControllerTemplate))
	pageControllerTmpl = template.Must(template.New("page_controller").Parse(pageControllerTemplate))

	staticCSSTmpl = template.Must(template.New("static_css").Parse(staticCSSTemplate))
	staticJSTmpl = template.Must(template.New("static_js").Parse(staticJSTemplate))
}

type TemplateData struct {
	ModulePath  string
	ProjectName string
	LitecoreVer string
}

func render(tmpl *template.Template, data *TemplateData) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("渲染模板失败: %w", err)
	}
	return buf.String(), nil
}

func GoMod(data *TemplateData) (string, error) {
	return render(goModTmpl, data)
}

func Readme(data *TemplateData) (string, error) {
	return render(readmeTmpl, data)
}

func ConfigYaml(data *TemplateData) (string, error) {
	return render(configYamlTmpl, data)
}

func Gitignore(data *TemplateData) (string, error) {
	return render(gitignoreTmpl, data)
}

func ServerMain(data *TemplateData) (string, error) {
	return render(serverMainTmpl, data)
}

func GenerateMain(data *TemplateData) (string, error) {
	return render(generateMainTmpl, data)
}

func Entity(data *TemplateData) (string, error) {
	return render(entityTmpl, data)
}

func Repository(data *TemplateData) (string, error) {
	return render(repositoryTmpl, data)
}

func Service(data *TemplateData) (string, error) {
	return render(serviceTmpl, data)
}

func Controller(data *TemplateData) (string, error) {
	return render(controllerTmpl, data)
}

func Middleware(data *TemplateData) (string, error) {
	return render(middlewareTmpl, data)
}

func Listener(data *TemplateData) (string, error) {
	return render(listenerTmpl, data)
}

func Scheduler(data *TemplateData) (string, error) {
	return render(schedulerTmpl, data)
}

func StaticCSS(data *TemplateData) (string, error) {
	return render(staticCSSTmpl, data)
}

func StaticJS(data *TemplateData) (string, error) {
	return render(staticJSTmpl, data)
}

func StaticController(data *TemplateData) (string, error) {
	return render(staticControllerTmpl, data)
}

func HTMLTemplateService(data *TemplateData) (string, error) {
	return render(htmlTemplateServiceTmpl, data)
}

func PageController(data *TemplateData) (string, error) {
	return render(pageControllerTmpl, data)
}

func HealthController(data *TemplateData) (string, error) {
	return render(healthControllerTmpl, data)
}
