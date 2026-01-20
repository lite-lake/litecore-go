package common

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockConfigProvider struct {
	configs map[string]any
}

func (m *mockConfigProvider) ConfigProviderName() string {
	return "MockConfigProvider"
}

func (m *mockConfigProvider) Get(key string) (any, error) {
	if val, exists := m.configs[key]; exists {
		return val, nil
	}
	return nil, errors.New("配置项不存在")
}

func (m *mockConfigProvider) Has(key string) bool {
	_, exists := m.configs[key]
	return exists
}

type nestedConfigProvider struct {
	configs map[string]any
}

func (n *nestedConfigProvider) ConfigProviderName() string {
	return "NestedConfigProvider"
}

func (n *nestedConfigProvider) Get(key string) (any, error) {
	val, exists := n.configs[key]
	if !exists {
		return nil, errors.New("配置项不存在")
	}
	return val, nil
}

func (n *nestedConfigProvider) Has(key string) bool {
	_, exists := n.configs[key]
	return exists
}

func TestIBaseConfigProvider_基础接口实现(t *testing.T) {
	provider := &mockConfigProvider{
		configs: map[string]any{
			"app.name": "test-app",
			"app.port": 8080,
		},
	}

	assert.Equal(t, "MockConfigProvider", provider.ConfigProviderName())
	assert.True(t, provider.Has("app.name"))
	assert.False(t, provider.Has("nonexistent"))

	val, err := provider.Get("app.name")
	assert.NoError(t, err)
	assert.Equal(t, "test-app", val)
}

func TestIBaseConfigProvider_Get方法(t *testing.T) {
	tests := []struct {
		name      string
		provider  IBaseConfigProvider
		key       string
		wantValue any
		wantErr   bool
	}{
		{
			name: "获取字符串配置",
			provider: &mockConfigProvider{
				configs: map[string]any{"string.key": "value"},
			},
			key:       "string.key",
			wantValue: "value",
			wantErr:   false,
		},
		{
			name: "获取整数配置",
			provider: &mockConfigProvider{
				configs: map[string]any{"int.key": 123},
			},
			key:       "int.key",
			wantValue: 123,
			wantErr:   false,
		},
		{
			name: "获取布尔配置",
			provider: &mockConfigProvider{
				configs: map[string]any{"bool.key": true},
			},
			key:       "bool.key",
			wantValue: true,
			wantErr:   false,
		},
		{
			name: "获取不存在的配置",
			provider: &mockConfigProvider{
				configs: map[string]any{"exist": "value"},
			},
			key:       "nonexistent",
			wantValue: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.provider.Get(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, val)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantValue, val)
			}
		})
	}
}

func TestIBaseConfigProvider_Has方法(t *testing.T) {
	tests := []struct {
		name     string
		provider IBaseConfigProvider
		key      string
		wantBool bool
	}{
		{
			name: "存在的配置项",
			provider: &mockConfigProvider{
				configs: map[string]any{"exists": true},
			},
			key:      "exists",
			wantBool: true,
		},
		{
			name: "不存在的配置项",
			provider: &mockConfigProvider{
				configs: map[string]any{"exists": true},
			},
			key:      "notexists",
			wantBool: false,
		},
		{
			name: "空键",
			provider: &mockConfigProvider{
				configs: map[string]any{"": "value"},
			},
			key:      "",
			wantBool: true,
		},
		{
			name: "嵌套路径键",
			provider: &nestedConfigProvider{
				configs: map[string]any{
					"aaa.bbb.ccc": "nested-value",
				},
			},
			key:      "aaa.bbb.ccc",
			wantBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantBool, tt.provider.Has(tt.key))
		})
	}
}

func TestIBaseConfigProvider_空实现(t *testing.T) {
	tests := []struct {
		name     string
		provider IBaseConfigProvider
	}{
		{
			name: "空配置提供者",
			provider: &mockConfigProvider{
				configs: map[string]any{},
			},
		},
		{
			name: "空映射配置提供者",
			provider: &mockConfigProvider{
				configs: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.provider.ConfigProviderName())
			assert.False(t, tt.provider.Has("any.key"))

			_, err := tt.provider.Get("any.key")
			assert.Error(t, err)
		})
	}
}

func TestIBaseConfigProvider_接口组合(t *testing.T) {
	provider := &mockConfigProvider{
		configs: map[string]any{
			"combined": "value",
		},
	}

	var iface IBaseConfigProvider = provider
	assert.Equal(t, "MockConfigProvider", iface.ConfigProviderName())
	assert.True(t, iface.Has("combined"))
}
