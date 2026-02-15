# Liteservice

内置服务组件集合，提供常用的业务服务。

## 组件列表

| 子包 | 说明 |
|------|------|
| [litebloomsvc](./litebloomsvc) | 布隆过滤器服务 |
| [litei18nsvc](./litei18nsvc) | 国际化服务 |
| [litehtmltemplatesvc](./litehtmltemplatesvc) | HTML 模板服务 |

## 设计原则

- **统一接口** - 所有服务实现 `common.IBaseService`
- **依赖注入** - 支持通过 `inject:""` 标签注入 Manager 和其他组件
- **生命周期管理** - 实现 `OnStart()` / `OnStop()` 钩子
- **配置驱动** - 支持 `Config` 结构配置，提供 `NewXxx` / `NewXxxWithConfig` 工厂函数

## 快速开始

### litebloomsvc - 布隆过滤器

```go
import "github.com/lite-lake/litecore-go/component/liteservice/litebloomsvc"

// 创建服务
bloom := litebloomsvc.NewLiteBloomService()

// 创建过滤器
bloom.CreateFilterWithConfig("cache", &litebloomsvc.FilterConfig{
    ExpectedItems:     uintPtr(100000),
    FalsePositiveRate: floatPtr(0.01),
    TTL:               durationPtr(5 * time.Minute),
})

// 使用
bloom.AddString("cache", "key1")
exists, _ := bloom.ContainsString("cache", "key1") // true
```

### litei18nsvc - 国际化

```go
import "github.com/lite-lake/litecore-go/component/liteservice/litei18nsvc"

// 创建服务
i18n := litei18nsvc.NewLiteI18nService()

// 加载语言包
i18n.LoadLocale("en", map[string]string{"hello": "Hello"})
i18n.LoadLocale("zh", map[string]string{"hello": "你好"})

// 使用
i18n.T("zh", "hello")  // "你好"
i18n.T("en", "hello")  // "Hello"
```

### litehtmltemplatesvc - HTML 模板

```go
import "github.com/lite-lake/litecore-go/component/liteservice/litehtmltemplatesvc"

// 创建服务
tmpl := litehtmltemplatesvc.NewLiteHTMLTemplateServiceWithConfig(&litehtmltemplatesvc.Config{
    TemplatePath: strPtr("templates/**/*.html"),
})

// 设置 Gin 引擎
tmpl.SetGinEngine(ginEngine)

// 渲染
tmpl.Render(ctx, "index.html", gin.H{"title": "Hello"})
```

## 依赖关系

所有服务组件可选依赖 `LoggerMgr`：

```go
type MyService struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}
```

litehtmltemplatesvc 可选依赖 I18nService 以提供翻译函数：

```go
type liteHTMLTemplateServiceImpl struct {
    LoggerMgr loggermgr.ILoggerManager         `inject:""`
    I18nSvc   litehtmltemplatesvc.I18nService  `inject:""`  // 可选
}
```

## 测试

```bash
# 测试所有服务
go test ./component/liteservice/... -v

# 测试单个服务
go test ./component/liteservice/litebloomsvc/... -v
go test ./component/liteservice/litei18nsvc/... -v
go test ./component/liteservice/litehtmltemplatesvc/... -v
```
