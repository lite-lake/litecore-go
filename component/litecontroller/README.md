# Litecontroller

内置控制器组件，提供健康检查、性能分析、指标监控和资源管理功能。

## 特性

- **健康检查** - 提供 `/health` 端点，检查所有管理器的健康状态，返回详细的健康报告
- **性能分析** - 集成 pprof 工具，支持堆、协程、内存、锁等性能分析
- **指标监控** - 提供 `/metrics` 端点，展示服务器运行状态和组件数量
- **资源管理** - 支持 HTML 模板渲染和静态文件服务
- **依赖注入** - 通过 `inject:""` 标签注入 Manager 和其他组件
- **自动路由** - 根据 `GetRouter()` 定义的路由规则自动注册到 Gin 路由器

## 快速开始

```go
package main

import (
	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/component/litecontroller"
	"github.com/lite-lake/litecore-go/container"
	"github.com/lite-lake/litecore-go/server"
)

func main() {
	// 创建容器
	entityContainer := container.NewEntityContainer()
	repositoryContainer := container.NewRepositoryContainer(entityContainer)
	serviceContainer := container.NewServiceContainer(repositoryContainer)
	controllerContainer := container.NewControllerContainer(serviceContainer)

	// 注册内置控制器
	controllerContainer.RegisterController(
		litecontroller.NewHealthController(),
		litecontroller.NewMetricsController(),
		litecontroller.NewPprofIndexController(),
		litecontroller.NewPprofHeapController(),
		litecontroller.NewPprofGoroutineController(),
	)

	// 创建引擎
	engine := server.NewEngine(
		&server.BuiltinConfig{
			Driver:   "yaml",
			FilePath: "configs/config.yaml",
		},
		entityContainer,
		repositoryContainer,
		serviceContainer,
		controllerContainer,
		nil, // middlewareContainer
		nil, // listenerContainer
		nil, // schedulerContainer
	)

	// 启动服务
	engine.Run()
}
```

## 健康检查控制器

健康检查控制器提供 `/health` 端点，用于检查所有管理器的健康状态。

```go
import "github.com/lite-lake/litecore-go/component/litecontroller"

healthCtrl := litecontroller.NewHealthController()
controllerContainer.RegisterController(healthCtrl)
```

### 响应示例

```json
{
  "status": "ok",
  "timestamp": "2026-01-25T10:30:00+08:00",
  "managers": {
    "DatabaseManager": "ok",
    "CacheManager": "ok",
    "LoggerManager": "ok"
  }
}
```

### 响应状态码

- `200 OK` - 所有管理器健康
- `503 Service Unavailable` - 至少一个管理器不健康

## 性能分析控制器

性能分析控制器集成 pprof 工具，提供多个端点用于性能分析。

```go
import "github.com/lite-lake/litecore-go/component/litecontroller"

// 注册性能分析控制器
controllerContainer.RegisterController(
	litecontroller.NewPprofIndexController(),        // /debug/pprof
	litecontroller.NewPprofHeapController(),         // /debug/pprof/heap
	litecontroller.NewPprofGoroutineController(),    // /debug/pprof/goroutine
	litecontroller.NewPprofAllocsController(),      // /debug/pprof/allocs
	litecontroller.NewPprofBlockController(),        // /debug/pprof/block
	litecontroller.NewPprofMutexController(),        // /debug/pprof/mutex
	litecontroller.NewPprofProfileController(),      // /debug/pprof/profile
	litecontroller.NewPprofTraceController(),        // /debug/pprof/trace
	litecontroller.NewPprofSymbolController(),       // /debug/pprof/symbol
	litecontroller.NewPprofSymbolPostController(),   // /debug/pprof/symbol (POST)
	litecontroller.NewPprofCmdlineController(),     // /debug/pprof/cmdline
	litecontroller.NewPprofThreadcreateController(), // /debug/pprof/threadcreate
)
```

### 端点说明

| 端点 | 方法 | 说明 |
|------|------|------|
| `/debug/pprof` | GET | pprof 索引页面 |
| `/debug/pprof/heap` | GET | 堆内存分析 |
| `/debug/pprof/goroutine` | GET | 协程栈跟踪 |
| `/debug/pprof/allocs` | GET | 内存分配分析 |
| `/debug/pprof/block` | GET | 阻塞分析 |
| `/debug/pprof/mutex` | GET | 互斥锁分析 |
| `/debug/pprof/profile` | GET | CPU 性能分析 |
| `/debug/pprof/trace` | GET | 执行跟踪 |
| `/debug/pprof/symbol` | GET/POST | 符号表查询 |
| `/debug/pprof/cmdline` | GET | 命令行参数 |
| `/debug/pprof/threadcreate` | GET | 线程创建分析 |

### 使用示例

```bash
# 查看 pprof 索引
curl http://localhost:8080/debug/pprof/

# 获取堆内存快照
go tool pprof http://localhost:8080/debug/pprof/heap

# 获取协程栈
go tool pprof http://localhost:8080/debug/pprof/goroutine

# 进行 CPU 性能分析（30秒）
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof

# 生成 SVG 可视化
go tool pprof -http=:8080 http://localhost:8080/debug/pprof/heap
```

## 指标控制器

指标控制器提供 `/metrics` 端点，用于展示服务器运行状态。

```go
import "github.com/lite-lake/litecore-go/component/litecontroller"

metricsCtrl := litecontroller.NewMetricsController()
controllerContainer.RegisterController(metricsCtrl)
```

### 响应示例

```json
{
  "server": "litecore-go",
  "status": "running",
  "version": "1.0.0",
  "managers": 5,
  "services": 3
}
```

## 静态文件控制器

静态文件控制器用于提供静态文件服务，如 CSS、JS、图片等。

```go
import "github.com/lite-lake/litecore-go/component/litecontroller"

// 创建静态文件控制器
// URLPath: 访问路径前缀，如 /static
// FilePath: 文件系统路径，如 ./static
staticCtrl := litecontroller.NewResourceStaticController("/static", "./static")
controllerContainer.RegisterController(staticCtrl)
```

### 访问示例

```bash
# 访问 static 目录下的文件
http://localhost:8080/static/css/style.css
http://localhost:8080/static/js/app.js
http://localhost:8080/static/images/logo.png
```

### 配置结构

```go
type ResourceStaticConfig struct {
    URLPath  string // URL路径前缀
    FilePath string // 文件系统路径
}
```

## HTML 模板控制器

HTML 模板控制器用于加载和渲染 HTML 模板。

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/lite-lake/litecore-go/component/litecontroller"
)

// 创建 HTML 模板控制器
htmlCtrl := litecontroller.NewResourceHTMLController("templates/*")

// 加载模板（需要在注册控制器前调用）
htmlCtrl.LoadTemplates(ginEngine)

// 注册控制器
controllerContainer.RegisterController(htmlCtrl)

// 在业务代码中使用
func (c *MyController) Handle(ctx *gin.Context) {
    data := gin.H{
        "Title":   "首页",
        "Content": "欢迎访问 litecore-go",
    }
    htmlCtrl.Render(ctx, "index.html", data)
}
```

### 模板目录结构

```
templates/
├── index.html
├── user/
│   ├── profile.html
│   └── settings.html
└── admin/
    └── dashboard.html
```

### 模板示例

```html
<!-- templates/index.html -->
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
</head>
<body>
    <h1>{{.Title}}</h1>
    <p>{{.Content}}</p>
</body>
</html>
```

### 配置结构

```go
type ResourceHTMLConfig struct {
    TemplatePath string // 模板文件路径模式
}
```

## API

### 控制器接口

所有控制器都实现 `common.IBaseController` 接口：

```go
type IBaseController interface {
    ControllerName() string  // 返回控制器名称
    GetRouter() string       // 返回路由定义，如 "/health [GET]"
    Handle(ctx *gin.Context) // 处理 HTTP 请求
}
```

### 健康检查控制器

```go
// 创建健康检查控制器
func NewHealthController() IHealthController

// 健康检查响应
type HealthResponse struct {
    Status    string            `json:"status"`     // ok 或 degraded
    Timestamp string            `json:"timestamp"`  // RFC3339 格式时间
    Managers  map[string]string `json:"managers"`   // 各管理器状态
}
```

### 指标控制器

```go
// 创建指标控制器
func NewMetricsController() IMetricsController
```

### 性能分析控制器

```go
// 创建 pprof 索引控制器
func NewPprofIndexController() IPprofIndexController

// 创建堆内存分析控制器
func NewPprofHeapController() IPprofHeapController

// 创建协程分析控制器
func NewPprofGoroutineController() IPprofGoroutineController

// 创建内存分配分析控制器
func NewPprofAllocsController() IPprofAllocsController

// 创建阻塞分析控制器
func NewPprofBlockController() IPprofBlockController

// 创建互斥锁分析控制器
func NewPprofMutexController() IPprofMutexController

// 创建 CPU 性能分析控制器
func NewPprofProfileController() IPprofProfileController

// 创建执行跟踪控制器
func NewPprofTraceController() IPprofTraceController

// 创建符号表查询控制器 (GET)
func NewPprofSymbolController() IPprofSymbolController

// 创建符号表查询控制器 (POST)
func NewPprofSymbolPostController() IPprofSymbolPostController

// 创建命令行参数控制器
func NewPprofCmdlineController() IPprofCmdlineController

// 创建线程创建分析控制器
func NewPprofThreadcreateController() IPprofThreadcreateController
```

### 静态文件控制器

```go
// 创建静态文件控制器
func NewResourceStaticController(urlPath, filePath string) *ResourceStaticController

// 获取配置
func (c *ResourceStaticController) GetConfig() *ResourceStaticConfig

// 配置结构
type ResourceStaticConfig struct {
    URLPath  string // URL路径前缀
    FilePath string // 文件系统路径
}
```

### HTML 模板控制器

```go
// 创建 HTML 模板控制器
func NewResourceHTMLController(templatePath string) *ResourceHTMLController

// 加载模板
func (c *ResourceHTMLController) LoadTemplates(engine *gin.Engine)

// 渲染模板
func (c *ResourceHTMLController) Render(ctx *gin.Context, name string, data interface{})

// 获取配置
func (c *ResourceHTMLController) GetConfig() *ResourceHTMLConfig

// 配置结构
type ResourceHTMLConfig struct {
    TemplatePath string // 模板文件路径模式
}
```

## 错误处理

所有控制器都支持依赖注入日志管理器，用于记录错误信息：

```go
type HealthController struct {
    ManagerContainer common.IBaseManager      `inject:""`
    LoggerMgr        loggermgr.ILoggerManager `inject:""`
}
```

### 错误示例

```json
{
  "status": "degraded",
  "timestamp": "2026-01-25T10:30:00+08:00",
  "managers": {
    "DatabaseManager": "unhealthy: connection refused",
    "CacheManager": "ok"
  }
}
```

## 最佳实践

### 1. 生产环境配置

生产环境中建议：

```go
// 仅在开发环境启用性能分析
if gin.Mode() == gin.DebugMode {
    controllerContainer.RegisterController(
        litecontroller.NewPprofIndexController(),
        litecontroller.NewPprofHeapController(),
        litecontroller.NewPprofGoroutineController(),
    )
}

// 生产环境启用健康检查和指标
controllerContainer.RegisterController(
    litecontroller.NewHealthController(),
    litecontroller.NewMetricsController(),
)
```

### 2. 安全考虑

性能分析端点可能暴露敏感信息，建议：

- 使用中间件进行身份验证
- 限制访问 IP 范围
- 仅在需要时启用

```go
// 使用 IP 限制中间件
router.GET("/debug/pprof/*filepath", middleware.IPWhitelistMiddleware(), pprofHandler)
```

### 3. 静态文件缓存

静态文件控制器使用 Gin 的默认缓存策略，如需自定义缓存头：

```go
// 在业务控制器中添加缓存头
func (c *MyController) Handle(ctx *gin.Context) {
    ctx.Header("Cache-Control", "public, max-age=3600")
    staticCtrl.Handle(ctx)
}
```

### 4. HTML 模板优化

```go
// 使用模板继承
// templates/layouts/base.html
{{define "base"}}
<!DOCTYPE html>
<html>
<head>
    <title>{{block "title" .}}默认标题{{end}}</title>
</head>
<body>
    {{block "content" .}}默认内容{{end}}
</body>
</html>
{{end}}

// templates/index.html
{{define "title"}}首页{{end}}
{{define "content"}}
<h1>欢迎</h1>
{{end}}
{{template "base" .}}
```

## 相关文档

- [AGENTS.md](../../AGENTS.md) - 项目开发指南
- [README.md](../../README.md) - 项目主文档
- [common/README.md](../../common/README.md) - 公共接口说明
- [container/README.md](../../container/README.md) - 依赖注入容器说明
