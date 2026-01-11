package container

import "testing"

// TestManagerContainer 测试 ManagerContainer（含依赖注入）
func TestManagerContainer(t *testing.T) {
	configContainer := NewConfigContainer()
	managerContainer := NewManagerContainer(configContainer)

	// 注册配置
	config := &MockConfigProvider{name: "app-config"}
	configContainer.Register(config)

	// 注册管理器（依赖配置）
	manager := &MockManager{name: "db-manager"}
	err := managerContainer.Register(manager)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 注入依赖
	err = managerContainer.InjectAll()
	if err != nil {
		t.Fatalf("InjectAll failed: %v", err)
	}

	// 验证依赖已注入
	if manager.Config != config {
		t.Fatalf("Config was not injected into manager. Expected: %v, Got: %v", config, manager.Config)
	}

	// 测试重复 InjectAll（应该幂等）
	err = managerContainer.InjectAll()
	if err != nil {
		t.Fatalf("Second InjectAll failed: %v", err)
	}
}
