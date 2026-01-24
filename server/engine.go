package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/container"
	"github.com/lite-lake/litecore-go/logger"
	"github.com/lite-lake/litecore-go/manager/configmgr"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/manager/schedulermgr"
)

// Engine 服务引擎
type Engine struct {
	// 内置配置（在 Initialize 时用于初始化内置组件）
	builtinConfig *BuiltinConfig

	// 容器
	Manager    *container.ManagerContainer // 内置组件（在 Initialize 时初始化）
	Entity     *container.EntityContainer
	Repository *container.RepositoryContainer
	Service    *container.ServiceContainer
	Controller *container.ControllerContainer
	Middleware *container.MiddlewareContainer
	Listener   *container.ListenerContainer
	Scheduler  *container.SchedulerContainer

	// HTTP 服务器
	httpServer *http.Server
	ginEngine  *gin.Engine

	// 配置
	serverConfig    *serverConfig
	shutdownTimeout time.Duration
	autoMigrateDB   bool // 是否自动迁移数据库

	// 生命周期管理
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	mu      sync.RWMutex

	// 启动日志配置
	startupLogConfig *StartupLogConfig

	// 启动时间统计
	startupStartTime time.Time
	phaseDurations   map[StartupPhase]time.Duration
	phaseStartTimes  map[StartupPhase]time.Time

	// 日志器（统一使用 logger.ILogger）
	internalLogger logger.ILogger
	isStartup      bool         // 标识是否处于启动阶段
	loggerMu       sync.RWMutex // 保护日志器的并发访问

	// 异步日志器
	asyncLogger *AsyncStartupLogger
}

func NewEngine(
	builtinConfig *BuiltinConfig,
	entity *container.EntityContainer,
	repository *container.RepositoryContainer,
	service *container.ServiceContainer,
	controller *container.ControllerContainer,
	middleware *container.MiddlewareContainer,
	listener *container.ListenerContainer,
	scheduler *container.SchedulerContainer,
) *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	defaultConfig := defaultServerConfig()

	return &Engine{
		Entity:           entity,
		Repository:       repository,
		Service:          service,
		Controller:       controller,
		Middleware:       middleware,
		Listener:         listener,
		Scheduler:        scheduler,
		serverConfig:     defaultConfig,
		shutdownTimeout:  defaultConfig.ShutdownTimeout,
		ctx:              ctx,
		cancel:           cancel,
		builtinConfig:    builtinConfig,
		startupLogConfig: defaultConfig.StartupLog,
		phaseDurations:   make(map[StartupPhase]time.Duration),
		phaseStartTimes:  make(map[StartupPhase]time.Time),
	}
}

func (e *Engine) logger() logger.ILogger {
	return e.getLogger()
}

// setLogger 设置日志器（线程安全）
func (e *Engine) setLogger(logger logger.ILogger) {
	e.loggerMu.Lock()
	defer e.loggerMu.Unlock()
	e.internalLogger = logger
}

// getLogger 获取日志器（线程安全）
func (e *Engine) getLogger() logger.ILogger {
	e.loggerMu.RLock()
	defer e.loggerMu.RUnlock()
	return e.internalLogger
}

// Initialize 初始化引擎（实现 liteServer 接口）
// - 初始化内置组件（BuiltinConfig、Logger、Telemetry、Database、Cache）
// - 创建 Gin 引擎
// - 注册全局中间件
// - 注册系统路由
// - 注册控制器路由
func (e *Engine) Initialize() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 初始化启动时间统计
	e.startupStartTime = time.Now()

	// 初始化前使用默认日志器
	e.setLogger(logger.NewDefaultLogger("Engine"))
	e.isStartup = true

	// 1. 初始化内置组件
	builtInManagerContainer, err := Initialize(e.builtinConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize builtin components: %w", err)
	}
	e.Manager = builtInManagerContainer

	// 读取自动迁移配置
	e.autoMigrateDB = false
	if configMgr := e.Manager.GetByType(reflect.TypeOf((*configmgr.IConfigManager)(nil)).Elem()); configMgr != nil {
		if mgr, ok := configMgr.(configmgr.IConfigManager); ok {
			if autoMigrate, err := mgr.Get("database.auto_migrate"); err == nil {
				if autoMigrateBool, ok := autoMigrate.(bool); ok {
					e.autoMigrateDB = autoMigrateBool
				}
			}
		}
	}

	// 切换到结构化日志
	if loggerMgr, err := container.GetManager[loggermgr.ILoggerManager](e.Manager); err == nil {
		e.setLogger(loggerMgr.Ins())
		e.isStartup = false
		e.getLogger().Info("Switched to structured logging system")

		// 初始化异步日志器
		if e.startupLogConfig.Async {
			e.asyncLogger = NewAsyncStartupLogger(e.getLogger(), e.startupLogConfig.Buffer)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Failed to get logger manager: %v, using default logger\n", err)
	}

	// 2. 验证 Scheduler 配置（在依赖注入之前）
	if e.Scheduler != nil {
		e.logPhaseStart(PhaseValidation, "Starting to validate Scheduler configuration")

		// Manager 验证
		schedulerMgr, err := container.GetManager[schedulermgr.ISchedulerManager](e.Manager)
		if err == nil {
			schedulers := e.Scheduler.GetAll()
			for _, scheduler := range schedulers {
				if err := schedulerMgr.ValidateScheduler(scheduler); err != nil {
					panic(fmt.Sprintf("scheduler %s crontab validation failed: %v", scheduler.SchedulerName(), err))
				}
			}
		}

		e.logPhaseEnd(PhaseValidation, "Scheduler configuration validation complete", logger.F("count", e.Scheduler.Count()))
	}

	// 3. 自动依赖注入
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

	// 添加 NoRoute 处理器
	e.ginEngine.NoRoute(func(c *gin.Context) {
		e.logger().Warn("Route not found", "path", c.Request.URL.Path, "method", c.Request.Method)
		c.JSON(common.HTTPStatusNotFound, gin.H{
			"error":  "route not found",
			"path":   c.Request.URL.Path,
			"method": c.Request.Method,
		})
	})

	// 注册控制器路由
	if err := e.registerControllers(); err != nil {
		return fmt.Errorf("register controllers failed: %w", err)
	}

	// 初始化需要 Gin 引擎的服务（如 HTML 模板服务）
	e.initializeGinEngineServices()

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
	e.logPhaseStart(PhaseInjection, "Starting dependency injection")

	// 1. Entity 层（无需依赖注入）

	// 2. Repository 层（依赖 Manager + Entity）
	e.Repository.SetManagerContainer(e.Manager)
	if err := e.Repository.InjectAll(); err != nil {
		return fmt.Errorf("repository inject failed: %w", err)
	}
	repos := e.Repository.GetAll()
	for _, repo := range repos {
		e.logStartup(PhaseInjection, fmt.Sprintf("[%s layer] %s: injection complete", "Repository", repo.RepositoryName()))
	}

	// 3. Service 层（依赖 Manager + Repository + 同层）
	e.Service.SetManagerContainer(e.Manager)
	if err := e.Service.InjectAll(); err != nil {
		return fmt.Errorf("service inject failed: %w", err)
	}
	svcs := e.Service.GetAll()
	for _, svc := range svcs {
		e.logStartup(PhaseInjection, fmt.Sprintf("[%s layer] %s: injection complete", "Service", svc.ServiceName()))
	}

	// 4. Controller 层（依赖 Manager + Service）
	e.Controller.SetManagerContainer(e.Manager)
	if err := e.Controller.InjectAll(); err != nil {
		return fmt.Errorf("controller inject failed: %w", err)
	}
	ctrls := e.Controller.GetAll()
	for _, ctrl := range ctrls {
		e.logStartup(PhaseInjection, fmt.Sprintf("[%s layer] %s: injection complete", "Controller", ctrl.ControllerName()))
	}

	// 5. Middleware 层（依赖 Manager + Service）
	e.Middleware.SetManagerContainer(e.Manager)
	if err := e.Middleware.InjectAll(); err != nil {
		return fmt.Errorf("middleware inject failed: %w", err)
	}
	mws := e.Middleware.GetAll()
	for _, mw := range mws {
		e.logStartup(PhaseInjection, fmt.Sprintf("[%s layer] %s: injection complete", "Middleware", mw.MiddlewareName()))
	}

	// 6. Listener 层
	if e.Listener != nil {
		e.Listener.SetManagerContainer(e.Manager)
		if err := e.Listener.InjectAll(); err != nil {
			return fmt.Errorf("listener inject failed: %w", err)
		}
		listeners := e.Listener.GetAll()
		for _, listener := range listeners {
			e.logStartup(PhaseInjection, fmt.Sprintf("[%s layer] %s: injection complete", "Listener", listener.ListenerName()))
		}
	}

	// 7. Scheduler 层（新增）
	if e.Scheduler != nil {
		e.Scheduler.SetManagerContainer(e.Manager)
		if err := e.Scheduler.InjectAll(); err != nil {
			return fmt.Errorf("scheduler inject failed: %w", err)
		}
		schedulers := e.Scheduler.GetAll()
		for _, scheduler := range schedulers {
			e.logStartup(PhaseInjection, fmt.Sprintf("[%s layer] %s: injection complete", "Scheduler", scheduler.SchedulerName()))
		}
	}

	totalCount := len(repos) + len(svcs) + len(ctrls) + len(mws)
	if e.Listener != nil {
		totalCount += len(e.Listener.GetAll())
	}
	if e.Scheduler != nil {
		totalCount += len(e.Scheduler.GetAll())
	}
	e.logPhaseEnd(PhaseInjection, "Dependency injection complete", logger.F("count", totalCount))

	return nil
}

// Start 启动引擎（实现 liteServer 接口）
// - 启动所有 Manager
// - 启动所有 Repository
// - 启动所有 Service
// - 启动所有 Middleware
// - 启动 HTTP 服务器
func (e *Engine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.started {
		return fmt.Errorf("engine already started")
	}

	// 1. 启动所有 Manager
	if err := e.startManagers(); err != nil {
		return fmt.Errorf("start managers failed: %w", err)
	}

	// 2. 自动迁移数据库（如果启用）
	if e.autoMigrateDB {
		if err := e.autoMigrateDatabase(); err != nil {
			return fmt.Errorf("auto migrate database failed: %w", err)
		}
	}

	// 3. 启动所有 Repository
	if err := e.startRepositories(); err != nil {
		return fmt.Errorf("start repositories failed: %w", err)
	}

	// 3. 启动所有 Service
	if err := e.startServices(); err != nil {
		return fmt.Errorf("start services failed: %w", err)
	}

	// 4. 启动所有 Middleware
	if err := e.startMiddlewares(); err != nil {
		return fmt.Errorf("start middlewares failed: %w", err)
	}

	// 5. 启动所有 Scheduler（新增）
	if err := e.startSchedulers(); err != nil {
		return fmt.Errorf("start schedulers failed: %w", err)
	}

	// 6. 启动所有 Listener
	if err := e.startListeners(); err != nil {
		return fmt.Errorf("start listeners failed: %w", err)
	}

	// 停止异步日志器
	if e.asyncLogger != nil {
		e.asyncLogger.Stop()
		e.asyncLogger = nil
	}

	// 6. 启动 HTTP 服务器
	e.logger().Info("HTTP server listening", "addr", e.httpServer.Addr)

	errChan := make(chan error, 1)
	go func() {
		if err := e.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			e.logger().Error("HTTP server error", "error", err)
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	select {
	case err := <-errChan:
		return fmt.Errorf("HTTP server failed to start: %w", err)
	case <-time.After(100 * time.Millisecond):
		e.logger().Debug("HTTP server started successfully")
	}

	// 记录启动完成汇总
	totalDuration := time.Since(e.startupStartTime)
	e.logPhaseStart(PhaseStartup, "Service startup complete, starting to serve requests",
		logger.F("addr", e.httpServer.Addr),
		logger.F("total_duration", totalDuration.String()))

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

// getGinEngine 获取 Gin 引擎
func (e *Engine) getGinEngine() *gin.Engine {
	return e.ginEngine
}

// registerControllers 注册所有控制器路由
func (e *Engine) registerControllers() error {
	e.logPhaseStart(PhaseRouter, "Starting to register routes")

	controllers := e.Controller.GetAll()
	registeredCount := 0

	for _, ctrl := range controllers {
		route := ctrl.GetRouter()
		if route == "" {
			continue
		}

		method, path, err := parseRoute(route)
		if err != nil {
			e.getLogger().Warn("Invalid route format",
				logger.F("controller", ctrl.ControllerName()),
				logger.F("error", err))
			continue
		}

		handler := ctrl.Handle
		e.registerRoute(method, path, handler)

		e.logStartup(PhaseRouter, "Registered route",
			logger.F("method", method),
			logger.F("path", path),
			logger.F("controller", ctrl.ControllerName()))
		registeredCount++
	}

	e.logPhaseEnd(PhaseRouter, "Route registration complete",
		logger.F("route_count", registeredCount),
		logger.F("controller_count", len(controllers)))

	return nil
}

// parseRoute 解析路由字符串
// 支持格式: "/path [METHOD]" (OpenAPI 风格)
// 返回: method (大写), path, error
func parseRoute(route string) (string, string, error) {
	for i := len(route) - 1; i >= 0; i-- {
		if route[i] == '[' {
			if i+1 < len(route) && route[len(route)-1] == ']' {
				method := route[i+1 : len(route)-1]
				path := route[:i]
				path = strings.TrimSpace(path)
				method = strings.TrimSpace(method)
				if path == "" {
					return "", "", fmt.Errorf("path cannot be empty")
				}
				if method == "" {
					return "", "", fmt.Errorf("method cannot be empty")
				}
				return strings.ToUpper(method), path, nil
			}
		}
	}
	return "", "", fmt.Errorf("invalid route format, expected 'path [METHOD]'")
}

// initializeGinEngineServices 初始化需要 Gin 引擎的服务
func (e *Engine) initializeGinEngineServices() {
	services := e.Service.GetAll()
	for _, svc := range services {
		if ginEngineSetter, ok := svc.(interface{ SetGinEngine(*gin.Engine) }); ok {
			ginEngineSetter.SetGinEngine(e.ginEngine)
		}
	}
}

// WaitForShutdown 等待关闭信号
func (e *Engine) WaitForShutdown() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-sigs
	e.logger().Info("Received shutdown signal", "signal", sig)

	if err := e.Stop(); err != nil {
		e.logger().Fatal("Shutdown error", "error", err)
		os.Exit(1)
	}
}
