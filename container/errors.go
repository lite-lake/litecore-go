package container

import (
	"fmt"
	"reflect"
	"strings"
)

// DependencyNotFoundError 依赖缺失错误
type DependencyNotFoundError struct {
	InstanceName  string       // 当前实例名称
	FieldName     string       // 缺失依赖的字段名
	FieldType     reflect.Type // 期望的依赖类型
	ContainerType string       // 应该从哪个容器查找
}

func (e *DependencyNotFoundError) Error() string {
	return fmt.Sprintf("dependency not found for %s.%s: need type %s from %s container",
		e.InstanceName, e.FieldName, e.FieldType, e.ContainerType)
}

// CircularDependencyError 循环依赖错误
type CircularDependencyError struct {
	Cycle []string // 循环依赖链
}

func (e *CircularDependencyError) Error() string {
	if len(e.Cycle) == 0 {
		return "circular dependency detected"
	}
	return fmt.Sprintf("circular dependency detected: %s → %s",
		strings.Join(e.Cycle, " → "), e.Cycle[0])
}

// AmbiguousMatchError 多重匹配错误
type AmbiguousMatchError struct {
	InstanceName string
	FieldName    string
	FieldType    reflect.Type
	Candidates   []string // 匹配的候选实例名称
}

func (e *AmbiguousMatchError) Error() string {
	return fmt.Sprintf("ambiguous match for %s.%s: type %s matches multiple instances: %s",
		e.InstanceName, e.FieldName, e.FieldType, strings.Join(e.Candidates, ", "))
}

// DuplicateRegistrationError 重复注册错误
type DuplicateRegistrationError struct {
	Name     string
	Existing interface{}
	New      interface{}
}

func (e *DuplicateRegistrationError) Error() string {
	return fmt.Sprintf("duplicate registration: name '%s' already exists", e.Name)
}

// InstanceNotFoundError 实例未找到错误
type InstanceNotFoundError struct {
	Name  string
	Layer string
}

func (e *InstanceNotFoundError) Error() string {
	return fmt.Sprintf("%s instance not found: '%s'", e.Layer, e.Name)
}
