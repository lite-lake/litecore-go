package container

import (
	"fmt"
)

// InstanceIterator 实例迭代器接口
// 用于泛型场景下的实例迭代
type InstanceIterator interface {
	// Range 遍历所有实例
	Range(func(name string, instance interface{}) bool)
}

// topologicalSort 使用 Kahn 算法进行拓扑排序
// graph: 依赖图，key 为节点名，value 为该节点依赖的节点列表
// 返回: 拓扑排序后的节点列表
func topologicalSort(graph map[string][]string) ([]string, error) {
	// 初始化入度和邻接表
	inDegree := make(map[string]int)
	adjList := make(map[string][]string)

	// 为所有节点初始化入度为 0
	for node := range graph {
		inDegree[node] = 0
	}

	// 构建邻接表和入度
	// graph[A] = [B, C] 表示 A 依赖 B 和 C
	// 即 B -> A, C -> A
	for node, deps := range graph {
		for _, dep := range deps {
			adjList[dep] = append(adjList[dep], node)
			inDegree[node]++
		}
	}

	// 找到所有入度为 0 的节点（无依赖）
	var queue []string
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	var result []string
	for len(queue) > 0 {
		// 取出一个节点
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		// 减少邻接节点的入度
		for _, neighbor := range adjList[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// 检测循环依赖：如果结果数量不等于图节点数，说明存在循环
	if len(result) != len(graph) {
		// 找出剩余的节点（参与循环的节点）
		var remainingNodes []string
		for node, degree := range inDegree {
			if degree > 0 {
				remainingNodes = append(remainingNodes, node)
			}
		}
		return nil, &CircularDependencyError{
			Cycle: remainingNodes,
		}
	}

	return result, nil
}

// buildDependencyGraphFromMap 从 map 构建依赖图
func buildDependencyGraphFromMap(
	instances map[string]interface{},
	getDependencies func(interface{}) ([]string, error),
) (map[string][]string, error) {
	graph := make(map[string][]string)

	for name, instance := range instances {
		deps, err := getDependencies(instance)
		if err != nil {
			return nil, fmt.Errorf("build graph for %s failed: %w", name, err)
		}
		graph[name] = deps
	}

	return graph, nil
}

// buildDependencyGraphFromIterator 从迭代器构建依赖图
func buildDependencyGraphFromIterator(
	iterator InstanceIterator,
	getDependencies func(interface{}) ([]string, error),
) (map[string][]string, error) {
	graph := make(map[string][]string)

	iterator.Range(func(name string, instance interface{}) bool {
		deps, err := getDependencies(instance)
		if err != nil {
			// 将错误保存并停止迭代
			return false
		}
		graph[name] = deps
		return true
	})

	return graph, nil
}
