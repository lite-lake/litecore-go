package lockmgr

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lite-lake/litecore-go/manager/cachemgr"
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
		mgr, err := Build("", map[string]any{}, nil, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.IsType(t, &lockManagerMemoryImpl{}, mgr)
	})

	t.Run("创建 Redis 驱动且 cache 底座为 nil 被契约自检拦截", func(t *testing.T) {
		redisConfig := map[string]any{
			"host":              "localhost",
			"port":              6379,
			"password":          "test-pass",
			"db":                0,
			"max_idle_conns":    10,
			"max_open_conns":    100,
			"conn_max_lifetime": "30s",
		}
		_, err := Build("redis", redisConfig, nil, nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "lock.driver=redis requires cache.driver=redis")
		assert.Contains(t, err.Error(), "当前 cache 底座=nil")
	})

	t.Run("创建 Memory 驱动", func(t *testing.T) {
		memoryConfig := map[string]any{
			"max_backups": 1000,
		}
		mgr, err := Build("memory", memoryConfig, nil, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.IsType(t, &lockManagerMemoryImpl{}, mgr)
	})

	t.Run("创建的实例实现了 ILockManager 接口", func(t *testing.T) {
		memoryMgr, err := Build("memory", map[string]any{}, nil, nil, nil)
		assert.NoError(t, err)
		assert.Implements(t, (*ILockManager)(nil), memoryMgr)
	})

	t.Run("不支持的驱动类型返回错误", func(t *testing.T) {
		mgr, err := Build("invalid-driver", map[string]any{}, nil, nil, nil)
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
				name:   "Redis 配置但 cache 底座为 nil（契约违反）",
				driver: "redis",
				config: map[string]any{
					"host": "127.0.0.1",
					"port": 6380,
					"db":   1,
				},
				wantErr: "lock.driver=redis requires cache.driver=redis",
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
				mgr, err := Build(tc.driver, tc.config, nil, nil, nil)
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
	t.Run("Redis 驱动且 cache 底座为 nil 被契约自检拦截", func(t *testing.T) {
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

		_, err := BuildWithConfigProvider(mockCfg, nil, nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "lock.driver=redis requires cache.driver=redis")
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

		mgr, err := BuildWithConfigProvider(mockCfg, nil, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.IsType(t, &lockManagerMemoryImpl{}, mgr)
	})

	t.Run("配置提供者为 nil 返回错误", func(t *testing.T) {
		mgr, err := BuildWithConfigProvider(nil, nil, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "configProvider cannot be nil")
	})

	t.Run("获取 lock.driver 失败返回错误", func(t *testing.T) {
		mockCfg := &mockConfigManager{
			data: map[string]any{},
			err:  fmt.Errorf("get error"),
		}

		mgr, err := BuildWithConfigProvider(mockCfg, nil, nil, nil)
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

		mgr, err := BuildWithConfigProvider(mockCfg, nil, nil, nil)
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

		mgr, err := BuildWithConfigProvider(mockCfg, nil, nil, nil)
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

		mgr, err := BuildWithConfigProvider(mockCfg, nil, nil, nil)
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

		mgr, err := BuildWithConfigProvider(mockCfg, nil, nil, nil)
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

		mgr, err := BuildWithConfigProvider(mockCfg, nil, nil, nil)
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

		mgr, err := BuildWithConfigProvider(mockCfg, nil, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "unsupported driver type")
	})

	t.Run("完整的 Redis 配置但 cache 底座为 nil 被契约自检拦截", func(t *testing.T) {
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

		_, err := BuildWithConfigProvider(mockCfg, nil, nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "lock.driver=redis requires cache.driver=redis")
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

		mgr, err := BuildWithConfigProvider(mockCfg, nil, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.IsType(t, &lockManagerMemoryImpl{}, mgr)
	})
}

// TestBuildRedisContractCheck 测试 lock=redis 时的驱动契约自检（内部PRD B组）：
// bazi（cache=none → SetNX 恒失败 → 登录 30s 挂死）与 screenlocker（cache=memory → 锁退化进程内）
// 两类历史事故形态必须在构建期拦截，不再等到运行时哑火。
func TestBuildRedisContractCheck(t *testing.T) {
	redisConfig := map[string]any{
		"host": "localhost",
		"port": 6379,
		"db":   0,
	}

	t.Run("cache 底座为 none（bazi 事故形态）被拦截", func(t *testing.T) {
		cacheMgr := cachemgr.NewCacheManagerNoneImpl(nil, nil)
		mgr, err := Build("redis", redisConfig, nil, nil, cacheMgr)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "lock.driver=redis requires cache.driver=redis")
		assert.Contains(t, err.Error(), "当前 cache 底座=none")
		assert.Contains(t, err.Error(), "liteauth")
	})

	t.Run("cache 底座为 memory（screenlocker 事故形态）被拦截", func(t *testing.T) {
		cacheMgr := cachemgr.NewCacheManagerMemoryImpl(5*time.Minute, time.Minute, nil, nil)
		mgr, err := Build("redis", redisConfig, nil, nil, cacheMgr)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "lock.driver=redis requires cache.driver=redis")
		assert.Contains(t, err.Error(), "当前 cache 底座=memory")
	})

	t.Run("cache 底座为非内置实现（unknown）被拦截", func(t *testing.T) {
		// mockCacheManager 定义于 redis_impl_test.go，属外部自定义实现形态
		mgr, err := Build("redis", redisConfig, nil, nil, newMockCacheManager())
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "当前 cache 底座=unknown")
	})

	t.Run("DriverKind 类型判别正确", func(t *testing.T) {
		assert.Equal(t, "", cachemgr.DriverKind(nil))
		assert.Equal(t, "none", cachemgr.DriverKind(cachemgr.NewCacheManagerNoneImpl(nil, nil)))
		assert.Equal(t, "memory", cachemgr.DriverKind(cachemgr.NewCacheManagerMemoryImpl(5*time.Minute, time.Minute, nil, nil)))
		assert.Equal(t, "unknown", cachemgr.DriverKind(newMockCacheManager()))
	})
}
