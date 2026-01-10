package cachemgr

import (
	"context"
	"testing"
	"time"

	"com.litelake.litecore/manager/cachemgr/internal/drivers"
)

func TestCacheManagerAdapter_WithMemoryDriver(t *testing.T) {
	memoryMgr := drivers.NewMemoryManager(5*time.Minute, 10*time.Minute)
	adapter := NewCacheManagerAdapter(memoryMgr)

	ctx := context.Background()

	// 测试基本操作
	err := adapter.Set(ctx, "key1", "value1", time.Minute)
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	// 使用 GetAny 获取
	value, found := memoryMgr.GetAny(ctx, "key1")
	if !found {
		t.Error("key should exist")
	}
	if value != "value1" {
		t.Errorf("value = %v, want %v", value, "value1")
	}

	// 测试 Exists
	exists, err := adapter.Exists(ctx, "key1")
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("key should exist")
	}

	// 测试 Delete
	err = adapter.Delete(ctx, "key1")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}
}

func TestCacheManagerAdapter_WithNoneDriver(t *testing.T) {
	noneMgr := drivers.NewNoneManager()
	adapter := NewCacheManagerAdapter(noneMgr)

	ctx := context.Background()

	// 测试 Manager 接口方法
	if adapter.ManagerName() != "none-cache" {
		t.Errorf("ManagerName() = %v, want %v", adapter.ManagerName(), "none-cache")
	}

	if err := adapter.Health(); err != nil {
		t.Errorf("Health() error = %v", err)
	}

	// 测试 Get - 应该返回错误
	var dest string
	err := adapter.Get(ctx, "test-key", &dest)
	if err == nil {
		t.Error("Get() should return error for none cache")
	}

	// 测试 Set - 不应该返回错误
	err = adapter.Set(ctx, "test-key", "test-value", time.Minute)
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	// 测试 SetNX - 应该返回 false
	set, err := adapter.SetNX(ctx, "test-key", "test-value", time.Minute)
	if err != nil {
		t.Errorf("SetNX() error = %v", err)
	}
	if set {
		t.Error("SetNX() should return false for none cache")
	}
}

func TestCacheManagerAdapter_BatchOperations(t *testing.T) {
	memoryMgr := drivers.NewMemoryManager(5*time.Minute, 10*time.Minute)
	adapter := NewCacheManagerAdapter(memoryMgr)

	ctx := context.Background()

	// 测试 SetMultiple
	items := map[string]any{
		"key1": "value1",
		"key2": "value2",
	}
	err := adapter.SetMultiple(ctx, items, time.Minute)
	if err != nil {
		t.Errorf("SetMultiple() error = %v", err)
	}

	// 测试 GetMultiple
	result, err := adapter.GetMultiple(ctx, []string{"key1", "key2"})
	if err != nil {
		t.Errorf("GetMultiple() error = %v", err)
	}
	if len(result) != 2 {
		t.Errorf("GetMultiple() returned %d items, want 2", len(result))
	}

	// 测试 DeleteMultiple
	err = adapter.DeleteMultiple(ctx, []string{"key1", "key2"})
	if err != nil {
		t.Errorf("DeleteMultiple() error = %v", err)
	}
}

func TestCacheManagerAdapter_CounterOperations(t *testing.T) {
	memoryMgr := drivers.NewMemoryManager(5*time.Minute, 10*time.Minute)
	adapter := NewCacheManagerAdapter(memoryMgr)

	ctx := context.Background()

	// 测试 Increment
	val, err := adapter.Increment(ctx, "counter", 1)
	if err != nil {
		t.Errorf("Increment() error = %v", err)
	}
	if val != 1 {
		t.Errorf("Increment() = %d, want 1", val)
	}

	// 测试 Decrement
	val, err = adapter.Decrement(ctx, "counter", 1)
	if err != nil {
		t.Errorf("Decrement() error = %v", err)
	}
	if val != 0 {
		t.Errorf("Decrement() = %d, want 0", val)
	}
}

func TestCacheManagerAdapter_TTL(t *testing.T) {
	memoryMgr := drivers.NewMemoryManager(5*time.Minute, 10*time.Minute)
	adapter := NewCacheManagerAdapter(memoryMgr)

	ctx := context.Background()

	// 设置一个值
	memoryMgr.Set(ctx, "key1", "value1", time.Minute)

	// 获取 TTL
	ttl, err := adapter.TTL(ctx, "key1")
	if err != nil {
		t.Errorf("TTL() error = %v", err)
	}
	// Memory 驱动的 TTL 实现返回 0（因为 go-cache 限制）
	if ttl != 0 {
		t.Logf("TTL() = %v (Memory driver returns 0)", ttl)
	}

	// 设置过期时间
	err = adapter.Expire(ctx, "key1", 5*time.Minute)
	if err != nil {
		t.Errorf("Expire() error = %v", err)
	}
}

func TestCacheManagerAdapter_Clear(t *testing.T) {
	memoryMgr := drivers.NewMemoryManager(5*time.Minute, 10*time.Minute)
	adapter := NewCacheManagerAdapter(memoryMgr)

	ctx := context.Background()

	// 添加一些数据
	items := map[string]any{
		"key1": "value1",
		"key2": "value2",
	}
	adapter.SetMultiple(ctx, items, time.Minute)

	// 清空
	err := adapter.Clear(ctx)
	if err != nil {
		t.Errorf("Clear() error = %v", err)
	}

	// 验证清空
	count := memoryMgr.ItemCount()
	if count != 0 {
		t.Errorf("ItemCount() = %d, want 0 after Clear()", count)
	}
}

func TestCacheManagerAdapter_Close(t *testing.T) {
	memoryMgr := drivers.NewMemoryManager(5*time.Minute, 10*time.Minute)
	adapter := NewCacheManagerAdapter(memoryMgr)

	err := adapter.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
