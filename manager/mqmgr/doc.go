package mqmgr

// Package mqmgr 提供统一的消息队列管理功能，支持 RabbitMQ 和内存两种驱动。
//
// 核心特性：
//   - 多驱动支持：支持 RabbitMQ（分布式）、Memory（高性能内存）两种消息队列驱动
//   - 统一接口：提供统一的 IMQManager 接口，便于切换消息队列实现
//   - 可观测性：内置日志、指标和链路追踪支持
//   - 消息确认：支持手动 Ack/Nack 机制
//   - 选项模式：使用选项模式配置发布和订阅行为
//
// 基本用法：
//
//	// 使用内存队列
//	mgr := mqmgr.NewMessageQueueManagerMemoryImpl(&mqmgr.MemoryConfig{
//	    MaxQueueSize:  10000,
//	    ChannelBuffer: 100,
//	})
//	defer mgr.Close()
//
//	ctx := context.Background()
//
//	// 发布消息
//	err := mgr.Publish(ctx, "user_queue", []byte("hello world"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 订阅队列
//	msgCh, err := mgr.Subscribe(ctx, "user_queue")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 消费消息
//	for msg := range msgCh {
//	    fmt.Println(string(msg.Body()))
//	    mgr.Ack(ctx, msg)
//	}
//
// 使用 RabbitMQ：
//
//	cfg := &mqmgr.RabbitMQConfig{
//	    URL:     "amqp://guest:guest@localhost:5672/",
//	    Durable: true,
//	}
//	mgr, err := mqmgr.NewMessageQueueManagerRabbitMQImpl(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer mgr.Close()
//
// 配置驱动类型：
//
//	// 通过 Build 函数创建
//	mgr, err := mqmgr.Build("memory", map[string]any{
//	    "max_queue_size": 10000,
//	    "channel_buffer": 100,
//	})
//
//	// 通过配置提供者创建
//	mgr, err := mqmgr.BuildWithConfigProvider(configProvider)
//
// 使用回调函数处理消息：
//
//	err := mgr.SubscribeWithCallback(ctx, "user_queue", func(ctx context.Context, msg mqmgr.Message) error {
//	    fmt.Println(string(msg.Body()))
//	    return nil
//	})
//
// 配置选项：
//
//	// 发布选项
//	mgr.Publish(ctx, "queue", []byte("data"),
//	    mqmgr.WithPublishHeaders(map[string]any{"key": "value"}),
//	    mqmgr.WithPublishDurable(true),
//	)
//
//	// 订阅选项
//	mgr.SubscribeWithCallback(ctx, "queue", handler,
//	    mqmgr.WithAutoAck(false),
//	    mqmgr.WithSubscribeDurable(true),
//	)
//
// 消息确认：
//
//	// 手动确认
//	msgCh, _ := mgr.Subscribe(ctx, "queue", mqmgr.WithAutoAck(false))
//	for msg := range msgCh {
//	    // 处理消息
//	    if err := process(msg); err != nil {
//	        mgr.Nack(ctx, msg, true) // 重新入队
//	    } else {
//	        mgr.Ack(ctx, msg) // 确认
//	    }
//	}
//
// 队列操作：
//
//	// 获取队列长度
//	length, err := mgr.QueueLength(ctx, "queue")
//
//	// 清空队列
//	err = mgr.Purge(ctx, "queue")
