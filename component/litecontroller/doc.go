// Package litecontroller 提供内置控制器组件，用于健康检查、性能分析、指标监控和资源管理。
//
// 核心特性：
//   - 健康检查：提供 /health 端点，检查所有管理器的健康状态
//   - 性能分析：集成 pprof 工具，支持堆、协程、内存等性能分析
//   - 指标监控：提供 /metrics 端点，展示服务器运行状态和组件数量
//   - 资源管理：支持 HTML 模板渲染和静态文件服务
//   - 依赖注入：通过 inject 标签注入 Manager 和其他组件
//
// 基本用法：
//
//	// 注册健康检查控制器
//	healthCtrl := litecontroller.NewHealthController()
//	controllerContainer.RegisterController(healthCtrl)
//
//	// 注册性能分析控制器
//	pprofIndexCtrl := litecontroller.NewPprofIndexController()
//	pprofHeapCtrl := litecontroller.NewPprofHeapController()
//	pprofGoroutineCtrl := litecontroller.NewPprofGoroutineController()
//	controllerContainer.RegisterController(pprofIndexCtrl, pprofHeapCtrl, pprofGoroutineCtrl)
//
//	// 注册指标控制器
//	metricsCtrl := litecontroller.NewMetricsController()
//	controllerContainer.RegisterController(metricsCtrl)
//
//	// 注册静态文件控制器
//	staticCtrl := litecontroller.NewResourceStaticController("/static", "./static")
//	controllerContainer.RegisterController(staticCtrl)
//
//	// 注册 HTML 模板控制器
//	htmlCtrl := litecontroller.NewResourceHTMLController("templates/*")
//	htmlCtrl.LoadTemplates(ginEngine)
//	controllerContainer.RegisterController(htmlCtrl)
//
// 控制器接口：
//
//	所有控制器都实现 common.IBaseController 接口，包含以下方法：
//	- ControllerName() string：返回控制器名称
//	- GetRouter() string：返回路由定义
//	- Handle(ctx *gin.Context)：处理 HTTP 请求
//
// 路由自动注册：
//
//	Engine 启动时会自动扫描 ControllerContainer 中的控制器，根据 GetRouter() 定义的
//	路由规则自动注册到 Gin 路由器。例如："/health [GET]" 会注册为 GET /health。
//
// 依赖注入：
//
//	控制器可以通过 inject 标签注入 Manager 和其他组件：
//	type HealthController struct {
//	    ManagerContainer common.IBaseManager      `inject:""`
//	    LoggerMgr        loggermgr.ILoggerManager `inject:""`
//	}
package litecontroller
