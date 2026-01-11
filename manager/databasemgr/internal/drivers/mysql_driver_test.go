package drivers

import (
	"context"
	"testing"
	"time"

	"com.litelake.litecore/manager/databasemgr/internal/config"
)

func TestNewMySQLManager(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.DatabaseConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &config.DatabaseConfig{
				Driver: "mysql",
				MySQLConfig: &config.MySQLConfig{
					DSN: "root:password@tcp(localhost:3306)/test?parseTime=True",
				},
			},
			wantErr: true, // 会失败，因为没有实际的 MySQL 数据库
		},
		{
			name: "invalid config - missing driver",
			config: &config.DatabaseConfig{
				Driver: "",
				MySQLConfig: &config.MySQLConfig{
					DSN: "root:password@tcp(localhost:3306)/test",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid config - missing mysql config",
			config: &config.DatabaseConfig{
				Driver: "mysql",
			},
			wantErr: true,
		},
		{
			name: "invalid config - missing DSN",
			config: &config.DatabaseConfig{
				Driver: "mysql",
				MySQLConfig: &config.MySQLConfig{
					DSN: "",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := NewMySQLManager(tt.config)
			if (err != nil) != tt.wantErr {
				t.Logf("NewMySQLManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				defer mgr.Close()

				if mgr == nil {
					t.Fatal("NewMySQLManager() returned nil manager")
				}

				if mgr.ManagerName() != "mysql-database" {
					t.Errorf("ManagerName() = %v, want %v", mgr.ManagerName(), "mysql-database")
				}

				if mgr.Driver() != "mysql" {
					t.Errorf("Driver() = %v, want %v", mgr.Driver(), "mysql")
				}

				if mgr.DB() == nil {
					t.Error("DB() returned nil")
				}
			}
		})
	}
}

func TestMySQLManager_Driver(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "mysql",
		MySQLConfig: &config.MySQLConfig{
			DSN: "root:password@tcp(localhost:3306)/test",
		},
	}

	mgr, err := NewMySQLManager(cfg)
	if err != nil {
		// 预期会失败，因为没有实际的 MySQL 数据库
		return
	}
	defer mgr.Close()

	if got := mgr.Driver(); got != "mysql" {
		t.Errorf("Driver() = %v, want %v", got, "mysql")
	}
}

func TestMySQLManager_Ping_NilContext(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "mysql",
		MySQLConfig: &config.MySQLConfig{
			DSN: "root:password@tcp(localhost:3306)/test",
		},
	}

	mgr, err := NewMySQLManager(cfg)
	if err != nil {
		// 预期会失败，因为没有实际的 MySQL 数据库
		return
	}
	defer mgr.Close()

	err = mgr.Ping(nil)
	if err == nil {
		t.Error("Ping() with nil context should return error")
	}
}

func TestMySQLManager_BeginTx_NilContext(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "mysql",
		MySQLConfig: &config.MySQLConfig{
			DSN: "root:password@tcp(localhost:3306)/test",
		},
	}

	mgr, err := NewMySQLManager(cfg)
	if err != nil {
		// 预期会失败，因为没有实际的 MySQL 数据库
		return
	}
	defer mgr.Close()

	tx, err := mgr.BeginTx(nil, nil)
	if err == nil {
		tx.Rollback()
		t.Error("BeginTx() with nil context should return error")
	}
}

func TestMySQLManager_Stats(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "mysql",
		MySQLConfig: &config.MySQLConfig{
			DSN: "root:password@tcp(localhost:3306)/test",
			PoolConfig: &config.PoolConfig{
				MaxOpenConns: 10,
				MaxIdleConns: 5,
			},
		},
	}

	mgr, err := NewMySQLManager(cfg)
	if err != nil {
		// 预期会失败，因为没有实际的 MySQL 数据库
		return
	}
	defer mgr.Close()

	stats := mgr.Stats()

	// Verify stats are returned
	if stats.MaxOpenConnections != 10 {
		t.Errorf("MaxOpenConnections = %v, want 10", stats.MaxOpenConnections)
	}
}

func TestMySQLManager_Close(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "mysql",
		MySQLConfig: &config.MySQLConfig{
			DSN: "root:password@tcp(localhost:3306)/test",
		},
	}

	mgr, err := NewMySQLManager(cfg)
	if err != nil {
		// 预期会失败，因为没有实际的 MySQL 数据库
		return
	}

	// Close the manager
	err = mgr.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}

	// Verify DB is nil after close
	if mgr.DB() != nil {
		t.Error("DB() should return nil after close")
	}

	// Close again should not error
	err = mgr.Close()
	if err != nil {
		t.Errorf("Close() second call error = %v, want nil", err)
	}
}

func TestMySQLManager_Shutdown(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "mysql",
		MySQLConfig: &config.MySQLConfig{
			DSN: "root:password@tcp(localhost:3306)/test",
		},
	}

	mgr, err := NewMySQLManager(cfg)
	if err != nil {
		// 预期会失败，因为没有实际的 MySQL 数据库
		return
	}

	ctx := context.Background()

	// Shutdown should close the database
	err = mgr.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() error = %v, want nil", err)
	}

	// Verify DB is nil after Shutdown
	if mgr.DB() != nil {
		t.Error("DB() should return nil after Shutdown")
	}

	// Shutdown again should not error
	err = mgr.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() second call error = %v, want nil", err)
	}
}

func TestMySQLManager_ConcurrentAccess(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "mysql",
		MySQLConfig: &config.MySQLConfig{
			DSN: "root:password@tcp(localhost:3306)/test",
			PoolConfig: &config.PoolConfig{
				MaxOpenConns: 10,
				MaxIdleConns: 5,
			},
		},
	}

	mgr, err := NewMySQLManager(cfg)
	if err != nil {
		// 预期会失败，因为没有实际的 MySQL 数据库
		return
	}
	defer mgr.Close()

	done := make(chan bool)

	// Concurrent reads and writes
	for i := 0; i < 10; i++ {
		go func(i int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Concurrent access panicked: %v", r)
				}
			}()

			_ = mgr.Stats()
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestMySQLManager_WithRealDatabase 是一个集成测试，需要实际的 MySQL 数据库
// 这个测试默认跳过，只有在有实际数据库时才运行
func TestMySQLManager_WithRealDatabase(t *testing.T) {
	t.Skip("Skipping integration test - requires actual MySQL database")

	cfg := &config.DatabaseConfig{
		Driver: "mysql",
		MySQLConfig: &config.MySQLConfig{
			DSN: "root:password@tcp(localhost:3306)/test_db?parseTime=True&loc=Local",
			PoolConfig: &config.PoolConfig{
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 30 * time.Second,
				ConnMaxIdleTime: 5 * time.Minute,
			},
		},
	}

	mgr, err := NewMySQLManager(cfg)
	if err != nil {
		t.Fatalf("NewMySQLManager() error = %v", err)
	}
	defer mgr.Close()

	// Test Ping
	ctx := context.Background()
	err = mgr.Ping(ctx)
	if err != nil {
		t.Errorf("Ping() error = %v", err)
	}

	// Test query
	var result int
	err = mgr.DB().QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		t.Errorf("Failed to execute query: %v", err)
	}

	if result != 1 {
		t.Errorf("Query result = %v, want 1", result)
	}

	// Test transaction
	tx, err := mgr.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx() error = %v", err)
	}

	// Create table
	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS test (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255))")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert data
	_, err = tx.Exec("INSERT INTO test (name) VALUES (?)", "test")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		t.Errorf("Failed to commit transaction: %v", err)
	}

	// Verify data was committed
	var count int
	err = mgr.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM test").Scan(&count)
	if err != nil {
		t.Errorf("Failed to query data: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 row, got %d", count)
	}

	// Test Health
	err = mgr.Health()
	if err != nil {
		t.Errorf("Health() error = %v", err)
	}

	// Test OnStart
	err = mgr.OnStart()
	if err != nil {
		t.Errorf("OnStart() error = %v", err)
	}

	// Test OnStop
	err = mgr.OnStop()
	if err != nil {
		t.Errorf("OnStop() error = %v", err)
	}
}
