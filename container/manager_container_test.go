package container

import (
	"reflect"
	"testing"

	"com.litelake.litecore/common"
)

// TestManagerContainer 测试 ManagerContainer（含依赖注入）
func TestManagerContainer(t *testing.T) {
	configContainer := NewConfigContainer()
	managerContainer := NewManagerContainer(configContainer)

	// 注册配置
	config := &MockConfigProvider{name: "app-config"}
	err := configContainer.RegisterByType(reflect.TypeOf((*common.BaseConfigProvider)(nil)).Elem(), config)
	if err != nil {
		t.Fatalf("Register config failed: %v", err)
	}

	// 注册管理器（依赖配置）
	manager := &MockManager{name: "db-manager"}
	err = managerContainer.RegisterByType(reflect.TypeOf((*common.BaseManager)(nil)).Elem(), manager)
	if err != nil {
		t.Fatalf("Register manager failed: %v", err)
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
