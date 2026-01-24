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
		return fmt.Errorf("get project absolute path failed: path contains invalid characters")
	}

	if !isValidPath(cfg.OutputDir) {
		return fmt.Errorf("get output directory absolute path failed: path contains invalid characters")
	}

	absProjectPath, err := filepath.Abs(cfg.ProjectPath)
	if err != nil {
		return fmt.Errorf("get project absolute path failed: %w", err)
	}

	var absOutputDir string
	if filepath.IsAbs(cfg.OutputDir) {
		absOutputDir = cfg.OutputDir
	} else {
		absOutputDir = filepath.Join(absProjectPath, cfg.OutputDir)
		absOutputDir, err = filepath.Abs(absOutputDir)
		if err != nil {
			return fmt.Errorf("get output directory absolute path failed: %w", err)
		}
	}

	moduleName, err := FindModuleName(absProjectPath)
	if err != nil {
		return fmt.Errorf("find module name failed: %w", err)
	}

	parser := NewParser(absProjectPath)
	info, err := parser.Parse(moduleName)
	if err != nil {
		return fmt.Errorf("parse project failed: %w", err)
	}

	builder := NewBuilder(absProjectPath, absOutputDir, cfg.PackageName, moduleName, cfg.ConfigPath)
	if err := builder.Generate(info); err != nil {
		return fmt.Errorf("generate code failed: %w", err)
	}

	fmt.Printf("Successfully generated container code to %s\n", absOutputDir)
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
