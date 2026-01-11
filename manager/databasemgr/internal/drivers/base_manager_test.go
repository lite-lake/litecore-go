package drivers

import (
	"context"
	"testing"
)

func TestNewBaseManager(t *testing.T) {
	name := "test-manager"
	mgr := NewBaseManager(name)

	if mgr == nil {
		t.Fatal("NewBaseManager() returned nil")
	}

	if mgr.ManagerName() != name {
		t.Errorf("ManagerName() = %v, want %v", mgr.ManagerName(), name)
	}
}

func TestBaseManager_ManagerName(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"test-manager-1"},
		{"test-manager-2"},
		{""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewBaseManager(tt.name)
			if got := mgr.ManagerName(); got != tt.name {
				t.Errorf("BaseManager.ManagerName() = %v, want %v", got, tt.name)
			}
		})
	}
}

func TestBaseManager_Health(t *testing.T) {
	mgr := NewBaseManager("test-manager")

	if err := mgr.Health(); err != nil {
		t.Errorf("BaseManager.Health() error = %v, want nil", err)
	}
}

func TestBaseManager_OnStart(t *testing.T) {
	mgr := NewBaseManager("test-manager")

	if err := mgr.OnStart(); err != nil {
		t.Errorf("BaseManager.OnStart() error = %v, want nil", err)
	}
}

func TestBaseManager_OnStop(t *testing.T) {
	mgr := NewBaseManager("test-manager")

	if err := mgr.OnStop(); err != nil {
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
			name:    "normal context",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "nil context",
			ctx:     nil,
			wantErr: false, // BaseManager.Shutdown doesn't validate context
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewBaseManager("test-manager")
			err := mgr.Shutdown(tt.ctx)
			if (err != nil) != tt.wantErr {
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
			ctx:     context.TODO(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateContext(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
