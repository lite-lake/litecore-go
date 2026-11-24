package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetString_获取字符串(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		want      string
		wantError bool
	}{
		{
			name:      "正常字符串",
			value:     "hello",
			want:      "hello",
			wantError: false,
		},
		{
			name:      "空字符串",
			value:     "",
			want:      "",
			wantError: false,
		},
		{
			name:      "整型值",
			value:     123,
			want:      "",
			wantError: true,
		},
		{
			name:      "浮点型值",
			value:     3.14,
			want:      "",
			wantError: true,
		},
		{
			name:      "布尔值",
			value:     true,
			want:      "",
			wantError: true,
		},
		{
			name:      "nil值",
			value:     nil,
			want:      "",
			wantError: true,
		},
		{
			name:      "切片类型",
			value:     []string{"a", "b"},
			want:      "",
			wantError: true,
		},
		{
			name:      "Map类型",
			value:     map[string]any{"key": "value"},
			want:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetString(tt.value)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetStringOrDefault_获取字符串或默认值(t *testing.T) {
	tests := []struct {
		name         string
		value        any
		defaultValue string
		want         string
	}{
		{
			name:         "正常字符串",
			value:        "hello",
			defaultValue: "default",
			want:         "hello",
		},
		{
			name:         "空字符串",
			value:        "",
			defaultValue: "default",
			want:         "",
		},
		{
			name:         "整型值返回默认值",
			value:        123,
			defaultValue: "default",
			want:         "default",
		},
		{
			name:         "nil值返回默认值",
			value:        nil,
			defaultValue: "default",
			want:         "default",
		},
		{
			name:         "布尔值返回默认值",
			value:        true,
			defaultValue: "default",
			want:         "default",
		},
		{
			name:         "空默认值",
			value:        nil,
			defaultValue: "",
			want:         "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetStringOrDefault(tt.value, tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetMap_获取Map(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		want      map[string]any
		wantError bool
	}{
		{
			name: "正常Map",
			value: map[string]any{
				"key1": "value1",
				"key2": 123,
			},
			want: map[string]any{
				"key1": "value1",
				"key2": 123,
			},
			wantError: false,
		},
		{
			name:      "空Map",
			value:     map[string]any{},
			want:      map[string]any{},
			wantError: false,
		},
		{
			name:      "字符串值",
			value:     "string",
			want:      nil,
			wantError: true,
		},
		{
			name:      "整型值",
			value:     123,
			want:      nil,
			wantError: true,
		},
		{
			name:      "nil值",
			value:     nil,
			want:      nil,
			wantError: true,
		},
		{
			name:      "切片类型",
			value:     []string{"a", "b"},
			want:      nil,
			wantError: true,
		},
		{
			name:      "整数Key的Map",
			value:     map[int]string{1: "one"},
			want:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMap(tt.value)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetMapOrDefault_获取Map或默认值(t *testing.T) {
	tests := []struct {
		name         string
		value        any
		defaultValue map[string]any
		want         map[string]any
	}{
		{
			name: "正常Map",
			value: map[string]any{
				"key1": "value1",
			},
			defaultValue: map[string]any{"default": "value"},
			want: map[string]any{
				"key1": "value1",
			},
		},
		{
			name:         "空Map",
			value:        map[string]any{},
			defaultValue: map[string]any{"default": "value"},
			want:         map[string]any{},
		},
		{
			name:         "字符串值返回默认值",
			value:        "string",
			defaultValue: map[string]any{"default": "value"},
			want:         map[string]any{"default": "value"},
		},
		{
			name:         "nil值返回默认值",
			value:        nil,
			defaultValue: map[string]any{"default": "value"},
			want:         map[string]any{"default": "value"},
		},
		{
			name:         "整型值返回默认值",
			value:        123,
			defaultValue: map[string]any{"default": "value"},
			want:         map[string]any{"default": "value"},
		},
		{
			name:         "空默认值",
			value:        nil,
			defaultValue: map[string]any{},
			want:         map[string]any{},
		},
		{
			name:         "nil默认值",
			value:        nil,
			defaultValue: nil,
			want:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetMapOrDefault(tt.value, tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTypeUtils_嵌套类型(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		wantError bool
	}{
		{
			name:      "嵌套Map包含字符串",
			value:     map[string]any{"nested": "string"},
			wantError: false,
		},
		{
			name:      "嵌套Map包含Map",
			value:     map[string]any{"nested": map[string]any{"inner": "value"}},
			wantError: false,
		},
		{
			name:      "嵌套Map包含切片",
			value:     map[string]any{"nested": []string{"a", "b"}},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetMap(tt.value)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
