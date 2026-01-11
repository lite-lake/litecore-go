// Package server 提供统一的服务引擎，让业务系统能够通过简单的方式创建和启动服务。
//
// # 核心功能
//
// 自动容器管理 - Engine 内部自动创建和管理所有容器
// 自动依赖注入 - Engine 自动处理注入顺序和依赖关系
// 生命周期管理 - 自动管理 Manager 的启动和停止
// HTTP 服务集成 - 自动注册 Controller 和 Middleware 到 Gin 引擎
// 健康检查 - 提供统一的健康检查端点
// 优雅关闭 - 支持信号处理和优雅关闭
//
// # 基本使用
//
//	engine, err := server.NewEngine(
//	    server.WithConfigFile("config.json"),
//	    server.RegisterManagers(
//	        &manager.LoggerManager{},
//	        &manager.DatabaseManager{},
//	    ),
//	    server.RegisterEntities(
//	        &entity.User{},
//	    ),
//	    server.RegisterRepositories(
//	        &repository.UserRepository{},
//	    ),
//	    server.RegisterServices(
//	        &service.UserService{},
//	    ),
//	    server.RegisterControllers(
//	        &controller.UserController{},
//	    ),
//	    server.RegisterMiddlewares(
//	        &middleware.AuthMiddleware{},
//	    ),
//	)
//	if err != nil {
//	    panic(err)
//	}
//
//	if err := engine.Run(); err != nil {
//	    panic(err)
//	}
//
// # 自定义配置
//
//	engine, err := server.NewEngine(
//	    server.WithConfigFile("config.json"),
//	    server.RegisterManagers(&manager.LoggerManager{}),
//	    server.WithServerConfig(&server.ServerConfig{
//	        Host:            "0.0.0.0",
//	        Port:            9090,
//	        Mode:            "debug",
//	        EnableMetrics:   true,
//	        EnableHealth:    true,
//	        EnablePprof:     true,
//	        ShutdownTimeout: 60 * time.Second,
//	    }),
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
