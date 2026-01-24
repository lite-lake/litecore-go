package databasemgr

import (
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestMySQLImplementation_Coverage(t *testing.T) {
	t.Run("NewDatabaseManagerMySQLImpl with nil config", func(t *testing.T) {
		_, err := NewDatabaseManagerMySQLImpl(nil, nil, nil)
		if err == nil {
			t.Error("Expected error for nil config")
		}
	})

	t.Run("NewDatabaseManagerMySQLImpl with empty DSN", func(t *testing.T) {
		cfg := &MySQLConfig{
			DSN: "",
		}
		_, err := NewDatabaseManagerMySQLImpl(cfg, nil, nil)
		if err == nil {
			t.Error("Expected error for empty DSN")
		}
	})

	t.Run("NewDatabaseManagerMySQLImpl with invalid DSN", func(t *testing.T) {
		cfg := &MySQLConfig{
			DSN: "invalid-dsn",
		}
		_, err := NewDatabaseManagerMySQLImpl(cfg, nil, nil)
		if err == nil {
			t.Error("Expected error for invalid DSN")
		}
	})

	t.Run("NewDatabaseManagerMySQLImpl with unreachable host", func(t *testing.T) {
		cfg := &MySQLConfig{
			DSN: "root:password@tcp(unreachable-host:3306)/test",
		}
		_, err := NewDatabaseManagerMySQLImpl(cfg, nil, nil)
		if err == nil {
			t.Error("Expected error for unreachable host")
		}
	})

	t.Run("NewDatabaseManagerMySQLImpl with valid config", func(t *testing.T) {
		cfg := &MySQLConfig{
			DSN: "root:password@tcp(localhost:3306)/test?parseTime=True",
		}
		_, err := NewDatabaseManagerMySQLImpl(cfg, nil, nil)
		if err == nil {
			t.Error("Expected error for non-existent database")
		}
	})
}

func TestPostgreSQLImplementation_Coverage(t *testing.T) {
	t.Run("NewDatabaseManagerPostgreSQLImpl with nil config", func(t *testing.T) {
		_, err := NewDatabaseManagerPostgreSQLImpl(nil, nil, nil)
		if err == nil {
			t.Error("Expected error for nil config")
		}
	})

	t.Run("NewDatabaseManagerPostgreSQLImpl with empty DSN", func(t *testing.T) {
		cfg := &PostgreSQLConfig{
			DSN: "",
		}
		_, err := NewDatabaseManagerPostgreSQLImpl(cfg, nil, nil)
		if err == nil {
			t.Error("Expected error for empty DSN")
		}
	})

	t.Run("NewDatabaseManagerPostgreSQLImpl with invalid DSN", func(t *testing.T) {
		cfg := &PostgreSQLConfig{
			DSN: "invalid-dsn",
		}
		_, err := NewDatabaseManagerPostgreSQLImpl(cfg, nil, nil)
		if err == nil {
			t.Error("Expected error for invalid DSN")
		}
	})

	t.Run("NewDatabaseManagerPostgreSQLImpl with unreachable host", func(t *testing.T) {
		cfg := &PostgreSQLConfig{
			DSN: "host=unreachable-host port=5432",
		}
		_, err := NewDatabaseManagerPostgreSQLImpl(cfg, nil, nil)
		if err == nil {
			t.Error("Expected error for unreachable host")
		}
	})
}

func TestFactory_Coverage(t *testing.T) {
	t.Run("Build with MySQL and valid config", func(t *testing.T) {
		cfg := map[string]any{
			"dsn": "root:password@tcp(localhost:3306)/test",
		}
		_, err := Build("mysql", cfg, nil, nil)
		if err == nil {
			t.Error("Expected error for non-existent database")
		}
	})

	t.Run("Build with PostgreSQL and valid config", func(t *testing.T) {
		cfg := map[string]any{
			"dsn": "host=localhost port=5432",
		}
		_, err := Build("postgresql", cfg, nil, nil)
		if err == nil {
			t.Error("Expected error for non-existent database")
		}
	})

	t.Run("BuildWithConfigProvider error cases", func(t *testing.T) {
		t.Run("nil provider", func(t *testing.T) {
			_, err := BuildWithConfigProvider(nil, nil, nil)
			if err == nil {
				t.Error("Expected error for nil provider")
			}
		})

		t.Run("missing driver key", func(t *testing.T) {
			provider := &MockConfigProvider{
				data: map[string]any{},
			}
			_, err := BuildWithConfigProvider(provider, nil, nil)
			if err == nil {
				t.Error("Expected error for missing driver")
			}
		})

		t.Run("missing database config for MySQL", func(t *testing.T) {
			provider := &MockConfigProvider{
				data: map[string]any{
					"database.driver": "mysql",
				},
			}
			_, err := BuildWithConfigProvider(provider, nil, nil)
			if err == nil {
				t.Error("Expected error for missing MySQL config")
			}
		})

		t.Run("missing database config for PostgreSQL", func(t *testing.T) {
			provider := &MockConfigProvider{
				data: map[string]any{
					"database.driver": "postgresql",
				},
			}
			_, err := BuildWithConfigProvider(provider, nil, nil)
			if err == nil {
				t.Error("Expected error for missing PostgreSQL config")
			}
		})

		t.Run("missing database config for SQLite", func(t *testing.T) {
			provider := &MockConfigProvider{
				data: map[string]any{
					"database.driver": "sqlite",
				},
			}
			_, err := BuildWithConfigProvider(provider, nil, nil)
			if err == nil {
				t.Error("Expected error for missing SQLite config")
			}
		})
	})
}

func TestConfig_Coverage(t *testing.T) {
	t.Run("ParseDatabaseConfigFromMap with all drivers", func(t *testing.T) {
		tests := []struct {
			name  string
			cfg   map[string]any
			check func(*DatabaseConfig) bool
		}{
			{
				name: "MySQL config",
				cfg: map[string]any{
					"driver": "mysql",
					"mysql_config": map[string]any{
						"dsn": "root:password@tcp(localhost:3306)/test",
					},
				},
				check: func(c *DatabaseConfig) bool {
					return c.Driver == "mysql" && c.MySQLConfig != nil
				},
			},
			{
				name: "PostgreSQL config",
				cfg: map[string]any{
					"driver": "postgresql",
					"postgresql_config": map[string]any{
						"dsn": "host=localhost port=5432",
					},
				},
				check: func(c *DatabaseConfig) bool {
					return c.Driver == "postgresql" && c.PostgreSQLConfig != nil
				},
			},
			{
				name: "SQLite config",
				cfg: map[string]any{
					"driver": "sqlite",
					"sqlite_config": map[string]any{
						"dsn": ":memory:",
					},
				},
				check: func(c *DatabaseConfig) bool {
					return c.Driver == "sqlite" && c.SQLiteConfig != nil
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cfg, err := ParseDatabaseConfigFromMap(tt.cfg)
				if err != nil {
					t.Errorf("ParseDatabaseConfigFromMap() error = %v", err)
				}
				if !tt.check(cfg) {
					t.Error("Config check failed")
				}
			})
		}
	})

	t.Run("ParseObservabilityConfig with edge values", func(t *testing.T) {
		tests := []struct {
			name  string
			cfg   map[string]any
			check func(*ObservabilityConfig) bool
		}{
			{
				name: "minimum values",
				cfg: map[string]any{
					"slow_query_threshold": 0,
					"log_sql":              false,
					"sample_rate":          0.0,
				},
				check: func(c *ObservabilityConfig) bool {
					return c.SlowQueryThreshold == 0 && !c.LogSQL && c.SampleRate == 0.0
				},
			},
			{
				name: "maximum values",
				cfg: map[string]any{
					"slow_query_threshold": "1h",
					"log_sql":              true,
					"sample_rate":          1.0,
				},
				check: func(c *ObservabilityConfig) bool {
					return c.SlowQueryThreshold == 1*time.Hour && c.LogSQL && c.SampleRate == 1.0
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cfg, err := parseObservabilityConfig(tt.cfg)
				if err != nil {
					t.Errorf("parseObservabilityConfig() error = %v", err)
				}
				if !tt.check(cfg) {
					t.Error("Config check failed")
				}
			})
		}
	})

	t.Run("ParsePoolConfig with all data types", func(t *testing.T) {
		tests := []struct {
			name  string
			cfg   map[string]any
			check func(*PoolConfig) bool
		}{
			{
				name: "int values",
				cfg: map[string]any{
					"max_open_conns":     10,
					"max_idle_conns":     5,
					"conn_max_lifetime":  3600,
					"conn_max_idle_time": 300,
				},
				check: func(c *PoolConfig) bool {
					return c.MaxOpenConns == 10 && c.MaxIdleConns == 5
				},
			},
			{
				name: "float64 values",
				cfg: map[string]any{
					"max_open_conns":     float64(10),
					"max_idle_conns":     float64(5),
					"conn_max_lifetime":  float64(3600),
					"conn_max_idle_time": float64(300),
				},
				check: func(c *PoolConfig) bool {
					return c.MaxOpenConns == 10 && c.MaxIdleConns == 5
				},
			},
			{
				name: "string duration",
				cfg: map[string]any{
					"conn_max_lifetime":  "1h",
					"conn_max_idle_time": "5m",
				},
				check: func(c *PoolConfig) bool {
					return c.ConnMaxLifetime == 1*time.Hour && c.ConnMaxIdleTime == 5*time.Minute
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cfg, err := parsePoolConfig(tt.cfg)
				if err != nil {
					t.Errorf("parsePoolConfig() error = %v", err)
				}
				if !tt.check(cfg) {
					t.Error("Config check failed")
				}
			})
		}
	})
}

func TestSQLiteImplementation_AdvancedCoverage(t *testing.T) {
	skipIfCGONotAvailable(t)

	t.Run("NewDatabaseManagerSQLiteImpl with invalid pool config", func(t *testing.T) {
		cfg := &SQLiteConfig{
			DSN: ":memory:",
			PoolConfig: &PoolConfig{
				MaxOpenConns: -1,
			},
		}
		_, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
		if err == nil {
			t.Error("Expected error for invalid pool config")
		}
	})
}

func TestDatabaseManager_LifecycleCoverage(t *testing.T) {
	skipIfCGONotAvailable(t)

	t.Run("OnStart calls Health", func(t *testing.T) {
		cfg := &SQLiteConfig{
			DSN: ":memory:",
		}

		mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
		if err != nil {
			t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
		}

		if err := mgr.OnStart(); err != nil {
			t.Errorf("OnStart() error = %v", err)
		}

		mgr.Close()
	})

	t.Run("OnStop closes connection", func(t *testing.T) {
		cfg := &SQLiteConfig{
			DSN: ":memory:",
		}

		mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
		if err != nil {
			t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
		}

		if err := mgr.OnStop(); err != nil {
			t.Errorf("OnStop() error = %v", err)
		}

		stats := mgr.Stats()
		if stats.MaxOpenConnections != 0 {
			t.Error("Connection should be closed")
		}
	})

	t.Run("Close can be called multiple times", func(t *testing.T) {
		cfg := &SQLiteConfig{
			DSN: ":memory:",
		}

		mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
		if err != nil {
			t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
		}

		if err := mgr.Close(); err != nil {
			t.Errorf("First Close() error = %v", err)
		}

		if err := mgr.Close(); err != nil {
			t.Errorf("Second Close() error = %v", err)
		}
	})
}

func TestDatabaseManager_GORMCoverage(t *testing.T) {
	skipIfCGONotAvailable(t)

	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	type User struct {
		ID   uint
		Name string
	}

	if err := mgr.AutoMigrate(&User{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	t.Run("Model with nil value", func(t *testing.T) {
		result := mgr.Model(nil)
		if result == nil {
			t.Error("Model() should not return nil")
		}
	})

	t.Run("Table with empty name", func(t *testing.T) {
		result := mgr.Table("")
		if result == nil {
			t.Error("Table() should not return nil")
		}
	})

	t.Run("WithContext with nil context", func(t *testing.T) {
		result := mgr.WithContext(nil)
		if result == nil {
			t.Error("WithContext() should not return nil")
		}
	})

	t.Run("AutoMigrate with no models", func(t *testing.T) {
		err := mgr.AutoMigrate()
		if err != nil {
			t.Errorf("AutoMigrate() error = %v", err)
		}
	})
}
