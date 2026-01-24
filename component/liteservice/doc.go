// Package liteservice 提供内置服务组件，用于 HTML 模板渲染和模板管理。
//
// 核心特性：
//   - HTML 模板渲染：基于 Gin 的 HTML 模板渲染功能
//   - 模板加载：支持通配符路径加载多个模板文件
//   - 生命周期管理：实现 OnStart/OnStop 钩子，自动加载和释放模板
//   - 依赖注入：通过 inject 标签注入 Manager 和其他组件
//   - 配置驱动：支持通过配置指定模板文件路径
//
// 基本用法：
//
//	// 创建 HTML 模板服务
//	htmlService := liteservice.NewHTMLTemplateService("templates/*")
//
//	// 设置 Gin 引擎（由 Engine 自动调用）
//	ginEngine := gin.New()
//	htmlService.SetGinEngine(ginEngine)
//
//	// 启动服务（加载模板）
//	err := htmlService.OnStart()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 在控制器中使用模板渲染
//	router.GET("/page", func(ctx *gin.Context) {
//	    data := gin.H{
//	        "title": "欢迎",
//	        "message": "Hello World",
//	    }
//	    htmlService.Render(ctx, "index.html", data)
//	})
//
// 使用依赖注入：
//
//	// 在 Service 中注入 HTMLTemplateService
//	type PageService struct {
//	    HTMLService liteservice.IHTMLTemplateService `inject:""`
//	    LoggerMgr   loggermgr.ILoggerManager        `inject:""`
//	}
//
//	func (s *PageService) RenderPage(ctx *gin.Context) {
//	    data := gin.H{"title": "示例页面"}
//	    s.HTMLService.Render(ctx, "page.html", data)
//	}
//
// 模板配置：
//
//	模板文件路径支持通配符语法：
//	- "templates/*"：加载 templates 目录下所有模板
//	- "views/**/*.html"：递归加载 views 目录下所有 HTML 文件
//	- "templates/*.html"：只加载 templates 目录下 .html 后缀的文件
//
// 注册到容器：
//
//	// 创建服务实例
//	htmlService := liteservice.NewHTMLTemplateService("templates/*")
//
//	// 注册到服务容器
//	serviceContainer.RegisterService(htmlService)
//
//	// Engine 启动时会自动调用 OnStart 加载模板
package liteservice
