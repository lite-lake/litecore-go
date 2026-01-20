package common

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockManager struct {
	name    string
	healthy bool
}

func (m *mockManager) ManagerName() string {
	return m.name
}

func (m *mockManager) Health() error {
	if m.healthy {
		return nil
	}
	return errors.New("管理器不健康")
}

func (m *mockManager) OnStart() error {
	return nil
}

func (m *mockManager) OnStop() error {
	return nil
}

type failingManager struct{}

func (f *failingManager) ManagerName() string {
	return "FailingManager"
}

func (f *failingManager) Health() error {
	return errors.New("健康检查失败")
}

func (f *failingManager) OnStart() error {
	return errors.New("启动失败")
}

func (f *failingManager) OnStop() error {
	return errors.New("停止失败")
}

func TestIBaseManager_基础接口实现(t *testing.T) {
	manager := &mockManager{
		name:    "TestManager",
		healthy: true,
	}

	assert.Equal(t, "TestManager", manager.ManagerName())
	assert.NoError(t, manager.Health())
	assert.NoError(t, manager.OnStart())
	assert.NoError(t, manager.OnStop())
}

func TestIBaseManager_健康检查(t *testing.T) {
	tests := []struct {
		name    string
		manager IBaseManager
		wantErr bool
	}{
		{
			name:    "健康的管理器",
			manager: &mockManager{name: "Healthy", healthy: true},
			wantErr: false,
		},
		{
			name:    "不健康的管理器",
			manager: &mockManager{name: "Unhealthy", healthy: false},
			wantErr: true,
		},
		{
			name:    "健康检查失败的管理器",
			manager: &failingManager{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.manager.Health()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIBaseManager_生命周期方法(t *testing.T) {
	tests := []struct {
		name    string
		manager IBaseManager
		wantErr bool
	}{
		{
			name:    "正常启动和停止",
			manager: &mockManager{name: "LifecycleTest"},
			wantErr: false,
		},
		{
			name:    "启动失败的管理器",
			manager: &failingManager{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.manager.OnStart()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			err = tt.manager.OnStop()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIBaseManager_接口组合(t *testing.T) {
	manager := &mockManager{
		name:    "CombinedManager",
		healthy: true,
	}

	var iface IBaseManager = manager
	assert.Equal(t, "CombinedManager", iface.ManagerName())
	assert.NoError(t, iface.Health())
}
