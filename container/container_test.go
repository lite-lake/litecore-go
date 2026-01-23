package container

import (
	"reflect"
	"testing"
)

// 模拟测试接口
type testInterface interface {
	testMethod() string
}

type testImplementation struct {
	name string
}

func (t *testImplementation) testMethod() string {
	return t.name
}

// 测试 TypedContainer
func TestTypedContainer(t *testing.T) {
	t.Run("基本注册和获取", func(t *testing.T) {
		container := NewTypedContainer(func(item testInterface) string {
			return item.testMethod()
		})

		ifaceType := reflect.TypeOf((*testInterface)(nil)).Elem()
		impl := &testImplementation{name: "test"}

		err := container.Register(ifaceType, impl)
		if err != nil {
			t.Errorf("注册失败: %v", err)
		}

		retrieved := container.GetByType(ifaceType)
		if retrieved == nil {
			t.Fatal("获取失败，返回 nil")
		}

		if retrieved.testMethod() != "test" {
			t.Errorf("获取的实例不正确，期望: test, 实际: %s", retrieved.testMethod())
		}
	})

	t.Run("注册 nil 实现", func(t *testing.T) {
		container := NewTypedContainer(func(item testInterface) string { return "" })
		ifaceType := reflect.TypeOf((*testInterface)(nil)).Elem()

		err := container.Register(ifaceType, nil)
		if err == nil {
			t.Error("期望返回错误，但没有")
		}
	})

	t.Run("注册不实现接口的实现", func(t *testing.T) {
		container := NewTypedContainer(func(item testInterface) string { return "" })

		type otherInterface interface{}
		otherType := reflect.TypeOf((*otherInterface)(nil)).Elem()

		impl := &testImplementation{name: "test"}
		err := container.Register(otherType, impl)
		if err == nil {
			t.Error("期望返回错误，但没有")
		}
	})

	t.Run("重复注册", func(t *testing.T) {
		container := NewTypedContainer(func(item testInterface) string { return "" })
		ifaceType := reflect.TypeOf((*testInterface)(nil)).Elem()

		impl1 := &testImplementation{name: "test1"}
		impl2 := &testImplementation{name: "test2"}

		err1 := container.Register(ifaceType, impl1)
		if err1 != nil {
			t.Errorf("第一次注册失败: %v", err1)
		}

		err2 := container.Register(ifaceType, impl2)
		if err2 == nil {
			t.Error("期望返回重复注册错误，但没有")
		}
	})

	t.Run("获取所有实例", func(t *testing.T) {
		container := NewTypedContainer(func(item testInterface) string { return "" })
		ifaceType := reflect.TypeOf((*testInterface)(nil)).Elem()

		impl := &testImplementation{name: "test"}
		container.Register(ifaceType, impl)

		all := container.GetAll()
		if len(all) != 1 {
			t.Errorf("期望 1 个实例，实际: %d", len(all))
		}
	})

	t.Run("获取实例数量", func(t *testing.T) {
		container := NewTypedContainer(func(item testInterface) string { return "" })
		ifaceType := reflect.TypeOf((*testInterface)(nil)).Elem()

		if container.Count() != 0 {
			t.Errorf("期望 0 个实例，实际: %d", container.Count())
		}

		container.Register(ifaceType, &testImplementation{name: "test"})

		if container.Count() != 1 {
			t.Errorf("期望 1 个实例，实际: %d", container.Count())
		}
	})

	t.Run("注入状态", func(t *testing.T) {
		container := NewTypedContainer(func(item testInterface) string { return "" })

		if container.IsInjected() {
			t.Error("期望未注入状态")
		}

		container.setInjected(true)

		if !container.IsInjected() {
			t.Error("期望已注入状态")
		}
	})

	t.Run("获取名称", func(t *testing.T) {
		container := NewTypedContainer(func(item testInterface) string {
			return item.testMethod()
		})
		ifaceType := reflect.TypeOf((*testInterface)(nil)).Elem()

		container.Register(ifaceType, &testImplementation{name: "test"})

		names := container.GetNames()
		if len(names) != 1 {
			t.Errorf("期望 1 个名称，实际: %d", len(names))
		}

		if names[0] != "test" {
			t.Errorf("期望名称 test，实际: %s", names[0])
		}
	})

	t.Run("遍历所有实例", func(t *testing.T) {
		container := NewTypedContainer(func(item testInterface) string { return "" })
		ifaceType := reflect.TypeOf((*testInterface)(nil)).Elem()

		container.Register(ifaceType, &testImplementation{name: "test"})

		count := 0
		container.RangeItems(func(ifaceType reflect.Type, item testInterface) bool {
			count++
			return true
		})

		if count != 1 {
			t.Errorf("期望遍历 1 个实例，实际: %d", count)
		}
	})
}

// 测试 NamedContainer
func TestNamedContainer(t *testing.T) {
	t.Run("基本注册和获取", func(t *testing.T) {
		container := NewNamedContainer(func(item testInterface) string {
			return item.testMethod()
		})

		impl := &testImplementation{name: "test"}

		err := container.Register(impl)
		if err != nil {
			t.Errorf("注册失败: %v", err)
		}

		retrieved, err := container.GetByName("test")
		if err != nil {
			t.Errorf("获取失败: %v", err)
		}

		if retrieved.testMethod() != "test" {
			t.Errorf("获取的实例不正确，期望: test, 实际: %s", retrieved.testMethod())
		}
	})

	t.Run("注册空名称", func(t *testing.T) {
		container := NewNamedContainer(func(item testInterface) string { return "" })
		impl := &testImplementation{name: ""}

		err := container.Register(impl)
		if err == nil {
			t.Error("期望返回错误，但没有")
		}
	})

	t.Run("重复注册", func(t *testing.T) {
		container := NewNamedContainer(func(item testInterface) string { return "" })

		impl1 := &testImplementation{name: "test"}
		impl2 := &testImplementation{name: "test"}

		err1 := container.Register(impl1)
		if err1 != nil {
			t.Errorf("第一次注册失败: %v", err1)
		}

		err2 := container.Register(impl2)
		if err2 == nil {
			t.Error("期望返回重复注册错误，但没有")
		}
	})

	t.Run("获取不存在的实例", func(t *testing.T) {
		container := NewNamedContainer(func(item testInterface) string { return "" })

		_, err := container.GetByName("nonexistent")
		if err == nil {
			t.Error("期望返回错误，但没有")
		}
	})

	t.Run("获取所有实例", func(t *testing.T) {
		container := NewNamedContainer(func(item testInterface) string { return "" })

		container.Register(&testImplementation{name: "test1"})
		container.Register(&testImplementation{name: "test2"})

		all := container.GetAll()
		if len(all) != 2 {
			t.Errorf("期望 2 个实例，实际: %d", len(all))
		}
	})

	t.Run("获取实例数量", func(t *testing.T) {
		container := NewNamedContainer(func(item testInterface) string { return "" })

		if container.Count() != 0 {
			t.Errorf("期望 0 个实例，实际: %d", container.Count())
		}

		container.Register(&testImplementation{name: "test"})

		if container.Count() != 1 {
			t.Errorf("期望 1 个实例，实际: %d", container.Count())
		}
	})

	t.Run("获取名称", func(t *testing.T) {
		container := NewNamedContainer(func(item testInterface) string {
			return item.testMethod()
		})

		container.Register(&testImplementation{name: "test1"})
		container.Register(&testImplementation{name: "test2"})

		names := container.GetNames()
		if len(names) != 2 {
			t.Errorf("期望 2 个名称，实际: %d", len(names))
		}
	})
}

// 测试拓扑排序
func TestTopologicalSortByInterfaceType(t *testing.T) {
	t.Run("简单依赖图", func(t *testing.T) {
		type interfaceA interface{}
		type interfaceB interface{}

		aType := reflect.TypeOf((*interfaceA)(nil)).Elem()
		bType := reflect.TypeOf((*interfaceB)(nil)).Elem()

		graph := map[reflect.Type][]reflect.Type{
			bType: {aType}, // B 依赖 A
			aType: {},
		}

		result, err := topologicalSortByInterfaceType(graph)
		if err != nil {
			t.Errorf("拓扑排序失败: %v", err)
		}

		if len(result) != 2 {
			t.Errorf("期望 2 个结果，实际: %d", len(result))
		}

		if result[0] != aType {
			t.Error("期望 A 排在第一个")
		}
	})

	t.Run("复杂依赖图", func(t *testing.T) {
		type interfaceA interface{}
		type interfaceB interface{}
		type interfaceC interface{}

		aType := reflect.TypeOf((*interfaceA)(nil)).Elem()
		bType := reflect.TypeOf((*interfaceB)(nil)).Elem()
		cType := reflect.TypeOf((*interfaceC)(nil)).Elem()

		graph := map[reflect.Type][]reflect.Type{
			cType: {aType, bType}, // C 依赖 A 和 B
			aType: {},
			bType: {aType}, // B 依赖 A
		}

		result, err := topologicalSortByInterfaceType(graph)
		if err != nil {
			t.Errorf("拓扑排序失败: %v", err)
		}

		if len(result) != 3 {
			t.Errorf("期望 3 个结果，实际: %d", len(result))
		}

		if result[0] != aType {
			t.Error("期望 A 排在第一个")
		}
	})

	t.Run("循环依赖", func(t *testing.T) {
		type interfaceA interface{}
		type interfaceB interface{}

		aType := reflect.TypeOf((*interfaceA)(nil)).Elem()
		bType := reflect.TypeOf((*interfaceB)(nil)).Elem()

		graph := map[reflect.Type][]reflect.Type{
			aType: {bType}, // A 依赖 B
			bType: {aType}, // B 依赖 A
		}

		_, err := topologicalSortByInterfaceType(graph)
		if err == nil {
			t.Error("期望返回循环依赖错误，但没有")
		}

		_, isCircular := err.(*CircularDependencyError)
		if !isCircular {
			t.Error("期望返回 CircularDependencyError")
		}
	})

	t.Run("空图", func(t *testing.T) {
		graph := map[reflect.Type][]reflect.Type{}

		result, err := topologicalSortByInterfaceType(graph)
		if err != nil {
			t.Errorf("拓扑排序失败: %v", err)
		}

		if len(result) != 0 {
			t.Errorf("期望 0 个结果，实际: %d", len(result))
		}
	})
}

// 测试错误类型
func TestErrors(t *testing.T) {
	t.Run("DependencyNotFoundError", func(t *testing.T) {
		err := &DependencyNotFoundError{
			InstanceName:  "TestService",
			FieldName:     "repo",
			FieldType:     reflect.TypeOf(struct{}{}),
			ContainerType: "Repository",
		}

		msg := err.Error()
		if msg == "" {
			t.Error("错误消息为空")
		}
	})

	t.Run("CircularDependencyError", func(t *testing.T) {
		err := &CircularDependencyError{
			Cycle: []string{"A", "B", "C"},
		}

		msg := err.Error()
		if msg == "" {
			t.Error("错误消息为空")
		}

		if !contains(msg, "circular dependency") {
			t.Error("错误消息不包含循环依赖信息")
		}
	})

	t.Run("AmbiguousMatchError", func(t *testing.T) {
		err := &AmbiguousMatchError{
			InstanceName: "TestService",
			FieldName:    "entity",
			FieldType:    reflect.TypeOf(struct{}{}),
			Candidates:   []string{"EntityA", "EntityB"},
		}

		msg := err.Error()
		if msg == "" {
			t.Error("错误消息为空")
		}
	})

	t.Run("DuplicateRegistrationError", func(t *testing.T) {
		err := &DuplicateRegistrationError{
			Name:     "test",
			Existing: &testImplementation{name: "test1"},
			New:      &testImplementation{name: "test2"},
		}

		msg := err.Error()
		if msg == "" {
			t.Error("错误消息为空")
		}
	})

	t.Run("InstanceNotFoundError", func(t *testing.T) {
		err := &InstanceNotFoundError{
			Name:  "TestService",
			Layer: "Service",
		}

		msg := err.Error()
		if msg == "" {
			t.Error("错误消息为空")
		}
	})

	t.Run("InterfaceAlreadyRegisteredError", func(t *testing.T) {
		err := &InterfaceAlreadyRegisteredError{
			InterfaceType: reflect.TypeOf((*testInterface)(nil)).Elem(),
			ExistingImpl:  &testImplementation{name: "test1"},
			NewImpl:       &testImplementation{name: "test2"},
		}

		msg := err.Error()
		if msg == "" {
			t.Error("错误消息为空")
		}
	})

	t.Run("ImplementationDoesNotImplementInterfaceError", func(t *testing.T) {
		err := &ImplementationDoesNotImplementInterfaceError{
			InterfaceType:  reflect.TypeOf((*testInterface)(nil)).Elem(),
			Implementation: &struct{}{},
		}

		msg := err.Error()
		if msg == "" {
			t.Error("错误消息为空")
		}
	})

	t.Run("InterfaceNotRegisteredError", func(t *testing.T) {
		err := &InterfaceNotRegisteredError{
			InterfaceType: reflect.TypeOf((*testInterface)(nil)).Elem(),
		}

		msg := err.Error()
		if msg == "" {
			t.Error("错误消息为空")
		}
	})

	t.Run("ManagerContainerNotSetError", func(t *testing.T) {
		err := &ManagerContainerNotSetError{
			Layer: "Service",
		}

		msg := err.Error()
		if msg == "" {
			t.Error("错误消息为空")
		}
	})
}

// 测试辅助函数
func TestExtractNameFromType(t *testing.T) {
	t.Run("指针类型", func(t *testing.T) {
		type TestStruct struct{}
		typ := reflect.TypeOf(&TestStruct{})
		name := extractNameFromType(typ)
		if name != "TestStruct" {
			t.Errorf("期望 TestStruct，实际: %s", name)
		}
	})

	t.Run("结构体类型", func(t *testing.T) {
		type TestStruct struct{}
		typ := reflect.TypeOf(TestStruct{})
		name := extractNameFromType(typ)
		if name != "TestStruct" {
			t.Errorf("期望 TestStruct，实际: %s", name)
		}
	})

	t.Run("接口类型", func(t *testing.T) {
		name := extractNameFromType(reflect.TypeOf((*testInterface)(nil)).Elem())
		if name == "" {
			t.Error("期望非空名称")
		}
	})
}

func TestToLowerCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Test", "test"},
		{"UserService", "userService"},
		{"A", "a"},
		{"", ""},
	}

	for _, test := range tests {
		result := toLowerCamelCase(test.input)
		if result != test.expected {
			t.Errorf("输入: %s, 期望: %s, 实际: %s", test.input, test.expected, result)
		}
	}
}

func TestGenericDependencyResolver(t *testing.T) {
	t.Run("多个依赖源", func(t *testing.T) {
		source1 := &mockContainerSource{
			types: map[reflect.Type]interface{}{
				reflect.TypeOf((*testInterface)(nil)).Elem(): &testImplementation{name: "source1"},
			},
		}

		source2 := &mockContainerSource{
			types: map[reflect.Type]interface{}{
				reflect.TypeOf((*testInterface)(nil)).Elem(): &testImplementation{name: "source2"},
			},
		}

		resolver := NewGenericDependencyResolver(source1, source2)

		ifaceType := reflect.TypeOf((*testInterface)(nil)).Elem()
		dep, err := resolver.ResolveDependency(ifaceType, reflect.TypeOf(struct{}{}), "test")
		if err != nil {
			t.Errorf("解析依赖失败: %v", err)
		}

		impl, ok := dep.(testInterface)
		if !ok {
			t.Fatal("期望 testInterface 类型")
		}

		if impl.testMethod() != "source1" {
			t.Errorf("期望从 source1 获取，实际: %s", impl.testMethod())
		}
	})

	t.Run("找不到依赖", func(t *testing.T) {
		source := &mockContainerSource{
			types: map[reflect.Type]interface{}{},
		}

		resolver := NewGenericDependencyResolver(source)

		ifaceType := reflect.TypeOf((*testInterface)(nil)).Elem()
		_, err := resolver.ResolveDependency(ifaceType, reflect.TypeOf(struct{}{}), "test")

		if err == nil {
			t.Error("期望返回错误，但没有")
		}
	})
}

type mockContainerSource struct {
	types map[reflect.Type]interface{}
}

func (m *mockContainerSource) GetDependency(fieldType reflect.Type) (interface{}, error) {
	if impl, exists := m.types[fieldType]; exists {
		return impl, nil
	}
	return nil, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || (len(s) > len(substr) && contains(s[1:], substr))
}
