package databasemgr

import (
	"testing"

	"com.litelake.litecore/manager/databasemgr/internal/config"
)

func TestNewFactory(t *testing.T) {
	factory := NewFactory()

	if factory == nil {
		t.Fatal("NewFactory() returned nil")
	}
}

func TestFactory_Build(t *testing.T) {
	factory := NewFactory()

	tests := []struct {
		name           string
		driver         string
		config         map[string]any
		expectedDriver string
	}{
		{
			name:   "sqlite driver",
			driver: "sqlite",
			config: map[string]any{
				"sqlite_config": map[string]any{
					"dsn": "file::memory:?cache=shared",
				},
			},
			expectedDriver: "sqlite",
		},
		{
			name:   "mysql driver",
			driver: "mysql",
			config: map[string]any{
				"mysql_config": map[string]any{
					"dsn": "root:password@tcp(localhost:3306)/test",
				},
			},
			expectedDriver: "mysql", // 会降级到 none 驱动，但驱动类型仍然是 mysql
		},
		{
			name:   "postgresql driver",
			driver: "postgresql",
			config: map[string]any{
				"postgresql_config": map[string]any{
					"dsn": "host=localhost port=5432 user=postgres dbname=test",
				},
			},
			expectedDriver: "postgresql", // 会降级到 none 驱动，但驱动类型仍然是 postgresql
		},
		{
			name:           "none driver",
			driver:         "none",
			config:         map[string]any{},
			expectedDriver: "none",
		},
		{
			name:           "invalid config - missing driver config",
			driver:         "sqlite",
			config:         map[string]any{},
			expectedDriver: "none", // 会降级到 none
		},
		{
			name:           "unknown driver",
			driver:         "unknown",
			config:         map[string]any{},
			expectedDriver: "none", // 会降级到 none
		},
		{
			name:   "empty driver",
			driver: "",
			config: map[string]any{
				"driver": "none",
			},
			expectedDriver: "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := factory.Build(tt.driver, tt.config)

			if mgr == nil {
				t.Fatal("Build() returned nil")
			}

			// 转换为 DatabaseManager 接口
			dbMgr, ok := mgr.(DatabaseManager)
			if !ok {
				t.Fatal("Build() should return DatabaseManager")
			}

			// 验证驱动类型
			if got := dbMgr.Driver(); got != tt.expectedDriver {
				t.Errorf("DatabaseManager.Driver() = %v, want %v", got, tt.expectedDriver)
			}

			// 验证 common.Manager 接口
			if got := mgr.ManagerName(); got == "" {
				t.Error("ManagerName() should not be empty")
			}

			// 对于 SQLite 驱动，验证 DB() 不为 nil
			if tt.expectedDriver == "sqlite" {
				if dbMgr.DB() == nil {
					t.Error("DB() should not be nil for sqlite driver")
				}
			}
		})
	}
}

func TestFactory_BuildWithConfig(t *testing.T) {
	factory := NewFactory()

	tests := []struct {
		name           string
		config         *config.DatabaseConfig
		wantErr        bool
		expectedDriver string
	}{
		{
			name: "valid sqlite config",
			config: &config.DatabaseConfig{
				Driver: "sqlite",
				SQLiteConfig: &config.SQLiteConfig{
					DSN: "file::memory:?cache=shared",
				},
			},
			wantErr:        false,
			expectedDriver: "sqlite",
		},
		{
			name: "valid mysql config",
			config: &config.DatabaseConfig{
				Driver: "mysql",
				MySQLConfig: &config.MySQLConfig{
					DSN: "root:password@tcp(localhost:3306)/test",
				},
			},
			wantErr:        false, // 初始化不会失败，只有在实际查询时才会失败
			expectedDriver: "mysql",
		},
		{
			name: "valid postgresql config",
			config: &config.DatabaseConfig{
				Driver: "postgresql",
				PostgreSQLConfig: &config.PostgreSQLConfig{
					DSN: "host=localhost port=5432 user=postgres dbname=test",
				},
			},
			wantErr:        false, // 初始化不会失败，只有在实际查询时才会失败
			expectedDriver: "postgresql",
		},
		{
			name: "valid none config",
			config: &config.DatabaseConfig{
				Driver: "none",
			},
			wantErr:        false,
			expectedDriver: "none",
		},
		{
			name: "invalid config - missing driver",
			config: &config.DatabaseConfig{
				Driver: "",
			},
			wantErr:        true,
			expectedDriver: "",
		},
		{
			name: "invalid config - unknown driver",
			config: &config.DatabaseConfig{
				Driver: "unknown",
			},
			wantErr:        true,
			expectedDriver: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := factory.BuildWithConfig(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("BuildWithConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if mgr == nil {
					t.Fatal("BuildWithConfig() returned nil")
				}

				// 转换为 DatabaseManager 接口
				dbMgr, ok := mgr.(DatabaseManager)
				if !ok {
					t.Fatal("BuildWithConfig() should return DatabaseManager")
				}

				// 验证驱动类型
				if got := dbMgr.Driver(); got != tt.expectedDriver {
					t.Errorf("DatabaseManager.Driver() = %v, want %v", got, tt.expectedDriver)
				}
			}
		})
	}
}

func TestFactory_Build_Degradation(t *testing.T) {
	factory := NewFactory()

	// 测试降级场景
	tests := []struct {
		name   string
		driver string
		config map[string]any
	}{
		{
			name:   "nil config",
			driver: "sqlite",
			config: nil,
		},
		{
			name:   "empty config",
			driver: "sqlite",
			config: map[string]any{},
		},
		{
			name:   "invalid config type",
			driver: "sqlite",
			config: map[string]any{
				"sqlite_config": "invalid",
			},
		},
		{
			name:   "missing DSN",
			driver: "sqlite",
			config: map[string]any{
				"sqlite_config": map[string]any{
					"dsn": "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := factory.Build(tt.driver, tt.config)

			if mgr == nil {
				t.Fatal("Build() returned nil")
			}

			// 应该降级到 none 驱动
			dbMgr, ok := mgr.(DatabaseManager)
			if !ok {
				t.Fatal("Build() should return DatabaseManager")
			}

			if got := dbMgr.Driver(); got != "none" {
				t.Errorf("Expected driver 'none' after degradation, got %v", got)
			}

			// 验证 DB() 返回 nil
			if dbMgr.DB() != nil {
				t.Error("DB() should return nil for none driver")
			}
		})
	}
}
