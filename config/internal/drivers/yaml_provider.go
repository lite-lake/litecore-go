package drivers

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
	"com.litelake.litecore/config/common"
)

type YamlConfigProvider struct {
	base *BaseConfigProvider
}

func NewYamlConfigProvider(filePath string) (common.ConfigProvider, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read yaml file: %w", err)
	}

	var configData map[string]any
	if err := yaml.Unmarshal(data, &configData); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	return &YamlConfigProvider{
		base: NewBaseConfigProvider(configData),
	}, nil
}

func (p *YamlConfigProvider) Get(key string) (any, error) {
	return p.base.Get(key)
}

func (p *YamlConfigProvider) Has(key string) bool {
	return p.base.Has(key)
}
