package container

import "testing"

// TestEntityContainer 测试 EntityContainer
func TestEntityContainer(t *testing.T) {
	container := NewEntityContainer()

	// 测试注册
	entity := &MockEntity{name: "test-entity", id: "1"}
	err := container.Register(entity)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 测试获取
	retrieved, err := container.GetByName("test-entity")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if retrieved != entity {
		t.Fatal("Retrieved entity is not the same as registered")
	}
}
