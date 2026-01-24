package scaffold

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lite-lake/litecore-go/cli/generator"
)

func Run(cfg *Config) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	if cfg.Interactive {
		if err := RunInteractive(cfg); err != nil {
			return err
		}
	}

	templateData := &TemplateData{
		ModulePath:  cfg.ModulePath,
		ProjectName: cfg.ProjectName,
		LitecoreVer: cfg.LitecoreGoVer,
	}

	if err := createProjectStructure(cfg, templateData); err != nil {
		return fmt.Errorf("创建项目结构失败: %w", err)
	}

	absOutputDir, err := filepath.Abs(cfg.OutputDir)
	if err != nil {
		absOutputDir = cfg.OutputDir
	}

	fmt.Printf("\n项目 %s 创建成功!\n", cfg.ProjectName)
	fmt.Printf("输出目录: %s\n\n", absOutputDir)
	fmt.Println("接下来可以执行:")
	fmt.Printf("  cd %s\n", absOutputDir)
	fmt.Println("  go run ./cmd/generate  # 生成容器代码")
	fmt.Println("  go run ./cmd/server    # 启动应用")

	return nil
}

func (cfg *Config) Validate() error {
	if cfg.ModulePath == "" {
		return fmt.Errorf("模块路径不能为空")
	}
	if cfg.ProjectName == "" {
		return fmt.Errorf("项目名称不能为空")
	}
	return cfg.TemplateType.Validate()
}

func createProjectStructure(cfg *Config, data *TemplateData) error {
	basePath := cfg.OutputDir

	dirs := []string{
		"cmd/generate",
		"cmd/server",
		"configs",
		"internal/entities",
		"internal/repositories",
		"internal/services",
		"internal/controllers",
		"internal/middlewares",
		"internal/listeners",
		"internal/schedulers",
		"internal/application",
		"data",
		"logs",
	}

	for _, dir := range dirs {
		dirPath := filepath.Join(basePath, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("创建目录 %s 失败: %w", dir, err)
		}
	}

	files := []struct {
		path    string
		content func(*TemplateData) (string, error)
	}{
		{"go.mod", GoMod},
		{"README.md", Readme},
		{"configs/config.yaml", ConfigYaml},
		{".gitignore", Gitignore},
		{"cmd/server/main.go", ServerMain},
		{"cmd/generate/main.go", GenerateMain},
	}

	for _, f := range files {
		content, err := f.content(data)
		if err != nil {
			return fmt.Errorf("生成 %s 内容失败: %w", f.path, err)
		}
		filePath := filepath.Join(basePath, f.path)
		if err := writeFile(filePath, content); err != nil {
			return fmt.Errorf("写入文件 %s 失败: %w", f.path, err)
		}
	}

	switch cfg.TemplateType {
	case TemplateTypeStandard:
		if err := generateStandardTemplate(basePath, data); err != nil {
			return err
		}
	case TemplateTypeFull:
		if err := generateStandardTemplate(basePath, data); err != nil {
			return err
		}
		if err := generateFullTemplate(basePath, data); err != nil {
			return err
		}
	}

	if cfg.TemplateType == TemplateTypeStandard || cfg.TemplateType == TemplateTypeFull {
		absBasePath, err := filepath.Abs(basePath)
		if err != nil {
			return fmt.Errorf("获取项目绝对路径失败: %w", err)
		}

		goGenCfg := &generator.Config{
			ProjectPath: absBasePath,
			OutputDir:   "internal/application",
			PackageName: "application",
			ConfigPath:  "configs/config.yaml",
		}
		if err := generator.Run(goGenCfg); err != nil {
			return fmt.Errorf("生成容器代码失败: %w", err)
		}
	}

	return nil
}

func generateStandardTemplate(basePath string, data *TemplateData) error {
	middlewareContent, err := Middleware(data)
	if err != nil {
		return err
	}

	middlewarePath := filepath.Join(basePath, "internal/middlewares", "recovery_middleware.go")
	if err := writeFile(middlewarePath, middlewareContent); err != nil {
		return fmt.Errorf("写入中间件文件失败: %w", err)
	}

	return nil
}

func generateFullTemplate(basePath string, data *TemplateData) error {
	entityContent, err := Entity(data)
	if err != nil {
		return err
	}

	entityPath := filepath.Join(basePath, "internal/entities", "example_entity.go")
	if err := writeFile(entityPath, entityContent); err != nil {
		return fmt.Errorf("写入实体文件失败: %w", err)
	}

	repositoryContent, err := Repository(data)
	if err != nil {
		return err
	}

	repositoryPath := filepath.Join(basePath, "internal/repositories", "example_repository.go")
	if err := writeFile(repositoryPath, repositoryContent); err != nil {
		return fmt.Errorf("写入仓储文件失败: %w", err)
	}

	serviceContent, err := Service(data)
	if err != nil {
		return err
	}

	servicePath := filepath.Join(basePath, "internal/services", "example_service.go")
	if err := writeFile(servicePath, serviceContent); err != nil {
		return fmt.Errorf("写入服务文件失败: %w", err)
	}

	controllerContent, err := Controller(data)
	if err != nil {
		return err
	}

	controllerPath := filepath.Join(basePath, "internal/controllers", "example_controller.go")
	if err := writeFile(controllerPath, controllerContent); err != nil {
		return fmt.Errorf("写入控制器文件失败: %w", err)
	}

	listenerContent, err := Listener(data)
	if err != nil {
		return err
	}

	listenerPath := filepath.Join(basePath, "internal/listeners", "example_listener.go")
	if err := writeFile(listenerPath, listenerContent); err != nil {
		return fmt.Errorf("写入监听器文件失败: %w", err)
	}

	schedulerContent, err := Scheduler(data)
	if err != nil {
		return err
	}

	schedulerPath := filepath.Join(basePath, "internal/schedulers", "example_scheduler.go")
	if err := writeFile(schedulerPath, schedulerContent); err != nil {
		return fmt.Errorf("写入调度器文件失败: %w", err)
	}

	return nil
}

func writeFile(path, content string) error {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return err
	}
	return nil
}
