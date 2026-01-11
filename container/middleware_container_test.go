package container

import "testing"

// TestMiddlewareContainer 测试 MiddlewareContainer
func TestMiddlewareContainer(t *testing.T) {
	configContainer := NewConfigContainer()
	managerContainer := NewManagerContainer(configContainer)
	serviceContainer := NewServiceContainer(configContainer, managerContainer, NewRepositoryContainer(configContainer, managerContainer, NewEntityContainer()))
	middlewareContainer := NewMiddlewareContainer(configContainer, managerContainer, serviceContainer)

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

	// 注册中间件
	middleware := &MockMiddleware{name: "auth-middleware"}
	err := middlewareContainer.Register(middleware)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 注入依赖
	err = middlewareContainer.InjectAll()
	if err != nil {
		t.Fatalf("InjectAll failed: %v", err)
	}

	// 验证服务已注入
	if middleware.Service != service {
		t.Fatal("Service was not injected into middleware")
	}
}
