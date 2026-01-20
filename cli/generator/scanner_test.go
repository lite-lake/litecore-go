package generator

import (
	go_parser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"com.litelake.litecore/cli/analyzer"
)

func TestNewScanner(t *testing.T) {
	t.Run("创建扫描器", func(t *testing.T) {
		scanner := NewScanner("/test/path", "test.module")
		assert.NotNil(t, scanner)
		assert.Equal(t, "/test/path", scanner.projectPath)
		assert.Equal(t, "test.module", scanner.moduleName)
		assert.NotNil(t, scanner.analyzer)
	})
}

func TestScanner_Scan(t *testing.T) {
	t.Run("扫描项目", func(t *testing.T) {
		tempDir := t.TempDir()
		setupScannerTestProject(t, tempDir)

		scanner := NewScanner(tempDir, "test.module")
		info, err := scanner.Scan()

		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "test.module", info.ModuleName)
	})

	t.Run("扫描空目录", func(t *testing.T) {
		tempDir := t.TempDir()

		scanner := NewScanner(tempDir, "test.module")
		info, err := scanner.Scan()

		assert.NoError(t, err)
		assert.NotNil(t, info)
	})
}

func TestFindGoFiles_Scanner(t *testing.T) {
	t.Run("查找Go文件", func(t *testing.T) {
		tempDir := t.TempDir()

		os.WriteFile(filepath.Join(tempDir, "test1.go"), []byte("// test"), 0644)
		os.WriteFile(filepath.Join(tempDir, "test2.go"), []byte("// test"), 0644)
		os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("// test"), 0644)
		os.Mkdir(filepath.Join(tempDir, "subdir"), 0755)
		os.WriteFile(filepath.Join(tempDir, "subdir", "test3.go"), []byte("// test"), 0644)

		scanner := NewScanner(tempDir, "test.module")
		files := scanner.findGoFiles(tempDir)

		assert.Len(t, files, 3)
	})

	t.Run("不存在的目录", func(t *testing.T) {
		scanner := NewScanner("/nonexistent", "test.module")
		files := scanner.findGoFiles("/nonexistent")

		assert.Nil(t, files)
	})
}

func TestGetPackagePath_Scanner(t *testing.T) {
	projectPath := filepath.Join("", "test", "project")
	scanner := NewScanner(projectPath, "test.module")

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
			path := scanner.getPackagePath(tt.filename)
			assert.Equal(t, tt.wantPath, path)
		})
	}
}

func TestScanConfig(t *testing.T) {
	t.Run("扫描配置目录", func(t *testing.T) {
		tempDir := t.TempDir()

		configsDir := filepath.Join(tempDir, "configs")
		os.Mkdir(configsDir, 0755)

		configCode := `package configs

func NewConfigProvider() *ConfigProvider {
	return &ConfigProvider{}
}
`
		os.WriteFile(filepath.Join(configsDir, "config.go"), []byte(configCode), 0644)

		scanner := NewScanner(tempDir, "test.module")
		scanner.scanConfig()

		assert.NotNil(t, scanner)
	})
}

func TestScanManagers(t *testing.T) {
	t.Run("扫描管理器目录", func(t *testing.T) {
		tempDir := t.TempDir()

		managerDir := filepath.Join(tempDir, "internal", "infras", "managers")
		os.MkdirAll(managerDir, 0755)

		managerCode := `package managers

type IDatabaseManager interface {
	Connect() error
}

func NewDatabaseManager() IDatabaseManager {
	return &DatabaseManager{}
}
`
		os.WriteFile(filepath.Join(managerDir, "database.go"), []byte(managerCode), 0644)

		scanner := NewScanner(tempDir, "test.module")
		scanner.scanManagers()

		assert.NotNil(t, scanner)
	})
}

func TestScanComponents(t *testing.T) {
	t.Run("扫描组件目录", func(t *testing.T) {
		tempDir := t.TempDir()

		entitiesDir := filepath.Join(tempDir, "internal", "entities")
		os.MkdirAll(entitiesDir, 0755)

		entityCode := `package entities

type User struct {
	ID   string
	Name string
}
`
		os.WriteFile(filepath.Join(entitiesDir, "user.go"), []byte(entityCode), 0644)

		scanner := NewScanner(tempDir, "test.module")
		info := &analyzer.ProjectInfo{
			Layers: map[analyzer.Layer][]*analyzer.ComponentInfo{
				analyzer.LayerEntity: {},
			},
		}

		scanner.scanComponents(info)

		assert.NotNil(t, scanner)
	})
}

func TestExtractFactoryFunc(t *testing.T) {
	t.Run("提取工厂函数", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.go")

		testCode := `package test

type IService interface {
	Execute() error
}

func NewService() IService {
	return &ServiceImpl{}
}
`

		err := os.WriteFile(testFile, []byte(testCode), 0644)
		require.NoError(t, err)

		fset := token.NewFileSet()
		node, err := go_parser.ParseFile(fset, testFile, nil, go_parser.ParseComments)
		require.NoError(t, err)

		comp := &analyzer.ComponentInfo{
			InterfaceName: "IService",
			FileName:      testFile,
		}

		scanner := NewScanner(tempDir, "test.module")
		scanner.extractFactoryFunc(comp, node)

		assert.Equal(t, "NewService", comp.FactoryFunc)
	})
}

func setupScannerTestProject(t *testing.T, tempDir string) {
	t.Helper()

	goMod := `module test.module
go 1.21
`
	err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0644)
	require.NoError(t, err)

	os.MkdirAll(filepath.Join(tempDir, "configs"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "internal", "infras", "managers"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "internal", "entities"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "internal", "repositories"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "internal", "services"), 0755)

	configCode := `package configs

func NewConfigProvider() *ConfigProvider {
	return &ConfigProvider{}
}
`
	os.WriteFile(filepath.Join(tempDir, "configs", "config.go"), []byte(configCode), 0644)

	managerCode := `package managers

type IDatabaseManager interface {
	Connect() error
}

func NewDatabaseManager() IDatabaseManager {
	return &DatabaseManager{}
}
`
	os.WriteFile(filepath.Join(tempDir, "internal", "infras", "managers", "database.go"), []byte(managerCode), 0644)

	entityCode := `package entities

type User struct {
	ID   string
	Name string
}
`
	os.WriteFile(filepath.Join(tempDir, "internal", "entities", "user.go"), []byte(entityCode), 0644)
}
