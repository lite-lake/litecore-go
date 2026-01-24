package scaffold

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lite-lake/litecore-go/cli/generator"
)

func Run(cfg *Config) error {
	if cfg.Interactive {
		if err := RunInteractive(cfg); err != nil {
			return err
		}
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
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

	if cfg.WithStatic {
		dirs = append(dirs, "static/css", "static/js", "templates")
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

	if cfg.WithStatic || cfg.WithHTML {
		if err := generateWebTemplate(basePath, data, cfg.WithStatic, cfg.WithHTML); err != nil {
			return err
		}
	}

	if cfg.WithHealth {
		if err := generateHealthTemplate(basePath, data); err != nil {
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

func generateWebTemplate(basePath string, data *TemplateData, withStatic, withHTML bool) error {
	if withStatic {
		staticCSSContent, err := StaticCSS(data)
		if err != nil {
			return err
		}

		staticCSSPath := filepath.Join(basePath, "static/css/style.css")
		if err := writeFile(staticCSSPath, staticCSSContent); err != nil {
			return fmt.Errorf("写入静态CSS文件失败: %w", err)
		}

		staticJSContent, err := StaticJS(data)
		if err != nil {
			return err
		}

		staticJSPath := filepath.Join(basePath, "static/js/app.js")
		if err := writeFile(staticJSPath, staticJSContent); err != nil {
			return fmt.Errorf("写入静态JS文件失败: %w", err)
		}

		staticControllerContent, err := StaticController(data)
		if err != nil {
			return err
		}

		staticControllerPath := filepath.Join(basePath, "internal/controllers", "static_controller.go")
		if err := writeFile(staticControllerPath, staticControllerContent); err != nil {
			return fmt.Errorf("写入静态文件控制器失败: %w", err)
		}
	}

	if withHTML {
		htmlTemplateServiceContent, err := HTMLTemplateService(data)
		if err != nil {
			return err
		}

		htmlTemplateServicePath := filepath.Join(basePath, "internal/services", "html_template_service.go")
		if err := writeFile(htmlTemplateServicePath, htmlTemplateServiceContent); err != nil {
			return fmt.Errorf("写入HTML模板服务失败: %w", err)
		}

		pageControllerContent, err := PageController(data)
		if err != nil {
			return err
		}

		pageControllerPath := filepath.Join(basePath, "internal/controllers", "page_controller.go")
		if err := writeFile(pageControllerPath, pageControllerContent); err != nil {
			return fmt.Errorf("写入页面控制器失败: %w", err)
		}

		indexHTMLContent := generateIndexHTML(data.ProjectName)
		indexHTMLPath := filepath.Join(basePath, "templates/index.html")
		if err := writeFile(indexHTMLPath, indexHTMLContent); err != nil {
			return fmt.Errorf("写入index.html失败: %w", err)
		}
	}

	return nil
}

func generateHealthTemplate(basePath string, data *TemplateData) error {
	healthControllerContent, err := HealthController(data)
	if err != nil {
		return err
	}

	healthControllerPath := filepath.Join(basePath, "internal/controllers", "health_controller.go")
	if err := writeFile(healthControllerPath, healthControllerContent); err != nil {
		return fmt.Errorf("写入健康检查控制器失败: %w", err)
	}

	return nil
}

func generateIndexHTML(projectName string) string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.title}}</title>
    <!-- Bootstrap 5 CSS -->
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <!-- 自定义样式 -->
    <link href="/static/css/style.css" rel="stylesheet">
</head>
<body>
    <div class="container">
        <header class="text-center py-5">
            <h1 class="display-4 fw-light">{{.title}}</h1>
            <p class="text-muted">欢迎使用 LiteCore 框架</p>
        </header>

        <!-- 留言列表 -->
        <section id="messages-section" class="mb-5">
            <h2 class="h4 mb-4 fw-light">最新留言</h2>
            <div id="message-list" class="message-list">
                <div class="text-center text-muted py-5">
                    <div class="spinner-border spinner-border-sm" role="status"></div>
                    <p class="mt-2">加载中...</p>
                </div>
            </div>
        </section>

        <!-- 提交表单 -->
        <section id="form-section" class="mt-5">
            <h2 class="h4 mb-4 fw-light">发表留言</h2>
            <div class="card border-0 shadow-sm">
                <div class="card-body p-4">
                    <form id="message-form">
                        <div class="mb-3">
                            <label for="nickname" class="form-label">昵称</label>
                            <input type="text" class="form-control" id="nickname" name="nickname"
                                   placeholder="请输入您的昵称（2-20个字符）" required minlength="2" maxlength="20">
                        </div>
                        <div class="mb-3">
                            <label for="content" class="form-label">留言内容</label>
                            <textarea class="form-control" id="content" name="content" rows="5"
                                      placeholder="请输入留言内容（5-500个字符）" required minlength="5" maxlength="500"></textarea>
                        </div>
                        <button type="submit" class="btn btn-primary w-100">提交留言</button>
                    </form>
                </div>
            </div>
        </section>

        <footer class="text-center py-5 mt-5 text-muted">
            <small>&copy; 2025 {{.title}}. All rights reserved.</small>
        </footer>
    </div>

    <!-- Bootstrap 5 JS -->
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
    <!-- jQuery -->
    <script src="https://code.jquery.com/jquery-3.7.0.min.js"></script>
    <!-- jQuery Validation -->
    <script src="https://cdn.jsdelivr.net/npm/jquery-validation@1.21.0/dist/jquery.validate.min.js"></script>
    <script src="/static/js/app.js"></script>
</body>
</html>
 `
}
