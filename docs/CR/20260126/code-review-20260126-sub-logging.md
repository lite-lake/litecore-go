# 代码审查报告 - 日志规范维度

## 审查概览
- **审查日期**: 2026-01-26
- **审查维度**: 日志规范
- **评分**: 82/100
- **严重问题**: 4 个
- **重要问题**: 2 个
- **建议**: 5 个

## 评分细则

| 检查项 | 得分 | 说明 |
|--------|------|------|
| 日志使用规范 | 80/100 | 大部分代码正确使用依赖注入，但存在使用 fmt.Printf 和标准库 log 的违规行为 |
| 日志级别使用 | 90/100 | 级别使用合理，Debug/Info/Warn/Error/Fatal 应用恰当 |
| 敏感信息处理 | 65/100 | 存在多处直接记录 token 等敏感信息的问题，未实现脱敏机制 |
| 日志格式 | 95/100 | 支持 gin/json/default 三种格式，配置完善，时间格式统一 |
| 日志内容 | 85/100 | 消息清晰，上下文信息丰富，但 With 使用较少 |
| 日志性能 | 90/100 | 基于 Zap 高性能日志库，支持异步日志，配置合理 |

## 问题清单

### 🔴 严重问题

#### 问题 1: 定时任务中直接使用 fmt.Printf 记录错误和 panic
- **位置**: `manager/schedulermgr/cron_impl.go:212,217`
- **描述**: 在定时任务执行器中使用 fmt.Printf 直接输出 panic 和错误信息，绕过了结构化日志系统
- **影响**:
  - 违反日志规范，日志无法被统一管理和分析
  - 错误信息缺少上下文，难以追踪问题
  - panic 信息可能包含敏感数据
- **建议**: 改用注入的 LoggerMgr 记录错误和 panic 信息
- **代码示例**:
```go
// 问题代码 (manager/schedulermgr/cron_impl.go:212)
func (s *schedulerManagerImpl) executeTask(scheduler common.IBaseScheduler, tickID int64) {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                var err error
                if e, ok := r.(error); ok {
                    err = e
                } else {
                    err = fmt.Errorf("panic recovered: %v", r)
                }
                fmt.Printf("[Scheduler] %s panic: %v\n", scheduler.SchedulerName(), err)  // ❌ 违反规范
            }
        }()

        if err := scheduler.OnTick(tickID); err != nil {
            fmt.Printf("[Scheduler] %s OnTick error: %v\n", scheduler.SchedulerName(), err)  // ❌ 违反规范
        }
    }()
}

// 建议修改
type schedulerManagerImpl struct {
    loggerMgr loggermgr.ILoggerManager `inject:""`
    // ...
}

func (s *schedulerManagerImpl) executeTask(scheduler common.IBaseScheduler, tickID int64) {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                var err error
                if e, ok := r.(error); ok {
                    err = e
                } else {
                    err = fmt.Errorf("panic recovered: %v", r)
                }
                s.loggerMgr.Ins().Error("Scheduler panic recovered",
                    "scheduler", scheduler.SchedulerName(),
                    "tick_id", tickID,
                    "panic", err)
            }
        }()

        if err := scheduler.OnTick(tickID); err != nil {
            s.loggerMgr.Ins().Error("Scheduler OnTick error",
                "scheduler", scheduler.SchedulerName(),
                "tick_id", tickID,
                "error", err)
        }
    }()
}
```

#### 问题 2: 认证服务中日志记录完整 token（敏感信息泄露）
- **位置**:
  - `samples/messageboard/internal/services/auth_service.go:72`
  - `samples/messageboard/internal/controllers/admin_auth_controller.go:55`
- **描述**: 在用户登录成功后，直接记录完整的 token 到日志中，存在严重的安全风险
- **影响**:
  - Token 可能被日志收集系统保存，导致持久化泄露
  - 日志文件可能被未授权人员访问
  - 违反安全最佳实践
- **建议**: token 应该脱敏记录或完全不记录
- **代码示例**:
```go
// 问题代码 (samples/messageboard/internal/services/auth_service.go:72)
func (s *authServiceImpl) Login(password string) (string, error) {
    // ...
    s.LoggerMgr.Ins().Info("Login successful", "token", token)  // ❌ 直接记录完整 token
    return token, nil
}

// 建议修改
func (s *authServiceImpl) Login(password string) (string, error) {
    // ...
    // 方案1：不记录 token
    s.LoggerMgr.Ins().Info("Login successful", "expires_at", time.Now().Add(3600*time.Second))

    // 方案2：记录脱敏后的 token（只显示前4位和后4位）
    maskedToken := maskToken(token)
    s.LoggerMgr.Ins().Info("Login successful", "token", maskedToken)
    return token, nil
}

// token 脱敏函数
func maskToken(token string) string {
    if len(token) <= 8 {
        return "***"
    }
    return token[:4] + "..." + token[len(token)-4:]
}
```

#### 问题 3: 会话服务中多处日志记录完整 token
- **位置**:
  - `samples/messageboard/internal/services/session_service.go:70,73,85,90,95,102`
- **描述**: 在会话创建、验证、删除的各个阶段，都记录了完整的 token
- **影响**: 同问题 2，增加敏感信息泄露的风险面
- **建议**: 统一使用脱敏函数或完全不记录 token
- **代码示例**:
```go
// 问题代码片段
func (s *sessionServiceImpl) CreateSession() (string, error) {
    token := uuid.New().String()
    // ...
    s.LoggerMgr.Ins().Error("Failed to create session", "token", token, "error", err)  // ❌
    s.LoggerMgr.Ins().Info("Session created successfully", "token", token, "expires_at", session.ExpiresAt)  // ❌
    // ...
}

// 建议修改
func (s *sessionServiceImpl) CreateSession() (string, error) {
    token := uuid.New().String()
    // ...
    s.LoggerMgr.Ins().Error("Failed to create session", "error", err)
    s.LoggerMgr.Ins().Info("Session created successfully", "expires_at", session.ExpiresAt)
    // ...
}
```

#### 问题 4: 默认日志实现使用标准库 log.Fatal/log.Printf
- **位置**: `logger/default_logger.go:29,38,47,56,62,64`
- **描述**: DefaultLogger 实现中使用了标准库的 log.Fatal 和 log.Printf
- **影响**:
  - 虽然实际项目使用 Zap，但 DefaultLogger 作为后备实现应该符合规范
  - log.Fatal 会直接调用 os.Exit(1)，可能绕过优雅关闭流程
  - log.Printf 不是结构化日志，不符合框架设计理念
- **建议**: 将 DefaultLogger 改为仅用于开发调试，并在文档中明确说明
- **代码示例**:
```go
// 问题代码 (logger/default_logger.go:62-64)
func (l *DefaultLogger) Fatal(msg string, args ...any) {
    allArgs := append(l.extraArgs, args...)
    log.Printf(l.prefix+"FATAL: %s %v", msg, allArgs)
    args = append([]any{l.prefix + "FATAL: " + msg}, args...)
    log.Fatal(args...)  // ❌ 使用标准库 log.Fatal
}

// 建议：在文档中说明 DefaultLogger 仅用于开发调试
/*
DefaultLogger 是一个简单的日志实现，仅用于开发调试阶段。

生产环境必须使用基于 Zap 的日志管理器（driver: "zap"），
通过 loggermgr.ILoggerManager 接口进行依赖注入。

使用 DefaultLogger 的限制：
1. 仅用于项目初期，依赖注入尚未完成时
2. 不得在生产环境使用
3. 不保证性能和功能完整性
*/
```

### 🟡 重要问题

#### 问题 1: With 使用较少，未能充分利用结构化日志上下文
- **位置**: 全局
- **描述**: 大部分代码直接使用 `LoggerMgr.Ins().Info()`，很少使用 `With()` 方法添加固定上下文
- **影响**:
  - 日志缺少必要的固定上下文（如用户ID、请求ID等）
  - 难以在日志分析时关联同一请求的多条日志
- **建议**: 在 Controller/Service 层使用 With 创建带上下文的 logger
- **代码示例**:
```go
// 当前用法
func (c *msgCreateControllerImpl) Handle(ctx *gin.Context) {
    c.LoggerMgr.Ins().Debug("Starting to create message", "nickname", req.Nickname)
    c.LoggerMgr.Ins().Info("Message created successfully", "id", message.ID)
}

// 建议用法
func (c *msgCreateControllerImpl) Handle(ctx *gin.Context) {
    logger := c.LoggerMgr.Ins().With("request_id", c.GetString("request_id"))
    logger.Debug("Starting to create message", "nickname", req.Nickname)
    logger.Info("Message created successfully", "id", message.ID)
}

// 或者在 Service 层使用 With
func (s *messageServiceImpl) CreateMessage(nickname, content string) (*Message, error) {
    logger := s.LoggerMgr.Ins().With("service", "MessageService")
    // ...
    logger.Info("Message created successfully", "id", message.ID)
}
```

#### 问题 2: CLI 工具中使用 fmt.Printf/Println（虽然不是日志相关，但值得统一）
- **位置**:
  - `cli/scaffold/interactive.go:11-13,171`
  - `cli/scaffold/scaffold.go:37-42`
  - `cli/generator/run.go:67`
  - `cli/cmd/version.go:17,34-68`
  - `samples/messageboard/cmd/genpasswd/main.go:15-16,38-42,58,63`
- **描述**: CLI 工具大量使用 fmt.Printf/Println 输出用户界面信息
- **影响**:
  - 虽然这些是用户交互输出，不属于日志，但容易与日志混淆
  - 建议明确区分日志输出和用户界面输出
- **建议**:
  - 在 AGENTS.md 中明确说明 CLI 工具可以使用 fmt.Printf/Println 输出用户界面信息
  - 或者考虑使用专门的 UI 输出库

### 🟢 建议

#### 建议 1: 统一日志消息的中英文规范
- **位置**: 全局
- **描述**: 当前日志消息都是中文，但有些地方可能需要考虑国际化
- **建议**: 在 AGENTS.md 中明确日志消息统一使用中文，或者考虑支持多语言

#### 建议 2: 添加日志采样配置
- **位置**: `samples/messageboard/configs/config.yaml`
- **描述**: 日志配置中没有采样率配置，高并发场景下可能产生大量日志
- **建议**: 参考 database.observability_config.sample_rate，为日志添加采样配置
- **代码示例**:
```yaml
logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "info"
      format: "gin"
      color: true
      time_format: "2006-01-02 15:04:05.000"
      sample_rate: 0.1  # 新增：采样率 10%（仅记录 10% 的日志）
```

#### 建议 3: 完善日志字段命名规范
- **位置**: 全局
- **描述**: 当前日志字段命名基本一致，但可以更规范
- **建议**: 在 AGENTS.md 中补充日志字段命名规范
  - user_id / request_id / message_id: 使用下划线分隔
  - nickname / status: 使用小写单词
  - clientIP / userAgent: 使用驼峰命名（当前使用）

#### 建议 4: 添加慢请求日志
- **位置**: `component/litemiddleware/request_logger_middleware.go`
- **描述**: 当前 RequestLoggerMiddleware 记录所有请求的日志，但没有慢请求特别标识
- **建议**: 当请求耗时超过阈值时，自动升级为 WARN 级别
- **代码示例**:
```go
// request_logger_middleware.go:143-212
latency := time.Since(start)
// ...

// 慢请求检测
slowThreshold := 1 * time.Second  // 可配置
if latency > slowThreshold {
    logFunc = m.LoggerMgr.Ins().Warn
    fields = append(fields, "slow_request", true, "slow_threshold", slowThreshold)
}

logFunc("Request processed successfully", fields...)
```

#### 建议 5: 添加日志脱敏工具函数
- **位置**: 新建 `util/logger/mask.go`
- **描述**: 提供统一的敏感信息脱敏函数，避免各处自行实现
- **建议**: 在 logger 包中提供常用的脱敏函数
- **代码示例**:
```go
// util/logger/mask.go
package logger

import (
    "strings"
)

// MaskToken 脱敏 token，只显示前4位和后4位
func MaskToken(token string) string {
    if len(token) <= 8 {
        return "***"
    }
    return token[:4] + "..." + token[len(token)-4:]
}

// MaskEmail 脱敏邮箱
func MaskEmail(email string) string {
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return "***"
    }
    if len(parts[0]) <= 3 {
        return "***@" + parts[1]
    }
    return parts[0][:3] + "***@" + parts[1]
}

// MaskPhone 脱敏手机号
func MaskPhone(phone string) string {
    if len(phone) != 11 {
        return "***"
    }
    return phone[:3] + "****" + phone[7:]
}

// MaskString 脱敏字符串，保留首尾各 n 位
func MaskString(s string, keepPrefix, keepSuffix int) string {
    if len(s) <= keepPrefix+keepSuffix {
        return strings.Repeat("*", len(s))
    }
    return s[:keepPrefix] + strings.Repeat("*", len(s)-keepPrefix-keepSuffix) + s[len(s)-keepSuffix:]
}

// 使用示例
s.LoggerMgr.Ins().Info("Login successful", "email", logger.MaskEmail(user.Email))
```

## 亮点总结

1. **日志架构设计优秀**: 采用依赖注入模式，通过 loggermgr.ILoggerManager 统一管理，各层通过 `inject:""` 标签自动注入，符合框架设计理念

2. **结构化日志应用良好**: 绝大多数日志使用 `logger.Info("msg", "key", value)` 的结构化格式，便于日志分析和查询

3. **日志格式配置完善**: 支持 gin（竖线分隔符）、json（适合日志分析）、default（ConsoleEncoder）三种格式，时间格式统一为 `2006-01-02 15:04:05.000`

4. **日志级别使用合理**: Debug（调试信息）、Info（正常业务流程）、Warn（降级处理）、Error（业务错误）、Fatal（致命错误）使用恰当，符合最佳实践

5. **日志性能优秀**: 基于 Uber Zap 高性能日志库，支持异步日志（startup_log.async: true），配置合理

6. **各层日志职责清晰**:
   - Controller 层：请求开始、参数错误、处理成功
   - Service 层：业务逻辑、验证失败、数据库操作
   - Listener 层：MQ 消息接收、处理
   - Scheduler 层：定时任务执行
   - Middleware 层：请求日志、panic 恢复

7. **日志字段命名一致**: 普遍使用小写加下划线的命名方式（如 message_id、nickname、status），便于统一解析

## 改进建议优先级

1. **[P0-立即修复]** 敏感信息脱敏：token、密码等敏感信息必须脱敏记录
   - 影响：安全风险
   - 位置：`samples/messageboard/internal/services/auth_service.go`, `session_service.go`, `controllers/admin_auth_controller.go`

2. **[P0-立即修复]** Scheduler 中使用 fmt.Printf 改为结构化日志
   - 影响：违反日志规范，日志无法统一管理
   - 位置：`manager/schedulermgr/cron_impl.go:212,217`

3. **[P1-短期改进]** 推广使用 With 添加上下文
   - 影响：日志关联性不足，难以追踪请求链路
   - 建议：在 Controller/Service 层使用 With 创建带 request_id 的 logger

4. **[P1-短期改进]** 添加慢请求日志
   - 影响：无法快速定位性能瓶颈
   - 建议：在 RequestLoggerMiddleware 中添加慢请求检测和 WARN 级别记录

5. **[P2-长期优化]** 完善日志脱敏工具
   - 影响：提高开发效率，减少重复代码
   - 建议：在 logger 包中提供统一的脱敏函数（MaskToken、MaskEmail、MaskPhone 等）

6. **[P2-长期优化]** 添加日志采样配置
   - 影响：高并发场景下日志量过大
   - 建议：参考 database.observability_config.sample_rate，为日志添加采样配置

## 审查人员
- 审查人：日志规范审查 Agent
- 审查时间：2026-01-26
