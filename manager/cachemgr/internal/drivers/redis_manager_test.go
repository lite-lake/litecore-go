package drivers

import (
	"context"
	"strconv"
	"testing"
	"time"

	"com.litelake.litecore/manager/cachemgr/internal/config"
	"github.com/alicebob/miniredis/v2"
)

func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *RedisManager) {
	// 启动 miniredis
	s := miniredis.RunT(t)

	// 解析端口号（miniredis v2 的 Port() 返回字符串）
	port, err := strconv.Atoi(s.Port())
	if err != nil {
		t.Fatalf("Failed to parse port: %v", err)
	}

	// 创建 Redis 配置
	cfg := &config.RedisConfig{
		Host:            s.Host(),
		Port:            port,
		Password:        "",
		DB:              0,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 30 * time.Second,
	}

	// 创建 Redis 管理器
	mgr, err := NewRedisManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create Redis manager: %v", err)
	}

	return s, mgr
}

func TestNewRedisManager(t *testing.T) {
	_, mgr := setupTestRedis(t)

	if mgr.ManagerName() != "redis-cache" {
		t.Errorf("ManagerName() = %v, want %v", mgr.ManagerName(), "redis-cache")
	}

	// 测试健康检查
	if err := mgr.Health(); err != nil {
		t.Errorf("Health() error = %v", err)
	}
}

func TestRedisManager_BasicOperations(t *testing.T) {
	s, mgr := setupTestRedis(t)
	defer s.Close()

	ctx := context.Background()

	// 测试 Set 和 Get
	err := mgr.Set(ctx, "key1", "value1", time.Minute)
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	var value string
	err = mgr.Get(ctx, "key1", &value)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if value != "value1" {
		t.Errorf("Get() = %v, want %v", value, "value1")
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

func TestRedisManager_SetNX(t *testing.T) {
	s, mgr := setupTestRedis(t)
	defer s.Close()

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
	var value string
	err = mgr.Get(ctx, "key1", &value)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if value != "value1" {
		t.Errorf("value = %v, want %v", value, "value1")
	}
}

func TestRedisManager_BatchOperations(t *testing.T) {
	s, mgr := setupTestRedis(t)
	defer s.Close()

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
	count, _ := mgr.client.Exists(ctx, "key1", "key2", "key3").Result()
	if count != 1 {
		t.Errorf("After DeleteMultiple(), got %d items, want 1", count)
	}
}

func TestRedisManager_CounterOperations(t *testing.T) {
	s, mgr := setupTestRedis(t)
	defer s.Close()

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

func TestRedisManager_TTL(t *testing.T) {
	s, mgr := setupTestRedis(t)
	defer s.Close()

	ctx := context.Background()

	// 设置一个带过期时间的值
	err := mgr.Set(ctx, "key1", "value1", time.Minute)
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	// 获取 TTL
	ttl, err := mgr.TTL(ctx, "key1")
	if err != nil {
		t.Errorf("TTL() error = %v", err)
	}
	if ttl <= 0 {
		t.Errorf("TTL() = %v, want > 0", ttl)
	}

	// 修改过期时间
	err = mgr.Expire(ctx, "key1", 5*time.Minute)
	if err != nil {
		t.Errorf("Expire() error = %v", err)
	}

	// 验证新的 TTL
	ttl, err = mgr.TTL(ctx, "key1")
	if err != nil {
		t.Errorf("TTL() error = %v", err)
	}
	if ttl <= 0 {
		t.Errorf("TTL() after Expire() = %v, want > 0", ttl)
	}
}

func TestRedisManager_Clear(t *testing.T) {
	s, mgr := setupTestRedis(t)
	defer s.Close()

	ctx := context.Background()

	// 添加一些数据
	items := map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	mgr.SetMultiple(ctx, items, time.Minute)

	// 清空
	err := mgr.Clear(ctx)
	if err != nil {
		t.Errorf("Clear() error = %v", err)
	}

	// 验证清空
	count, _ := mgr.client.DBSize(ctx).Result()
	if count != 0 {
		t.Errorf("After Clear(), DBSize() = %d, want 0", count)
	}
}

func TestRedisManager_ContextValidation(t *testing.T) {
	_, mgr := setupTestRedis(t)

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

func TestRedisManager_KeyValidation(t *testing.T) {
	_, mgr := setupTestRedis(t)
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

func TestRedisManager_Serialization(t *testing.T) {
	s, mgr := setupTestRedis(t)
	defer s.Close()

	ctx := context.Background()

	// 测试不同类型的序列化
	testCases := []struct {
		name  string
		key   string
		value any
		dest  any
	}{
		{
			name:  "string value",
			key:   "str_key",
			value: "test string",
			dest:  new(string),
		},
		{
			name:  "int value",
			key:   "int_key",
			value: 42,
			dest:  new(int),
		},
		{
			name:  "float value",
			key:   "float_key",
			value: 3.14,
			dest:  new(float64),
		},
		{
			name:  "bool value",
			key:   "bool_key",
			value: true,
			dest:  new(bool),
		},
		{
			name:  "slice value",
			key:   "slice_key",
			value: []int{1, 2, 3, 4, 5},
			dest:  new([]int),
		},
		{
			name:  "map value",
			key:   "map_key",
			value: map[string]string{"a": "apple", "b": "banana"},
			dest:  new(map[string]string),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 设置值
			err := mgr.Set(ctx, tc.key, tc.value, time.Minute)
			if err != nil {
				t.Errorf("Set() error = %v", err)
			}

			// 获取值
			err = mgr.Get(ctx, tc.key, tc.dest)
			if err != nil {
				t.Errorf("Get() error = %v", err)
			}

			// 验证值 - 通过类型断言获取实际值
			var destValue any
			switch d := tc.dest.(type) {
			case *string:
				destValue = *d
			case *int:
				destValue = *d
			case *float64:
				destValue = *d
			case *bool:
				destValue = *d
			case *[]int:
				destValue = *d
			case *map[string]string:
				destValue = *d
			}

			if !compareValues(destValue, tc.value) {
				t.Errorf("Get() = %v, want %v", destValue, tc.value)
			}
		})
	}
}

func TestRedisManager_ErrorHandling(t *testing.T) {
	s, mgr := setupTestRedis(t)
	defer s.Close()

	ctx := context.Background()

	// 测试获取不存在的键
	var value string
	err := mgr.Get(ctx, "nonexistent_key", &value)
	if err == nil {
		t.Error("Get() with nonexistent key should return error")
	}

	// 测试获取不存在的键的 TTL
	ttl, err := mgr.TTL(ctx, "nonexistent_key")
	if err != nil {
		t.Errorf("TTL() with nonexistent key returned error: %v", err)
	}
	// Redis 返回 -2 表示键不存在
	if ttl != -2*time.Second && ttl != -2 {
		t.Logf("TTL() for nonexistent key = %v (Redis typically returns -2)", ttl)
	}

	// 测试对不存在的键设置过期时间
	err = mgr.Expire(ctx, "nonexistent_key", time.Minute)
	if err != nil {
		// Redis 通常允许对不存在的键设置过期时间,但返回 0
		t.Logf("Expire() with nonexistent key returned error: %v", err)
	}

	// 测试空批量操作
	result, err := mgr.GetMultiple(ctx, []string{})
	if err != nil {
		t.Errorf("GetMultiple() with empty keys returned error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("GetMultiple() with empty keys should return empty map, got %d items", len(result))
	}

	err = mgr.SetMultiple(ctx, map[string]any{}, time.Minute)
	if err != nil {
		t.Errorf("SetMultiple() with empty items returned error: %v", err)
	}

	err = mgr.DeleteMultiple(ctx, []string{})
	if err != nil {
		t.Errorf("DeleteMultiple() with empty keys returned error: %v", err)
	}
}

func TestRedisManager_Close(t *testing.T) {
	s, mgr := setupTestRedis(t)
	defer s.Close()

	// 测试关闭
	err := mgr.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// 关闭后应该无法正常使用
	err = mgr.Health()
	if err == nil {
		t.Error("Health() should return error after Close()")
	}
}

// compareValues 比较两个值是否相等
func compareValues(a, b any) bool {
	switch va := a.(type) {
	case []byte:
		if vb, ok := b.([]byte); ok {
			if len(va) != len(vb) {
				return false
			}
			for i := range va {
				if va[i] != vb[i] {
					return false
				}
			}
			return true
		}
		return false
	case []int:
		if vb, ok := b.([]int); ok {
			if len(va) != len(vb) {
				return false
			}
			for i := range va {
				if va[i] != vb[i] {
					return false
				}
			}
			return true
		}
		return false
	case []string:
		if vb, ok := b.([]string); ok {
			if len(va) != len(vb) {
				return false
			}
			for i := range va {
				if va[i] != vb[i] {
					return false
				}
			}
			return true
		}
		return false
	case map[string]string:
		if vb, ok := b.(map[string]string); ok {
			if len(va) != len(vb) {
				return false
			}
			for k, v := range va {
				if vb[k] != v {
					return false
				}
			}
			return true
		}
		return false
	default:
		// 使用简单的相等比较
		return a == b
	}
}
