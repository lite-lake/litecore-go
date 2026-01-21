package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/lite-lake/litecore-go/common"
)

type JsonConfigProvider struct {
	base *BaseConfigProvider
}

func NewJsonConfigProvider(filePath string) (common.IBaseConfigProvider, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read json file: %w", err)
	}

	var configData map[string]any
	if err := json.Unmarshal(data, &configData); err != nil {
		return nil, fmt.Errorf("failed to parse json: %w", err)
	}

	return &JsonConfigProvider{
		base: NewBaseConfigProvider(configData),
	}, nil
}

func (p *JsonConfigProvider) Get(key string) (any, error) {
	return p.base.Get(key)
}

func (p *JsonConfigProvider) Has(key string) bool {
	return p.base.Has(key)
}

func (p *JsonConfigProvider) ConfigProviderName() string {
	return "JsonConfigProvider"
}
