package drivers

import (
	"context"
	"testing"
	"time"

	"com.litelake.litecore/manager/databasemgr/internal/config"
)

func TestNewSQLiteManager(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.DatabaseConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &config.DatabaseConfig{
				Driver: "sqlite",
				SQLiteConfig: &config.SQLiteConfig{
					DSN: "file::memory:?cache=shared",
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with pool config",
			config: &config.DatabaseConfig{
				Driver: "sqlite",
				SQLiteConfig: &config.SQLiteConfig{
					DSN: "file::memory:?cache=shared",
					PoolConfig: &config.PoolConfig{
						MaxOpenConns:    1,
						MaxIdleConns:    1,
						ConnMaxLifetime: 30 * time.Second,
						ConnMaxIdleTime: 5 * time.Minute,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid config - missing driver",
			config: &config.DatabaseConfig{
				Driver: "",
				SQLiteConfig: &config.SQLiteConfig{
					DSN: "file::memory:?cache=shared",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid config - missing sqlite config",
			config: &config.DatabaseConfig{
				Driver: "sqlite",
			},
			wantErr: true,
		},
		{
			name: "invalid config - missing DSN",
			config: &config.DatabaseConfig{
				Driver: "sqlite",
				SQLiteConfig: &config.SQLiteConfig{
					DSN: "",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := NewSQLiteManager(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSQLiteManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				defer mgr.Close()

				if mgr == nil {
					t.Fatal("NewSQLiteManager() returned nil manager")
				}

				if mgr.ManagerName() != "sqlite-database" {
					t.Errorf("ManagerName() = %v, want %v", mgr.ManagerName(), "sqlite-database")
				}

				if mgr.Driver() != "sqlite" {
					t.Errorf("Driver() = %v, want %v", mgr.Driver(), "sqlite")
				}

				if mgr.DB() == nil {
					t.Error("DB() returned nil")
				}
			}
		})
	}
}

func TestSQLiteManager_DB(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: "file::memory:?cache=shared",
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
	}
	defer mgr.Close()

	db := mgr.DB()
	if db == nil {
		t.Fatal("DB() returned nil")
	}

	// Test that we can execute a query
	ctx := context.Background()
	var result int
	err = db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		t.Errorf("Failed to execute query: %v", err)
	}

	if result != 1 {
		t.Errorf("Query result = %v, want 1", result)
	}
}

func TestSQLiteManager_Driver(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: "file::memory:?cache=shared",
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
	}
	defer mgr.Close()

	if got := mgr.Driver(); got != "sqlite" {
		t.Errorf("Driver() = %v, want %v", got, "sqlite")
	}
}

func TestSQLiteManager_Ping(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: "file::memory:?cache=shared",
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
	}
	defer mgr.Close()

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "normal context",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "nil context",
			ctx:     nil,
			wantErr: true,
		},
		{
			name:    "timeout context",
			ctx:     func() context.Context { ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond); time.Sleep(10 * time.Millisecond); cancel(); return ctx }(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.Ping(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLiteManager.Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSQLiteManager_BeginTx(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: "file::memory:?cache=shared",
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
	}
	defer mgr.Close()

	ctx := context.Background()

	// Test successful transaction
	tx, err := mgr.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx() error = %v", err)
	}

	// Create table
	_, err = tx.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
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

	// Test rollback
	tx2, err := mgr.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx() error = %v", err)
	}

	_, err = tx2.Exec("INSERT INTO test (name) VALUES (?)", "test2")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	err = tx2.Rollback()
	if err != nil {
		t.Errorf("Failed to rollback transaction: %v", err)
	}

	// Verify data was not committed
	err = mgr.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM test").Scan(&count)
	if err != nil {
		t.Errorf("Failed to query data: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 row after rollback, got %d", count)
	}
}

func TestSQLiteManager_BeginTx_NilContext(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: "file::memory:?cache=shared",
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
	}
	defer mgr.Close()

	tx, err := mgr.BeginTx(nil, nil)
	if err == nil {
		tx.Rollback()
		t.Error("BeginTx() with nil context should return error")
	}
}

func TestSQLiteManager_Stats(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: "file::memory:?cache=shared",
			PoolConfig: &config.PoolConfig{
				MaxOpenConns: 1,
				MaxIdleConns: 1,
			},
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
	}
	defer mgr.Close()

	stats := mgr.Stats()

	// Verify stats are returned
	if stats.MaxOpenConnections != 1 {
		t.Errorf("MaxOpenConnections = %v, want 1", stats.MaxOpenConnections)
	}
}

func TestSQLiteManager_Close(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: "file::memory:?cache=shared",
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
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

func TestSQLiteManager_Health(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: "file::memory:?cache=shared",
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
	}
	defer mgr.Close()

	// Health should pass
	err = mgr.Health()
	if err != nil {
		t.Errorf("Health() error = %v, want nil", err)
	}

	// Close the manager
	mgr.Close()

	// Health should fail after close
	err = mgr.Health()
	if err == nil {
		t.Error("Health() should return error after close")
	}
}

func TestSQLiteManager_OnStart(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: "file::memory:?cache=shared",
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
	}
	defer mgr.Close()

	// OnStart should pass
	err = mgr.OnStart()
	if err != nil {
		t.Errorf("OnStart() error = %v, want nil", err)
	}
}

func TestSQLiteManager_OnStop(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: "file::memory:?cache=shared",
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
	}

	// OnStop should close the database
	err = mgr.OnStop()
	if err != nil {
		t.Errorf("OnStop() error = %v, want nil", err)
	}

	// Verify DB is nil after OnStop
	if mgr.DB() != nil {
		t.Error("DB() should return nil after OnStop")
	}
}

func TestSQLiteManager_Shutdown(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: "file::memory:?cache=shared",
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
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

func TestSQLiteManager_ConcurrentAccess(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Driver: "sqlite",
		SQLiteConfig: &config.SQLiteConfig{
			DSN: "file::memory:?cache=shared",
			PoolConfig: &config.PoolConfig{
				MaxOpenConns: 1, // SQLite 通常设置为 1
				MaxIdleConns: 1,
			},
		},
	}

	mgr, err := NewSQLiteManager(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteManager() error = %v", err)
	}
	defer mgr.Close()

	// Create table
	ctx := context.Background()
	_, err = mgr.DB().ExecContext(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	done := make(chan bool)

	// Concurrent reads and writes
	for i := 0; i < 10; i++ {
		go func(i int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Concurrent access panicked: %v", r)
				}
			}()

			// Query data (read operation)
			var count int
			err = mgr.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM test").Scan(&count)
			if err != nil {
				t.Errorf("Failed to query data: %v", err)
			}

			_ = mgr.Stats()
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}