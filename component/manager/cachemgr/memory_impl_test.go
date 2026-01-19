package cachemgr

import (
	"context"
	"sync"
	"testing"
	"time"
)

// setupMemoryManager 创建一个内存缓存管理器用于测试
func setupMemoryManager(t *testing.T) ICacheManager {
	return NewCacheManagerMemoryImpl(10*time.Minute, 5*time.Minute)
}

// TestMemoryManager_ManagerName 测试管理器名称
func TestMemoryManager_ManagerName(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	if mgr.ManagerName() != "cacheManagerMemoryImpl" {
		t.Errorf("expected name 'cacheManagerMemoryImpl', got '%s'", mgr.ManagerName())
	}
}

// TestMemoryManager_Health 测试健康检查
func TestMemoryManager_Health(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	if err := mgr.Health(); err != nil {
		t.Errorf("Health() should not return error, got %v", err)
	}
}

// TestMemoryManager_Lifecycle 测试生命周期
func TestMemoryManager_Lifecycle(t *testing.T) {
	mgr := setupMemoryManager(t)

	// 测试 OnStart
	if err := mgr.OnStart(); err != nil {
		t.Errorf("OnStart() error = %v", err)
	}

	// 测试 OnStop
	if err := mgr.OnStop(); err != nil {
		t.Errorf("OnStop() error = %v", err)
	}

	// 测试 Close
	if err := mgr.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

// TestMemoryManager_SetAndGet 测试基本的设置和获取
func TestMemoryManager_SetAndGet(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	// 测试设置和获取字符串
	err := mgr.Set(ctx, "key1", "value1", 5*time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	var result any
	err = mgr.Get(ctx, "key1", &result)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if result != "value1" {
		t.Errorf("expected 'value1', got '%v'", result)
	}

	// 测试获取不存在的键
	err = mgr.Get(ctx, "nonexistent", &result)
	if err == nil {
		t.Error("expected error when getting non-existent key, got nil")
	}
}

// TestMemoryManager_GetWithNilContext 测试空上下文
func TestMemoryManager_GetWithNilContext(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	var result any
	err := mgr.Get(nil, "key", &result)
	if err == nil {
		t.Error("expected error with nil context, got nil")
	}
}

// TestMemoryManager_SetWithNilContext 测试空上下文设置
func TestMemoryManager_SetWithNilContext(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	err := mgr.Set(nil, "key", "value", 5*time.Minute)
	if err == nil {
		t.Error("expected error with nil context, got nil")
	}
}

// TestMemoryManager_SetWithEmptyKey 测试空键
func TestMemoryManager_SetWithEmptyKey(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Set(ctx, "", "value", 5*time.Minute)
	if err == nil {
		t.Error("expected error with empty key, got nil")
	}
}

// TestMemoryManager_SetNX 测试 SetNX
func TestMemoryManager_SetNX(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	// 首次设置，应该成功
	set, err := mgr.SetNX(ctx, "key_nx", "value1", 5*time.Minute)
	if err != nil {
		t.Fatalf("SetNX() error = %v", err)
	}
	if !set {
		t.Error("expected SetNX to return true on first call, got false")
	}

	// 再次设置，应该失败
	set, err = mgr.SetNX(ctx, "key_nx", "value2", 5*time.Minute)
	if err != nil {
		t.Fatalf("SetNX() error = %v", err)
	}
	if set {
		t.Error("expected SetNX to return false on second call, got true")
	}

	// 验证值没有被覆盖
	var result any
	err = mgr.Get(ctx, "key_nx", &result)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if result != "value1" {
		t.Errorf("expected 'value1', got '%v'", result)
	}
}

// TestMemoryManager_Delete 测试删除
func TestMemoryManager_Delete(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	// 设置一个值
	err := mgr.Set(ctx, "key_delete", "value", 5*time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// 删除它
	err = mgr.Delete(ctx, "key_delete")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// 验证已删除
	var result any
	err = mgr.Get(ctx, "key_delete", &result)
	if err == nil {
		t.Error("expected error when getting deleted key, got nil")
	}
}

// TestMemoryManager_Exists 测试检查键是否存在
func TestMemoryManager_Exists(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	// 检查不存在的键
	exists, err := mgr.Exists(ctx, "key_exists")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if exists {
		t.Error("expected key to not exist, got true")
	}

	// 设置键
	err = mgr.Set(ctx, "key_exists", "value", 5*time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// 再次检查
	exists, err = mgr.Exists(ctx, "key_exists")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("expected key to exist, got false")
	}
}

// TestMemoryManager_Expire 测试设置过期时间
func TestMemoryManager_Expire(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	// 设置一个值
	err := mgr.Set(ctx, "key_expire", "value", 5*time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// 修改过期时间
	err = mgr.Expire(ctx, "key_expire", 10*time.Minute)
	if err != nil {
		t.Fatalf("Expire() error = %v", err)
	}

	// 验证键仍然存在
	exists, err := mgr.Exists(ctx, "key_expire")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("expected key to exist after updating expiration")
	}

	// 尝试为不存在的键设置过期时间
	err = mgr.Expire(ctx, "nonexistent", 5*time.Minute)
	if err == nil {
		t.Error("expected error when expiring non-existent key, got nil")
	}
}

// TestMemoryManager_TTL 测试获取 TTL
func TestMemoryManager_TTL(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	// 设置一个值
	err := mgr.Set(ctx, "key_ttl", "value", 5*time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// 获取 TTL（内存实现返回 0）
	ttl, err := mgr.TTL(ctx, "key_ttl")
	if err != nil {
		t.Fatalf("TTL() error = %v", err)
	}
	// go-cache 不直接支持 TTL 查询，返回 0
	if ttl != 0 {
		t.Logf("Note: TTL returned %v (expected 0 for memory cache)", ttl)
	}
}

// TestMemoryManager_Clear 测试清空缓存
func TestMemoryManager_Clear(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	// 设置多个键
	for i := 0; i < 5; i++ {
		key := "key_clear_" + string(rune('0'+i))
		err := mgr.Set(ctx, key, i, 5*time.Minute)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}
	}

	// 清空所有缓存
	err := mgr.Clear(ctx)
	if err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	// 验证所有键都已删除
	for i := 0; i < 5; i++ {
		key := "key_clear_" + string(rune('0'+i))
		exists, err := mgr.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists() error = %v", err)
		}
		if exists {
			t.Errorf("expected key '%s' to be deleted", key)
		}
	}
}

// TestMemoryManager_GetMultiple 测试批量获取
func TestMemoryManager_GetMultiple(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	// 设置多个键
	items := map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	for key, value := range items {
		err := mgr.Set(ctx, key, value, 5*time.Minute)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}
	}

	// 批量获取
	keys := []string{"key1", "key2", "key3", "key4"}
	result, err := mgr.GetMultiple(ctx, keys)
	if err != nil {
		t.Fatalf("GetMultiple() error = %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 items, got %d", len(result))
	}

	if result["key1"] != "value1" {
		t.Errorf("expected key1 = 'value1', got '%v'", result["key1"])
	}
	if result["key2"] != "value2" {
		t.Errorf("expected key2 = 'value2', got '%v'", result["key2"])
	}
	if result["key4"] != nil {
		t.Error("expected key4 to not exist in result")
	}

	// 测试空键列表
	result, err = mgr.GetMultiple(ctx, []string{})
	if err != nil {
		t.Fatalf("GetMultiple() with empty keys error = %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d items", len(result))
	}
}

// TestMemoryManager_SetMultiple 测试批量设置
func TestMemoryManager_SetMultiple(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	// 批量设置
	items := map[string]any{
		"batch1": "value1",
		"batch2": "value2",
		"batch3": "value3",
	}
	err := mgr.SetMultiple(ctx, items, 5*time.Minute)
	if err != nil {
		t.Fatalf("SetMultiple() error = %v", err)
	}

	// 验证所有键都已设置
	for key, expectedValue := range items {
		var result any
		err := mgr.Get(ctx, key, &result)
		if err != nil {
			t.Errorf("Get() for key '%s' error = %v", key, err)
		}
		if result != expectedValue {
			t.Errorf("expected '%s' = '%v', got '%v'", key, expectedValue, result)
		}
	}

	// 测试空 map
	err = mgr.SetMultiple(ctx, map[string]any{}, 5*time.Minute)
	if err != nil {
		t.Fatalf("SetMultiple() with empty items error = %v", err)
	}
}

// TestMemoryManager_DeleteMultiple 测试批量删除
func TestMemoryManager_DeleteMultiple(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	// 设置多个键
	keys := []string{"del1", "del2", "del3", "keep1"}
	for _, key := range keys {
		err := mgr.Set(ctx, key, "value", 5*time.Minute)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}
	}

	// 批量删除部分键
	err := mgr.DeleteMultiple(ctx, []string{"del1", "del2", "del3"})
	if err != nil {
		t.Fatalf("DeleteMultiple() error = %v", err)
	}

	// 验证已删除
	for _, key := range []string{"del1", "del2", "del3"} {
		exists, err := mgr.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists() error = %v", err)
		}
		if exists {
			t.Errorf("expected key '%s' to be deleted", key)
		}
	}

	// 验证未删除的键仍然存在
	exists, err := mgr.Exists(ctx, "keep1")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("expected key 'keep1' to still exist")
	}

	// 测试空键列表
	err = mgr.DeleteMultiple(ctx, []string{})
	if err != nil {
		t.Fatalf("DeleteMultiple() with empty keys error = %v", err)
	}
}

// TestMemoryManager_Increment 测试自增
func TestMemoryManager_Increment(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	// 自增一个不存在的键（从 0 开始）
	result, err := mgr.Increment(ctx, "counter", 10)
	if err != nil {
		t.Fatalf("Increment() error = %v", err)
	}
	if result != 10 {
		t.Errorf("expected 10, got %d", result)
	}

	// 再次自增
	result, err = mgr.Increment(ctx, "counter", 5)
	if err != nil {
		t.Fatalf("Increment() error = %v", err)
	}
	if result != 15 {
		t.Errorf("expected 15, got %d", result)
	}

	// 验证值
	var value any
	err = mgr.Get(ctx, "counter", &value)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if value != int64(15) {
		t.Errorf("expected value to be 15, got %v", value)
	}
}

// TestMemoryManager_IncrementNonInt64 测试自增非 int64 值
func TestMemoryManager_IncrementNonInt64(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	// 设置一个非 int64 的值
	err := mgr.Set(ctx, "not_int", "string_value", 5*time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// 尝试自增
	_, err = mgr.Increment(ctx, "not_int", 5)
	if err == nil {
		t.Error("expected error when incrementing non-int64 value, got nil")
	}
}

// TestMemoryManager_Decrement 测试自减
func TestMemoryManager_Decrement(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	// 先设置一个值
	_, err := mgr.Increment(ctx, "decr_counter", 100)
	if err != nil {
		t.Fatalf("Increment() error = %v", err)
	}

	// 自减
	result, err := mgr.Decrement(ctx, "decr_counter", 30)
	if err != nil {
		t.Fatalf("Decrement() error = %v", err)
	}
	if result != 70 {
		t.Errorf("expected 70, got %d", result)
	}

	// 再次自减（可能变成负数）
	result, err = mgr.Decrement(ctx, "decr_counter", 100)
	if err != nil {
		t.Fatalf("Decrement() error = %v", err)
	}
	if result != -30 {
		t.Errorf("expected -30, got %d", result)
	}
}

// TestMemoryManager_Expiration 测试过期
func TestMemoryManager_Expiration(t *testing.T) {
	mgr := NewCacheManagerMemoryImpl(100*time.Millisecond, 50*time.Millisecond)
	defer mgr.Close()

	ctx := context.Background()

	// 设置一个值，过期时间短于清理间隔
	err := mgr.Set(ctx, "expire_test", "value", 50*time.Millisecond)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// 立即获取应该成功
	var result any
	err = mgr.Get(ctx, "expire_test", &result)
	if err != nil {
		t.Errorf("Get() immediately after Set() failed: %v", err)
	}

	// 等待过期
	time.Sleep(200 * time.Millisecond)

	// 再次获取应该失败
	err = mgr.Get(ctx, "expire_test", &result)
	if err == nil {
		t.Error("expected error when getting expired key, got nil")
	}
}

// TestMemoryManager_ConcurrentOperations 测试并发操作
func TestMemoryManager_ConcurrentOperations(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()
	const numGoroutines = 100
	const numOperations = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// 并发写入
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := "concurrent_key_" + string(rune('0'+id%10))
				err := mgr.Set(ctx, key, id, 5*time.Minute)
				if err != nil {
					t.Errorf("Set() error = %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	// 验证数据一致性
	for i := 0; i < 10; i++ {
		key := "concurrent_key_" + string(rune('0'+i))
		exists, err := mgr.Exists(ctx, key)
		if err != nil {
			t.Errorf("Exists() error = %v", err)
		}
		if !exists {
			t.Errorf("expected key '%s' to exist", key)
		}
	}
}

// TestMemoryManager_ComplexValues 测试复杂类型
func TestMemoryManager_ComplexValues(t *testing.T) {
	mgr := setupMemoryManager(t)
	defer mgr.Close()

	ctx := context.Background()

	// 测试 map
	mapValue := map[string]int{"a": 1, "b": 2}
	err := mgr.Set(ctx, "map_key", mapValue, 5*time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	var mapResult any
	err = mgr.Get(ctx, "map_key", &mapResult)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// 测试 slice
	sliceValue := []string{"x", "y", "z"}
	err = mgr.Set(ctx, "slice_key", sliceValue, 5*time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	var sliceResult any
	err = mgr.Get(ctx, "slice_key", &sliceResult)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// 测试 struct
	type Person struct {
		Name string
		Age  int
	}
	structValue := Person{Name: "Alice", Age: 30}
	err = mgr.Set(ctx, "struct_key", structValue, 5*time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	var structResult any
	err = mgr.Get(ctx, "struct_key", &structResult)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
}

// TestMemoryManager_ItemCount 测试获取缓存项数量
func TestMemoryManager_ItemCount(t *testing.T) {
	// 使用类型断言获取内存实现
	mgr := NewCacheManagerMemoryImpl(10*time.Minute, 5*time.Minute)
	defer mgr.Close()

	// 尝试类型断言
	if memImpl, ok := mgr.(*cacheManagerMemoryImpl); ok {
		ctx := context.Background()

		// 初始计数应该为 0
		if count := memImpl.ItemCount(); count != 0 {
			t.Errorf("expected initial count to be 0, got %d", count)
		}

		// 添加一些项
		for i := 0; i < 5; i++ {
			key := "count_key_" + string(rune('0'+i))
			err := mgr.Set(ctx, key, i, 5*time.Minute)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}
		}

		// 计数应该是 5
		if count := memImpl.ItemCount(); count != 5 {
			t.Errorf("expected count to be 5, got %d", count)
		}

		// 删除一个项
		err := mgr.Delete(ctx, "count_key_0")
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		// 计数应该是 4
		if count := memImpl.ItemCount(); count != 4 {
			t.Errorf("expected count to be 4 after deletion, got %d", count)
		}
	}
}
