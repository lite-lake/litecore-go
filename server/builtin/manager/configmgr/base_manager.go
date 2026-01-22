package configmgr

import (
	"fmt"
	"regexp"
	"strconv"
)

// 预编译路径正则表达式，提升性能
var pathPattern = regexp.MustCompile(`([^\.\[\]]+)(?:\[(\d+)\])?`)

// baseConfigManager 提供配置查询的公共实现
// 配置数据在创建后不可变，因此可以安全地在多个 goroutine 之间共享使用
type baseConfigManager struct {
	managerName string
	configData  map[string]any
}

// newBaseConfigManager 创建基础配置管理器
func newBaseConfigManager(managerName string, handler IConfigLoadHandler) (IConfigManager, error) {
	data, err := handler()
	if err != nil {
		return nil, err
	}
	return &baseConfigManager{
		managerName: managerName,
		configData:  data,
	}, nil
}

func (p *baseConfigManager) ManagerName() string {
	return p.managerName
}

func (p *baseConfigManager) Health() error {
	return nil
}

func (p *baseConfigManager) OnStart() error {
	return nil
}

func (p *baseConfigManager) OnStop() error {
	return nil
}

// Get 获取配置项
// 支持路径语法：
//   - 点分隔: aaa.bbb.ccc
//   - 数组索引: servers[0].port, items[2]
func (p *baseConfigManager) Get(key string) (any, error) {
	if key == "" {
		return p.configData, nil
	}

	return p.navigatePath(p.configData, key)
}

// pathPart 表示路径中的一个部分
type pathPart struct {
	key      string
	index    int // 数组索引，-1 表示非数组
	hasIndex bool
}

// parsePath 解析路径字符串为路径部分列表
// 例如: "servers[0].port" -> [{key: "servers", index: 0}, {key: "port", index: -1}]
func (p *baseConfigManager) parsePath(path string) ([]pathPart, error) {
	var parts []pathPart

	// 使用预编译的正则表达式匹配: key[index] 或 key
	// 例如: servers[0], port, database.host
	matches := pathPattern.FindAllStringSubmatch(path, -1)

	if matches == nil {
		return nil, fmt.Errorf("invalid path syntax: %s", path)
	}

	for _, match := range matches {
		part := pathPart{
			key:      match[1],
			index:    -1,
			hasIndex: false,
		}

		// 如果有数组索引
		if match[2] != "" {
			idx, err := strconv.Atoi(match[2])
			if err != nil {
				return nil, fmt.Errorf("invalid array index in path: %s", path)
			}
			part.index = idx
			part.hasIndex = true
		}

		parts = append(parts, part)
	}

	return parts, nil
}

// navigatePath 在配置数据中导航到指定路径
func (p *baseConfigManager) navigatePath(data map[string]any, path string) (any, error) {
	parts, err := p.parsePath(path)
	if err != nil {
		return nil, err
	}

	var current any = data

	for i, part := range parts {
		// 获取当前层级（必须是 map[string]any）
		currentMap, ok := current.(map[string]any)
		if !ok {
			return nil, p.pathError(path, i, "expected object, got %T", current)
		}

		value, exists := currentMap[part.key]
		if !exists {
			return nil, fmt.Errorf("configmgr key '%s' not found", path)
		}

		// 如果是最后一部分，返回值（可能需要提取数组元素）
		if i == len(parts)-1 {
			if part.hasIndex {
				return p.getArrayElement(value, part.key, part.index, path)
			}
			return value, nil
		}

		// 继续向下导航
		if part.hasIndex {
			arrValue, err := p.getArrayElement(value, part.key, part.index, path)
			if err != nil {
				return nil, err
			}
			current = arrValue
		} else {
			current = value
		}
	}

	return current, nil
}

// getArrayElement 从数组中获取指定索引的元素
func (p *baseConfigManager) getArrayElement(value any, key string, index int, path string) (any, error) {
	arr, ok := value.([]any)
	if !ok {
		return nil, p.pathError(path, -1, "key '%s' is not an array (got %T)", key, value)
	}

	if index < 0 || index >= len(arr) {
		return nil, p.pathError(path, -1, "array index %d out of bounds (length: %d) for key '%s'", index, len(arr), key)
	}

	return arr[index], nil
}

// pathError 生成路径相关的错误信息
func (p *baseConfigManager) pathError(path string, part int, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("configmgr path '%s': %s", path, msg)
}

// Has 检查配置项是否存在
func (p *baseConfigManager) Has(key string) bool {
	_, err := p.Get(key)
	return err == nil
}

var _ IConfigManager = (*baseConfigManager)(nil)
