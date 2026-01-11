package drivers

import (
	"context"
	"testing"

	"com.litelake.litecore/common"
)

func TestNewBaseManager(t *testing.T) {
	name := "test-manager"
	bm := NewBaseManager(name)

	if bm == nil {
		t.Fatal("NewBaseManager() returned nil")
	}
	if bm.name != name {
		t.Errorf("NewBaseManager() name = %v, want %v", bm.name, name)
	}
}

func TestBaseManager_ManagerName(t *testing.T) {
	name := "test-manager"
	bm := NewBaseManager(name)

	if got := bm.ManagerName(); got != name {
		t.Errorf("BaseManager.ManagerName() = %v, want %v", got, name)
	}
}

func TestBaseManager_Health(t *testing.T) {
	bm := NewBaseManager("test-manager")

	if err := bm.Health(); err != nil {
		t.Errorf("BaseManager.Health() error = %v, want nil", err)
	}
}

func TestBaseManager_OnStart(t *testing.T) {
	bm := NewBaseManager("test-manager")

	if err := bm.OnStart(); err != nil {
		t.Errorf("BaseManager.OnStart() error = %v, want nil", err)
	}
}

func TestBaseManager_OnStop(t *testing.T) {
	bm := NewBaseManager("test-manager")

	if err := bm.OnStop(); err != nil {
		t.Errorf("BaseManager.OnStop() error = %v, want nil", err)
	}
}

func TestBaseManager_Shutdown(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "valid context",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "nil context",
			ctx:     nil,
			wantErr: false, // BaseManager.Shutdown doesn't validate context
		},
		{
			name:    "cancelled context",
			ctx:     func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm := NewBaseManager("test-manager")
			if err := bm.Shutdown(tt.ctx); (err != nil) != tt.wantErr {
				t.Errorf("BaseManager.Shutdown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "valid context",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "nil context",
			ctx:     nil,
			wantErr: true,
		},
		{
			name:    "context with timeout",
			ctx:     func() context.Context { ctx, _ := context.WithTimeout(context.Background(), 10); return ctx }(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateContext(tt.ctx); (err != nil) != tt.wantErr {
				t.Errorf("ValidateContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBaseManagerImplementsManagerInterface(t *testing.T) {
	// Compile-time check that BaseManager implements common.BaseManager
	var _ common.BaseManager = (*BaseManager)(nil)
}
