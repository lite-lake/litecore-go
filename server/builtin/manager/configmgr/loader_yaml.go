package configmgr

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadYAML 加载 YAML 配置文件并返回配置数据
func LoadYAML(filePath string) (map[string]any, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read yaml file: %w", err)
	}

	var configData map[string]any
	if err := yaml.Unmarshal(data, &configData); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	return configData, nil
}
