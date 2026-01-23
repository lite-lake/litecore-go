# 错误处理维度代码审查报告

**审查日期**: 2026-01-24
**审查维度**: 错误处理
**项目**: litecore-go

## 1. 概述

本次审查从错误检查、错误传递、错误信息、错误类型、Panic 处理、错误日志和用户反馈七个维度，对 litecore-go 项目的错误处理机制进行了深度分析。

总体评价：项目错误处理机制较为完善，大部分代码遵循了 Go 的错误处理最佳实践，使用 `fmt.Errorf` 和 `%w` 包装错误，错误信息清晰且包含上下文，有结构化日志和敏感信息脱敏机制。但也存在一些需要改进的地方，特别是 Panic 的使用不当、部分错误检查缺失以及用户反馈不够友好等问题。

---

## 2. 详细审查结果

### 2.1 错误检查

#### ✅ 优点

1. **大部分代码正确检查了所有返回的错误**
   - 核心业务代码如 `message_service.go` 中的所有数据库操作都正确处理了错误
   - Manager 层、Service 层、Repository 层的错误检查覆盖率很高

2. **使用 `go vet` 检查未发现问题**
   - 项目通过了 `go vet` 检查，没有明显的未检查错误问题

#### ⚠️ 问题

1. **忽略错误返回值**

**位置**: `manager/databasemgr/impl_base.go:50`

```go
sqlDB, _ := db.DB()
```

**问题**: 在创建基础数据库管理器时，`db.DB()` 可能返回错误，但被忽略。

**风险**: 如果数据库连接获取失败，后续操作将出现 `nil pointer` 错误，且错误信息不够明确。

**建议修改**:

```go
sqlDB, err := db.DB()
if err != nil {
    return nil, fmt.Errorf("failed to get *sql.DB: %w", err)
}
```

---

2. **消息队列中的错误静默处理**

**位置**: `manager/mqmgr/memory_impl.go:196,218`

```go
func() {
    defer recover()
    messageCh <- msg
}()
```

**问题**: 使用 `defer recover()` 但没有检查 recover 的返回值，导致 panic 后错误信息丢失。

**风险**: 无法诊断消息队列中的 panic 原因。

**建议修改**:

```go
func() {
    if err := recover(); err != nil {
        m.LoggerMgr.Ins().Error("发送消息到 channel 时发生 panic",
            "queue", q.name,
            "error", err)
    } else {
        messageCh <- msg
    }
}()
```

---

3. **部分控制器未验证错误类型**

**位置**: `samples/messageboard/internal/controllers/msg_create_controller.go:41,50`

```go
if err := ctx.ShouldBindJSON(&req); err != nil {
    c.LoggerMgr.Ins().Error("创建留言失败：参数绑定失败", "error", err)
    ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
    return
}

message, err := c.MessageService.CreateMessage(req.Nickname, req.Content)
if err != nil {
    c.LoggerMgr.Ins().Error("创建留言失败", "nickname", req.Nickname, "error", err)
    ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
    return
}
```

**问题**: 将所有错误都映射到 HTTP 400，没有区分验证错误、业务错误和系统错误。

**风险**: 客户端无法正确处理不同类型的错误。

**建议修改**: 根据错误类型返回不同的 HTTP 状态码：

```go
if err := ctx.ShouldBindJSON(&req); err != nil {
    c.LoggerMgr.Ins().Warn("创建留言失败：参数绑定失败", "error", err)
    ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, "请求参数格式错误"))
    return
}

message, err := c.MessageService.CreateMessage(req.Nickname, req.Content)
if err != nil {
    c.LoggerMgr.Ins().Error("创建留言失败", "nickname", req.Nickname, "error", err)
    // 根据错误类型判断状态码
    if errors.Is(err, gorm.ErrRecordNotFound) {
        ctx.JSON(common.HTTPStatusNotFound, dtos.ErrorResponse(common.HTTPStatusNotFound, "留言不存在"))
    } else {
        ctx.JSON(common.HTTPStatusInternalServerError, dtos.ErrorResponse(common.HTTPStatusInternalServerError, "服务器内部错误"))
    }
    return
}
```

---

### 2.2 错误传递

#### ✅ 优点

1. **正确使用错误包装**

**位置**: `server/engine.go` 等多处

```go
return fmt.Errorf("failed to initialize builtin components: %w", err)
return fmt.Errorf("auto inject failed: %w", err)
return fmt.Errorf("failed to start manager %s: %w", mgr.ManagerName(), err)
```

**优点**:
- 使用 `%w` 动词正确包装错误
- 保留错误链，支持 `errors.Is` 和 `errors.As`
- 错误信息包含足够的上下文

2. **错误链完整**

从 Controller → Service → Repository → Manager 的错误传递链清晰，每层都正确包装了错误。

---

#### ⚠️ 问题

1. **部分错误包装信息不够具体**

**位置**: `manager/loggermgr/factory.go:39,43`

```go
return nil, fmt.Errorf("failed to get logger.driver: %w", err)
return nil, fmt.Errorf("logger.driver: %w", err)
```

**问题**: 第二条错误信息没有说明失败原因，只是重复了配置项名称。

**建议修改**:

```go
return nil, fmt.Errorf("logger.driver configuration is missing or invalid: %w", err)
```

---

### 2.3 错误信息

#### ✅ 优点

1. **错误信息清晰且包含上下文**

**位置**: `server/lifecycle.go:50`

```go
return fmt.Errorf("failed to start manager %s: %w", mgr.ManagerName(), err)
```

**优点**:
- 包含失败的操作（"failed to start manager"）
- 包含具体的管理器名称
- 包装了底层错误

2. **敏感信息脱敏机制**

**位置**: `manager/databasemgr/impl_base.go:441-461`

```go
func sanitizeSQL(sql string) string {
    // 脱敏密码字段
    passwordPatterns := []string{
        `password\s*=\s*'[^']*'`,
        `password\s*=\s*"[^"]*"`,
        `pwd\s*=\s*'[^']*'`,
        ...
    }
    // 脱敏字符串值中的敏感字段
    sensitiveFields := []string{"password", "pwd", "token", "secret", "api_key"}
    ...
}
```

**优点**:
- 在记录 SQL 时自动脱敏敏感字段
- 使用正则表达式匹配常见模式
- 保护密码、token 等敏感信息不被记录到日志中

---

#### ⚠️ 问题

1. **部分错误信息可能泄露内部细节**

**位置**: `samples/messageboard/internal/controllers/msg_create_controller.go:51`

```go
ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
```

**问题**: 直接将 `err.Error()` 返回给客户端，可能包含数据库结构、内部逻辑等敏感信息。

**风险**:
- 数据库字段名暴露：`Error 1054: Unknown column 'xxx' in 'field list'`
- 内部服务名暴露：`failed to connect to database xxx`
- 堆栈信息（如果有的话）

**建议修改**: 定义用户友好的错误消息，而不是直接返回 `err.Error()`:

```go
// 在 errors 包中定义错误类型
type UserError struct {
    Code    string
    Message string
}

func (e *UserError) Error() string {
    return e.Message
}

// 在 Service 层返回 UserError
func (s *messageService) CreateMessage(nickname, content string) (*entities.Message, error) {
    if len(nickname) < 2 || len(nickname) > 20 {
        return nil, &UserError{
            Code:    "NICKNAME_INVALID",
            Message: "昵称长度必须在 2-20 个字符之间",
        }
    }
    ...
}

// 在 Controller 层处理
if err != nil {
    if userErr, ok := err.(*UserError); ok {
        ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, userErr.Message, "code", userErr.Code))
    } else {
        c.LoggerMgr.Ins().Error("创建留言失败", "error", err)
        ctx.JSON(common.HTTPStatusInternalServerError, dtos.ErrorResponse(common.HTTPStatusInternalServerError, "服务器内部错误"))
    }
    return
}
```

---

2. **错误信息缺少国际化支持**

**位置**: 多处硬编码的中文错误消息

```go
return errors.New("昵称长度必须在 2-20 个字符之间")
return errors.New("留言内容长度必须在 5-500 个字符之间")
```

**问题**: 所有错误消息都是硬编码的中文，无法支持多语言。

**建议修改**: 使用错误码 + 消息模板，支持国际化：

```go
// 定义错误码
const (
    ErrCodeNicknameTooShort = "NICKNAME_TOO_SHORT"
    ErrCodeNicknameTooLong   = "NICKNAME_TOO_LONG"
)

// 返回错误码和参数
return fmt.Errorf("validation failed: %s", ErrCodeNicknameTooShort)

// 在中间件或响应层根据 Accept-Language 头部翻译消息
```

---

### 2.4 错误类型

#### ✅ 优点

1. **自定义错误类型定义完善**

**位置**: `container/errors.go`

```go
type DependencyNotFoundError struct {
    InstanceName  string
    FieldName     string
    FieldType     reflect.Type
    ContainerType string
}

type CircularDependencyError struct {
    Cycle []string
}

type AmbiguousMatchError struct {
    InstanceName string
    FieldName    string
    FieldType    reflect.Type
    Candidates   []string
}
...
```

**优点**:
- 容器层有完整的自定义错误类型
- 错误类型语义清晰，便于调用方判断
- 每种错误都有专门的场景

2. **验证器自定义错误类型**

**位置**: `util/validator/validator.go:91-98`

```go
type ValidationError struct {
    Message string
    Errors  validator.ValidationErrors
}

func (ve *ValidationError) Error() string {
    return ve.Message
}
```

**优点**: 将验证错误包装为自定义类型，便于统一处理。

---

#### ⚠️ 问题

1. **业务层缺少统一错误类型体系**

**问题**: 当前项目只在容器层和验证器层定义了自定义错误类型，业务层（Service、Repository）直接返回 `error` 或使用 `errors.New` 创建错误。

**影响**:
- Controller 层无法区分业务错误和系统错误
- 错误处理逻辑重复
- 难以实现统一的错误码体系

**建议**: 定义统一的业务错误类型：

```go
// common/errors.go
package common

// BizError 业务错误
type BizError struct {
    Code    string // 错误码
    Message string // 用户可见的错误消息
    Err     error  // 原始错误（可选）
}

func (e *BizError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
    }
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 实现 errors.Unwrap 接口
func (e *BizError) Unwrap() error {
    return e.Err
}

// 预定义错误
var (
    ErrInvalidParams    = &BizError{Code: "INVALID_PARAMS", Message: "请求参数无效"}
    ErrNotFound         = &BizError{Code: "NOT_FOUND", Message: "资源不存在"}
    ErrPermissionDenied = &BizError{Code: "PERMISSION_DENIED", Message: "权限不足"}
    ErrInternalError    = &BizError{Code: "INTERNAL_ERROR", Message: "服务器内部错误"}
)

// NewBizError 创建业务错误
func NewBizError(code, message string, err error) *BizError {
    return &BizError{
        Code:    code,
        Message: message,
        Err:     err,
    }
}
```

---

### 2.5 Panic 处理

#### ⚠️ 严重问题

1. **不恰当的 Panic 使用**

**位置**: `manager/cachemgr/memory_impl.go:44-46`

```go
cache, err := ristretto.NewCache(&ristretto.Config[string, any]{...})
if err != nil {
    panic(fmt.Sprintf("failed to create ristretto cache: %v", err))
}
```

**问题**: 在正常初始化流程中使用 panic，导致程序直接崩溃，无法优雅降级。

**风险**:
- 服务启动失败无法恢复
- 生产环境中 panic 会直接终止进程
- 无法返回有意义的错误信息

**建议修改**:

```go
func NewCacheManagerMemoryImpl(defaultExpiration, cleanupInterval time.Duration) (ICacheManager, error) {
    cache, err := ristretto.NewCache(&ristretto.Config[string, any]{...})
    if err != nil {
        return nil, fmt.Errorf("failed to create ristretto cache: %w", err)
    }
    return &cacheManagerMemoryImpl{...}, nil
}
```

**影响范围**: 需要修改所有调用 `NewCacheManagerMemoryImpl` 的地方，检查返回的错误。

---

2. **依赖注入验证使用 Panic**

**位置**: `container/injector.go:52-58`

```go
if !fieldVal.CanInterface() || fieldVal.IsZero() || fieldVal.IsNil() {
    panic(&UninjectedFieldError{
        InstanceName: instanceName,
        FieldName:    field.Name,
        FieldType:    field.Type,
    })
}
```

**问题**: 在依赖注入验证阶段使用 panic，导致注入失败时程序崩溃。

**风险**:
- 启动时配置错误会导致服务完全无法启动
- 错误信息无法被上层处理
- 无法实现优雅降级

**建议修改**: 将 panic 改为返回错误，或者将 `UninjectedFieldError` 改为普通错误类型：

```go
// 选项 1: 返回错误
func verifyInjectTags(instance interface{}) error {
    val := reflect.ValueOf(instance)
    ...
    for i := 0; i < val.NumField(); i++ {
        ...
        if !fieldVal.CanInterface() || fieldVal.IsZero() || fieldVal.IsNil() {
            return &UninjectedFieldError{...}
        }
    }
    return nil
}

// 在调用处处理错误
if err := verifyInjectTags(svc); err != nil {
    return fmt.Errorf("injection verification failed: %w", err)
}
```

---

3. **ManagerContainer 未设置时 Panic**

**位置**: `container/service_container.go:57-59`

```go
func (s *ServiceContainer) InjectAll() error {
    if s.managerContainer == nil {
        panic(&ManagerContainerNotSetError{Layer: "Service"})
    }
    ...
}
```

**问题**: 使用 panic 而不是返回错误，违反了 Go 的错误处理惯例。

**建议修改**:

```go
func (s *ServiceContainer) InjectAll() error {
    if s.managerContainer == nil {
        return &ManagerContainerNotSetError{Layer: "Service"}
    }
    ...
}
```

类似问题也出现在 `container/repository_container.go:57`。

---

4. **CLI 生成器使用 Panic**

**位置**: `cli/generator/run.go:71-75`

```go
func MustRun(g *Generator, configPath string) {
    if err := g.Run(configPath); err != nil {
        panic(err)
    }
}
```

**问题**: `MustRun` 命名暗示这是可以 panic 的，但在生产环境应避免使用。

**建议修改**: 提供两个函数：
- `Run() error` - 正常使用，返回错误
- `MustRun()` - 仅在初始化代码中使用，明确标注会 panic

并添加注释说明：

```go
// MustRun 运行代码生成器，失败时 panic
// 注意：此函数仅用于初始化阶段，不在生产环境使用
func MustRun(g *Generator, configPath string) {
    if err := g.Run(configPath); err != nil {
        panic(err)
    }
}
```

---

#### ✅ 优点

1. **Recovery 中间件实现良好**

**位置**: `component/litemiddleware/recovery_middleware.go`

```go
defer func() {
    if err := recover(); err != nil {
        stack := debug.Stack()
        // 记录详细的 panic 信息
        fields := []interface{}{
            "panic", err,
            "method", method,
            "path", path,
            ...
        }
        if m.cfg.PrintStack != nil && *m.cfg.PrintStack {
            fields = append(fields, "stack", string(stack))
        }
        m.LoggerMgr.Ins().Error("PANIC recovered", fields...)
        ...
        c.Abort()
    }
}()
```

**优点**:
- 正确捕获 panic
- 记录完整的堆栈信息
- 返回友好的错误响应
- 可配置是否打印堆栈

---

### 2.6 错误日志

#### ✅ 优点

1. **统一使用结构化日志**

**位置**: 多处示例

```go
c.LoggerMgr.Ins().Error("创建留言失败", "nickname", nickname, "error", err)
s.LoggerMgr.Ins().Info("创建留言成功", "id", message.ID, "nickname", message.Nickname)
```

**优点**:
- 日志格式统一，便于解析
- 包含上下文信息
- 支持结构化字段

2. **日志级别使用正确**

- `Debug`: 开发调试信息（如 `s.LoggerMgr.Ins().Debug("获取已审核留言列表")`）
- `Info`: 正常业务流程（如 `s.LoggerMgr.Ins().Info("创建留言成功")`）
- `Warn`: 降级处理、慢查询（如数据库慢查询日志）
- `Error`: 业务错误、操作失败

3. **数据库操作日志完善**

**位置**: `manager/databasemgr/impl_base.go:364-399`

```go
if err != nil {
    logArgs := []any{
        "operation", operation,
        "table", db.Statement.Table,
        "error", err.Error(),
        "duration", duration,
    }
    if p.logSQL {
        logArgs = append(logArgs, "sql", sanitizeSQL(db.Statement.SQL.String()))
    }
    p.logger.Error("database operation failed", logArgs...)
} else {
    // 慢查询使用 Warn 级别
    if p.slowQueryThreshold > 0 && time.Since(start) >= p.slowQueryThreshold {
        p.logger.Warn("slow database query detected", logArgs...)
    } else {
        // 正常操作使用 Debug 级别
        p.logger.Debug("database operation success", ...)
    }
}
```

**优点**:
- 错误时记录详细上下文
- 慢查询单独告警
- SQL 自动脱敏
- 指标和日志联动

---

#### ⚠️ 问题

1. **部分日志缺少必要的上下文**

**位置**: `samples/messageboard/internal/services/auth_service.go:68`

```go
if err := s.SessionService.StoreSession(token, adminID); err != nil {
    s.LoggerMgr.Ins().Error("登录失败：创建会话失败", "error", err)
    return "", fmt.Errorf("failed to create session: %w", err)
}
```

**问题**: 日志中没有记录 adminID，难以追踪具体是哪个用户的登录失败。

**建议修改**:

```go
s.LoggerMgr.Ins().Error("登录失败：创建会话失败", "admin_id", adminID, "error", err)
```

---

2. **错误日志未包含请求 ID**

虽然 Recovery 中间件记录了 `requestID`，但部分业务日志中没有包含。

**建议**: 通过中间件在上下文中注入 Request ID，所有日志自动包含：

```go
// 在中间件中
ctx.Set("request_id", generateRequestID())

// 在日志中使用
c.LoggerMgr.Ins().Error("创建留言失败", "request_id", c.GetString("request_id"), "nickname", nickname, "error", err)
```

---

### 2.7 用户反馈

#### ✅ 优点

1. **Recovery 中间件提供友好的错误响应**

**位置**: `component/litemiddleware/recovery_middleware.go:134-141`

```go
if m.cfg.CustomErrorBody != nil && *m.cfg.CustomErrorBody {
    c.JSON(common.HTTPStatusInternalServerError, gin.H{
        "error": *m.cfg.ErrorMessage,
        "code":  *m.cfg.ErrorCode,
    })
} else {
    c.String(common.HTTPStatusInternalServerError, *m.cfg.ErrorMessage)
}
```

**优点**:
- 可以配置错误消息
- 提供错误码
- 不暴露内部堆栈

2. **限流中间件提供明确的用户提示**

**位置**: `component/litemiddleware/rate_limiter_middleware.go:155-158`

```go
c.JSON(http.StatusTooManyRequests, gin.H{
    "error": fmt.Sprintf("请求过于频繁，请 %v 后再试", *m.config.Window),
    "code":  "RATE_LIMIT_EXCEEDED",
})
c.Header("Retry-After", fmt.Sprintf("%d", int(m.config.Window.Seconds())))
```

**优点**:
- 错误消息清晰
- 提供 Retry-After 头
- 包含错误码

---

#### ⚠️ 问题

1. **控制器错误响应不够友好**

**位置**: `samples/messageboard/internal/controllers/msg_create_controller.go:41-51`

```go
if err := ctx.ShouldBindJSON(&req); err != nil {
    c.LoggerMgr.Ins().Error("创建留言失败：参数绑定失败", "error", err)
    ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
    return
}

message, err := c.MessageService.CreateMessage(req.Nickname, req.Content)
if err != nil {
    c.LoggerMgr.Ins().Error("创建留言失败", "nickname", req.Nickname, "error", err)
    ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
    return
}
```

**问题**:
- 直接返回 `err.Error()`，可能暴露内部细节
- 参数绑定错误和业务错误都返回 400，难以区分
- 没有提供错误码
- 没有提供解决建议

**建议修改**: 建立统一的错误响应中间件：

```go
// 统一错误响应
type ErrorResponse struct {
    Error   string `json:"error"`
    Code    string `json:"code"`
    Details any    `json:"details,omitempty"`
}

func HandleError(ctx *gin.Context, err error) {
    var response ErrorResponse

    // 1. 判断错误类型
    switch e := err.(type) {
    case *BizError:
        response = ErrorResponse{
            Error: e.Message,
            Code:  e.Code,
        }
        // 根据错误码映射 HTTP 状态码
        switch e.Code {
        case "INVALID_PARAMS":
            ctx.JSON(http.StatusBadRequest, response)
        case "NOT_FOUND":
            ctx.JSON(http.StatusNotFound, response)
        case "PERMISSION_DENIED":
            ctx.JSON(http.StatusForbidden, response)
        default:
            ctx.JSON(http.StatusInternalServerError, ErrorResponse{
                Error: "服务器内部错误",
                Code:  "INTERNAL_ERROR",
            })
        }
    case validator.ValidationErrors:
        // 验证错误
        var details []string
        for _, fieldErr := range e {
            details = append(details, fmt.Sprintf("%s: %s", fieldErr.Field(), fieldErr.Tag()))
        }
        response = ErrorResponse{
            Error:   "请求参数验证失败",
            Code:    "VALIDATION_ERROR",
            Details: details,
        }
        ctx.JSON(http.StatusBadRequest, response)
    default:
        // 其他错误
        ctx.JSON(http.StatusInternalServerError, ErrorResponse{
            Error: "服务器内部错误",
            Code:  "INTERNAL_ERROR",
        })
    }
}

// 在控制器中使用
if err := ctx.ShouldBindJSON(&req); err != nil {
    HandleError(ctx, err)
    return
}
```

---

2. **缺少错误文档和错误码列表**

**问题**: 项目中没有定义完整的错误码和对应的文档，客户端开发者难以正确处理错误。

**建议**: 创建错误码文档 `docs/errors.md`:

```markdown
# 错误码参考

## 系统错误

| 错误码 | HTTP 状态码 | 说明 | 解决方案 |
|--------|-----------|------|----------|
| INTERNAL_ERROR | 500 | 服务器内部错误 | 稍后重试 |
| SERVICE_UNAVAILABLE | 503 | 服务不可用 | 稍后重试 |

## 参数错误

| 错误码 | HTTP 状态码 | 说明 | 解决方案 |
|--------|-----------|------|----------|
| INVALID_PARAMS | 400 | 请求参数无效 | 检查请求参数格式 |
| VALIDATION_ERROR | 400 | 参数验证失败 | 根据详情修改请求参数 |

## 业务错误

| 错误码 | HTTP 状态码 | 说明 | 解决方案 |
|--------|-----------|------|----------|
| NICKNAME_TOO_SHORT | 400 | 昵称太短 | 昵称长度需 2-20 个字符 |
| NICKNAME_TOO_LONG | 400 | 昵称太长 | 昵称长度需 2-20 个字符 |
| CONTENT_TOO_SHORT | 400 | 留言内容太短 | 内容长度需 5-500 个字符 |
| CONTENT_TOO_LONG | 400 | 留言内容太长 | 内容长度需 5-500 个字符 |
| MESSAGE_NOT_FOUND | 404 | 留言不存在 | 检查留言 ID |

## 限流错误

| 错误码 | HTTP 状态码 | 说明 | 解决方案 |
|--------|-----------|------|----------|
| RATE_LIMIT_EXCEEDED | 429 | 请求过于频繁 | 根据 Retry-After 头等待后重试 |
```

---

## 3. 优先级建议

### P0 - 严重问题（必须修复）

1. **消除不恰当的 Panic 使用**
   - `manager/cachemgr/memory_impl.go:45` - 创建缓存失败应返回错误
   - `container/injector.go:53` - 依赖注入验证应返回错误
   - `container/service_container.go:58` - ManagerContainer 未设置应返回错误
   - `container/repository_container.go:57` - 同上

2. **修复忽略错误的问题**
   - `manager/databasemgr/impl_base.go:50` - 获取 sqlDB 应检查错误
   - `manager/mqmgr/memory_impl.go:196,218` - panic recover 应记录错误

---

### P1 - 高优先级问题（建议尽快修复）

1. **建立统一的错误响应机制**
   - 创建统一的错误处理中间件
   - 定义业务错误类型体系
   - 避免直接返回 `err.Error()`

2. **完善错误信息上下文**
   - 所有日志记录包含请求 ID
   - 关键操作日志包含必要的上下文（如 adminID）

---

### P2 - 中等优先级问题（建议修复）

1. **国际化支持**
   - 错误消息支持多语言
   - 使用错误码而非硬编码消息

2. **错误文档**
   - 建立完整的错误码列表
   - 提供错误处理指南

---

### P3 - 低优先级问题（可优化）

1. **错误信息规范化**
   - 统一错误消息格式
   - 改进部分错误消息的具体性

---

## 4. 修复示例

### 示例 1: 消除缓存管理器的 Panic

**修改前**:

```go
// manager/cachemgr/memory_impl.go
func NewCacheManagerMemoryImpl(defaultExpiration, cleanupInterval time.Duration) ICacheManager {
    cache, err := ristretto.NewCache(&ristretto.Config[string, any]{...})
    if err != nil {
        panic(fmt.Sprintf("failed to create ristretto cache: %v", err))
    }
    return &cacheManagerMemoryImpl{...}
}
```

**修改后**:

```go
// manager/cachemgr/memory_impl.go
func NewCacheManagerMemoryImpl(defaultExpiration, cleanupInterval time.Duration) (ICacheManager, error) {
    cache, err := ristretto.NewCache(&ristretto.Config[string, any]{...})
    if err != nil {
        return nil, fmt.Errorf("failed to create ristretto cache: %w", err)
    }
    return &cacheManagerMemoryImpl{...}, nil
}
```

---

### 示例 2: 统一的错误处理中间件

**新增**: `component/litemiddleware/error_handler_middleware.go`

```go
package litemiddleware

import (
    "errors"
    "github.com/gin-gonic/gin"
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
)

type errorResponse struct {
    Error   string `json:"error"`
    Code    string `json:"code"`
    Details any    `json:"details,omitempty"`
}

type ErrorHandlerConfig struct {
    Name           *string // 中间件名称
    Order          *int    // 执行顺序
    IncludeDetails *bool   // 是否包含错误详情（开发环境）
}

type errorHandlerMiddleware struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
    cfg       *ErrorHandlerConfig
}

func NewErrorHandlerMiddleware(config *ErrorHandlerConfig) common.IBaseMiddleware {
    // 配置处理...
    return &errorHandlerMiddleware{cfg: cfg}
}

func (m *errorHandlerMiddleware) Wrapper() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        // 检查是否有错误
        if len(c.Errors) == 0 {
            return
        }

        // 只处理第一个错误
        err := c.Errors.Last().Err

        // 构建响应
        response := m.buildErrorResponse(err)

        // 记录日志
        m.logError(c, err, response)

        // 返回响应
        c.JSON(m.getHTTPStatusCode(response), response)
    }
}

func (m *errorHandlerMiddleware) buildErrorResponse(err error) errorResponse {
    // 1. 检查是否是自定义业务错误
    var bizErr *BizError
    if errors.As(err, &bizErr) {
        return errorResponse{
            Error: bizErr.Message,
            Code:  bizErr.Code,
        }
    }

    // 2. 检查是否是验证错误
    var valErr validator.ValidationErrors
    if errors.As(err, &valErr) {
        details := make([]string, len(valErr))
        for i, fe := range valErr {
            details[i] = fmt.Sprintf("%s: %s", fe.Field(), fe.Tag())
        }
        return errorResponse{
            Error:   "请求参数验证失败",
            Code:    "VALIDATION_ERROR",
            Details: details,
        }
    }

    // 3. 默认返回系统错误
    return errorResponse{
        Error: "服务器内部错误",
        Code:  "INTERNAL_ERROR",
    }
}

func (m *errorHandlerMiddleware) getHTTPStatusCode(response errorResponse) int {
    switch response.Code {
    case "INVALID_PARAMS", "VALIDATION_ERROR":
        return common.HTTPStatusBadRequest
    case "NOT_FOUND":
        return common.HTTPStatusNotFound
    case "PERMISSION_DENIED":
        return common.HTTPStatusForbidden
    case "RATE_LIMIT_EXCEEDED":
        return 429
    default:
        return common.HTTPStatusInternalServerError
    }
}
```

---

### 示例 3: 业务错误类型体系

**新增**: `common/errors.go`

```go
package common

import (
    "fmt"
)

// BizError 业务错误
type BizError struct {
    Code    string // 错误码
    Message string // 用户可见的错误消息
    Err     error  // 原始错误（可选）
}

func (e *BizError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
    }
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 实现 errors.Unwrap 接口
func (e *BizError) Unwrap() error {
    return e.Err
}

// Is 实现 errors.Is 接口
func (e *BizError) Is(target error) bool {
    t, ok := target.(*BizError)
    return ok && e.Code == t.Code
}

// 预定义错误
var (
    ErrInvalidParams    = &BizError{Code: "INVALID_PARAMS", Message: "请求参数无效"}
    ErrNotFound         = &BizError{Code: "NOT_FOUND", Message: "资源不存在"}
    ErrPermissionDenied = &BizError{Code: "PERMISSION_DENIED", Message: "权限不足"}
    ErrInternalError    = &BizError{Code: "INTERNAL_ERROR", Message: "服务器内部错误"}
)

// NewBizError 创建业务错误
func NewBizError(code, message string, err error) *BizError {
    return &BizError{
        Code:    code,
        Message: message,
        Err:     err,
    }
}

// BizErrorIs 判断错误是否是指定的业务错误
func BizErrorIs(err error, code string) bool {
    var bizErr *BizError
    if errors.As(err, &bizErr) {
        return bizErr.Code == code
    }
    return false
}
```

---

## 5. 总结

### 优点

1. 错误包装机制完善，使用 `fmt.Errorf` 和 `%w` 正确包装错误
2. 错误日志使用结构化日志，包含丰富的上下文信息
3. 有敏感信息脱敏机制，保护密码、token 等敏感数据
4. 自定义错误类型定义完善，特别是在容器层
5. Recovery 中间件实现良好，能有效捕获和处理 panic

### 主要问题

1. **Panic 使用不当**：多处应该在启动阶段返回错误的地方使用了 panic
2. **忽略错误返回值**：少数地方忽略了重要函数的错误返回
3. **用户反馈不够友好**：直接返回 `err.Error()` 可能泄露内部信息
4. **缺少统一的错误类型体系**：业务层没有统一的错误类型
5. **错误信息未国际化**：所有错误消息都是硬编码的中文

### 建议行动计划

1. **短期（1-2 周）**：
   - 修复所有 P0 级别的 panic 问题
   - 修复忽略错误的问题
   - 添加请求 ID 到所有日志

2. **中期（1 个月）**：
   - 建立统一的业务错误类型体系
   - 实现统一的错误处理中间件
   - 改进控制器的错误响应

3. **长期（2-3 个月）**：
   - 实现错误信息国际化
   - 建立完整的错误码文档
   - 完善错误监控和告警机制

---

**审查人**: OpenCode
**审查时间**: 2026-01-24
