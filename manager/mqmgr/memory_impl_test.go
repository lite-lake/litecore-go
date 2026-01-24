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
	return NewMessageQueueManagerMemoryImpl(config, nil, nil)
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
	mgr := NewMessageQueueManagerMemoryImpl(config, nil, nil)
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

func TestMemoryManager_RemoveMessageById(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Publish(ctx, "test_queue", []byte("msg1"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	msgCh, err := mgr.Subscribe(ctx, "test_queue", WithAutoAck(false))
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	var messageID string
	select {
	case msg := <-msgCh:
		messageID = msg.ID()
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message")
	}

	if mgrImpl, ok := mgr.(*messageQueueManagerMemoryImpl); ok {
		mgrImpl.removeMessageById(messageID)

		length, _ := mgr.QueueLength(ctx, "test_queue")
		if length != 0 {
			t.Errorf("expected queue length 0 after removing message by ID, got %d", length)
		}
	}
}

func TestMemoryManager_SubscribeContextCancellation(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgCh, err := mgr.Subscribe(ctx, "test_queue")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	cancel()

	select {
	case _, ok := <-msgCh:
		if ok {
			t.Error("expected channel to be closed after context cancellation")
		}
	case <-time.After(100 * time.Millisecond):
	}
}

func TestMemoryManager_MultipleQueues(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Publish(ctx, "queue1", []byte("msg1"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	err = mgr.Publish(ctx, "queue2", []byte("msg2"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	length1, _ := mgr.QueueLength(ctx, "queue1")
	if length1 != 1 {
		t.Errorf("expected queue1 length 1, got %d", length1)
	}

	length2, _ := mgr.QueueLength(ctx, "queue2")
	if length2 != 1 {
		t.Errorf("expected queue2 length 1, got %d", length2)
	}

	err = mgr.Purge(ctx, "queue1")
	if err != nil {
		t.Fatalf("Purge() error = %v", err)
	}

	length1, _ = mgr.QueueLength(ctx, "queue1")
	if length1 != 0 {
		t.Errorf("expected queue1 length 0 after purge, got %d", length1)
	}

	length2, _ = mgr.QueueLength(ctx, "queue2")
	if length2 != 1 {
		t.Errorf("expected queue2 length 1 after purge of queue1, got %d", length2)
	}
}

func TestMemoryManager_NilHeaders(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Publish(ctx, "test_queue", []byte("msg"), WithPublishHeaders(nil))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	msgCh, err := mgr.Subscribe(ctx, "test_queue")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	select {
	case msg := <-msgCh:
		if msg.Headers() != nil {
			t.Errorf("expected nil headers, got %v", msg.Headers())
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestMemoryManager_DuplicateNack(t *testing.T) {
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
			t.Errorf("first Nack() error = %v", err)
		}

		err = mgr.Nack(ctx, msg, false)
		if err != nil {
			t.Errorf("second Nack() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestMemoryManager_AckAfterNack(t *testing.T) {
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

		err = mgr.Ack(ctx, msg)
		if err != nil {
			t.Errorf("Ack() after Nack() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestMemoryManager_NackAfterAck(t *testing.T) {
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

		err = mgr.Nack(ctx, msg, false)
		if err != nil {
			t.Errorf("Nack() after Ack() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestMemoryManager_ZeroChannelBuffer(t *testing.T) {
	config := &MemoryConfig{
		MaxQueueSize:  1000,
		ChannelBuffer: 0,
	}
	mgr := NewMessageQueueManagerMemoryImpl(config, nil, nil)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Publish(ctx, "test_queue", []byte("msg1"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	msgCh, err := mgr.Subscribe(ctx, "test_queue")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	select {
	case msg := <-msgCh:
		if string(msg.Body()) != "msg1" {
			t.Errorf("expected 'msg1', got '%s'", string(msg.Body()))
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestMemoryManager_EmptyMessage(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Publish(ctx, "test_queue", []byte{})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	msgCh, err := mgr.Subscribe(ctx, "test_queue")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	select {
	case msg := <-msgCh:
		if len(msg.Body()) != 0 {
			t.Errorf("expected empty body, got %v", msg.Body())
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestMemoryManager_NonExistentQueueLength(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	length, err := mgr.QueueLength(ctx, "non_existent_queue")
	if err != nil {
		t.Errorf("QueueLength() error = %v", err)
	}

	if length != 0 {
		t.Errorf("expected queue length 0 for non-existent queue, got %d", length)
	}
}

func TestMemoryManager_InvalidMessageNack(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Nack(ctx, nil, false)
	if err != nil {
		t.Errorf("Nack() with nil message should not return error, got %v", err)
	}

	err = mgr.Ack(ctx, nil)
	if err != nil {
		t.Errorf("Ack() with nil message should not return error, got %v", err)
	}
}

func TestMemoryManager_SubscribeWithDurable(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	msgCh, err := mgr.Subscribe(ctx, "test_queue", WithSubscribeDurable(true))
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

func TestMemoryManager_LargeMessage(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	largeBody := make([]byte, 10000)
	for i := range largeBody {
		largeBody[i] = byte(i % 256)
	}

	err := mgr.Publish(ctx, "test_queue", largeBody)
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	msgCh, err := mgr.Subscribe(ctx, "test_queue")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	select {
	case msg := <-msgCh:
		if len(msg.Body()) != 10000 {
			t.Errorf("expected body length 10000, got %d", len(msg.Body()))
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestMemoryManager_RapidPublish(t *testing.T) {
	config := &MemoryConfig{
		MaxQueueSize:  10000,
		ChannelBuffer: 1000,
	}
	mgr := NewMessageQueueManagerMemoryImpl(config, nil, nil)
	defer mgr.Close()

	ctx := context.Background()

	const numMessages = 1000
	for i := 0; i < numMessages; i++ {
		err := mgr.Publish(ctx, "test_queue", []byte(fmt.Sprintf("msg%d", i)))
		if err != nil {
			t.Fatalf("Publish() error = %v", err)
		}
	}

	length, err := mgr.QueueLength(ctx, "test_queue")
	if err != nil {
		t.Fatalf("QueueLength() error = %v", err)
	}

	if length != numMessages {
		t.Errorf("expected queue length %d, got %d", numMessages, length)
	}
}

func TestMemoryManager_ConcurrentAccess(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			queueName := fmt.Sprintf("queue%d", idx)
			for j := 0; j < 10; j++ {
				err := mgr.Publish(ctx, queueName, []byte(fmt.Sprintf("msg%d", j)))
				if err != nil {
					t.Errorf("Publish() error = %v", err)
				}
			}

			length, err := mgr.QueueLength(ctx, queueName)
			if err != nil {
				t.Errorf("QueueLength() error = %v", err)
			}

			if length != 10 {
				t.Errorf("expected queue length 10, got %d", length)
			}
		}(i)
	}

	wg.Wait()
}

func TestMemoryManager_RepeatedSubscribe(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	for i := 0; i < 5; i++ {
		msgCh, err := mgr.Subscribe(ctx, "test_queue")
		if err != nil {
			t.Fatalf("Subscribe() error = %v", err)
		}

		go func(ch <-chan Message) {
			for range ch {
			}
		}(msgCh)
	}

	err := mgr.Publish(ctx, "test_queue", []byte("msg1"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	time.Sleep(500 * time.Millisecond)
}

func TestMemoryManager_PublishAfterClose(t *testing.T) {
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
}

func TestMemoryManager_SubscribeAfterClose(t *testing.T) {
	mgr := setupMemoryManager(t)
	ctx := context.Background()

	err := mgr.Close()
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	_, err = mgr.Subscribe(ctx, "test_queue")
	if err == nil {
		t.Error("expected error after close, got nil")
	}
}

func TestMemoryManager_AckWithInvalidType(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	invalidMsg := &invalidMessage{}
	err := mgr.Ack(ctx, invalidMsg)
	if err != nil {
		t.Error("Ack() with invalid message type should not error")
	}
}

func TestMemoryManager_NackWithInvalidType(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	invalidMsg := &invalidMessage{}
	err := mgr.Nack(ctx, invalidMsg, false)
	if err != nil {
		t.Error("Nack() with invalid message type should not error")
	}
}

type invalidMessage struct{}

func (m *invalidMessage) ID() string              { return "invalid" }
func (m *invalidMessage) Body() []byte            { return []byte("invalid") }
func (m *invalidMessage) Headers() map[string]any { return nil }

func TestMemoryManager_PurgeWithNilContext(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	err := mgr.Purge(nil, "test_queue")
	if err == nil {
		t.Error("expected error with nil context, got nil")
	}
}

func TestMemoryManager_PurgeWithEmptyQueue(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Purge(ctx, "")
	if err == nil {
		t.Error("expected error with empty queue name, got nil")
	}
}

func TestMemoryManager_QueueLengthWithNilContext(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	_, err := mgr.QueueLength(nil, "test_queue")
	if err == nil {
		t.Error("expected error with nil context, got nil")
	}
}

func TestMemoryManager_QueueLengthWithEmptyQueue(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	_, err := mgr.QueueLength(ctx, "")
	if err == nil {
		t.Error("expected error with empty queue name, got nil")
	}
}

func TestMemoryManager_SubscribeWithNilContext(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	_, err := mgr.Subscribe(nil, "test_queue")
	if err == nil {
		t.Error("expected error with nil context, got nil")
	}
}

func TestMemoryManager_SubscribeWithEmptyQueue(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	_, err := mgr.Subscribe(ctx, "")
	if err == nil {
		t.Error("expected error with empty queue name, got nil")
	}
}

func TestMemoryManager_SubscribeWithCallbackNilContext(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	err := mgr.SubscribeWithCallback(nil, "test_queue", func(ctx context.Context, msg Message) error {
		return nil
	})
	if err == nil {
		t.Error("expected error with nil context, got nil")
	}
}

func TestMemoryManager_SubscribeWithCallbackEmptyQueue(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.SubscribeWithCallback(ctx, "", func(ctx context.Context, msg Message) error {
		return nil
	})
	if err == nil {
		t.Error("expected error with empty queue name, got nil")
	}
}

func TestMemoryManager_SubscribeWithCallbackNilHandler(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.SubscribeWithCallback(ctx, "test_queue", nil)
	if err != nil {
		t.Errorf("expected no error with nil handler, got %v", err)
	}
}

func TestMemoryManager_SubscribeChannelFull(t *testing.T) {
	config := &MemoryConfig{
		MaxQueueSize:  1000,
		ChannelBuffer: 1,
	}
	mgr := NewMessageQueueManagerMemoryImpl(config, nil, nil)
	defer mgr.Close()

	ctx := context.Background()

	msgCh, err := mgr.Subscribe(ctx, "test_queue", WithAutoAck(false))
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	for i := 0; i < 10; i++ {
		err := mgr.Publish(ctx, "test_queue", []byte(fmt.Sprintf("msg%d", i)))
		if err != nil {
			t.Fatalf("Publish() error = %v", err)
		}
	}

	count := 0
	for i := 0; i < 10; i++ {
		select {
		case <-msgCh:
			count++
		case <-time.After(100 * time.Millisecond):
			break
		}
	}

	if count < 1 {
		t.Errorf("expected at least 1 message, got %d", count)
	}
}

func TestMemoryManager_PublishToMultipleQueues(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	queues := []string{"queue1", "queue2", "queue3"}
	for _, queue := range queues {
		err := mgr.Publish(ctx, queue, []byte(fmt.Sprintf("msg for %s", queue)))
		if err != nil {
			t.Fatalf("Publish() error = %v", err)
		}
	}

	for _, queue := range queues {
		length, err := mgr.QueueLength(ctx, queue)
		if err != nil {
			t.Fatalf("QueueLength() error = %v", err)
		}
		if length != 1 {
			t.Errorf("expected queue %s length 1, got %d", queue, length)
		}
	}
}

func TestMemoryManager_RemoveNonExistentMessage(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	if mgrImpl, ok := mgr.(*messageQueueManagerMemoryImpl); ok {
		mgrImpl.removeMessageById("non-existent-id")
	}
}

func TestMemoryManager_PublishAndConsumeMultipleQueues(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	queues := []string{"queue1", "queue2", "queue3"}
	for _, queue := range queues {
		for i := 0; i < 5; i++ {
			err := mgr.Publish(ctx, queue, []byte(fmt.Sprintf("%s-msg%d", queue, i)))
			if err != nil {
				t.Fatalf("Publish() error = %v", err)
			}
		}
	}

	for _, queue := range queues {
		length, err := mgr.QueueLength(ctx, queue)
		if err != nil {
			t.Fatalf("QueueLength() error = %v", err)
		}
		if length != 5 {
			t.Errorf("expected queue %s length 5, got %d", queue, length)
		}
	}
}

func TestMemoryManager_RapidPublishAndSubscribe(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	const numMessages = 20

	for i := 0; i < numMessages; i++ {
		err := mgr.Publish(ctx, "test_queue", []byte(fmt.Sprintf("msg%d", i)))
		if err != nil {
			t.Fatalf("Publish() error = %v", err)
		}
	}

	msgCh, err := mgr.Subscribe(ctx, "test_queue")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	count := 0
	timeout := time.After(2 * time.Second)
	for {
		select {
		case <-msgCh:
			count++
			if count >= numMessages {
				return
			}
		case <-timeout:
			if count < numMessages {
				t.Errorf("expected at least %d messages, got %d", numMessages, count)
			}
			return
		}
	}
}

func TestMemoryManager_PublishAndSubscribeWithDifferentOptions(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	tests := []struct {
		name          string
		publishOpts   []PublishOption
		subscribeOpts []SubscribeOption
	}{
		{"默认选项", nil, nil},
		{"带 headers", []PublishOption{WithPublishHeaders(map[string]any{"key": "value"})}, nil},
		{"非持久化", []PublishOption{WithPublishDurable(false)}, nil},
		{"非自动确认", nil, []SubscribeOption{WithAutoAck(false)}},
		{"所有选项", []PublishOption{WithPublishHeaders(map[string]any{"key": "value"}), WithPublishDurable(false)}, []SubscribeOption{WithAutoAck(false)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgCh, err := mgr.Subscribe(ctx, tt.name, tt.subscribeOpts...)
			if err != nil {
				t.Fatalf("Subscribe() error = %v", err)
			}

			err = mgr.Publish(ctx, tt.name, []byte("test"), tt.publishOpts...)
			if err != nil {
				t.Fatalf("Publish() error = %v", err)
			}

			select {
			case msg := <-msgCh:
				if string(msg.Body()) != "test" {
					t.Errorf("expected 'test', got '%s'", string(msg.Body()))
				}
			case <-time.After(2 * time.Second):
				t.Error("timeout waiting for message")
			}
		})
	}
}

func TestMemoryManager_ImmediatePublishAndSubscribe(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	msgCh, err := mgr.Subscribe(ctx, "test_queue")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	err = mgr.Publish(ctx, "test_queue", []byte("immediate"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	select {
	case msg := <-msgCh:
		if string(msg.Body()) != "immediate" {
			t.Errorf("expected 'immediate', got '%s'", string(msg.Body()))
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("timeout waiting for message")
	}
}

func TestMemoryManager_MultipleMessagesWithHeaders(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	headers1 := map[string]any{"type": "email", "priority": 1}
	headers2 := map[string]any{"type": "sms", "priority": 2}

	err := mgr.Publish(ctx, "test_queue", []byte("msg1"), WithPublishHeaders(headers1))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	err = mgr.Publish(ctx, "test_queue", []byte("msg2"), WithPublishHeaders(headers2))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	msgCh, err := mgr.Subscribe(ctx, "test_queue")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	messages := make([]Message, 0, 2)
	for i := 0; i < 2; i++ {
		select {
		case msg := <-msgCh:
			messages = append(messages, msg)
		case <-time.After(2 * time.Second):
			t.Error("timeout waiting for message")
		}
	}

	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}
}

func TestMemoryManager_RequeueMessage(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Publish(ctx, "test_queue", []byte("requeue"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	msgCh, err := mgr.Subscribe(ctx, "test_queue", WithAutoAck(false))
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	var msg Message
	select {
	case msg = <-msgCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message")
	}

	err = mgr.Nack(ctx, msg, true)
	if err != nil {
		t.Errorf("Nack() error = %v", err)
	}

	length, _ := mgr.QueueLength(ctx, "test_queue")
	if length != 1 {
		t.Errorf("expected queue length 1 after requeue, got %d", length)
	}
}

func TestMemoryManager_SubscribeWithCallbackHandlerPanic(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.SubscribeWithCallback(ctx, "test_queue", func(ctx context.Context, msg Message) error {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Handler panic recovered: %v", r)
			}
		}()
		panic("handler panic")
	})
	if err != nil {
		t.Fatalf("SubscribeWithCallback() error = %v", err)
	}

	err = mgr.Publish(ctx, "test_queue", []byte("panic test"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	time.Sleep(500 * time.Millisecond)
}

func TestMemoryManager_SubscribeChannelSendPanic(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	msgCh, err := mgr.Subscribe(ctx, "test_queue")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	for i := 0; i < 1000; i++ {
		err := mgr.Publish(ctx, "test_queue", []byte(fmt.Sprintf("msg%d", i)))
		if err != nil {
			t.Fatalf("Publish() error = %v", err)
		}
	}

	count := 0
	for i := 0; i < 10; i++ {
		select {
		case <-msgCh:
			count++
		case <-time.After(100 * time.Millisecond):
			break
		}
	}

	if count < 10 {
		t.Logf("received %d messages (may be expected due to channel buffer)", count)
	}
}

func TestMemoryManager_RemoveMessageByIdFromMultipleQueues(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	if mgrImpl, ok := mgr.(*messageQueueManagerMemoryImpl); ok {
		err := mgr.Publish(ctx, "queue1", []byte("msg1"))
		if err != nil {
			t.Fatalf("Publish() error = %v", err)
		}

		err = mgr.Publish(ctx, "queue2", []byte("msg2"))
		if err != nil {
			t.Fatalf("Publish() error = %v", err)
		}

		msgCh1, err := mgr.Subscribe(ctx, "queue1", WithAutoAck(false))
		if err != nil {
			t.Fatalf("Subscribe() error = %v", err)
		}

		var msg1 Message
		select {
		case msg1 = <-msgCh1:
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for message")
		}

		mgrImpl.removeMessageById(msg1.ID())

		length, _ := mgr.QueueLength(ctx, "queue1")
		if length != 0 {
			t.Errorf("expected queue1 length 0, got %d", length)
		}

		length2, _ := mgr.QueueLength(ctx, "queue2")
		if length2 != 1 {
			t.Errorf("expected queue2 length 1, got %d", length2)
		}
	}
}
