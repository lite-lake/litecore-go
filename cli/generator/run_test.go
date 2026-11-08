package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("默认配置", func(t *testing.T) {
		cfg := DefaultConfig()

		assert.NotNil(t, cfg)
		assert.Equal(t, ".", cfg.ProjectPath)
		assert.Equal(t, "internal/application", cfg.OutputDir)
		assert.Equal(t, "application", cfg.PackageName)
		assert.Equal(t, "configs/config.yaml", cfg.ConfigPath)
	})
}

func TestRun(t *testing.T) {
	t.Run("运行生成器", func(t *testing.T) {
		tempDir := t.TempDir()
		outputDir := filepath.Join(tempDir, "output")
		setupRunTestProject(t, tempDir)

		cfg := &Config{
			ProjectPath: tempDir,
			OutputDir:   outputDir,
			PackageName: "app",
			ConfigPath:  "configs/config.yaml",
		}

		err := Run(cfg)

		require.NoError(t, err)

		assert.FileExists(t, filepath.Join(outputDir, "entity_container.go"))
		assert.FileExists(t, filepath.Join(outputDir, "repository_container.go"))
		assert.FileExists(t, filepath.Join(outputDir, "service_container.go"))
		assert.FileExists(t, filepath.Join(outputDir, "controller_container.go"))
		assert.FileExists(t, filepath.Join(outputDir, "middleware_container.go"))
		assert.FileExists(t, filepath.Join(outputDir, "engine.go"))
	})

	t.Run("项目路径无效", func(t *testing.T) {
		cfg := &Config{
			ProjectPath: "\x00invalid",
			OutputDir:   "/output",
			PackageName: "app",
		}

		err := Run(cfg)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "获取项目绝对路径失败")
	})

	t.Run("输出路径无效", func(t *testing.T) {
		tempDir := t.TempDir()
		setupRunTestProject(t, tempDir)

		cfg := &Config{
			ProjectPath: tempDir,
			OutputDir:   "\x00invalid",
			PackageName: "app",
		}

		err := Run(cfg)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "获取输出目录绝对路径失败")
	})

	t.Run("没有go.mod文件", func(t *testing.T) {
		tempDir := t.TempDir()

		cfg := &Config{
			ProjectPath: tempDir,
			OutputDir:   filepath.Join(tempDir, "output"),
			PackageName: "app",
		}

		err := Run(cfg)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "查找模块名失败")
	})
}

func TestMustRun(t *testing.T) {
	t.Run("成功运行", func(t *testing.T) {
		tempDir := t.TempDir()
		outputDir := filepath.Join(tempDir, "output")
		setupRunTestProject(t, tempDir)

		cfg := &Config{
			ProjectPath: tempDir,
			OutputDir:   outputDir,
			PackageName: "app",
			ConfigPath:  "configs/config.yaml",
		}

		assert.NotPanics(t, func() {
			MustRun(cfg)
		})

		assert.FileExists(t, filepath.Join(outputDir, "engine.go"))
	})

	t.Run("失败时panic", func(t *testing.T) {
		cfg := &Config{
			ProjectPath: "/nonexistent",
			OutputDir:   "/output",
			PackageName: "app",
		}

		assert.Panics(t, func() {
			MustRun(cfg)
		})
	})
}

func setupRunTestProject(t *testing.T, tempDir string) {
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

	os.MkdirAll(filepath.Join(tempDir, "configs"), 0755)
	configFile := []byte(`server:
  port: 8080
`)
	os.WriteFile(filepath.Join(tempDir, "configs", "config.yaml"), configFile, 0644)

	entityCode := `package entities

type User struct {
	ID   string
	Name string
}
`
	os.WriteFile(filepath.Join(tempDir, "internal", "entities", "user.go"), []byte(entityCode), 0644)

	repoCode := `package repositories

import "entities"

type IUserRepository interface {
	GetByID(id string) (*entities.User, error)
}

func NewUserRepository() IUserRepository {
	return &UserRepository{}
}

type UserRepository struct{}

func (r *UserRepository) GetByID(id string) (*entities.User, error) {
	return &entities.User{ID: id}, nil
}
`
	os.WriteFile(filepath.Join(tempDir, "internal", "repositories", "user_repo.go"), []byte(repoCode), 0644)

	serviceCode := `package services

type IMessageService interface {
	Send(message string) error
}

func NewMessageService() IMessageService {
	return &MessageService{}
}

type MessageService struct{}

func (s *MessageService) Send(message string) error {
	return nil
}
`
	os.WriteFile(filepath.Join(tempDir, "internal", "services", "message.go"), []byte(serviceCode), 0644)

	controllerCode := `package controllers

type IHomeController interface {
	Index() string
}

func NewHomeController() IHomeController {
	return &HomeController{}
}

type HomeController struct{}

func (c *HomeController) Index() string {
	return "Hello"
}
`
	os.WriteFile(filepath.Join(tempDir, "internal", "controllers", "home.go"), []byte(controllerCode), 0644)

	middlewareCode := `package middlewares

type IAuthMiddleware interface {
	Process() bool
}

func NewAuthMiddleware() IAuthMiddleware {
	return &AuthMiddleware{}
}

type AuthMiddleware struct{}

func (m *AuthMiddleware) Process() bool {
	return true
}
`
	os.WriteFile(filepath.Join(tempDir, "internal", "middlewares", "auth.go"), []byte(middlewareCode), 0644)
}
