# Manager 接口速查

## Manager 列表

| Manager | 接口 | 驱动 |
|---------|------|------|
| ConfigManager | `IConfigManager` | yaml, json |
| LoggerManager | `ILoggerManager` | zap, default, none |
| DatabaseManager | `IDatabaseManager` | mysql, postgresql, sqlite, none |
| CacheManager | `ICacheManager` | redis, memory, none |
| LockManager | `ILockManager` | redis, memory |
| LimiterManager | `ILimiterManager` | redis, memory |
| MQManager | `IMQManager` | rabbitmq, memory |
| TelemetryManager | `ITelemetryManager` | otel, none |

## 常用接口

### ConfigManager
```go
name, _ := configmgr.Get[string](s.Config, "app.name")
timeout := configmgr.GetWithDefault(s.Config, "server.timeout", 30)
```

### LoggerManager
```go
s.LoggerMgr.Ins().Info("操作开始", "user_id", 123)
s.LoggerMgr.Ins().Error("操作失败", "error", err)
```

### DatabaseManager
```go
r.DBManager.DB().Create(entity)
r.DBManager.DB().Where("id = ?", id).First(&entity)
r.DBManager.DB().Transaction(func(tx *gorm.DB) error { ... })
```

### CacheManager
```go
s.CacheMgr.Get(ctx, "session:"+token, &session)
s.CacheMgr.Set(ctx, "session:"+token, session, time.Hour)
s.CacheMgr.Delete(ctx, "session:"+token)
```

### LockManager
```go
if err := s.LockMgr.Lock(ctx, "resource:"+id, 10*time.Second); err != nil { return err }
defer s.LockMgr.Unlock(ctx, "resource:"+id)
```

### LimiterManager
```go
if allowed, _ := s.LimiterMgr.Allow(ctx, "api:"+userID, 10, time.Minute); !allowed {
    return errors.New("请求过于频繁")
}
```

### MQManager
```go
s.MQMgr.Publish(ctx, "notifications", body)
s.MQMgr.SubscribeWithCallback(ctx, "my_queue", handler)
```

## 注入方式

```go
type MyService struct {
    Config     configmgr.IConfigManager     `inject:""`
    DBManager  databasemgr.IDatabaseManager `inject:""`
    LoggerMgr  loggermgr.ILoggerManager     `inject:""`
    CacheMgr   cachemgr.ICacheManager       `inject:""`
    LockMgr    lockmgr.ILockManager         `inject:""`
    LimiterMgr limitermgr.ILimiterManager   `inject:""`
    MQMgr      mqmgr.IMQManager             `inject:""`
}
```
