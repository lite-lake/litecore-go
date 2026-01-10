package drivers

import (
	"context"
	"testing"
	"time"
)

func TestNoneManager(t *testing.T) {
	mgr := NewNoneManager()

	ctx := context.Background()

	// 测试 Manager 接口
	if mgr.ManagerName() != "none-cache" {
		t.Errorf("ManagerName() = %v, want %v", mgr.ManagerName(), "none-cache")
	}

	if err := mgr.Health(); err != nil {
		t.Errorf("Health() error = %v", err)
	}

	if err := mgr.OnStart(); err != nil {
		t.Errorf("OnStart() error = %v", err)
	}

	if err := mgr.OnStop(); err != nil {
		t.Errorf("OnStop() error = %v", err)
	}

	if err := mgr.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// 测试 Get - 应该返回错误
	var dest string
	err := mgr.Get(ctx, "test-key", &dest)
	if err == nil {
		t.Error("Get() should return error for none cache")
	}

	// 测试 Set - 不应该返回错误
	err = mgr.Set(ctx, "test-key", "test-value", time.Minute)
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	// 测试 SetNX - 应该返回 false
	set, err := mgr.SetNX(ctx, "test-key", "test-value", time.Minute)
	if err != nil {
		t.Errorf("SetNX() error = %v", err)
	}
	if set {
		t.Error("SetNX() should return false for none cache")
	}

	// 测试 Delete
	err = mgr.Delete(ctx, "test-key")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// 测试 Exists
	exists, err := mgr.Exists(ctx, "test-key")
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if exists {
		t.Error("Exists() should return false for none cache")
	}

	// 测试 Expire
	err = mgr.Expire(ctx, "test-key", time.Minute)
	if err != nil {
		t.Errorf("Expire() error = %v", err)
	}

	// 测试 TTL
	ttl, err := mgr.TTL(ctx, "test-key")
	if err != nil {
		t.Errorf("TTL() error = %v", err)
	}
	if ttl != 0 {
		t.Errorf("TTL() = %v, want 0", ttl)
	}

	// 测试 Clear
	err = mgr.Clear(ctx)
	if err != nil {
		t.Errorf("Clear() error = %v", err)
	}

	// 测试 GetMultiple
	result, err := mgr.GetMultiple(ctx, []string{"key1", "key2"})
	if err != nil {
		t.Errorf("GetMultiple() error = %v", err)
	}
	if result == nil || len(result) != 0 {
		t.Errorf("GetMultiple() should return empty map, got %v", result)
	}

	// 测试 SetMultiple
	items := map[string]any{
		"key1": "value1",
		"key2": "value2",
	}
	err = mgr.SetMultiple(ctx, items, time.Minute)
	if err != nil {
		t.Errorf("SetMultiple() error = %v", err)
	}

	// 测试 DeleteMultiple
	err = mgr.DeleteMultiple(ctx, []string{"key1", "key2"})
	if err != nil {
		t.Errorf("DeleteMultiple() error = %v", err)
	}

	// 测试 Increment
	val, err := mgr.Increment(ctx, "counter", 1)
	if err != nil {
		t.Errorf("Increment() error = %v", err)
	}
	if val != 0 {
		t.Errorf("Increment() = %v, want 0", val)
	}

	// 测试 Decrement
	val, err = mgr.Decrement(ctx, "counter", 1)
	if err != nil {
		t.Errorf("Decrement() error = %v", err)
	}
	if val != 0 {
		t.Errorf("Decrement() = %v, want 0", val)
	}
}
