// Package server 提供统一的 HTTP 服务引擎，支持自动依赖注入、生命周期管理和中间件集成。
//
// 核心特性：
//   - 容器管理：集成 Config、Entity、Manager、Repository、Service、Controller、Middleware 七层容器
//   - 自动注入：按依赖顺序自动处理组件注入（Entity → Manager → Repository → Service → Controller/Middleware）
//   - 生命周期管理：统一管理各层组件的启动和停止，支持健康检查
//   - 中间件集成：自动排序并注册全局中间件到 Gin 引擎
//   - 路由管理：自动注册控制器路由，支持自定义路由扩展
//   - 优雅关闭：支持信号处理，超时控制的安全关闭机制
//
// 基本用法：
//
//	// 创建容器
//	configContainer := container.NewConfigContainer()
//	entityContainer := container.NewEntityContainer()
//	managerContainer := container.NewManagerContainer(configContainer)
//	repositoryContainer := container.NewRepositoryContainer(configContainer, managerContainer, entityContainer)
//	serviceContainer := container.NewServiceContainer(configContainer, managerContainer, repositoryContainer)
//	controllerContainer := container.NewControllerContainer(configContainer, managerContainer, serviceContainer)
//	middlewareContainer := container.NewMiddlewareContainer(configContainer, managerContainer, serviceContainer)
//
//	// 注册配置和管理器
//	configProvider, _ := config.NewConfigProvider("yaml", "config.yaml")
//	container.RegisterConfig[common.BaseConfigProvider](configContainer, configProvider)
//	dbMgr := databasemgr.NewDatabaseManager()
//	container.RegisterManager[databasemgr.DatabaseManager](managerContainer, dbMgr)
//
//	// 注册其他组件...
//
//	// 创建并启动引擎
//	engine := server.NewEngine(configContainer, entityContainer, managerContainer,
//	    repositoryContainer, serviceContainer, controllerContainer, middlewareContainer)
//	if err := engine.Run(); err != nil {
//	    panic(err)
//	}
//
// 服务器配置：
//
// Engine 使用默认配置 DefaultServerConfig，包含以下配置项：
//
//	Host/Port         - HTTP 监听地址和端口（默认 0.0.0.0:8080）
//	Mode              - Gin 运行模式：debug/release/test（默认 release）
//	EnableRecovery    - 是否启用 panic 恢复（默认 true）
//	ReadTimeout       - HTTP 读取超时（默认 10s）
//	WriteTimeout      - HTTP 写入超时（默认 10s）
//	IdleTimeout       - HTTP 空闲超时（默认 60s）
//	ShutdownTimeout   - 优雅关闭超时（默认 30s）
//
// 自定义路由：
//
//	// 初始化引擎后获取 Gin 引擎进行扩展
//	if err := engine.Initialize(); err != nil {
//	    panic(err)
//	}
//	r := engine.GetGinEngine()
//	r.GET("/custom", func(c *gin.Context) {
//	    c.JSON(200, gin.H{"message": "custom route"})
//	})
//	if err := engine.Run(); err != nil {
//	    panic(err)
//	}
//
// 生命周期控制：
//
//	// 分步启动
//	if err := engine.Initialize(); err != nil {
//	    panic(err)
//	}
//	if err := engine.Start(); err != nil {
//	    panic(err)
//	}
//	engine.WaitForShutdown()
//
//	// 健康检查
//	if err := engine.Health(); err != nil {
//	    log.Printf("Health check failed: %v", err)
//	}
package server
