package databasemgr

import (
	"os"
	"testing"
	"time"
)

// TestNewDatabaseManagerPostgreSQLImpl 测试 PostgreSQL 管理器创建
// 这是一个集成测试，需要实际的 PostgreSQL 数据库
// 运行测试前需要设置环境变量:
//
//	POSTGRESQL_DSN - PostgreSQL 连接字符串 (可选，默认使用 test)
func TestNewDatabaseManagerPostgreSQLImpl(t *testing.T) {
	// 检查是否应该运行集成测试
	if shouldSkipPostgreSQLIntegrationTest() {
		t.Skip("skipping PostgreSQL integration test (set POSTGRESQL_DSN or TEST_INTEGRATION to run)")
	}

	dsn := getPostgreSQLDSN()

	cfg := &PostgreSQLConfig{
		DSN: dsn,
		PoolConfig: &PoolConfig{
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 30 * time.Second,
			ConnMaxIdleTime: 5 * time.Minute,
		},
	}

	mgr, err := NewDatabaseManagerPostgreSQLImpl(cfg)
	if err != nil {
		t.Fatalf("NewDatabaseManagerPostgreSQLImpl() error = %v", err)
	}
	defer mgr.Close()

	if mgr.ManagerName() != "databaseManagerPostgresqlImpl" {
		t.Errorf("ManagerName() = %v, want 'databaseManagerPostgresqlImpl'", mgr.ManagerName())
	}

	if mgr.Driver() != "postgresql" {
		t.Errorf("Driver() = %v, want 'postgresql'", mgr.Driver())
	}
}

// TestPostgreSQLImpl_ConfigValidation 测试 PostgreSQL 配置验证
func TestPostgreSQLImpl_ConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *PostgreSQLConfig
		wantErr bool
	}{
		{
			name:    "nil configmgr",
			cfg:     nil,
			wantErr: true,
		},
		{
			name: "empty DSN",
			cfg: &PostgreSQLConfig{
				DSN: "",
			},
			wantErr: true,
		},
		{
			name: "valid DSN",
			cfg: &PostgreSQLConfig{
				DSN: "host=localhost port=5432 user=postgres password=password dbname=test sslmode=disable",
			},
			wantErr: false,
		},
		{
			name: "valid DSN with pool configmgr",
			cfg: &PostgreSQLConfig{
				DSN: "host=localhost port=5432",
				PoolConfig: &PoolConfig{
					MaxOpenConns: 50,
					MaxIdleConns: 10,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 不实际连接，只测试配置验证
			if tt.cfg == nil {
				return
			}
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PostgreSQLConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestPostgreSQLImpl_PoolConfig 测试连接池配置
func TestPostgreSQLImpl_PoolConfig(t *testing.T) {
	if shouldSkipPostgreSQLIntegrationTest() {
		t.Skip("skipping PostgreSQL integration test")
	}

	dsn := getPostgreSQLDSN()

	cfg := &PostgreSQLConfig{
		DSN: dsn,
		PoolConfig: &PoolConfig{
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: 10 * time.Second,
			ConnMaxIdleTime: 3 * time.Minute,
		},
	}

	mgr, err := NewDatabaseManagerPostgreSQLImpl(cfg)
	if err != nil {
		t.Fatalf("NewDatabaseManagerPostgreSQLImpl() error = %v", err)
	}
	defer mgr.Close()

	stats := mgr.Stats()
	if stats.MaxOpenConnections != 5 {
		t.Errorf("Stats().MaxOpenConnections = %v, want 5", stats.MaxOpenConnections)
	}
}

// getPostgreSQLDSN 获取 PostgreSQL DSN
func getPostgreSQLDSN() string {
	if dsn := os.Getenv("POSTGRESQL_DSN"); dsn != "" {
		return dsn
	}
	// 默认使用本地 PostgreSQL 测试数据库
	return "host=localhost port=5432 user=postgres password=password dbname=test sslmode=disable"
}

// shouldSkipPostgreSQLIntegrationTest 检查是否应该跳过集成测试
func shouldSkipPostgreSQLIntegrationTest() bool {
	return os.Getenv("POSTGRESQL_DSN") == "" && os.Getenv("TEST_INTEGRATION") == ""
}

// TestPostgreSQLImpl_ErrorHandling 测试错误处理
func TestPostgreSQLImpl_ErrorHandling(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		wantErr bool
	}{
		{
			name:    "invalid host",
			dsn:     "host=invalid-host port=5432",
			wantErr: true,
		},
		{
			name:    "invalid port",
			dsn:     "host=localhost port=99999",
			wantErr: true,
		},
		{
			name:    "wrong password",
			dsn:     "host=localhost port=5432 user=postgres password=wrongpassword",
			wantErr: true,
		},
		{
			name:    "empty DSN",
			dsn:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &PostgreSQLConfig{DSN: tt.dsn}
			_, err := NewDatabaseManagerPostgreSQLImpl(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDatabaseManagerPostgreSQLImpl() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
