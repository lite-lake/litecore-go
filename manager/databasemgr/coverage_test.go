package databasemgr

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"gorm.io/gorm"

	_ "github.com/mattn/go-sqlite3"
)

func TestDatabaseManagerFactory_Coverage(t *testing.T) {
	t.Run("Build MySQL with nil config", func(t *testing.T) {
		_, err := Build("mysql", nil, nil, nil)
		if err == nil {
			t.Error("Build should return error for nil MySQL config")
		}
	})

	t.Run("Build MySQL with empty DSN", func(t *testing.T) {
		cfg := map[string]any{
			"dsn": "",
		}
		_, err := Build("mysql", cfg, nil, nil)
		if err == nil {
			t.Error("Build should return error for empty DSN")
		}
	})

	t.Run("Build PostgreSQL with nil config", func(t *testing.T) {
		_, err := Build("postgresql", nil, nil, nil)
		if err == nil {
			t.Error("Build should return error for nil PostgreSQL config")
		}
	})

	t.Run("Build PostgreSQL with empty DSN", func(t *testing.T) {
		cfg := map[string]any{
			"dsn": "",
		}
		_, err := Build("postgresql", cfg, nil, nil)
		if err == nil {
			t.Error("Build should return error for empty DSN")
		}
	})

	t.Run("Build with invalid driver", func(t *testing.T) {
		_, err := Build("invalid", nil, nil, nil)
		if err == nil {
			t.Error("Build should return error for invalid driver")
		}
	})
}

func TestDatabaseManagerConfig_Coverage(t *testing.T) {
	t.Run("PoolConfig with max values", func(t *testing.T) {
		cfg := &PoolConfig{
			MaxOpenConns:    1000,
			MaxIdleConns:    500,
			ConnMaxLifetime: 1 * time.Hour,
			ConnMaxIdleTime: 30 * time.Minute,
		}
		err := cfg.Validate()
		if err != nil {
			t.Errorf("Validate() error = %v", err)
		}
	})

	t.Run("PoolConfig with zero values", func(t *testing.T) {
		cfg := &PoolConfig{
			MaxOpenConns:    0,
			MaxIdleConns:    0,
			ConnMaxLifetime: 0,
			ConnMaxIdleTime: 0,
		}
		err := cfg.Validate()
		if err != nil {
			t.Errorf("Validate() error = %v", err)
		}
	})

	t.Run("ParsePoolConfig with all types", func(t *testing.T) {
		cfg := map[string]any{
			"max_open_conns":     float64(10),
			"max_idle_conns":     int(5),
			"conn_max_lifetime":  "1h",
			"conn_max_idle_time": float64(300),
		}
		poolCfg, err := parsePoolConfig(cfg)
		if err != nil {
			t.Errorf("parsePoolConfig() error = %v", err)
		}
		if poolCfg.MaxOpenConns != 10 {
			t.Errorf("MaxOpenConns = %v, want 10", poolCfg.MaxOpenConns)
		}
		if poolCfg.MaxIdleConns != 5 {
			t.Errorf("MaxIdleConns = %v, want 5", poolCfg.MaxIdleConns)
		}
		if poolCfg.ConnMaxLifetime != 1*time.Hour {
			t.Errorf("ConnMaxLifetime = %v, want 1h", poolCfg.ConnMaxLifetime)
		}
		if poolCfg.ConnMaxIdleTime != 5*time.Minute {
			t.Errorf("ConnMaxIdleTime = %v, want 5m", poolCfg.ConnMaxIdleTime)
		}
	})

	t.Run("ParseObservabilityConfig with edge values", func(t *testing.T) {
		cfg := map[string]any{
			"slow_query_threshold": 0,
			"log_sql":              true,
			"sample_rate":          0.0,
		}
		obsCfg, err := parseObservabilityConfig(cfg)
		if err != nil {
			t.Errorf("parseObservabilityConfig() error = %v", err)
		}
		if obsCfg.SlowQueryThreshold != 0 {
			t.Errorf("SlowQueryThreshold = %v, want 0", obsCfg.SlowQueryThreshold)
		}
		if !obsCfg.LogSQL {
			t.Error("LogSQL should be true")
		}
		if obsCfg.SampleRate != 0.0 {
			t.Errorf("SampleRate = %v, want 0.0", obsCfg.SampleRate)
		}
	})

	t.Run("ParseObservabilityConfig with max sample rate", func(t *testing.T) {
		cfg := map[string]any{
			"slow_query_threshold": "1s",
			"log_sql":              false,
			"sample_rate":          1.0,
		}
		obsCfg, err := parseObservabilityConfig(cfg)
		if err != nil {
			t.Errorf("parseObservabilityConfig() error = %v", err)
		}
		if obsCfg.SlowQueryThreshold != 1*time.Second {
			t.Errorf("SlowQueryThreshold = %v, want 1s", obsCfg.SlowQueryThreshold)
		}
		if obsCfg.LogSQL {
			t.Error("LogSQL should be false")
		}
		if obsCfg.SampleRate != 1.0 {
			t.Errorf("SampleRate = %v, want 1.0", obsCfg.SampleRate)
		}
	})

	t.Run("ParseMySQLConfig with all types", func(t *testing.T) {
		cfg := map[string]any{
			"dsn": "root:password@tcp(localhost:3306)/test",
			"pool_config": map[string]any{
				"max_open_conns": 100,
				"max_idle_conns": 10,
			},
		}
		mysqlCfg, err := parseMySQLConfig(cfg)
		if err != nil {
			t.Errorf("parseMySQLConfig() error = %v", err)
		}
		if mysqlCfg.DSN != "root:password@tcp(localhost:3306)/test" {
			t.Errorf("DSN = %v", mysqlCfg.DSN)
		}
	})

	t.Run("ParsePostgreSQLConfig with all types", func(t *testing.T) {
		cfg := map[string]any{
			"dsn": "host=localhost port=5432",
			"pool_config": map[string]any{
				"max_open_conns": 50,
				"max_idle_conns": 5,
			},
		}
		postgresCfg, err := parsePostgreSQLConfig(cfg)
		if err != nil {
			t.Errorf("parsePostgreSQLConfig() error = %v", err)
		}
		if postgresCfg.DSN != "host=localhost port=5432" {
			t.Errorf("DSN = %v", postgresCfg.DSN)
		}
	})

	t.Run("ParseSQLiteConfig with file path", func(t *testing.T) {
		cfg := map[string]any{
			"dsn": "file:test.db?mode=rwc",
			"pool_config": map[string]any{
				"max_open_conns": 1,
			},
		}
		sqliteCfg, err := parseSQLiteConfig(cfg)
		if err != nil {
			t.Errorf("parseSQLiteConfig() error = %v", err)
		}
		if sqliteCfg.DSN != "file:test.db?mode=rwc" {
			t.Errorf("DSN = %v", sqliteCfg.DSN)
		}
	})
}

func TestDatabaseManager_SQLiteEdgeCases(t *testing.T) {
	skipIfCGONotAvailable(t)

	t.Run("SQLite with file database", func(t *testing.T) {
		cfg := &SQLiteConfig{
			DSN: "file:test.db?mode=rwc",
		}
		mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
		if err != nil {
			t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
		}
		defer mgr.Close()

		if mgr.Driver() != "sqlite" {
			t.Errorf("Driver() = %v, want 'sqlite'", mgr.Driver())
		}
	})

	t.Run("SQLite with shared cache", func(t *testing.T) {
		cfg := &SQLiteConfig{
			DSN: "file:test2.db?cache=shared&mode=rwc",
		}
		mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
		if err != nil {
			t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
		}
		defer mgr.Close()
	})
}

func TestDatabaseManager_MigrationEdgeCases(t *testing.T) {
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
		ID        uint
		Name      string
		Email     string
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	t.Run("AutoMigrate with multiple models", func(t *testing.T) {
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

		err := mgr.AutoMigrate(&User{}, &Product{}, &Order{})
		if err != nil {
			t.Errorf("AutoMigrate() error = %v", err)
		}
	})

	t.Run("AutoMigrate with same model twice", func(t *testing.T) {
		err := mgr.AutoMigrate(&User{})
		if err != nil {
			t.Errorf("AutoMigrate() error = %v", err)
		}
	})

	t.Run("Migrator operations", func(t *testing.T) {
		migrator := mgr.Migrator()
		if migrator == nil {
			t.Fatal("Migrator() returned nil")
		}

		hasTable := migrator.HasTable(&User{})
		if !hasTable {
			t.Error("Table should exist")
		}
	})
}

func TestDatabaseManager_QueryEdgeCases(t *testing.T) {
	skipIfCGONotAvailable(t)

	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	type Item struct {
		ID   uint
		Name string
	}

	if err := mgr.AutoMigrate(&Item{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	db := mgr.DB()

	t.Run("Find with no results", func(t *testing.T) {
		var items []Item
		if err := db.Find(&items).Error; err != nil {
			t.Errorf("Find() error = %v", err)
		}
		if len(items) != 0 {
			t.Errorf("Expected 0 items, got %d", len(items))
		}
	})

	t.Run("First with no results", func(t *testing.T) {
		var item Item
		err := db.First(&item).Error
		if err != gorm.ErrRecordNotFound {
			t.Errorf("Expected ErrRecordNotFound, got %v", err)
		}
	})

	t.Run("Find with empty slice", func(t *testing.T) {
		var items []Item
		if err := db.Where("name = ?", "nonexistent").Find(&items).Error; err != nil {
			t.Errorf("Find() error = %v", err)
		}
		if len(items) != 0 {
			t.Errorf("Expected 0 items, got %d", len(items))
		}
	})
}

func TestDatabaseManager_ConnectionEdgeCases(t *testing.T) {
	skipIfCGONotAvailable(t)

	t.Run("MaxOpenConns zero", func(t *testing.T) {
		cfg := &SQLiteConfig{
			DSN: ":memory:",
			PoolConfig: &PoolConfig{
				MaxOpenConns:    0,
				MaxIdleConns:    0,
				ConnMaxLifetime: 0,
				ConnMaxIdleTime: 0,
			},
		}

		mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
		if err != nil {
			t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
		}
		defer mgr.Close()

		stats := mgr.Stats()
		if stats.MaxOpenConnections != 0 {
			t.Errorf("MaxOpenConnections = %v, want 0", stats.MaxOpenConnections)
		}
	})

	t.Run("Ping with timeout", func(t *testing.T) {
		cfg := &SQLiteConfig{
			DSN: ":memory:",
		}

		mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
		if err != nil {
			t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
		}
		defer mgr.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(10 * time.Millisecond)

		err = mgr.Ping(ctx)
		if err == nil {
			t.Error("Expected error for expired context")
		}
	})
}

func TestDatabaseManager_TransactionEdgeCases(t *testing.T) {
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

	t.Run("Transaction with tx options", func(t *testing.T) {
		opts := &sql.TxOptions{
			Isolation: sql.LevelSerializable,
			ReadOnly:  false,
		}
		err := mgr.Transaction(func(tx *gorm.DB) error {
			return nil
		}, opts)
		if err != nil {
			t.Errorf("Transaction() error = %v", err)
		}
	})

	t.Run("Begin with tx options", func(t *testing.T) {
		opts := &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly:  true,
		}
		tx := mgr.Begin(opts)
		if tx == nil {
			t.Fatal("Begin() returned nil")
		}
		tx.Rollback()
	})
}

func TestDatabaseManager_RawSQLEdgeCases(t *testing.T) {
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

	t.Run("Raw with no results", func(t *testing.T) {
		var results []User
		result := mgr.Raw("SELECT * FROM users WHERE 1=0").Scan(&results)
		if result.Error != nil {
			t.Errorf("Raw() error = %v", result.Error)
		}
		if len(results) != 0 {
			t.Errorf("Expected 0 results, got %d", len(results))
		}
	})

	t.Run("Exec with no rows affected", func(t *testing.T) {
		result := mgr.Exec("UPDATE users SET name = 'test' WHERE 1=0")
		if result.Error != nil {
			t.Errorf("Exec() error = %v", result.Error)
		}
		if result.RowsAffected != 0 {
			t.Errorf("RowsAffected = %v, want 0", result.RowsAffected)
		}
	})
}

func TestDatabaseManager_Concurrency(t *testing.T) {
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

	t.Run("Concurrent DB calls", func(t *testing.T) {
		done := make(chan bool, 5)
		for i := 0; i < 5; i++ {
			go func() {
				var users []User
				db.Find(&users)
				done <- true
			}()
		}
		for i := 0; i < 5; i++ {
			<-done
		}
	})

	t.Run("Concurrent transactions", func(t *testing.T) {
		done := make(chan bool, 3)
		for i := 0; i < 3; i++ {
			go func() {
				mgr.Transaction(func(tx *gorm.DB) error {
					var users []User
					return tx.Find(&users).Error
				})
				done <- true
			}()
		}
		for i := 0; i < 3; i++ {
			<-done
		}
	})
}

func TestDatabaseManager_DriverInfo(t *testing.T) {
	skipIfCGONotAvailable(t)

	cfg := &SQLiteConfig{
		DSN: ":memory:",
	}

	mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	defer mgr.Close()

	t.Run("Driver returns sqlite", func(t *testing.T) {
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

func TestDatabaseManager_Lifecycle(t *testing.T) {
	skipIfCGONotAvailable(t)

	t.Run("OnStart returns no error", func(t *testing.T) {
		cfg := &SQLiteConfig{
			DSN: ":memory:",
		}

		mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
		if err != nil {
			t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
		}
		defer mgr.Close()

		if err := mgr.OnStart(); err != nil {
			t.Errorf("OnStart() error = %v", err)
		}
	})

	t.Run("OnStop returns no error", func(t *testing.T) {
		cfg := &SQLiteConfig{
			DSN: ":memory:",
		}

		mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
		if err != nil {
			t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
		}

		if err := mgr.OnStop(); err != nil {
			t.Errorf("OnStop() error = %v", err)
		}
	})

	t.Run("Close can be called multiple times", func(t *testing.T) {
		cfg := &SQLiteConfig{
			DSN: ":memory:",
		}

		mgr, err := NewDatabaseManagerSQLiteImpl(cfg, nil, nil)
		if err != nil {
			t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
		}

		if err := mgr.Close(); err != nil {
			t.Errorf("First Close() error = %v", err)
		}

		if err := mgr.Close(); err != nil {
			t.Errorf("Second Close() error = %v", err)
		}
	})
}
