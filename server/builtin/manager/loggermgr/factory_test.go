package loggermgr

import (
	"testing"

	"github.com/lite-lake/litecore-go/server/builtin/manager/telemetrymgr"
	"github.com/stretchr/testify/assert"
)

type mockConfigProvider struct {
	data map[string]any
}

func (m *mockConfigProvider) Get(key string) (any, error) {
	return m.data[key], nil
}

func (m *mockConfigProvider) Has(key string) bool {
	_, ok := m.data[key]
	return ok
}

func TestBuildWithConfigProvider(t *testing.T) {
	t.Run("zap_driver_with_config", func(t *testing.T) {
		provider := &mockConfigProvider{
			data: map[string]any{
				"logger.driver": "zap",
				"logger.zap_config": map[string]any{
					"console_enabled": true,
					"console_config": map[string]any{
						"level": "info",
					},
				},
			},
		}

		mgr, err := BuildWithConfigProvider(provider, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.Equal(t, "LoggerZapManager", mgr.ManagerName())
	})

	t.Run("zap_driver_with_file_config", func(t *testing.T) {
		provider := &mockConfigProvider{
			data: map[string]any{
				"logger.driver": "zap",
				"logger.zap_config": map[string]any{
					"file_enabled": true,
					"file_config": map[string]any{
						"level": "debug",
						"path":  "/tmp/test.log",
						"rotation": map[string]any{
							"max_size":    100,
							"max_age":     30,
							"max_backups": 10,
							"compress":    true,
						},
					},
				},
			},
		}

		mgr, err := BuildWithConfigProvider(provider, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.Equal(t, "LoggerZapManager", mgr.ManagerName())
	})

	t.Run("zap_driver_with_telemetry", func(t *testing.T) {
		provider := &mockConfigProvider{
			data: map[string]any{
				"logger.driver": "zap",
				"logger.zap_config": map[string]any{
					"telemetry_enabled": true,
					"telemetry_config": map[string]any{
						"level": "info",
					},
					"console_enabled": true,
					"console_config": map[string]any{
						"level": "info",
					},
				},
			},
		}

		telemetryMgr := telemetrymgr.NewTelemetryManagerNoneImpl()
		mgr, err := BuildWithConfigProvider(provider, telemetryMgr)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.Equal(t, "LoggerZapManager", mgr.ManagerName())
	})

	t.Run("default_driver", func(t *testing.T) {
		provider := &mockConfigProvider{
			data: map[string]any{
				"logger.driver": "default",
			},
		}

		mgr, err := BuildWithConfigProvider(provider, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.Equal(t, "LoggerDefaultManager", mgr.ManagerName())
	})

	t.Run("none_driver", func(t *testing.T) {
		provider := &mockConfigProvider{
			data: map[string]any{
				"logger.driver": "none",
			},
		}

		mgr, err := BuildWithConfigProvider(provider, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.Equal(t, "LoggerNoneManager", mgr.ManagerName())
	})

	t.Run("unknown_driver", func(t *testing.T) {
		provider := &mockConfigProvider{
			data: map[string]any{
				"logger.driver": "unknown",
			},
		}

		mgr, err := BuildWithConfigProvider(provider, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "unsupported driver type")
	})

	t.Run("nil_config_provider", func(t *testing.T) {
		mgr, err := BuildWithConfigProvider(nil, nil)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "configProvider cannot be nil")
	})

	t.Run("zap_driver_with_case_insensitive", func(t *testing.T) {
		provider := &mockConfigProvider{
			data: map[string]any{
				"logger.driver": "ZAP",
				"logger.zap_config": map[string]any{
					"console_enabled": true,
					"console_config": map[string]any{
						"level": "info",
					},
				},
			},
		}

		mgr, err := BuildWithConfigProvider(provider, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.Equal(t, "LoggerZapManager", mgr.ManagerName())
	})

	t.Run("zap_driver_with_whitespace", func(t *testing.T) {
		provider := &mockConfigProvider{
			data: map[string]any{
				"logger.driver": " zap ",
				"logger.zap_config": map[string]any{
					"console_enabled": true,
					"console_config": map[string]any{
						"level": "info",
					},
				},
			},
		}

		mgr, err := BuildWithConfigProvider(provider, nil)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.Equal(t, "LoggerZapManager", mgr.ManagerName())
	})
}
