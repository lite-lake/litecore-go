package container

import (
	"errors"
	"testing"
)

// TestConfigContainer 测试 ConfigContainer
func TestConfigContainer(t *testing.T) {
	container := NewConfigContainer()

	// 测试注册
	config := &MockConfigProvider{name: "test-config"}
	err := container.Register(config)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 测试重复注册
	err = container.Register(config)
	if err == nil {
		t.Fatal("Expected duplicate registration error")
	}
	var dupErr *DuplicateRegistrationError
	if !errors.As(err, &dupErr) {
		t.Fatalf("Expected DuplicateRegistrationError, got %T", err)
	}

	// 测试获取
	retrieved, err := container.GetByName("test-config")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if retrieved != config {
		t.Fatal("Retrieved config is not the same as registered")
	}

	// 测试获取不存在的实例
	_, err = container.GetByName("non-existent")
	if err == nil {
		t.Fatal("Expected error when getting non-existent config")
	}
	var notFoundErr *InstanceNotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("Expected InstanceNotFoundError, got %T", err)
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
