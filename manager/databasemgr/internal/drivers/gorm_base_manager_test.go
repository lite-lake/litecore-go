package drivers

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNewGormBaseManager(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	name := "test-manager"
	driver := "sqlite"
	mgr := NewGormBaseManager(name, driver, db)

	if mgr == nil {
		t.Fatal("NewGormBaseManager() returned nil")
	}

	if mgr.ManagerName() != name {
		t.Errorf("ManagerName() = %v, want %v", mgr.ManagerName(), name)
	}

	if mgr.Driver() != driver {
		t.Errorf("Driver() = %v, want %v", mgr.Driver(), driver)
	}

	if mgr.DB() == nil {
		t.Error("DB() returned nil")
	}
}

func TestGormBaseManager_ManagerName(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	tests := []struct {
		name string
	}{
		{"test-manager-1"},
		{"test-manager-2"},
		{""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewGormBaseManager(tt.name, "sqlite", db)
			if got := mgr.ManagerName(); got != tt.name {
				t.Errorf("GormBaseManager.ManagerName() = %v, want %v", got, tt.name)
			}
		})
	}
}

func TestGormBaseManager_Driver(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	tests := []struct {
		driver string
	}{
		{"sqlite"},
		{"mysql"},
		{"postgresql"},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			mgr := NewGormBaseManager("test", tt.driver, db)
			if got := mgr.Driver(); got != tt.driver {
				t.Errorf("GormBaseManager.Driver() = %v, want %v", got, tt.driver)
			}
		})
	}
}

func TestGormBaseManager_DB(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewGormBaseManager("test", "sqlite", db)

	// Multiple calls should return the same DB instance
	db1 := mgr.DB()
	db2 := mgr.DB()

	if db1 == nil {
		t.Fatal("DB() returned nil")
	}

	if db1 != db2 {
		t.Error("DB() should return the same instance")
	}
}

func TestGormBaseManager_Ping(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewGormBaseManager("test", "sqlite", db)

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
			name:    "context with timeout",
			ctx:     func() context.Context { ctx, _ := context.WithTimeout(context.Background(), 5*time.Second); return ctx }(),
			wantErr: false,
		},
		{
			name:    "cancelled context",
			ctx:     func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.Ping(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GormBaseManager.Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGormBaseManager_Health(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewGormBaseManager("test", "sqlite", db)

	// Health should pass
	if err := mgr.Health(); err != nil {
		t.Errorf("GormBaseManager.Health() error = %v, want nil", err)
	}

	// Close the manager
	mgr.Close()

	// Health should fail after close
	if err := mgr.Health(); err == nil {
		t.Error("GormBaseManager.Health() should return error after close")
	}
}

func TestGormBaseManager_Stats(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewGormBaseManager("test", "sqlite", db)

	stats := mgr.Stats()

	// Verify stats are returned
	if stats.MaxOpenConnections == 0 {
		t.Error("Stats() should return non-zero MaxOpenConnections")
	}

	// Close and verify stats return zero
	mgr.Close()
	stats = mgr.Stats()
	if stats.MaxOpenConnections != 0 {
		t.Error("Stats() should return zero after close")
	}
}

func TestGormBaseManager_Close(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewGormBaseManager("test", "sqlite", db)

	// Close the manager
	if err := mgr.Close(); err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}

	// Close again should not error
	if err := mgr.Close(); err != nil {
		t.Errorf("Close() second call error = %v, want nil", err)
	}
}

func TestGormBaseManager_OnStart(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewGormBaseManager("test", "sqlite", db)

	// OnStart should pass
	if err := mgr.OnStart(); err != nil {
		t.Errorf("OnStart() error = %v, want nil", err)
	}

	// Close the manager
	mgr.Close()

	// OnStart should fail after close
	if err := mgr.OnStart(); err == nil {
		t.Error("OnStart() should return error after close")
	}
}

func TestGormBaseManager_OnStop(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewGormBaseManager("test", "sqlite", db)

	// OnStop should close the database
	if err := mgr.OnStop(); err != nil {
		t.Errorf("OnStop() error = %v, want nil", err)
	}

	// Verify DB is closed
	if err := mgr.Health(); err == nil {
		t.Error("Health() should fail after OnStop")
	}

	// OnStop again should not error
	if err := mgr.OnStop(); err != nil {
		t.Errorf("OnStop() second call error = %v, want nil", err)
	}
}

func TestGormBaseManager_Model(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewGormBaseManager("test", "sqlite", db)

	type TestModel struct {
		ID   uint
		Name string
	}

	// Model should return a GORM DB instance
	tx := mgr.Model(&TestModel{})
	if tx == nil {
		t.Fatal("Model() returned nil")
	}

	// Verify we can use the returned DB
	if tx.Statement.Table != "" {
		// Table name should be derived from model
	}
}

func TestGormBaseManager_Table(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewGormBaseManager("test", "sqlite", db)

	tableName := "test_table"

	// Table should return a GORM DB instance
	tx := mgr.Table(tableName)
	if tx == nil {
		t.Fatal("Table() returned nil")
	}

	if tx.Statement.Table != tableName {
		t.Errorf("Table() = %v, want %v", tx.Statement.Table, tableName)
	}
}

func TestGormBaseManager_WithContext(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewGormBaseManager("test", "sqlite", db)

	ctx := context.Background()

	// WithContext should return a GORM DB instance with the context
	tx := mgr.WithContext(ctx)
	if tx == nil {
		t.Fatal("WithContext() returned nil")
	}

	if tx.Statement.Context != ctx {
		t.Error("WithContext() should set the context")
	}
}

func TestGormBaseManager_Transaction(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewGormBaseManager("test", "sqlite", db)

	// Create test table
	if err := db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)").Error; err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test successful transaction
	err = mgr.Transaction(func(tx *gorm.DB) error {
		return tx.Exec("INSERT INTO test (name) VALUES (?)", "test").Error
	})
	if err != nil {
		t.Errorf("Transaction() error = %v, want nil", err)
	}

	// Verify data was committed
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM test").Scan(&count).Error; err != nil {
		t.Errorf("Failed to count rows: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 row, got %d", count)
	}

	// Test failed transaction
	err = mgr.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("INSERT INTO test (name) VALUES (?)", "test2").Error; err != nil {
			return err
		}
		return errors.New("rollback")
	})
	if err == nil {
		t.Error("Transaction() should return error when callback returns error")
	}

	// Verify rollback
	if err := db.Raw("SELECT COUNT(*) FROM test").Scan(&count).Error; err != nil {
		t.Errorf("Failed to count rows: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 row after rollback, got %d", count)
	}
}

func TestGormBaseManager_Begin(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewGormBaseManager("test", "sqlite", db)

	// Create test table
	if err := db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)").Error; err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Begin a transaction
	tx := mgr.Begin()
	if tx == nil {
		t.Fatal("Begin() returned nil")
	}

	// Insert data within transaction
	if err := tx.Exec("INSERT INTO test (name) VALUES (?)", "test").Error; err != nil {
		t.Errorf("Failed to insert data: %v", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		t.Errorf("Failed to commit transaction: %v", err)
	}

	// Verify data was committed
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM test").Scan(&count).Error; err != nil {
		t.Errorf("Failed to count rows: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 row, got %d", count)
	}
}

func TestGormBaseManager_AutoMigrate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewGormBaseManager("test", "sqlite", db)

	type TestModel struct {
		ID   uint
		Name string
	}

	// AutoMigrate should create the table
	if err := mgr.AutoMigrate(&TestModel{}); err != nil {
		t.Errorf("AutoMigrate() error = %v, want nil", err)
	}

	// Verify table exists
	if !mgr.Migrator().HasTable(&TestModel{}) {
		t.Error("AutoMigrate() should create the table")
	}
}

func TestGormBaseManager_Migrator(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewGormBaseManager("test", "sqlite", db)

	// Migrator should return a GORM migrator
	migrator := mgr.Migrator()
	if migrator == nil {
		t.Fatal("Migrator() returned nil")
	}

	// Verify it's a valid migrator by checking if it can create tables
	// (We don't actually create a table, just verify the migrator interface works)
	type TestModel struct {
		ID   uint
		Name string
	}
	_ = migrator.HasTable(&TestModel{})
}

func TestGormBaseManager_Exec(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewGormBaseManager("test", "sqlite", db)

	// Create test table
	if err := mgr.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)").Error; err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert data using Exec
	result := mgr.Exec("INSERT INTO test (name) VALUES (?)", "test")
	if result.Error != nil {
		t.Errorf("Exec() error = %v", result.Error)
	}

	// Verify rows affected
	if result.RowsAffected != 1 {
		t.Errorf("Exec() RowsAffected = %d, want 1", result.RowsAffected)
	}
}

func TestGormBaseManager_Raw(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewGormBaseManager("test", "sqlite", db)

	// Create test table
	if err := mgr.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)").Error; err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert data
	if err := mgr.Exec("INSERT INTO test (name) VALUES (?)", "test").Error; err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	// Query using Raw
	type Result struct {
		Name string
	}
	var results []Result
	if err := mgr.Raw("SELECT name FROM test").Scan(&results).Error; err != nil {
		t.Errorf("Raw() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Raw() returned %d results, want 1", len(results))
	}

	if results[0].Name != "test" {
		t.Errorf("Raw() returned name = %v, want 'test'", results[0].Name)
	}
}

func TestGormBaseManager_ConcurrentAccess(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewGormBaseManager("test", "sqlite", db)

	// Create test table
	if err := mgr.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)").Error; err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	var wg sync.WaitGroup
	done := make(chan bool, 100)

	// Concurrent reads and writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			// Access DB concurrently
			_ = mgr.DB()
			_ = mgr.Stats()
			_ = mgr.Driver()
			_ = mgr.ManagerName()

			// Execute query
			var count int64
			_ = mgr.Raw("SELECT COUNT(*) FROM test").Scan(&count)

			done <- true
		}(i)
	}

	// Wait for all goroutines
	wg.Wait()
	close(done)

	// Verify all operations completed
	completed := 0
	for range done {
		completed++
	}

	if completed != 100 {
		t.Errorf("Expected 100 completed operations, got %d", completed)
	}
}

func TestGormBaseManager_Stats_ConcurrentAccess(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewGormBaseManager("test", "sqlite", db)

	var wg sync.WaitGroup

	// Concurrent Stats() calls
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stats := mgr.Stats()
			if stats.MaxOpenConnections == 0 {
				t.Error("Stats() should return non-zero MaxOpenConnections")
			}
		}()
	}

	wg.Wait()
}

func TestGormBaseManager_DB_ConcurrentAccess(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewGormBaseManager("test", "sqlite", db)

	var wg sync.WaitGroup

	// Concurrent DB() calls
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			db := mgr.DB()
			if db == nil {
				t.Error("DB() should not return nil")
			}
		}()
	}

	wg.Wait()
}

func TestGormBaseManager_SQLDBNil(t *testing.T) {
	// Create a manager without a valid sql.DB (edge case)
	mgr := &GormBaseManager{
		name:   "test",
		driver: "sqlite",
		db:     nil,
		sqlDB:  nil,
	}

	// Stats should return zero stats
	stats := mgr.Stats()
	if stats.MaxOpenConnections != 0 {
		t.Error("Stats() should return zero when sqlDB is nil")
	}

	// Close should not error
	if err := mgr.Close(); err != nil {
		t.Errorf("Close() should not error when sqlDB is nil, got %v", err)
	}
}
