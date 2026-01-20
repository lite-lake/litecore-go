package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"com.litelake.litecore/common"
	"com.litelake.litecore/container"
)

// Engine 服务引擎
type Engine struct {
	// 容器
	Config     *container.ConfigContainer
	Entity     *container.EntityContainer
	Manager    *container.ManagerContainer
	Repository *container.RepositoryContainer
	Service    *container.ServiceContainer
	Controller *container.ControllerContainer
	Middleware *container.MiddlewareContainer

	// HTTP 服务器
	httpServer *http.Server
	ginEngine  *gin.Engine

	// 配置
	serverConfig    *serverConfig
	shutdownTimeout time.Duration

	// 生命周期管理
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	mu      sync.RWMutex
}

// NewEngine 创建服务引擎
func NewEngine(
	config *container.ConfigContainer,
	entity *container.EntityContainer,
	manager *container.ManagerContainer,
	repository *container.RepositoryContainer,
	service *container.ServiceContainer,
	controller *container.ControllerContainer,
	middleware *container.MiddlewareContainer,
) *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	defaultConfig := defaultServerConfig()

	return &Engine{
		Config:          config,
		Entity:          entity,
		Manager:         manager,
		Repository:      repository,
		Service:         service,
		Controller:      controller,
		Middleware:      middleware,
		serverConfig:    defaultConfig,
		shutdownTimeout: defaultConfig.ShutdownTimeout,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// Initialize 初始化引擎（实现 liteServer 接口）
// - 创建 Gin 引擎
// - 注册全局中间件
// - 注册系统路由
// - 注册控制器路由
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

	// 添加 NoRoute 处理器用于调试
	e.ginEngine.NoRoute(func(c *gin.Context) {
		fmt.Printf("[NoRoute] Path: %s, Method: %s\n", c.Request.URL.Path, c.Request.Method)
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
	// 按依赖顺序自动注入
	// 1. Config 层（无依赖）- 已经在 NewEngine 中完成
	// 2. Entity 层（无依赖）
	if err := e.Entity.InjectAll(); err != nil {
		return fmt.Errorf("entity inject failed: %w", err)
	}

	// 3. Manager 层（依赖 Config + 同层）
	if err := e.Manager.InjectAll(); err != nil {
		return fmt.Errorf("manager inject failed: %w", err)
	}

	// 4. Repository 层（依赖 Config + Manager + Entity）
	if err := e.Repository.InjectAll(); err != nil {
		return fmt.Errorf("repository inject failed: %w", err)
	}

	// 5. Service 层（依赖 Config + Manager + Repository + 同层）
	if err := e.Service.InjectAll(); err != nil {
		return fmt.Errorf("service inject failed: %w", err)
	}

	// 6. Controller 层（依赖 Config + Manager + Service）
	if err := e.Controller.InjectAll(); err != nil {
		return fmt.Errorf("controller inject failed: %w", err)
	}

	// 7. Middleware 层（依赖 Config + Manager + Service）
	if err := e.Middleware.InjectAll(); err != nil {
		return fmt.Errorf("middleware inject failed: %w", err)
	}

	return nil
}

// Start 启动引擎（实现 liteServer 接口）
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
	errChan := make(chan error, 1)
	go func() {
		fmt.Println("[DEBUG] HTTP server listening on", e.httpServer.Addr)
		if err := e.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("[ERROR] HTTP server error: %v\n", err)
			errChan <- fmt.Errorf("HTTP server error: %w", err)
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

// getGinEngine 获取 Gin 引擎
func (e *Engine) getGinEngine() *gin.Engine {
	return e.ginEngine
}

// registerControllers 注册所有控制器路由
func (e *Engine) registerControllers() error {
	controllers := e.Controller.GetAll()
	for _, ctrl := range controllers {
		route := ctrl.GetRouter()
		if route == "" {
			continue
		}

		method, path, err := parseRoute(route)
		if err != nil {
			fmt.Printf("[WARNING] Controller %s has invalid route format: %v\n", ctrl.ControllerName(), err)
			continue
		}

		handler := ctrl.Handle
		e.registerRoute(method, path, handler)
		fmt.Printf("[DEBUG] Registered controller: %s -> %s %s\n", ctrl.ControllerName(), method, path)
	}

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
