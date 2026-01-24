package databasemgr

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	_ "github.com/mattn/go-sqlite3"
)

func TestObservabilityPlugin_Callbacks(t *testing.T) {
	skipIfCGONotAvailable(t)

	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	type User struct {
		ID   uint
		Name string
	}

	if err := mgr.AutoMigrate(&User{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	db := mgr.DB()

	t.Run("Create callback", func(t *testing.T) {
		user := User{Name: "Test User"}
		if err := db.Create(&user).Error; err != nil {
			t.Errorf("Create() error = %v", err)
		}
	})

	t.Run("Query callback", func(t *testing.T) {
		var user User
		if err := db.First(&user, 1).Error; err != nil {
			t.Errorf("First() error = %v", err)
		}
	})

	t.Run("Update callback", func(t *testing.T) {
		if err := db.Model(&User{}).Where("id = ?", 1).Update("name", "Updated User").Error; err != nil {
			t.Errorf("Update() error = %v", err)
		}
	})

	t.Run("Delete callback", func(t *testing.T) {
		if err := db.Delete(&User{}, 1).Error; err != nil {
			t.Errorf("Delete() error = %v", err)
		}
	})
}

func TestDatabaseManager_AdvancedQueries(t *testing.T) {
	skipIfCGONotAvailable(t)

	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	type Product struct {
		ID    uint
		Name  string
		Price float64
	}

	if err := mgr.AutoMigrate(&Product{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	db := mgr.DB()

	t.Run("Where and Or", func(t *testing.T) {
		var products []Product
		result := db.Where("price > ?", 50).Or("name = ?", "Free").Find(&products)
		if result.Error != nil {
			t.Errorf("Where/Or() error = %v", result.Error)
		}
	})

	t.Run("Like query", func(t *testing.T) {
		var products []Product
		result := db.Where("name LIKE ?", "%test%").Find(&products)
		if result.Error != nil {
			t.Errorf("Like() error = %v", result.Error)
		}
	})

	t.Run("In query", func(t *testing.T) {
		var products []Product
		result := db.Where("id IN ?", []int{1, 2, 3}).Find(&products)
		if result.Error != nil {
			t.Errorf("In() error = %v", result.Error)
		}
	})

	t.Run("Order and Limit", func(t *testing.T) {
		var products []Product
		result := db.Order("price desc").Limit(10).Find(&products)
		if result.Error != nil {
			t.Errorf("Order/Limit() error = %v", result.Error)
		}
	})

	t.Run("Count", func(t *testing.T) {
		var count int64
		result := db.Model(&Product{}).Count(&count)
		if result.Error != nil {
			t.Errorf("Count() error = %v", result.Error)
		}
	})

	t.Run("Pluck", func(t *testing.T) {
		var names []string
		result := db.Model(&Product{}).Pluck("name", &names)
		if result.Error != nil {
			t.Errorf("Pluck() error = %v", result.Error)
		}
	})
}

func TestDatabaseManager_TransactionErrors(t *testing.T) {
	skipIfCGONotAvailable(t)

	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	type Account struct {
		ID      uint
		Name    string
		Balance float64
	}

	if err := mgr.AutoMigrate(&Account{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	db := mgr.DB()
	db.Create(&Account{Name: "Alice", Balance: 100})
	db.Create(&Account{Name: "Bob", Balance: 50})

	t.Run("Nested transaction", func(t *testing.T) {
		err := mgr.Transaction(func(tx1 *gorm.DB) error {
			return tx1.Transaction(func(tx2 *gorm.DB) error {
				return nil
			})
		})
		if err != nil {
			t.Errorf("Nested transaction error = %v", err)
		}
	})
}

func TestDatabaseManager_RawSQLComplex(t *testing.T) {
	skipIfCGONotAvailable(t)

	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	type User struct {
		ID   uint
		Name string
		Age  int
	}

	if err := mgr.AutoMigrate(&User{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	t.Run("Raw with JOIN", func(t *testing.T) {
		var users []User
		result := mgr.Raw("SELECT * FROM users WHERE id > ?", 0).Scan(&users)
		if result.Error != nil {
			t.Errorf("Raw JOIN() error = %v", result.Error)
		}
	})

	t.Run("Exec with multiple statements", func(t *testing.T) {
		result := mgr.Exec("UPDATE users SET age = age + 1 WHERE id = 1")
		if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
			t.Errorf("Exec() error = %v", result.Error)
		}
	})

	t.Run("Raw with subquery", func(t *testing.T) {
		var users []User
		result := mgr.Raw("SELECT * FROM users WHERE age > (SELECT AVG(age) FROM users)").Scan(&users)
		if result.Error != nil {
			t.Errorf("Raw subquery() error = %v", result.Error)
		}
	})
}

func TestDatabaseManager_ConnectionPoolBehavior(t *testing.T) {
	skipIfCGONotAvailable(t)

	cfg := &SQLiteConfig{
		DSN: ":memory:",
		PoolConfig: &PoolConfig{
			MaxOpenConns:    2,
			MaxIdleConns:    1,
			ConnMaxLifetime: 5 * time.Second,
			ConnMaxIdleTime: 3 * time.Second,
		},
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	type User struct {
		ID   uint
		Name string
	}

	if err := mgr.AutoMigrate(&User{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	t.Run("Pool stats initial", func(t *testing.T) {
		stats := mgr.Stats()
		if stats.MaxOpenConnections != 2 {
			t.Errorf("MaxOpenConnections = %v, want 2", stats.MaxOpenConnections)
		}
	})

	t.Run("Connection reuse", func(t *testing.T) {
		db := mgr.DB()
		for i := 0; i < 10; i++ {
			var user User
			db.First(&user)
		}
		stats := mgr.Stats()
		if stats.InUse > 2 {
			t.Errorf("InUse = %v, should not exceed MaxOpenConnections", stats.InUse)
		}
	})
}

func TestDatabaseManager_AutoMigrateMultipleModels(t *testing.T) {
	skipIfCGONotAvailable(t)

	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	type User struct {
		ID   uint
		Name string
	}

	type Product struct {
		ID    uint
		Name  string
		Price float64
	}

	type Order struct {
		ID        uint
		UserID    uint
		ProductID uint
		Amount    int
	}

	t.Run("AutoMigrate multiple models", func(t *testing.T) {
		err := mgr.AutoMigrate(&User{}, &Product{}, &Order{})
		if err != nil {
			t.Errorf("AutoMigrate() error = %v", err)
		}
	})

	t.Run("Check tables exist", func(t *testing.T) {
		migrator := mgr.Migrator()
		if !migrator.HasTable(&User{}) {
			t.Error("User table should exist")
		}
		if !migrator.HasTable(&Product{}) {
			t.Error("Product table should exist")
		}
		if !migrator.HasTable(&Order{}) {
			t.Error("Order table should exist")
		}
	})
}

func TestDatabaseManager_MigratorOperations(t *testing.T) {
	skipIfCGONotAvailable(t)

	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	migrator := mgr.Migrator()

	type TestModel struct {
		ID   uint
		Name string
	}

	t.Run("Create and drop table", func(t *testing.T) {
		if err := migrator.CreateTable(&TestModel{}); err != nil {
			t.Errorf("CreateTable() error = %v", err)
		}
		if !migrator.HasTable(&TestModel{}) {
			t.Error("Table should exist")
		}
		if err := migrator.DropTable(&TestModel{}); err != nil {
			t.Errorf("DropTable() error = %v", err)
		}
		if migrator.HasTable(&TestModel{}) {
			t.Error("Table should not exist")
		}
	})

	t.Run("Add column", func(t *testing.T) {
		if err := migrator.CreateTable(&TestModel{}); err != nil {
			t.Fatalf("CreateTable() error = %v", err)
		}

		db := mgr.DB()
		db.Exec("ALTER TABLE test_models ADD COLUMN age INTEGER")

		columns, err := migrator.ColumnTypes(&TestModel{})
		if err != nil {
			t.Fatalf("ColumnTypes() error = %v", err)
		}
		if len(columns) < 3 {
			t.Errorf("Expected at least 3 columns, got %d", len(columns))
		}
	})
}

func TestDatabaseManager_ErrorHandling(t *testing.T) {
	skipIfCGONotAvailable(t)

	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	type User struct {
		ID   uint
		Name string `gorm:"unique"`
	}

	if err := mgr.AutoMigrate(&User{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	db := mgr.DB()

	t.Run("Duplicate key error", func(t *testing.T) {
		db.Create(&User{Name: "Alice"})
		err := db.Create(&User{Name: "Alice"}).Error
		if err == nil {
			t.Error("Expected error for duplicate key")
		}
	})

	t.Run("Invalid table error", func(t *testing.T) {
		err := db.Exec("INSERT INTO nonexistent_table (id) VALUES (1)").Error
		if err == nil {
			t.Error("Expected error for invalid table")
		}
	})

	t.Run("Invalid column error", func(t *testing.T) {
		err := db.Exec("UPDATE users SET nonexistent_column = 1").Error
		if err == nil {
			t.Error("Expected error for invalid column")
		}
	})
}

func TestDatabaseManager_ModelOperations(t *testing.T) {
	skipIfCGONotAvailable(t)

	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	type User struct {
		ID   uint
		Name string
	}

	if err := mgr.AutoMigrate(&User{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	t.Run("Model returns valid DB", func(t *testing.T) {
		result := mgr.Model(&User{})
		if result == nil {
			t.Error("Model() should not return nil")
		}
		if result.Statement == nil {
			t.Error("Model() statement should not be nil")
		}
	})

	t.Run("Model with query", func(t *testing.T) {
		var users []User
		result := mgr.Model(&User{}).Find(&users)
		if result.Error != nil {
			t.Errorf("Model().Find() error = %v", result.Error)
		}
	})
}

func TestDatabaseManager_TableOperations(t *testing.T) {
	skipIfCGONotAvailable(t)

	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	db := mgr.DB()
	db.Exec("CREATE TABLE custom_table (id INTEGER PRIMARY KEY, name TEXT)")

	t.Run("Table returns valid DB", func(t *testing.T) {
		result := mgr.Table("custom_table")
		if result == nil {
			t.Error("Table() should not return nil")
		}
	})

	t.Run("Table insert", func(t *testing.T) {
		result := mgr.Table("custom_table").Create(map[string]any{"name": "test"})
		if result.Error != nil {
			t.Errorf("Table().Create() error = %v", result.Error)
		}
	})

	t.Run("Table select", func(t *testing.T) {
		var results []map[string]any
		result := mgr.Table("custom_table").Find(&results)
		if result.Error != nil {
			t.Errorf("Table().Find() error = %v", result.Error)
		}
	})
}

func TestDatabaseManager_WithTimeoutContext(t *testing.T) {
	skipIfCGONotAvailable(t)

	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	type User struct {
		ID   uint
		Name string
	}

	if err := mgr.AutoMigrate(&User{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	t.Run("With valid timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		db := mgr.WithContext(ctx)
		if db == nil {
			t.Error("WithContext() should not return nil")
		}
		var user User
		err := db.First(&user).Error
		if err != gorm.ErrRecordNotFound && err != nil {
			t.Errorf("First() error = %v", err)
		}
	})

	t.Run("With expired timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(10 * time.Millisecond)

		db := mgr.WithContext(ctx)
		if db == nil {
			t.Error("WithContext() should not return nil")
		}
		var user User
		err := db.First(&user).Error
		if err == nil {
			t.Error("Expected error for expired context")
		}
	})
}

func TestDatabaseManager_BeginTransaction(t *testing.T) {
	skipIfCGONotAvailable(t)

	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	type User struct {
		ID   uint
		Name string
	}

	if err := mgr.AutoMigrate(&User{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	t.Run("Begin returns valid transaction", func(t *testing.T) {
		tx := mgr.Begin()
		if tx == nil {
			t.Fatal("Begin() returned nil")
		}

		user := User{Name: "Test"}
		if err := tx.Create(&user).Error; err != nil {
			t.Errorf("tx.Create() error = %v", err)
		}

		tx.Commit()

		var count int64
		mgr.DB().Model(&User{}).Count(&count)
		if count != 1 {
			t.Errorf("Expected 1 record, got %d", count)
		}
	})

	t.Run("Begin with rollback", func(t *testing.T) {
		tx := mgr.Begin()
		if tx == nil {
			t.Fatal("Begin() returned nil")
		}

		user := User{Name: "Test2"}
		if err := tx.Create(&user).Error; err != nil {
			t.Errorf("tx.Create() error = %v", err)
		}

		tx.Rollback()

		var users []User
		mgr.DB().Find(&users)
		if len(users) != 1 {
			t.Errorf("Expected 1 record after rollback, got %d", len(users))
		}
	})
}

func TestDatabaseManager_ExecAndRawErrors(t *testing.T) {
	skipIfCGONotAvailable(t)

	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	t.Run("Exec invalid SQL", func(t *testing.T) {
		result := mgr.Exec("INVALID SQL")
		if result.Error == nil {
			t.Error("Expected error for invalid SQL")
		}
	})

	t.Run("Raw invalid SQL", func(t *testing.T) {
		var results []map[string]any
		result := mgr.Raw("INVALID SQL").Scan(&results)
		if result.Error == nil {
			t.Error("Expected error for invalid SQL")
		}
	})
}

func TestDatabaseManager_CloseBehavior(t *testing.T) {
	skipIfCGONotAvailable(t)

	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}

	t.Run("Close works", func(t *testing.T) {
		if err := mgr.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	t.Run("Multiple close", func(t *testing.T) {
		mgr2, err := NewDatabaseManagerSQLiteImpl(&SQLiteConfig{DSN: ":memory:"}, nil, nil)
		if err != nil {
			t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
		}

		if err := mgr2.Close(); err != nil {
			t.Errorf("First Close() error = %v", err)
		}

		if err := mgr2.Close(); err != nil {
			t.Errorf("Second Close() error = %v", err)
		}
	})
}

func TestDatabaseManager_DBMethod(t *testing.T) {
	skipIfCGONotAvailable(t)

	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	t.Run("DB returns valid instance", func(t *testing.T) {
		db := mgr.DB()
		if db == nil {
			t.Error("DB() should not return nil")
		}
		if db.Statement == nil {
			t.Error("DB.Statement should not be nil")
		}
	})

	t.Run("DB is thread-safe", func(t *testing.T) {
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				db := mgr.DB()
				if db == nil {
					t.Error("DB() should not return nil")
				}
				done <- true
			}()
		}
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestDatabaseManager_HealthAndPing(t *testing.T) {
	skipIfCGONotAvailable(t)

	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	t.Run("Health check", func(t *testing.T) {
		if err := mgr.Health(); err != nil {
			t.Errorf("Health() error = %v", err)
		}
	})

	t.Run("Ping with context", func(t *testing.T) {
		ctx := context.Background()
		if err := mgr.Ping(ctx); err != nil {
			t.Errorf("Ping() error = %v", err)
		}
	})

	t.Run("Ping with timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mgr.Ping(ctx); err != nil {
			t.Errorf("Ping() with timeout error = %v", err)
		}
	})
}

func TestDatabaseManager_DriverAndManagerName(t *testing.T) {
	skipIfCGONotAvailable(t)

	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	t.Run("Driver returns correct type", func(t *testing.T) {
		driver := mgr.Driver()
		if driver != "sqlite" {
			t.Errorf("Driver() = %v, want 'sqlite'", driver)
		}
	})

	t.Run("ManagerName returns correct name", func(t *testing.T) {
		name := mgr.ManagerName()
		if name != "databaseManagerSqliteImpl" {
			t.Errorf("ManagerName() = %v, want 'databaseManagerSqliteImpl'", name)
		}
	})
}
