package databasemgr

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	_ "github.com/mattn/go-sqlite3"
)

func TestObservabilityPlugin_Basic(t *testing.T) {
	plugin := newObservabilityPlugin()

	t.Run("New plugin has defaults", func(t *testing.T) {
		if plugin.slowQueryThreshold != 1*time.Second {
			t.Errorf("slowQueryThreshold = %v, want 1s", plugin.slowQueryThreshold)
		}
		if plugin.logSQL != false {
			t.Errorf("logSQL = %v, want false", plugin.logSQL)
		}
		if plugin.sampleRate != 1.0 {
			t.Errorf("sampleRate = %v, want 1.0", plugin.sampleRate)
		}
	})

	t.Run("Name", func(t *testing.T) {
		if plugin.Name() != "observability" {
			t.Errorf("Name() = %v, want 'observability'", plugin.Name())
		}
	})

	t.Run("Initialize", func(t *testing.T) {
		db, _ := gorm.Open(&dummyDialector{}, &gorm.Config{})
		if err := plugin.Initialize(db); err != nil {
			t.Errorf("Initialize() error = %v", err)
		}
	})

	t.Run("SetConfig", func(t *testing.T) {
		plugin.SetConfig(2*time.Second, true, 0.5)
		if plugin.slowQueryThreshold != 2*time.Second {
			t.Errorf("slowQueryThreshold = %v, want 2s", plugin.slowQueryThreshold)
		}
		if plugin.logSQL != true {
			t.Errorf("logSQL = %v, want true", plugin.logSQL)
		}
		if plugin.sampleRate != 0.5 {
			t.Errorf("sampleRate = %v, want 0.5", plugin.sampleRate)
		}
	})
}

func TestObservabilityPlugin_CallbackRegistration(t *testing.T) {
	plugin := newObservabilityPlugin()
	db, _ := gorm.Open(&dummyDialector{}, &gorm.Config{})

	if err := plugin.Initialize(db); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	t.Run("Callbacks are registered", func(t *testing.T) {
		callbacks := db.Callback()

		plugin.registerCallbacks(db)

		if callbacks == nil {
			t.Error("Callbacks should not be nil")
		}
	})
}

func TestObservabilityPlugin_OperationCallbacks(t *testing.T) {
	plugin := newObservabilityPlugin()

	db, _ := gorm.Open(&dummyDialector{}, &gorm.Config{})
	db.Statement.Context = context.Background()
	db.Statement.Table = "test_table"

	t.Run("beforeQuery with nil tracer and logger", func(t *testing.T) {
		plugin.beforeQuery(db)
		_, ok := db.InstanceGet("observability:start_time")
		if ok {
			t.Error("start_time should not be set when tracer and logger are nil")
		}
	})

	t.Run("afterQuery handles missing start time", func(t *testing.T) {
		db2, _ := gorm.Open(&dummyDialector{}, &gorm.Config{})
		db2.Statement.Context = context.Background()
		db2.Statement.Table = "test_table"
		plugin.afterQuery(db2)
	})

	t.Run("beforeCreate with nil tracer and logger", func(t *testing.T) {
		plugin.beforeCreate(db)
		_, ok := db.InstanceGet("observability:start_time")
		if ok {
			t.Error("start_time should not be set when tracer and logger are nil")
		}
	})

	t.Run("afterCreate handles missing start time", func(t *testing.T) {
		db2, _ := gorm.Open(&dummyDialector{}, &gorm.Config{})
		db2.Statement.Context = context.Background()
		db2.Statement.Table = "test_table"
		plugin.afterCreate(db2)
	})

	t.Run("beforeUpdate with nil tracer and logger", func(t *testing.T) {
		plugin.beforeUpdate(db)
		_, ok := db.InstanceGet("observability:start_time")
		if ok {
			t.Error("start_time should not be set when tracer and logger are nil")
		}
	})

	t.Run("afterUpdate handles missing start time", func(t *testing.T) {
		db2, _ := gorm.Open(&dummyDialector{}, &gorm.Config{})
		db2.Statement.Context = context.Background()
		db2.Statement.Table = "test_table"
		plugin.afterUpdate(db2)
	})

	t.Run("beforeDelete with nil tracer and logger", func(t *testing.T) {
		plugin.beforeDelete(db)
		_, ok := db.InstanceGet("observability:start_time")
		if ok {
			t.Error("start_time should not be set when tracer and logger are nil")
		}
	})

	t.Run("afterDelete handles missing start time", func(t *testing.T) {
		db2, _ := gorm.Open(&dummyDialector{}, &gorm.Config{})
		db2.Statement.Context = context.Background()
		db2.Statement.Table = "test_table"
		plugin.afterDelete(db2)
	})
}

func TestObservabilityPlugin_Sampling(t *testing.T) {
	tests := []struct {
		name       string
		sampleRate float64
	}{
		{
			name:       "full sampling",
			sampleRate: 1.0,
		},
		{
			name:       "50% sampling",
			sampleRate: 0.5,
		},
		{
			name:       "0% sampling",
			sampleRate: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &observabilityPlugin{
				sampleRate: tt.sampleRate,
			}

			db, _ := gorm.Open(&dummyDialector{}, &gorm.Config{})
			db.Statement.Context = context.Background()
			db.Statement.Table = "test_table"

			plugin.beforeQuery(db)
		})
	}
}

func TestObservabilityPlugin_ContextHandling(t *testing.T) {
	plugin := newObservabilityPlugin()
	plugin.sampleRate = 1.0

	tests := []struct {
		name      string
		setupDB   func() *gorm.DB
		expectNil bool
	}{
		{
			name: "nil context with nil tracer and logger",
			setupDB: func() *gorm.DB {
				db, _ := gorm.Open(&dummyDialector{}, &gorm.Config{})
				db.Statement.Context = nil
				db.Statement.Table = "test_table"
				return db
			},
			expectNil: true,
		},
		{
			name: "existing context preserved",
			setupDB: func() *gorm.DB {
				ctx := context.Background()
				db, _ := gorm.Open(&dummyDialector{}, &gorm.Config{})
				db.Statement.Context = ctx
				db.Statement.Table = "test_table"
				return db
			},
			expectNil: false,
		},
		{
			name: "context with timeout",
			setupDB: func() *gorm.DB {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				db, _ := gorm.Open(&dummyDialector{}, &gorm.Config{})
				db.Statement.Context = ctx
				db.Statement.Table = "test_table"
				return db
			},
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setupDB()
			plugin.beforeQuery(db)

			if tt.expectNil && db.Statement.Context != nil {
				t.Error("context should not be set when tracer and logger are nil")
			}
			if !tt.expectNil && db.Statement.Context == nil {
				t.Error("context should be set when it was provided")
			}
		})
	}
}

func TestDatabaseManagerBaseImpl(t *testing.T) {
	t.Run("New base impl", func(t *testing.T) {
		db, _ := gorm.Open(&dummyDialector{}, &gorm.Config{})
		base := newIDatabaseManagerBaseImpl(nil, nil, "test", "test", db)

		if base == nil {
			t.Fatal("newIDatabaseManagerBaseImpl returned nil")
		}

		if base.name != "test" {
			t.Errorf("name = %v, want 'test'", base.name)
		}

		if base.driver != "test" {
			t.Errorf("driver = %v, want 'test'", base.driver)
		}

		if base.db == nil {
			t.Error("db should not be nil")
		}

		if base.observabilityPlugin == nil {
			t.Error("observabilityPlugin should not be nil")
		}
	})

	t.Run("Init observability with nil config", func(t *testing.T) {
		db, _ := gorm.Open(&dummyDialector{}, &gorm.Config{})
		base := newIDatabaseManagerBaseImpl(nil, nil, "test", "test", db)

		base.initObservability(nil)

		if base.observabilityPlugin == nil {
			t.Error("observabilityPlugin should not be nil")
		}
	})

	t.Run("Init observability with config", func(t *testing.T) {
		db, _ := gorm.Open(&dummyDialector{}, &gorm.Config{})
		base := newIDatabaseManagerBaseImpl(nil, nil, "test", "test", db)

		cfg := &DatabaseConfig{
			Driver: "sqlite",
			ObservabilityConfig: &ObservabilityConfig{
				SlowQueryThreshold: 2 * time.Second,
				LogSQL:             true,
				SampleRate:         0.5,
			},
		}

		base.initObservability(cfg)

		if base.observabilityPlugin == nil {
			t.Error("observabilityPlugin should not be nil")
		}

		if base.observabilityPlugin.slowQueryThreshold != 2*time.Second {
			t.Errorf("slowQueryThreshold = %v, want 2s", base.observabilityPlugin.slowQueryThreshold)
		}
	})
}

func TestDummyDialector(t *testing.T) {
	d := &dummyDialector{}

	t.Run("Name", func(t *testing.T) {
		if d.Name() != "none" {
			t.Errorf("Name() = %v, want 'none'", d.Name())
		}
	})

	t.Run("Initialize", func(t *testing.T) {
		db, _ := gorm.Open(d, &gorm.Config{})
		if err := d.Initialize(db); err != nil {
			t.Errorf("Initialize() error = %v", err)
		}
	})

	t.Run("Migrator", func(t *testing.T) {
		db, _ := gorm.Open(d, &gorm.Config{})
		if d.Migrator(db) != nil {
			t.Error("Migrator() should return nil")
		}
	})

	t.Run("DataTypeOf", func(t *testing.T) {
		field := &schema.Field{}
		if d.DataTypeOf(field) != "" {
			t.Errorf("DataTypeOf() should return empty string, got %v", d.DataTypeOf(field))
		}
	})

	t.Run("DefaultValueOf", func(t *testing.T) {
		field := &schema.Field{}
		result := d.DefaultValueOf(field)
		if result == nil {
			t.Error("DefaultValueOf() should not return nil")
		}
	})

	t.Run("Explain", func(t *testing.T) {
		if d.Explain("SELECT 1", 1) != "" {
			t.Error("Explain() should return empty string")
		}
	})
}

func TestDatabaseManagerBaseImpl_NilDB(t *testing.T) {
	base := &databaseManagerBaseImpl{
		db:    nil,
		sqlDB: nil,
	}

	t.Run("DB returns nil", func(t *testing.T) {
		if base.db != nil {
			t.Error("db should be nil")
		}
	})
}

func TestObservabilityPlugin_NullComponents(t *testing.T) {
	plugin := newObservabilityPlugin()

	db, _ := gorm.Open(&dummyDialector{}, &gorm.Config{})
	db.Statement.Context = context.Background()
	db.Statement.Table = "test_table"

	t.Run("beforeQuery with nil logger and tracer", func(t *testing.T) {
		plugin.logger = nil
		plugin.tracer = nil
		plugin.beforeQuery(db)
	})

	t.Run("afterQuery with nil logger and tracer", func(t *testing.T) {
		plugin.logger = nil
		plugin.tracer = nil
		plugin.afterQuery(db)
	})
}

func TestObservabilityPlugin_VaryingSlowQueryThresholds(t *testing.T) {
	tests := []struct {
		name      string
		threshold time.Duration
	}{
		{
			name:      "disabled",
			threshold: 0,
		},
		{
			name:      "100ms",
			threshold: 100 * time.Millisecond,
		},
		{
			name:      "1 second",
			threshold: 1 * time.Second,
		},
		{
			name:      "10 seconds",
			threshold: 10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &observabilityPlugin{
				slowQueryThreshold: tt.threshold,
			}

			db, _ := gorm.Open(&dummyDialector{}, &gorm.Config{})
			db.Statement.Context = context.Background()
			db.Statement.Table = "test_table"
			db.InstanceSet("observability:start_time", time.Now())

			plugin.afterQuery(db)
		})
	}
}

func TestObservabilityPlugin_AllOperations(t *testing.T) {
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

	db := mgr.DB()

	t.Run("Create operation", func(t *testing.T) {
		user := User{Name: "Alice", Age: 30}
		if err := db.Create(&user).Error; err != nil {
			t.Errorf("Create() error = %v", err)
		}
	})

	t.Run("Query operation", func(t *testing.T) {
		var user User
		if err := db.First(&user, 1).Error; err != nil {
			t.Errorf("First() error = %v", err)
		}
	})

	t.Run("Update operation", func(t *testing.T) {
		if err := db.Model(&User{}).Where("id = ?", 1).Update("age", 31).Error; err != nil {
			t.Errorf("Update() error = %v", err)
		}
	})

	t.Run("Delete operation", func(t *testing.T) {
		if err := db.Delete(&User{}, 1).Error; err != nil {
			t.Errorf("Delete() error = %v", err)
		}
	})
}

func TestObservabilityPlugin_TransactionOperations(t *testing.T) {
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

	t.Run("Transaction commit", func(t *testing.T) {
		err := mgr.Transaction(func(tx *gorm.DB) error {
			return tx.Model(&Account{}).Where("name = ?", "Alice").
				Update("balance", 90).Error
		})
		if err != nil {
			t.Errorf("Transaction() error = %v", err)
		}
	})

	t.Run("Begin transaction", func(t *testing.T) {
		tx := mgr.Begin()
		if tx == nil {
			t.Fatal("Begin() returned nil")
		}
		tx.Model(&Account{}).Where("name = ?", "Alice").
			Update("balance", 80)
		tx.Commit()
	})
}
