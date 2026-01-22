package container

import (
	"container/list"
	"reflect"
)

// topologicalSortByInterfaceType 使用 Kahn 算法进行拓扑排序（接口类型版本）
// graph: 依赖图，key 和 value 都是接口类型
// 返回: 拓扑排序后的接口类型列表
func topologicalSortByInterfaceType(graph map[reflect.Type][]reflect.Type) ([]reflect.Type, error) {
	inDegree := make(map[reflect.Type]int)
	adjList := make(map[reflect.Type][]reflect.Type)

	for node := range graph {
		inDegree[node] = 0
	}

	for node, deps := range graph {
		for _, dep := range deps {
			adjList[dep] = append(adjList[dep], node)
			inDegree[node]++
		}
	}

	queue := list.New()
	for node, degree := range inDegree {
		if degree == 0 {
			queue.PushBack(node)
		}
	}

	var result []reflect.Type
	for queue.Len() > 0 {
		node := queue.Remove(queue.Front()).(reflect.Type)
		result = append(result, node)

		for _, neighbor := range adjList[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue.PushBack(neighbor)
			}
		}
	}

	if len(result) != len(graph) {
		var remainingNodes []string
		for node, degree := range inDegree {
			if degree > 0 {
				remainingNodes = append(remainingNodes, node.String())
			}
		}
		return nil, &CircularDependencyError{
			Cycle: remainingNodes,
		}
	}

	return result, nil
}
