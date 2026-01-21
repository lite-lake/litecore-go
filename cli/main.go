// Package main 是 litecore 提供的通用代码生成器命令行工具
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lite-lake/litecore-go/cli/generator"
)

const (
	version = "1.0.0"
)

func main() {
	var showVersion bool

	cfg := generator.DefaultConfig()

	flag.BoolVar(&showVersion, "version", false, "显示版本信息")
	flag.BoolVar(&showVersion, "v", false, "显示版本信息（简写）")
	flag.StringVar(&cfg.ProjectPath, "project", cfg.ProjectPath, "项目路径")
	flag.StringVar(&cfg.ProjectPath, "p", cfg.ProjectPath, "项目路径（简写）")
	flag.StringVar(&cfg.OutputDir, "output", cfg.OutputDir, "输出目录")
	flag.StringVar(&cfg.OutputDir, "o", cfg.OutputDir, "输出目录（简写）")
	flag.StringVar(&cfg.PackageName, "package", cfg.PackageName, "包名")
	flag.StringVar(&cfg.PackageName, "pkg", cfg.PackageName, "包名（简写）")
	flag.StringVar(&cfg.ConfigPath, "config", cfg.ConfigPath, "配置文件路径")
	flag.StringVar(&cfg.ConfigPath, "c", cfg.ConfigPath, "配置文件路径（简写）")

	flag.Parse()

	if showVersion {
		fmt.Printf("litecore-generate version %s\n", version)
		os.Exit(0)
	}

	if err := generator.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}
