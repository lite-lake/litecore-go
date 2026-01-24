package limitermgr

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockConfigManager struct {
	config map[string]any
	err    error
}

func (m *mockConfigManager) ManagerName() string {
	return "mockConfigManager"
}

func (m *mockConfigManager) Health() error {
	return nil
}

func (m *mockConfigManager) OnStart() error {
	return nil
}

func (m *mockConfigManager) OnStop() error {
	return nil
}

func (m *mockConfigManager) Get(key string) (any, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.config == nil {
		return nil, fmt.Errorf("config not found: %s", key)
	}
	val, ok := m.config[key]
	if !ok {
		return nil, fmt.Errorf("config not found: %s", key)
	}
	return val, nil
}

func (m *mockConfigManager) Has(key string) bool {
	if m.config == nil {
		return false
	}
	_, ok := m.config[key]
	return ok
}

func TestBuild(t *testing.T) {
	t.Run("空字符串驱动类型", func(t *testing.T) {
		mgr, err := Build("", nil, nil, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "driver type is required")
	})

	t.Run("Redis驱动配置正确", func(t *testing.T) {
		config := map[string]any{
			"host":              "localhost",
			"port":              6379,
			"password":          "pass123",
			"db":                0,
			"max_idle_conns":    10,
			"max_open_conns":    100,
			"conn_max_lifetime": 30,
		}
		mgr, err := Build("redis", config, nil, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.IsType(t, &limiterManagerRedisImpl{}, mgr)
		assert.Equal(t, "limiterManagerRedisImpl", mgr.ManagerName())
		assert.Implements(t, (*ILimiterManager)(nil), mgr)
	})

	t.Run("Redis驱动配置为nil", func(t *testing.T) {
		mgr, err := Build("redis", nil, nil, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.IsType(t, &limiterManagerRedisImpl{}, mgr)
	})

	t.Run("Memory驱动配置正确", func(t *testing.T) {
		config := map[string]any{
			"max_backups": 1000,
		}
		mgr, err := Build("memory", config, nil, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.IsType(t, &limiterManagerMemoryImpl{}, mgr)
		assert.Equal(t, "limiterManagerMemoryImpl", mgr.ManagerName())
		assert.Implements(t, (*ILimiterManager)(nil), mgr)
	})

	t.Run("Memory驱动配置为nil", func(t *testing.T) {
		mgr, err := Build("memory", nil, nil, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.IsType(t, &limiterManagerMemoryImpl{}, mgr)
	})

	t.Run("不支持的驱动类型-mysql", func(t *testing.T) {
		mgr, err := Build("mysql", nil, nil, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "unsupported driver type: mysql")
	})

	t.Run("不支持的驱动类型-postgres", func(t *testing.T) {
		mgr, err := Build("postgres", nil, nil, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "unsupported driver type: postgres")
	})

	t.Run("不支持的驱动类型-未知字符串", func(t *testing.T) {
		mgr, err := Build("unknown", nil, nil, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "unsupported driver type: unknown")
	})
}

func TestBuildWithConfigProvider(t *testing.T) {
	t.Run("configProvider为nil", func(t *testing.T) {
		mgr, err := BuildWithConfigProvider(nil, nil, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "configProvider cannot be nil")
	})

	t.Run("获取limiter.driver失败", func(t *testing.T) {
		mockMgr := &mockConfigManager{
			err: errors.New("config error"),
		}
		mgr, err := BuildWithConfigProvider(mockMgr, nil, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "failed to get limiter.driver")
	})

	t.Run("limiter.driver不是字符串类型", func(t *testing.T) {
		mockMgr := &mockConfigManager{
			config: map[string]any{
				"limiter.driver": 123,
			},
		}
		mgr, err := BuildWithConfigProvider(mockMgr, nil, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "limiter.driver")
	})

	t.Run("Redis驱动-获取redis_config失败", func(t *testing.T) {
		mockMgr := &mockConfigManager{
			config: map[string]any{
				"limiter.driver": "redis",
			},
		}
		mgr, err := BuildWithConfigProvider(mockMgr, nil, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "failed to get limiter.redis_config")
	})

	t.Run("Redis驱动-redis_config不是map类型", func(t *testing.T) {
		mockMgr := &mockConfigManager{
			config: map[string]any{
				"limiter.driver":       "redis",
				"limiter.redis_config": "not_a_map",
			},
		}
		mgr, err := BuildWithConfigProvider(mockMgr, nil, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "limiter.redis_config")
	})

	t.Run("Redis驱动-配置正确", func(t *testing.T) {
		mockMgr := &mockConfigManager{
			config: map[string]any{
				"limiter.driver": "redis",
				"limiter.redis_config": map[string]any{
					"host":              "localhost",
					"port":              6379,
					"password":          "pass123",
					"db":                0,
					"max_idle_conns":    10,
					"max_open_conns":    100,
					"conn_max_lifetime": 30,
				},
			},
		}
		mgr, err := BuildWithConfigProvider(mockMgr, nil, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.IsType(t, &limiterManagerRedisImpl{}, mgr)
		assert.Equal(t, "limiterManagerRedisImpl", mgr.ManagerName())
	})

	t.Run("Redis驱动-空配置map", func(t *testing.T) {
		mockMgr := &mockConfigManager{
			config: map[string]any{
				"limiter.driver":       "redis",
				"limiter.redis_config": map[string]any{},
			},
		}
		mgr, err := BuildWithConfigProvider(mockMgr, nil, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.IsType(t, &limiterManagerRedisImpl{}, mgr)
	})

	t.Run("Memory驱动-获取memory_config失败", func(t *testing.T) {
		mockMgr := &mockConfigManager{
			config: map[string]any{
				"limiter.driver": "memory",
			},
		}
		mgr, err := BuildWithConfigProvider(mockMgr, nil, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "failed to get limiter.memory_config")
	})

	t.Run("Memory驱动-memory_config不是map类型", func(t *testing.T) {
		mockMgr := &mockConfigManager{
			config: map[string]any{
				"limiter.driver":        "memory",
				"limiter.memory_config": "not_a_map",
			},
		}
		mgr, err := BuildWithConfigProvider(mockMgr, nil, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "limiter.memory_config")
	})

	t.Run("Memory驱动-配置正确", func(t *testing.T) {
		mockMgr := &mockConfigManager{
			config: map[string]any{
				"limiter.driver": "memory",
				"limiter.memory_config": map[string]any{
					"max_backups": 1000,
				},
			},
		}
		mgr, err := BuildWithConfigProvider(mockMgr, nil, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.IsType(t, &limiterManagerMemoryImpl{}, mgr)
		assert.Equal(t, "limiterManagerMemoryImpl", mgr.ManagerName())
	})

	t.Run("Memory驱动-空配置map", func(t *testing.T) {
		mockMgr := &mockConfigManager{
			config: map[string]any{
				"limiter.driver":        "memory",
				"limiter.memory_config": map[string]any{},
			},
		}
		mgr, err := BuildWithConfigProvider(mockMgr, nil, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.IsType(t, &limiterManagerMemoryImpl{}, mgr)
	})

	t.Run("不支持的驱动类型", func(t *testing.T) {
		mockMgr := &mockConfigManager{
			config: map[string]any{
				"limiter.driver": "mysql",
			},
		}
		mgr, err := BuildWithConfigProvider(mockMgr, nil, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "unsupported driver type: mysql")
	})
}
