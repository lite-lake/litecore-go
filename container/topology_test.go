package container

import (
	"errors"
	"reflect"
	"testing"

	"com.litelake.litecore/common"
)

// TestCircularDependencyDetection 测试循环依赖检测
func TestCircularDependencyDetection(t *testing.T) {
	configContainer := NewConfigContainer()
	managerContainer := NewManagerContainer(configContainer)
	_ = NewServiceContainer(configContainer, managerContainer,
		NewRepositoryContainer(configContainer, managerContainer, NewEntityContainer()))

	// 注册配置
	config := &MockConfigProvider{name: "app-config"}
	err := configContainer.RegisterByType(reflect.TypeOf((*common.IBaseConfigProvider)(nil)).Elem(), config)
	if err != nil {
		t.Fatalf("Register config failed: %v", err)
	}

	// 注入下层容器
	err = managerContainer.InjectAll()
	if err != nil {
		t.Fatalf("Manager InjectAll failed: %v", err)
	}

	// 这里需要创建真正有循环依赖的服务
	// 由于 MockService 无法直接表达同层依赖关系，
	// 这个测试需要更复杂的 Mock 实现
	// 当前跳过此测试
	t.Skip("Need more complex mock to test circular dependency")
}

// TestTopologicalSort 测试拓扑排序
func TestTopologicalSort(t *testing.T) {
	tests := []struct {
		name      string
		graph     map[string][]string
		wantErr   bool
		wantOrder []string
	}{
		{
			name: "simple linear",
			graph: map[string][]string{
				"A": {"B"},
				"B": {"C"},
				"C": {},
			},
			wantErr:   false,
			wantOrder: []string{"C", "B", "A"}, // 或其他有效顺序
		},
		{
			name: "diamond",
			graph: map[string][]string{
				"A": {"B", "C"},
				"B": {"D"},
				"C": {"D"},
				"D": {},
			},
			wantErr:   false,
			wantOrder: []string{"D", "B", "C", "A"}, // D 必须在最后，A 必须在最前
		},
		{
			name: "circular",
			graph: map[string][]string{
				"A": {"B"},
				"B": {"C"},
				"C": {"A"},
			},
			wantErr: true,
		},
		{
			name:      "empty",
			graph:     map[string][]string{},
			wantErr:   false,
			wantOrder: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := topologicalSort(tt.graph)

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				var circErr *CircularDependencyError
				if !errors.As(err, &circErr) {
					t.Fatalf("Expected CircularDependencyError, got %T", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(order) != len(tt.graph) {
				t.Fatalf("Expected %d nodes, got %d", len(tt.graph), len(order))
			}

			// 验证顺序满足依赖关系
			pos := make(map[string]int)
			for i, node := range order {
				pos[node] = i
			}
			for node, deps := range tt.graph {
				for _, dep := range deps {
					if pos[node] <= pos[dep] {
						t.Errorf("Invalid order: %s should come after %s", node, dep)
					}
				}
			}
		})
	}
}

// TestNilRegistration 测试注册 nil
func TestNilRegistration(t *testing.T) {
	configContainer := NewConfigContainer()

	err := configContainer.RegisterByType(reflect.TypeOf((*common.IBaseConfigProvider)(nil)).Elem(), nil)
	if err == nil {
		t.Fatal("Expected error when registering nil")
	}
}
