# Middleware / Listener / Scheduler 模板

## Middleware

> 位置: `internal/middlewares/YOUR_middleware.go`

```go
package middlewares

type IYOUR_Middleware interface { common.IBaseMiddleware }

type yourMiddlewareImpl struct {
	SomeService services.ISomeService `inject:""`
}

func NewYOUR_Middleware() IYOUR_Middleware { return &yourMiddlewareImpl{} }

func (m *yourMiddlewareImpl) MiddlewareName() string { return "YOUR_Middleware" }
func (m *yourMiddlewareImpl) Order() int             { return 300 }  // 从 300 开始

func (m *yourMiddlewareImpl) Handle(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.AbortWithStatusJSON(401, gin.H{"error": "未授权"})
		return
	}
	ctx.Next()
}
```

## Listener

> 位置: `internal/listeners/YOUR_listener.go`

```go
package listeners

type IYOUR_Listener interface { common.IBaseListener }

type yourListenerImpl struct {
	MQMgr      mqmgr.IMQManager       `inject:""`
	SomeService services.ISomeService `inject:""`
}

func NewYOUR_Listener() IYOUR_Listener { return &yourListenerImpl{} }

func (l *yourListenerImpl) ListenerName() string { return "YOUR_Listener" }
func (l *yourListenerImpl) QueueName() string    { return "YOUR_QUEUE" }

func (l *yourListenerImpl) Handle(ctx context.Context, msg mqmgr.Message) error {
	return l.SomeService.Process(msg.Body)
}
```

## Scheduler

> 位置: `internal/schedulers/YOUR_scheduler.go`

```go
package schedulers

type IYOUR_Scheduler interface { common.IBaseScheduler }

type yourSchedulerImpl struct {
	SomeService services.ISomeService `inject:""`
}

func NewYOUR_Scheduler() IYOUR_Scheduler { return &yourSchedulerImpl{} }

func (s *yourSchedulerImpl) SchedulerName() string  { return "YOUR_Scheduler" }
func (s *yourSchedulerImpl) CronExpression() string { return "0 0 * * *" }  // 每天 0 点

func (s *yourSchedulerImpl) Handle(ctx context.Context) error {
	return s.SomeService.Cleanup()
}
```
