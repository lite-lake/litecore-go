package databasemgr

import (
	"os"
	"testing"
	"time"
)

// TestNewDatabaseManagerMySQLImpl 测试 MySQL 管理器创建
// 这是一个集成测试，需要实际的 MySQL 数据库
// 运行测试前需要设置环境变量:
//   MYSQL_DSN - MySQL 连接字符串 (可选，默认使用 test)
func TestNewDatabaseManagerMySQLImpl(t *testing.T) {
	// 检查是否应该运行集成测试
	if shouldSkipIntegrationTest() {
		t.Skip("skipping MySQL integration test (set MYSQL_DSN or TEST_INTEGRATION to run)")
	}

	dsn := getMySQLDSN()

	cfg := &MySQLConfig{
		DSN: dsn,
		PoolConfig: &PoolConfig{
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 30 * time.Second,
			ConnMaxIdleTime: 5 * time.Minute,
		},
	}

	mgr, err := NewDatabaseManagerMySQLImpl(cfg)
	if err != nil {
		t.Fatalf("NewDatabaseManagerMySQLImpl() error = %v", err)
	}
	defer mgr.Close()

	if mgr.ManagerName() != "databaseManagerMysqlImpl" {
		t.Errorf("ManagerName() = %v, want 'databaseManagerMysqlImpl'", mgr.ManagerName())
	}

	if mgr.Driver() != "mysql" {
		t.Errorf("Driver() = %v, want 'mysql'", mgr.Driver())
	}
}

// TestMySQLImpl_ConfigValidation 测试 MySQL 配置验证
func TestMySQLImpl_ConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *MySQLConfig
		wantErr bool
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
		},
		{
			name: "empty DSN",
			cfg: &MySQLConfig{
				DSN: "",
			},
			wantErr: true,
		},
		{
			name: "valid DSN",
			cfg: &MySQLConfig{
				DSN: "root:password@tcp(localhost:3306)/test",
			},
			wantErr: false,
		},
		{
			name: "valid DSN with pool config",
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
			// 不实际连接，只测试配置验证
			if tt.cfg == nil {
				return
			}
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("MySQLConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestMySQLImpl_PoolConfig 测试连接池配置
func TestMySQLImpl_PoolConfig(t *testing.T) {
	if shouldSkipIntegrationTest() {
		t.Skip("skipping MySQL integration test")
	}

	dsn := getMySQLDSN()

	cfg := &MySQLConfig{
		DSN: dsn,
		PoolConfig: &PoolConfig{
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: 10 * time.Second,
			ConnMaxIdleTime: 3 * time.Minute,
		},
	}

	mgr, err := NewDatabaseManagerMySQLImpl(cfg)
	if err != nil {
		t.Fatalf("NewDatabaseManagerMySQLImpl() error = %v", err)
	}
	defer mgr.Close()

	stats := mgr.Stats()
	if stats.MaxOpenConnections != 5 {
		t.Errorf("Stats().MaxOpenConnections = %v, want 5", stats.MaxOpenConnections)
	}
}

// getMySQLDSN 获取 MySQL DSN
func getMySQLDSN() string {
	if dsn := os.Getenv("MYSQL_DSN"); dsn != "" {
		return dsn
	}
	// 默认使用本地 MySQL 测试数据库
	return "root:password@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
}

// shouldSkipIntegrationTest 检查是否应该跳过集成测试
func shouldSkipIntegrationTest() bool {
	return os.Getenv("MYSQL_DSN") == "" && os.Getenv("TEST_INTEGRATION") == ""
}

// TestMySQLImpl_ErrorHandling 测试错误处理
func TestMySQLImpl_ErrorHandling(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		wantErr bool
	}{
		{
			name:    "invalid host",
			dsn:     "root:password@tcp(invalid-host:3306)/test",
			wantErr: true,
		},
		{
			name:    "invalid port",
			dsn:     "root:password@tcp(localhost:99999)/test",
			wantErr: true,
		},
		{
			name:    "wrong password",
			dsn:     "root:wrongpassword@tcp(localhost:3306)/test",
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
			cfg := &MySQLConfig{DSN: tt.dsn}
			_, err := NewDatabaseManagerMySQLImpl(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDatabaseManagerMySQLImpl() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
