package generate

import (
	"context"
	"fmt"
	"os"

	"github.com/lite-lake/litecore-go/cli/generator"
	"github.com/urfave/cli/v3"
)

func GetCommand() *cli.Command {
	var projectPath string
	var outputDir string
	var packageName string
	var configPath string

	return &cli.Command{
		Name:  "generate",
		Usage: "生成依赖注入容器代码",
		Description: `扫描项目中的 Entity、Repository、Service、Controller、Middleware、Listener、Scheduler 组件，
并自动生成依赖注入容器的初始化代码`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "project",
				Aliases:     []string{"p"},
				Value:       ".",
				Usage:       "项目路径",
				Destination: &projectPath,
			},
			&cli.StringFlag{
				Name:        "output",
				Aliases:     []string{"o"},
				Value:       "internal/application",
				Usage:       "输出目录",
				Destination: &outputDir,
			},
			&cli.StringFlag{
				Name:        "package",
				Value:       "application",
				Usage:       "包名",
				Destination: &packageName,
			},
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Value:       "configs/config.yaml",
				Usage:       "配置文件路径",
				Destination: &configPath,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg := &generator.Config{
				ProjectPath: projectPath,
				OutputDir:   outputDir,
				PackageName: packageName,
				ConfigPath:  configPath,
			}

			if err := generator.Run(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				return cli.Exit("", 1)
			}
			return nil
		},
	}
}
