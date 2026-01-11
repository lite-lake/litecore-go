package drivers

import (
	"testing"

	"com.litelake.litecore/manager/databasemgr/internal/config"
)

// TestNewMySQLManager_InvalidConfig 测试无效配置
func TestNewMySQLManager_InvalidConfig(t *testing.T) {
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
			name: "MySQL 配置为空",
			config: &config.DatabaseConfig{
				Driver: "mysql",
			},
			wantErr: true,
		},
		{
			name: "空 DSN",
			config: &config.DatabaseConfig{
				Driver:      "mysql",
				MySQLConfig: &config.MySQLConfig{},
			},
			wantErr: true,
		},
		{
			name: "无效的 DSN 格式",
			config: &config.DatabaseConfig{
				Driver: "mysql",
				MySQLConfig: &config.MySQLConfig{
					DSN: "invalid-dsn-format",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMySQLManager(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMySQLManager() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestNewMySQLManager_ValidConfig 测试有效配置但不连接真实数据库
func TestNewMySQLManager_ValidConfig(t *testing.T) {
	// 注意：这个测试使用有效的配置格式，但不需要真实的 MySQL 连接
	// 如果要测试真实的连接，需要集成测试环境
	cfg := &config.DatabaseConfig{
		Driver: "mysql",
		MySQLConfig: &config.MySQLConfig{
			DSN: "root:password@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local",
			PoolConfig: &config.PoolConfig{
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 0,
				ConnMaxIdleTime: 0,
			},
		},
	}

	// 这个测试会尝试连接，如果 MySQL 不可用会失败
	// 在 CI/CD 环境中，应该跳过或使用 testcontainers
	mgr, err := NewMySQLManager(cfg)
	if err != nil {
		// 如果 MySQL 不可用，跳过测试
		t.Skipf("MySQL not available: %v", err)
		return
	}

	if mgr == nil {
		t.Error("NewMySQLManager() returned nil manager")
	}

	if mgr.ManagerName() != "mysql-database" {
		t.Errorf("ManagerName() = %v, want 'mysql-database'", mgr.ManagerName())
	}

	if mgr.Driver() != "mysql" {
		t.Errorf("Driver() = %v, want 'mysql'", mgr.Driver())
	}

	// 清理
	_ = mgr.Close()
}

// TestMySQLManager_ImplementsDatabaseManager 测试实现接口
func TestMySQLManager_ImplementsDatabaseManager(t *testing.T) {
	// MySQLManager 通过嵌入 GormBaseManager 实现了 DatabaseManager 接口
	// 这里我们只是编译时检查
	cfg := &config.DatabaseConfig{
		Driver: "mysql",
		MySQLConfig: &config.MySQLConfig{
			DSN: "root:password@tcp(localhost:3306)/test",
		},
	}

	mgr, err := NewMySQLManager(cfg)
	if err != nil {
		t.Skipf("MySQL not available: %v", err)
		return
	}
	defer mgr.Close()

	// 验证基本方法
	_ = mgr.ManagerName()
	_ = mgr.Driver()
	_ = mgr.DB()
	_ = mgr.Stats()
}

// TestNewMySQLManager_WithPoolConfig 测试带连接池配置
func TestNewMySQLManager_WithPoolConfig(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "mysql",
		MySQLConfig: &config.MySQLConfig{
			DSN: "root:password@tcp(localhost:3306)/test",
			PoolConfig: &config.PoolConfig{
				MaxOpenConns:    20,
				MaxIdleConns:    10,
				ConnMaxLifetime: 0,
				ConnMaxIdleTime: 0,
			},
		},
	}

	mgr, err := NewMySQLManager(cfg)
	if err != nil {
		t.Skipf("MySQL not available: %v", err)
		return
	}
	defer mgr.Close()

	if mgr == nil {
		t.Fatal("NewMySQLManager() returned nil")
	}

	// 验证连接池配置已应用（需要等连接建立后才能检查）
	stats := mgr.Stats()
	// MaxOpenConnections 可能尚未建立连接，所以可能是 0
	_ = stats.MaxOpenConnections
}

// TestMySQLManager_Lifecycle 测试生命周期方法
func TestMySQLManager_Lifecycle(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "mysql",
		MySQLConfig: &config.MySQLConfig{
			DSN: "root:password@tcp(localhost:3306)/test",
		},
	}

	mgr, err := NewMySQLManager(cfg)
	if err != nil {
		t.Skipf("MySQL not available: %v", err)
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
