package mqmgr

import (
	"context"

	"github.com/lite-lake/litecore-go/common"
)

// IMQManager 消息队列管理器接口
type IMQManager interface {
	common.IBaseManager
	// Publish 发布消息到指定队列
	Publish(ctx context.Context, queue string, message []byte, options ...PublishOption) error

	// Subscribe 订阅指定队列，返回消息通道
	Subscribe(ctx context.Context, queue string, options ...SubscribeOption) (<-chan Message, error)

	// SubscribeWithCallback 使用回调函数订阅指定队列
	SubscribeWithCallback(ctx context.Context, queue string, handler MessageHandler, options ...SubscribeOption) error

	// Ack 确认消息已处理
	Ack(ctx context.Context, message Message) error

	// Nack 拒绝消息，可选择是否重新入队
	Nack(ctx context.Context, message Message, requeue bool) error

	// QueueLength 获取队列长度
	QueueLength(ctx context.Context, queue string) (int64, error)

	// Purge 清空队列
	Purge(ctx context.Context, queue string) error

	// Close 关闭管理器
	Close() error
}

// Message 消息接口
type Message interface {
	// ID 获取消息 ID
	ID() string
	// Body 获取消息体
	Body() []byte
	// Headers 获取消息头
	Headers() map[string]any
}

// MessageHandler 消息处理函数类型
type MessageHandler func(ctx context.Context, msg Message) error

// PublishOption 发布选项函数类型
type PublishOption func(*PublishOptions)

// PublishOptions 发布选项
type PublishOptions struct {
	// Headers 消息头
	Headers map[string]any
	// Durable 是否持久化
	Durable bool
}

// SubscribeOption 订阅选项函数类型
type SubscribeOption func(*SubscribeOptions)

// SubscribeOptions 订阅选项
type SubscribeOptions struct {
	// Durable 是否持久化
	Durable bool
	// AutoAck 是否自动确认
	AutoAck bool
}
