package config

import (
	"testing"
	"time"
)

// TestDatabaseConfig_Validate 测试 DatabaseConfig 验证
func TestDatabaseConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *DatabaseConfig
		wantErr bool
	}{
		{
			name: "空驱动",
			config: &DatabaseConfig{
				Driver: "",
			},
			wantErr: true,
		},
		{
			name: "无效驱动",
			config: &DatabaseConfig{
				Driver: "invalid",
			},
			wantErr: true,
		},
		{
			name: "SQLite 驱动无配置",
			config: &DatabaseConfig{
				Driver: "sqlite",
			},
			wantErr: true,
		},
		{
			name: "SQLite 驱动有效配置",
			config: &DatabaseConfig{
				Driver: "sqlite",
				SQLiteConfig: &SQLiteConfig{
					DSN: "file::memory:?cache=shared",
				},
			},
			wantErr: false,
		},
		{
			name: "PostgreSQL 驱动无配置",
			config: &DatabaseConfig{
				Driver: "postgresql",
			},
			wantErr: true,
		},
		{
			name: "PostgreSQL 驱动有效配置",
			config: &DatabaseConfig{
				Driver: "postgresql",
				PostgreSQLConfig: &PostgreSQLConfig{
					DSN: "host=localhost port=5432",
				},
			},
			wantErr: false,
		},
		{
			name: "MySQL 驱动无配置",
			config: &DatabaseConfig{
				Driver: "mysql",
			},
			wantErr: true,
		},
		{
			name: "MySQL 驱动有效配置",
			config: &DatabaseConfig{
				Driver: "mysql",
				MySQLConfig: &MySQLConfig{
					DSN: "root:password@tcp(localhost:3306)/test",
				},
			},
			wantErr: false,
		},
		{
			name: "none 驱动不需要配置",
			config: &DatabaseConfig{
				Driver: "none",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("DatabaseConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSQLiteConfig_Validate 测试 SQLiteConfig 验证
func TestSQLiteConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *SQLiteConfig
		wantErr bool
	}{
		{
			name:    "空 DSN",
			config:  &SQLiteConfig{},
			wantErr: true,
		},
		{
			name: "有效 DSN",
			config: &SQLiteConfig{
				DSN: "file::memory:?cache=shared",
			},
			wantErr: false,
		},
		{
			name: "有效 DSN 带连接池配置",
			config: &SQLiteConfig{
				DSN: "file:./test.db",
				PoolConfig: &PoolConfig{
					MaxOpenConns:    1,
					MaxIdleConns:    1,
					ConnMaxLifetime: time.Hour,
					ConnMaxIdleTime: 5 * time.Minute,
				},
			},
			wantErr: false,
		},
		{
			name: "无效连接池配置",
			config: &SQLiteConfig{
				DSN: "file:./test.db",
				PoolConfig: &PoolConfig{
					MaxOpenConns: -1,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("SQLiteConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestPostgreSQLConfig_Validate 测试 PostgreSQLConfig 验证
func TestPostgreSQLConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *PostgreSQLConfig
		wantErr bool
	}{
		{
			name:    "空 DSN",
			config:  &PostgreSQLConfig{},
			wantErr: true,
		},
		{
			name: "有效 DSN",
			config: &PostgreSQLConfig{
				DSN: "host=localhost port=5432",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("PostgreSQLConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestMySQLConfig_Validate 测试 MySQLConfig 验证
func TestMySQLConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *MySQLConfig
		wantErr bool
	}{
		{
			name:    "空 DSN",
			config:  &MySQLConfig{},
			wantErr: true,
		},
		{
			name: "有效 DSN",
			config: &MySQLConfig{
				DSN: "root:password@tcp(localhost:3306)/test",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("MySQLConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestPoolConfig_Validate 测试 PoolConfig 验证
func TestPoolConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *PoolConfig
		wantErr bool
	}{
		{
			name: "有效配置",
			config: &PoolConfig{
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: time.Hour,
				ConnMaxIdleTime: 5 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "MaxOpenConns 为负数",
			config: &PoolConfig{
				MaxOpenConns: -1,
			},
			wantErr: true,
		},
		{
			name: "MaxIdleConns 为负数",
			config: &PoolConfig{
				MaxIdleConns: -1,
			},
			wantErr: true,
		},
		{
			name: "MaxIdleConns 大于 MaxOpenConns",
			config: &PoolConfig{
				MaxOpenConns: 5,
				MaxIdleConns: 10,
			},
			wantErr: true,
		},
		{
			name: "ConnMaxLifetime 为负数",
			config: &PoolConfig{
				ConnMaxLifetime: -1,
			},
			wantErr: true,
		},
		{
			name: "ConnMaxIdleTime 为负数",
			config: &PoolConfig{
				ConnMaxIdleTime: -1,
			},
			wantErr: true,
		},
		{
			name: "MaxOpenConns 为 0（无限制）",
			config: &PoolConfig{
				MaxOpenConns: 0,
				MaxIdleConns: 5,
			},
			wantErr: false,
		},
		{
			name: "零值配置（全部使用默认值）",
			config: &PoolConfig{
				MaxOpenConns:    0,
				MaxIdleConns:    0,
				ConnMaxLifetime: 0,
				ConnMaxIdleTime: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
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
		want    *DatabaseConfig
		wantErr bool
	}{
		{
			name:    "空配置",
			cfg:     nil,
			want:    &DatabaseConfig{Driver: "none"},
			wantErr: false,
		},
		{
			name:    "空 map",
			cfg:     map[string]any{},
			want:    &DatabaseConfig{Driver: "none"},
			wantErr: false,
		},
		{
			name: "SQLite 配置",
			cfg: map[string]any{
				"driver": "sqlite",
				"sqlite_config": map[string]any{
					"dsn": "file::memory:?cache=shared",
				},
			},
			want: &DatabaseConfig{
				Driver: "sqlite",
				SQLiteConfig: &SQLiteConfig{
					DSN: "file::memory:?cache=shared",
					PoolConfig: &PoolConfig{
						MaxOpenConns:    1,
						MaxIdleConns:    1,
						ConnMaxLifetime: DefaultConnMaxLifetime,
						ConnMaxIdleTime: DefaultConnMaxIdleTime,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "MySQL 配置带连接池",
			cfg: map[string]any{
				"driver": "mysql",
				"mysql_config": map[string]any{
					"dsn": "root:password@tcp(localhost:3306)/test",
					"pool_config": map[string]any{
						"max_open_conns":    20,
						"max_idle_conns":    10,
						"conn_max_lifetime": 60,
						"conn_max_idle_time": "5m",
					},
				},
			},
			want: &DatabaseConfig{
				Driver: "mysql",
				MySQLConfig: &MySQLConfig{
					DSN: "root:password@tcp(localhost:3306)/test",
					PoolConfig: &PoolConfig{
						MaxOpenConns:    20,
						MaxIdleConns:    10,
						ConnMaxLifetime: 60 * time.Second,
						ConnMaxIdleTime: 5 * time.Minute,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "PostgreSQL 配置",
			cfg: map[string]any{
				"driver": "postgresql",
				"postgresql_config": map[string]any{
					"dsn": "host=localhost port=5432",
				},
			},
			want: &DatabaseConfig{
				Driver: "postgresql",
				PostgreSQLConfig: &PostgreSQLConfig{
					DSN: "host=localhost port=5432",
					PoolConfig: &PoolConfig{
						MaxOpenConns:    DefaultMaxOpenConns,
						MaxIdleConns:    DefaultMaxIdleConns,
						ConnMaxLifetime: DefaultConnMaxLifetime,
						ConnMaxIdleTime: DefaultConnMaxIdleTime,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDatabaseConfigFromMap(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDatabaseConfigFromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Driver != tt.want.Driver {
					t.Errorf("ParseDatabaseConfigFromMap() Driver = %v, want %v", got.Driver, tt.want.Driver)
				}
				// 简单验证子配置是否存在
				switch tt.want.Driver {
				case "sqlite":
					if got.SQLiteConfig == nil || got.SQLiteConfig.DSN != tt.want.SQLiteConfig.DSN {
						t.Errorf("ParseDatabaseConfigFromMap() SQLiteConfig mismatch")
					}
				case "mysql":
					if got.MySQLConfig == nil || got.MySQLConfig.DSN != tt.want.MySQLConfig.DSN {
						t.Errorf("ParseDatabaseConfigFromMap() MySQLConfig mismatch")
					}
				case "postgresql":
					if got.PostgreSQLConfig == nil || got.PostgreSQLConfig.DSN != tt.want.PostgreSQLConfig.DSN {
						t.Errorf("ParseDatabaseConfigFromMap() PostgreSQLConfig mismatch")
					}
				}
			}
		})
	}
}

// TestParsePoolConfig 测试连接池配置解析
func TestParsePoolConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     map[string]any
		want    *PoolConfig
		wantErr bool
	}{
		{
			name: "默认配置",
			cfg:  map[string]any{},
			want: &PoolConfig{
				MaxOpenConns:    DefaultMaxOpenConns,
				MaxIdleConns:    DefaultMaxIdleConns,
				ConnMaxLifetime: DefaultConnMaxLifetime,
				ConnMaxIdleTime: DefaultConnMaxIdleTime,
			},
			wantErr: false,
		},
		{
			name: "整数参数",
			cfg: map[string]any{
				"max_open_conns":    20,
				"max_idle_conns":    10,
				"conn_max_lifetime": 60,
				"conn_max_idle_time": 300,
			},
			want: &PoolConfig{
				MaxOpenConns:    20,
				MaxIdleConns:    10,
				ConnMaxLifetime: 60 * time.Second,
				ConnMaxIdleTime: 300 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "浮点数参数",
			cfg: map[string]any{
				"max_open_conns": 20.0,
				"max_idle_conns": 10.0,
			},
			want: &PoolConfig{
				MaxOpenConns:    20,
				MaxIdleConns:    10,
				ConnMaxLifetime: DefaultConnMaxLifetime,
				ConnMaxIdleTime: DefaultConnMaxIdleTime,
			},
			wantErr: false,
		},
		{
			name: "时间间隔字符串",
			cfg: map[string]any{
				"conn_max_lifetime": "1h30m",
				"conn_max_idle_time": "5m",
			},
			want: &PoolConfig{
				MaxOpenConns:    DefaultMaxOpenConns,
				MaxIdleConns:    DefaultMaxIdleConns,
				ConnMaxLifetime: 90 * time.Minute,
				ConnMaxIdleTime: 5 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "无效时间格式",
			cfg: map[string]any{
				"conn_max_lifetime": "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePoolConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePoolConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.MaxOpenConns != tt.want.MaxOpenConns {
					t.Errorf("parsePoolConfig() MaxOpenConns = %v, want %v", got.MaxOpenConns, tt.want.MaxOpenConns)
				}
				if got.MaxIdleConns != tt.want.MaxIdleConns {
					t.Errorf("parsePoolConfig() MaxIdleConns = %v, want %v", got.MaxIdleConns, tt.want.MaxIdleConns)
				}
				if got.ConnMaxLifetime != tt.want.ConnMaxLifetime {
					t.Errorf("parsePoolConfig() ConnMaxLifetime = %v, want %v", got.ConnMaxLifetime, tt.want.ConnMaxLifetime)
				}
				if got.ConnMaxIdleTime != tt.want.ConnMaxIdleTime {
					t.Errorf("parsePoolConfig() ConnMaxIdleTime = %v, want %v", got.ConnMaxIdleTime, tt.want.ConnMaxIdleTime)
				}
			}
		})
	}
}

// TestIsValidDriver 测试驱动验证
func TestIsValidDriver(t *testing.T) {
	tests := []struct {
		driver string
		want   bool
	}{
		{"mysql", true},
		{"postgresql", true},
		{"sqlite", true},
		{"none", true},
		{"invalid", false},
		{"", false},
		{"MYSQL", false}, // 大小写敏感
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			if got := isValidDriver(tt.driver); got != tt.want {
				t.Errorf("isValidDriver() = %v, want %v", got, tt.want)
			}
		})
	}
}
