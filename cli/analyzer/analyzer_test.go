package analyzer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestNewAnalyzer(t *testing.T) {
	t.Run("创建分析器", func(t *testing.T) {
		analyzer := NewAnalyzer("/test/path", "test.module")
		assert.NotNil(t, analyzer)
		assert.Equal(t, "/test/path", analyzer.projectPath)
		assert.Equal(t, "test.module", analyzer.moduleName)
		assert.NotNil(t, analyzer.info)
		assert.Equal(t, "test.module", analyzer.info.ModuleName)
		assert.NotNil(t, analyzer.info.Layers)
	})
}

func TestAnalyze(t *testing.T) {
	tests := []struct {
		name       string
		projectDir string
		moduleName string
		wantErr    bool
	}{
		{"有效项目", "testdata/valid_project", "test.module", false},
		{"无效目录", "testdata/nonexistent", "test.module", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewAnalyzer(tt.projectDir, tt.moduleName)
			info, err := analyzer.Analyze()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, info)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, info)
				assert.Equal(t, tt.moduleName, info.ModuleName)
			}
		})
	}
}

func TestDetectLayer(t *testing.T) {
	analyzer := NewAnalyzer("/test", "test.module")

	tests := []struct {
		name        string
		filename    string
		packageName string
		wantLayer   Layer
	}{
		{"实体层", "internal/entities/entity.go", "entities", LayerEntity},
		{"仓储层", "internal/repositories/repo.go", "repositories", LayerRepository},
		{"服务层", "internal/services/service.go", "services", LayerService},
		{"控制器层", "internal/controllers/controller.go", "controllers", LayerController},
		{"中间件层", "internal/middlewares/middleware.go", "middlewares", LayerMiddleware},
		{"配置层", "infras/config_provider.go", "configproviders", LayerConfig},
		{"管理器层", "infras/managers/manager.go", "managers", LayerManager},
		{"未知层", "internal/unknown/unknown.go", "unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layer := analyzer.detectLayer(tt.filename, tt.packageName)
			assert.Equal(t, tt.wantLayer, layer)
		})
	}
}

func TestGetPackagePath(t *testing.T) {
	analyzer := NewAnalyzer("/test/project", "test.module")

	tests := []struct {
		name     string
		filename string
		wantPath string
	}{
		{"根目录文件", "/test/project/main.go", "test.module"},
		{"子目录文件", "/test/project/internal/pkg/file.go", "test.module/internal/pkg"},
		{"多级目录", "/test/project/internal/pkg/sub/file.go", "test.module/internal/pkg/sub"},
		{"Windows路径", "/test/project/internal/pkg/file.go", "test.module/internal/pkg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := analyzer.getPackagePath(tt.filename)
			assert.Equal(t, tt.wantPath, path)
		})
	}
}

func TestFindFactoryFunc(t *testing.T) {
	t.Run("查找工厂函数", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.go")

		testCode := `package test

type ITest interface {
	Method()
}

func NewTest() ITest {
	return nil
}

func NewAnother() string {
	return ""
}
`

		err := os.WriteFile(testFile, []byte(testCode), 0644)
		require.NoError(t, err)

		analyzer := NewAnalyzer(tempDir, "test.module")
		fn := analyzer.findFactoryFunc(testFile, "Test")

		assert.NotNil(t, fn)
		assert.Equal(t, "NewTest", fn.Name.Name)
	})

	t.Run("查找不存在的函数", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.go")

		testCode := `package test

type ITest interface {
	Method()
}
`

		err := os.WriteFile(testFile, []byte(testCode), 0644)
		require.NoError(t, err)

		analyzer := NewAnalyzer(tempDir, "test.module")
		fn := analyzer.findFactoryFunc(testFile, "Test")

		assert.Nil(t, fn)
	})
}

func TestAnalyzeRealCode(t *testing.T) {
	t.Run("分析真实Go代码", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.go")

		testCode := `package test

type ITestService interface {
	DoWork() error
}

type TestServiceImpl struct{}

func NewTestService() ITestService {
	return &TestServiceImpl{}
}
`

		err := os.WriteFile(testFile, []byte(testCode), 0644)
		require.NoError(t, err)

		analyzer := NewAnalyzer(tempDir, "test.module")
		info, err := analyzer.Analyze()

		require.NoError(t, err)
		assert.NotNil(t, info)
	})
}
