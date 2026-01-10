# CacheManager 缓存管理器

CacheManager 是 LiteCore 框架的缓存管理组件，负责缓存的创建、管理和生命周期控制。

## 功能特性

- 支持 **Redis**、**Memory** 和 **None** 三种缓存驱动
- 提供统一的缓存操作接口
- 支持基本 CRUD 操作（Get/Set/Delete）
- 支持批量操作（GetMultiple/SetMultiple/DeleteMultiple）
- 支持计数器操作（Increment/Decrement）
- 支持过期时间管理（TTL/Expire）
- 支持分布式锁（SetNX）
- 自动降级到 None 驱动（当配置或初始化失败时）

## 快速开始

### 1. 使用内存缓存

```go
package main

import (
    "context"
    "fmt"
    "time"

    "com.litelake.litecore/manager/cachemgr"
)

func main() {
    // 创建内存缓存管理器
    cacheMgr := cachemgr.BuildMemory(5*time.Minute, 10*time.Minute)

    ctx := context.Background()

    // 设置缓存
    err := cacheMgr.Set(ctx, "user:123", "John Doe", 10*time.Minute)
    if err != nil {
        panic(err)
    }

    // 获取缓存（使用内部方法）
    // 注意：Get 方法需要传入指针参数，实际使用时可能需要根据驱动类型选择
    fmt.Println("Cache manager created:", cacheMgr.ManagerName())
}
```

### 2. 使用 Redis 缓存

```go
// 创建 Redis 缓存管理器
cacheMgr := cachemgr.BuildRedis("localhost", 6379, "", 0)

// 或使用配置方式
cfg := map[string]any{
    "driver": "redis",
    "redis_config": map[string]any{
        "host":     "localhost",
        "port":     6379,
        "password": "",
        "db":       0,
    },
}
cacheMgr := cachemgr.Build(cfg)
```

### 3. 使用配置创建

```go
cfg := map[string]any{
    "driver": "memory",
    "memory_config": map[string]any{
        "max_size": 100,    // MB
        "max_age":  "30d",  // 30 days
    },
}
cacheMgr := cachemgr.Build(cfg)
```

## 配置说明

### Redis 配置

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| host | string | localhost | Redis 主机地址 |
| port | int | 6379 | Redis 端口 |
| password | string | "" | Redis 密码 |
| db | int | 0 | Redis 数据库编号（0-15） |
| max_idle_conns | int | 10 | 最大空闲连接数 |
| max_open_conns | int | 100 | 最大打开连接数 |
| conn_max_lifetime | duration | 30s | 连接最大存活时间 |

### Memory 配置

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| max_size | int | 100 | 最大缓存大小（MB） |
| max_age | duration | 30d | 最大缓存时间 |
| max_backups | int | 1000 | 最大备份项数 |
| compress | bool | false | 是否压缩 |

## API 参考

### 基本操作

#### Set 设置缓存值

```go
err := cacheMgr.Set(ctx, "key", "value", 10*time.Minute)
```

#### Get 获取缓存值

```go
var value string
err := cacheMgr.Get(ctx, "key", &value)
```

#### Delete 删除缓存值

```go
err := cacheMgr.Delete(ctx, "key")
```

#### Exists 检查键是否存在

```go
exists, err := cacheMgr.Exists(ctx, "key")
```

### 过期时间管理

#### TTL 获取剩余过期时间

```go
ttl, err := cacheMgr.TTL(ctx, "key")
fmt.Println("剩余时间:", ttl)
```

#### Expire 设置过期时间

```go
err := cacheMgr.Expire(ctx, "key", 5*time.Minute)
```

### 批量操作

#### SetMultiple 批量设置

```go
items := map[string]any{
    "user:123": user1,
    "user:456": user2,
}
err := cacheMgr.SetMultiple(ctx, items, 10*time.Minute)
```

#### GetMultiple 批量获取

```go
keys := []string{"user:123", "user:456"}
values, err := cacheMgr.GetMultiple(ctx, keys)
```

#### DeleteMultiple 批量删除

```go
keys := []string{"user:123", "user:456"}
err := cacheMgr.DeleteMultiple(ctx, keys)
```

### 计数器操作

#### Increment 自增

```go
counter, err := cacheMgr.Increment(ctx, "views:page:1", 1)
```

#### Decrement 自减

```go
counter, err := cacheMgr.Decrement(ctx, "views:page:1", 1)
```

### 分布式锁

#### SetNX 仅当键不存在时设置

```go
// 返回值表示是否设置成功
set, err := cacheMgr.SetNX(ctx, "lock:resource", "locked", 10*time.Second)
if set {
    // 获取锁成功
    defer cacheMgr.Delete(ctx, "lock:resource")
    // 执行业务逻辑
}
```

### 其他操作

#### Clear 清空所有缓存

```go
err := cacheMgr.Clear(ctx)
```

#### Close 关闭连接

```go
err := cacheMgr.Close()
```

## 驱动特性对比

| 特性 | Redis | Memory | None |
|------|-------|--------|------|
| 分布式支持 | ✅ | ❌ | ❌ |
| 持久化 | ✅ | ❌ | ❌ |
| TTL 精确查询 | ✅ | ⚠️ | ❌ |
| 高性能 | ⚠️ | ✅ | ✅ |
| 降级支持 | ❌ | ❌ | ✅ |

## 最佳实践

### 1. 选择合适的驱动

- **Redis**：需要分布式缓存、数据持久化的场景
- **Memory**：单机应用、需要高性能的场景
- **None**：测试环境、不需要缓存功能的场景

### 2. 合理设置过期时间

```go
// 短期缓存（热点数据）
cacheMgr.Set(ctx, "hot:data", value, 5*time.Minute)

// 中期缓存（一般数据）
cacheMgr.Set(ctx, "user:123", value, 30*time.Minute)

// 长期缓存（静态数据）
cacheMgr.Set(ctx, "config:static", value, 24*time.Hour)
```

### 3. 使用批量操作提高性能

```go
// ❌ 不推荐：多次单独操作
for _, user := range users {
    cacheMgr.Set(ctx, "user:"+user.ID, user, time.Hour)
}

// ✅ 推荐：批量操作
items := make(map[string]any)
for _, user := range users {
    items["user:"+user.ID] = user
}
cacheMgr.SetMultiple(ctx, items, time.Hour)
```

### 4. 处理缓存错误

```go
value, err := cacheMgr.Get(ctx, "key", &dest)
if err != nil {
    // 缓存未命中，从数据库加载
    dest, err = loadFromDB("key")
    if err == nil {
        // 写入缓存
        cacheMgr.Set(ctx, "key", dest, 10*time.Minute)
    }
}
```

### 5. 使用 SetNX 实现分布式锁

```go
func acquireLock(cacheMgr cachemgr.CacheManager, lockKey string, expiry time.Duration) (bool, error) {
    return cacheMgr.SetNX(context.Background(), lockKey, "locked", expiry)
}

func releaseLock(cacheMgr cachemgr.CacheManager, lockKey string) error {
    return cacheMgr.Delete(context.Background(), lockKey)
}

// 使用
locked, err := acquireLock(cacheMgr, "lock:resource", 10*time.Second)
if err != nil {
    return err
}
if locked {
    defer releaseLock(cacheMgr, "lock:resource")
    // 执行业务逻辑
}
```

## 降级策略

当以下情况发生时，CacheManager 会自动降级到 None 驱动：

1. 配置解析失败
2. 驱动初始化失败
3. Redis 连接创建失败

None 驱动的行为：
- 所有 Get 操作返回 "cache not available" 错误
- 所有 Set/Delete 操作为空操作（不返回错误）
- Health() 返回错误
- OnStart() 和 OnStop() 无操作

## 注意事项

1. **序列化**：Redis 驱动使用 gob 编码进行序列化，复杂类型需要注册
2. **并发安全**：Memory 驱动基于 go-cache，是并发安全的
3. **内存限制**：Memory 驱动受限于进程内存，不适合存储大量数据
4. **TTL 限制**：Memory 驱动的 TTL 查询返回 0（go-cache 限制）
5. **密码安全**：不要在代码中硬编码 Redis 密码，使用环境变量或配置文件

## 依赖

- Redis 驱动：`github.com/redis/go-redis/v9`
- Memory 驱动：`github.com/patrickmn/go-cache`

## 测试

运行测试：

```bash
# 运行所有测试
go test ./manager/cachemgr/...

# 运行特定测试
go test ./manager/cachemgr/internal/drivers/ -run TestMemoryManager

# 运行 Redis 测试（需要 miniredis）
go test ./manager/cachemgr/internal/drivers/ -run TestRedisManager
```

## 许可证

本项目采用 MIT 许可证。
