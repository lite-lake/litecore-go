package mqmgr

import (
	"context"

	"github.com/lite-lake/litecore-go/common"
)

// IMQManager 消息队列管理器接口
type IMQManager interface {
	common.IBaseManager
	Publish(ctx context.Context, queue string, message []byte, options ...PublishOption) error

	Subscribe(ctx context.Context, queue string, options ...SubscribeOption) (<-chan Message, error)

	SubscribeWithCallback(ctx context.Context, queue string, handler MessageHandler, options ...SubscribeOption) error

	Ack(ctx context.Context, message Message) error

	Nack(ctx context.Context, message Message, requeue bool) error

	QueueLength(ctx context.Context, queue string) (int64, error)

	Purge(ctx context.Context, queue string) error

	Close() error
}

type Message interface {
	ID() string
	Body() []byte
	Headers() map[string]any
}

type MessageHandler func(ctx context.Context, msg Message) error

type PublishOption func(*PublishOptions)

type PublishOptions struct {
	Headers map[string]any
	Durable bool
}

type SubscribeOption func(*SubscribeOptions)

type SubscribeOptions struct {
	Durable bool
	AutoAck bool
}
