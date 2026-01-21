package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lite-lake/litecore-go/cli/analyzer"
)

func TestNewBuilder(t *testing.T) {
	t.Run("创建生成器", func(t *testing.T) {
		builder := NewBuilder("/test/path", "/output", "app", "test.module", "configs/config.yaml")
		assert.NotNil(t, builder)
		assert.Equal(t, "/test/path", builder.projectPath)
		assert.Equal(t, "/output", builder.outputDir)
		assert.Equal(t, "app", builder.packageName)
		assert.Equal(t, "test.module", builder.moduleName)
		assert.Equal(t, "configs/config.yaml", builder.configPath)
	})
}

func TestBuilder_Generate(t *testing.T) {
	t.Run("生成所有容器代码", func(t *testing.T) {
		tempDir := t.TempDir()
		outputDir := filepath.Join(tempDir, "output")

		info := &analyzer.ProjectInfo{
			ModuleName: "test.module",
			Layers: map[analyzer.Layer][]*analyzer.ComponentInfo{
				analyzer.LayerEntity: {
					{
						InterfaceName: "User",
						FileName:      "user.go",
						PackagePath:   "test.module/entities",
					},
				},
				analyzer.LayerRepository: {
					{
						InterfaceName: "IUserRepository",
						FactoryFunc:   "NewUserRepository",
						PackagePath:   "test.module/repositories",
					},
				},
				analyzer.LayerService: {
					{
						InterfaceName: "IMessageService",
						FactoryFunc:   "NewMessageService",
						PackagePath:   "test.module/services",
					},
				},
				analyzer.LayerController: {
					{
						InterfaceName: "IMessageController",
						FactoryFunc:   "NewMessageController",
						PackagePath:   "test.module/controllers",
					},
				},
				analyzer.LayerMiddleware: {
					{
						InterfaceName: "IAuthMiddleware",
						FactoryFunc:   "NewAuthMiddleware",
						PackagePath:   "test.module/middlewares",
					},
				},
			},
		}

		builder := NewBuilder(tempDir, outputDir, "app", "test.module", "configs/config.yaml")
		err := builder.Generate(info)

		require.NoError(t, err)

		assert.FileExists(t, filepath.Join(outputDir, "entity_container.go"))
		assert.FileExists(t, filepath.Join(outputDir, "repository_container.go"))
		assert.FileExists(t, filepath.Join(outputDir, "service_container.go"))
		assert.FileExists(t, filepath.Join(outputDir, "controller_container.go"))
		assert.FileExists(t, filepath.Join(outputDir, "middleware_container.go"))
		assert.FileExists(t, filepath.Join(outputDir, "engine.go"))
	})
}

func TestConvertComponents(t *testing.T) {
	builder := NewBuilder("/test", "/output", "app", "test.module", "")

	components := []*analyzer.ComponentInfo{
		{
			InterfaceName: "IUserService",
			InterfaceType: "services.IUserService",
			PackagePath:   "test.module/services",
			FactoryFunc:   "NewUserService",
			Layer:         analyzer.LayerService,
		},
		{
			InterfaceName: "Message",
			InterfaceType: "entities.Message",
			PackagePath:   "test.module/entities",
			FactoryFunc:   "",
			Layer:         analyzer.LayerEntity,
			FileName:      "message.go",
		},
	}

	result := builder.convertComponents(components)

	assert.Len(t, result, 2)
	assert.Equal(t, "UserService", result[0].TypeName)
	assert.Equal(t, "IUserService", result[0].InterfaceName)
	assert.Equal(t, "services", result[0].PackageAlias)
	assert.Equal(t, "Message", result[1].TypeName)
	assert.Equal(t, "Message", result[1].InterfaceName)
	assert.Equal(t, "entities", result[1].PackageAlias)
}

func TestGetPackageAlias(t *testing.T) {
	builder := NewBuilder("/test", "/output", "app", "test.module", "")

	tests := []struct {
		name        string
		packagePath string
		wantAlias   string
	}{
		{"标准路径", "test.module/services", "services"},
		{"单级路径", "services", "services"},
		{"多级路径", "test.module/internal/services", "services"},
		{"根模块", "test.module", "test.module"},
		{"空路径", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alias := builder.getPackageAlias(tt.packagePath)
			assert.Equal(t, tt.wantAlias, alias)
		})
	}
}

func TestCollectImports(t *testing.T) {
	builder := NewBuilder("/test", "/output", "app", "test.module", "")

	tests := []struct {
		name    string
		info    *analyzer.ProjectInfo
		layer   analyzer.Layer
		wantLen int
	}{
		{"无组件", &analyzer.ProjectInfo{
			Layers: map[analyzer.Layer][]*analyzer.ComponentInfo{},
		}, analyzer.LayerService, 0},
		{"有组件", &analyzer.ProjectInfo{
			Layers: map[analyzer.Layer][]*analyzer.ComponentInfo{
				analyzer.LayerService: {
					{
						InterfaceName: "IService",
						PackagePath:   "test.module/services",
						InterfaceType: "IService",
					},
				},
			},
		}, analyzer.LayerService, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imports := builder.collectImports(tt.info, tt.layer)
			assert.Len(t, imports, tt.wantLen)
		})
	}
}

func TestWriteFile(t *testing.T) {
	t.Run("写入文件", func(t *testing.T) {
		tempDir := t.TempDir()
		builder := NewBuilder(tempDir, tempDir, "app", "test.module", "")

		content := "test content"
		filename := "test.go"
		err := builder.writeFile(filename, content)

		require.NoError(t, err)

		filePath := filepath.Join(tempDir, filename)
		assert.FileExists(t, filePath)

		data, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, content, string(data))
	})

	t.Run("创建目录并写入", func(t *testing.T) {
		tempDir := t.TempDir()
		outputDir := filepath.Join(tempDir, "output", "nested")
		builder := NewBuilder(tempDir, outputDir, "app", "test.module", "")

		content := "test content"
		filename := "test.go"
		err := builder.writeFile(filename, content)

		require.NoError(t, err)

		filePath := filepath.Join(outputDir, filename)
		assert.FileExists(t, filePath)
	})
}
