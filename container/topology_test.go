package container

import (
	"testing"
)

// TestTopologicalSort 测试拓扑排序
func TestTopologicalSort(t *testing.T) {
	t.Run("空图_返回空列表", func(t *testing.T) {
		graph := make(map[string][]string)
		order, err := topologicalSort(graph)
		if err != nil {
			t.Fatalf("期望成功, 实际错误: %v", err)
		}
		if len(order) != 0 {
			t.Errorf("期望空列表, 实际长度: %d", len(order))
		}
	})

	t.Run("单个节点_返回单个节点", func(t *testing.T) {
		graph := map[string][]string{
			"A": {},
		}
		order, err := topologicalSort(graph)
		if err != nil {
			t.Fatalf("期望成功, 实际错误: %v", err)
		}
		if len(order) != 1 || order[0] != "A" {
			t.Errorf("期望 ['A'], 实际: %v", order)
		}
	})

	t.Run("简单链_正确排序", func(t *testing.T) {
		graph := map[string][]string{
			"A": {"B"},
			"B": {"C"},
			"C": {},
		}
		order, err := topologicalSort(graph)
		if err != nil {
			t.Fatalf("期望成功, 实际错误: %v", err)
		}
		expectedOrder := []string{"C", "B", "A"}
		for i, v := range order {
			if v != expectedOrder[i] {
				t.Errorf("期望 %v, 实际: %v", expectedOrder, order)
				break
			}
		}
	})
}
