package cmd

import (
	"context"
	"os"

	"github.com/lite-lake/litecore-go/cli/cmd/generate"
	"github.com/lite-lake/litecore-go/cli/cmd/scaffold"
	"github.com/lite-lake/litecore-go/cli/cmd/upgrade"
	"github.com/urfave/cli/v3"
)

func NewApp() *cli.Command {
	return &cli.Command{
		Name:  "litecore-cli",
		Usage: "LiteCore-Go 框架命令行工具",
		Description: `LiteCore-CLI 是LiteCore配套的命令行工具，提供代码生成和项目脚手架功能。

代码生成：自动扫描项目并生成依赖注入容器代码
项目脚手架：快速创建符合 LiteCore 架构的新项目`,
		Commands: []*cli.Command{
			generate.GetCommand(),
			scaffold.GetCommand(),
			upgrade.GetCommand(),
			GetVersionCommand(),
			GetCompletionCommand(),
		},
	}
}

func Execute() {
	app := NewApp()
	if err := app.Run(context.Background(), os.Args); err != nil {
		os.Exit(1)
	}
}
