package common

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockRepository struct {
	name string
}

func (m *mockRepository) RepositoryName() string {
	return m.name
}

func (m *mockRepository) OnStart() error {
	return nil
}

func (m *mockRepository) OnStop() error {
	return nil
}

type failingRepository struct{}

func (f *failingRepository) RepositoryName() string {
	return "FailingRepository"
}

func (f *failingRepository) OnStart() error {
	return errors.New("存储库启动失败")
}

func (f *failingRepository) OnStop() error {
	return errors.New("存储库停止失败")
}

func TestIBaseRepository_基础接口实现(t *testing.T) {
	repo := &mockRepository{
		name: "TestRepository",
	}

	assert.Equal(t, "TestRepository", repo.RepositoryName())
	assert.NoError(t, repo.OnStart())
	assert.NoError(t, repo.OnStop())
}

func TestIBaseRepository_生命周期方法(t *testing.T) {
	tests := []struct {
		name    string
		repo    IBaseRepository
		wantErr bool
	}{
		{
			name:    "正常启动和停止",
			repo:    &mockRepository{name: "LifecycleTest"},
			wantErr: false,
		},
		{
			name:    "启动失败的存储库",
			repo:    &failingRepository{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.repo.OnStart()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			err = tt.repo.OnStop()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIBaseRepository_空实现(t *testing.T) {
	tests := []struct {
		name string
		repo IBaseRepository
	}{
		{
			name: "空存储库实例",
			repo: &mockRepository{},
		},
		{
			name: "带有空名称的存储库",
			repo: &mockRepository{name: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.repo.RepositoryName())
		})
	}
}

func TestIBaseRepository_接口组合(t *testing.T) {
	repo := &mockRepository{
		name: "CombinedRepository",
	}

	var iface IBaseRepository = repo
	assert.Equal(t, "CombinedRepository", iface.RepositoryName())
	assert.NoError(t, iface.OnStart())
}
