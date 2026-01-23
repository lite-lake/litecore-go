package mqmgr

import (
	"context"
	"fmt"
	"sync"
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

func TestMemoryManager_PublishWithHeaders(t *testing.T) {
	mgr := setupMemoryManager(t)
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

	msgCh, err := mgr.Subscribe(ctx, "test_queue")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	select {
	case msg := <-msgCh:
		if msg.Headers()["key1"] != "value1" {
			t.Errorf("expected key1='value1', got %v", msg.Headers()["key1"])
		}
		if msg.Headers()["key2"] != 123 {
			t.Errorf("expected key2=123, got %v", msg.Headers()["key2"])
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestMemoryManager_PublishWithDurable(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Publish(ctx, "test_queue", []byte("hello world"), WithPublishDurable(true))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
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

func TestMemoryManager_SubscribeWithOptions(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	msgCh, err := mgr.Subscribe(ctx, "test_queue", WithSubscribeDurable(true), WithAutoAck(false))
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

func TestMemoryManager_MultipleConsumers(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	var wg sync.WaitGroup
	receivedCount := 0
	var mu sync.Mutex

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msgCh, err := mgr.Subscribe(ctx, "test_queue")
			if err != nil {
				t.Errorf("Subscribe() error = %v", err)
				return
			}

			select {
			case <-msgCh:
				mu.Lock()
				receivedCount++
				mu.Unlock()
			case <-time.After(2 * time.Second):
				t.Error("timeout waiting for message")
			}
		}()
	}

	time.Sleep(100 * time.Millisecond)
	err := mgr.Publish(ctx, "test_queue", []byte("hello"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	wg.Wait()

	if receivedCount != 3 {
		t.Errorf("expected 3 consumers to receive message, got %d", receivedCount)
	}
}

func TestMemoryManager_OperationsAfterClose(t *testing.T) {
	mgr := setupMemoryManager(t)
	ctx := context.Background()

	err := mgr.Close()
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	err = mgr.Publish(ctx, "test_queue", []byte("hello"))
	if err == nil {
		t.Error("expected error after close, got nil")
	}

	_, err = mgr.Subscribe(ctx, "test_queue")
	if err == nil {
		t.Error("expected error after close, got nil")
	}
}

func TestMemoryManager_QueueFull(t *testing.T) {
	config := &MemoryConfig{
		MaxQueueSize:  2,
		ChannelBuffer: 10,
	}
	mgr := NewMessageQueueManagerMemoryImpl(config)
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

	err = mgr.Publish(ctx, "test_queue", []byte("msg3"))
	if err == nil {
		t.Error("expected error when queue is full, got nil")
	}
}

func TestMemoryManager_ConcurrentPublish(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()
	const numMessages = 100
	var wg sync.WaitGroup

	for i := 0; i < numMessages; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			err := mgr.Publish(ctx, "test_queue", []byte(fmt.Sprintf("msg%d", idx)))
			if err != nil {
				t.Errorf("Publish() error = %v", err)
			}
		}(i)
	}

	wg.Wait()

	length, err := mgr.QueueLength(ctx, "test_queue")
	if err != nil {
		t.Fatalf("QueueLength() error = %v", err)
	}

	if length != numMessages {
		t.Errorf("expected queue length %d, got %d", numMessages, length)
	}
}

func TestMemoryManager_ConcurrentSubscribe(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()
	const numConsumers = 10
	const numMessages = 50

	var wg sync.WaitGroup
	receivedCounts := make([]int, numConsumers)
	var mu sync.Mutex

	for i := 0; i < numConsumers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			msgCh, err := mgr.Subscribe(ctx, "test_queue")
			if err != nil {
				t.Errorf("Subscribe() error = %v", err)
				return
			}

			count := 0
			timeout := time.After(2 * time.Second)
			for {
				select {
				case <-msgCh:
					count++
					mu.Lock()
					receivedCounts[idx] = count
					mu.Unlock()
				case <-timeout:
					return
				}
			}
		}(i)
	}

	time.Sleep(100 * time.Millisecond)

	for i := 0; i < numMessages; i++ {
		err := mgr.Publish(ctx, "test_queue", []byte(fmt.Sprintf("msg%d", i)))
		if err != nil {
			t.Fatalf("Publish() error = %v", err)
		}
	}

	wg.Wait()

	totalReceived := 0
	for _, count := range receivedCounts {
		totalReceived += count
	}

	if totalReceived < numMessages {
		t.Errorf("expected at least %d messages received, got %d", numMessages, totalReceived)
	}
}

func TestMemoryManager_MessageID(t *testing.T) {
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
		if msg.ID() == "" {
			t.Error("expected non-empty message ID")
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestMemoryManager_DuplicateAck(t *testing.T) {
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
			t.Errorf("first Ack() error = %v", err)
		}

		err = mgr.Ack(ctx, msg)
		if err != nil {
			t.Errorf("second Ack() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestMemoryManager_PurgeNonExistentQueue(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Purge(ctx, "non_existent_queue")
	if err != nil {
		t.Errorf("Purge() on non-existent queue should not return error, got %v", err)
	}
}

func TestMemoryManager_QueueLengthNonExistentQueue(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	length, err := mgr.QueueLength(ctx, "non_existent_queue")
	if err != nil {
		t.Errorf("QueueLength() on non-existent queue should not return error, got %v", err)
	}

	if length != 0 {
		t.Errorf("expected queue length 0 for non-existent queue, got %d", length)
	}
}

func TestMemoryManager_SubscribeWithCallbackError(t *testing.T) {
	mgr := setupMemoryManager(t)
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

func TestMemoryManager_ExistingMessages(t *testing.T) {
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

	msgCh, err := mgr.Subscribe(ctx, "test_queue")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	received := make([]string, 0, 2)
	for i := 0; i < 2; i++ {
		select {
		case msg := <-msgCh:
			received = append(received, string(msg.Body()))
		case <-time.After(2 * time.Second):
			t.Error("timeout waiting for message")
		}
	}

	if len(received) != 2 {
		t.Errorf("expected 2 messages, got %d", len(received))
	}

	length, _ := mgr.QueueLength(ctx, "test_queue")
	if length != 0 {
		t.Errorf("expected queue length 0 after consuming all messages, got %d", length)
	}
}
