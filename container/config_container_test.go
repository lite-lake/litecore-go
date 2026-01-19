package container

import (
	"reflect"
	"testing"

	"com.litelake.litecore/common"
)

// TestConfigContainer 测试 ConfigContainer
func TestConfigContainer(t *testing.T) {
	container := NewConfigContainer()

	// 测试注册
	config := &MockConfigProvider{name: "test-config"}
	baseConfigType := reflect.TypeOf((*common.IBaseConfigProvider)(nil)).Elem()
	err := container.RegisterByType(baseConfigType, config)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 测试重复注册
	err = container.RegisterByType(baseConfigType, config)
	if err == nil {
		t.Fatal("Expected duplicate registration error")
	}

	// 测试获取
	retrieved := container.GetByType(baseConfigType)
	if retrieved != config {
		t.Fatal("Retrieved config is not the same as registered")
	}

	// 测试获取不存在的接口
	nonExistentType := reflect.TypeOf((*testInterface)(nil)).Elem()
	retrieved = container.GetByType(nonExistentType)
	if retrieved != nil {
		t.Fatal("Expected nil for non-existent interface")
	}

	// 测试 GetAll
	all := container.GetAll()
	if len(all) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(all))
	}

	// 测试 Count
	if container.Count() != 1 {
		t.Fatal("Count should be 1")
	}
}

// testInterface 测试接口
type testInterface interface {
	TestMethod()
}
