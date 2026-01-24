package mqmgr

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/lite-lake/litecore-go/manager/loggermgr"
	"github.com/lite-lake/litecore-go/manager/telemetrymgr"
	amqp "github.com/rabbitmq/amqp091-go"
)

// rabbitMQMessage RabbitMQ 消息
type rabbitMQMessage struct {
	delivery *amqp.Delivery
}

// ID 获取消息 ID
func (m *rabbitMQMessage) ID() string {
	return fmt.Sprintf("%d", m.delivery.DeliveryTag)
}

// Body 获取消息体
func (m *rabbitMQMessage) Body() []byte {
	return m.delivery.Body
}

// Headers 获取消息头
func (m *rabbitMQMessage) Headers() map[string]any {
	return m.delivery.Headers
}

// messageQueueManagerRabbitMQImpl RabbitMQ 消息队列管理器实现
type messageQueueManagerRabbitMQImpl struct {
	*mqManagerBaseImpl
	conn     *amqp.Connection
	chanMu   sync.RWMutex
	channels map[string]*amqp.Channel
	name     string
	config   *RabbitMQConfig
	closed   atomic.Bool
}

// NewMessageQueueManagerRabbitMQImpl 创建 RabbitMQ 消息队列管理器
// 参数：
//   - config: RabbitMQ 配置
//   - loggerMgr: 日志管理器
//   - telemetryMgr: 遥测管理器
//
// 返回 IMQManager 接口实例和可能的错误
func NewMessageQueueManagerRabbitMQImpl(
	config *RabbitMQConfig,
	loggerMgr loggermgr.ILoggerManager,
	telemetryMgr telemetrymgr.ITelemetryManager,
) (IMQManager, error) {
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	impl := &messageQueueManagerRabbitMQImpl{
		mqManagerBaseImpl: newMqManagerBaseImpl(loggerMgr, telemetryMgr),
		conn:              conn,
		channels:          make(map[string]*amqp.Channel),
		name:              "messageQueueManagerRabbitMQImpl",
		config:            config,
	}
	impl.initObservability()

	return impl, nil
}

func (m *messageQueueManagerRabbitMQImpl) ManagerName() string {
	return m.name
}

func (m *messageQueueManagerRabbitMQImpl) Health() error {
	if m.closed.Load() {
		return fmt.Errorf("connection is closed")
	}
	if m.conn == nil || m.conn.IsClosed() {
		return fmt.Errorf("connection is not established")
	}
	return nil
}

func (m *messageQueueManagerRabbitMQImpl) OnStart() error {
	return nil
}

func (m *messageQueueManagerRabbitMQImpl) OnStop() error {
	return m.Close()
}

func (m *messageQueueManagerRabbitMQImpl) Publish(ctx context.Context, queue string, message []byte, options ...PublishOption) error {
	return m.recordOperation(ctx, "rabbitmq", "publish", queue, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateQueue(queue); err != nil {
			return err
		}

		ch, err := m.getChannel(queue)
		if err != nil {
			return err
		}

		_, err = ch.QueueDeclare(
			queue,
			m.config.Durable,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue: %w", err)
		}

		opts := &PublishOptions{}
		for _, opt := range options {
			opt(opts)
		}

		publishing := amqp.Publishing{
			ContentType:  "text/plain",
			Body:         message,
			DeliveryMode: amqp.Persistent,
			Headers:      opts.Headers,
		}

		if !opts.Durable {
			publishing.DeliveryMode = amqp.Transient
		}

		err = ch.PublishWithContext(
			ctx,
			"",
			queue,
			false,
			false,
			publishing,
		)
		if err != nil {
			return fmt.Errorf("failed to publish message: %w", err)
		}

		m.recordPublish(ctx, "rabbitmq")
		return nil
	})
}

func (m *messageQueueManagerRabbitMQImpl) Subscribe(ctx context.Context, queue string, options ...SubscribeOption) (<-chan Message, error) {
	if err := ValidateContext(ctx); err != nil {
		return nil, err
	}
	if err := ValidateQueue(queue); err != nil {
		return nil, err
	}

	opts := &SubscribeOptions{AutoAck: true}
	for _, opt := range options {
		opt(opts)
	}

	ch, err := m.getChannel(queue)
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(
		queue,
		m.config.Durable,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	msgs, err := ch.Consume(
		queue,
		"",
		opts.AutoAck,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register consumer: %w", err)
	}

	messageCh := make(chan Message, 100)

	go func() {
		defer close(messageCh)

		for {
			select {
			case <-ctx.Done():
				return
			case delivery, ok := <-msgs:
				if !ok {
					return
				}

				msg := &rabbitMQMessage{delivery: &delivery}
				m.recordConsume(ctx, "rabbitmq")

				if opts.AutoAck {
					m.recordAck(ctx, "rabbitmq")
				}

				messageCh <- msg
			}
		}
	}()

	return messageCh, nil
}

func (m *messageQueueManagerRabbitMQImpl) SubscribeWithCallback(ctx context.Context, queue string, handler MessageHandler, options ...SubscribeOption) error {
	msgCh, err := m.Subscribe(ctx, queue, options...)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgCh:
				if !ok {
					return
				}

				if err := handler(ctx, msg); err != nil {
					m.Nack(ctx, msg, false)
				} else {
					m.Ack(ctx, msg)
				}
			}
		}
	}()

	return nil
}

func (m *messageQueueManagerRabbitMQImpl) Ack(ctx context.Context, message Message) error {
	if rabbitMsg, ok := message.(*rabbitMQMessage); ok {
		err := rabbitMsg.delivery.Ack(false)
		if err != nil {
			return fmt.Errorf("failed to ack message: %w", err)
		}
		m.recordAck(ctx, "rabbitmq")
	}
	return nil
}

func (m *messageQueueManagerRabbitMQImpl) Nack(ctx context.Context, message Message, requeue bool) error {
	if rabbitMsg, ok := message.(*rabbitMQMessage); ok {
		err := rabbitMsg.delivery.Nack(false, requeue)
		if err != nil {
			return fmt.Errorf("failed to nack message: %w", err)
		}
		m.recordNack(ctx, "rabbitmq")
	}
	return nil
}

func (m *messageQueueManagerRabbitMQImpl) QueueLength(ctx context.Context, queue string) (int64, error) {
	var result int64

	err := m.recordOperation(ctx, "rabbitmq", "queue_length", queue, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateQueue(queue); err != nil {
			return err
		}

		ch, err := m.getChannel(queue)
		if err != nil {
			return err
		}

		_, err = ch.QueueDeclare(
			queue,
			m.config.Durable,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue: %w", err)
		}

		info, err := ch.QueueInspect(queue)
		if err != nil {
			return fmt.Errorf("failed to inspect queue: %w", err)
		}

		result = int64(info.Messages)
		return nil
	})

	return result, err
}

func (m *messageQueueManagerRabbitMQImpl) Purge(ctx context.Context, queue string) error {
	return m.recordOperation(ctx, "rabbitmq", "purge", queue, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateQueue(queue); err != nil {
			return err
		}

		ch, err := m.getChannel(queue)
		if err != nil {
			return err
		}

		_, err = ch.QueuePurge(queue, false)
		if err != nil {
			return fmt.Errorf("failed to purge queue: %w", err)
		}

		return nil
	})
}

func (m *messageQueueManagerRabbitMQImpl) Close() error {
	if m.closed.CompareAndSwap(false, true) {
		m.chanMu.Lock()
		for queue, ch := range m.channels {
			ch.Close()
			delete(m.channels, queue)
		}
		m.chanMu.Unlock()

		if m.conn != nil && !m.conn.IsClosed() {
			return m.conn.Close()
		}
	}
	return nil
}

// getChannel 获取或创建通道
func (m *messageQueueManagerRabbitMQImpl) getChannel(queue string) (*amqp.Channel, error) {
	m.chanMu.RLock()
	ch, ok := m.channels[queue]
	m.chanMu.RUnlock()

	if ok && ch != nil && !ch.IsClosed() {
		return ch, nil
	}

	m.chanMu.Lock()
	defer m.chanMu.Unlock()

	if ch, ok := m.channels[queue]; ok && ch != nil && !ch.IsClosed() {
		return ch, nil
	}

	newCh, err := m.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	m.channels[queue] = newCh
	return newCh, nil
}

var _ IMQManager = (*messageQueueManagerRabbitMQImpl)(nil)
