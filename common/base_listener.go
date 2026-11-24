package common

import (
	"context"
)

// IMessageListener 消息监听器接口
// 定义消息队列相关的消息和订阅选项
// 此接口避免了与 manager 包的循环依赖
type IMessageListener interface {
	// ID 获取消息 ID
	ID() string
	// Body 获取消息体
	Body() []byte
	// Headers 获取消息头
	Headers() map[string]any
}

// ISubscribeOption 订阅选项接口
type ISubscribeOption interface{}

// IBaseListener 基础监听器接口
// 所有 Listener 类必须继承此接口并实现相关方法
// 用于定义监听器的基础行为和契约
type IBaseListener interface {
	// ListenerName 返回监听器名称
	// 格式：xxxListenerImpl（小驼峰，带 Impl 后缀）
	ListenerName() string

	// GetQueue 返回监听的队列名称
	// 返回值示例："message.created", "user.registered"
	GetQueue() string

	// GetSubscribeOptions 返回订阅选项
	// 可配置是否持久化、是否自动确认、并发消费者数量等
	GetSubscribeOptions() []ISubscribeOption

	// Handle 处理队列消息
	// ctx: 上下文
	// msg: 消息对象，包含 ID、Body、Headers
	// 返回: 处理错误（返回 error 会触发 Nack）
	Handle(ctx context.Context, msg IMessageListener) error

	// OnStart 在服务器启动时触发
	OnStart() error
	// OnStop 在服务器停止时触发
	OnStop() error
}
