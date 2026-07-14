# Liteservice

内置服务组件，提供 HTML 模板渲染和模板管理功能。

## 特性

- **HTML 模板渲染** - 基于 Gin 框架，支持 Go 原生模板语法
- **灵活的模板加载** - 支持通配符路径加载多个模板文件
- **生命周期管理** - 自动管理模板的加载和释放
- **依赖注入** - 支持通过 `inject:""` 标签注入 Manager 和其他组件
- **配置驱动** - 支持通过配置指定模板文件路径
- **错误处理** - 提供完善的错误处理和日志记录

## 快速开始

### 通过容器注册（推荐）

在 5 层依赖注入架构中，HTMLTemplateService 由容器管理：

```go
import (
    "github.com/lite-lake/litecore-go/component/liteservice"
    "github.com/lite-lake/litecore-go/container"
    "github.com/lite-lake/litecore-go/server"
)

// 创建服务容器
serviceContainer := container.NewServiceContainer(repositoryContainer)

// 注册 HTML 模板服务
htmlService := liteservice.NewHTMLTemplateService("templates/*")
serviceContainer.RegisterService(htmlService)

// 创建 Engine 并启动
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
)
engine.Run()
```

### 在 Controller 中使用

通过依赖注入获取 HTMLTemplateService：

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/lite-lake/litecore-go/component/liteservice"
)

type PageController struct {
    HTMLService liteservice.IHTMLTemplateService `inject:""`
}

func (c *PageController) ControllerName() string {
    return "PageController"
}

func (c *PageController) GetRouter() string {
    return "/page [GET]"
}

func (c *PageController) Handle(ctx *gin.Context) {
    data := gin.H{
        "title":   "示例页面",
        "content": "欢迎使用 liteservice 组件",
    }
    c.HTMLService.Render(ctx, "page.html", data)
}
```

### 在 Service 中使用

```go
type PageService struct {
    HTMLService liteservice.IHTMLTemplateService `inject:""`
    LoggerMgr   loggermgr.ILoggerManager        `inject:""`
}

func (s *PageService) RenderWelcomePage(ctx *gin.Context) error {
    data := gin.H{
        "title":   "欢迎",
        "message": "Hello World",
    }

    s.HTMLService.Render(ctx, "welcome.html", data)
    s.LoggerMgr.Ins().Info("渲染欢迎页面", "page", "welcome.html")

    return nil
}
```

## 核心功能

### 模板加载

模板加载在服务的 `OnStart` 钩子中自动完成，支持通配符路径：

```go
// 创建服务，指定模板路径
htmlService := liteservice.NewHTMLTemplateService("templates/*")

// Engine 启动时会自动调用 OnStart
err := htmlService.OnStart()
if err != nil {
    log.Fatal(err)
}
```

**路径语法示例：**

- `templates/*` - 加载 templates 目录下所有文件
- `views/**/*.html` - 递归加载 views 目录下所有 HTML 文件
- `templates/*.tmpl` - 只加载 templates 目录下 .tmpl 后缀的文件

### 模板渲染

使用 `Render` 方法渲染模板：

```go
// 渲染模板
data := gin.H{
    "title":   "页面标题",
    "user":    gin.H{"name": "张三", "age": 30},
    "items":   []string{"A", "B", "C"},
}

htmlService.Render(ctx, "index.html", data)
```

**错误处理：**

如果 Gin 引擎未设置或模板未加载，`Render` 方法会返回 500 错误：

```go
// 模板未加载时，自动返回错误响应
// HTTP/1.1 500 Internal Server Error
// {"error": "HTML templates not loaded"}
```

### 配置管理

通过 `GetConfig` 方法获取配置：

```go
htmlService := liteservice.NewHTMLTemplateService("templates/*")
config := htmlService.GetConfig()

// 访问配置
fmt.Println("模板路径:", config.TemplatePath)
```

### 生命周期管理

HTMLTemplateService 实现了 `common.IBaseService` 接口：

```go
type IBaseService interface {
    ServiceName() string
    OnStart() error
    OnStop() error
}
```

**OnStart：**

- 加载指定的模板文件
- 由 Engine 在启动时自动调用

**OnStop：**

- 释放 Gin 引擎引用
- 由 Engine 在停止时自动调用

## API 说明

### 工厂函数

| 函数 | 说明 |
|------|------|
| `NewHTMLTemplateService(templatePath string)` | 创建 HTML 模板服务实例 |

### 接口定义

#### IHTMLTemplateService

```go
type IHTMLTemplateService interface {
    common.IBaseService
    Render(ctx *gin.Context, name string, data interface{})
}
```

| 方法 | 说明 |
|------|------|
| `ServiceName() string` | 返回服务名称（固定为 "HTMLTemplateService"） |
| `OnStart() error` | 启动服务，加载 HTML 模板 |
| `OnStop() error` | 停止服务，释放资源 |
| `Render(ctx *gin.Context, name string, data interface{})` | 渲染 HTML 模板 |
| `SetGinEngine(engine *gin.Engine)` | 设置 Gin 引擎 |
| `GetConfig() *HTMLTemplateConfig` | 获取服务配置 |

### 配置结构

#### HTMLTemplateConfig

```go
type HTMLTemplateConfig struct {
    TemplatePath string // 模板文件路径模式
}
```

## 错误处理

### 模板未加载错误

如果 Gin 引擎未设置或模板未加载，`Render` 方法会返回错误：

```go
// 检查模板是否已加载
htmlService.Render(ctx, "page.html", data)

// 如果模板未加载，响应为：
// HTTP/1.1 500 Internal Server Error
// {"error": "HTML templates not loaded"}
```

### 模板加载错误

模板加载失败会在 `OnStart` 方法中返回错误：

```go
err := htmlService.OnStart()
if err != nil {
    // 模板加载失败
    log.Fatal("加载模板失败:", err)
}
```

## 模板示例

### 简单模板

```html
<!-- templates/index.html -->
<!DOCTYPE html>
<html>
<head>
    <title>{{ .title }}</title>
</head>
<body>
    <h1>{{ .title }}</h1>
    <p>{{ .message }}</p>
</body>
</html>
```

**渲染代码：**

```go
data := gin.H{
    "title":   "欢迎",
    "message": "Hello World",
}
htmlService.Render(ctx, "index.html", data)
```

### 列表渲染

```html
<!-- templates/list.html -->
<ul>
{{ range .items }}
    <li>{{ . }}</li>
{{ end }}
</ul>
```

**渲染代码：**

```go
data := gin.H{
    "items": []string{"Apple", "Banana", "Cherry"},
}
htmlService.Render(ctx, "list.html", data)
```

### 条件渲染

```html
<!-- templates/profile.html -->
{{ if .user }}
    <p>欢迎, {{ .user.name }}!</p>
{{ else }}
    <p>请先登录</p>
{{ end }}
```

**渲染代码：**

```go
data := gin.H{
    "user": gin.H{
        "name": "张三",
        "email": "zhangsan@example.com",
    },
}
htmlService.Render(ctx, "profile.html", data)
```

## 使用场景

### 渲染静态页面

```go
type HomeController struct {
    HTMLService liteservice.IHTMLTemplateService `inject:""`
}

func (c *HomeController) Handle(ctx *gin.Context) {
    c.HTMLService.Render(ctx, "home.html", gin.H{
        "title": "首页",
    })
}
```

### 渲染动态数据页面

```go
type ArticleController struct {
    HTMLService liteservice.IHTMLTemplateService `inject:""`
    ArticleRepo repository.IArticleRepository   `inject:""`
}

func (c *ArticleController) Handle(ctx *gin.Context) {
    // 获取文章数据
    article, err := c.ArticleRepo.GetByID(1)
    if err != nil {
        ctx.JSON(500, gin.H{"error": err.Error()})
        return
    }

    // 渲染模板
    c.HTMLService.Render(ctx, "article.html", gin.H{
        "title":   article.Title,
        "content": article.Content,
        "author":  article.Author,
    })
}
```

### 错误页面渲染

```go
func RenderErrorPage(ctx *gin.Context, htmlService liteservice.IHTMLTemplateService, code int, message string) {
    data := gin.H{
        "code":    code,
        "message": message,
    }
    htmlService.Render(ctx, "error.html", data)
}
```

## 最佳实践

### 1. 模板文件组织

推荐将模板文件按功能分类组织：

```
templates/
├── layout/
│   ├── base.html
│   └── header.html
├── pages/
│   ├── index.html
│   └── about.html
└── partials/
    ├── footer.html
    └── sidebar.html
```

### 2. 使用依赖注入

在 Service 中注入 HTMLTemplateService，便于测试和维护：

```go
type EmailService struct {
    HTMLService liteservice.IHTMLTemplateService `inject:""`
    LoggerMgr   loggermgr.ILoggerManager        `inject:""`
}

func (s *EmailService) SendWelcomeEmail(to, username string) error {
    data := gin.H{
        "username": username,
        "date":     time.Now().Format("2006-01-02"),
    }

    // 渲染邮件模板
    var buf bytes.Buffer
    // ... 渲染逻辑

    s.LoggerMgr.Ins().Info("发送欢迎邮件", "to", to)
    return nil
}
```

### 3. 错误处理

在 Controller 中处理模板渲染错误：

```go
func (c *PageController) Handle(ctx *gin.Context) {
    data := gin.H{
        "title": "示例页面",
    }

    // 使用 defer-recover 捕获渲染错误
    defer func() {
        if r := recover(); r != nil {
            ctx.JSON(500, gin.H{"error": "页面渲染失败"})
        }
    }()

    c.HTMLService.Render(ctx, "page.html", data)
}
```

### 4. 配置管理

通过配置文件管理模板路径：

```yaml
# config.yaml
templates:
  path: "templates/**/*.html"
```

**在代码中读取配置：**

```go
type ConfigService struct {
    Config configmgr.IConfigManager `inject:""`
}

func (s *ConfigService) GetTemplatePath() string {
    return configmgr.GetWithDefault[string](s.Config, "templates.path", "templates/*")
}
```

## 依赖关系

HTMLTemplateService 依赖以下组件：

```go
type HTMLTemplateService struct {
    config    *HTMLTemplateConfig
    ginEngine *gin.Engine
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}
```

- **LoggerMgr** - 日志记录（可选）

## 注意事项

1. **Gin 引擎设置**：`OnStart` 之前必须调用 `SetGinEngine`，否则模板无法加载
2. **模板路径**：使用相对路径时，相对于程序运行目录
3. **并发安全**：模板渲染是并发安全的，但模板加载不是
4. **错误处理**：模板不存在或语法错误会在加载时返回
5. **内存占用**：大量模板会增加内存占用，合理组织模板结构
