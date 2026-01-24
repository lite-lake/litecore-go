package cmd

import (
	"testing"

	"github.com/lite-lake/litecore-go/cli/internal/version"
	"github.com/urfave/cli/v3"
)

func TestGetVersionCommand(t *testing.T) {
	t.Run("创建版本命令", func(t *testing.T) {
		cmd := GetVersionCommand()

		if cmd.Name != "version" {
			t.Errorf("期望命令名为 'version', 实际: %s", cmd.Name)
		}

		if cmd.Usage != "显示版本" {
			t.Errorf("期望用法为 '显示版本', 实际: %s", cmd.Usage)
		}

		if cmd.Description == "" {
			t.Error("Description 不能为空")
		}

		if cmd.Action == nil {
			t.Error("Action 不能为 nil")
		}
	})
}

func TestGetCompletionCommand(t *testing.T) {
	t.Run("创建补全命令", func(t *testing.T) {
		cmd := GetCompletionCommand()

		if cmd.Name != "completion" {
			t.Errorf("期望命令名为 'completion', 实际: %s", cmd.Name)
		}

		if cmd.Usage != "生成命令自动补全脚本" {
			t.Errorf("期望用法为 '生成命令自动补全脚本', 实际: %s", cmd.Usage)
		}

		if cmd.Description == "" {
			t.Error("Description 不能为空")
		}

		expectedSubcommands := []string{"bash", "zsh", "fish", "powershell"}
		if len(cmd.Commands) != len(expectedSubcommands) {
			t.Errorf("期望 %d 个子命令, 实际: %d", len(expectedSubcommands), len(cmd.Commands))
		}

		subcmdMap := make(map[string]bool)
		for _, subcmd := range cmd.Commands {
			subcmdMap[subcmd.Name] = true
		}

		for _, expected := range expectedSubcommands {
			if !subcmdMap[expected] {
				t.Errorf("缺少子命令: %s", expected)
			}
		}
	})
}

func TestGetCompletionCommand_Bash(t *testing.T) {
	t.Run("Bash 补全命令", func(t *testing.T) {
		completionCmd := GetCompletionCommand()
		var bashCmd *cli.Command

		for _, cmd := range completionCmd.Commands {
			if cmd.Name == "bash" {
				bashCmd = cmd
				break
			}
		}

		if bashCmd == nil {
			t.Fatal("未找到 bash 子命令")
		}

		if bashCmd.Usage != "生成 bash 补全脚本" {
			t.Errorf("期望用法为 '生成 bash 补全脚本', 实际: %s", bashCmd.Usage)
		}

		if bashCmd.Action == nil {
			t.Error("Action 不能为 nil")
		}

		if bashCmd.Description == "" {
			t.Error("Description 不能为空")
		}
	})
}

func TestGetCompletionCommand_Zsh(t *testing.T) {
	t.Run("Zsh 补全命令", func(t *testing.T) {
		completionCmd := GetCompletionCommand()
		var zshCmd *cli.Command

		for _, cmd := range completionCmd.Commands {
			if cmd.Name == "zsh" {
				zshCmd = cmd
				break
			}
		}

		if zshCmd == nil {
			t.Fatal("未找到 zsh 子命令")
		}

		if zshCmd.Usage != "生成 zsh 补全脚本" {
			t.Errorf("期望用法为 '生成 zsh 补全脚本', 实际: %s", zshCmd.Usage)
		}

		if zshCmd.Action == nil {
			t.Error("Action 不能为 nil")
		}

		if zshCmd.Description == "" {
			t.Error("Description 不能为空")
		}
	})
}

func TestGetCompletionCommand_Fish(t *testing.T) {
	t.Run("Fish 补全命令", func(t *testing.T) {
		completionCmd := GetCompletionCommand()
		var fishCmd *cli.Command

		for _, cmd := range completionCmd.Commands {
			if cmd.Name == "fish" {
				fishCmd = cmd
				break
			}
		}

		if fishCmd == nil {
			t.Fatal("未找到 fish 子命令")
		}

		if fishCmd.Usage != "生成 fish 补全脚本" {
			t.Errorf("期望用法为 '生成 fish 补全脚本', 实际: %s", fishCmd.Usage)
		}

		if fishCmd.Action == nil {
			t.Error("Action 不能为 nil")
		}

		if fishCmd.Description == "" {
			t.Error("Description 不能为空")
		}
	})
}

func TestGetCompletionCommand_PowerShell(t *testing.T) {
	t.Run("PowerShell 补全命令", func(t *testing.T) {
		completionCmd := GetCompletionCommand()
		var psCmd *cli.Command

		for _, cmd := range completionCmd.Commands {
			if cmd.Name == "powershell" {
				psCmd = cmd
				break
			}
		}

		if psCmd == nil {
			t.Fatal("未找到 powershell 子命令")
		}

		if psCmd.Usage != "生成 powershell 补全脚本" {
			t.Errorf("期望用法为 '生成 powershell 补全脚本', 实际: %s", psCmd.Usage)
		}

		if psCmd.Action == nil {
			t.Error("Action 不能为 nil")
		}

		if psCmd.Description == "" {
			t.Error("Description 不能为空")
		}
	})
}

func TestVersionConstant(t *testing.T) {
	t.Run("验证版本常量", func(t *testing.T) {
		if version.Version == "" {
			t.Error("版本号不能为空")
		}
	})
}
