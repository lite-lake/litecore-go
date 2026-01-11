package drivers

import (
	"context"
	"database/sql"
	"testing"
)

func TestNewNoneDatabaseManager(t *testing.T) {
	mgr := NewNoneDatabaseManager()

	if mgr == nil {
		t.Fatal("NewNoneDatabaseManager() returned nil")
	}

	if mgr.BaseManager == nil {
		t.Error("NoneDatabaseManager.BaseManager is nil")
	}
}

func TestNoneDatabaseManager_ManagerName(t *testing.T) {
	mgr := NewNoneDatabaseManager()

	expected := "none-database"
	if got := mgr.ManagerName(); got != expected {
		t.Errorf("NoneDatabaseManager.ManagerName() = %v, want %v", got, expected)
	}
}

func TestNoneDatabaseManager_DB(t *testing.T) {
	mgr := NewNoneDatabaseManager()

	db := mgr.DB()
	if db != nil {
		t.Error("NoneDatabaseManager.DB() should return nil")
	}
}

func TestNoneDatabaseManager_Driver(t *testing.T) {
	mgr := NewNoneDatabaseManager()

	expected := "none"
	if got := mgr.Driver(); got != expected {
		t.Errorf("NoneDatabaseManager.Driver() = %v, want %v", got, expected)
	}
}

func TestNoneDatabaseManager_Ping(t *testing.T) {
	mgr := NewNoneDatabaseManager()

	ctx := context.Background()
	err := mgr.Ping(ctx)

	if err == nil {
		t.Error("NoneDatabaseManager.Ping() should return error")
	}

	expectedErrMsg := "database not available"
	if err.Error()[:len(expectedErrMsg)] != expectedErrMsg {
		t.Errorf("NoneDatabaseManager.Ping() error = %v, want prefix %v", err.Error(), expectedErrMsg)
	}
}

func TestNoneDatabaseManager_BeginTx(t *testing.T) {
	mgr := NewNoneDatabaseManager()

	ctx := context.Background()
	tx, err := mgr.BeginTx(ctx, nil)

	if tx != nil {
		t.Error("NoneDatabaseManager.BeginTx() should return nil transaction")
	}

	if err == nil {
		t.Error("NoneDatabaseManager.BeginTx() should return error")
	}

	expectedErrMsg := "database not available"
	if err.Error()[:len(expectedErrMsg)] != expectedErrMsg {
		t.Errorf("NoneDatabaseManager.BeginTx() error = %v, want prefix %v", err.Error(), expectedErrMsg)
	}
}

func TestNoneDatabaseManager_Stats(t *testing.T) {
	mgr := NewNoneDatabaseManager()

	stats := mgr.Stats()

	expected := sql.DBStats{}
	if stats != expected {
		t.Errorf("NoneDatabaseManager.Stats() = %v, want %v", stats, expected)
	}
}

func TestNoneDatabaseManager_Close(t *testing.T) {
	mgr := NewNoneDatabaseManager()

	err := mgr.Close()
	if err != nil {
		t.Errorf("NoneDatabaseManager.Close() error = %v, want nil", err)
	}
}

func TestNoneDatabaseManager_Health(t *testing.T) {
	mgr := NewNoneDatabaseManager()

	err := mgr.Health()
	if err == nil {
		t.Error("NoneDatabaseManager.Health() should return error")
	}

	expectedErrMsg := "database not available"
	if err.Error()[:len(expectedErrMsg)] != expectedErrMsg {
		t.Errorf("NoneDatabaseManager.Health() error = %v, want prefix %v", err.Error(), expectedErrMsg)
	}
}

func TestNoneDatabaseManager_Shutdown(t *testing.T) {
	mgr := NewNoneDatabaseManager()

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
			wantErr: false, // NoneDatabaseManager doesn't validate context
		},
		{
			name:    "cancelled context",
			ctx:     func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.Shutdown(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("NoneDatabaseManager.Shutdown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNoneDatabaseManager_CommonManager(t *testing.T) {
	mgr := NewNoneDatabaseManager()

	// Test common.Manager interface methods
	if err := mgr.OnStart(); err != nil {
		t.Errorf("NoneDatabaseManager.OnStart() error = %v, want nil", err)
	}

	if err := mgr.OnStop(); err != nil {
		t.Errorf("NoneDatabaseManager.OnStop() error = %v, want nil", err)
	}
}

func TestNoneDatabaseManager_ConcurrentUsage(t *testing.T) {
	mgr := NewNoneDatabaseManager()
	done := make(chan bool)

	// Concurrent calls to ensure thread safety
	for i := 0; i < 100; i++ {
		go func(i int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Concurrent usage panicked: %v", r)
				}
			}()

			switch i % 4 {
			case 0:
				_ = mgr.DB()
			case 1:
				_ = mgr.Driver()
			case 2:
				_ = mgr.Stats()
			case 3:
				_ = mgr.Close()
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}
