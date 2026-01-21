package container

import (
	"reflect"
	"testing"

	"github.com/lite-lake/litecore-go/common"
)

// TestRepositoryContainer 测试 RepositoryContainer（含依赖注入）
func TestRepositoryContainer(t *testing.T) {
	configContainer := NewConfigContainer()
	managerContainer := NewManagerContainer(configContainer)
	entityContainer := NewEntityContainer()
	repositoryContainer := NewRepositoryContainer(configContainer, managerContainer, entityContainer)

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

	// 注册存储库（依赖配置、管理器、实体）
	repo := &MockRepository{name: "user-repo"}
	err = repositoryContainer.RegisterByType(reflect.TypeOf((*common.IBaseRepository)(nil)).Elem(), repo)
	if err != nil {
		t.Fatalf("Register repository failed: %v", err)
	}

	// 需要先注入 Manager（因为 Repository 依赖 Manager）
	err = managerContainer.InjectAll()
	if err != nil {
		t.Fatalf("Manager InjectAll failed: %v", err)
	}

	// 注入存储库依赖
	err = repositoryContainer.InjectAll()
	if err != nil {
		t.Fatalf("InjectAll failed: %v", err)
	}

	// 验证依赖已注入
	if repo.Config != config {
		t.Fatal("Config was not injected into repository")
	}
	if repo.Manager != manager {
		t.Fatal("Manager was not injected into repository")
	}
	if repo.Entity != entity {
		t.Fatal("Entity was not injected into repository")
	}
}
