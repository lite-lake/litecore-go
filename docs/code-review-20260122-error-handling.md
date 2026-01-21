# 错误处理与日志代码审查报告

**审查日期**: 2026-01-22
**审查范围**: /Users/kentzhu/Projects/lite-lake/litecore-go 全项目
**审查维度**: 错误处理、日志使用、敏感信息处理、日志级别规范

---

## 审查总结

本项目在错误处理和日志方面整体表现良好，遵循了 Go 语言的错误处理最佳实践，使用了结构化日志，并且通过依赖注入统一管理日志组件。然而，仍存在一些严重问题需要立即修复，特别是：

1. `util/logger/default_logger.go` 使用了禁止的标准库 `log.Printf` 和 `log.Fatal`
2. 部分 Repository 层直接返回 GORM 原始错误，缺少上下文信息
3. 个别地方记录了 token 等敏感信息
4. 日志级别使用不够规范（参数验证失败使用了 Warn）

整体评分：**7.5/10**

---

## 问题清单

### 🔴 严重问题

#### 1. 使用了禁止的日志方法

**问题描述**: `util/logger/default_logger.go` 中使用了标准库的 `log.Printf` 和 `log.Fatal`，违反了项目的日志使用规范。

**位置**: `util/logger/default_logger.go:16-34`

**影响**:
- 违反了项目日志使用规范（AGENTS.md 中明确禁止使用 `log.Fatal/Print/Printf/Println`）
- `log.Fatal` 会直接调用 `os.Exit(1)`，不会执行 defer 语句和资源清理，可能导致数据丢失或资源泄露
- 无法使用项目的日志管理器统一配置日志级别、格式和输出目标

**代码示例**:
```go
func (l *defaultLogger) Debug(msg string, args ...any) {
	log.Printf(l.prefix+"DEBUG: %s %v", msg, args)  // ❌ 使用了禁止的 log.Printf
}

func (l *defaultLogger) Fatal(msg string, args ...any) {
	log.Printf(l.prefix+"FATAL: %s %v", msg, args)
	args = append([]any{l.prefix}, args...)
	log.Fatal(args...)  // ❌ 使用了禁止的 log.Fatal，会直接 os.Exit
}
```

**建议**:
1. 直接移除 `defaultLogger`，因为项目中已经有完善的 `zap` 和 `none` 实现
2. 如果必须保留，应该使用标准输出（`fmt.Fprint`），而不是 `log.Fatal`

**修复建议**:
```go
func (l *defaultLogger) Debug(msg string, args ...any) {
	fmt.Printf(l.prefix+"DEBUG: %s %v\n", msg, args)
}

func (l *defaultLogger) Fatal(msg string, args ...any) {
	fmt.Printf(l.prefix+"FATAL: %s %v\n", msg, args)
	os.Exit(1)
}
```

---

#### 2. CLI 工具使用 fmt.Printf/Println 输出

**问题描述**: 多个 CLI 工具使用 `fmt.Printf` 和 `fmt.Println` 进行输出，虽然是 CLI 工具，但不符合项目规范。

**位置**:
- `samples/messageboard/cmd/genpasswd/main.go:14-79`
- `cli/main.go:35`
- `cli/generator/run.go:61`

**影响**:
- 违反了项目日志使用规范（AGENTS.md 明确禁止 `fmt.Printf/fmt.Println`，仅限开发调试）
- 如果需要在生产环境使用 CLI 工具，无法统一管理日志输出
- 注释中也明确标注了这是示例用途

**代码示例**:
```go
// samples/messageboard/cmd/genpasswd/main.go
fmt.Println("=== 留言板管理员密码生成工具 ===")
fmt.Printf("加密后的密码: %s\n", hashedPassword)

// cli/main.go
fmt.Printf("litecore-generate version %s\n", version)

// cli/generator/run.go
fmt.Printf("成功生成容器代码到 %s\n", absOutputDir)
```

**建议**:
1. 对于 CLI 工具，可以使用 `fmt.Fprint` 输出到标准输出/错误输出
2. 或者引入一个专门的 CLI 输出组件，统一管理 CLI 工具的输出格式
3. 在 AGENTS.md 中明确 CLI 工具的输出规范

**修复建议**:
```go
// 使用 fmt.Fprint 输出
fmt.Fprintln(os.Stdout, "=== 留言板管理员密码生成工具 ===")
fmt.Fprintf(os.Stdout, "加密后的密码: %s\n", hashedPassword)
```

---

#### 3. Repository 层直接返回 GORM 原始错误

**问题描述**: Repository 层的多个方法直接返回 GORM 的原始错误，没有包装或添加上下文信息。

**位置**:
- `samples/messageboard/internal/repositories/message_repository.go:52-55`
- `samples/messageboard/internal/repositories/message_repository.go:62-66`
- `samples/messageboard/internal/repositories/message_repository.go:71-73`
- `samples/messageboard/internal/repositories/message_repository.go:77-80`
- `samples/messageboard/internal/repositories/message_repository.go:84`
- `samples/messageboard/internal/repositories/message_repository.go:90-93`

**影响**:
- 错误信息缺少上下文，难以定位问题
- Service 层收到的错误信息可能包含数据库敏感信息
- 错误信息不够友好，不利于用户理解

**代码示例**:
```go
func (r *messageRepository) GetByID(id uint) (*entities.Message, error) {
	db := r.Manager.DB()
	var message entities.Message
	err := db.First(&message, id).Error  // ❌ 直接返回 GORM 错误
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) Delete(id uint) error {
	db := r.Manager.DB()
	return db.Delete(&entities.Message{}, id).Error  // ❌ 直接返回 GORM 错误
}
```

**建议**:
1. 在 Repository 层包装错误，添加操作类型和上下文信息
2. 使用自定义错误类型，区分数据库错误、记录不存在等场景
3. 隐藏数据库底层细节，提供清晰的错误信息

**修复建议**:
```go
func (r *messageRepository) GetByID(id uint) (*entities.Message, error) {
	db := r.Manager.DB()
	var message entities.Message
	err := db.First(&message, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("留言不存在: id=%d", id)
		}
		return nil, fmt.Errorf("查询留言失败: id=%d, error=%w", id, err)
	}
	return &message, nil
}

func (r *messageRepository) Delete(id uint) error {
	db := r.Manager.DB()
	result := db.Delete(&entities.Message{}, id)
	if result.Error != nil {
		return fmt.Errorf("删除留言失败: id=%d, error=%w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("留言不存在: id=%d", id)
	}
	return nil
}
```

---

#### 4. Service 层错误未包装

**问题描述**: `GetStatistics` 方法中有多个错误返回点，但都直接返回原始错误，没有包装或添加上下文。

**位置**: `samples/messageboard/internal/services/message_service.go:199-211`

**影响**:
- 错误信息缺少上下文，无法确定是哪个状态统计失败
- Service 层没有提供统一的错误处理

**代码示例**:
```go
func (s *messageService) GetStatistics() (map[string]int64, error) {
	pendingCount, err := s.Repository.CountByStatus("pending")
	if err != nil {
		return nil, err  // ❌ 没有包装错误
	}

	approvedCount, err := s.Repository.CountByStatus("approved")
	if err != nil {
		return nil, err  // ❌ 没有包装错误
	}

	rejectedCount, err := s.Repository.CountByStatus("rejected")
	if err != nil {
		return nil, err  // ❌ 没有包装错误
	}

	// ...
}
```

**建议**:
1. 包装错误，添加状态类型信息
2. 或者使用统一的方法包装所有错误

**修复建议**:
```go
func (s *messageService) GetStatistics() (map[string]int64, error) {
	pendingCount, err := s.Repository.CountByStatus("pending")
	if err != nil {
		return nil, fmt.Errorf("统计留言数量失败: status=pending, error=%w", err)
	}

	approvedCount, err := s.Repository.CountByStatus("approved")
	if err != nil {
		return nil, fmt.Errorf("统计留言数量失败: status=approved, error=%w", err)
	}

	rejectedCount, err := s.Repository.CountByStatus("rejected")
	if err != nil {
		return nil, fmt.Errorf("统计留言数量失败: status=rejected, error=%w", err)
	}

	// ...
}
```

---

### 🟡 中等问题

#### 5. 日志级别使用不当 - 参数验证失败使用 Warn

**问题描述**: 参数验证失败时使用了 `Warn` 级别，但根据日志级别规范，参数验证失败应该使用 `Debug` 或 `Info`。

**位置**:
- `samples/messageboard/internal/services/message_service.go:52, 58`
- `samples/messageboard/internal/controllers/msg_create_controller.go:40`

**影响**:
- 滥用 `Warn` 级别会导致日志噪音
- `Warn` 应该用于降级处理、慢查询、重试等场景
- 参数验证失败是正常的业务逻辑，不应该作为警告

**代码示例**:
```go
func (s *messageService) CreateMessage(nickname, content string) (*entities.Message, error) {
	if len(nickname) < 2 || len(nickname) > 20 {
		if s.Logger != nil {
			s.Logger.Warn("创建留言失败：昵称长度不符合要求", "nickname_length", len(nickname))  // ❌ 应该使用 Debug
		}
		return nil, errors.New("昵称长度必须在 2-20 个字符之间")
	}
	// ...
}
```

**建议**:
1. 参数验证失败使用 `Debug` 级别
2. 如果需要监控验证失败的频率，可以使用 `Info` 级别

**修复建议**:
```go
func (s *messageService) CreateMessage(nickname, content string) (*entities.Message, error) {
	if len(nickname) < 2 || len(nickname) > 20 {
		if s.Logger != nil {
			s.Logger.Debug("创建留言失败：昵称长度不符合要求", "nickname_length", len(nickname))  // ✅ 使用 Debug
		}
		return nil, errors.New("昵称长度必须在 2-20 个字符之间")
	}
	// ...
}
```

---

#### 6. 日志级别使用不当 - 密码错误使用 Warn

**问题描述**: 密码验证失败时使用了 `Warn` 级别，应该使用 `Debug` 级别。

**位置**: `samples/messageboard/internal/services/auth_service.go:60`

**影响**:
- 密码错误是正常的安全场景，使用 `Warn` 会产生大量噪音
- `Warn` 应该用于需要关注的异常情况

**代码示例**:
```go
func (s *authService) Login(password string) (string, error) {
	if !s.VerifyPassword(password) {
		if s.Logger != nil {
			s.Logger.Warn("登录失败：密码错误")  // ❌ 应该使用 Debug
		}
		return "", fmt.Errorf("invalid password")
	}
	// ...
}
```

**建议**:
使用 `Debug` 级别记录密码验证失败。

**修复建议**:
```go
func (s *authService) Login(password string) (string, error) {
	if !s.VerifyPassword(password) {
		if s.Logger != nil {
			s.Logger.Debug("登录失败：密码错误")  // ✅ 使用 Debug
		}
		return "", fmt.Errorf("invalid password")
	}
	// ...
}
```

---

#### 7. 记录敏感信息 - Token

**问题描述**: 登录和退出登录时记录了 token，这是敏感信息，可能导致安全问题。

**位置**:
- `samples/messageboard/internal/services/auth_service.go:74`
- `samples/messageboard/internal/services/auth_service.go:82`

**影响**:
- 可能导致 token 泄露
- 违反了安全最佳实践

**代码示例**:
```go
func (s *authService) Login(password string) (string, error) {
	// ...
	if s.Logger != nil {
		s.Logger.Info("登录成功", "token", token)  // ❌ 记录了敏感的 token
	}
	return token, nil
}

func (s *authService) Logout(token string) error {
	if s.Logger != nil {
		s.Logger.Info("退出登录", "token", token)  // ❌ 记录了敏感的 token
	}
	return s.SessionService.DeleteSession(token)
}
```

**建议**:
1. 不记录完整的 token
2. 只记录 token 的前几位或脱敏后的信息
3. 或者只记录操作成功/失败，不记录 token

**修复建议**:
```go
func (s *authService) Login(password string) (string, error) {
	// ...
	if s.Logger != nil {
		// 方案1：只记录部分信息
		tokenPrefix := ""
		if len(token) > 8 {
			tokenPrefix = token[:8] + "..."
		}
		s.Logger.Info("登录成功", "token_prefix", tokenPrefix)

		// 方案2：不记录 token
		// s.Logger.Info("登录成功")
	}
	return token, nil
}

func (s *authService) Logout(token string) error {
	if s.Logger != nil {
		// 方案1：只记录部分信息
		tokenPrefix := ""
		if len(token) > 8 {
			tokenPrefix = token[:8] + "..."
		}
		s.Logger.Info("退出登录", "token_prefix", tokenPrefix)

		// 方案2：不记录 token
		// s.Logger.Info("退出登录")
	}
	return s.SessionService.DeleteSession(token)
}
```

---

#### 8. 错误信息不够具体 - 认证失败

**问题描述**: 认证失败时返回的错误信息不够具体，无法区分是 token 格式错误还是 token 无效。

**位置**: `samples/messageboard/internal/middlewares/auth_middleware.go:47-76`

**影响**:
- 客户端无法根据错误类型进行不同的处理
- 调试困难

**代码示例**:
```go
authHeader := c.GetHeader("Authorization")
if authHeader == "" {
	c.JSON(common.HTTPStatusUnauthorized, gin.H{
		"code":    common.HTTPStatusUnauthorized,
		"message": "未提供认证令牌",
	})
	c.Abort()
	return
}

parts := strings.SplitN(authHeader, " ", 2)
if len(parts) != 2 || parts[0] != "Bearer" {
	c.JSON(common.HTTPStatusUnauthorized, gin.H{
		"code":    common.HTTPStatusUnauthorized,
		"message": "认证令牌格式错误",
	})
	c.Abort()
	return
}

token := parts[1]

session, err := m.AuthService.ValidateToken(token)
if err != nil {
	c.JSON(common.HTTPStatusUnauthorized, gin.H{
		"code":    common.HTTPStatusUnauthorized,
		"message": "认证令牌无效或已过期",  // ❌ 错误信息不够具体
	})
	c.Abort()
	return
}
```

**建议**:
根据错误类型返回不同的错误信息，或者使用错误代码。

**修复建议**:
```go
token := parts[1]

session, err := m.AuthService.ValidateToken(token)
if err != nil {
	// 根据错误类型返回不同的错误信息
	errMsg := "认证失败"
	if errors.Is(err, common.ErrTokenExpired) {
		errMsg = "认证令牌已过期"
	} else if errors.Is(err, common.ErrTokenInvalid) {
		errMsg = "认证令牌无效"
	} else {
		errMsg = "认证失败"
	}

	c.JSON(common.HTTPStatusUnauthorized, gin.H{
		"code":    common.HTTPStatusUnauthorized,
		"message": errMsg,
	})
	c.Abort()
	return
}
```

---

#### 9. Controller 层错误日志使用了 Error 级别

**问题描述**: Controller 层对于参数验证失败等业务错误使用了 `Error` 级别，应该根据错误严重程度选择合适的级别。

**位置**:
- `samples/messageboard/internal/controllers/msg_create_controller.go:40`
- `samples/messageboard/internal/controllers/msg_delete_controller.go:42`

**影响**:
- 混淆了真正的系统错误和业务错误
- 可能导致日志监控误报

**代码示例**:
```go
func (c *msgCreateControllerImpl) Handle(ctx *gin.Context) {
	var req dtos.CreateMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		if c.Logger != nil {
			c.Logger.Error("创建留言失败：参数绑定失败", "error", err)  // ❌ 应该使用 Warn 或 Debug
		}
		ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
		return
	}
	// ...
}
```

**建议**:
根据错误类型选择合适的日志级别：
- 参数验证失败：`Debug` 或 `Warn`
- 业务逻辑错误：`Warn`
- 系统错误：`Error`

**修复建议**:
```go
func (c *msgCreateControllerImpl) Handle(ctx *gin.Context) {
	var req dtos.CreateMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		if c.Logger != nil {
			c.Logger.Warn("创建留言失败：参数绑定失败", "error", err)  // ✅ 使用 Warn
		}
		ctx.JSON(common.HTTPStatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
		return
	}
	// ...
}
```

---

### 🟢 轻微问题

#### 10. 重复的 Logger nil 检查

**问题描述**: 在每个使用 Logger 的地方都进行了 `if s.Logger != nil` 检查，代码重复且不够优雅。

**位置**: 多个 Service、Controller、Middleware 文件

**影响**:
- 代码冗余
- 可读性降低

**代码示例**:
```go
func (s *messageService) CreateMessage(nickname, content string) (*entities.Message, error) {
	if len(nickname) < 2 || len(nickname) > 20 {
		if s.Logger != nil {  // ❌ 重复的 nil 检查
			s.Logger.Warn("创建留言失败：昵称长度不符合要求", "nickname_length", len(nickname))
		}
		return nil, errors.New("昵称长度必须在 2-20 个字符之间")
	}

	if err := s.Repository.Create(message); err != nil {
		if s.Logger != nil {  // ❌ 重复的 nil 检查
			s.Logger.Error("创建留言失败", "nickname", nickname, "error", err)
		}
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	if s.Logger != nil {  // ❌ 重复的 nil 检查
		s.Logger.Info("创建留言成功", "id", message.ID, "nickname", message.Nickname, "status", message.Status)
	}

	return message, nil
}
```

**建议**:
1. 使用 `initLogger()` 方法初始化 logger（AGENTS.md 中推荐的模式）
2. 或者使用空对象模式（NoOpLogger）

**修复建议**:
```go
type messageService struct {
	Config     common.IBaseConfigProvider      `inject:""`
	Repository repositories.IMessageRepository `inject:""`
	Logger     logger.ILogger                  `inject:""`
	logger     logger.ILogger  // 内部使用的 logger
}

// initLogger 初始化 logger（遵循 AGENTS.md 推荐的模式）
func (s *messageService) initLogger() {
	if s.Logger != nil {
		s.logger = s.Logger
	} else {
		s.logger = logger.NewNoOpLogger()
	}
}

func (s *messageService) CreateMessage(nickname, content string) (*entities.Message, error) {
	s.initLogger()  // 调用一次即可

	if len(nickname) < 2 || len(nickname) > 20 {
		s.logger.Warn("创建留言失败：昵称长度不符合要求", "nickname_length", len(nickname))  // ✅ 无需 nil 检查
		return nil, errors.New("昵称长度必须在 2-20 个字符之间")
	}

	if err := s.Repository.Create(message); err != nil {
		s.logger.Error("创建留言失败", "nickname", nickname, "error", err)  // ✅ 无需 nil 检查
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	s.logger.Info("创建留言成功", "id", message.ID, "nickname", message.Nickname, "status", message.Status)  // ✅ 无需 nil 检查

	return message, nil
}
```

---

#### 11. 日志中的重复字段

**问题描述**: 在某些日志中，有些字段可能重复或冗余。

**位置**: 多处

**影响**:
- 日志体积增大
- 查询效率降低

**建议**:
检查日志字段是否必要，避免冗余字段。

---

#### 12. RecoveryMiddleware 中的 panic 记录使用了 Error 级别

**问题描述**: panic 恢复后使用了 `Error` 级别记录，可能需要更严重的级别（如 `Fatal`）或保持 `Error`。

**位置**: `component/middleware/recovery_middleware.go:53`

**影响**:
- 可能需要更高级别来引起注意

**建议**:
保持当前实现即可，`Error` 级别已经足够。如果需要更严重，可以改为 `Fatal`，但要注意 `Fatal` 可能会导致程序退出。

---

#### 13. analyzer.go 中错误处理缺失上下文

**问题描述**: `findFactoryFunc` 方法中解析失败直接返回 `nil`，没有记录错误或返回错误信息。

**位置**: `cli/analyzer/analyzer.go:252-274`

**影响**:
- 调试困难，不知道为什么找不到工厂函数

**代码示例**:
```go
func (a *Analyzer) findFactoryFunc(filename, typeName string) *ast.FuncDecl {
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil  // ❌ 没有记录错误
	}

	var found *ast.FuncDecl

	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Name.Name == "New"+typeName {
				found = fn
				return false
			}
		}
		return true
	})

	return found
}
```

**建议**:
1. 添加日志记录
2. 或者在调用处处理错误

**修复建议**:
```go
func (a *Analyzer) findFactoryFunc(filename, typeName string) *ast.FuncDecl {
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		// 添加日志记录（如果项目中支持）
		fmt.Fprintf(os.Stderr, "警告: 解析文件失败 %s: %v\n", filename, err)
		return nil
	}

	var found *ast.FuncDecl

	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Name.Name == "New"+typeName {
				found = fn
				return false
			}
		}
		return true
	})

	return found
}
```

---

## 优秀实践

### ✅ 1. 良好的错误包装模式

项目中大部分地方都使用了 `%w` 进行错误包装，这是 Go 1.13+ 推荐的错误处理方式。

**示例**:
```go
// server/engine.go:102
return fmt.Errorf("failed to initialize builtin components: %w", err)

// component/manager/databasemgr/mysql_impl.go:26
return nil, fmt.Errorf("invalid mysql config: %w", err)
```

**优点**:
- 保留了原始错误信息
- 支持错误链追踪（`errors.Is` 和 `errors.As`）
- 错误信息层次清晰

---

### ✅ 2. 结构化日志

项目使用了结构化日志（zap），日志信息以键值对形式记录，便于查询和分析。

**示例**:
```go
// component/middleware/recovery_middleware.go:53-65
m.Logger.Error(
	"PANIC recovered",
	"panic", err,
	"method", method,
	"path", path,
	"query", query,
	"ip", clientIP,
	"userAgent", userAgent,
	"requestID", requestID,
	"timestamp", time.Now().Format(time.RFC3339Nano),
	"stack", string(stack),
)
```

**优点**:
- 结构化日志易于解析
- 便于日志查询和分析
- 支持日志聚合和监控

---

### ✅ 3. 依赖注入日志管理器

项目通过依赖注入的方式统一管理日志组件，符合项目的架构设计。

**示例**:
```go
// samples/messageboard/internal/services/auth_service.go
type authService struct {
	Config         common.IBaseConfigProvider `inject:""`
	SessionService ISessionService            `inject:""`
	Logger         logger.ILogger             `inject:""`
}
```

**优点**:
- 解耦日志组件
- 易于测试（可以注入 Mock Logger）
- 统一管理日志配置

---

### ✅ 4. SQL 脱敏

数据库日志中实现了 SQL 脱敏功能，避免敏感信息泄露。

**示例**:
```go
// component/manager/databasemgr/impl_base.go:419-462
func sanitizeSQL(sql string) string {
	// 脱敏密码参数（常见模式）
	passwordPatterns := []string{
		`password\s*=\s*'[^']*'`,
		`password\s*=\s*"[^"]*"`,
		`pwd\s*=\s*'[^']*'`,
		`pwd\s*=\s*"[^"]*"`,
		`token\s*=\s*'[^']*'`,
		`token\s*=\s*"[^"]*"`,
		// ...
	}

	for _, pattern := range passwordPatterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		sql = re.ReplaceAllString(sql, "***")
	}

	return strings.TrimSpace(sql)
}
```

**优点**:
- 防止敏感信息泄露
- 符合安全合规要求
- 日志信息仍保留核心内容

---

### ✅ 5. Panic 恢复中间件

项目中实现了 panic 恢复中间件，防止程序崩溃。

**示例**:
```go
// component/middleware/recovery_middleware.go:36-76
func (m *RecoveryMiddleware) Wrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()

				// 记录 panic 信息
				if m.Logger != nil {
					m.Logger.Error(
						"PANIC recovered",
						"panic", err,
						"stack", string(stack),
					)
				}

				// 返回友好错误
				c.JSON(common.HTTPStatusInternalServerError, gin.H{
					"error": "内部服务器错误",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
```

**优点**:
- 防止程序崩溃
- 记录 panic 信息便于调试
- 返回友好错误信息

---

### ✅ 6. 自定义错误类型

项目中定义了多个自定义错误类型，便于错误处理和区分。

**示例**:
```go
// container/errors.go:9-20
type DependencyNotFoundError struct {
	InstanceName  string
	FieldName     string
	FieldType     reflect.Type
	ContainerType string
}

func (e *DependencyNotFoundError) Error() string {
	return fmt.Sprintf("dependency not found for %s.%s: need type %s from %s container",
		e.InstanceName, e.FieldName, e.FieldType, e.ContainerType)
}
```

**优点**:
- 错误类型清晰
- 支持错误类型判断
- 错误信息具体

---

### ✅ 7. 慢查询日志

数据库操作中实现了慢查询检测和日志记录。

**示例**:
```go
// component/manager/databasemgr/impl_base.go:382-393
if p.slowQueryThreshold > 0 && time.Since(start) >= p.slowQueryThreshold {
	logArgs := []any{
		"operation", operation,
		"table", db.Statement.Table,
		"duration", duration,
		"threshold", p.slowQueryThreshold.Seconds(),
	}
	if p.logSQL {
		logArgs = append(logArgs, "sql", sanitizeSQL(db.Statement.SQL.String()))
	}
	p.logger.Warn("slow database query detected", logArgs...)
}
```

**优点**:
- 帮助发现性能问题
- 符合日志级别规范（慢查询使用 Warn）
- 包含详细的性能数据

---

### ✅ 8. 请求日志中间件

项目中实现了请求日志中间件，记录所有请求的关键信息。

**示例**:
```go
// component/middleware/request_logger_middleware.go:56-78
if len(c.Errors) > 0 {
	for _, e := range c.Errors {
		m.Logger.Error("请求处理错误",
			"request_id", requestID,
			"method", method,
			"path", path,
			"client_ip", clientIP,
			"status", status,
			"latency", latency,
			"error", e.Error(),
		)
	}
} else {
	m.Logger.Info("请求处理完成",
		"request_id", requestID,
		"method", method,
		"path", path,
		"client_ip", clientIP,
		"status", status,
		"latency", latency,
	)
}
```

**优点**:
- 完整的请求追踪
- 包含请求 ID、耗时等关键信息
- 区分成功和失败请求

---

### ✅ 9. 日志级别使用正确 - 慢查询

慢查询正确地使用了 `Warn` 级别。

**示例**:
```go
// component/manager/databasemgr/impl_base.go:393
p.logger.Warn("slow database query detected", logArgs...)
```

**优点**:
- 符合日志级别规范（Warn: 降级处理、慢查询、重试）

---

### ✅ 10. 日志级别使用正确 - 调试信息

数据库操作成功使用了 `Debug` 级别。

**示例**:
```go
// component/manager/databasemgr/impl_base.go:395-401
p.logger.Debug("database operation success",
	"operation", operation,
	"table", db.Statement.Table,
	"duration", duration,
)
```

**优点**:
- 符合日志级别规范（Debug: 开发调试信息）
- 避免生产环境日志噪音

---

## 改进建议

### 🎯 短期改进（1-2 周）

#### 1. 修复 defaultLogger

**优先级**: 高

**行动项**:
1. 移除 `util/logger/default_logger.go` 或改用 `fmt.Fprint` 替代 `log.Printf/log.Fatal`
2. 确保所有地方使用 `ILoggerManager` 而不是标准库的 `log` 包

---

#### 2. 统一 Repository 层错误处理

**优先级**: 高

**行动项**:
1. 在所有 Repository 方法中包装错误
2. 添加上下文信息（操作类型、参数等）
3. 区分"记录不存在"和其他错误

**示例**:
```go
// 创建统一的错误包装函数
func wrapDBError(operation string, err error, context ...string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%s: 记录不存在 %v", operation, context)
	}
	return fmt.Errorf("%s: %v %w", operation, context, err)
}

// 使用示例
func (r *messageRepository) GetByID(id uint) (*entities.Message, error) {
	db := r.Manager.DB()
	var message entities.Message
	err := db.First(&message, id).Error
	if err != nil {
		return nil, wrapDBError("查询留言", err, fmt.Sprintf("id=%d", id))
	}
	return &message, nil
}
```

---

#### 3. 修复敏感信息记录

**优先级**: 高

**行动项**:
1. 移除或脱敏所有 token 记录
2. 检查是否有其他敏感信息被记录（密码、密钥等）

---

#### 4. 修复日志级别使用不当

**优先级**: 中

**行动项**:
1. 将参数验证失败的日志从 `Warn` 改为 `Debug`
2. 将密码错误从 `Warn` 改为 `Debug`
3. 检查其他日志级别使用是否正确

---

### 🎯 中期改进（1-2 个月）

#### 5. 引入 initLogger 模式

**优先级**: 中

**行动项**:
1. 在所有 Service、Controller、Middleware 中实现 `initLogger()` 方法
2. 移除重复的 `nil` 检查
3. 考虑使用 NoOpLogger

---

#### 6. 建立统一的错误代码体系

**优先级**: 中

**行动项**:
1. 定义常见的错误代码（如 `ERR_TOKEN_INVALID`, `ERR_RECORD_NOT_FOUND`）
2. 在错误包装时添加错误代码
3. 在 API 响应中返回错误代码，便于客户端处理

**示例**:
```go
// 定义错误代码
const (
	ErrCodeTokenInvalid   = "TOKEN_INVALID"
	ErrCodeTokenExpired    = "TOKEN_EXPIRED"
	ErrCodeRecordNotFound  = "RECORD_NOT_FOUND"
)

// 定义带错误代码的错误
type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// 使用示例
return &AppError{
	Code:    ErrCodeRecordNotFound,
	Message: "留言不存在",
	Err:     err,
}
```

---

#### 7. 添加日志采样

**优先级**: 低

**行动项**:
1. 对于高频日志（如 Debug 级别的数据库操作），添加采样机制
2. 避免日志量过大

---

### 🎯 长期改进（3-6 个月）

#### 8. 建立错误监控和告警

**优先级**: 中

**行动项**:
1. 集成监控系统（如 Prometheus、Grafana）
2. 对错误日志进行聚合和告警
3. 建立错误趋势分析

---

#### 9. 添加日志审计功能

**优先级**: 低

**行动项**:
1. 对关键操作（登录、删除、修改等）添加审计日志
2. 审计日志应该持久化且不可篡改
3. 审计日志应该包含操作人、操作时间、操作内容等

---

#### 10. 改进错误文档

**优先级**: 低

**行动项**:
1. 编写错误处理指南
2. 列出常见的错误类型和处理方式
3. 提供错误处理最佳实践

---

## 附录：日志级别规范总结

| 级别 | 使用场景 | 项目中的示例 |
|------|---------|-------------|
| **Debug** | 开发调试信息 | 数据库操作成功、参数验证失败 |
| **Info** | 正常业务流程 | 请求完成、资源创建、登录成功 |
| **Warn** | 降级处理、慢查询、重试 | 慢查询、认证失败（应改为 Debug） |
| **Error** | 业务错误、操作失败 | 数据库查询失败、panic 恢复 |
| **Fatal** | 致命错误，需要立即终止 | 关闭时的错误（server/signal.go:18） |

---

## 附录：错误包装最佳实践

### ✅ 推荐

```go
// 使用 %w 包装错误
return fmt.Errorf("failed to create message: %w", err)

// 提供上下文信息
return fmt.Errorf("查询留言失败: id=%d, error=%w", id, err)

// 定义自定义错误类型
type AppError struct {
	Code    string
	Message string
	Err     error
}
```

### ❌ 不推荐

```go
// 直接返回原始错误
return err

// 使用 %s 或 %v 包装，会丢失原始错误
return fmt.Errorf("failed: %s", err)

// 忽略错误
db.Exec("UPDATE ...")
```

---

## 总结

本次审查发现了 **13 个问题**，其中：
- 严重问题：4 个
- 中等问题：5 个
- 轻微问题：4 个

项目整体在错误处理和日志方面表现良好，但仍需要针对上述问题进行改进。建议优先修复严重问题，然后逐步解决中等问题和轻微问题。

---

**审查人员**: AI Code Reviewer
**审查工具**: OpenCode
**审查标准**: AGENTS.md - 日志使用规范、Go 错误处理最佳实践
