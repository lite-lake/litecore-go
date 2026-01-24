package scaffold

import (
	"context"
	"fmt"
	"os"

	"github.com/lite-lake/litecore-go/cli/scaffold"
	"github.com/urfave/cli/v3"
)

func GetCommand() *cli.Command {
	var modulePath string
	var projectName string
	var outputDir string
	var templateType string
	var interactive bool

	return &cli.Command{
		Name:  "scaffold",
		Usage: "创建新项目",
		Description: `创建一个新的 LiteCore 项目结构，包含目录结构、配置文件和示例代码

模板类型：
  - basic: 基础模板（目录结构 + go.mod + README）
  - standard: 标准模板（基础 + 配置文件 + 基础中间件）
  - full: 完整模板（标准 + 完整示例代码）

如果不指定参数，将进入交互式模式`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "module",
				Aliases:     []string{"m"},
				Usage:       "模块路径 (如 github.com/user/app)",
				Destination: &modulePath,
			},
			&cli.StringFlag{
				Name:        "project",
				Aliases:     []string{"n"},
				Usage:       "项目名称",
				Destination: &projectName,
			},
			&cli.StringFlag{
				Name:        "output",
				Aliases:     []string{"o"},
				Usage:       "输出目录",
				Destination: &outputDir,
			},
			&cli.StringFlag{
				Name:        "template",
				Aliases:     []string{"t"},
				Usage:       "模板类型 (basic/standard/full)",
				Destination: &templateType,
			},
			&cli.BoolFlag{
				Name:        "interactive",
				Aliases:     []string{"i"},
				Usage:       "交互式模式",
				Destination: &interactive,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg := scaffold.DefaultConfig()

			cfg.ModulePath = modulePath
			cfg.ProjectName = projectName
			cfg.OutputDir = outputDir
			cfg.Interactive = interactive

			if templateType != "" {
				cfg.TemplateType = scaffold.TemplateType(templateType)
			}

			if modulePath == "" || projectName == "" || templateType == "" {
				cfg.Interactive = true
			}

			if err := scaffold.Run(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				return cli.Exit("", 1)
			}
			return nil
		},
	}
}
