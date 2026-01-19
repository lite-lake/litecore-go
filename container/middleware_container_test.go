package container

import (
	"reflect"
	"testing"

	"com.litelake.litecore/common"
)

// TestMiddlewareContainer 测试 MiddlewareContainer
func TestMiddlewareContainer(t *testing.T) {
	configContainer := NewConfigContainer()
	managerContainer := NewManagerContainer(configContainer)
	entityContainer := NewEntityContainer()
	repositoryContainer := NewRepositoryContainer(configContainer, managerContainer, entityContainer)
	serviceContainer := NewServiceContainer(configContainer, managerContainer, repositoryContainer)
	middlewareContainer := NewMiddlewareContainer(configContainer, managerContainer, serviceContainer)

	// 注册配置
	config := &MockConfigProvider{name: "app-config"}
	err := configContainer.RegisterByType(reflect.TypeOf((*common.IBaseConfigProvider)(nil)).Elem(), config)
	if err != nil {
		t.Fatalf("Register config failed: %v", err)
	}

	// 注册管理器
	manager := &MockManager{name: "db-manager"}
	err = managerContainer.RegisterByType(reflect.TypeOf((*common.IBaseManager)(nil)).Elem(), manager)
	if err != nil {
		t.Fatalf("Register manager failed: %v", err)
	}

	// 注册实体
	entity := &MockEntity{name: "user-entity", id: "1"}
	err = entityContainer.Register(entity)
	if err != nil {
		t.Fatalf("Register entity failed: %v", err)
	}

	// 注册存储库
	repo := &MockRepository{name: "user-repo"}
	err = repositoryContainer.RegisterByType(reflect.TypeOf((*common.IBaseRepository)(nil)).Elem(), repo)
	if err != nil {
		t.Fatalf("Register repository failed: %v", err)
	}

	// 注册服务
	service := &MockService{name: "user-service"}
	err = serviceContainer.RegisterByType(reflect.TypeOf((*common.IBaseService)(nil)).Elem(), service)
	if err != nil {
		t.Fatalf("Register service failed: %v", err)
	}

	// 注入下层容器
	err = managerContainer.InjectAll()
	if err != nil {
		t.Fatalf("Manager InjectAll failed: %v", err)
	}

	err = repositoryContainer.InjectAll()
	if err != nil {
		t.Fatalf("Repository InjectAll failed: %v", err)
	}

	err = serviceContainer.InjectAll()
	if err != nil {
		t.Fatalf("Service InjectAll failed: %v", err)
	}

	// 注册中间件
	middleware := &MockMiddleware{name: "auth-middleware"}
	err = middlewareContainer.RegisterByType(reflect.TypeOf((*common.IBaseMiddleware)(nil)).Elem(), middleware)
	if err != nil {
		t.Fatalf("Register middleware failed: %v", err)
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
