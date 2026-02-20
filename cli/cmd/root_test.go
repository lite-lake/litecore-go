package cmd

import (
	"testing"

	"github.com/lite-lake/litecore-go/cli/cmd/generate"
	"github.com/lite-lake/litecore-go/cli/cmd/scaffold"
	"github.com/lite-lake/litecore-go/cli/cmd/upgrade"
)

func TestNewApp(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "创建主应用",
			want: "litecore-cli",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp()
			if app.Name != tt.want {
				t.Errorf("NewApp() = %v, want %v", app.Name, tt.want)
			}
		})
	}
}

func TestNewApp_Commands(t *testing.T) {
	t.Run("验证子命令", func(t *testing.T) {
		app := NewApp()
		expectedCommands := []string{
			generate.GetCommand().Name,
			scaffold.GetCommand().Name,
			upgrade.GetCommand().Name,
			GetVersionCommand().Name,
			GetCompletionCommand().Name,
		}

		if len(app.Commands) != len(expectedCommands) {
			t.Errorf("期望 %d 个命令，实际 %d 个", len(expectedCommands), len(app.Commands))
		}

		cmdMap := make(map[string]bool)
		for _, cmd := range app.Commands {
			cmdMap[cmd.Name] = true
		}

		for _, expected := range expectedCommands {
			if !cmdMap[expected] {
				t.Errorf("缺少命令: %s", expected)
			}
		}
	})
}

func TestNewApp_Description(t *testing.T) {
	t.Run("验证描述", func(t *testing.T) {
		app := NewApp()
		if app.Usage == "" {
			t.Error("Usage 不能为空")
		}
		if app.Description == "" {
			t.Error("Description 不能为空")
		}
	})
}

func TestNewApp_Name(t *testing.T) {
	t.Run("验证应用名称", func(t *testing.T) {
		app := NewApp()
		if app.Name != "litecore-cli" {
			t.Errorf("期望名称为 'litecore-cli', 实际: %s", app.Name)
		}
	})
}

func TestNewApp_Usage(t *testing.T) {
	t.Run("验证应用用法", func(t *testing.T) {
		app := NewApp()
		if app.Usage != "LiteCore-Go 框架命令行工具" {
			t.Errorf("期望用法为 'LiteCore-Go 框架命令行工具', 实际: %s", app.Usage)
		}
	})
}
