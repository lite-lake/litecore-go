package drivers

import (
	"com.litelake.litecore/config/common"
)

type JsonConfigProvider struct {
	configData map[string]any
}

func NewJsonConfigProvider(filePath string) (common.ConfigProvider, error) {
	// TODO 从文件里读取并初始化
	demoData := map[string]any{
		"key1": "value1",
		"key2": 123,
	}
	return &JsonConfigProvider{
		configData: demoData,
	}, nil
}

func (p *JsonConfigProvider) GetConfig(key string) (any, error) {
	// TODO
	return nil, nil
}
