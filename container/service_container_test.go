package container

import (
	"reflect"
	"testing"

	"github.com/lite-lake/litecore-go/common"
)

// TestServiceContainer 测试 ServiceContainer
func TestServiceContainer(t *testing.T) {
	entityContainer := NewEntityContainer()
	repositoryContainer := NewRepositoryContainer(entityContainer)
	serviceContainer := NewServiceContainer(repositoryContainer)

	// 注册实体
	entity := &MockEntity{name: "user-entity", id: "1"}
	err := entityContainer.Register(entity)
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

	// 设置 BuiltinProvider
	config := &MockConfigProvider{name: "app-config"}
	manager := &MockManager{name: "db-manager"}
	builtinProvider := &MockBuiltinProvider{
		configProvider: config,
		managers:       []interface{}{manager},
	}

	repositoryContainer.SetBuiltinProvider(builtinProvider)
	serviceContainer.SetBuiltinProvider(builtinProvider)

	// 注入依赖
	err = repositoryContainer.InjectAll()
	if err != nil {
		t.Fatalf("Repository InjectAll failed: %v", err)
	}

	err = serviceContainer.InjectAll()
	if err != nil {
		t.Fatalf("Service InjectAll failed: %v", err)
	}

	// 验证存储库已注入
	if service.Repo != repo {
		t.Fatal("Repository was not injected into service")
	}
}
