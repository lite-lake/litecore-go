package cachemgr

import (
	"context"
	"testing"
	"time"

	"com.litelake.litecore/common"
)

// TestManagerImplementsInterfaces 验证 Manager 实现了所有必需的接口
func TestManagerImplementsInterfaces(t *testing.T) {
	var _ CacheManager = (*Manager)(nil)
	var _ common.BaseManager = (*Manager)(nil)
}

// TestNewManager 测试 Manager 构造函数
func TestNewManager(t *testing.T) {
	mgr := NewManager("test")
	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}

	if mgr.ManagerName() != "test" {
		t.Errorf("expected manager name 'test', got '%s'", mgr.ManagerName())
	}
}

// TestManagerOnStartWithoutConfig 测试无配置时的启动
func TestManagerOnStartWithoutConfig(t *testing.T) {
	mgr := NewManager("test")

	// 无配置启动，应使用默认配置（内存缓存）
	err := mgr.OnStart()
	if err != nil {
		t.Fatalf("OnStart failed: %v", err)
	}

	// 验证健康检查
	if err := mgr.Health(); err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	// 停止管理器
	if err := mgr.OnStop(); err != nil {
		t.Errorf("OnStop failed: %v", err)
	}
}

// TestManagerBasicOperations 测试基本缓存操作
func TestManagerBasicOperations(t *testing.T) {
	mgr := NewManager("test")

	// 启动管理器
	if err := mgr.OnStart(); err != nil {
		t.Fatalf("OnStart failed: %v", err)
	}
	defer mgr.OnStop()

	ctx := context.Background()

	// 测试 Set
	err := mgr.Set(ctx, "test_key", "test_value", 5*time.Minute)
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	// 测试 Get
	var value string
	err = mgr.Get(ctx, "test_key", &value)
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if value != "test_value" {
		t.Errorf("expected 'test_value', got '%s'", value)
	}

	// 测试 Exists
	exists, err := mgr.Exists(ctx, "test_key")
	if err != nil {
		t.Errorf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("expected key to exist")
	}

	// 测试 Delete
	err = mgr.Delete(ctx, "test_key")
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	// 验证删除后不存在
	exists, err = mgr.Exists(ctx, "test_key")
	if err != nil {
		t.Errorf("Exists after Delete failed: %v", err)
	}
	if exists {
		t.Error("expected key to not exist after delete")
	}
}

// TestManagerSetNX 测试 SetNX 操作
func TestManagerSetNX(t *testing.T) {
	mgr := NewManager("test")

	if err := mgr.OnStart(); err != nil {
		t.Fatalf("OnStart failed: %v", err)
	}
	defer mgr.OnStop()

	ctx := context.Background()

	// 第一次设置，应成功
	ok, err := mgr.SetNX(ctx, "test_key", "value1", 5*time.Minute)
	if err != nil {
		t.Errorf("SetNX failed: %v", err)
	}
	if !ok {
		t.Error("expected SetNX to succeed on first call")
	}

	// 第二次设置，应失败（键已存在）
	ok, err = mgr.SetNX(ctx, "test_key", "value2", 5*time.Minute)
	if err != nil {
		t.Errorf("SetNX failed: %v", err)
	}
	if ok {
		t.Error("expected SetNX to fail on second call")
	}

	// 验证值是第一次设置的值
	var value string
	err = mgr.Get(ctx, "test_key", &value)
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if value != "value1" {
		t.Errorf("expected 'value1', got '%s'", value)
	}
}

// TestManagerIncrementDecrement 测试自增和自减操作
func TestManagerIncrementDecrement(t *testing.T) {
	mgr := NewManager("test")

	if err := mgr.OnStart(); err != nil {
		t.Fatalf("OnStart failed: %v", err)
	}
	defer mgr.OnStop()

	ctx := context.Background()

	// 设置初始值
	err := mgr.Set(ctx, "counter", int64(10), 5*time.Minute)
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	// 自增
	newVal, err := mgr.Increment(ctx, "counter", 5)
	if err != nil {
		t.Errorf("Increment failed: %v", err)
	}
	if newVal != 15 {
		t.Errorf("expected 15, got %d", newVal)
	}

	// 自减
	newVal, err = mgr.Decrement(ctx, "counter", 3)
	if err != nil {
		t.Errorf("Decrement failed: %v", err)
	}
	if newVal != 12 {
		t.Errorf("expected 12, got %d", newVal)
	}
}

// TestManagerBatchOperations 测试批量操作
func TestManagerBatchOperations(t *testing.T) {
	mgr := NewManager("test")

	if err := mgr.OnStart(); err != nil {
		t.Fatalf("OnStart failed: %v", err)
	}
	defer mgr.OnStop()

	ctx := context.Background()

	// 批量设置
	items := map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	err := mgr.SetMultiple(ctx, items, 5*time.Minute)
	if err != nil {
		t.Errorf("SetMultiple failed: %v", err)
	}

	// 批量获取
	keys := []string{"key1", "key2", "key3", "key4"}
	results, err := mgr.GetMultiple(ctx, keys)
	if err != nil {
		t.Errorf("GetMultiple failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// 批量删除
	err = mgr.DeleteMultiple(ctx, keys)
	if err != nil {
		t.Errorf("DeleteMultiple failed: %v", err)
	}

	// 验证删除
	results, err = mgr.GetMultiple(ctx, keys)
	if err != nil {
		t.Errorf("GetMultiple after DeleteMultiple failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results after delete, got %d", len(results))
	}
}
