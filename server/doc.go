// Package server 提供统一的 HTTP 服务引擎，支持自动依赖注入、生命周期管理和中间件集成。
//
// 核心特性：
//   - 五层容器管理：集成 Entity、Repository、Service、Controller、Middleware
//   - 内置组件：BuiltinConfig 和 Manager 作为内置组件，由引擎自动初始化和注入
//   - 自动依赖注入：按依赖顺序自动处理组件注入
//   - 生命周期管理：统一管理各层组件的启动和停止
//   - 中间件集成：自动排序并注册全局中间件到 Gin 引擎
//   - 路由管理：自动注册控制器路由
//   - 优雅关闭：支持信号处理，超时控制的安全关闭机制
//   - 自动初始化：自动初始化需要 Gin 引擎的服务（如 HTML 模板服务）
//
// 基本用法：
//
//	loggermgr "github.com/lite-lake/litecore-go/component/manager/loggermgr"
//
//	loggerMgr := loggermgr.GetLoggerManager()
//	logger := loggerMgr.Logger("main")
//
//	// 创建应用引擎（由 CLI 工具生成）
//	engine, err := app.NewEngine()
//	if err != nil {
//	    logger.Fatal("创建引擎失败", "error", err)
//	}
//
//	// 启动服务
//	if err := engine.Run(); err != nil {
//	    logger.Fatal("引擎启动失败", "error", err)
//	}
//
// 分步启动（需要自定义初始化时）：
//
//	// 初始化引擎
//	if err := engine.Initialize(); err != nil {
//	    logger.Fatal("初始化引擎失败", "error", err)
//	}
//
//	// 启动服务
//	if err := engine.Start(); err != nil {
//	    logger.Fatal("启动引擎失败", "error", err)
//	}
//
//	// 等待关闭信号
//	engine.WaitForShutdown()
//
// 依赖注入：
//
// 各层组件通过 inject:"" 标签声明依赖，Manager 由引擎自动注入：
//
//	type UserServiceImpl struct {
//	    BuiltinConfig    configmgr.IConfigManager    `inject:""`  // 内置组件
//	    DBManager databasemgr.IDatabaseManager `inject:""`  // 内置组件
//	    UserRepo  repository.IUserRepository `inject:""`
//	}
//
// 中间件排序：
//
// 中间件通过 Order() 方法定义执行顺序（越小越先执行）：
//
//	type AuthMiddleware struct {}
//
//	func (m *AuthMiddleware) Order() int {
//	    return 100
//	}
//
// 控制器路由：
//
// 控制器通过 GetRouter() 方法定义路由，格式为："/path [METHOD]"：
//
//	func (ctrl *UserController) GetRouter() string {
//	    return "/users [GET]"
//	}
//
// 自动初始化服务：
//
// 服务可实现 SetGinEngine(*gin.Engine) 接口，Engine.Initialize() 会自动调用：
//
//	type HTMLTemplateService struct {
//	    ginEngine *gin.Engine
//	}
//
//	func (s *HTMLTemplateService) SetGinEngine(engine *gin.Engine) {
//	    s.ginEngine = engine
//	    // 初始化 HTML 模板
//	    engine.LoadHTMLGlob("templates/*.html")
//	}
package server
