package drivers

import (
	"com.litelake.litecore/config/common"
)

type YamlConfigProvider struct {
	configData map[string]any
}

func NewYamlConfigProvider(filePath string) (common.ConfigProvider, error) {
	// TODO 从文件里读取并初始化
	demoData := map[string]any{
		"key1": "value1",
		"key2": 123,
	}
	return &YamlConfigProvider{
		configData: demoData,
	}, nil
}

func (p *YamlConfigProvider) GetConfig(key string) (any, error) {
	// TODO
	return nil, nil
}
