package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLayerString(t *testing.T) {
	tests := []struct {
		layer Layer
		want  string
	}{
		{LayerConfig, "config"},
		{LayerEntity, "entity"},
		{LayerManager, "manager"},
		{LayerRepository, "repository"},
		{LayerService, "service"},
		{LayerController, "controller"},
		{LayerMiddleware, "middleware"},
	}

	for _, tt := range tests {
		t.Run(string(tt.layer), func(t *testing.T) {
			assert.Equal(t, tt.want, string(tt.layer))
		})
	}
}

func TestIsLitecoreLayer(t *testing.T) {
	tests := []struct {
		name  string
		layer Layer
		want  bool
	}{
		{"配置层", LayerConfig, true},
		{"实体层", LayerEntity, true},
		{"管理器层", LayerManager, true},
		{"仓储层", LayerRepository, true},
		{"服务层", LayerService, true},
		{"控制器层", LayerController, true},
		{"中间件层", LayerMiddleware, true},
		{"未知层", Layer("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsLitecoreLayer(tt.layer))
		})
	}
}

func TestGetBaseInterface(t *testing.T) {
	tests := []struct {
		name  string
		layer Layer
		want  string
	}{
		{"配置层", LayerConfig, "BaseConfigProvider"},
		{"实体层", LayerEntity, "BaseEntity"},
		{"管理器层", LayerManager, "BaseManager"},
		{"仓储层", LayerRepository, "BaseRepository"},
		{"服务层", LayerService, "BaseService"},
		{"控制器层", LayerController, "BaseController"},
		{"中间件层", LayerMiddleware, "BaseMiddleware"},
		{"未知层", Layer("unknown"), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetBaseInterface(tt.layer))
		})
	}
}

func TestGetContainerName(t *testing.T) {
	tests := []struct {
		name  string
		layer Layer
		want  string
	}{
		{"配置层", LayerConfig, "ConfigContainer"},
		{"实体层", LayerEntity, "EntityContainer"},
		{"管理器层", LayerManager, "ManagerContainer"},
		{"仓储层", LayerRepository, "RepositoryContainer"},
		{"服务层", LayerService, "ServiceContainer"},
		{"控制器层", LayerController, "ControllerContainer"},
		{"中间件层", LayerMiddleware, "MiddlewareContainer"},
		{"未知层", Layer("unknown"), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetContainerName(tt.layer))
		})
	}
}

func TestGetRegisterFunction(t *testing.T) {
	tests := []struct {
		name  string
		layer Layer
		want  string
	}{
		{"配置层", LayerConfig, "RegisterConfig"},
		{"实体层", LayerEntity, "RegisterEntity"},
		{"管理器层", LayerManager, "RegisterManager"},
		{"仓储层", LayerRepository, "RegisterRepository"},
		{"服务层", LayerService, "RegisterService"},
		{"控制器层", LayerController, "RegisterController"},
		{"中间件层", LayerMiddleware, "RegisterMiddleware"},
		{"未知层", Layer("unknown"), "Register"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetRegisterFunction(tt.layer))
		})
	}
}
