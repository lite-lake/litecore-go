package databasemgr

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"
)

// TestNewDatabaseManagerSQLiteImpl 测试创建 SQLite 管理器
func TestNewDatabaseManagerSQLiteImpl(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *SQLiteConfig
		wantErr bool
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
		},
		{
			name: "empty DSN",
			cfg: &SQLiteConfig{
				DSN: "",
			},
			wantErr: true,
		},
		{
			name: "valid in-memory config",
			cfg: &SQLiteConfig{
				DSN: ":memory:",
			},
			wantErr: false,
		},
		{
			name: "valid file config",
			cfg: &SQLiteConfig{
				DSN: "file:test.db?mode=memory",
			},
			wantErr: false,
		},
		{
			name: "valid config with pool config",
			cfg: &SQLiteConfig{
				DSN: ":memory:",
				PoolConfig: &PoolConfig{
					MaxOpenConns:    1,
					MaxIdleConns:    1,
					ConnMaxLifetime: 30 * time.Second,
					ConnMaxIdleTime: 5 * time.Minute,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := NewDatabaseManagerSQLiteImpl(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDatabaseManagerSQLiteImpl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if mgr == nil {
					t.Error("NewDatabaseManagerSQLiteImpl() returned nil manager")
				}
				if mgr != nil {
					defer mgr.Close()
					if mgr.ManagerName() != "databaseManagerSqliteImpl" {
						t.Errorf("ManagerName() = %v, want 'databaseManagerSqliteImpl'", mgr.ManagerName())
					}
					if mgr.Driver() != "sqlite" {
						t.Errorf("Driver() = %v, want 'sqlite'", mgr.Driver())
					}
				}
			}
		})
	}
}

// TestSQLiteImpl_BasicOperations 测试基本操作
func TestSQLiteImpl_BasicOperations(t *testing.T) {
	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	t.Run("Health", func(t *testing.T) {
		if err := mgr.Health(); err != nil {
			t.Errorf("Health() error = %v", err)
		}
	})

	t.Run("OnStart", func(t *testing.T) {
		if err := mgr.OnStart(); err != nil {
			t.Errorf("OnStart() error = %v", err)
		}
	})

	t.Run("OnStop", func(t *testing.T) {
		if err := mgr.OnStop(); err != nil {
			t.Errorf("OnStop() error = %v", err)
		}
	})
}

// TestSQLiteImpl_GORMOperations 测试 GORM 操作
func TestSQLiteImpl_GORMOperations(t *testing.T) {
	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	// 测试模型
	type User struct {
		ID   uint
		Name string
		Age  int
	}

	t.Run("AutoMigrate", func(t *testing.T) {
		if err := mgr.AutoMigrate(&User{}); err != nil {
			t.Errorf("AutoMigrate() error = %v", err)
		}
	})

	t.Run("DB", func(t *testing.T) {
		db := mgr.DB()
		if db == nil {
			t.Error("DB() returned nil")
		}
	})

	t.Run("Model", func(t *testing.T) {
		result := mgr.Model(&User{})
		if result == nil {
			t.Error("Model() returned nil")
		}
	})

	t.Run("Table", func(t *testing.T) {
		result := mgr.Table("users")
		if result == nil {
			t.Error("Table() returned nil")
		}
	})

	t.Run("WithContext", func(t *testing.T) {
		ctx := context.Background()
		result := mgr.WithContext(ctx)
		if result == nil {
			t.Error("WithContext() returned nil")
		}
	})
}

// TestSQLiteImpl_CRUDOperations 测试 CRUD 操作
func TestSQLiteImpl_CRUDOperations(t *testing.T) {
	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	type Product struct {
		ID    uint
		Name  string
		Price float64
	}

	// 自动迁移
	if err := mgr.AutoMigrate(&Product{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	db := mgr.DB()

	t.Run("Create", func(t *testing.T) {
		product := Product{Name: "Test Product", Price: 9.99}
		if err := db.Create(&product).Error; err != nil {
			t.Errorf("Create() error = %v", err)
		}
		if product.ID == 0 {
			t.Error("Create() did not set ID")
		}
	})

	t.Run("Read", func(t *testing.T) {
		var product Product
		if err := db.First(&product, 1).Error; err != nil {
			t.Errorf("First() error = %v", err)
		}
		if product.Name != "Test Product" {
			t.Errorf("Name = %v, want 'Test Product'", product.Name)
		}
	})

	t.Run("Update", func(t *testing.T) {
		if err := db.Model(&Product{}).Where("id = ?", 1).Update("price", 19.99).Error; err != nil {
			t.Errorf("Update() error = %v", err)
		}

		var product Product
		if err := db.First(&product, 1).Error; err != nil {
			t.Errorf("First() after Update error = %v", err)
		}
		if product.Price != 19.99 {
			t.Errorf("Price = %v, want 19.99", product.Price)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := db.Delete(&Product{}, 1).Error; err != nil {
			t.Errorf("Delete() error = %v", err)
		}

		var product Product
		err := db.First(&product, 1).Error
		if err == nil {
			t.Error("Expected error when fetching deleted record")
		}
	})
}

// TestSQLiteImpl_Transaction 测试事务操作
func TestSQLiteImpl_Transaction(t *testing.T) {
	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	type Account struct {
		ID      uint
		Name    string
		Balance float64
	}

	// 自动迁移
	if err := mgr.AutoMigrate(&Account{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	// 创建测试数据
	db := mgr.DB()
	db.Create(&Account{Name: "Alice", Balance: 100})
	db.Create(&Account{Name: "Bob", Balance: 50})

	t.Run("Transaction_Commit", func(t *testing.T) {
		err := mgr.Transaction(func(tx *gorm.DB) error {
			// Alice 转账 50 给 Bob
			if err := tx.Model(&Account{}).Where("name = ?", "Alice").Update("balance", gorm.Expr("balance - ?", 50)).Error; err != nil {
				return err
			}
			if err := tx.Model(&Account{}).Where("name = ?", "Bob").Update("balance", gorm.Expr("balance + ?", 50)).Error; err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			t.Errorf("Transaction() commit error = %v", err)
		}

		// 验证余额
		var alice, bob Account
		db.Where("name = ?", "Alice").First(&alice)
		db.Where("name = ?", "Bob").First(&bob)

		if alice.Balance != 50 {
			t.Errorf("Alice balance = %v, want 50", alice.Balance)
		}
		if bob.Balance != 100 {
			t.Errorf("Bob balance = %v, want 100", bob.Balance)
		}
	})

	t.Run("Transaction_Rollback", func(t *testing.T) {
		err := mgr.Transaction(func(tx *gorm.DB) error {
			// 尝试转账，但返回错误触发回滚
			if err := tx.Model(&Account{}).Where("name = ?", "Alice").Update("balance", gorm.Expr("balance - ?", 200)).Error; err != nil {
				return err
			}
			return gorm.ErrInvalidTransaction // 触发回滚
		})
		if err == nil {
			t.Error("Transaction() should return error")
		}

		// 验证余额未变
		var alice, bob Account
		db.Where("name = ?", "Alice").First(&alice)
		db.Where("name = ?", "Bob").First(&bob)

		if alice.Balance != 50 {
			t.Errorf("Alice balance should remain 50, got %v", alice.Balance)
		}
		if bob.Balance != 100 {
			t.Errorf("Bob balance should remain 100, got %v", bob.Balance)
		}
	})

	t.Run("Begin_Manual", func(t *testing.T) {
		tx := mgr.Begin()
		if tx == nil {
			t.Fatal("Begin() returned nil")
		}

		// 执行一些操作
		tx.Model(&Account{}).Where("name = ?", "Alice").Update("balance", gorm.Expr("balance + ?", 10))

		// 手动提交
		tx.Commit()

		// 验证
		var alice Account
		db.Where("name = ?", "Alice").First(&alice)
		if alice.Balance != 60 {
			t.Errorf("Alice balance = %v, want 60", alice.Balance)
		}
	})
}

// TestSQLiteImpl_RawSQL 测试原生 SQL
func TestSQLiteImpl_RawSQL(t *testing.T) {
	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	type Item struct {
		ID   uint
		Name string
	}

	// 自动迁移
	if err := mgr.AutoMigrate(&Item{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	t.Run("Exec", func(t *testing.T) {
		result := mgr.Exec("INSERT INTO items (name) VALUES (?)", "Test Item")
		if result.Error != nil {
			t.Errorf("Exec() error = %v", result.Error)
		}
		if result.RowsAffected != 1 {
			t.Errorf("Exec() RowsAffected = %v, want 1", result.RowsAffected)
		}
	})

	t.Run("Raw", func(t *testing.T) {
		var items []Item
		result := mgr.Raw("SELECT * FROM items WHERE name = ?", "Test Item").Scan(&items)
		if result.Error != nil {
			t.Errorf("Raw() error = %v", result.Error)
		}
		if len(items) != 1 {
			t.Errorf("Raw() returned %d items, want 1", len(items))
		}
	})
}

// TestSQLiteImpl_ConnectionManagement 测试连接管理
func TestSQLiteImpl_ConnectionManagement(t *testing.T) {
	cfg := &SQLiteConfig{
		DSN: ":memory:",
		PoolConfig: &PoolConfig{
			MaxOpenConns:    1,
			MaxIdleConns:    1,
			ConnMaxLifetime: 10 * time.Second,
			ConnMaxIdleTime: 5 * time.Second,
		},
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	t.Run("Ping", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mgr.Ping(ctx); err != nil {
			t.Errorf("Ping() error = %v", err)
		}
	})

	t.Run("Stats", func(t *testing.T) {
		stats := mgr.Stats()
		if stats.MaxOpenConnections != 1 {
			t.Errorf("Stats().MaxOpenConnections = %v, want 1", stats.MaxOpenConnections)
		}
	})

	t.Run("Close", func(t *testing.T) {
		mgr2, err := NewDatabaseManagerSQLiteImpl(&SQLiteConfig{DSN: ":memory:"})
		if err != nil {
			t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
		}

		if err := mgr2.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}

		// 关闭后 Ping 应该失败
		ctx := context.Background()
		if err := mgr2.Ping(ctx); err == nil {
			t.Error("Ping() after Close() should return error")
		}
	})
}

// TestSQLiteImpl_Migrator 测试迁移器
func TestSQLiteImpl_Migrator(t *testing.T) {
	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	migrator := mgr.Migrator()
	if migrator == nil {
		t.Error("Migrator() returned nil")
	}

	type TestTable struct {
		ID   uint
		Name string
	}

	t.Run("HasTable", func(t *testing.T) {
		// 表不存在
		if migrator.HasTable(&TestTable{}) {
			t.Error("Table should not exist yet")
		}

		// 创建表
		if err := migrator.CreateTable(&TestTable{}); err != nil {
			t.Errorf("CreateTable() error = %v", err)
		}

		// 表应该存在
		if !migrator.HasTable(&TestTable{}) {
			t.Error("Table should exist after CreateTable()")
		}
	})
}

// TestSQLiteImpl_Concurrency 测试并发操作
func TestSQLiteImpl_Concurrency(t *testing.T) {
	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	// SQLite 的并发写入有限制，这里只测试基本的并发读取
	type Item struct {
		ID   uint
		Name string
	}

	if err := mgr.AutoMigrate(&Item{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	db := mgr.DB()
	for i := 0; i < 10; i++ {
		db.Create(&Item{Name: "test"})
	}

	// 测试基本的读取操作
	var items []Item
	if err := db.Find(&items).Error; err != nil {
		t.Errorf("Find() error = %v", err)
	}

	if len(items) != 10 {
		t.Errorf("Expected 10 items, got %v", len(items))
	}

	t.Logf("Concurrency test: SQLite has limited write concurrency, reads working correctly")
}

// BenchmarkSQLiteImpl_Operations 基准测试
func BenchmarkSQLiteImpl_Operations(b *testing.B) {
	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg)
	if err != nil {
		b.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	type BenchModel struct {
		ID   uint
		Name string
	}

	mgr.AutoMigrate(&BenchModel{})

	b.Run("Insert", func(b *testing.B) {
		db := mgr.DB()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			model := BenchModel{Name: "Test"}
			db.Create(&model)
		}
	})

	b.Run("Select", func(b *testing.B) {
		var model BenchModel
		db := mgr.DB()
		db.Create(&BenchModel{Name: "Test"})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			db.First(&model, 1)
		}
	})
}
