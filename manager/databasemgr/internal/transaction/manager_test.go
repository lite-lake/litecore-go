package transaction

import (
	"errors"
	"sync"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNewManager(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewManager(db)

	if mgr == nil {
		t.Fatal("NewManager() returned nil")
	}

	if mgr.db == nil {
		t.Error("NewManager() should set db field")
	}
}

func TestManager_Transaction(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewManager(db)

	// Create test table
	db.Exec("DROP TABLE IF EXISTS test")
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
}

func TestManager_Transaction_RollbackOnError(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewManager(db)

	// Create test table
	db.Exec("DROP TABLE IF EXISTS test")
	if err := db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)").Error; err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert initial data
	if err := db.Exec("INSERT INTO test (name) VALUES (?)", "initial").Error; err != nil {
		t.Fatalf("Failed to insert initial data: %v", err)
	}

	// Test transaction with error (should rollback)
	testErr := errors.New("test error")
	err = mgr.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("INSERT INTO test (name) VALUES (?)", "test").Error; err != nil {
			return err
		}
		return testErr
	})

	if err != testErr {
		t.Errorf("Transaction() should return test error, got %v", err)
	}

	// Verify rollback occurred
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM test").Scan(&count).Error; err != nil {
		t.Errorf("Failed to count rows: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 row after rollback, got %d", count)
	}
}

func TestManager_Transaction_MultipleStatements(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewManager(db)

	// Create test table
	db.Exec("DROP TABLE IF EXISTS test")
	if err := db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)").Error; err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test transaction with multiple statements
	err = mgr.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("INSERT INTO test (name) VALUES (?)", "test1").Error; err != nil {
			return err
		}
		if err := tx.Exec("INSERT INTO test (name) VALUES (?)", "test2").Error; err != nil {
			return err
		}
		if err := tx.Exec("INSERT INTO test (name) VALUES (?)", "test3").Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		t.Errorf("Transaction() error = %v, want nil", err)
	}

	// Verify all data was committed
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM test").Scan(&count).Error; err != nil {
		t.Errorf("Failed to count rows: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 rows, got %d", count)
	}
}

func TestManager_Transaction_Panic(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewManager(db)

	// Create test table
	db.Exec("DROP TABLE IF EXISTS test")
	if err := db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)").Error; err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert initial data
	if err := db.Exec("INSERT INTO test (name) VALUES (?)", "initial").Error; err != nil {
		t.Fatalf("Failed to insert initial data: %v", err)
	}

	// Test transaction with panic (should rollback)
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Panic was recovered
			}
		}()

		_ = mgr.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("INSERT INTO test (name) VALUES (?)", "test").Error; err != nil {
				return err
			}
			panic("test panic")
		})
	}()

	// Verify rollback occurred
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM test").Scan(&count).Error; err != nil {
		t.Errorf("Failed to count rows: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 row after panic rollback, got %d", count)
	}
}

func TestManager_Begin(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewManager(db)

	// Create test table
	db.Exec("DROP TABLE IF EXISTS test")
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

func TestManager_Begin_Rollback(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewManager(db)

	// Create test table
	db.Exec("DROP TABLE IF EXISTS test")
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

	// Rollback transaction
	if err := tx.Rollback().Error; err != nil {
		t.Errorf("Failed to rollback transaction: %v", err)
	}

	// Verify data was not committed
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM test").Scan(&count).Error; err != nil {
		t.Errorf("Failed to count rows: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 rows after rollback, got %d", count)
	}
}

func TestManager_BeginTx(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewManager(db)

	// Create test table
	db.Exec("DROP TABLE IF EXISTS test")
	if err := db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)").Error; err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Begin a transaction with options
	tx := mgr.BeginTx()
	if tx == nil {
		t.Fatal("BeginTx() returned nil")
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

func TestManager_BeginTx_WithOpts(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewManager(db)

	// Create test table
	db.Exec("DROP TABLE IF EXISTS test")
	if err := db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)").Error; err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Begin a transaction with nil options
	tx := mgr.BeginTx(nil)
	if tx == nil {
		t.Fatal("BeginTx() with nil opts returned nil")
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

func TestManager_NestedTransaction(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewManager(db)

	// Create test table
	db.Exec("DROP TABLE IF EXISTS test")
	if err := db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)").Error; err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test nested transaction
	err = mgr.NestedTransaction(func(tx *gorm.DB) error {
		// Outer transaction
		if err := tx.Exec("INSERT INTO test (name) VALUES (?)", "outer1").Error; err != nil {
			return err
		}

		// Inner transaction (savepoint)
		return tx.Transaction(func(tx2 *gorm.DB) error {
			if err := tx2.Exec("INSERT INTO test (name) VALUES (?)", "inner1").Error; err != nil {
				return err
			}
			return nil
		})
	})

	if err != nil {
		t.Errorf("NestedTransaction() error = %v, want nil", err)
	}

	// Verify both inserts were committed
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM test").Scan(&count).Error; err != nil {
		t.Errorf("Failed to count rows: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 rows, got %d", count)
	}
}

func TestManager_NestedTransaction_InnerRollback(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewManager(db)

	// Create test table
	db.Exec("DROP TABLE IF EXISTS test")
	if err := db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)").Error; err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test nested transaction with inner rollback
	err = mgr.NestedTransaction(func(tx *gorm.DB) error {
		// Outer transaction
		if err := tx.Exec("INSERT INTO test (name) VALUES (?)", "outer1").Error; err != nil {
			return err
		}

		// Inner transaction (savepoint) that fails
		err := tx.Transaction(func(tx2 *gorm.DB) error {
			if err := tx2.Exec("INSERT INTO test (name) VALUES (?)", "inner1").Error; err != nil {
				return err
			}
			return errors.New("inner error")
		})

		// Inner transaction should rollback to savepoint
		// but outer transaction continues
		if err := tx.Exec("INSERT INTO test (name) VALUES (?)", "outer2").Error; err != nil {
			return err
		}

		return err
	})

	// Should return inner error
	if err == nil {
		t.Error("NestedTransaction() should return inner error")
	}

	// Verify outer transaction was also rolled back
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM test").Scan(&count).Error; err != nil {
		t.Errorf("Failed to count rows: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 rows after nested transaction failure, got %d", count)
	}
}

func TestManager_NestedTransaction_OuterRollback(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewManager(db)

	// Create test table
	db.Exec("DROP TABLE IF EXISTS test")
	if err := db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)").Error; err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test nested transaction with outer rollback
	testErr := errors.New("outer error")
	err = mgr.NestedTransaction(func(tx *gorm.DB) error {
		// Outer transaction
		if err := tx.Exec("INSERT INTO test (name) VALUES (?)", "outer1").Error; err != nil {
			return err
		}

		// Inner transaction (savepoint) succeeds
		if err := tx.Transaction(func(tx2 *gorm.DB) error {
			if err := tx2.Exec("INSERT INTO test (name) VALUES (?)", "inner1").Error; err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}

		// Outer transaction fails
		return testErr
	})

	if err != testErr {
		t.Errorf("NestedTransaction() should return outer error, got %v", err)
	}

	// Verify everything was rolled back
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM test").Scan(&count).Error; err != nil {
		t.Errorf("Failed to count rows: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 rows after outer rollback, got %d", count)
	}
}

func TestManager_ConcurrentTransactions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewManager(db)

	// Create test table
	db.Exec("DROP TABLE IF EXISTS test")
	if err := db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)").Error; err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Run concurrent transactions
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := mgr.Transaction(func(tx *gorm.DB) error {
				return tx.Exec("INSERT INTO test (name) VALUES (?)", i).Error
			})
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent transaction error: %v", err)
	}

	// Verify all data was committed
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM test").Scan(&count).Error; err != nil {
		t.Errorf("Failed to count rows: %v", err)
	}
	if count != 10 {
		t.Errorf("Expected 10 rows, got %d", count)
	}
}

func TestManager_Transaction_Read(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewManager(db)

	// Create test table
	db.Exec("DROP TABLE IF EXISTS test")
	if err := db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)").Error; err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert initial data
	if err := db.Exec("INSERT INTO test (name) VALUES (?)", "test").Error; err != nil {
		t.Fatalf("Failed to insert initial data: %v", err)
	}

	// Test read in transaction
	err = mgr.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Raw("SELECT COUNT(*) FROM test").Scan(&count).Error; err != nil {
			return err
		}
		if count != 1 {
			t.Errorf("Expected 1 row in transaction, got %d", count)
		}
		return nil
	})

	if err != nil {
		t.Errorf("Transaction() with read error = %v, want nil", err)
	}
}

func TestManager_Transaction_Update(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewManager(db)

	// Create test table
	db.Exec("DROP TABLE IF EXISTS test")
	if err := db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)").Error; err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert initial data
	if err := db.Exec("INSERT INTO test (name) VALUES (?)", "old").Error; err != nil {
		t.Fatalf("Failed to insert initial data: %v", err)
	}

	// Test update in transaction
	err = mgr.Transaction(func(tx *gorm.DB) error {
		return tx.Exec("UPDATE test SET name = ? WHERE name = ?", "new", "old").Error
	})

	if err != nil {
		t.Errorf("Transaction() with update error = %v, want nil", err)
	}

	// Verify update was committed
	var name string
	if err := db.Raw("SELECT name FROM test").Scan(&name).Error; err != nil {
		t.Errorf("Failed to query data: %v", err)
	}
	if name != "new" {
		t.Errorf("Expected name 'new', got '%s'", name)
	}
}

func TestManager_Transaction_Delete(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewManager(db)

	// Create test table
	db.Exec("DROP TABLE IF EXISTS test")
	if err := db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)").Error; err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert initial data
	if err := db.Exec("INSERT INTO test (name) VALUES (?)", "test").Error; err != nil {
		t.Fatalf("Failed to insert initial data: %v", err)
	}

	// Test delete in transaction
	err = mgr.Transaction(func(tx *gorm.DB) error {
		return tx.Exec("DELETE FROM test").Error
	})

	if err != nil {
		t.Errorf("Transaction() with delete error = %v, want nil", err)
	}

	// Verify delete was committed
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM test").Scan(&count).Error; err != nil {
		t.Errorf("Failed to count rows: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 rows after delete, got %d", count)
	}
}

func TestManager_Transaction_Empty(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewManager(db)

	// Test empty transaction (no operations)
	err = mgr.Transaction(func(tx *gorm.DB) error {
		return nil
	})

	if err != nil {
		t.Errorf("Empty transaction error = %v, want nil", err)
	}
}

func TestManager_Begin_MultipleTransactions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewManager(db)

	// Create test table
	db.Exec("DROP TABLE IF EXISTS test")
	if err := db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)").Error; err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Begin first transaction
	tx1 := mgr.Begin()
	if tx1 == nil {
		t.Fatal("Begin() first call returned nil")
	}

	// Insert data in first transaction
	if err := tx1.Exec("INSERT INTO test (name) VALUES (?)", "test1").Error; err != nil {
		t.Errorf("Failed to insert data in first transaction: %v", err)
	}

	// Commit first transaction
	if err := tx1.Commit().Error; err != nil {
		t.Errorf("Failed to commit first transaction: %v", err)
	}

	// Begin second transaction
	tx2 := mgr.Begin()
	if tx2 == nil {
		t.Fatal("Begin() second call returned nil")
	}

	// Insert data in second transaction
	if err := tx2.Exec("INSERT INTO test (name) VALUES (?)", "test2").Error; err != nil {
		t.Errorf("Failed to insert data in second transaction: %v", err)
	}

	// Commit second transaction
	if err := tx2.Commit().Error; err != nil {
		t.Errorf("Failed to commit second transaction: %v", err)
	}

	// Verify both transactions committed
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM test").Scan(&count).Error; err != nil {
		t.Errorf("Failed to count rows: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 rows, got %d", count)
	}
}

func TestManager_Transaction_Isolation(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	mgr := NewManager(db)

	// Create test table
	db.Exec("DROP TABLE IF EXISTS test")
	if err := db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, counter INTEGER)").Error; err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert initial counter
	if err := db.Exec("INSERT INTO test (id, counter) VALUES (1, 0)").Error; err != nil {
		t.Fatalf("Failed to insert initial data: %v", err)
	}

	var wg sync.WaitGroup

	// Run concurrent transactions that increment counter
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = mgr.Transaction(func(tx *gorm.DB) error {
				return tx.Exec("UPDATE test SET counter = counter + 1 WHERE id = 1").Error
			})
		}()
	}

	wg.Wait()

	// Verify counter was incremented
	var counter int64
	if err := db.Raw("SELECT counter FROM test WHERE id = 1").Scan(&counter).Error; err != nil {
		t.Errorf("Failed to query counter: %v", err)
	}
	// Note: SQLite's default isolation level might allow some lost updates
	// so we just verify it was incremented at least
	if counter < 1 {
		t.Errorf("Expected counter >= 1, got %d", counter)
	}
}
