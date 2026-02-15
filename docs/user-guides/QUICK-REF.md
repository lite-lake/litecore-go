# 速查卡

## 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 接口 | `I` 前缀 | `IMessageService`、`IUserRepository` |
| 实现 | 小写 | `messageServiceImpl` |
| 工厂 | `New` + 组件名 | `NewMessageService()` |
| 名称方法 | 返回组件名 | `ServiceName() string { return "MessageService" }` |

## 依赖规则

```
Entity → Repository → Service → Controller/Middleware/Listener/Scheduler
```

| 层级 | 可依赖 | 禁止 |
|------|--------|------|
| Entity | 无 | 所有 |
| Repository | Entity, Manager | Service, 交互层 |
| Service | Repository, Manager, Service | 交互层 |
| 交互层 | Service, Manager | Repository, Entity |

## 注入模式

```go
type messageServiceImpl struct {
    Repository repositories.IMessageRepository `inject:""`
    LoggerMgr  loggermgr.ILoggerManager        `inject:""`
}
```

## 接口签名速查

```go
// Entity
func (e *Message) EntityName() string      { return "Message" }
func (Message) TableName() string          { return "messages" }  // 非指针
func (e *Message) GetId() string           { return e.ID }

// Repository
func (r *messageRepositoryImpl) RepositoryName() string { return "MessageRepository" }

// Service
func (s *messageServiceImpl) ServiceName() string { return "MessageService" }

// Controller
func (c *messageControllerImpl) ControllerName() string { return "MessageController" }
func (c *messageControllerImpl) GetRouter() string {
    return "/api/messages [POST],/api/messages [GET],/api/messages/:id [GET]"
}

// Middleware
func (m *authMiddlewareImpl) MiddlewareName() string { return "AuthMiddleware" }
func (m *authMiddlewareImpl) Order() int             { return 300 }  // 从 300 开始

// Listener
func (l *notificationListenerImpl) ListenerName() string { return "NotificationListener" }
func (l *notificationListenerImpl) QueueName() string    { return "notifications" }

// Scheduler
func (s *cleanupSchedulerImpl) SchedulerName() string  { return "CleanupScheduler" }
func (s *cleanupSchedulerImpl) CronExpression() string { return "0 0 * * *" }
```

## 禁止清单

- ❌ 手动设置 `ID`、`CreatedAt`、`UpdatedAt`
- ❌ 使用 `uint/int` 作为 ID
- ❌ `First(&entity, id)` — 用 `Where("id = ?", id).First(&entity)`
- ❌ `log.Fatal/Println` — 用 `LoggerMgr.Ins().Info/Error`
- ❌ `fmt.Println` — 仅限开发调试
- ❌ `return err` — 包装为 `fmt.Errorf("操作失败: %w", err)`
- ❌ Middleware Order < 300 — 会覆盖框架中间件
