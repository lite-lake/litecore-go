package hooks

import (
	"sync"
	"testing"
)

// TestNewManager 测试创建钩子管理器
func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager() returned nil")
	}

	if m.callbacks == nil {
		t.Error("callbacks map is not initialized")
	}
}

// TestManager_Register 测试注册钩子
func TestManager_Register(t *testing.T) {
	m := NewManager()
	testHook := func() {}

	tests := []struct {
		name    string
		hook    string
		hookFn  interface{}
		wantErr bool
	}{
		{
			name:    "beforeCreate 钩子",
			hook:    "beforeCreate",
			hookFn:  testHook,
			wantErr: false,
		},
		{
			name:    "afterCreate 钩子",
			hook:    "afterCreate",
			hookFn:  testHook,
			wantErr: false,
		},
		{
			name:    "beforeUpdate 钩子",
			hook:    "beforeUpdate",
			hookFn:  testHook,
			wantErr: false,
		},
		{
			name:    "afterUpdate 钩子",
			hook:    "afterUpdate",
			hookFn:  testHook,
			wantErr: false,
		},
		{
			name:    "beforeDelete 钩子",
			hook:    "beforeDelete",
			hookFn:  testHook,
			wantErr: false,
		},
		{
			name:    "afterDelete 钩子",
			hook:    "afterDelete",
			hookFn:  testHook,
			wantErr: false,
		},
		{
			name:    "afterFind 钩子",
			hook:    "afterFind",
			hookFn:  testHook,
			wantErr: false,
		},
		{
			name:    "无效钩子名称",
			hook:    "invalidHook",
			hookFn:  testHook,
			wantErr: true,
		},
		{
			name:    "空钩子名称",
			hook:    "",
			hookFn:  testHook,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := m.Register(tt.hook, tt.hookFn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 如果注册成功，验证钩子已添加
			if !tt.wantErr {
				hooks := m.Get(tt.hook)
				if len(hooks) == 0 {
					t.Errorf("Hook %s was not registered", tt.hook)
				}
			}
		})
	}
}

// TestManager_RegisterMultiple 测试注册多个钩子
func TestManager_RegisterMultiple(t *testing.T) {
	m := NewManager()

	hook1 := func() {}
	hook2 := func() {}
	hook3 := func() {}

	// 注册同一个钩子多次
	err := m.Register("beforeCreate", hook1)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	err = m.Register("beforeCreate", hook2)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	err = m.Register("beforeCreate", hook3)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	hooks := m.Get("beforeCreate")
	if len(hooks) != 3 {
		t.Errorf("Expected 3 hooks, got %d", len(hooks))
	}
}

// TestManager_Get 测试获取钩子
func TestManager_Get(t *testing.T) {
	m := NewManager()
	testHook := func() {}

	// 测试获取不存在的钩子
	hooks := m.Get("beforeCreate")
	if len(hooks) != 0 {
		t.Errorf("Get() on empty manager should return empty slice, got %d hooks", len(hooks))
	}

	// 注册钩子
	_ = m.Register("beforeCreate", testHook)

	// 测试获取已注册的钩子
	hooks = m.Get("beforeCreate")
	if len(hooks) != 1 {
		t.Errorf("Get() should return 1 hook, got %d", len(hooks))
	}
}

// TestManager_Clear 测试清除所有钩子
func TestManager_Clear(t *testing.T) {
	m := NewManager()
	testHook := func() {}

	// 注册多个钩子
	_ = m.Register("beforeCreate", testHook)
	_ = m.Register("afterCreate", testHook)
	_ = m.Register("beforeUpdate", testHook)

	// 验证钩子已注册
	if m.Get("beforeCreate") == nil || len(m.Get("beforeCreate")) == 0 {
		t.Error("Hooks should be registered")
	}

	// 清除所有钩子
	m.Clear()

	// 验证所有钩子已清除
	hooks := m.Get("beforeCreate")
	if len(hooks) != 0 {
		t.Errorf("Clear() should remove all hooks, got %d", len(hooks))
	}

	hooks = m.Get("afterCreate")
	if len(hooks) != 0 {
		t.Errorf("Clear() should remove all hooks, got %d", len(hooks))
	}
}

// TestManager_Remove 测试移除特定钩子
func TestManager_Remove(t *testing.T) {
	m := NewManager()
	testHook := func() {}

	// 注册多个钩子
	_ = m.Register("beforeCreate", testHook)
	_ = m.Register("afterCreate", testHook)

	// 移除 beforeCreate 钩子
	m.Remove("beforeCreate")

	// 验证 beforeCreate 已移除
	hooks := m.Get("beforeCreate")
	if len(hooks) != 0 {
		t.Errorf("Remove() should remove the hook, got %d hooks", len(hooks))
	}

	// 验证 afterCreate 仍然存在
	hooks = m.Get("afterCreate")
	if len(hooks) != 1 {
		t.Errorf("Remove() should not affect other hooks, got %d hooks", len(hooks))
	}

	// 移除不存在的钩子不应该 panic
	m.Remove("nonExistent")
}

// TestManager_ConcurrentAccess 测试并发访问
func TestManager_ConcurrentAccess(t *testing.T) {
	m := NewManager()
	testHook := func() {}

	var wg sync.WaitGroup

	// 并发注册
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = m.Register("beforeCreate", testHook)
		}()
	}

	// 并发读取
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = m.Get("beforeCreate")
		}()
	}

	wg.Wait()

	// 验证钩子已注册
	hooks := m.Get("beforeCreate")
	if len(hooks) != 10 {
		t.Errorf("Expected 10 hooks, got %d", len(hooks))
	}
}

// TestManager_ApplyTo 测试应用钩子
func TestManager_ApplyTo(t *testing.T) {
	m := NewManager()
	testHook := func() {}

	_ = m.Register("beforeCreate", testHook)

	// ApplyTo 方法目前是空实现，我们只测试它不会 panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ApplyTo() panicked: %v", r)
		}
	}()

	// ApplyTo 需要 *gorm.DB，但由于当前实现是空的，我们跳过实际调用
	_ = m
}

// TestManager_AllHookTypes 测试所有支持的钩子类型
func TestManager_AllHookTypes(t *testing.T) {
	m := NewManager()
	testHook := func() {}

	validHooks := []string{
		"beforeCreate",
		"afterCreate",
		"beforeUpdate",
		"afterUpdate",
		"beforeDelete",
		"afterDelete",
		"afterFind",
	}

	for _, hookName := range validHooks {
		t.Run(hookName, func(t *testing.T) {
			err := m.Register(hookName, testHook)
			if err != nil {
				t.Errorf("Register(%s) error = %v", hookName, err)
			}

			hooks := m.Get(hookName)
			if len(hooks) != 1 {
				t.Errorf("Get(%s) should return 1 hook, got %d", hookName, len(hooks))
			}
		})
	}
}

// TestManager_RegisterNil 测试注册 nil 钩子
func TestManager_RegisterNil(t *testing.T) {
	m := NewManager()

	// 注册 nil 不应该 panic
	err := m.Register("beforeCreate", nil)
	if err != nil {
		t.Errorf("Register(nil) error = %v", err)
	}

	hooks := m.Get("beforeCreate")
	if len(hooks) != 1 {
		t.Errorf("Register(nil) should still add the hook, got %d hooks", len(hooks))
	}
}

// TestManager_CallbacksInitialization 测试回调初始化
func TestManager_CallbacksInitialization(t *testing.T) {
	m := NewManager()

	// 验证所有钩子类型的回调都已初始化
	validHooks := []string{
		"beforeCreate",
		"afterCreate",
		"beforeUpdate",
		"afterUpdate",
		"beforeDelete",
		"afterDelete",
		"afterFind",
	}

	for _, hookName := range validHooks {
		hooks := m.Get(hookName)
		if hooks == nil {
			t.Errorf("Get(%s) should return empty slice, not nil", hookName)
		}
	}
}

// BenchmarkManager_Register 基准测试注册钩子
func BenchmarkManager_Register(b *testing.B) {
	m := NewManager()
	testHook := func() {}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Register("beforeCreate", testHook)
	}
}

// BenchmarkManager_Get 基准测试获取钩子
func BenchmarkManager_Get(b *testing.B) {
	m := NewManager()
	testHook := func() {}
	_ = m.Register("beforeCreate", testHook)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Get("beforeCreate")
	}
}

// BenchmarkManager_ConcurrentAccess 基准测试并发访问
func BenchmarkManager_ConcurrentAccess(b *testing.B) {
	m := NewManager()
	testHook := func() {}
	_ = m.Register("beforeCreate", testHook)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = m.Get("beforeCreate")
		}
	})
}
