package container

import (
	"reflect"
	"testing"

	"com.litelake.litecore/common"
	"github.com/stretchr/testify/assert"
)

// TestConfigContainer 测试 ConfigContainer
func TestConfigContainer(t *testing.T) {
	t.Run("注册配置", testRegister)
	t.Run("获取配置", testGet)
	t.Run("获取所有配置", testGetAll)
	t.Run("统计数量", testCount)
	t.Run("注入所有依赖", testInjectAll)
	t.Run("获取依赖", testGetDependency)
}

// testRegister 测试注册功能
func testRegister(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*ConfigContainer)
		ifaceType   reflect.Type
		impl        common.IBaseConfigProvider
		wantErr     error
		errContains string
	}{
		{
			name:      "正常注册配置",
			ifaceType: reflect.TypeOf((*common.IBaseConfigProvider)(nil)).Elem(),
			impl:      &MockConfigProvider{name: "config-1"},
			wantErr:   nil,
		},
		{
			name: "重复注册相同接口",
			setup: func(c *ConfigContainer) {
				c.RegisterByType(reflect.TypeOf((*common.IBaseConfigProvider)(nil)).Elem(), &MockConfigProvider{name: "config-1"})
			},
			ifaceType:   reflect.TypeOf((*common.IBaseConfigProvider)(nil)).Elem(),
			impl:        &MockConfigProvider{name: "config-2"},
			errContains: "already registered",
		},
		{
			name:        "注册nil实现",
			ifaceType:   reflect.TypeOf((*common.IBaseConfigProvider)(nil)).Elem(),
			impl:        nil,
			errContains: "nil",
		},
		{
			name:        "注册不实现接口的实现",
			ifaceType:   reflect.TypeOf((*testInterface)(nil)).Elem(),
			impl:        &MockConfigProvider{name: "config-1"},
			errContains: "does not implement",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfigContainer()
			if tt.setup != nil {
				tt.setup(c)
			}
			err := c.RegisterByType(tt.ifaceType, tt.impl)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else if tt.errContains != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// testGet 测试获取功能
func testGet(t *testing.T) {
	c := NewConfigContainer()
	config1 := &MockConfigProvider{name: "config-1"}

	baseConfigType := reflect.TypeOf((*common.IBaseConfigProvider)(nil)).Elem()
	c.RegisterByType(baseConfigType, config1)

	tests := []struct {
		name      string
		ifaceType reflect.Type
		want      common.IBaseConfigProvider
	}{
		{
			name:      "获取已注册的配置",
			ifaceType: baseConfigType,
			want:      config1,
		},
		{
			name:      "获取不存在的配置",
			ifaceType: reflect.TypeOf((*testInterface)(nil)).Elem(),
			want:      nil,
		},
		{
			name:      "获取空类型",
			ifaceType: nil,
			want:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.GetByType(tt.ifaceType)
			assert.Equal(t, tt.want, got)
		})
	}
}

// testGetAll 测试获取所有配置
func testGetAll(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*ConfigContainer)
		wantLen  int
		wantName []string
	}{
		{
			name:     "空容器获取所有",
			setup:    func(c *ConfigContainer) {},
			wantLen:  0,
			wantName: nil,
		},
		{
			name: "获取单个配置",
			setup: func(c *ConfigContainer) {
				c.RegisterByType(reflect.TypeOf((*common.IBaseConfigProvider)(nil)).Elem(), &MockConfigProvider{name: "config-b"})
			},
			wantLen:  1,
			wantName: []string{"config-b"},
		},
		{
			name: "获取多个配置并验证排序",
			setup: func(c *ConfigContainer) {
				c.RegisterByType(reflect.TypeOf((*common.IBaseConfigProvider)(nil)).Elem(), &MockConfigProvider{name: "config-b"})
				c.RegisterByType(reflect.TypeOf((*IMockConfig)(nil)).Elem(), &MockConfig{name: "config-a"})
			},
			wantLen:  2,
			wantName: []string{"config-a", "config-b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfigContainer()
			tt.setup(c)
			got := c.GetAll()
			assert.Equal(t, tt.wantLen, len(got))
			if tt.wantName != nil {
				for i, name := range tt.wantName {
					assert.Equal(t, name, got[i].ConfigProviderName())
				}
			}
		})
	}
}

// testCount 测试统计数量
func testCount(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*ConfigContainer)
		wantCnt int
	}{
		{
			name:    "空容器统计",
			setup:   func(c *ConfigContainer) {},
			wantCnt: 0,
		},
		{
			name: "单个配置统计",
			setup: func(c *ConfigContainer) {
				c.RegisterByType(reflect.TypeOf((*common.IBaseConfigProvider)(nil)).Elem(), &MockConfigProvider{name: "config-1"})
			},
			wantCnt: 1,
		},
		{
			name: "多个配置统计",
			setup: func(c *ConfigContainer) {
				c.RegisterByType(reflect.TypeOf((*common.IBaseConfigProvider)(nil)).Elem(), &MockConfigProvider{name: "config-1"})
				c.RegisterByType(reflect.TypeOf((*IMockConfig)(nil)).Elem(), &MockConfig{name: "config-2"})
			},
			wantCnt: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfigContainer()
			tt.setup(c)
			assert.Equal(t, tt.wantCnt, c.Count())
		})
	}
}

// testInjectAll 测试注入所有依赖
func testInjectAll(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*ConfigContainer)
		wantErr error
	}{
		{
			name:    "空容器注入",
			setup:   func(c *ConfigContainer) {},
			wantErr: nil,
		},
		{
			name: "正常配置注入",
			setup: func(c *ConfigContainer) {
				c.RegisterByType(reflect.TypeOf((*common.IBaseConfigProvider)(nil)).Elem(), &MockConfigProvider{name: "config-1"})
			},
			wantErr: nil,
		},
		{
			name: "多个配置注入",
			setup: func(c *ConfigContainer) {
				c.RegisterByType(reflect.TypeOf((*common.IBaseConfigProvider)(nil)).Elem(), &MockConfigProvider{name: "config-1"})
				c.RegisterByType(reflect.TypeOf((*IMockConfig)(nil)).Elem(), &MockConfig{name: "config-2"})
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfigContainer()
			tt.setup(c)
			err := c.InjectAll()
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

// testGetDependency 测试获取依赖
func testGetDependency(t *testing.T) {
	c := NewConfigContainer()
	config1 := &MockConfigProvider{name: "config-1"}
	baseConfigType := reflect.TypeOf((*common.IBaseConfigProvider)(nil)).Elem()
	c.RegisterByType(baseConfigType, config1)

	tests := []struct {
		name      string
		fieldType reflect.Type
		want      interface{}
		wantErr   error
	}{
		{
			name:      "获取已注册的基础配置类型依赖",
			fieldType: baseConfigType,
			want:      config1,
			wantErr:   nil,
		},
		{
			name:      "获取未注册的基础配置类型依赖",
			fieldType: reflect.TypeOf((*IMockConfig)(nil)).Elem(),
			want:      nil,
			wantErr:   &DependencyNotFoundError{},
		},
		{
			name:      "获取非配置类型依赖",
			fieldType: reflect.TypeOf(0),
			want:      nil,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.GetDependency(tt.fieldType)
			if tt.wantErr != nil {
				assert.Error(t, err)
				_, ok := err.(*DependencyNotFoundError)
				assert.True(t, ok)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

// testInterface 测试接口
type testInterface interface {
	TestMethod()
}

// IMockConfig Mock 配置接口
type IMockConfig interface {
	common.IBaseConfigProvider
}

// MockConfig Mock 配置实现
type MockConfig struct {
	name string
}

func (m *MockConfig) ConfigProviderName() string {
	return m.name
}

func (m *MockConfig) Get(key string) (any, error) {
	return nil, nil
}

func (m *MockConfig) Has(key string) bool {
	return false
}
