package common

import (
	"errors"
	"testing"

	"github.com/lite-lake/litecore-go/util/logger"
	"github.com/stretchr/testify/assert"
)

type mockService struct {
	name   string
	Logger logger.ILogger `inject:""`
}

func (m *mockService) ServiceName() string {
	return m.name
}

func (m *mockService) OnStart() error {
	return nil
}

func (m *mockService) OnStop() error {
	return nil
}

type failingService struct {
	Logger logger.ILogger `inject:""`
}

func (f *failingService) ServiceName() string {
	return "FailingService"
}

func (f *failingService) OnStart() error {
	return errors.New("服务启动失败")
}

func (f *failingService) OnStop() error {
	return errors.New("服务停止失败")
}

func TestIBaseService_基础接口实现(t *testing.T) {
	service := &mockService{
		name: "TestService",
	}

	assert.Equal(t, "TestService", service.ServiceName())
	assert.NoError(t, service.OnStart())
	assert.NoError(t, service.OnStop())
}

func TestIBaseService_生命周期方法(t *testing.T) {
	tests := []struct {
		name    string
		service IBaseService
		wantErr bool
	}{
		{
			name:    "正常启动和停止",
			service: &mockService{name: "LifecycleTest"},
			wantErr: false,
		},
		{
			name:    "启动失败的服务",
			service: &failingService{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.service.OnStart()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			err = tt.service.OnStop()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIBaseService_空实现(t *testing.T) {
	tests := []struct {
		name    string
		service IBaseService
	}{
		{
			name:    "空服务实例",
			service: &mockService{},
		},
		{
			name:    "带有空名称的服务",
			service: &mockService{name: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.service.ServiceName())
		})
	}
}

func TestIBaseService_接口组合(t *testing.T) {
	service := &mockService{
		name: "CombinedService",
	}

	var iface IBaseService = service
	assert.Equal(t, "CombinedService", iface.ServiceName())
	assert.NoError(t, iface.OnStart())
}
