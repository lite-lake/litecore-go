// Package server 提供统一的服务引擎，让业务系统能够通过简单的方式创建和启动服务。
//
// # 核心功能
//
// 容器管理 - 业务系统创建 Container 并注册组件，然后传给 Server
// 自动依赖注入 - Engine 自动处理注入顺序和依赖关系
// 生命周期管理 - 自动管理 Manager、Repository、Service 的启动和停止
// 中间件集成 - 自动注册 Middleware 到 Gin 引擎
// 优雅关闭 - 支持信号处理和优雅关闭
//
// # 基本使用
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
//	// 注册配置
//	configProvider, _ := config.NewConfigProvider("yaml", "config.yaml")
//	container.RegisterConfig[common.BaseConfigProvider](configContainer, configProvider)
//
//	// 注册管理器
//	dbMgr := databasemgr.NewDatabaseManager()
//	container.RegisterManager[databasemgr.DatabaseManager](managerContainer, dbMgr)
//
//	// 注册实体
//	entityContainer.Register(&entity.User{})
//
//	// 注册仓储
//	userRepo := repository.NewUserRepository()
//	container.RegisterRepository[repository.IUserRepository](repositoryContainer, userRepo)
//
//	// 注册服务
//	userService := service.NewUserService()
//	container.RegisterService[service.IUserService](serviceContainer, userService)
//
//	// 注册控制器（可选组件）
//	userController := controller.NewUserController()
//	container.RegisterController[controller.IUserController](controllerContainer, userController)
//
//	// 注册中间件
//	authMiddleware := middleware.NewAuthMiddleware()
//	container.RegisterMiddleware[middleware.IAuthMiddleware](middlewareContainer, authMiddleware)
//
//	// 创建引擎，传入容器
//	engine := server.NewEngine(
//	    configContainer,
//	    entityContainer,
//	    managerContainer,
//	    repositoryContainer,
//	    serviceContainer,
//	    controllerContainer,
//	    middlewareContainer,
//	)
//
//	// 启动引擎
//	if err := engine.Run(); err != nil {
//	    panic(err)
//	}
//
// # 服务器配置
//
// Engine 使用内部默认配置 DefaultServerConfig，支持以下配置项：
//
//	Host/Port           - HTTP 监听地址和端口
//	Mode                - Gin 运行模式（debug/release/test）
//	EnableRecovery      - 是否启用 panic 恢复
//	ReadTimeout         - HTTP 读取超时
//	WriteTimeout        - HTTP 写入超时
//	IdleTimeout         - HTTP 空闲超时
//	ShutdownTimeout     - 优雅关闭超时
//
// # 自定义路由扩展
//
//	if err := engine.Initialize(); err != nil {
//	    panic(err)
//	}
//
//	// 获取 Gin 引擎进行自定义扩展
//	r := engine.GetGinEngine()
//	r.GET("/custom", func(c *gin.Context) {
//	    c.JSON(200, gin.H{"message": "custom route"})
//	})
//
//	if err := engine.Run(); err != nil {
//	    panic(err)
//	}
package server
