package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	"com.litelake.litecore/container"
)

// Engine 服务引擎
type Engine struct {
	// 容器
	config     *container.ConfigContainer
	entity     *container.EntityContainer
	manager    *container.ManagerContainer
	repository *container.RepositoryContainer
	service    *container.ServiceContainer
	controller *container.ControllerContainer
	middleware *container.MiddlewareContainer

	// HTTP 服务器
	httpServer *http.Server
	ginEngine  *gin.Engine

	// 配置
	serverConfig    *ServerConfig
	shutdownTimeout time.Duration

	// 生命周期管理
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	mu      sync.RWMutex

	// 初始化错误（用于延迟报告）
	initErrors []error
}

// NewEngine 创建服务引擎
func NewEngine(opts ...EngineOption) (*Engine, error) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 创建引擎
	engine := &Engine{
		serverConfig:    DefaultServerConfig(),
		shutdownTimeout: DefaultServerConfig().ShutdownTimeout,
		ctx:             ctx,
		cancel:          cancel,
		initErrors:      make([]error, 0),
	}

	// 创建容器（两步初始化以避免依赖问题）
	engine.config = container.NewConfigContainer()
	engine.entity = container.NewEntityContainer()
	engine.manager = container.NewManagerContainer(engine.config)

	// 继续初始化其他容器（依赖于已初始化的容器）
	engine.repository = container.NewRepositoryContainer(
		engine.config,
		engine.manager,
		engine.entity,
	)
	engine.service = container.NewServiceContainer(
		engine.config,
		engine.manager,
		engine.repository,
	)
	engine.controller = container.NewControllerContainer(
		engine.config,
		engine.manager,
		engine.service,
	)
	engine.middleware = container.NewMiddlewareContainer(
		engine.config,
		engine.manager,
		engine.service,
	)

	// 应用选项
	for _, opt := range opts {
		if opt != nil {
			opt(engine)
		}
	}

	// 检查初始化错误
	if len(engine.initErrors) > 0 {
		return nil, fmt.Errorf("engine initialization failed: %v", engine.initErrors)
	}

	return engine, nil
}

// Initialize 初始化引擎（实现 LiteServer 接口）
// - 创建 Gin 引擎
// - 注册全局中间件
// - 注册业务中间件和控制器
// - 注册系统路由
func (e *Engine) Initialize() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 自动依赖注入
	if err := e.autoInject(); err != nil {
		return fmt.Errorf("auto inject failed: %w", err)
	}

	// 设置 Gin 模式
	gin.SetMode(e.serverConfig.Mode)

	// 创建 Gin 引擎
	e.ginEngine = gin.New()

	// 注册中间件
	if err := e.registerMiddlewares(); err != nil {
		return fmt.Errorf("register middlewares failed: %w", err)
	}

	// 注册控制器
	e.registerControllers()

	// 添加 NoRoute 处理器用于调试
	e.ginEngine.NoRoute(func(c *gin.Context) {
		fmt.Printf("[NoRoute] Path: %s, Method: %s\n", c.Request.URL.Path, c.Request.Method)
		c.JSON(404, gin.H{"error": "route not found", "path": c.Request.URL.Path, "method": c.Request.Method})
	})

	// 注册系统路由
	if e.serverConfig.EnableHealth {
		e.registerHealthRoute()
	}

	if e.serverConfig.EnableMetrics {
		e.registerMetricsRoute()
	}

	if e.serverConfig.EnablePprof {
		e.registerPprofRoutes()
	}

	// 创建 HTTP 服务器
	e.httpServer = &http.Server{
		Addr:         e.serverConfig.Address(),
		Handler:      e.ginEngine,
		ReadTimeout:  e.serverConfig.ReadTimeout,
		WriteTimeout: e.serverConfig.WriteTimeout,
		IdleTimeout:  e.serverConfig.IdleTimeout,
	}

	return nil
}

// autoInject 自动依赖注入
func (e *Engine) autoInject() error {
	// 按依赖顺序自动注入
	// 1. Config 层（无依赖）- 已经在 NewEngine 中完成
	// 2. Entity 层（无依赖）
	if err := e.entity.InjectAll(); err != nil {
		return fmt.Errorf("entity inject failed: %w", err)
	}

	// 3. Manager 层（依赖 Config + 同层）
	if err := e.manager.InjectAll(); err != nil {
		return fmt.Errorf("manager inject failed: %w", err)
	}

	// 4. Repository 层（依赖 Config + Manager + Entity）
	if err := e.repository.InjectAll(); err != nil {
		return fmt.Errorf("repository inject failed: %w", err)
	}

	// 5. Service 层（依赖 Config + Manager + Repository + 同层）
	if err := e.service.InjectAll(); err != nil {
		return fmt.Errorf("service inject failed: %w", err)
	}

	// 6. Controller 层（依赖 Config + Manager + Service）
	if err := e.controller.InjectAll(); err != nil {
		return fmt.Errorf("controller inject failed: %w", err)
	}

	// 7. Middleware 层（依赖 Config + Manager + Service）
	if err := e.middleware.InjectAll(); err != nil {
		return fmt.Errorf("middleware inject failed: %w", err)
	}

	return nil
}

// Start 启动引擎（实现 LiteServer 接口）
// - 启动所有 Manager
// - 启动所有 Repository
// - 启动所有 Service
// - 启动 HTTP 服务器
func (e *Engine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.started {
		return fmt.Errorf("engine already started")
	}

	fmt.Println("[DEBUG] Starting all managers...")
	// 1. 启动所有 Manager
	if err := e.startManagers(); err != nil {
		return fmt.Errorf("start managers failed: %w", err)
	}
	fmt.Println("[DEBUG] All managers started successfully")

	fmt.Println("[DEBUG] Starting all repositories...")
	// 2. 启动所有 Repository
	if err := e.startRepositories(); err != nil {
		return fmt.Errorf("start repositories failed: %w", err)
	}
	fmt.Println("[DEBUG] All repositories started successfully")

	fmt.Println("[DEBUG] Starting all services...")
	// 3. 启动所有 Service
	if err := e.startServices(); err != nil {
		return fmt.Errorf("start services failed: %w", err)
	}
	fmt.Println("[DEBUG] All services started successfully")

	// 4. 启动 HTTP 服务器
	go func() {
		if err := e.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			e.cancel()
		}
	}()

	e.started = true
	return nil
}

// Run 简化的启动方法
// 等价于 Initialize() + Start() + 等待信号
func (e *Engine) Run() error {
	// 初始化
	if err := e.Initialize(); err != nil {
		return err
	}

	// 启动
	if err := e.Start(); err != nil {
		return err
	}

	// 等待关闭信号
	e.WaitForShutdown()

	return nil
}

// GetGinEngine 获取 Gin 引擎（用于自定义扩展）
func (e *Engine) GetGinEngine() *gin.Engine {
	return e.ginEngine
}

// GetConfig 获取配置
func (e *Engine) GetConfig() common.BaseConfigProvider {
	configs := e.config.GetAll()
	if len(configs) == 0 {
		return nil
	}
	return configs[0]
}

// GetLogger 获取日志记录器
// 如果有 LoggerManager，返回其默认 logger；否则返回 nil
func (e *Engine) GetLogger() interface{} {
	// 尝试获取 LoggerManager
	mgrs := e.manager.GetAll()
	for _, mgr := range mgrs {
		// 通过类型断言检查是否为 LoggerManager
		if loggerMgr, ok := mgr.(interface{ Logger(name string) interface{} }); ok {
			return loggerMgr.Logger("server")
		}
	}
	return nil
}

// Health 健康检查
func (e *Engine) Health() error {
	managers := e.manager.GetAll()
	for _, mgr := range managers {
		if err := mgr.Health(); err != nil {
			return fmt.Errorf("manager %s health check failed: %w", mgr.ManagerName(), err)
		}
	}
	return nil
}

// NewEngineWithConfig 使用配置提供者创建引擎
func NewEngineWithConfig(cfgProvider common.BaseConfigProvider, opts ...EngineOption) (*Engine, error) {
	allOpts := append([]EngineOption{WithConfig(cfgProvider)}, opts...)
	return NewEngine(allOpts...)
}

// SetMode 设置 Gin 运行模式
func (e *Engine) SetMode(mode string) {
	e.serverConfig.Mode = mode
	gin.SetMode(mode)
}

// IsStarted 检查引擎是否已启动
func (e *Engine) IsStarted() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.started
}
