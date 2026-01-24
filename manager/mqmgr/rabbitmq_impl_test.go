package mqmgr

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type mockRabbitMQConnection struct {
	closed bool
}

type mockRabbitMQChannel struct {
	closed bool
	queue  string
}

func setupRabbitMQManager(t *testing.T) IMQManager {
	config := &RabbitMQConfig{
		URL:     "amqp://guest:guest@localhost:5672/",
		Durable: true,
	}
	mgr, err := NewMessageQueueManagerRabbitMQImpl(config, nil, nil)
	if err != nil {
		t.Skip("RabbitMQ not available:", err)
	}
	return mgr
}

func TestRabbitMQManager_ManagerName(t *testing.T) {
	mgr := setupRabbitMQManager(t)
	defer mgr.Close()

	if mgr.ManagerName() != "messageQueueManagerRabbitMQImpl" {
		t.Errorf("expected name 'messageQueueManagerRabbitMQImpl', got '%s'", mgr.ManagerName())
	}
}

func TestRabbitMQManager_Health(t *testing.T) {
	mgr := setupRabbitMQManager(t)
	defer mgr.Close()

	if err := mgr.Health(); err != nil {
		t.Errorf("Health() should not return error, got %v", err)
	}
}

func TestRabbitMQManager_Lifecycle(t *testing.T) {
	mgr := setupRabbitMQManager(t)

	if err := mgr.OnStart(); err != nil {
		t.Errorf("OnStart() error = %v", err)
	}

	if err := mgr.OnStop(); err != nil {
		t.Errorf("OnStop() error = %v", err)
	}

	if err := mgr.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestRabbitMQManager_Publish(t *testing.T) {
	mgr := setupRabbitMQManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Publish(ctx, "test_queue", []byte("hello world"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	length, err := mgr.QueueLength(ctx, "test_queue")
	if err != nil {
		t.Fatalf("QueueLength() error = %v", err)
	}

	if length != 1 {
		t.Errorf("expected queue length 1, got %d", length)
	}
}

func TestRabbitMQManager_PublishWithHeaders(t *testing.T) {
	mgr := setupRabbitMQManager(t)
	defer mgr.Close()

	ctx := context.Background()

	headers := map[string]any{
		"key1": "value1",
		"key2": 123,
	}

	err := mgr.Publish(ctx, "test_queue", []byte("hello world"), WithPublishHeaders(headers))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
}

func TestRabbitMQManager_PublishWithNilContext(t *testing.T) {
	mgr := setupRabbitMQManager(t)
	defer mgr.Close()

	err := mgr.Publish(nil, "test_queue", []byte("hello world"))
	if err == nil {
		t.Error("expected error with nil context, got nil")
	}
}

func TestRabbitMQManager_PublishWithEmptyQueue(t *testing.T) {
	mgr := setupRabbitMQManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Publish(ctx, "", []byte("hello world"))
	if err == nil {
		t.Error("expected error with empty queue name, got nil")
	}
}

func TestRabbitMQManager_Subscribe(t *testing.T) {
	mgr := setupRabbitMQManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Publish(ctx, "test_queue", []byte("hello"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	msgCh, err := mgr.Subscribe(ctx, "test_queue")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	select {
	case msg := <-msgCh:
		if string(msg.Body()) != "hello" {
			t.Errorf("expected 'hello', got '%s'", string(msg.Body()))
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestRabbitMQManager_Ack(t *testing.T) {
	mgr := setupRabbitMQManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Publish(ctx, "test_queue", []byte("hello"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	msgCh, err := mgr.Subscribe(ctx, "test_queue", WithAutoAck(false))
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	select {
	case msg := <-msgCh:
		err := mgr.Ack(ctx, msg)
		if err != nil {
			t.Errorf("Ack() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestRabbitMQManager_Nack(t *testing.T) {
	mgr := setupRabbitMQManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Publish(ctx, "test_queue", []byte("hello"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	msgCh, err := mgr.Subscribe(ctx, "test_queue", WithAutoAck(false))
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	select {
	case msg := <-msgCh:
		err := mgr.Nack(ctx, msg, false)
		if err != nil {
			t.Errorf("Nack() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestRabbitMQManager_NackWithRequeue(t *testing.T) {
	mgr := setupRabbitMQManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Publish(ctx, "test_queue", []byte("hello"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	msgCh, err := mgr.Subscribe(ctx, "test_queue", WithAutoAck(false))
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	select {
	case msg := <-msgCh:
		err := mgr.Nack(ctx, msg, true)
		if err != nil {
			t.Errorf("Nack() error = %v", err)
		}

		length, _ := mgr.QueueLength(ctx, "test_queue")
		if length != 1 {
			t.Errorf("expected queue length 1 after nack with requeue, got %d", length)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestRabbitMQManager_QueueLength(t *testing.T) {
	mgr := setupRabbitMQManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Publish(ctx, "test_queue", []byte("msg1"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	err = mgr.Publish(ctx, "test_queue", []byte("msg2"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	length, err := mgr.QueueLength(ctx, "test_queue")
	if err != nil {
		t.Fatalf("QueueLength() error = %v", err)
	}

	if length != 2 {
		t.Errorf("expected queue length 2, got %d", length)
	}
}

func TestRabbitMQManager_Purge(t *testing.T) {
	mgr := setupRabbitMQManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Publish(ctx, "test_queue", []byte("msg1"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	err = mgr.Publish(ctx, "test_queue", []byte("msg2"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	err = mgr.Purge(ctx, "test_queue")
	if err != nil {
		t.Fatalf("Purge() error = %v", err)
	}

	length, _ := mgr.QueueLength(ctx, "test_queue")
	if length != 0 {
		t.Errorf("expected queue length 0 after purge, got %d", length)
	}
}

func TestRabbitMQManager_SubscribeWithCallback(t *testing.T) {
	mgr := setupRabbitMQManager(t)
	defer mgr.Close()

	ctx := context.Background()

	received := make(chan []byte, 1)

	err := mgr.SubscribeWithCallback(ctx, "test_queue", func(ctx context.Context, msg Message) error {
		received <- msg.Body()
		return nil
	})
	if err != nil {
		t.Fatalf("SubscribeWithCallback() error = %v", err)
	}

	err = mgr.Publish(ctx, "test_queue", []byte("hello"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	select {
	case body := <-received:
		if string(body) != "hello" {
			t.Errorf("expected 'hello', got '%s'", string(body))
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestRabbitMQManager_SubscribeWithCallbackError(t *testing.T) {
	mgr := setupRabbitMQManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.SubscribeWithCallback(ctx, "test_queue", func(ctx context.Context, msg Message) error {
		return fmt.Errorf("handler error")
	})
	if err != nil {
		t.Fatalf("SubscribeWithCallback() error = %v", err)
	}

	err = mgr.Publish(ctx, "test_queue", []byte("hello"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	time.Sleep(500 * time.Millisecond)
}

func TestRabbitMQManager_Close(t *testing.T) {
	mgr := setupRabbitMQManager(t)

	err := mgr.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	ctx := context.Background()
	err = mgr.Publish(ctx, "test_queue", []byte("hello"))
	if err == nil {
		t.Error("expected error after close, got nil")
	}
}
