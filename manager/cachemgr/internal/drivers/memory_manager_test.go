package drivers

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewMemoryManager(t *testing.T) {
	mgr := NewMemoryManager(5*time.Minute, 10*time.Minute)

	if mgr.ManagerName() != "memory-cache" {
		t.Errorf("ManagerName() = %v, want %v", mgr.ManagerName(), "memory-cache")
	}
}

func TestMemoryManager_BasicOperations(t *testing.T) {
	mgr := NewMemoryManager(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	// 测试 Set 和 Get
	err := mgr.Set(ctx, "key1", "value1", time.Minute)
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	// 使用 GetAny 获取
	value, found := mgr.GetAny(ctx, "key1")
	if !found {
		t.Error("GetAny() should find the key")
	}
	if value != "value1" {
		t.Errorf("GetAny() = %v, want %v", value, "value1")
	}

	// 测试 Exists
	exists, err := mgr.Exists(ctx, "key1")
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Exists() should return true")
	}

	// 测试 Delete
	err = mgr.Delete(ctx, "key1")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// 验证删除后不存在
	exists, err = mgr.Exists(ctx, "key1")
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if exists {
		t.Error("Exists() should return false after delete")
	}
}

func TestMemoryManager_SetNX(t *testing.T) {
	mgr := NewMemoryManager(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	// 第一次 SetNX 应该成功
	set, err := mgr.SetNX(ctx, "key1", "value1", time.Minute)
	if err != nil {
		t.Errorf("SetNX() error = %v", err)
	}
	if !set {
		t.Error("SetNX() should return true on first call")
	}

	// 第二次 SetNX 应该失败（键已存在）
	set, err = mgr.SetNX(ctx, "key1", "value2", time.Minute)
	if err != nil {
		t.Errorf("SetNX() error = %v", err)
	}
	if set {
		t.Error("SetNX() should return false on second call")
	}

	// 验证值未被覆盖
	value, _ := mgr.GetAny(ctx, "key1")
	if value != "value1" {
		t.Errorf("value = %v, want %v", value, "value1")
	}
}

func TestMemoryManager_BatchOperations(t *testing.T) {
	mgr := NewMemoryManager(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	// 测试 SetMultiple
	items := map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	err := mgr.SetMultiple(ctx, items, time.Minute)
	if err != nil {
		t.Errorf("SetMultiple() error = %v", err)
	}

	// 测试 GetMultiple
	result, err := mgr.GetMultiple(ctx, []string{"key1", "key2", "key3", "key4"})
	if err != nil {
		t.Errorf("GetMultiple() error = %v", err)
	}
	if len(result) != 3 {
		t.Errorf("GetMultiple() returned %d items, want 3", len(result))
	}

	// 测试 DeleteMultiple
	err = mgr.DeleteMultiple(ctx, []string{"key1", "key2"})
	if err != nil {
		t.Errorf("DeleteMultiple() error = %v", err)
	}

	// 验证删除
	result, _ = mgr.GetMultiple(ctx, []string{"key1", "key2", "key3"})
	if len(result) != 1 {
		t.Errorf("After DeleteMultiple(), got %d items, want 1", len(result))
	}
}

func TestMemoryManager_CounterOperations(t *testing.T) {
	mgr := NewMemoryManager(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	// 测试 Increment
	val, err := mgr.Increment(ctx, "counter", 1)
	if err != nil {
		t.Errorf("Increment() error = %v", err)
	}
	if val != 1 {
		t.Errorf("Increment() = %d, want 1", val)
	}

	val, err = mgr.Increment(ctx, "counter", 5)
	if err != nil {
		t.Errorf("Increment() error = %v", err)
	}
	if val != 6 {
		t.Errorf("Increment() = %d, want 6", val)
	}

	// 测试 Decrement
	val, err = mgr.Decrement(ctx, "counter", 2)
	if err != nil {
		t.Errorf("Decrement() error = %v", err)
	}
	if val != 4 {
		t.Errorf("Decrement() = %d, want 4", val)
	}
}

func TestMemoryManager_Clear(t *testing.T) {
	mgr := NewMemoryManager(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	// 添加一些数据
	items := map[string]any{
		"key1": "value1",
		"key2": "value2",
	}
	mgr.SetMultiple(ctx, items, time.Minute)

	// 清空
	err := mgr.Clear(ctx)
	if err != nil {
		t.Errorf("Clear() error = %v", err)
	}

	// 验证清空
	count := mgr.ItemCount()
	if count != 0 {
		t.Errorf("ItemCount() = %d, want 0 after Clear()", count)
	}
}

func TestMemoryManager_Expire(t *testing.T) {
	mgr := NewMemoryManager(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	// 设置一个值
	mgr.Set(ctx, "key1", "value1", time.Minute)

	// 修改过期时间
	err := mgr.Expire(ctx, "key1", 5*time.Minute)
	if err != nil {
		t.Errorf("Expire() error = %v", err)
	}

	// 验证值仍然存在
	value, found := mgr.GetAny(ctx, "key1")
	if !found {
		t.Error("key should still exist after Expire()")
	}
	if value != "value1" {
		t.Errorf("value = %v, want %v", value, "value1")
	}
}

func TestMemoryManager_ContextValidation(t *testing.T) {
	mgr := NewMemoryManager(5*time.Minute, 10*time.Minute)

	// 测试 nil context
	err := mgr.Set(nil, "key1", "value1", time.Minute)
	if err == nil {
		t.Error("Set() with nil context should return error")
	}

	err = mgr.Get(nil, "key1", nil)
	if err == nil {
		t.Error("Get() with nil context should return error")
	}

	err = mgr.Delete(nil, "key1")
	if err == nil {
		t.Error("Delete() with nil context should return error")
	}
}

func TestMemoryManager_KeyValidation(t *testing.T) {
	mgr := NewMemoryManager(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	// 测试空键
	err := mgr.Set(ctx, "", "value1", time.Minute)
	if err == nil {
		t.Error("Set() with empty key should return error")
	}

	err = mgr.Get(ctx, "", nil)
	if err == nil {
		t.Error("Get() with empty key should return error")
	}

	_, err = mgr.Exists(ctx, "")
	if err == nil {
		t.Error("Exists() with empty key should return error")
	}
}

func TestMemoryManager_ConcurrentOperations(t *testing.T) {
	mgr := NewMemoryManager(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	// 测试并发写入
	const numGoroutines = 100
	const numOperations = 100

	done := make(chan bool, numGoroutines)

	// 并发 Set 操作
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key-%d", id*numOperations+j)
				value := fmt.Sprintf("value-%d-%d", id, j)
				err := mgr.Set(ctx, key, value, time.Minute)
				if err != nil {
					t.Errorf("Set() error = %v", err)
				}
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// 验证数据一致性
	expectedCount := numGoroutines * numOperations
	actualCount := mgr.ItemCount()
	if actualCount != expectedCount {
		t.Errorf("ItemCount() = %d, want %d", actualCount, expectedCount)
	}
}

func TestMemoryManager_ConcurrentReads(t *testing.T) {
	mgr := NewMemoryManager(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	// 预设一些数据
	testKeys := []string{}
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("read-key-%d", i)
		value := fmt.Sprintf("read-value-%d", i)
		err := mgr.Set(ctx, key, value, time.Minute)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}
		testKeys = append(testKeys, key)
	}

	const numReaders = 50
	done := make(chan bool, numReaders)

	// 并发读取
	for i := 0; i < numReaders; i++ {
		go func(id int) {
			for _, key := range testKeys {
				value, found := mgr.GetAny(ctx, key)
				if !found {
					t.Errorf("GetAny() should find key %s", key)
				}
				if value == nil {
					t.Errorf("GetAny() returned nil value for key %s", key)
				}
			}
			done <- true
		}(i)
	}

	// 等待所有读取完成
	for i := 0; i < numReaders; i++ {
		<-done
	}
}

func TestMemoryManager_ConcurrentWritesAndReads(t *testing.T) {
	mgr := NewMemoryManager(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	const numWriters = 20
	const numReaders = 20
	const numOperations = 50

	done := make(chan bool, numWriters+numReaders)

	// 并发写入
	for i := 0; i < numWriters; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("mixed-key-%d", id*numOperations+j)
				value := fmt.Sprintf("mixed-value-%d-%d", id, j)
				mgr.Set(ctx, key, value, time.Minute)
			}
			done <- true
		}(i)
	}

	// 并发读取
	for i := 0; i < numReaders; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("mixed-key-%d", j)
				mgr.GetAny(ctx, key)
			}
			done <- true
		}(i)
	}

	// 等待所有操作完成
	for i := 0; i < numWriters+numReaders; i++ {
		<-done
	}

	// 验证没有数据损坏
	count := mgr.ItemCount()
	if count == 0 {
		t.Error("ItemCount() should be > 0 after concurrent operations")
	}
}

func TestMemoryManager_ConcurrentCounter(t *testing.T) {
	mgr := NewMemoryManager(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	const numGoroutines = 100
	const incrementPerGoroutine = 100

	done := make(chan bool, numGoroutines)

	// 并发自增
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < incrementPerGoroutine; j++ {
				mgr.Increment(ctx, "counter", 1)
			}
			done <- true
		}(i)
	}

	// 等待所有自增完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// 验证最终值
	expectedValue := int64(numGoroutines * incrementPerGoroutine)
	value, _ := mgr.GetAny(ctx, "counter")
	if value == nil {
		t.Error("GetAny() should find counter key")
	} else {
		if val, ok := value.(int64); ok {
			if val != expectedValue {
				t.Errorf("Counter = %d, want %d", val, expectedValue)
			}
		} else {
			t.Errorf("Counter is not int64, got %T", value)
		}
	}
}

func TestMemoryManager_ConcurrentSetNX(t *testing.T) {
	mgr := NewMemoryManager(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	const numGoroutines = 100
	done := make(chan bool, numGoroutines)

	// 多个 goroutine 尝试设置同一个键
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			mgr.SetNX(ctx, "lock-key", fmt.Sprintf("value-%d", id), time.Minute)
			done <- true
		}(i)
	}

	// 等待所有操作完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// 验证只有一个成功
	exists, _ := mgr.Exists(ctx, "lock-key")
	if !exists {
		t.Error("lock-key should exist")
	}
}
