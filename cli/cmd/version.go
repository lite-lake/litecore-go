package cmd

import (
	"context"
	"fmt"

	"github.com/lite-lake/litecore-go/cli/internal/version"
	"github.com/urfave/cli/v3"
)

func GetVersionCommand() *cli.Command {
	return &cli.Command{
		Name:        "version",
		Usage:       "显示版本",
		Description: `显示 LiteCore CLI 工具的版本信息`,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Printf("litecore-cli 版本 %s\n", version.Version)
			return nil
		},
	}
}

func GetCompletionCommand() *cli.Command {
	return &cli.Command{
		Name:        "completion",
		Usage:       "生成命令自动补全脚本",
		Description: `为指定的 shell 生成自动补全脚本`,
		Commands: []*cli.Command{
			{
				Name:        "bash",
				Usage:       "生成 bash 补全脚本",
				Description: `生成 bash shell 的自动补全脚本`,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("# 生成的 bash 补全脚本需要被 source 才能生效")
					fmt.Println("# 将以下内容添加到 ~/.bashrc 或 ~/.bash_profile：")
					fmt.Println("# source <(litecore-cli completion bash)")
					return nil
				},
			},
			{
				Name:        "zsh",
				Usage:       "生成 zsh 补全脚本",
				Description: `生成 zsh shell 的自动补全脚本`,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("# 生成的 zsh 补全脚本需要被 source 才能生效")
					fmt.Println("# 将以下内容添加到 ~/.zshrc：")
					fmt.Println("# source <(litecore-cli completion zsh)")
					return nil
				},
			},
			{
				Name:        "fish",
				Usage:       "生成 fish 补全脚本",
				Description: `生成 fish shell 的自动补全脚本`,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("# 生成的 fish 补全脚本需要被 source 才能生效")
					fmt.Println("# 将以下内容添加到 ~/.config/fish/config.fish：")
					fmt.Println("# litecore-cli completion fish | source")
					return nil
				},
			},
			{
				Name:        "powershell",
				Usage:       "生成 powershell 补全脚本",
				Description: `生成 PowerShell 的自动补全脚本`,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("# 在 PowerShell 中运行以下命令：")
					fmt.Println("# litecore-cli completion powershell | Out-String | Invoke-Expression")
					return nil
				},
			},
		},
	}
}
