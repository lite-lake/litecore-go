// Package server 提供统一的服务引擎，让业务系统能够通过简单的方式创建和启动服务。
//
// # 核心功能
//
// 容器管理 - 业务系统创建 Container 并注册组件，然后传给 Server
// 自动依赖注入 - Engine 自动处理注入顺序和依赖关系
// 生命周期管理 - 自动管理 Manager 的启动和停止
// HTTP 服务集成 - 自动注册 Controller 和 Middleware 到 Gin 引擎
// 健康检查 - 提供统一的健康检查端点
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
//	configContainer.RegisterByType(reflect.TypeOf((*config.BaseConfigProvider)(nil)).Elem(), configProvider)
//
//	// 注册管理器
//	dbMgr := databasemgr.NewDatabaseManager()
//	managerContainer.RegisterByType(reflect.TypeOf((*databasemgr.DatabaseManager)(nil)).Elem(), dbMgr)
//
//	// 注册实体
//	entityContainer.Register(&entity.User{})
//
//	// 注册仓储
//	userRepo := repository.NewUserRepository()
//	repositoryContainer.RegisterByType(reflect.TypeOf((*repository.IUserRepository)(nil)).Elem(), userRepo)
//
//	// 注册服务
//	userService := service.NewUserService()
//	serviceContainer.RegisterByType(reflect.TypeOf((*service.IUserService)(nil)).Elem(), userService)
//
//	// 注册控制器
//	userController := controller.NewUserController()
//	controllerContainer.RegisterByType(reflect.TypeOf((*controller.IUserController)(nil)).Elem(), userController)
//
//	// 注册中间件
//	authMiddleware := middleware.NewAuthMiddleware()
//	middlewareContainer.RegisterByType(reflect.TypeOf((*middleware.IAuthMiddleware)(nil)).Elem(), authMiddleware)
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
// # 自定义配置
//
//	serverConfig := &server.ServerConfig{
//	    Host:            "0.0.0.0",
//	    Port:            9090,
//	    Mode:            "debug",
//	    EnableMetrics:   true,
//	    EnableHealth:    true,
//	    EnablePprof:     true,
//	    ShutdownTimeout: 60 * time.Second,
//	}
//	engine := server.NewEngineWithConfig(
//	    serverConfig,
//	    configContainer,
//	    entityContainer,
//	    managerContainer,
//	    repositoryContainer,
//	    serviceContainer,
//	    controllerContainer,
//	    middlewareContainer,
//	)
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
//
// # 系统路由
//
//	GET /health        - 健康检查
//	GET /healthz       - 健康检查（Kubernetes 标准）
//	GET /live          - 存活检查
//	GET /ready         - 就绪检查
//	GET /metrics       - Prometheus 指标
//	GET /debug/pprof/* - 性能分析（可选启用）
package server
