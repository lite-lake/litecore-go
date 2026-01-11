package container

import "testing"

// TestControllerContainer 测试 ControllerContainer
func TestControllerContainer(t *testing.T) {
	configContainer := NewConfigContainer()
	managerContainer := NewManagerContainer(configContainer)
	serviceContainer := NewServiceContainer(configContainer, managerContainer, NewRepositoryContainer(configContainer, managerContainer, NewEntityContainer()))
	controllerContainer := NewControllerContainer(configContainer, managerContainer, serviceContainer)

	// 注册配置
	config := &MockConfigProvider{name: "app-config"}
	configContainer.Register(config)

	// 注册管理器
	manager := &MockManager{name: "db-manager"}
	managerContainer.Register(manager)

	// 注册服务
	service := &MockService{name: "user-service"}
	serviceContainer.Register(service)

	// 注入下层容器
	managerContainer.InjectAll()
	serviceContainer.InjectAll()

	// 注册控制器
	controller := &MockController{name: "user-controller"}
	err := controllerContainer.Register(controller)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 注入依赖
	err = controllerContainer.InjectAll()
	if err != nil {
		t.Fatalf("InjectAll failed: %v", err)
	}

	// 验证服务已注入
	if controller.Service != service {
		t.Fatal("Service was not injected into controller")
	}
}
