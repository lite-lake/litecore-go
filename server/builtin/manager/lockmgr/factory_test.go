package lockmgr

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockConfigManager struct {
	data map[string]any
	err  error
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
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("key not found: %s", key)
}

func (m *mockConfigManager) Has(key string) bool {
	if m.err != nil {
		return false
	}
	_, ok := m.data[key]
	return ok
}

func (m *mockConfigManager) Set(key string, value any) error {
	if m.data == nil {
		m.data = make(map[string]any)
	}
	m.data[key] = value
	return nil
}

func TestBuild(t *testing.T) {
	t.Run("默认驱动（空字符串）返回 memory 驱动", func(t *testing.T) {
		mgr, err := Build("", map[string]any{})
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.IsType(t, &lockManagerMemoryImpl{}, mgr)
	})

	t.Run("创建 Redis 驱动", func(t *testing.T) {
		redisConfig := map[string]any{
			"host":              "localhost",
			"port":              6379,
			"password":          "test-pass",
			"db":                0,
			"max_idle_conns":    10,
			"max_open_conns":    100,
			"conn_max_lifetime": "30s",
		}
		mgr, err := Build("redis", redisConfig)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.IsType(t, &lockManagerRedisImpl{}, mgr)
	})

	t.Run("创建 Memory 驱动", func(t *testing.T) {
		memoryConfig := map[string]any{
			"max_backups": 1000,
		}
		mgr, err := Build("memory", memoryConfig)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.IsType(t, &lockManagerMemoryImpl{}, mgr)
	})

	t.Run("创建的实例实现了 ILockManager 接口", func(t *testing.T) {
		memoryMgr, err := Build("memory", map[string]any{})
		assert.NoError(t, err)
		assert.Implements(t, (*ILockManager)(nil), memoryMgr)

		redisConfig := map[string]any{
			"host": "localhost",
			"port": 6379,
			"db":   0,
		}
		redisMgr, err := Build("redis", redisConfig)
		assert.NoError(t, err)
		assert.Implements(t, (*ILockManager)(nil), redisMgr)
	})

	t.Run("不支持的驱动类型返回错误", func(t *testing.T) {
		mgr, err := Build("invalid-driver", map[string]any{})
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "unsupported driver type")
	})

	t.Run("配置解析错误处理", func(t *testing.T) {
		testCases := []struct {
			name    string
			driver  string
			config  map[string]any
			wantErr string
		}{
			{
				name:    "nil 配置",
				driver:  "memory",
				config:  nil,
				wantErr: "",
			},
			{
				name:    "空配置",
				driver:  "memory",
				config:  map[string]any{},
				wantErr: "",
			},
			{
				name:   "有效的 Redis 配置",
				driver: "redis",
				config: map[string]any{
					"host": "127.0.0.1",
					"port": 6380,
					"db":   1,
				},
				wantErr: "",
			},
			{
				name:   "有效的 Memory 配置",
				driver: "memory",
				config: map[string]any{
					"max_backups": 500,
				},
				wantErr: "",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mgr, err := Build(tc.driver, tc.config)
				if tc.wantErr != "" {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tc.wantErr)
					assert.Nil(t, mgr)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, mgr)
					assert.Implements(t, (*ILockManager)(nil), mgr)
				}
			})
		}
	})
}

func TestBuildWithConfigProvider(t *testing.T) {
	t.Run("成功创建 Redis 驱动", func(t *testing.T) {
		mockCfg := &mockConfigManager{
			data: map[string]any{
				"lock.driver": "redis",
				"lock.redis_config": map[string]any{
					"host":              "localhost",
					"port":              6379,
					"password":          "redis-pass",
					"db":                0,
					"max_idle_conns":    10,
					"max_open_conns":    100,
					"conn_max_lifetime": 30 * time.Second,
				},
			},
		}

		mgr, err := BuildWithConfigProvider(mockCfg)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.IsType(t, &lockManagerRedisImpl{}, mgr)
	})

	t.Run("成功创建 Memory 驱动", func(t *testing.T) {
		mockCfg := &mockConfigManager{
			data: map[string]any{
				"lock.driver": "memory",
				"lock.memory_config": map[string]any{
					"max_backups": 1000,
				},
			},
		}

		mgr, err := BuildWithConfigProvider(mockCfg)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.IsType(t, &lockManagerMemoryImpl{}, mgr)
	})

	t.Run("配置提供者为 nil 返回错误", func(t *testing.T) {
		mgr, err := BuildWithConfigProvider(nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "configProvider cannot be nil")
	})

	t.Run("获取 lock.driver 失败返回错误", func(t *testing.T) {
		mockCfg := &mockConfigManager{
			data: map[string]any{},
			err:  fmt.Errorf("get error"),
		}

		mgr, err := BuildWithConfigProvider(mockCfg)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "failed to get lock.driver")
	})

	t.Run("lock.driver 类型错误返回错误", func(t *testing.T) {
		mockCfg := &mockConfigManager{
			data: map[string]any{
				"lock.driver": 12345,
			},
		}

		mgr, err := BuildWithConfigProvider(mockCfg)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "lock.driver")
	})

	t.Run("获取 Redis 配置失败返回错误", func(t *testing.T) {
		mockCfg := &mockConfigManager{
			data: map[string]any{
				"lock.driver": "redis",
			},
		}

		mgr, err := BuildWithConfigProvider(mockCfg)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "failed to get lock.redis_config")
	})

	t.Run("lock.redis_config 类型错误返回错误", func(t *testing.T) {
		mockCfg := &mockConfigManager{
			data: map[string]any{
				"lock.driver":       "redis",
				"lock.redis_config": "invalid-config",
			},
		}

		mgr, err := BuildWithConfigProvider(mockCfg)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "lock.redis_config")
	})

	t.Run("获取 Memory 配置失败返回错误", func(t *testing.T) {
		mockCfg := &mockConfigManager{
			data: map[string]any{
				"lock.driver": "memory",
			},
		}

		mgr, err := BuildWithConfigProvider(mockCfg)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "failed to get lock.memory_config")
	})

	t.Run("lock.memory_config 类型错误返回错误", func(t *testing.T) {
		mockCfg := &mockConfigManager{
			data: map[string]any{
				"lock.driver":        "memory",
				"lock.memory_config": "invalid-config",
			},
		}

		mgr, err := BuildWithConfigProvider(mockCfg)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "lock.memory_config")
	})

	t.Run("不支持的驱动类型返回错误", func(t *testing.T) {
		mockCfg := &mockConfigManager{
			data: map[string]any{
				"lock.driver": "unsupported-driver",
			},
		}

		mgr, err := BuildWithConfigProvider(mockCfg)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "unsupported driver type")
	})

	t.Run("完整的 Redis 配置创建成功", func(t *testing.T) {
		mockCfg := &mockConfigManager{
			data: map[string]any{
				"lock.driver": "redis",
				"lock.redis_config": map[string]any{
					"host":              "192.168.1.100",
					"port":              6380,
					"password":          "secure-password",
					"db":                2,
					"max_idle_conns":    20,
					"max_open_conns":    200,
					"conn_max_lifetime": 60 * time.Second,
				},
			},
		}

		mgr, err := BuildWithConfigProvider(mockCfg)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.IsType(t, &lockManagerRedisImpl{}, mgr)
	})

	t.Run("完整的 Memory 配置创建成功", func(t *testing.T) {
		mockCfg := &mockConfigManager{
			data: map[string]any{
				"lock.driver": "memory",
				"lock.memory_config": map[string]any{
					"max_backups": 5000,
				},
			},
		}

		mgr, err := BuildWithConfigProvider(mockCfg)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.IsType(t, &lockManagerMemoryImpl{}, mgr)
	})
}
