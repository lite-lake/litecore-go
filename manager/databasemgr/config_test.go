package databasemgr

import (
	"testing"
	"time"
)

// TestDefaultConfig 测试默认配置
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Driver != "none" {
		t.Errorf("DefaultConfig().Driver = %v, want 'none'", cfg.Driver)
	}

	if cfg.ObservabilityConfig == nil {
		t.Fatal("DefaultConfig().ObservabilityConfig should not be nil")
	}

	if cfg.ObservabilityConfig.SlowQueryThreshold != 1*time.Second {
		t.Errorf("DefaultConfig().ObservabilityConfig.SlowQueryThreshold = %v, want 1s",
			cfg.ObservabilityConfig.SlowQueryThreshold)
	}

	if cfg.ObservabilityConfig.LogSQL != false {
		t.Errorf("DefaultConfig().ObservabilityConfig.LogSQL = %v, want false", cfg.ObservabilityConfig.LogSQL)
	}

	if cfg.ObservabilityConfig.SampleRate != 1.0 {
		t.Errorf("DefaultConfig().ObservabilityConfig.SampleRate = %v, want 1.0", cfg.ObservabilityConfig.SampleRate)
	}
}

// TestDatabaseConfig_Validate 测试 DatabaseConfig 验证
func TestDatabaseConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *DatabaseConfig
		wantErr bool
	}{
		{
			name:    "empty driver",
			cfg:     &DatabaseConfig{Driver: ""},
			wantErr: true,
		},
		{
			name:    "invalid driver",
			cfg:     &DatabaseConfig{Driver: "invalid"},
			wantErr: true,
		},
		{
			name:    "none driver without configmgr",
			cfg:     &DatabaseConfig{Driver: "none"},
			wantErr: false,
		},
		{
			name: "sqlite driver without configmgr",
			cfg: &DatabaseConfig{
				Driver:       "sqlite",
				SQLiteConfig: nil,
			},
			wantErr: true,
		},
		{
			name: "sqlite driver with valid configmgr",
			cfg: &DatabaseConfig{
				Driver: "sqlite",
				SQLiteConfig: &SQLiteConfig{
					DSN: ":memory:",
				},
			},
			wantErr: false,
		},
		{
			name: "postgresql driver without configmgr",
			cfg: &DatabaseConfig{
				Driver:           "postgresql",
				PostgreSQLConfig: nil,
			},
			wantErr: true,
		},
		{
			name: "postgresql driver with valid configmgr",
			cfg: &DatabaseConfig{
				Driver: "postgresql",
				PostgreSQLConfig: &PostgreSQLConfig{
					DSN: "host=localhost",
				},
			},
			wantErr: false,
		},
		{
			name: "mysql driver without configmgr",
			cfg: &DatabaseConfig{
				Driver:      "mysql",
				MySQLConfig: nil,
			},
			wantErr: true,
		},
		{
			name: "mysql driver with valid configmgr",
			cfg: &DatabaseConfig{
				Driver: "mysql",
				MySQLConfig: &MySQLConfig{
					DSN: "root:password@tcp(localhost:3306)/test",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSQLiteConfig_Validate 测试 SQLiteConfig 验证
func TestSQLiteConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *SQLiteConfig
		wantErr bool
	}{
		{
			name:    "nil configmgr",
			cfg:     &SQLiteConfig{DSN: ""},
			wantErr: true,
		},
		{
			name:    "empty DSN",
			cfg:     &SQLiteConfig{DSN: ""},
			wantErr: true,
		},
		{
			name:    "valid DSN",
			cfg:     &SQLiteConfig{DSN: ":memory:"},
			wantErr: false,
		},
		{
			name: "valid DSN with pool configmgr",
			cfg: &SQLiteConfig{
				DSN: ":memory:",
				PoolConfig: &PoolConfig{
					MaxOpenConns: 1,
					MaxIdleConns: 1,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid pool configmgr",
			cfg: &SQLiteConfig{
				DSN: ":memory:",
				PoolConfig: &PoolConfig{
					MaxOpenConns: -1,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLiteConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestPostgreSQLConfig_Validate 测试 PostgreSQLConfig 验证
func TestPostgreSQLConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *PostgreSQLConfig
		wantErr bool
	}{
		{
			name:    "empty DSN",
			cfg:     &PostgreSQLConfig{DSN: ""},
			wantErr: true,
		},
		{
			name:    "valid DSN",
			cfg:     &PostgreSQLConfig{DSN: "host=localhost"},
			wantErr: false,
		},
		{
			name: "valid DSN with pool configmgr",
			cfg: &PostgreSQLConfig{
				DSN: "host=localhost",
				PoolConfig: &PoolConfig{
					MaxOpenConns: 10,
					MaxIdleConns: 5,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PostgreSQLConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestMySQLConfig_Validate 测试 MySQLConfig 验证
func TestMySQLConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *MySQLConfig
		wantErr bool
	}{
		{
			name:    "empty DSN",
			cfg:     &MySQLConfig{DSN: ""},
			wantErr: true,
		},
		{
			name:    "valid DSN",
			cfg:     &MySQLConfig{DSN: "root:password@tcp(localhost:3306)/test"},
			wantErr: false,
		},
		{
			name: "valid DSN with pool configmgr",
			cfg: &MySQLConfig{
				DSN: "root:password@tcp(localhost:3306)/test",
				PoolConfig: &PoolConfig{
					MaxOpenConns: 100,
					MaxIdleConns: 10,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("MySQLConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestPoolConfig_Validate 测试 PoolConfig 验证
func TestPoolConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *PoolConfig
		wantErr bool
	}{
		{
			name: "valid configmgr",
			cfg: &PoolConfig{
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 30 * time.Second,
				ConnMaxIdleTime: 5 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "negative max_open_conns",
			cfg: &PoolConfig{
				MaxOpenConns: -1,
			},
			wantErr: true,
		},
		{
			name: "negative max_idle_conns",
			cfg: &PoolConfig{
				MaxIdleConns: -1,
			},
			wantErr: true,
		},
		{
			name: "max_idle_conns greater than max_open_conns",
			cfg: &PoolConfig{
				MaxOpenConns: 5,
				MaxIdleConns: 10,
			},
			wantErr: true,
		},
		{
			name: "zero max_open_conns allows any max_idle_conns",
			cfg: &PoolConfig{
				MaxOpenConns: 0,
				MaxIdleConns: 10,
			},
			wantErr: false,
		},
		{
			name: "negative conn_max_lifetime",
			cfg: &PoolConfig{
				ConnMaxLifetime: -1,
			},
			wantErr: true,
		},
		{
			name: "negative conn_max_idle_time",
			cfg: &PoolConfig{
				ConnMaxIdleTime: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PoolConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestParseDatabaseConfigFromMap 测试从 map 解析配置
func TestParseDatabaseConfigFromMap(t *testing.T) {
	tests := []struct {
		name    string
		cfg     map[string]any
		check   func(*DatabaseConfig) bool
		wantErr bool
	}{
		{
			name:    "nil configmgr returns default",
			cfg:     nil,
			check:   func(c *DatabaseConfig) bool { return c.Driver == "none" },
			wantErr: false,
		},
		{
			name: "parse driver",
			cfg: map[string]any{
				"driver": "sqlite",
			},
			check:   func(c *DatabaseConfig) bool { return c.Driver == "sqlite" },
			wantErr: false,
		},
		{
			name: "parse sqlite configmgr",
			cfg: map[string]any{
				"driver": "sqlite",
				"sqlite_config": map[string]any{
					"dsn": ":memory:",
				},
			},
			check: func(c *DatabaseConfig) bool {
				return c.Driver == "sqlite" &&
					c.SQLiteConfig != nil &&
					c.SQLiteConfig.DSN == ":memory:"
			},
			wantErr: false,
		},
		{
			name: "parse mysql configmgr with pool",
			cfg: map[string]any{
				"driver": "mysql",
				"mysql_config": map[string]any{
					"dsn": "root:password@tcp(localhost:3306)/test",
					"pool_config": map[string]any{
						"max_open_conns": 100,
						"max_idle_conns": 10,
					},
				},
			},
			check: func(c *DatabaseConfig) bool {
				return c.Driver == "mysql" &&
					c.MySQLConfig != nil &&
					c.MySQLConfig.DSN == "root:password@tcp(localhost:3306)/test" &&
					c.MySQLConfig.PoolConfig != nil &&
					c.MySQLConfig.PoolConfig.MaxOpenConns == 100
			},
			wantErr: false,
		},
		{
			name: "parse postgresql configmgr",
			cfg: map[string]any{
				"driver": "postgresql",
				"postgresql_config": map[string]any{
					"dsn": "host=localhost",
				},
			},
			check: func(c *DatabaseConfig) bool {
				return c.Driver == "postgresql" &&
					c.PostgreSQLConfig != nil &&
					c.PostgreSQLConfig.DSN == "host=localhost"
			},
			wantErr: false,
		},
		{
			name: "parse observability configmgr",
			cfg: map[string]any{
				"observability_config": map[string]any{
					"slow_query_threshold": "2s",
					"log_sql":              true,
					"sample_rate":          0.5,
				},
			},
			check: func(c *DatabaseConfig) bool {
				return c.ObservabilityConfig != nil &&
					c.ObservabilityConfig.SlowQueryThreshold == 2*time.Second &&
					c.ObservabilityConfig.LogSQL == true &&
					c.ObservabilityConfig.SampleRate == 0.5
			},
			wantErr: false,
		},
		{
			name: "invalid sample rate",
			cfg: map[string]any{
				"observability_config": map[string]any{
					"sample_rate": 1.5,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid duration format",
			cfg: map[string]any{
				"observability_config": map[string]any{
					"slow_query_threshold": "invalid",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := ParseDatabaseConfigFromMap(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDatabaseConfigFromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.check(cfg) {
				t.Error("ParseDatabaseConfigFromMap() configmgr check failed")
			}
		})
	}
}

// TestParsePoolConfig 测试解析连接池配置
func TestParsePoolConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     map[string]any
		check   func(*PoolConfig) bool
		wantErr bool
	}{
		{
			name:    "empty configmgr uses defaults",
			cfg:     map[string]any{},
			check:   func(c *PoolConfig) bool { return c.MaxOpenConns == DefaultMaxOpenConns },
			wantErr: false,
		},
		{
			name: "parse int values",
			cfg: map[string]any{
				"max_open_conns": 50,
				"max_idle_conns": 25,
			},
			check: func(c *PoolConfig) bool {
				return c.MaxOpenConns == 50 && c.MaxIdleConns == 25
			},
			wantErr: false,
		},
		{
			name: "parse float64 values",
			cfg: map[string]any{
				"max_open_conns": float64(50),
				"max_idle_conns": float64(25),
			},
			check: func(c *PoolConfig) bool {
				return c.MaxOpenConns == 50 && c.MaxIdleConns == 25
			},
			wantErr: false,
		},
		{
			name: "parse duration as int (seconds)",
			cfg: map[string]any{
				"conn_max_lifetime": 3600,
			},
			check: func(c *PoolConfig) bool {
				return c.ConnMaxLifetime == 3600*time.Second
			},
			wantErr: false,
		},
		{
			name: "parse duration as string",
			cfg: map[string]any{
				"conn_max_lifetime": "1h30m",
			},
			check: func(c *PoolConfig) bool {
				return c.ConnMaxLifetime == 90*time.Minute
			},
			wantErr: false,
		},
		{
			name: "invalid duration format",
			cfg: map[string]any{
				"conn_max_lifetime": "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parsePoolConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePoolConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.check(cfg) {
				t.Error("parsePoolConfig() configmgr check failed")
			}
		})
	}
}

// TestIsValidDriver 测试驱动验证
func TestIsValidDriver(t *testing.T) {
	tests := []struct {
		driver string
		valid  bool
	}{
		{"mysql", true},
		{"postgresql", true},
		{"sqlite", true},
		{"none", true},
		{"invalid", false},
		{"", false},
		{"MYSQL", false},  // case sensitive
		{"SQLite", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			if got := isValidDriver(tt.driver); got != tt.valid {
				t.Errorf("isValidDriver(%q) = %v, want %v", tt.driver, got, tt.valid)
			}
		})
	}
}
