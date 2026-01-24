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

## 更多信息

- [LiteCore Go 文档](https://github.com/lite-lake/litecore-go)
- [CLI 工具文档](../../cli/README.md)
- [示例项目](../../samples/)
`

const configYamlTemplate = `app:
  name: "{{.ProjectName}}"
  version: "1.0.0"

server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"
  read_timeout: 60
  write_timeout: 60
  max_header_bytes: 1048576

database:
  driver: "sqlite"
  auto_migrate: true
  sqlite_config:
    dsn: "./data/app.db"
  gorm_config:
    skip_default_transaction: false
    prepare_stmt: true

logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "info"
      format: "gin"
      color: true
      time_format: "2006-01-02 15:04:05.000"
    file_enabled: true
    file_config:
      filename: "./logs/app.log"
      max_size: 100
      max_backups: 10
      max_age: 30
      compress: true
      level: "info"

cache:
  driver: "memory"
  memory_config:
    max_size: 100
    max_age: "720h"
    max_backups: 1000
    compress: false

limiter:
  driver: "memory"
  memory_config:
    max_requests: 1000
    window: 60

lock:
  driver: "memory"
  memory_config:
    max_locks: 10000

mq:
  driver: "memory"
  memory_config:
    max_queue_size: 10000
    channel_buffer: 100

telemetry:
  driver: "none"

scheduler:
  driver: "cron"
  cron_config:
    validate_on_startup: true
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
