package config

import (
	"testing"
	"time"
)

func TestDatabaseConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *DatabaseConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid none driver",
			config: &DatabaseConfig{
				Driver: "none",
			},
			wantErr: false,
		},
		{
			name: "valid sqlite driver",
			config: &DatabaseConfig{
				Driver: "sqlite",
				SQLiteConfig: &SQLiteConfig{
					DSN: "file:./test.db",
				},
			},
			wantErr: false,
		},
		{
			name: "valid postgresql driver",
			config: &DatabaseConfig{
				Driver: "postgresql",
				PostgreSQLConfig: &PostgreSQLConfig{
					DSN: "host=localhost port=5432 user=postgres dbname=test",
				},
			},
			wantErr: false,
		},
		{
			name: "valid mysql driver",
			config: &DatabaseConfig{
				Driver: "mysql",
				MySQLConfig: &MySQLConfig{
					DSN: "root:password@tcp(localhost:3306)/test",
				},
			},
			wantErr: false,
		},
		{
			name:    "missing driver",
			config:  &DatabaseConfig{},
			wantErr: true,
			errMsg:  "driver is required",
		},
		{
			name: "invalid driver",
			config: &DatabaseConfig{
				Driver: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid driver",
		},
		{
			name: "sqlite driver without config",
			config: &DatabaseConfig{
				Driver: "sqlite",
			},
			wantErr: true,
			errMsg:  "sqlite_config is required",
		},
		{
			name: "postgresql driver without config",
			config: &DatabaseConfig{
				Driver: "postgresql",
			},
			wantErr: true,
			errMsg:  "postgresql_config is required",
		},
		{
			name: "mysql driver without config",
			config: &DatabaseConfig{
				Driver: "mysql",
			},
			wantErr: true,
			errMsg:  "mysql_config is required",
		},
		{
			name: "sqlite with empty DSN",
			config: &DatabaseConfig{
				Driver: "sqlite",
				SQLiteConfig: &SQLiteConfig{
					DSN: "",
				},
			},
			wantErr: true,
			errMsg:  "invalid sqlite_config",
		},
		{
			name: "postgresql with empty DSN",
			config: &DatabaseConfig{
				Driver: "postgresql",
				PostgreSQLConfig: &PostgreSQLConfig{
					DSN: "",
				},
			},
			wantErr: true,
			errMsg:  "invalid postgresql_config",
		},
		{
			name: "mysql with empty DSN",
			config: &DatabaseConfig{
				Driver: "mysql",
				MySQLConfig: &MySQLConfig{
					DSN: "",
				},
			},
			wantErr: true,
			errMsg:  "invalid mysql_config",
		},
		{
			name: "invalid pool config - negative max_open_conns",
			config: &DatabaseConfig{
				Driver: "sqlite",
				SQLiteConfig: &SQLiteConfig{
					DSN: "file:./test.db",
					PoolConfig: &PoolConfig{
						MaxOpenConns: -1,
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid sqlite_config",
		},
		{
			name: "invalid pool config - negative max_idle_conns",
			config: &DatabaseConfig{
				Driver: "sqlite",
				SQLiteConfig: &SQLiteConfig{
					DSN: "file:./test.db",
					PoolConfig: &PoolConfig{
						MaxIdleConns: -1,
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid sqlite_config",
		},
		{
			name: "invalid pool config - max_idle_conns > max_open_conns",
			config: &DatabaseConfig{
				Driver: "sqlite",
				SQLiteConfig: &SQLiteConfig{
					DSN: "file:./test.db",
					PoolConfig: &PoolConfig{
						MaxOpenConns: 5,
						MaxIdleConns: 10,
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid sqlite_config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if got := err.Error(); got[:len(tt.errMsg)] != tt.errMsg {
					t.Errorf("DatabaseConfig.Validate() error = %v, want prefix %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestPoolConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *PoolConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid pool config",
			config: &PoolConfig{
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 30 * time.Second,
				ConnMaxIdleTime: 5 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "valid pool config with zero max_open_conns",
			config: &PoolConfig{
				MaxOpenConns:    0,
				MaxIdleConns:    0,
				ConnMaxLifetime: 30 * time.Second,
				ConnMaxIdleTime: 5 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "negative max_open_conns",
			config: &PoolConfig{
				MaxOpenConns: -1,
			},
			wantErr: true,
			errMsg:  "max_open_conns must be >= 0",
		},
		{
			name: "negative max_idle_conns",
			config: &PoolConfig{
				MaxIdleConns: -1,
			},
			wantErr: true,
			errMsg:  "max_idle_conns must be >= 0",
		},
		{
			name: "max_idle_conns > max_open_conns",
			config: &PoolConfig{
				MaxOpenConns: 5,
				MaxIdleConns: 10,
			},
			wantErr: true,
			errMsg:  "max_idle_conns must be <= max_open_conns",
		},
		{
			name: "negative conn_max_lifetime",
			config: &PoolConfig{
				ConnMaxLifetime: -1 * time.Second,
			},
			wantErr: true,
			errMsg:  "conn_max_lifetime must be >= 0",
		},
		{
			name: "negative conn_max_idle_time",
			config: &PoolConfig{
				ConnMaxIdleTime: -1 * time.Second,
			},
			wantErr: true,
			errMsg:  "conn_max_idle_time must be >= 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PoolConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if got := err.Error(); got[:len(tt.errMsg)] != tt.errMsg {
					t.Errorf("PoolConfig.Validate() error = %v, want prefix %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestParseDatabaseConfigFromMap(t *testing.T) {
	tests := []struct {
		name    string
		cfg     map[string]any
		want    *DatabaseConfig
		wantErr bool
	}{
		{
			name: "nil config",
			cfg:  nil,
			want: &DatabaseConfig{
				Driver: "none",
			},
			wantErr: false,
		},
		{
			name: "empty config",
			cfg:  map[string]any{},
			want: &DatabaseConfig{
				Driver: "none",
			},
			wantErr: false,
		},
		{
			name: "sqlite config",
			cfg: map[string]any{
				"driver": "sqlite",
				"sqlite_config": map[string]any{
					"dsn": "file:./test.db",
				},
			},
			want: &DatabaseConfig{
				Driver: "sqlite",
				SQLiteConfig: &SQLiteConfig{
					DSN: "file:./test.db",
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
			name: "postgresql config",
			cfg: map[string]any{
				"driver": "postgresql",
				"postgresql_config": map[string]any{
					"dsn": "host=localhost port=5432 user=postgres dbname=test",
				},
			},
			want: &DatabaseConfig{
				Driver: "postgresql",
				PostgreSQLConfig: &PostgreSQLConfig{
					DSN: "host=localhost port=5432 user=postgres dbname=test",
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
		{
			name: "mysql config",
			cfg: map[string]any{
				"driver": "mysql",
				"mysql_config": map[string]any{
					"dsn": "root:password@tcp(localhost:3306)/test",
				},
			},
			want: &DatabaseConfig{
				Driver: "mysql",
				MySQLConfig: &MySQLConfig{
					DSN: "root:password@tcp(localhost:3306)/test",
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
		{
			name: "sqlite config with pool config",
			cfg: map[string]any{
				"driver": "sqlite",
				"sqlite_config": map[string]any{
					"dsn": "file:./test.db",
					"pool_config": map[string]any{
						"max_open_conns":    5,
						"max_idle_conns":    2,
						"conn_max_lifetime": 60,
						"conn_max_idle_time": 300,
					},
				},
			},
			want: &DatabaseConfig{
				Driver: "sqlite",
				SQLiteConfig: &SQLiteConfig{
					DSN: "file:./test.db",
					PoolConfig: &PoolConfig{
						MaxOpenConns:    5,
						MaxIdleConns:    2,
						ConnMaxLifetime: 60 * time.Second,
						ConnMaxIdleTime: 300 * time.Second,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "pool config with duration string",
			cfg: map[string]any{
				"driver": "mysql",
				"mysql_config": map[string]any{
					"dsn": "root:password@tcp(localhost:3306)/test",
					"pool_config": map[string]any{
						"max_open_conns":    20,
						"max_idle_conns":    10,
						"conn_max_lifetime": "1m",
						"conn_max_idle_time": "10m",
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
						ConnMaxLifetime: 1 * time.Minute,
						ConnMaxIdleTime: 10 * time.Minute,
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
					t.Errorf("ParseDatabaseConfigFromMap() driver = %v, want %v", got.Driver, tt.want.Driver)
				}
				if got.SQLiteConfig != nil && tt.want.SQLiteConfig != nil {
					if got.SQLiteConfig.DSN != tt.want.SQLiteConfig.DSN {
						t.Errorf("ParseDatabaseConfigFromMap() sqlite DSN = %v, want %v", got.SQLiteConfig.DSN, tt.want.SQLiteConfig.DSN)
					}
				}
				if got.PostgreSQLConfig != nil && tt.want.PostgreSQLConfig != nil {
					if got.PostgreSQLConfig.DSN != tt.want.PostgreSQLConfig.DSN {
						t.Errorf("ParseDatabaseConfigFromMap() postgresql DSN = %v, want %v", got.PostgreSQLConfig.DSN, tt.want.PostgreSQLConfig.DSN)
					}
				}
				if got.MySQLConfig != nil && tt.want.MySQLConfig != nil {
					if got.MySQLConfig.DSN != tt.want.MySQLConfig.DSN {
						t.Errorf("ParseDatabaseConfigFromMap() mysql DSN = %v, want %v", got.MySQLConfig.DSN, tt.want.MySQLConfig.DSN)
					}
				}
			}
		})
	}
}

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
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			if got := isValidDriver(tt.driver); got != tt.want {
				t.Errorf("isValidDriver() = %v, want %v", got, tt.want)
			}
		})
	}
}