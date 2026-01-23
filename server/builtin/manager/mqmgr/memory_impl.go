package mqmgr

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

// memoryMessage 内存消息
type memoryMessage struct {
	id      string
	queue   string
	body    []byte
	headers map[string]any
	acked   atomic.Bool
	nacked  atomic.Bool
}

// ID 获取消息 ID
func (m *memoryMessage) ID() string {
	return m.id
}

// Body 获取消息体
func (m *memoryMessage) Body() []byte {
	return m.body
}

// Headers 获取消息头
func (m *memoryMessage) Headers() map[string]any {
	return m.headers
}

// memoryQueue 内存队列
type memoryQueue struct {
	name        string
	messages    []*memoryMessage
	messagesMu  sync.RWMutex
	consumers   map[chan *memoryMessage]struct{}
	consumersMu sync.Mutex
	maxSize     int
	bufferSize  int
	deliveryTag atomic.Int64
}

// messageQueueManagerMemoryImpl 内存消息队列管理器实现
type messageQueueManagerMemoryImpl struct {
	*mqManagerBaseImpl
	queues   sync.Map
	name     string
	config   *MemoryConfig
	shutdown atomic.Bool
}

// NewMessageQueueManagerMemoryImpl 创建内存消息队列管理器
func NewMessageQueueManagerMemoryImpl(config *MemoryConfig) IMQManager {
	return &messageQueueManagerMemoryImpl{
		mqManagerBaseImpl: newMqManagerBaseImpl(),
		queues:            sync.Map{},
		name:              "messageQueueManagerMemoryImpl",
		config:            config,
	}
}

func (m *messageQueueManagerMemoryImpl) ManagerName() string {
	return m.name
}

func (m *messageQueueManagerMemoryImpl) Health() error {
	return nil
}

func (m *messageQueueManagerMemoryImpl) OnStart() error {
	return nil
}

func (m *messageQueueManagerMemoryImpl) OnStop() error {
	m.shutdown.Store(true)
	return nil
}

func (m *messageQueueManagerMemoryImpl) Publish(ctx context.Context, queue string, message []byte, options ...PublishOption) error {
	return m.recordOperation(ctx, "memory", "publish", queue, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateQueue(queue); err != nil {
			return err
		}

		if m.shutdown.Load() {
			return fmt.Errorf("manager is shutting down")
		}

		opts := &PublishOptions{}
		for _, opt := range options {
			opt(opts)
		}

		q := m.getOrCreateQueue(queue)

		q.messagesMu.Lock()

		if q.maxSize > 0 && len(q.messages) >= q.maxSize {
			q.messagesMu.Unlock()
			return fmt.Errorf("queue is full: %s", queue)
		}

		msg := &memoryMessage{
			id:      uuid.New().String(),
			queue:   queue,
			body:    message,
			headers: opts.Headers,
		}

		q.messages = append(q.messages, msg)
		q.messagesMu.Unlock()

		m.recordPublish(ctx, "memory")

		q.consumersMu.Lock()
		if len(q.consumers) > 0 {
			for ch := range q.consumers {
				select {
				case ch <- msg:
				default:
				}
			}
		}
		q.consumersMu.Unlock()

		return nil
	})
}

func (m *messageQueueManagerMemoryImpl) Subscribe(ctx context.Context, queue string, options ...SubscribeOption) (<-chan Message, error) {
	if err := ValidateContext(ctx); err != nil {
		return nil, err
	}
	if err := ValidateQueue(queue); err != nil {
		return nil, err
	}

	if m.shutdown.Load() {
		return nil, fmt.Errorf("manager is shutting down")
	}

	opts := &SubscribeOptions{AutoAck: true}
	for _, opt := range options {
		opt(opts)
	}

	q := m.getOrCreateQueue(queue)

	q.consumersMu.Lock()
	bufferSize := q.bufferSize
	if bufferSize == 0 {
		bufferSize = 100
	}
	ch := make(chan *memoryMessage, bufferSize)
	q.consumers[ch] = struct{}{}
	q.consumersMu.Unlock()

	messageCh := make(chan Message, bufferSize)

	go func() {
		defer func() {
			q.consumersMu.Lock()
			delete(q.consumers, ch)
			q.consumersMu.Unlock()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}

				m.recordConsume(ctx, "memory")

				if opts.AutoAck {
					msg.acked.Store(true)
					m.recordAck(ctx, "memory")
					m.removeMessage(q, msg)
				} else {
					m.removeMessage(q, msg)
				}

				func() {
					defer recover()
					messageCh <- msg
				}()
			}
		}
	}()

	q.messagesMu.Lock()
	messagesToSend := make([]*memoryMessage, len(q.messages))
	copy(messagesToSend, q.messages)
	q.messages = q.messages[:0]
	q.messagesMu.Unlock()

	for _, msg := range messagesToSend {
		select {
		case ch <- msg:
			m.recordConsume(ctx, "memory")
			if opts.AutoAck {
				msg.acked.Store(true)
				m.recordAck(ctx, "memory")
			}
			func() {
				defer recover()
				messageCh <- msg
			}()
		default:
			q.messagesMu.Lock()
			q.messages = append(q.messages, msg)
			q.messagesMu.Unlock()
		}
	}

	return messageCh, nil
}

func (m *messageQueueManagerMemoryImpl) SubscribeWithCallback(ctx context.Context, queue string, handler MessageHandler, options ...SubscribeOption) error {
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

func (m *messageQueueManagerMemoryImpl) Ack(ctx context.Context, message Message) error {
	if memMsg, ok := message.(*memoryMessage); ok {
		if memMsg.acked.CompareAndSwap(false, true) {
			memMsg.nacked.Store(false)
			m.recordAck(ctx, "memory")
		}
	}
	return nil
}

func (m *messageQueueManagerMemoryImpl) Nack(ctx context.Context, message Message, requeue bool) error {
	if memMsg, ok := message.(*memoryMessage); ok {
		if memMsg.nacked.CompareAndSwap(false, true) {
			memMsg.acked.Store(false)
			m.recordNack(ctx, "memory")

			if requeue {
				m.requeueMessage(ctx, memMsg.queue, memMsg)
			}
		}
	}
	return nil
}

func (m *messageQueueManagerMemoryImpl) QueueLength(ctx context.Context, queue string) (int64, error) {
	var result int64

	err := m.recordOperation(ctx, "memory", "queue_length", queue, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateQueue(queue); err != nil {
			return err
		}

		q, ok := m.queues.Load(queue)
		if !ok {
			result = 0
			return nil
		}

		mq := q.(*memoryQueue)
		mq.messagesMu.RLock()
		result = int64(len(mq.messages))
		mq.messagesMu.RUnlock()

		return nil
	})

	return result, err
}

func (m *messageQueueManagerMemoryImpl) Purge(ctx context.Context, queue string) error {
	return m.recordOperation(ctx, "memory", "purge", queue, func() error {
		if err := ValidateContext(ctx); err != nil {
			return err
		}
		if err := ValidateQueue(queue); err != nil {
			return err
		}

		q, ok := m.queues.Load(queue)
		if !ok {
			return nil
		}

		mq := q.(*memoryQueue)
		mq.messagesMu.Lock()
		mq.messages = nil
		mq.messagesMu.Unlock()

		return nil
	})
}

func (m *messageQueueManagerMemoryImpl) Close() error {
	m.shutdown.Store(true)
	m.queues.Range(func(key, value any) bool {
		mq := value.(*memoryQueue)
		mq.consumersMu.Lock()
		for ch := range mq.consumers {
			close(ch)
		}
		mq.consumers = nil
		mq.consumersMu.Unlock()
		return true
	})
	return nil
}

// getOrCreateQueue 获取或创建队列
func (m *messageQueueManagerMemoryImpl) getOrCreateQueue(queue string) *memoryQueue {
	q, _ := m.queues.LoadOrStore(queue, &memoryQueue{
		name:       queue,
		messages:   make([]*memoryMessage, 0),
		consumers:  make(map[chan *memoryMessage]struct{}),
		maxSize:    m.config.MaxQueueSize,
		bufferSize: m.config.ChannelBuffer,
	})
	return q.(*memoryQueue)
}

// removeMessage 从队列中移除消息
func (m *messageQueueManagerMemoryImpl) removeMessage(q *memoryQueue, msg *memoryMessage) {
	q.messagesMu.Lock()
	for i, m := range q.messages {
		if m == msg {
			q.messages = append(q.messages[:i], q.messages[i+1:]...)
			break
		}
	}
	q.messagesMu.Unlock()
}

// requeueMessage 重新入队消息
func (m *messageQueueManagerMemoryImpl) requeueMessage(ctx context.Context, queue string, msg *memoryMessage) {
	mq := m.getOrCreateQueue(queue)
	mq.messagesMu.Lock()
	mq.messages = append([]*memoryMessage{msg}, mq.messages...)
	mq.messagesMu.Unlock()
}

// removeMessageById 根据 ID 移除消息
func (m *messageQueueManagerMemoryImpl) removeMessageById(messageID string) {
	m.queues.Range(func(key, value any) bool {
		q := value.(*memoryQueue)
		q.messagesMu.Lock()
		for i, msg := range q.messages {
			if msg.id == messageID {
				q.messages = append(q.messages[:i], q.messages[i+1:]...)
				break
			}
		}
		q.messagesMu.Unlock()
		return true
	})
}

var _ IMQManager = (*messageQueueManagerMemoryImpl)(nil)
