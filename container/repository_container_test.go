package container

import (
	"reflect"
	"testing"

	"github.com/lite-lake/litecore-go/common"
)

// TestRepositoryContainer 测试 RepositoryContainer
func TestRepositoryContainer(t *testing.T) {
	entityContainer := NewEntityContainer()
	repositoryContainer := NewRepositoryContainer(entityContainer)

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

	// 设置 BuiltinProvider
	config := &MockConfigProvider{name: "app-configmgr"}
	manager := &MockManager{name: "db-manager"}
	builtinProvider := &MockBuiltinProvider{
		configProvider: config,
		managers:       []interface{}{manager},
	}
	repositoryContainer.SetBuiltinProvider(builtinProvider)

	// 注入依赖
	err = repositoryContainer.InjectAll()
	if err != nil {
		t.Fatalf("InjectAll failed: %v", err)
	}

	// 验证实体已注入
	if repo.Entity != entity {
		t.Fatal("Entity was not injected into repository")
	}
}
