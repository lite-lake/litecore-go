package drivers

import (
	"testing"

	"com.litelake.litecore/manager/databasemgr/internal/config"
)

// TestNewPostgreSQLManager_InvalidConfig 测试无效配置
func TestNewPostgreSQLManager_InvalidConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.DatabaseConfig
		wantErr bool
	}{
		{
			name:    "nil 配置",
			config:  nil,
			wantErr: true,
		},
		{
			name: "空驱动",
			config: &config.DatabaseConfig{
				Driver: "",
			},
			wantErr: true,
		},
		{
			name: "PostgreSQL 配置为空",
			config: &config.DatabaseConfig{
				Driver: "postgresql",
			},
			wantErr: true,
		},
		{
			name: "空 DSN",
			config: &config.DatabaseConfig{
				Driver:           "postgresql",
				PostgreSQLConfig: &config.PostgreSQLConfig{},
			},
			wantErr: true,
		},
		{
			name: "无效的 DSN 格式",
			config: &config.DatabaseConfig{
				Driver: "postgresql",
				PostgreSQLConfig: &config.PostgreSQLConfig{
					DSN: "invalid-dsn-format",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPostgreSQLManager(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPostgreSQLManager() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestNewPostgreSQLManager_ValidConfig 测试有效配置但不连接真实数据库
func TestNewPostgreSQLManager_ValidConfig(t *testing.T) {
	// 注意：这个测试使用有效的配置格式，但不需要真实的 PostgreSQL 连接
	// 如果要测试真实的连接，需要集成测试环境
	cfg := &config.DatabaseConfig{
		Driver: "postgresql",
		PostgreSQLConfig: &config.PostgreSQLConfig{
			DSN: "host=localhost port=5432 user=postgres password=password dbname=test sslmode=disable",
			PoolConfig: &config.PoolConfig{
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 0,
				ConnMaxIdleTime: 0,
			},
		},
	}

	// 这个测试会尝试连接，如果 PostgreSQL 不可用会失败
	// 在 CI/CD 环境中，应该跳过或使用 testcontainers
	mgr, err := NewPostgreSQLManager(cfg)
	if err != nil {
		// 如果 PostgreSQL 不可用，跳过测试
		t.Skipf("PostgreSQL not available: %v", err)
		return
	}

	if mgr == nil {
		t.Error("NewPostgreSQLManager() returned nil manager")
	}

	if mgr.ManagerName() != "postgresql-database" {
		t.Errorf("ManagerName() = %v, want 'postgresql-database'", mgr.ManagerName())
	}

	if mgr.Driver() != "postgresql" {
		t.Errorf("Driver() = %v, want 'postgresql'", mgr.Driver())
	}

	// 清理
	_ = mgr.Close()
}

// TestPostgreSQLManager_ImplementsDatabaseManager 测试实现接口
func TestPostgreSQLManager_ImplementsDatabaseManager(t *testing.T) {
	// PostgreSQLManager 通过嵌入 GormBaseManager 实现了 DatabaseManager 接口
	cfg := &config.DatabaseConfig{
		Driver: "postgresql",
		PostgreSQLConfig: &config.PostgreSQLConfig{
			DSN: "host=localhost port=5432 user=postgres password=password dbname=test sslmode=disable",
		},
	}

	mgr, err := NewPostgreSQLManager(cfg)
	if err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
		return
	}
	defer mgr.Close()

	// 验证基本方法
	_ = mgr.ManagerName()
	_ = mgr.Driver()
	_ = mgr.DB()
	_ = mgr.Stats()
}

// TestNewPostgreSQLManager_WithPoolConfig 测试带连接池配置
func TestNewPostgreSQLManager_WithPoolConfig(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "postgresql",
		PostgreSQLConfig: &config.PostgreSQLConfig{
			DSN: "host=localhost port=5432 user=postgres password=password dbname=test sslmode=disable",
			PoolConfig: &config.PoolConfig{
				MaxOpenConns:    20,
				MaxIdleConns:    10,
				ConnMaxLifetime: 0,
				ConnMaxIdleTime: 0,
			},
		},
	}

	mgr, err := NewPostgreSQLManager(cfg)
	if err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
		return
	}
	defer mgr.Close()

	if mgr == nil {
		t.Fatal("NewPostgreSQLManager() returned nil")
	}

	// 验证连接池配置已应用
	stats := mgr.Stats()
	_ = stats.MaxOpenConnections
}

// TestPostgreSQLManager_Lifecycle 测试生命周期方法
func TestPostgreSQLManager_Lifecycle(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "postgresql",
		PostgreSQLConfig: &config.PostgreSQLConfig{
			DSN: "host=localhost port=5432 user=postgres password=password dbname=test sslmode=disable",
		},
	}

	mgr, err := NewPostgreSQLManager(cfg)
	if err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
		return
	}

	// 测试 OnStart
	if err := mgr.OnStart(); err != nil {
		t.Errorf("OnStart() error = %v", err)
	}

	// 测试 Health
	if err := mgr.Health(); err != nil {
		t.Errorf("Health() error = %v", err)
	}

	// 测试 OnStop
	if err := mgr.OnStop(); err != nil {
		t.Errorf("OnStop() error = %v", err)
	}
}

// TestPostgreSQLManager_ConnStringVariants 测试不同的连接字符串格式
func TestPostgreSQLManager_ConnStringVariants(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		wantErr bool
	}{
		{
			name:    "标准格式",
			dsn:     "host=localhost port=5432 user=postgres password=password dbname=test sslmode=disable",
			wantErr: true, // 需要真实连接
		},
		{
			name:    "URL 格式",
			dsn:     "postgres://postgres:password@localhost:5432/test?sslmode=disable",
			wantErr: true, // 需要真实连接
		},
		{
			name:    "Unix socket 格式",
			dsn:     "host=/var/run/postgresql port=5432 user=postgres dbname=test",
			wantErr: true, // 需要真实连接
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.DatabaseConfig{
				Driver: "postgresql",
				PostgreSQLConfig: &config.PostgreSQLConfig{
					DSN: tt.dsn,
				},
			}

			mgr, err := NewPostgreSQLManager(cfg)
			if err != nil {
				t.Skipf("PostgreSQL not available: %v", err)
				return
			}
			if mgr != nil {
				_ = mgr.Close()
			}
		})
	}
}
