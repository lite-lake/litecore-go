package configmgr

import (
	"fmt"
)

// NewConfigManager 根据驱动类型创建配置管理器
func NewConfigManager(driver string, filePath string) (IConfigManager, error) {
	switch driver {
	case "yaml":
		return newBaseConfigManager(
			"ConfigYamlManager",
			func() (map[string]any, error) {
				return LoadYAML(filePath)
			},
		)
	case "json":
		return newBaseConfigManager(
			"ConfigJsonManager",
			func() (map[string]any, error) {
				return LoadJSON(filePath)
			},
		)
	default:
		return nil, fmt.Errorf("unsupported config driver: '%s'", driver)
	}
}
