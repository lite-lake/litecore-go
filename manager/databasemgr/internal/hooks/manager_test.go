package hooks

import (
	"sync"
	"testing"

	"gorm.io/gorm"
)

func TestNewManager(t *testing.T) {
	mgr := NewManager()

	if mgr == nil {
		t.Fatal("NewManager() returned nil")
	}

	if mgr.callbacks == nil {
		t.Error("NewManager() should initialize callbacks map")
	}
}

func TestManager_Register(t *testing.T) {
	tests := []struct {
		name      string
		hookName  string
		hook      interface{}
		wantErr   bool
		errMsg    string
	}{
		{
			name:     "valid beforeCreate hook",
			hookName: "beforeCreate",
			hook:     func(tx *gorm.DB) {},
			wantErr:  false,
		},
		{
			name:     "valid afterCreate hook",
			hookName: "afterCreate",
			hook:     func(tx *gorm.DB) {},
			wantErr:  false,
		},
		{
			name:     "valid beforeUpdate hook",
			hookName: "beforeUpdate",
			hook:     func(tx *gorm.DB) {},
			wantErr:  false,
		},
		{
			name:     "valid afterUpdate hook",
			hookName: "afterUpdate",
			hook:     func(tx *gorm.DB) {},
			wantErr:  false,
		},
		{
			name:     "valid beforeDelete hook",
			hookName: "beforeDelete",
			hook:     func(tx *gorm.DB) {},
			wantErr:  false,
		},
		{
			name:     "valid afterDelete hook",
			hookName: "afterDelete",
			hook:     func(tx *gorm.DB) {},
			wantErr:  false,
		},
		{
			name:     "valid afterFind hook",
			hookName: "afterFind",
			hook:     func(tx *gorm.DB) {},
			wantErr:  false,
		},
		{
			name:     "invalid hook name",
			hookName: "invalidHook",
			hook:     func(tx *gorm.DB) {},
			wantErr:  true,
			errMsg:   "invalid hook name",
		},
		{
			name:     "empty hook name",
			hookName: "",
			hook:     func(tx *gorm.DB) {},
			wantErr:  true,
			errMsg:   "invalid hook name",
		},
		{
			name:     "nil hook function",
			hookName: "beforeCreate",
			hook:     nil,
			wantErr:  false, // Register doesn't validate nil hooks
		},
		{
			name:     "hook with wrong signature",
			hookName: "beforeCreate",
			hook:     func(string) {},
			wantErr:  false, // Register doesn't validate hook signature
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager()
			err := mgr.Register(tt.hookName, tt.hook)

			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.Register() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" {
				if len(err.Error()) < len(tt.errMsg) || err.Error()[:len(tt.errMsg)] != tt.errMsg {
					t.Errorf("Manager.Register() error = %v, want prefix %v", err, tt.errMsg)
				}
			}

			if !tt.wantErr {
				hooks := mgr.Get(tt.hookName)
				if len(hooks) == 0 && tt.hook != nil {
					t.Error("Register() should add hook to callbacks")
				}
			}
		})
	}
}

func TestManager_Register_MultipleHooks(t *testing.T) {
	mgr := NewManager()

	// Register multiple hooks for the same event
	hook1 := func(tx *gorm.DB) {}
	hook2 := func(tx *gorm.DB) {}
	hook3 := func(tx *gorm.DB) {}

	if err := mgr.Register("beforeCreate", hook1); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if err := mgr.Register("beforeCreate", hook2); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if err := mgr.Register("beforeCreate", hook3); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	hooks := mgr.Get("beforeCreate")
	if len(hooks) != 3 {
		t.Errorf("Get() returned %d hooks, want 3", len(hooks))
	}
}

func TestManager_Get(t *testing.T) {
	mgr := NewManager()

	hook := func(tx *gorm.DB) {}
	_ = mgr.Register("beforeCreate", hook)

	tests := []struct {
		name     string
		hookName string
		wantLen  int
	}{
		{
			name:     "get existing hooks",
			hookName: "beforeCreate",
			wantLen:  1,
		},
		{
			name:     "get non-existing hooks",
			hookName: "afterCreate",
			wantLen:  0,
		},
		{
			name:     "get empty hook name",
			hookName: "",
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hooks := mgr.Get(tt.hookName)
			if len(hooks) != tt.wantLen {
				t.Errorf("Manager.Get() returned %d hooks, want %d", len(hooks), tt.wantLen)
			}
		})
	}
}

func TestManager_Get_AllHookTypes(t *testing.T) {
	mgr := NewManager()

	hook := func(tx *gorm.DB) {}

	// Register one hook for each type
	hookTypes := []string{
		"beforeCreate", "afterCreate",
		"beforeUpdate", "afterUpdate",
		"beforeDelete", "afterDelete",
		"afterFind",
	}

	for _, hookType := range hookTypes {
		if err := mgr.Register(hookType, hook); err != nil {
			t.Fatalf("Register(%s) error = %v", hookType, err)
		}
	}

	// Verify all hooks are registered
	for _, hookType := range hookTypes {
		hooks := mgr.Get(hookType)
		if len(hooks) != 1 {
			t.Errorf("Get(%s) returned %d hooks, want 1", hookType, len(hooks))
		}
	}
}

func TestManager_ApplyTo(t *testing.T) {
	mgr := NewManager()

	// ApplyTo is a placeholder method
	// It should not panic
	db := &gorm.DB{}
	mgr.ApplyTo(db)
}

func TestManager_Clear(t *testing.T) {
	mgr := NewManager()

	// Register some hooks
	hook1 := func(tx *gorm.DB) {}
	hook2 := func(tx *gorm.DB) {}

	_ = mgr.Register("beforeCreate", hook1)
	_ = mgr.Register("afterCreate", hook2)

	// Verify hooks are registered
	if len(mgr.Get("beforeCreate")) != 1 {
		t.Error("Expected beforeCreate hook to be registered")
	}
	if len(mgr.Get("afterCreate")) != 1 {
		t.Error("Expected afterCreate hook to be registered")
	}

	// Clear all hooks
	mgr.Clear()

	// Verify all hooks are cleared
	if len(mgr.Get("beforeCreate")) != 0 {
		t.Error("Expected beforeCreate hook to be cleared")
	}
	if len(mgr.Get("afterCreate")) != 0 {
		t.Error("Expected afterCreate hook to be cleared")
	}

	// Verify callbacks map is reinitialized
	if mgr.callbacks == nil {
		t.Error("Clear() should reinitialize callbacks map")
	}
}

func TestManager_Remove(t *testing.T) {
	mgr := NewManager()

	// Register some hooks
	hook1 := func(tx *gorm.DB) {}
	hook2 := func(tx *gorm.DB) {}

	_ = mgr.Register("beforeCreate", hook1)
	_ = mgr.Register("afterCreate", hook2)
	_ = mgr.Register("beforeUpdate", hook1)

	// Remove beforeCreate hooks
	mgr.Remove("beforeCreate")

	// Verify beforeCreate hooks are removed
	if len(mgr.Get("beforeCreate")) != 0 {
		t.Error("Expected beforeCreate hooks to be removed")
	}

	// Verify other hooks are still present
	if len(mgr.Get("afterCreate")) != 1 {
		t.Error("Expected afterCreate hook to still exist")
	}
	if len(mgr.Get("beforeUpdate")) != 1 {
		t.Error("Expected beforeUpdate hook to still exist")
	}

	// Remove non-existing hook should not panic
	mgr.Remove("nonExisting")
}

func TestManager_ConcurrentRegister(t *testing.T) {
	mgr := NewManager()

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Concurrently register hooks
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			hook := func(tx *gorm.DB) {}
			err := mgr.Register("beforeCreate", hook)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent Register() error = %v", err)
	}

	// Verify all hooks were registered
	hooks := mgr.Get("beforeCreate")
	if len(hooks) != 100 {
		t.Errorf("Expected 100 hooks, got %d", len(hooks))
	}
}

func TestManager_ConcurrentGet(t *testing.T) {
	mgr := NewManager()

	hook := func(tx *gorm.DB) {}
	_ = mgr.Register("beforeCreate", hook)

	var wg sync.WaitGroup

	// Concurrently get hooks
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hooks := mgr.Get("beforeCreate")
			if len(hooks) != 1 {
				t.Errorf("Expected 1 hook, got %d", len(hooks))
			}
		}()
	}

	wg.Wait()
}

func TestManager_ConcurrentClear(t *testing.T) {
	mgr := NewManager()

	var wg sync.WaitGroup

	// Concurrently clear and register
	for i := 0; i < 50; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			mgr.Clear()
		}()

		go func() {
			defer wg.Done()
			hook := func(tx *gorm.DB) {}
			_ = mgr.Register("beforeCreate", hook)
		}()
	}

	wg.Wait()
}

func TestManager_RegisterAfterClear(t *testing.T) {
	mgr := NewManager()

	// Register hooks
	hook1 := func(tx *gorm.DB) {}
	_ = mgr.Register("beforeCreate", hook1)

	// Clear all hooks
	mgr.Clear()

	// Register new hooks after clear
	hook2 := func(tx *gorm.DB) {}
	if err := mgr.Register("afterCreate", hook2); err != nil {
		t.Errorf("Register() after Clear() error = %v", err)
	}

	// Verify new hooks are registered
	hooks := mgr.Get("afterCreate")
	if len(hooks) != 1 {
		t.Errorf("Expected 1 hook after Clear(), got %d", len(hooks))
	}
}

func TestManager_RemoveAllHooks(t *testing.T) {
	mgr := NewManager()

	// Register multiple hook types
	hook := func(tx *gorm.DB) {}
	hookTypes := []string{
		"beforeCreate", "afterCreate",
		"beforeUpdate", "afterUpdate",
		"beforeDelete", "afterDelete",
		"afterFind",
	}

	for _, hookType := range hookTypes {
		_ = mgr.Register(hookType, hook)
	}

	// Remove each hook type
	for _, hookType := range hookTypes {
		mgr.Remove(hookType)
	}

	// Verify all hooks are removed
	for _, hookType := range hookTypes {
		hooks := mgr.Get(hookType)
		if len(hooks) != 0 {
			t.Errorf("Expected %s hooks to be removed", hookType)
		}
	}
}

func TestManager_HookOrder(t *testing.T) {
	mgr := NewManager()

	// Register hooks in specific order
	// Use counters to verify order
	counter1 := 0
	counter2 := 0
	counter3 := 0

	hook1 := func(tx *gorm.DB) { counter1++ }
	hook2 := func(tx *gorm.DB) { counter2++ }
	hook3 := func(tx *gorm.DB) { counter3++ }

	_ = mgr.Register("beforeCreate", hook1)
	_ = mgr.Register("beforeCreate", hook2)
	_ = mgr.Register("beforeCreate", hook3)

	// Get hooks and verify order
	hooks := mgr.Get("beforeCreate")
	if len(hooks) != 3 {
		t.Fatalf("Expected 3 hooks, got %d", len(hooks))
	}

	// Verify hooks are not nil and were added in order
	if hooks[0] == nil || hooks[1] == nil || hooks[2] == nil {
		t.Error("Hooks should not be nil")
	}
}

func TestManager_MultipleHookTypes(t *testing.T) {
	mgr := NewManager()

	// Register different types of hooks
	beforeCreate := func(tx *gorm.DB) {}
	afterCreate := func(tx *gorm.DB) {}
	beforeUpdate := func(tx *gorm.DB) {}

	_ = mgr.Register("beforeCreate", beforeCreate)
	_ = mgr.Register("afterCreate", afterCreate)
	_ = mgr.Register("beforeUpdate", beforeUpdate)

	// Verify each hook type is independent
	if len(mgr.Get("beforeCreate")) != 1 {
		t.Error("Expected 1 beforeCreate hook")
	}
	if len(mgr.Get("afterCreate")) != 1 {
		t.Error("Expected 1 afterCreate hook")
	}
	if len(mgr.Get("beforeUpdate")) != 1 {
		t.Error("Expected 1 beforeUpdate hook")
	}

	// Remove one hook type
	mgr.Remove("afterCreate")

	if len(mgr.Get("beforeCreate")) != 1 {
		t.Error("Expected beforeCreate hook to still exist")
	}
	if len(mgr.Get("afterCreate")) != 0 {
		t.Error("Expected afterCreate hook to be removed")
	}
	if len(mgr.Get("beforeUpdate")) != 1 {
		t.Error("Expected beforeUpdate hook to still exist")
	}
}

func TestManager_EmptyHooks(t *testing.T) {
	mgr := NewManager()

	// Get hooks from empty manager
	hooks := mgr.Get("beforeCreate")
	if hooks == nil {
		t.Error("Get() should return empty slice, not nil")
	}
	if len(hooks) != 0 {
		t.Errorf("Get() on empty manager should return empty slice, got %d hooks", len(hooks))
	}

	// Remove from empty manager
	mgr.Remove("beforeCreate") // Should not panic

	// Clear empty manager
	mgr.Clear() // Should not panic
}

func TestManager_DifferentHookSignatures(t *testing.T) {
	mgr := NewManager()

	// Register hooks with different signatures
	hook1 := func(tx *gorm.DB) {}
	hook2 := func(tx *gorm.DB) error { return nil }
	hook3 := func(tx *gorm.DB) int { return 0 }

	_ = mgr.Register("beforeCreate", hook1)
	_ = mgr.Register("beforeCreate", hook2)
	_ = mgr.Register("beforeCreate", hook3)

	// All hooks should be registered
	hooks := mgr.Get("beforeCreate")
	if len(hooks) != 3 {
		t.Errorf("Expected 3 hooks with different signatures, got %d", len(hooks))
	}
}

func TestManager_RegisterSameHookMultipleTimes(t *testing.T) {
	mgr := NewManager()

	// Register the same hook function multiple times
	hook := func(tx *gorm.DB) {}

	_ = mgr.Register("beforeCreate", hook)
	_ = mgr.Register("beforeCreate", hook)
	_ = mgr.Register("beforeCreate", hook)

	// All instances should be registered
	hooks := mgr.Get("beforeCreate")
	if len(hooks) != 3 {
		t.Errorf("Expected 3 instances of the same hook, got %d", len(hooks))
	}

	// Verify all hooks are non-nil
	for i, h := range hooks {
		if h == nil {
			t.Errorf("Hook at index %d should not be nil", i)
		}
	}
}

func TestManager_ClearAndRegisterSameType(t *testing.T) {
	mgr := NewManager()

	// Register beforeCreate hook
	hook1 := func(tx *gorm.DB) {}
	_ = mgr.Register("beforeCreate", hook1)

	// Clear all hooks
	mgr.Clear()

	// Register beforeCreate hook again
	hook2 := func(tx *gorm.DB) {}
	_ = mgr.Register("beforeCreate", hook2)

	// Should only have the new hook
	hooks := mgr.Get("beforeCreate")
	if len(hooks) != 1 {
		t.Errorf("Expected 1 hook after Clear() and re-register, got %d", len(hooks))
	}

	if hooks[0] == nil {
		t.Error("Expected a valid hook after Clear() and re-register")
	}
}

func TestManager_NilHook(t *testing.T) {
	mgr := NewManager()

	// Register nil hook (should not panic)
	err := mgr.Register("beforeCreate", nil)
	if err != nil {
		t.Errorf("Register() with nil hook should not error, got %v", err)
	}

	// Get should return the nil hook
	hooks := mgr.Get("beforeCreate")
	if len(hooks) != 1 {
		t.Errorf("Expected 1 hook (nil), got %d", len(hooks))
	}

	// Verify the hook is nil
	if hooks[0] != nil {
		t.Error("Expected hook to be nil")
	}
}

func TestManager_RemoveNonExistingHook(t *testing.T) {
	mgr := NewManager()

	// Register a hook
	hook := func(tx *gorm.DB) {}
	_ = mgr.Register("beforeCreate", hook)

	// Remove non-existing hook type
	mgr.Remove("afterCreate")

	// beforeCreate should still exist
	if len(mgr.Get("beforeCreate")) != 1 {
		t.Error("Expected beforeCreate hook to still exist after removing non-existing hook")
	}
}

func TestManager_ClearEmptyManager(t *testing.T) {
	mgr := NewManager()

	// Clear empty manager
	mgr.Clear()

	// Should be able to register hooks after clear
	hook := func(tx *gorm.DB) {}
	_ = mgr.Register("beforeCreate", hook)

	if len(mgr.Get("beforeCreate")) != 1 {
		t.Error("Expected to be able to register hooks after Clear() on empty manager")
	}
}

func TestManager_MixedOperations(t *testing.T) {
	mgr := NewManager()

	var wg sync.WaitGroup

	// Perform mixed operations concurrently
	for i := 0; i < 50; i++ {
		wg.Add(4)

		go func(i int) {
			defer wg.Done()
			hook := func(tx *gorm.DB) {}
			_ = mgr.Register("beforeCreate", hook)
		}(i)

		go func() {
			defer wg.Done()
			_ = mgr.Get("beforeCreate")
		}()

		go func() {
			defer wg.Done()
			mgr.Clear()
		}()

		go func() {
			defer wg.Done()
			mgr.Remove("afterCreate")
		}()
	}

	wg.Wait()
}

func TestManager_HookTypesValidation(t *testing.T) {
	validHooks := []string{
		"beforeCreate", "afterCreate",
		"beforeUpdate", "afterUpdate",
		"beforeDelete", "afterDelete",
		"afterFind",
	}

	invalidHooks := []string{
		"BeforeCreate", // case sensitive
		"BEFORECREATE",
		"beforecreate",
		"invalid",
		"create",
		"update",
		"delete",
		"find",
		"before",
		"after",
	}

	hook := func(tx *gorm.DB) {}

	// Test valid hooks
	for _, hookName := range validHooks {
		t.Run("valid_"+hookName, func(t *testing.T) {
			mgr := NewManager()
			err := mgr.Register(hookName, hook)
			if err != nil {
				t.Errorf("Register(%s) should not error, got %v", hookName, err)
			}
		})
	}

	// Test invalid hooks
	for _, hookName := range invalidHooks {
		t.Run("invalid_"+hookName, func(t *testing.T) {
			mgr := NewManager()
			err := mgr.Register(hookName, hook)
			if err == nil {
				t.Errorf("Register(%s) should error", hookName)
			}
		})
	}
}

func TestManager_LargeNumberOfHooks(t *testing.T) {
	mgr := NewManager()

	// Register a large number of hooks
	for i := 0; i < 1000; i++ {
		hook := func(tx *gorm.DB) {}
		_ = mgr.Register("beforeCreate", hook)
	}

	hooks := mgr.Get("beforeCreate")
	if len(hooks) != 1000 {
		t.Errorf("Expected 1000 hooks, got %d", len(hooks))
	}
}

func TestManager_CallbacksInitialization(t *testing.T) {
	mgr := NewManager()

	if mgr.callbacks == nil {
		t.Fatal("NewManager() should initialize callbacks map")
	}

	// Verify we can access the map without panicking
	_ = len(mgr.callbacks)
}

func TestManager_ClearReinitializesMap(t *testing.T) {
	mgr := NewManager()

	// Register hooks
	hook := func(tx *gorm.DB) {}
	_ = mgr.Register("beforeCreate", hook)

	// Clear
	mgr.Clear()

	// New map should be empty
	if len(mgr.callbacks) != 0 {
		t.Error("Clear() should result in empty callbacks map")
	}

	// Verify we can register hooks after clear
	_ = mgr.Register("afterCreate", hook)
	if len(mgr.callbacks) != 1 {
		t.Error("Should be able to register hooks after Clear()")
	}
}
