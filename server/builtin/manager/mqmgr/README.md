# 消息队列管理器 (mqmgr)

提供统一的消息队列管理功能，支持内存队列和 RabbitMQ 两种实现方式。

## 特性

- **多驱动支持** - 支持 Memory（内存）和 RabbitMQ 两种消息队列实现
- **统一接口** - 提供一致的 API，方便在不同实现间切换
- **消息确认机制** - 支持 Ack/Nack 机制，可配置是否重新入队
- **观察性集成** - 内置日志、链路追踪和指标监控
- **可配置性** - 支持队列大小、通道缓冲区等参数配置

## 快速开始

```go
import (
    "context"
    "github.com/lite-lake/litecore-go/server/builtin/manager/mqmgr"
)

// 创建内存队列管理器
config := &mqmgr.MemoryConfig{
    MaxQueueSize:  10000,
    ChannelBuffer: 100,
}
manager := mqmgr.NewMessageQueueManagerMemoryImpl(config)
defer manager.Close()

// 发布消息
ctx := context.Background()
err := manager.Publish(ctx, "my_queue", []byte("hello world"))
if err != nil {
    log.Fatal(err)
}

// 订阅消息
msgCh, err := manager.Subscribe(ctx, "my_queue", mqmgr.WithAutoAck(false))
if err != nil {
    log.Fatal(err)
}

for msg := range msgCh {
    // 处理消息
    if err := handleMessage(msg); err != nil {
        manager.Nack(ctx, msg, true) // 拒绝并重新入队
    } else {
        manager.Ack(ctx, msg) // 确认消息
    }
}

// 使用回调函数订阅
err = manager.SubscribeWithCallback(ctx, "my_queue", func(ctx context.Context, msg mqmgr.Message) error {
    return handleMessage(msg)
})
```

## 创建管理器

### 内存队列

```go
config := &mqmgr.MemoryConfig{
    MaxQueueSize:  10000,  // 最大队列大小
    ChannelBuffer: 100,    // 通道缓冲区大小
}
manager := mqmgr.NewMessageQueueManagerMemoryImpl(config)
```

### RabbitMQ

```go
config := &mqmgr.RabbitMQConfig{
    URL:     "amqp://guest:guest@localhost:5672/",
    Durable: true,  // 持久化队列
}
manager, err := mqmgr.NewMessageQueueManagerRabbitMQImpl(config)
if err != nil {
    log.Fatal(err)
}
```

### 使用工厂方法

```go
// 根据驱动类型构建
driverConfig := map[string]any{
    "url":     "amqp://guest:guest@localhost:5672/",
    "durable": true,
}
manager, err := mqmgr.Build("rabbitmq", driverConfig)

// 从配置提供者构建
manager, err := mqmgr.BuildWithConfigProvider(configProvider)
```

## 发布消息

### 基本发布

```go
err := manager.Publish(ctx, "my_queue", []byte("message content"))
```

### 带选项发布

```go
headers := map[string]any{
    "content-type": "application/json",
    "user-id":      "12345",
}
err := manager.Publish(ctx, "my_queue", []byte("message content"),
    mqmgr.WithPublishHeaders(headers),
    mqmgr.WithPublishDurable(true),
)
```

## 订阅消息

### 通道订阅

```go
// 自动确认
msgCh, err := manager.Subscribe(ctx, "my_queue")

// 手动确认
msgCh, err := manager.Subscribe(ctx, "my_queue", mqmgr.WithAutoAck(false))

for msg := range msgCh {
    // 处理消息
    if err := process(msg); err != nil {
        manager.Nack(ctx, msg, true) // 重新入队
    } else {
        manager.Ack(ctx, msg)
    }
}
```

### 回调订阅

```go
err := manager.SubscribeWithCallback(ctx, "my_queue", func(ctx context.Context, msg mqmgr.Message) error {
    // 处理消息
    if err := process(msg); err != nil {
        return err  // 返回错误会自动 Nack
    }
    return nil
})
```

## 消息确认

### Ack - 确认消息

```go
err := manager.Ack(ctx, msg)
```

### Nack - 拒绝消息

```go
// 拒绝消息，不重新入队
err := manager.Nack(ctx, msg, false)

// 拒绝消息，重新入队
err := manager.Nack(ctx, msg, true)
```

## 队列管理

### 获取队列长度

```go
length, err := manager.QueueLength(ctx, "my_queue")
```

### 清空队列

```go
err := manager.Purge(ctx, "my_queue")
```

## API

### IMQManager 接口

| 方法 | 说明 |
|------|------|
| `ManagerName()` | 获取管理器名称 |
| `Health()` | 健康检查 |
| `OnStart()` | 启动管理器 |
| `OnStop()` | 停止管理器 |
| `Publish(ctx, queue, message, ...opts)` | 发布消息 |
| `Subscribe(ctx, queue, ...opts)` | 订阅消息 |
| `SubscribeWithCallback(ctx, queue, handler, ...opts)` | 使用回调订阅 |
| `Ack(ctx, message)` | 确认消息 |
| `Nack(ctx, message, requeue)` | 拒绝消息 |
| `QueueLength(ctx, queue)` | 获取队列长度 |
| `Purge(ctx, queue)` | 清空队列 |
| `Close()` | 关闭管理器 |

### PublishOptions 发布选项

| 选项 | 说明 |
|------|------|
| `WithPublishHeaders(headers)` | 设置消息头 |
| `WithPublishDurable(durable)` | 设置是否持久化 |

### SubscribeOptions 订阅选项

| 选项 | 说明 |
|------|------|
| `WithSubscribeDurable(durable)` | 设置队列是否持久化 |
| `WithAutoAck(autoAck)` | 设置是否自动确认 |

### Message 接口

| 方法 | 说明 |
|------|------|
| `ID()` | 获取消息 ID |
| `Body()` | 获取消息体 |
| `Headers()` | 获取消息头 |

## 配置

### MQConfig

```yaml
driver: "rabbitmq"  # "memory" 或 "rabbitmq"
rabbitmq_config:
  url: "amqp://guest:guest@localhost:5672/"
  durable: true
memory_config:
  max_queue_size: 10000
  channel_buffer: 100
```

## 驱动对比

| 特性 | Memory | RabbitMQ |
|------|--------|----------|
| 持久化 | 不支持 | 支持 |
| 分布式 | 不支持 | 支持 |
| 性能 | 极高 | 高 |
| 使用场景 | 开发测试、轻量级应用 | 生产环境、分布式系统 |

## 错误处理

所有方法返回错误，建议使用 `if err != nil` 进行检查：

```go
err := manager.Publish(ctx, "queue", []byte("message"))
if err != nil {
    // 处理错误
    log.Printf("发布失败: %v", err)
    return err
}
```

## 最佳实践

1. **选择合适的驱动** - 生产环境使用 RabbitMQ，开发测试使用 Memory
2. **手动确认重要消息** - 对于不能丢失的消息，使用 `WithAutoAck(false)` 手动确认
3. **合理设置队列大小** - 根据业务需求设置 `MaxQueueSize` 避免内存溢出
4. **处理错误** - 在回调函数中处理错误，决定是否重新入队
5. **及时关闭** - 使用 `defer manager.Close()` 确保资源释放
