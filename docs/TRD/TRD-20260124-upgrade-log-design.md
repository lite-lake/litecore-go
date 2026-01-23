# 日志增强设计方案 v2

**文档编号**: TRD-20260124
**版本**: v2
**创建日期**: 2026-01-24
**更新日期**: 2026-01-24
**项目**: litecore-go 框架
**状态**: 设计定稿

## 1. 背景与目标

### 1.1 当前问题
- 启动过程缺乏可见性，无法追踪组件初始化、依赖注入、路由注册等关键环节
- LoggerManager 初始化前无法使用结构化日志
- 没有启动耗时统计
- 配置文件加载路径等关键信息未记录

### 1.2 设计目标
- 提供完整的启动生命周期日志
- 支持启动阶段划分和耗时统计
- 优雅处理 LoggerManager 初始化前后的日志输出切换
- 记录组件注册、路由注册等关键节点
- 低侵入性设计，保持容器纯粹性
- 异步日志输出，不影响启动性能

## 2. 核心设计

### 2.1 启动阶段划分

```go
type StartupPhase int

const (
	PhaseConfig     StartupPhase = iota // 配置加载
	PhaseManagers                        // 管理器初始化
	PhaseInjection                       // 依赖注入
	PhaseRouter                          // 路由注册
	PhaseStartup                         // 组件启动
	PhaseRunning                         // 运行中
	PhaseShutdown                        // 关闭中
)
```

### 2.2 统一日志接口

**设计原则：始终使用 `logger.ILogger`，不创建独立的启动日志接口**

```go
type Engine struct {
	// ... 现有字段

	// 启动日志配置
	startupLogConfig *StartupLogConfig

	// 启动时间统计
	startupStartTime   time.Time
	phaseDurations     map[StartupPhase]time.Duration
	phaseStartTimes    map[StartupPhase]time.Time

	// 日志器（统一使用 logger.ILogger）
	logger      logger.ILogger
	isStartup   bool          // 标识是否处于启动阶段
	loggerMu    sync.RWMutex // 保护日志器的并发访问
}
```

### 2.3 日志切换时机

**关键：在 `e.Manager = builtInManagerContainer` 之后立即切换日志**

```go
func (e *Engine) Initialize() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.startupStartTime = time.Now()
	e.phaseStartTimes = make(map[StartupPhase]time.Time)
	e.phaseDurations = make(map[StartupPhase]time.Duration)

	// 1. 初始化前使用默认日志器
	e.logger = logger.NewDefaultLogger()
	e.isStartup = true

	// 2. 初始化内置组件
	e.phaseStartTimes[PhaseConfig] = time.Now()
	builtInManagerContainer, err := Initialize(e.builtinConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize builtin components: %w", err)
	}
	e.Manager = builtInManagerContainer
	e.phaseDurations[PhaseConfig] = time.Since(e.phaseStartTimes[PhaseConfig])

	// 3. 【关键】切换到结构化日志
	if loggerMgr, err := container.GetManager[loggermgr.ILoggerManager](e.Manager); err == nil {
		e.setLogger(loggerMgr.Ins())
		e.isStartup = false
		e.logger.Info("切换到结构化日志系统")
	} else {
		fmt.Fprintf(os.Stderr, "Failed to get logger manager: %v, using default logger\n", err)
	}

	// 4. 继续初始化流程...
}
```

### 2.4 异步日志输出

```go
// server/async_startup_logger.go

type AsyncStartupLogger struct {
	logger    logger.ILogger
	buffer    chan *StartupEvent
	wg        sync.WaitGroup
	closeCh   chan struct{}
	isClosed  bool
	mu        sync.Mutex
}

type StartupEvent struct {
	Phase    StartupPhase
	Message  string
	Fields   []logger.Field
}

func NewAsyncStartupLogger(baseLogger logger.ILogger, bufferSize int) *AsyncStartupLogger {
	l := &AsyncStartupLogger{
		logger:  baseLogger,
		buffer:  make(chan *StartupEvent, bufferSize),
		closeCh: make(chan struct{}),
	}
	l.wg.Add(1)
	go l.flushLoop()
	return l
}

func (l *AsyncStartupLogger) Log(phase StartupPhase, msg string, fields ...logger.Field) {
	l.mu.Lock()
	if l.isClosed {
		l.mu.Unlock()
		return
	}
	l.mu.Unlock()

	event := &StartupEvent{
		Phase:   phase,
		Message: msg,
		Fields:  fields,
	}

	select {
	case l.buffer <- event:
	default:
		l.logger.Warn("Startup log buffer full, dropping log", "msg", msg)
	}
}

func (l *AsyncStartupLogger) flushLoop() {
	defer l.wg.Done()
	for {
		select {
		case event := <-l.buffer:
			l.logger.Info(event.Message, event.Fields...)
		case <-l.closeCh:
			// 处理剩余事件
			for len(l.buffer) > 0 {
				event := <-l.buffer
				l.logger.Info(event.Message, event.Fields...)
			}
			return
		}
	}
}

func (l *AsyncStartupLogger) Stop() {
	l.mu.Lock()
	if !l.isClosed {
		l.isClosed = true
		close(l.closeCh)
	}
	l.mu.Unlock()
	l.wg.Wait()
}
```

### 2.5 启动日志配置

```go
type StartupLogConfig struct {
	Enabled bool   `yaml:"enabled"` // 是否启用启动日志
	Async   bool   `yaml:"async"`   // 是否异步输出（默认 true）
	Buffer  int    `yaml:"buffer"`  // 缓冲区大小（默认 100）
}

func DefaultStartupLogConfig() *StartupLogConfig {
	return &StartupLogConfig{
		Enabled: true,
		Async:   true,
		Buffer:  100,
	}
}
```

## 3. 日志点规划

### 3.1 Initialize() 流程日志

```
[配置加载] 开始初始化内置组件
[配置加载] 配置文件: configs/config.yaml
[配置加载] 配置驱动: yaml
[配置加载] 配置加载完成 (耗时 5ms)

[管理器初始化] 初始化完成: ConfigManager
[管理器初始化] 初始化完成: TelemetryManager
[管理器初始化] 初始化完成: LoggerManager
切换到结构化日志系统
[管理器初始化] 初始化完成: DatabaseManager
[管理器初始化] 初始化完成: CacheManager
[管理器初始化] 初始化完成: LockManager
[管理器初始化] 初始化完成: LimiterManager
[管理器初始化] 初始化完成: MQManager
[管理器初始化] 全部完成: 8 个管理器 (耗时 20ms)

[依赖注入] Repository 层: 1 个组件
[依赖注入]   MessageRepository: 注入完成
[依赖注入] Service 层: 4 个组件
[依赖注入]   MessageService: 注入完成
[依赖注入]   AuthService: 注入完成
[依赖注入]   SessionService: 注入完成
[依赖注入]   HTMLTemplateService: 注入完成
[依赖注入] Controller 层: 8 个组件
[依赖注入]   MessageListController: 注入完成
[依赖注入]   MessageCreateController: 注入完成
[依赖注入]   MessageAllController: 注入完成
[依赖注入]   MessageDeleteController: 注入完成
[依赖注入]   MessageStatusController: 注入完成
[依赖注入]   AdminAuthController: 注入完成
[依赖注入]   PageHomeController: 注入完成
[依赖注入]   PageAdminController: 注入完成
[依赖注入] Middleware 层: 6 个组件
[依赖注入]   RecoveryMiddleware: 注入完成
[依赖注入]   CORSMiddleware: 注入完成
[依赖注入]   RateLimiterMiddleware: 注入完成
[依赖注入]   AuthMiddleware: 注入完成
[依赖注入]   RequestLoggerMiddleware: 注入完成
[依赖注入]   SecurityHeadersMiddleware: 注入完成
[依赖注入] 完成: 19 个组件 (耗时 8ms)

[路由注册] 注册中间件: RecoveryMiddleware (全局)
[路由注册] 注册中间件: CORSMiddleware (全局)
[路由注册] 注册中间件: RateLimiterMiddleware (全局)
[路由注册] 注册路由: GET /api/messages -> MessageListController
[路由注册] 注册路由: POST /api/messages -> MessageCreateController
[路由注册] 注册路由: GET /api/admin/messages -> MessageAllController
[路由注册] 注册路由: POST /api/admin/messages/:id/status -> MessageStatusController
[路由注册] 注册路由: POST /api/admin/messages/:id/delete -> MessageDeleteController
[路由注册] 注册路由: POST /api/admin/login -> AdminAuthController
[路由注册] 注册路由: GET / -> PageHomeController
[路由注册] 注册路由: GET /admin.html -> PageAdminController
[路由注册] 完成: 7 个路由, 3 个中间件 (耗时 3ms)
```

### 3.2 Start() 流程日志

```
[组件启动] 开始启动各层组件
[组件启动] Manager 层: 8 个组件启动完成 (耗时 2ms)
[组件启动] Repository 层: 1 个组件启动完成 (耗时 1ms)
[组件启动] Service 层: 4 个组件启动完成 (耗时 1ms)
[组件启动] Middleware 层: 6 个组件启动完成 (耗时 0ms)
[组件启动] 全部启动完成 (总耗时 4ms)

HTTP server listening: 0.0.0.0:8080
启动完成 (总耗时 150ms)
```

### 3.3 Stop() 流程日志

```
收到关闭信号: interrupt
[关闭中] HTTP 服务器关闭...
[关闭中] Middleware 层: 6 个组件停止完成
[关闭中] Service 层: 4 个组件停止完成
[关闭中] Repository 层: 1 个组件停止完成
[关闭中] Manager 层: 8 个组件停止完成
[关闭中] 关闭完成 (耗时 50ms)
```

## 4. 实现细节

### 4.1 日志器管理

```go
func (e *Engine) setLogger(logger logger.ILogger) {
	e.loggerMu.Lock()
	defer e.loggerMu.Unlock()
	e.logger = logger
}

func (e *Engine) getLogger() logger.ILogger {
	e.loggerMu.RLock()
	defer e.loggerMu.RUnlock()
	return e.logger
}
```

### 4.2 日志输出方法

```go
func (e *Engine) logStartup(phase StartupPhase, msg string, fields ...logger.Field) {
	if !e.startupLogConfig.Enabled {
		return
	}

	if e.startupLogConfig.Async {
		e.asyncLogger.Log(phase, msg, fields...)
	} else {
		e.getLogger().Info(msg, fields...)
	}
}

func (e *Engine) logPhaseStart(phase StartupPhase, msg string, fields ...logger.Field) {
	e.phaseStartTimes[phase] = time.Now()
	e.logStartup(phase, msg, fields...)
}

func (e *Engine) logPhaseEnd(phase StartupPhase, msg string, extraFields ...logger.Field) {
	duration := time.Since(e.phaseStartTimes[phase])
	e.phaseDurations[phase] = duration

	fields := append(extraFields,
		logger.F("duration", duration.String()),
		logger.F("phase", phase.String()))
	e.logStartup(phase, msg, fields...)
}
```

### 4.3 容器注册日志（观察者模式）

**不侵入容器，在 Engine 层收集注册信息**

```go
func (e *Engine) logContainerInject(layer string, components []string) {
	for _, name := range components {
		e.logStartup(PhaseInjection, fmt.Sprintf("[%s 层] %s: 注入完成", layer, name))
	}
}

func (e *Engine) autoInject() error {
	e.logPhaseStart(PhaseInjection, "开始依赖注入")

	repoNames := e.logInjectRepositories()
	svcNames := e.logInjectServices()
	ctrlNames := e.logInjectControllers()
	mwNames := e.logInjectMiddlewares()

	e.logPhaseEnd(PhaseInjection, "依赖注入完成",
		logger.F("count", len(repoNames)+len(svcNames)+len(ctrlNames)+len(mwNames)))

	return nil
}
```

### 4.4 路由注册日志

```go
func (e *Engine) registerControllers() error {
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

		e.logStartup(PhaseRouter, "注册路由",
			logger.F("method", method),
			logger.F("path", path),
			logger.F("controller", ctrl.ControllerName()))
		registeredCount++
	}

	e.logStartup(PhaseRouter, "路由注册完成",
		logger.F("route_count", registeredCount),
		logger.F("controller_count", len(controllers)))

	return nil
}
```

### 4.5 组件启动日志

```go
func (e *Engine) startManagers() error {
	e.logPhaseStart(PhaseStartup, "开始启动 Manager 层")
	managers := e.Manager.GetAll()

	for _, mgr := range managers {
		if err := mgr.(common.IBaseManager).OnStart(); err != nil {
			return fmt.Errorf("failed to start manager %s: %w", mgr.(common.IBaseManager).ManagerName(), err)
		}
		e.logStartup(PhaseStartup, fmt.Sprintf("%s: 启动完成", mgr.(common.IBaseManager).ManagerName()))
	}

	e.logPhaseEnd(PhaseStartup, "Manager 层启动完成", logger.F("count", len(managers)))
	return nil
}
```

### 4.6 关闭日志

```go
func (e *Engine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.started {
		return nil
	}

	e.logStartup(PhaseShutdown, "HTTP 服务器关闭...")

	ctx, cancel := context.WithTimeout(context.Background(), e.shutdownTimeout)
	defer cancel()

	if e.httpServer != nil {
		if err := e.httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("HTTP server shutdown error: %w", err)
		}
	}

	e.logPhaseStart(PhaseShutdown, "开始停止各层组件")
	middlewareErrors := e.stopMiddlewares()
	serviceErrors := e.stopServices()
	repositoryErrors := e.stopRepositories()
	managerErrors := e.stopManagers()

	allErrors := make([]error, 0, len(middlewareErrors)+len(serviceErrors)+len(repositoryErrors)+len(managerErrors))
	allErrors = append(allErrors, middlewareErrors...)
	allErrors = append(allErrors, serviceErrors...)
	allErrors = append(allErrors, repositoryErrors...)
	allErrors = append(allErrors, managerErrors...)

	e.logPhaseEnd(PhaseShutdown, "关闭完成",
		logger.F("error_count", len(allErrors)),
		logger.F("total_duration", time.Since(e.startupStartTime).String()))

	if len(allErrors) > 0 {
		errorMessages := make([]string, len(allErrors))
		for i, err := range allErrors {
			errorMessages[i] = err.Error()
		}
		return fmt.Errorf("shutdown completed with %d error(s): %s", len(allErrors), strings.Join(errorMessages, "; "))
	}

	e.started = false
	return nil
}
```

## 5. 配置与扩展

### 5.1 配置项

```yaml
server:
  startup_log:
    enabled: true   # 是否启用启动日志
    async: true     # 是否异步输出
    buffer: 100     # 缓冲区大小
```

### 5.2 日志级别控制

- 临时日志（LoggerManager 初始化前）：固定输出到标准输出
- 结构化日志（LoggerManager 初始化后）：使用 LoggerManager 的配置

## 6. 文件结构

```
server/
  ├── engine.go                 # 改造，添加日志点和日志切换
  ├── builtin.go                # 改造，添加管理器初始化日志
  ├── lifecycle.go              # 改造，添加启动/关闭日志
  ├── async_startup_logger.go   # 新增，异步日志器
  └── startup_phase.go          # 新增，阶段定义和辅助方法

container/
  # 不需要修改，保持纯粹性
```

## 7. 关键改进点（相比 v1）

### 7.1 简化接口设计
- **v1**: IStartupLogger 接口包含 6 个方法（Phase、Info、Warn、Error、Progress、Done）
- **v2**: 统一使用 logger.ILogger，只添加辅助方法（logStartup、logPhaseStart、logPhaseEnd）

### 7.2 移除容器钩子
- **v1**: 在 TypedContainer.Register 中添加日志钩子，破坏容器纯粹性
- **v2**: 使用观察者模式，在 Engine 层收集注册信息，容器保持不变

### 7.3 统一日志接口
- **v1**: Engine 同时维护 startupLogger 和 logger 两个日志器
- **v2**: 始终使用 logger.ILogger，通过 setLogger 方法统一切换

### 7.4 异步日志输出
- **v1**: 同步日志输出，可能阻塞启动流程
- **v2**: 使用 channel 缓冲，异步输出日志，不影响启动性能

### 7.5 配置化控制
- **v1**: 配置项较少（enabled、timing、details）
- **v2**: 支持异步开关和缓冲区大小配置

## 8. 性能影响评估

- **启动耗时增加**: 预计 < 5ms（异步日志情况下）
- **内存占用**: 增加约 100KB（日志缓冲区）
- **IO 开销**: 异步输出，不阻塞启动流程

## 9. 潜在风险及应对

### 风险 1：LoggerManager 初始化失败
**应对**：
- 使用默认日志器兜底
- 记录到 stderr
- 不影响框架启动

### 风险 2：日志缓冲区满
**应对**：
- Warn 级别记录丢弃日志
- 可配置缓冲区大小
- 不影响主流程

### 风险 3：并发安全性
**应对**：
- Engine 字段使用 sync.RWMutex 保护
- AsyncStartupLogger 使用 channel 隔离
- 避免在日志回调中修改共享状态

## 10. 实施计划

### 阶段一：核心实现（优先级：高）
1. 实现 StartupPhase 枚举和辅助方法
2. 实现 AsyncStartupLogger
3. 改造 Engine，添加日志点
4. 实现日志切换机制

**预期工作量**：2-3 天

### 阶段二：日志点完善（优先级：中）
1. 在 builtin.go 中添加管理器初始化日志
2. 在 lifecycle.go 中添加启动/关闭日志
3. 在 engine.go 中添加注入和路由注册日志

**预期工作量**：2 天

### 阶段三：测试与优化（优先级：低）
1. 添加单元测试
2. 性能测试
3. 文档完善

**预期工作量**：2 天

**总工作量**：约 6-7 天

## 11. 后续优化

- [ ] 支持启动日志输出到文件
- [ ] 支持启动日志 JSON 格式
- [ ] 支持启动进度条（终端模式）
- [ ] 支持启动健康检查
