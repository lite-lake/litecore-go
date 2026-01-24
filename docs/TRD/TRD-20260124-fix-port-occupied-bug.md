# 端口占用导致启动失败但程序不退出 bug 修复技术需求文档

| 文档版本 | 日期 | 作者 |
|---------|------|------|
| 1.0 | 2026-01-24 | opencode |

## 1. 问题概述

### 1.1 问题描述

当服务端口被占用时，HTTP 服务器启动失败并返回错误到 `errChan`，但 `Start()` 方法从未读取该通道，导致：

1. **错误被静默忽略** - 端口占用错误仅记录到日志，不影响启动流程
2. **程序继续运行** - `Start()` 返回 nil，调用 `WaitForShutdown()` 进入无限等待
3. **用户体验极差** - 用户需要手动 Ctrl+C 终止程序，或者查看日志才能发现问题

### 1.2 受影响范围

| 场景 | 行为 | 后果 |
|------|------|------|
| 端口已被占用 | 错误记录到日志，程序不退出 | 用户误以为程序正常启动，但实际无法提供服务 |
| 端口权限不足 | 错误记录到日志，程序不退出 | 用户误以为程序正常启动，但实际无法提供服务 |
| 其他 HTTP 启动错误 | 错误记录到日志，程序不退出 | 用户误以为程序正常启动，但实际无法提供服务 |

### 1.3 根本原因

**代码逻辑缺陷：**

在 `server/engine.go:356-363` 启动 HTTP 服务器时：

```go
// Start() 方法中
errChan := make(chan error, 1)
go func() {
    if err := e.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        e.logger().Error("HTTP server error", "error", err)
        errChan <- fmt.Errorf("HTTP server error: %w", err)  // ← 错误发送到通道
        e.cancel()  // ← 取消上下文，但 WaitCtx() 不会退出
    }
}()

// ❌ 没有读取 errChan 的代码
// ❌ 直接返回 nil
e.started = true
return nil
```

在 `Run()` 方法中：

```go
func (e *Engine) Run() error {
    if err := e.Initialize(); err != nil {
        return err
    }

    if err := e.Start(); err != nil {  // ← Start() 在端口占用时返回 nil
        return err
    }

    e.WaitForShutdown()  // ← 调用 WaitCtx() 进入无限等待

    return nil
}
```

**问题链条：**

1. `httpServer.ListenAndServe()` 返回 `bind: address already in use` 错误
2. goroutine 将错误发送到 `errChan`，但无人读取
3. `Start()` 直接返回 `nil`，`e.started = true`
4. `Run()` 调用 `WaitForShutdown()` 进入信号等待
5. 程序永远不会退出，除非手动 Ctrl+C

## 2. 当前实现分析

### 2.1 Start() 方法实现

```go
// server/engine.go:303-373
func (e *Engine) Start() error {
    e.mu.Lock()
    defer e.mu.Unlock()

    if e.started {
        return fmt.Errorf("engine already started")
    }

    // 1-5. 启动各层组件...

    // 6. 启动 HTTP 服务器
    e.logger().Info("HTTP server listening", "addr", e.httpServer.Addr)

    errChan := make(chan error, 1)
    go func() {
        if err := e.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            e.logger().Error("HTTP server error", "error", err)
            errChan <- fmt.Errorf("HTTP server error: %w", err)
            e.cancel()
        }
    }()

    // ❌ 问题：没有读取 errChan 的代码

    e.started = true
    return nil
}
```

### 2.2 Run() 方法实现

```go
// server/engine.go:375-392
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
```

### 2.3 WaitForShutdown() 方法实现

```go
// server/engine.go:471-483
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
```

## 3. 修复方案

### 3.1 设计原则

1. **快速失败** - HTTP 服务器启动失败应立即返回错误，不继续运行
2. **超时保护** - 等待 HTTP 服务器启动成功的超时时间可配置
3. **向后兼容** - 不影响正常启动流程
4. **错误清晰** - 错误信息应明确指出是端口占用或其他原因

### 3.2 修复方案 A：启动超时检测（推荐）

#### 修改 Start() 方法

```go
// server/engine.go

// Start 启动引擎
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

    // 2. 启动所有 Repository
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

    // 5. 启动所有 Scheduler
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

    // 7. 启动 HTTP 服务器（新增启动超时检测）
    e.logger().Info("HTTP server listening", "addr", e.httpServer.Addr)

    errChan := make(chan error, 1)
    go func() {
        if err := e.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            e.logger().Error("HTTP server error", "error", err)
            errChan <- fmt.Errorf("HTTP server error: %w", err)
        }
    }()

    // 等待一小段时间，检查是否有立即发生的错误（如端口占用）
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
```

#### 修改 Run() 方法

```go
// Run 简化的启动方法
// 等价于 Initialize() + Start() + 等待信号
func (e *Engine) Run() error {
    if err := e.Initialize(); err != nil {
        return err
    }

    if err := e.Start(); err != nil {
        return err
    }

    e.WaitForShutdown()

    return nil
}
```

**优点：**
- 简单直接，只修改 `Start()` 方法
- 100ms 超时足够检测端口占用等立即错误
- 不影响正常启动流程
- 向后兼容

**缺点：**
- 100ms 延迟可能不够（极端情况下）
- 需要测试不同环境下的实际表现

### 3.3 修复方案 B：监听启动完成事件（备选）

#### 修改 HTTP 服务器配置

```go
// server/engine.go

type Engine struct {
    // ... 现有字段 ...

    // 新增：HTTP 服务器启动完成通道
    serverReadyChan chan struct{}
}

// Start 方法修改
func (e *Engine) Start() error {
    e.mu.Lock()
    defer e.mu.Unlock()

    if e.started {
        return fmt.Errorf("engine already started")
    }

    // ... 启动各层组件 ...

    // 初始化启动完成通道
    e.serverReadyChan = make(chan struct{}, 1)

    // 启动 HTTP 服务器
    e.logger().Info("HTTP server listening", "addr", e.httpServer.Addr)

    errChan := make(chan error, 1)
    go func() {
        listener, err := net.Listen("tcp", e.httpServer.Addr)
        if err != nil {
            e.logger().Error("Failed to listen on address", "addr", e.httpServer.Addr, "error", err)
            errChan <- fmt.Errorf("failed to listen on %s: %w", e.httpServer.Addr, err)
            return
        }

        e.serverReadyChan <- struct{}{}

        if err := e.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
            e.logger().Error("HTTP server error", "error", err)
            errChan <- fmt.Errorf("HTTP server error: %w", err)
        }
    }()

    // 等待服务器启动完成或失败
    select {
    case err := <-errChan:
        return fmt.Errorf("HTTP server failed to start: %w", err)
    case <-e.serverReadyChan:
        e.logger().Debug("HTTP server started successfully")
    }

    // ... 记录启动完成汇总 ...

    e.started = true
    return nil
}
```

**优点：**
- 准确检测服务器启动完成
- 无固定超时延迟
- 更可靠的错误检测

**缺点：**
- 需要手动创建 listener，修改较多代码
- 与 gin 原生 `ListenAndServe()` 接口不一致
- 可能影响现有功能（如自定义 listener）

### 3.4 方案选择

**推荐方案 A（启动超时检测）**，理由：

1. 修改量最小，风险最低
2. 端口占用错误会在调用 `ListenAndServe()` 时立即返回，100ms 足够检测
3. 保持与 gin 原生接口一致
4. 代码简洁，易于理解

## 4. 需要修改的文件清单

| 文件 | 修改内容 |
|------|---------|
| server/engine.go | 1. 修改 `Start()` 方法，增加启动错误检测逻辑<br>2. `Run()` 方法无需修改（已正确处理 Start() 的错误） |

## 5. 验证标准

### 5.1 功能验证

| 场景 | 预期行为 | 验证方法 |
|------|---------|---------|
| 端口未被占用 | 程序正常启动，监听端口 | `lsof -i:8080` 或 `netstat -an \| grep 8080` |
| 端口已被占用 | 程序启动失败，返回错误并退出 | `Run()` 返回非 nil 错误，进程退出 |
| 端口权限不足 | 程序启动失败，返回错误并退出 | 使用 1-1023 端口，无 root 权限时测试 |
| 启动后正常关闭 | Ctrl+C 触发优雅关闭 | 发送 SIGINT 信号，验证 `Stop()` 被调用 |

### 5.2 错误信息验证

```go
// 端口占用时的错误信息应包含
err.Error() 应包含：
- "address already in use" 或 "bind: address already in use"
- 端口号
- 明确的错误原因

示例：
"HTTP server failed to start: HTTP server error: listen tcp :8080: bind: address already in use"
```

### 5.3 代码质量验证

```bash
# 1. 构建成功
go build -o litecore ./...

# 2. 测试通过
go test ./...

# 3. 格式检查
go fmt ./...

# 4. 静态检查
go vet ./...
```

### 5.4 集成测试

```bash
# 测试 1：正常启动
./litecore &
PID=$!
sleep 2
# 验证进程正在运行且端口监听
kill $PID

# 测试 2：端口占用
# 先占用端口
nc -l 8080 > /dev/null 2>&1 &
LISTENER_PID=$!
# 尝试启动服务
./litecore
# 验证程序返回错误并退出（exit code 非 0）
echo $?
# 清理
kill $LISTENER_PID

# 测试 3：端口权限不足（需要 root 测试）
sudo ./litecore --port 80
# 验证程序返回错误并退出
```

## 6. 风险评估

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| 100ms 超时不够（慢速服务器） | 低 | 高 | 可通过配置参数调整超时时间 |
| 误判：服务器正常启动但超时 | 极低 | 中 | 100ms 远大于 `ListenAndServe()` 的启动时间 |
| 影响现有功能（如自定义启动流程） | 极低 | 中 | 仅修改 `Start()` 方法，不改变接口签名 |
| goroutine 泄漏 | 极低 | 低 | goroutine 在服务器关闭时自动退出 |

## 7. 预期收益

修复完成后，预期收益：

1. **快速失败** - 端口占用等错误立即被检测并返回，程序快速退出
2. **用户体验提升** - 用户能立即发现问题，无需手动终止程序
3. **错误信息清晰** - 错误信息明确指出端口占用等具体原因
4. **运维效率提升** - 部署脚本可以根据返回码判断启动成功/失败

## 8. 参考资料

- server/engine.go:303-392 - Start() 和 Run() 方法实现
- server/engine.go:471-483 - WaitForShutdown() 方法实现
- AGENTS.md - 项目架构和编码规范
- net/http 包文档 - ListenAndServe() 的错误处理说明
