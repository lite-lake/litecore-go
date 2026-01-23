package configmgr

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadJSON 加载 JSON 配置文件并返回配置数据
func LoadJSON(filePath string) (map[string]any, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read json file: %w", err)
	}

	var configData map[string]any
	if err := json.Unmarshal(data, &configData); err != nil {
		return nil, fmt.Errorf("failed to parse json: %w", err)
	}

	return configData, nil
}
