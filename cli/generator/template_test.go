package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplateData(t *testing.T) {
	data := &TemplateData{
		PackageName: "application",
		ConfigPath:  "configs/config.yaml",
		Imports:     map[string]string{"entities": "github.com/lite-lake/litecore-go/common"},
		Components: []ComponentTemplateData{
			{
				TypeName:      "Message",
				InterfaceName: "IMessage",
				InterfaceType: "entities.IMessage",
				PackagePath:   "github.com/lite-lake/litecore-go/entities",
				PackageAlias:  "entities",
				FactoryFunc:   "NewMessage",
			},
		},
	}

	assert.Equal(t, "application", data.PackageName)
	assert.Equal(t, "configs/config.yaml", data.ConfigPath)
	assert.Len(t, data.Imports, 1)
	assert.Len(t, data.Components, 1)
	assert.Equal(t, "Message", data.Components[0].TypeName)
}

func TestComponentTemplateData(t *testing.T) {
	comp := ComponentTemplateData{
		TypeName:      "Message",
		InterfaceName: "IMessage",
		InterfaceType: "entities.IMessage",
		PackagePath:   "github.com/lite-lake/litecore-go/entities",
		FactoryFunc:   "NewMessage",
	}

	assert.Equal(t, "Message", comp.TypeName)
	assert.Equal(t, "IMessage", comp.InterfaceName)
	assert.Equal(t, "entities.IMessage", comp.InterfaceType)
	assert.Equal(t, "github.com/lite-lake/litecore-go/entities", comp.PackagePath)
	assert.Equal(t, "NewMessage", comp.FactoryFunc)
}

func TestGenerateEngine(t *testing.T) {
	data := &TemplateData{
		PackageName: "application",
		ConfigPath:  "configs/config.yaml",
	}

	code, err := GenerateEngine(data)
	assert.NoError(t, err)
	assert.Contains(t, code, "package application")
	assert.Contains(t, code, "NewEngine")
	assert.Contains(t, code, "server.NewEngine")
	assert.Contains(t, code, "builtin.Config")
	assert.Contains(t, code, "configs/config.yaml")
}
func TestGenerateEntityContainer(t *testing.T) {
	data := &TemplateData{
		PackageName: "application",
		Imports:     map[string]string{},
		Components: []ComponentTemplateData{
			{
				TypeName:     "Message",
				PackageAlias: "entities",
			},
		},
	}

	code, err := GenerateEntityContainer(data)
	assert.NoError(t, err)
	assert.Contains(t, code, "package application")
	assert.Contains(t, code, "InitEntityContainer")
	assert.Contains(t, code, "RegisterEntity")
	assert.Contains(t, code, "&entities.Message{}")
}

func TestGenerateRepositoryContainer(t *testing.T) {
	data := &TemplateData{
		PackageName: "application",
		Imports:     map[string]string{},
		Components: []ComponentTemplateData{
			{
				InterfaceType: "repositories.IMessageRepository",
				PackagePath:   "github.com/lite-lake/litecore-go/repositories",
				PackageAlias:  "repositories",
				FactoryFunc:   "NewMessageRepository",
			},
		},
	}

	code, err := GenerateRepositoryContainer(data)
	assert.NoError(t, err)
	assert.Contains(t, code, "package application")
	assert.Contains(t, code, "InitRepositoryContainer")
	assert.Contains(t, code, "RegisterRepository")
	assert.Contains(t, code, "IMessageRepository")
}

func TestGenerateServiceContainer(t *testing.T) {
	data := &TemplateData{
		PackageName: "application",
		Imports:     map[string]string{},
		Components: []ComponentTemplateData{
			{
				InterfaceType: "services.IMessageService",
				PackagePath:   "github.com/lite-lake/litecore-go/services",
				PackageAlias:  "services",
				FactoryFunc:   "NewMessageService",
			},
		},
	}

	code, err := GenerateServiceContainer(data)
	assert.NoError(t, err)
	assert.Contains(t, code, "package application")
	assert.Contains(t, code, "InitServiceContainer")
	assert.Contains(t, code, "RegisterService")
	assert.Contains(t, code, "IMessageService")
}

func TestGenerateControllerContainer(t *testing.T) {
	data := &TemplateData{
		PackageName: "application",
		Imports:     map[string]string{},
		Components: []ComponentTemplateData{
			{
				InterfaceType: "controllers.ICreateMessageController",
				PackagePath:   "github.com/lite-lake/litecore-go/controllers",
				PackageAlias:  "controllers",
				FactoryFunc:   "NewCreateMessageController",
			},
		},
	}

	code, err := GenerateControllerContainer(data)
	assert.NoError(t, err)
	assert.Contains(t, code, "package application")
	assert.Contains(t, code, "InitControllerContainer")
	assert.Contains(t, code, "RegisterController")
	assert.Contains(t, code, "ICreateMessageController")
}

func TestGenerateMiddlewareContainer(t *testing.T) {
	data := &TemplateData{
		PackageName: "application",
		Imports:     map[string]string{},
		Components: []ComponentTemplateData{
			{
				InterfaceType: "middlewares.IAuthMiddleware",
				PackagePath:   "github.com/lite-lake/litecore-go/middlewares",
				PackageAlias:  "middlewares",
				FactoryFunc:   "NewAuthMiddleware",
			},
		},
	}

	code, err := GenerateMiddlewareContainer(data)
	assert.NoError(t, err)
	assert.Contains(t, code, "package application")
	assert.Contains(t, code, "InitMiddlewareContainer")
	assert.Contains(t, code, "RegisterMiddleware")
	assert.Contains(t, code, "IAuthMiddleware")
}
