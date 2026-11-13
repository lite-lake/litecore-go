package cachemgr

import (
	"context"
	"testing"
	"time"
)

// setupNoneManager 创建一个 None 缓存管理器用于测试
func setupNoneManager(t *testing.T) ICacheManager {
	return NewCacheManagerNoneImpl(nil, nil)
}

// TestNoneManager_ManagerName 测试管理器名称
func TestNoneManager_ManagerName(t *testing.T) {
	mgr := setupNoneManager(t)
	defer mgr.Close()

	if mgr.ManagerName() != "cacheManagerNoneImpl" {
		t.Errorf("expected name 'cacheManagerNoneImpl', got '%s'", mgr.ManagerName())
	}
}

// TestNoneManager_Health 测试健康检查
func TestNoneManager_Health(t *testing.T) {
	mgr := setupNoneManager(t)
	defer mgr.Close()

	if err := mgr.Health(); err != nil {
		t.Errorf("Health() should not return error, got %v", err)
	}
}

// TestNoneManager_Lifecycle 测试生命周期
func TestNoneManager_Lifecycle(t *testing.T) {
	mgr := setupNoneManager(t)

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

// TestNoneManager_Get 测试获取（应该返回错误）
func TestNoneManager_Get(t *testing.T) {
	mgr := setupNoneManager(t)
	defer mgr.Close()

	ctx := context.Background()
	var result any

	err := mgr.Get(ctx, "any_key", &result)
	if err == nil {
		t.Error("expected error from Ins(), got nil")
	}
}

// TestNoneManager_Set 测试设置（应该是空操作）
func TestNoneManager_Set(t *testing.T) {
	mgr := setupNoneManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Set(ctx, "any_key", "any_value", 5*time.Minute)
	if err != nil {
		t.Errorf("Set() should not return error, got %v", err)
	}

	// 验证值没有被存储
	var result any
	err = mgr.Get(ctx, "any_key", &result)
	if err == nil {
		t.Error("expected error when getting value that was supposedly set")
	}
}

// TestNoneManager_SetNX 测试 SetNX（应该返回 false）
func TestNoneManager_SetNX(t *testing.T) {
	mgr := setupNoneManager(t)
	defer mgr.Close()

	ctx := context.Background()

	set, err := mgr.SetNX(ctx, "any_key", "any_value", 5*time.Minute)
	if err != nil {
		t.Errorf("SetNX() should not return error, got %v", err)
	}
	if set {
		t.Error("expected SetNX to return false, got true")
	}
}

// TestNoneManager_Delete 测试删除（应该是空操作）
func TestNoneManager_Delete(t *testing.T) {
	mgr := setupNoneManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Delete(ctx, "any_key")
	if err != nil {
		t.Errorf("Delete() should not return error, got %v", err)
	}
}

// TestNoneManager_Exists 测试检查键是否存在（应该返回 false）
func TestNoneManager_Exists(t *testing.T) {
	mgr := setupNoneManager(t)
	defer mgr.Close()

	ctx := context.Background()

	// 即使之前"设置"过，也应该返回 false
	exists, err := mgr.Exists(ctx, "any_key")
	if err != nil {
		t.Errorf("Exists() should not return error, got %v", err)
	}
	if exists {
		t.Error("expected Exists to return false, got true")
	}
}

// TestNoneManager_Expire 测试设置过期时间（应该是空操作）
func TestNoneManager_Expire(t *testing.T) {
	mgr := setupNoneManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Expire(ctx, "any_key", 5*time.Minute)
	if err != nil {
		t.Errorf("Expire() should not return error, got %v", err)
	}
}

// TestNoneManager_TTL 测试获取 TTL（应该返回 0）
func TestNoneManager_TTL(t *testing.T) {
	mgr := setupNoneManager(t)
	defer mgr.Close()

	ctx := context.Background()

	ttl, err := mgr.TTL(ctx, "any_key")
	if err != nil {
		t.Errorf("TTL() should not return error, got %v", err)
	}
	if ttl != 0 {
		t.Errorf("expected TTL to be 0, got %v", ttl)
	}
}

// TestNoneManager_Clear 测试清空缓存（应该是空操作）
func TestNoneManager_Clear(t *testing.T) {
	mgr := setupNoneManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.Clear(ctx)
	if err != nil {
		t.Errorf("Clear() should not return error, got %v", err)
	}
}

// TestNoneManager_GetMultiple 测试批量获取（应该返回空 map）
func TestNoneManager_GetMultiple(t *testing.T) {
	mgr := setupNoneManager(t)
	defer mgr.Close()

	ctx := context.Background()

	result, err := mgr.GetMultiple(ctx, []string{"key1", "key2", "key3"})
	if err != nil {
		t.Errorf("GetMultiple() should not return error, got %v", err)
	}
	if result == nil {
		t.Error("expected result to be empty map, not nil")
	} else if len(result) != 0 {
		t.Errorf("expected empty result, got %d items", len(result))
	}

	// 测试空键列表
	result, err = mgr.GetMultiple(ctx, []string{})
	if err != nil {
		t.Errorf("GetMultiple() with empty keys should not return error, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d items", len(result))
	}
}

// TestNoneManager_SetMultiple 测试批量设置（应该是空操作）
func TestNoneManager_SetMultiple(t *testing.T) {
	mgr := setupNoneManager(t)
	defer mgr.Close()

	ctx := context.Background()

	items := map[string]any{
		"key1": "value1",
		"key2": "value2",
	}
	err := mgr.SetMultiple(ctx, items, 5*time.Minute)
	if err != nil {
		t.Errorf("SetMultiple() should not return error, got %v", err)
	}

	// 验证值没有被存储
	result, err := mgr.GetMultiple(ctx, []string{"key1", "key2"})
	if err != nil {
		t.Errorf("GetMultiple() should not return error, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d items", len(result))
	}

	// 测试空 map
	err = mgr.SetMultiple(ctx, map[string]any{}, 5*time.Minute)
	if err != nil {
		t.Errorf("SetMultiple() with empty items should not return error, got %v", err)
	}
}

// TestNoneManager_DeleteMultiple 测试批量删除（应该是空操作）
func TestNoneManager_DeleteMultiple(t *testing.T) {
	mgr := setupNoneManager(t)
	defer mgr.Close()

	ctx := context.Background()

	err := mgr.DeleteMultiple(ctx, []string{"key1", "key2", "key3"})
	if err != nil {
		t.Errorf("DeleteMultiple() should not return error, got %v", err)
	}

	// 测试空键列表
	err = mgr.DeleteMultiple(ctx, []string{})
	if err != nil {
		t.Errorf("DeleteMultiple() with empty keys should not return error, got %v", err)
	}
}

// TestNoneManager_Increment 测试自增（应该返回 0）
func TestNoneManager_Increment(t *testing.T) {
	mgr := setupNoneManager(t)
	defer mgr.Close()

	ctx := context.Background()

	result, err := mgr.Increment(ctx, "counter", 10)
	if err != nil {
		t.Errorf("Increment() should not return error, got %v", err)
	}
	if result != 0 {
		t.Errorf("expected Increment to return 0, got %d", result)
	}
}

// TestNoneManager_Decrement 测试自减（应该返回 0）
func TestNoneManager_Decrement(t *testing.T) {
	mgr := setupNoneManager(t)
	defer mgr.Close()

	ctx := context.Background()

	result, err := mgr.Decrement(ctx, "counter", 10)
	if err != nil {
		t.Errorf("Decrement() should not return error, got %v", err)
	}
	if result != 0 {
		t.Errorf("expected Decrement to return 0, got %d", result)
	}
}

// TestNoneManager_NilContext 测试空上下文
func TestNoneManager_NilContext(t *testing.T) {
	mgr := setupNoneManager(t)
	defer mgr.Close()

	// 大部分操作应该能处理 nil context（空操作）
	tests := []struct {
		name string
		fn   func() error
	}{
		{"Set", func() error { return mgr.Set(nil, "key", "value", 5*time.Minute) }},
		{"Delete", func() error { return mgr.Delete(nil, "key") }},
		{"Expire", func() error { return mgr.Expire(nil, "key", 5*time.Minute) }},
		{"Clear", func() error { return mgr.Clear(nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fn(); err != nil {
				// 某些操作可能会验证 context
				t.Logf("%s with nil context returned error: %v", tt.name, err)
			}
		})
	}
}

// TestNoneManager_EmptyKey 测试空键
func TestNoneManager_EmptyKey(t *testing.T) {
	mgr := setupNoneManager(t)
	defer mgr.Close()

	ctx := context.Background()

	// 大部分操作应该能处理空键（空操作）
	tests := []struct {
		name string
		fn   func() error
	}{
		{"Set", func() error { return mgr.Set(ctx, "", "value", 5*time.Minute) }},
		{"Delete", func() error { return mgr.Delete(ctx, "") }},
		{"Expire", func() error { return mgr.Expire(ctx, "", 5*time.Minute) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fn(); err != nil {
				// 某些操作可能会验证 key
				t.Logf("%s with empty key returned error: %v", tt.name, err)
			}
		})
	}
}

// TestNoneManager_ComplexValues 测试复杂类型（空操作，但不应该 panic）
func TestNoneManager_ComplexValues(t *testing.T) {
	mgr := setupNoneManager(t)
	defer mgr.Close()

	ctx := context.Background()

	// 测试各种复杂类型的设置（都应该成功，因为是空操作）
	tests := []struct {
		name  string
		value any
	}{
		{"map", map[string]int{"a": 1, "b": 2}},
		{"slice", []string{"x", "y", "z"}},
		{"struct", struct{ Name string }{Name: "Test"}},
		{"nil", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.Set(ctx, "key", tt.value, 5*time.Minute)
			if err != nil {
				t.Errorf("Set() with %s should not return error, got %v", tt.name, err)
			}
		})
	}
}

// TestNoneManager_WithNilContextAndEmptyKey 测试空上下文和空键的组合
func TestNoneManager_WithNilContextAndEmptyKey(t *testing.T) {
	mgr := setupNoneManager(t)
	defer mgr.Close()

	// 测试最极端的情况
	tests := []struct {
		name string
		fn   func() error
	}{
		{"Set", func() error { return mgr.Set(nil, "", "value", 5*time.Minute) }},
		{"Delete", func() error { return mgr.Delete(nil, "") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fn(); err != nil {
				t.Logf("%s with nil context and empty key returned error: %v", tt.name, err)
			}
		})
	}
}

// TestNoneManager_AllOperationsNoPanic 测试所有操作都不会 panic
func TestNoneManager_AllOperationsNoPanic(t *testing.T) {
	mgr := setupNoneManager(t)
	defer mgr.Close()

	ctx := context.Background()
	var result any

	// 确保所有操作都不会 panic
	operations := []func(){
		func() { mgr.Set(ctx, "key", "value", 5*time.Minute) },
		func() { mgr.Get(ctx, "key", &result) },
		func() { mgr.SetNX(ctx, "key", "value", 5*time.Minute) },
		func() { mgr.Delete(ctx, "key") },
		func() { mgr.Exists(ctx, "key") },
		func() { mgr.Expire(ctx, "key", 5*time.Minute) },
		func() { mgr.TTL(ctx, "key") },
		func() { mgr.Clear(ctx) },
		func() { mgr.GetMultiple(ctx, []string{"key1", "key2"}) },
		func() { mgr.SetMultiple(ctx, map[string]any{"key": "value"}, 5*time.Minute) },
		func() { mgr.DeleteMultiple(ctx, []string{"key1", "key2"}) },
		func() { mgr.Increment(ctx, "key", 1) },
		func() { mgr.Decrement(ctx, "key", 1) },
		func() { mgr.Health() },
		func() { mgr.OnStart() },
		func() { mgr.OnStop() },
		func() { mgr.ManagerName() },
		func() { mgr.Close() },
	}

	for _, op := range operations {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("operation panicked: %v", r)
				}
			}()
			op()
		}()
	}
}

// TestNoneManager_IdempotentOperations 测试操作的幂等性
func TestNoneManager_IdempotentOperations(t *testing.T) {
	mgr := setupNoneManager(t)
	defer mgr.Close()

	ctx := context.Background()

	// 多次执行相同操作，结果应该一致
	for i := 0; i < 10; i++ {
		// Set 应该总是成功
		err := mgr.Set(ctx, "same_key", "value", 5*time.Minute)
		if err != nil {
			t.Errorf("iteration %d: Set() returned error %v", i, err)
		}

		// Exists 应该总是返回 false
		exists, err := mgr.Exists(ctx, "same_key")
		if err != nil {
			t.Errorf("iteration %d: Exists() returned error %v", i, err)
		}
		if exists {
			t.Errorf("iteration %d: Exists() returned true", i)
		}

		// TTL 应该总是返回 0
		ttl, err := mgr.TTL(ctx, "same_key")
		if err != nil {
			t.Errorf("iteration %d: TTL() returned error %v", i, err)
		}
		if ttl != 0 {
			t.Errorf("iteration %d: TTL() returned %v", i, ttl)
		}
	}
}
