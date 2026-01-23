package generator

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Config 生成器配置
type Config struct {
	ProjectPath string
	OutputDir   string
	PackageName string
	ConfigPath  string
}

// isValidPath 检查路径是否有效
func isValidPath(path string) bool {
	if path == "" {
		return false
	}
	return !strings.ContainsAny(path, "\x00")
}

// Run 运行代码生成器
func Run(cfg *Config) error {
	if !isValidPath(cfg.ProjectPath) {
		return fmt.Errorf("获取项目绝对路径失败: 路径包含无效字符")
	}

	if !isValidPath(cfg.OutputDir) {
		return fmt.Errorf("获取输出目录绝对路径失败: 路径包含无效字符")
	}

	absProjectPath, err := filepath.Abs(cfg.ProjectPath)
	if err != nil {
		return fmt.Errorf("获取项目绝对路径失败: %w", err)
	}

	absOutputDir := filepath.Join(absProjectPath, cfg.OutputDir)
	absOutputDir, err = filepath.Abs(absOutputDir)
	if err != nil {
		return fmt.Errorf("获取输出目录绝对路径失败: %w", err)
	}

	moduleName, err := FindModuleName(absProjectPath)
	if err != nil {
		return fmt.Errorf("查找模块名失败: %w", err)
	}

	parser := NewParser(absProjectPath)
	info, err := parser.Parse(moduleName)
	if err != nil {
		return fmt.Errorf("解析项目失败: %w", err)
	}

	builder := NewBuilder(absProjectPath, absOutputDir, cfg.PackageName, moduleName, cfg.ConfigPath)
	if err := builder.Generate(info); err != nil {
		return fmt.Errorf("生成代码失败: %w", err)
	}

	fmt.Printf("成功生成容器代码到 %s\n", absOutputDir)
	return nil
}

// MustRun 运行代码生成器，失败时 panic
func MustRun(cfg *Config) {
	if err := Run(cfg); err != nil {
		panic(err)
	}
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		ProjectPath: ".",
		OutputDir:   "internal/application",
		PackageName: "application",
		ConfigPath:  "configs/config.yaml",
	}
}
