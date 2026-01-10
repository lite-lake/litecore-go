package cachemgr

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestIntegration_CompleteWorkflow(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]any
	}{
		{
			name: "memory cache workflow",
			config: map[string]any{
				"driver": "memory",
				"memory_config": map[string]any{
					"max_size": 100,
					"max_age":  "1h",
				},
			},
		},
		{
			name: "none cache workflow",
			config: map[string]any{
				"driver": "none",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. 创建缓存管理器
			mgr := Build(tt.config)
			if mgr == nil {
				t.Fatal("Build() should not return nil")
			}

			// 类型断言为 CacheManager 以访问缓存方法
			cacheMgr, ok := mgr.(CacheManager)
			if !ok {
				t.Fatal("Build() should return CacheManager interface")
			}

			// 2. 测试生命周期方法
			if err := mgr.OnStart(); err != nil {
				t.Errorf("OnStart() error = %v", err)
			}

			if err := mgr.Health(); err != nil {
				t.Logf("Health() error = %v (may be expected for some drivers)", err)
			}

			if err := mgr.OnStop(); err != nil {
				t.Errorf("OnStop() error = %v", err)
			}

			// 3. 测试缓存操作
			ctx := context.Background()

			// Set
			err := cacheMgr.Set(ctx, "test_key", "test_value", time.Minute)
			if tt.config["driver"] != "none" && err != nil {
				t.Errorf("Set() error = %v", err)
			}

			// Get
			var value string
			err = cacheMgr.Get(ctx, "test_key", &value)
			if tt.config["driver"] == "none" {
				if err == nil {
					t.Error("Get() should return error for none driver")
				}
			} else {
				if err != nil {
					t.Errorf("Get() error = %v", err)
				}
				if value != "test_value" {
					t.Errorf("Get() = %v, want %v", value, "test_value")
				}
			}

			// Exists
			exists, err := cacheMgr.Exists(ctx, "test_key")
			if err != nil {
				t.Errorf("Exists() error = %v", err)
			}
			if tt.config["driver"] == "none" {
				if exists {
					t.Error("Exists() should return false for none driver")
				}
			} else {
				if !exists {
					t.Error("Exists() should return true for existing key")
				}
			}

			// Delete
			err = cacheMgr.Delete(ctx, "test_key")
			if err != nil {
				t.Errorf("Delete() error = %v", err)
			}

			// 4. 测试批量操作
			items := map[string]any{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}
			err = cacheMgr.SetMultiple(ctx, items, time.Minute)
			if tt.config["driver"] != "none" && err != nil {
				t.Errorf("SetMultiple() error = %v", err)
			}

			results, err := cacheMgr.GetMultiple(ctx, []string{"key1", "key2", "key3"})
			if err != nil {
				t.Errorf("GetMultiple() error = %v", err)
			}
			if tt.config["driver"] != "none" && len(results) != 3 {
				t.Errorf("GetMultiple() returned %d items, want 3", len(results))
			}

			// 5. 测试计数器操作
			counter, err := cacheMgr.Increment(ctx, "counter", 1)
			if err != nil {
				t.Errorf("Increment() error = %v", err)
			}
			if counter < 0 {
				t.Errorf("Increment() returned %d, want >= 0", counter)
			}

			// 6. 关闭管理器
			if err := cacheMgr.Close(); err != nil {
				t.Errorf("Close() error = %v", err)
			}
		})
	}
}

func TestIntegration_DriverSwitching(t *testing.T) {
	// 测试在不同驱动之间切换的场景
	drivers := []string{"none", "memory", "none"}

	var prevMgr string
	for _, driver := range drivers {
		var cfg map[string]any
		if driver == "memory" {
			cfg = map[string]any{
				"driver": "memory",
				"memory_config": map[string]any{
					"max_size": 100,
					"max_age":  "1h",
				},
			}
		} else {
			cfg = map[string]any{
				"driver": driver,
			}
		}

		mgr := Build(cfg)
		currentMgr := mgr.ManagerName()

		if prevMgr != "" && prevMgr == currentMgr && drivers[0] != drivers[1] {
			t.Logf("Successfully switched from %s to %s", prevMgr, currentMgr)
		}
		prevMgr = currentMgr

		// 使用 CacheManager 接口的 Close 方法
		cacheMgr := mgr.(CacheManager)
		cacheMgr.Close()
	}
}

func TestIntegration_ConfigVariations(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]any
		expectMgr   string
		expectError bool
	}{
		{
			name: "minimal memory config",
			config: map[string]any{
				"driver": "memory",
			},
			expectMgr:   "memory-cache",
			expectError: false,
		},
		{
			name: "minimal none config",
			config: map[string]any{
				"driver": "none",
			},
			expectMgr:   "none-cache",
			expectError: false,
		},
		{
			name: "memory with all options",
			config: map[string]any{
				"driver": "memory",
				"memory_config": map[string]any{
					"max_size":    500,
					"max_age":     "2h",
					"max_backups": 2000,
					"compress":    true,
				},
			},
			expectMgr:   "memory-cache",
			expectError: false,
		},
		{
			name: "driver case insensitive",
			config: map[string]any{
				"driver": "MEMORY",
				"memory_config": map[string]any{
					"max_size": 100,
				},
			},
			expectMgr:   "memory-cache",
			expectError: false,
		},
		{
			name: "driver with spaces",
			config: map[string]any{
				"driver": "  memory  ",
				"memory_config": map[string]any{
					"max_size": 100,
				},
			},
			expectMgr:   "memory-cache",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := Build(tt.config)
			if mgr.ManagerName() != tt.expectMgr {
				t.Errorf("ManagerName() = %v, want %v", mgr.ManagerName(), tt.expectMgr)
			}

			// 验证管理器可以正常工作
			cacheMgr := mgr.(CacheManager)
			ctx := context.Background()
			err := cacheMgr.Set(ctx, "test", "value", time.Minute)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			}
			if tt.expectMgr != "none-cache" && err != nil {
				t.Errorf("Set() unexpected error = %v", err)
			}

			cacheMgr.Close()
		})
	}
}

func TestIntegration_LifecycleManagement(t *testing.T) {
	cfg := map[string]any{
		"driver": "memory",
		"memory_config": map[string]any{
			"max_size": 100,
			"max_age":  "1h",
		},
	}

	mgr := Build(cfg)
	cacheMgr := mgr.(CacheManager)
	ctx := context.Background()

	// 设置一些数据
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		if err := cacheMgr.Set(ctx, key, value, time.Minute); err != nil {
			t.Errorf("Set() error = %v", err)
		}
	}

	// OnStart 不应该影响已有数据
	if err := mgr.OnStart(); err != nil {
		t.Errorf("OnStart() error = %v", err)
	}

	// 验证数据仍然存在
	exists, err := cacheMgr.Exists(ctx, "key5")
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Data should exist after OnStart()")
	}

	// OnStop 不应该影响数据
	if err := mgr.OnStop(); err != nil {
		t.Errorf("OnStop() error = %v", err)
	}

	// 验证数据仍然存在
	exists, err = cacheMgr.Exists(ctx, "key5")
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Data should exist after OnStop()")
	}

	// Close 后可能无法访问数据
	if err := cacheMgr.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestIntegration_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	cfg := map[string]any{
		"driver": "memory",
		"memory_config": map[string]any{
			"max_size": 1000,
			"max_age":  "1h",
		},
	}

	mgr := Build(cfg)
	cacheMgr := mgr.(CacheManager)
	defer cacheMgr.Close()

	ctx := context.Background()
	const numOperations = 1000

	// 执行大量操作
	for i := 0; i < numOperations; i++ {
		key := fmt.Sprintf("stress_key_%d", i)
		value := fmt.Sprintf("stress_value_%d", i)

		if err := cacheMgr.Set(ctx, key, value, time.Minute); err != nil {
			t.Errorf("Set() error at iteration %d: %v", i, err)
		}

		if i%100 == 0 {
			if _, err := cacheMgr.Increment(ctx, "stress_counter", 1); err != nil {
				t.Errorf("Increment() error at iteration %d: %v", i, err)
			}
		}
	}

	// 验证计数器
	counter, err := cacheMgr.Increment(ctx, "stress_counter", 0)
	if err != nil {
		t.Errorf("Increment() error = %v", err)
	}
	expected := int64(numOperations / 100)
	if counter != expected {
		t.Errorf("Counter = %d, want %d", counter, expected)
	}
}

func TestIntegration_ErrorRecovery(t *testing.T) {
	// 测试错误恢复能力
	cfg := map[string]any{
		"driver": "memory",
		"memory_config": map[string]any{
			"max_size": 100,
			"max_age":  "1h",
		},
	}

	mgr := Build(cfg)
	cacheMgr := mgr.(CacheManager)
	defer cacheMgr.Close()

	ctx := context.Background()

	// 尝试一些可能失败的操作
	var invalidDest string
	err := cacheMgr.Get(ctx, "nonexistent_key", &invalidDest)
	if err == nil {
		t.Log("Get() of nonexistent key returned nil error (may be expected for some drivers)")
	}

	// 设置一些值
	if err := cacheMgr.Set(ctx, "key1", "value1", time.Minute); err != nil {
		t.Errorf("Set() error = %v", err)
	}

	// 验证后续操作仍然正常工作
	if err := cacheMgr.Set(ctx, "key2", "value2", time.Minute); err != nil {
		t.Errorf("Set() after error should still work: %v", err)
	}

	exists, err := cacheMgr.Exists(ctx, "key1")
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("key1 should exist")
	}
}
