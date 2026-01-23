package mqmgr

import (
	"context"
	"testing"
	"time"
)

func setupMemoryManager(t *testing.T) IMQManager {
	config := &MemoryConfig{
		MaxQueueSize:  1000,
		ChannelBuffer: 10,
	}
	return NewMessageQueueManagerMemoryImpl(config)
}

func TestMemoryManager_ManagerName(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	if mgr.ManagerName() != "messageQueueManagerMemoryImpl" {
		t.Errorf("expected name 'messageQueueManagerMemoryImpl', got '%s'", mgr.ManagerName())
	}
}

func TestMemoryManager_Health(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	if err := mgr.Health(); err != nil {
		t.Errorf("Health() should not return error, got %v", err)
	}
}

func TestMemoryManager_Lifecycle(t *testing.T) {
	mgr := setupMemoryManager(t)

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

func TestMemoryManager_Publish(t *testing.T) {
	mgr := setupMemoryManager(t)
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

func TestMemoryManager_PublishWithNilContext(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	err := mgr.Publish(nil, "test_queue", []byte("hello world"))
	if err == nil {
		t.Error("expected error with nil context, got nil")
	}
}

func TestMemoryManager_PublishWithEmptyQueue(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Publish(ctx, "", []byte("hello world"))
	if err == nil {
		t.Error("expected error with empty queue name, got nil")
	}
}

func TestMemoryManager_Subscribe(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	msgCh, err := mgr.Subscribe(ctx, "test_queue")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	err = mgr.Publish(ctx, "test_queue", []byte("hello"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
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

func TestMemoryManager_Ack(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	msgCh, err := mgr.Subscribe(ctx, "test_queue", WithAutoAck(false))
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	err = mgr.Publish(ctx, "test_queue", []byte("hello"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	select {
	case msg := <-msgCh:
		err := mgr.Ack(ctx, msg)
		if err != nil {
			t.Errorf("Ack() error = %v", err)
		}

		length, _ := mgr.QueueLength(ctx, "test_queue")
		if length != 0 {
			t.Errorf("expected queue length 0 after ack, got %d", length)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestMemoryManager_Nack(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	msgCh, err := mgr.Subscribe(ctx, "test_queue", WithAutoAck(false))
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	err = mgr.Publish(ctx, "test_queue", []byte("hello"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	select {
	case msg := <-msgCh:
		err := mgr.Nack(ctx, msg, false)
		if err != nil {
			t.Errorf("Nack() error = %v", err)
		}

		length, _ := mgr.QueueLength(ctx, "test_queue")
		if length != 0 {
			t.Errorf("expected queue length 0 after nack without requeue, got %d", length)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestMemoryManager_NackWithRequeue(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	msgCh, err := mgr.Subscribe(ctx, "test_queue", WithAutoAck(false))
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	err = mgr.Publish(ctx, "test_queue", []byte("hello"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
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

func TestMemoryManager_QueueLength(t *testing.T) {
	mgr := setupMemoryManager(t)
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

func TestMemoryManager_Purge(t *testing.T) {
	mgr := setupMemoryManager(t)
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

func TestMemoryManager_SubscribeWithCallback(t *testing.T) {
	mgr := setupMemoryManager(t)
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
