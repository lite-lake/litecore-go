// Package litehtmltemplatesvc 提供 HTML 模板渲染服务，基于 Gin 的模板引擎实现。
//
// 核心特性：
//   - 模板渲染：Render / RenderWithCode 渲染 HTML 模板
//   - 模板加载：支持 glob 模式和目录扫描两种方式
//   - 热重载：支持开发模式下自动重载模板
//   - 自定义分隔符：支持自定义模板分隔符
//   - 内置模板函数：lower/upper/trim/safe/formatDate/dict/json/default
//   - i18n 集成：可选注入 I18nService 提供 t/tWithData 翻译函数
//   - 依赖注入：通过 inject 标签注入 LoggerMgr
//
// 基本用法：
//
//	// 方式一：使用默认配置
//	service := litehtmltemplatesvc.NewService()
//
//	// 方式二：使用自定义配置
//	config := &litehtmltemplatesvc.Config{
//	    TemplatePath: strPtr("templates/**/*.html"),
//	    HotReload:    boolPtr(true),
//	}
//	service := litehtmltemplatesvc.NewServiceWithConfig(config)
//
// 模板渲染：
//
//	// 设置 Gin 引擎
//	service.SetGinEngine(ginEngine)
//
//	// 加载模板
//	service.ReloadTemplates()
//
//	// 渲染模板
//	service.Render(ctx, "index.html", gin.H{"title": "首页"})
//	service.RenderWithCode(ctx, 200, "index.html", data)
//
// 内置模板函数：
//
//	{{ "HELLO" | lower }}                    → "hello"
//	{{ "hello" | upper }}                    → "HELLO"
//	{{ "  hello  " | trim }}                 → "hello"
//	{{ "<b>bold</b>" | safe }}               → 原样输出 HTML
//	{{ .Time | formatDate "2006-01-02" }}    → 格式化日期
//	{{ dict "name" "John" "age" 30 }}        → {"name": "John", "age": 30}
//	{{ .Data | json }}                       → JSON 编码
//	{{ .Value | default "N/A" }}             → 默认值
//	{{ t "zh-CN" "welcome" }}                → 翻译（需 i18n 服务）
//	{{ tWithData "zh-CN" "hello" (dict "name" "张三") }} → 带参数翻译
//
// 添加自定义函数：
//
//	service.AddFunc("myFunc", func(s string) string {
//	    return "prefix_" + s
//	})
//
//	service.AddFuncMap(template.FuncMap{
//	    "func1": func1,
//	    "func2": func2,
//	})
//
// 模板管理：
//
//	names := service.GetTemplateNames()
//	exists := service.HasTemplate("index.html")
//
// 使用依赖注入（集成 i18n）：
//
//	type MyController struct {
//	    HTMLService litehtmltemplatesvc.IService `inject:""`
//	    I18nService litei18nsvc.IService         `inject:""`
//	}
package litehtmltemplatesvc
