package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"com.litelake.litecore/cli/analyzer"
)

func TestNewParser(t *testing.T) {
	t.Run("创建解析器", func(t *testing.T) {
		parser := NewParser("/test/path")
		assert.NotNil(t, parser)
		assert.Equal(t, "/test/path", parser.projectPath)
		assert.NotNil(t, parser.info)
		assert.NotNil(t, parser.info.Layers)
	})
}

func TestParse(t *testing.T) {
	t.Run("解析完整项目", func(t *testing.T) {
		tempDir := t.TempDir()
		setupTestProject(t, tempDir)

		parser := NewParser(tempDir)
		info, err := parser.Parse("test.module")

		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "test.module", info.ModuleName)
	})

	t.Run("解析空目录", func(t *testing.T) {
		tempDir := t.TempDir()

		parser := NewParser(tempDir)
		info, err := parser.Parse("test.module")

		require.NoError(t, err)
		assert.NotNil(t, info)
	})
}

func TestFindModuleName(t *testing.T) {
	tests := []struct {
		name       string
		goMod      string
		wantModule string
		wantErr    bool
	}{
		{"标准模块", "module test.module\ngo 1.21\n", "test.module", false},
		{"带注释", "module test.module // comment\ngo 1.21\n", "test.module // comment", false},
		{"复杂模块名", "module github.com/example/project\ngo 1.21\n", "github.com/example/project", false},
		{"无效格式", "invalid format", "", true},
		{"空文件", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			goModPath := filepath.Join(tempDir, "go.mod")

			err := os.WriteFile(goModPath, []byte(tt.goMod), 0644)
			require.NoError(t, err)

			moduleName, err := FindModuleName(tempDir)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantModule, moduleName)
			}
		})
	}
}

func TestGetPackagePath(t *testing.T) {
	projectPath := filepath.Join("", "test", "project")
	parser := NewParser(projectPath)
	parser.info.ModuleName = "test.module"

	tests := []struct {
		name     string
		filename string
		wantPath string
	}{
		{"根目录文件", filepath.Join(projectPath, "file.go"), "test.module"},
		{"子目录文件", filepath.Join(projectPath, "internal", "file.go"), "test.module/internal"},
		{"多级目录", filepath.Join(projectPath, "internal", "pkg", "file.go"), "test.module/internal/pkg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := parser.getPackagePath(tt.filename)
			assert.Equal(t, tt.wantPath, path)
		})
	}
}

func TestFindGoFiles(t *testing.T) {
	t.Run("查找Go文件", func(t *testing.T) {
		tempDir := t.TempDir()

		os.WriteFile(filepath.Join(tempDir, "test1.go"), []byte("// test"), 0644)
		os.WriteFile(filepath.Join(tempDir, "test2.go"), []byte("// test"), 0644)
		os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("// test"), 0644)
		os.WriteFile(filepath.Join(tempDir, "test_test.go"), []byte("// test"), 0644)
		os.Mkdir(filepath.Join(tempDir, "subdir"), 0755)
		os.WriteFile(filepath.Join(tempDir, "subdir", "test3.go"), []byte("// test"), 0644)

		parser := NewParser(tempDir)
		files, err := parser.findGoFiles(tempDir)

		require.NoError(t, err)
		assert.Len(t, files, 3)
	})

	t.Run("不存在的目录", func(t *testing.T) {
		parser := NewParser("/nonexistent")
		files, err := parser.findGoFiles("/nonexistent")

		assert.Error(t, err)
		assert.Nil(t, files)
	})
}

func TestParseEntityFile(t *testing.T) {
	t.Run("解析实体文件", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "entity.go")

		code := `package entities

type User struct {
	ID   string
	Name string
}

type Message struct {
	ID   string
	Text string
}
`

		err := os.WriteFile(testFile, []byte(code), 0644)
		require.NoError(t, err)

		parser := NewParser(tempDir)
		err = parser.parseEntityFile(testFile)

		require.NoError(t, err)
		assert.Len(t, parser.info.Layers[analyzer.LayerEntity], 2)
	})
}

func TestParseRepositoryFile(t *testing.T) {
	t.Run("解析仓储文件", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "repository.go")

		code := `package repositories

type IUserRepository interface {
	GetByID(id string) (*User, error)
}

func NewUserRepository() IUserRepository {
	return &UserRepository{}
}
`

		err := os.WriteFile(testFile, []byte(code), 0644)
		require.NoError(t, err)

		parser := NewParser(tempDir)
		err = parser.parseRepositoryFile(testFile)

		require.NoError(t, err)
		assert.Len(t, parser.info.Layers[analyzer.LayerRepository], 1)
	})
}

func TestParseServiceFile(t *testing.T) {
	t.Run("解析服务文件", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "service.go")

		code := `package services

type IMessageService interface {
	Send(message string) error
}

func NewMessageService() IMessageService {
	return &MessageServiceImpl{}
}
`

		err := os.WriteFile(testFile, []byte(code), 0644)
		require.NoError(t, err)

		parser := NewParser(tempDir)
		err = parser.parseServiceFile(testFile)

		require.NoError(t, err)
		assert.Len(t, parser.info.Layers[analyzer.LayerService], 1)
	})
}

func TestParseControllerFile(t *testing.T) {
	t.Run("解析控制器文件", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "controller.go")

		code := `package controllers

type IMessageController interface {
	HandleMessage() error
}

func NewMessageController() IMessageController {
	return &MessageControllerImpl{}
}
`

		err := os.WriteFile(testFile, []byte(code), 0644)
		require.NoError(t, err)

		parser := NewParser(tempDir)
		err = parser.parseControllerFile(testFile)

		require.NoError(t, err)
		assert.Len(t, parser.info.Layers[analyzer.LayerController], 1)
	})
}

func TestParseMiddlewareFile(t *testing.T) {
	t.Run("解析中间件文件", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "middleware.go")

		code := `package middlewares

type IAuthMiddleware interface {
	Authenticate() bool
}

func NewAuthMiddleware() IAuthMiddleware {
	return &AuthMiddlewareImpl{}
}
`

		err := os.WriteFile(testFile, []byte(code), 0644)
		require.NoError(t, err)

		parser := NewParser(tempDir)
		err = parser.parseMiddlewareFile(testFile)

		require.NoError(t, err)
		assert.Len(t, parser.info.Layers[analyzer.LayerMiddleware], 1)
	})
}

func TestParseInfrasFile(t *testing.T) {
	t.Run("解析基础设施文件-管理器", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "manager.go")

		code := `package infras

func NewDatabaseManager() *DatabaseManager {
	return &DatabaseManager{}
}
`

		err := os.WriteFile(testFile, []byte(code), 0644)
		require.NoError(t, err)

		parser := NewParser(tempDir)
		err = parser.parseInfrasFile(testFile)

		require.NoError(t, err)
		assert.Len(t, parser.info.Layers[analyzer.LayerManager], 1)
	})

	t.Run("解析基础设施文件-配置提供者", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "config.go")

		code := `package configproviders

func NewConfigProvider() *ConfigProvider {
	return &ConfigProvider{}
}
`

		err := os.WriteFile(testFile, []byte(code), 0644)
		require.NoError(t, err)

		parser := NewParser(tempDir)
		err = parser.parseInfrasFile(testFile)

		require.NoError(t, err)
		assert.Len(t, parser.info.Layers[analyzer.LayerConfig], 1)
	})
}

func setupTestProject(t *testing.T, tempDir string) {
	t.Helper()

	goMod := `module test.module
go 1.21
`
	err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0644)
	require.NoError(t, err)

	os.MkdirAll(filepath.Join(tempDir, "internal", "entities"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "internal", "repositories"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "internal", "services"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "internal", "controllers"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "internal", "middlewares"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "internal", "infras"), 0755)

	entityCode := `package entities

type User struct {
	ID   string
	Name string
}
`
	os.WriteFile(filepath.Join(tempDir, "internal", "entities", "user.go"), []byte(entityCode), 0644)

	repoCode := `package repositories

type IUserRepository interface {
	GetByID(id string) (*User, error)
}
`
	os.WriteFile(filepath.Join(tempDir, "internal", "repositories", "user_repo.go"), []byte(repoCode), 0644)

	serviceCode := `package services

type IMessageService interface {
	Send(message string) error
}
`
	os.WriteFile(filepath.Join(tempDir, "internal", "services", "message.go"), []byte(serviceCode), 0644)
}
