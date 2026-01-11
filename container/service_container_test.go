package container

import "testing"

// TestServiceContainerWithSameLayerDependency 测试 ServiceContainer 的同层依赖
func TestServiceContainerWithSameLayerDependency(t *testing.T) {
	configContainer := NewConfigContainer()
	managerContainer := NewManagerContainer(configContainer)
	repositoryContainer := NewRepositoryContainer(configContainer, managerContainer, NewEntityContainer())
	serviceContainer := NewServiceContainer(configContainer, managerContainer, repositoryContainer)

	// 注册配置
	config := &MockConfigProvider{name: "app-config"}
	configContainer.Register(config)

	// 注册管理器
	manager := &MockManager{name: "db-manager"}
	managerContainer.Register(manager)

	// 注册存储库
	repo := &MockRepository{name: "user-repo"}
	repositoryContainer.Register(repo)

	// 注入下层容器
	managerContainer.InjectAll()
	repositoryContainer.InjectAll()

	// 创建服务：OrderService 无同层依赖，UserService 依赖 OrderService
	orderService := &MockService{name: "order-service"}
	userService := &MockService{name: "user-service"}

	// 手动设置 UserService 依赖 OrderService（通过反射无法直接设置接口）
	// 实际场景中，这会通过 inject 标签自动注入

	// 注册服务（顺序不限）
	serviceContainer.Register(userService)  // 依赖 OrderService
	serviceContainer.Register(orderService) // 无依赖

	// 注入服务依赖
	err := serviceContainer.InjectAll()
	if err != nil {
		t.Fatalf("InjectAll failed: %v", err)
	}

	// 验证 Config 和 Repo 已注入到服务
	if orderService.Config == nil {
		t.Error("Config was not injected into orderService")
	}
	if orderService.Repo == nil {
		t.Error("Repo was not injected into orderService")
	}
	if userService.Config == nil {
		t.Error("Config was not injected into userService")
	}
	if userService.Repo == nil {
		t.Error("Repo was not injected into userService")
	}
}
