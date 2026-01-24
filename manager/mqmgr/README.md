# 消息队列管理器 (mqmgr)

提供统一的消息队列管理功能，支持 RabbitMQ 和 Memory 两种驱动。

## 特性

- **多驱动支持** - 支持 RabbitMQ（分布式）和 Memory（高性能内存）两种消息队列驱动
- **统一接口** - 提供 IMQManager 接口，便于切换不同的消息队列实现
- **消息确认机制** - 支持 Ack/Nack 机制，可配置是否重新入队
- **可观测性集成** - 内置日志记录、链路追踪和指标监控
- **选项模式** - 使用选项模式灵活配置发布和订阅行为
- **依赖注入** - 支持通过 DI 注入 LoggerManager 和 TelemetryManager

## 快速开始

### 使用内存队列

```go
import (
    "context"
    "github.com/lite-lake/litecore-go/manager/mqmgr"
)

// 创建内存队列管理器
config := &mqmgr.MemoryConfig{
    MaxQueueSize:  10000,
    ChannelBuffer: 100,
}
manager := mqmgr.NewMessageQueueManagerMemoryImpl(config, nil, nil)
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
```

### 使用 RabbitMQ

```go
config := &mqmgr.RabbitMQConfig{
    URL:     "amqp://guest:guest@localhost:5672/",
    Durable: true,
}
manager, err := mqmgr.NewMessageQueueManagerRabbitMQImpl(config, nil, nil)
if err != nil {
    log.Fatal(err)
}
defer manager.Close()

// 使用方式与内存队列相同
```

### 使用回调函数订阅

```go
err := manager.SubscribeWithCallback(ctx, "my_queue", func(ctx context.Context, msg mqmgr.Message) error {
    // 处理消息，返回 nil 自动 Ack，返回 error 自动 Nack
    return handleMessage(msg)
})
```

## 支持的驱动

### RabbitMQ

基于 `amqp091-go` 实现的分布式消息队列，适合生产环境使用。

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `URL` | string | RabbitMQ 连接地址 | `amqp://guest:guest@localhost:5672/` |
| `Durable` | bool | 队列是否持久化 | `true` |

### Memory

基于内存的高性能消息队列，适合开发测试和轻量级应用。

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `MaxQueueSize` | int | 最大队列大小 | `10000` |
| `ChannelBuffer` | int | 通道缓冲区大小 | `100` |

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
// 自动确认（默认）
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
        return err  // 返回 error 会自动 Nack（不重新入队）
    }
    return nil  // 返回 nil 会自动 Ack
})
```

## 消息确认机制

### Ack - 确认消息

确认消息已成功处理，消息从队列中移除。

```go
err := manager.Ack(ctx, msg)
```

### Nack - 拒绝消息

拒绝消息，可选择是否重新入队。

```go
// 拒绝消息，不重新入队（消息将被丢弃）
err := manager.Nack(ctx, msg, false)

// 拒绝消息，重新入队（消息将再次被消费）
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

## 工厂方法

### Build - 根据驱动类型构建

```go
// 构建 RabbitMQ 管理器
manager, err := mqmgr.Build("rabbitmq", map[string]any{
    "url":     "amqp://guest:guest@localhost:5672/",
    "durable": true,
}, loggerMgr, telemetryMgr)

// 构建 Memory 管理器
manager, err := mqmgr.Build("memory", map[string]any{
    "max_queue_size": 10000,
    "channel_buffer": 100,
}, loggerMgr, telemetryMgr)
```

### BuildWithConfigProvider - 从配置提供者构建

```go
manager, err := mqmgr.BuildWithConfigProvider(configProvider, loggerMgr, telemetryMgr)
```

配置路径：
- `mq.driver`: 驱动类型（`rabbitmq` 或 `memory`）
- `mq.rabbitmq_config`: RabbitMQ 配置
- `mq.memory_config`: Memory 配置

## API

### IMQManager 接口

| 方法 | 说明 |
|------|------|
| `ManagerName()` | 获取管理器名称 |
| `Health()` | 健康检查 |
| `OnStart()` | 启动管理器 |
| `OnStop()` | 停止管理器 |
| `Publish(ctx, queue, message, ...opts)` | 发布消息 |
| `Subscribe(ctx, queue, ...opts)` | 订阅消息，返回消息通道 |
| `SubscribeWithCallback(ctx, queue, handler, ...opts)` | 使用回调函数订阅 |
| `Ack(ctx, message)` | 确认消息 |
| `Nack(ctx, message, requeue)` | 拒绝消息 |
| `QueueLength(ctx, queue)` | 获取队列长度 |
| `Purge(ctx, queue)` | 清空队列 |
| `Close()` | 关闭管理器 |

### Message 接口

| 方法 | 说明 |
|------|------|
| `ID()` | 获取消息 ID |
| `Body()` | 获取消息体 |
| `Headers()` | 获取消息头 |

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

## 配置

### MQConfig

```yaml
mq:
  driver: "rabbitmq"  # "memory" 或 "rabbitmq"
  rabbitmq_config:
    url: "amqp://guest:guest@localhost:5672/"
    durable: true
  memory_config:
    max_queue_size: 10000
    channel_buffer: 100
```

## 可观测性

mqmgr 内置了完整的可观测性支持，包括：

- **日志记录**: 自动记录操作成功/失败信息，包含操作类型、队列名、耗时等
- **链路追踪**: 使用 OpenTelemetry 追踪消息队列操作，便于问题排查
- **指标监控**: 提供以下指标：
  - `mq.publish`: 消息发布次数
  - `mq.consume`: 消息消费次数
  - `mq.ack`: 消息确认次数
  - `mq.nack`: 消息拒绝次数
  - `mq.operation.duration`: 操作耗时（秒）

### 依赖注入

mqmgr 通过依赖注入接收 LoggerManager 和 TelemetryManager：

```go
type MyService struct {
    MQManager mqmgr.IMQManager `inject:""`
}

func (s *MyService) init() {
    // 注入的 MQManager 会自动初始化可观测性组件
}
```

## 与 IBaseListener 集成

在 5 层依赖注入架构中，可以实现 `common.IBaseListener` 接口来创建消息监听器：

```go
import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
)

type MyListener struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func (l *MyListener) ListenerName() string {
    return "MyListener"
}

func (l *MyListener) GetQueue() string {
    return "my.queue"
}

func (l *MyListener) GetSubscribeOptions() []common.ISubscribeOption {
    return []common.ISubscribeOption{}
}

func (l *MyListener) OnStart() error {
    l.LoggerMgr.Ins().Info("Listener started")
    return nil
}

func (l *MyListener) OnStop() error {
    l.LoggerMgr.Ins().Info("Listener stopped")
    return nil
}

func (l *MyListener) Handle(ctx context.Context, msg common.IMessageListener) error {
    l.LoggerMgr.Ins().Info("Received message",
        "message_id", msg.ID(),
        "body", string(msg.Body()),
        "headers", msg.Headers())
    return nil
}
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
3. **合理设置队列大小** - 根据 Memory 配置设置 `MaxQueueSize` 避免内存溢出
4. **处理错误** - 在回调函数中处理错误，决定是否重新入队
5. **及时关闭** - 使用 `defer manager.Close()` 确保资源释放
6. **并发安全** - 两个驱动都支持并发操作，但请注意内存队列的 `MaxQueueSize` 限制

## 测试

```bash
# 运行所有测试
go test ./manager/mqmgr/...

# 运行内存队列测试
go test ./manager/mqmgr/ -run TestMemoryManager

# 运行 RabbitMQ 测试（需要启动 RabbitMQ）
go test ./manager/mqmgr/ -run TestRabbitMQManager
```
