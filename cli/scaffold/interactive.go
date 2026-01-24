package scaffold

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
)

func RunInteractive(cfg *Config) error {
	fmt.Println("欢迎使用 LiteCore 项目脚手架!")
	fmt.Println("请按照提示输入项目信息，或按 Ctrl+C 取消。")
	fmt.Println()

	if cfg.ModulePath == "" {
		modulePath, err := promptModulePath()
		if err != nil {
			return err
		}
		cfg.ModulePath = modulePath
	}

	if cfg.ProjectName == "" {
		projName := extractProjectName(cfg.ModulePath)
		projectName, err := promptProjectName(projName)
		if err != nil {
			return err
		}
		cfg.ProjectName = projectName
	}

	if cfg.TemplateType == "" {
		templateType, err := promptTemplateType()
		if err != nil {
			return err
		}
		cfg.TemplateType = templateType
	}

	if cfg.OutputDir == "" || cfg.OutputDir == "." {
		outputDir, err := promptOutputDir(cfg.ProjectName)
		if err != nil {
			return err
		}
		cfg.OutputDir = outputDir
	}

	if !cfg.WithStatic {
		withStatic, err := promptWithStatic()
		if err != nil {
			return err
		}
		cfg.WithStatic = withStatic
	}

	if !cfg.WithHTML {
		withHTML, err := promptWithHTML()
		if err != nil {
			return err
		}
		cfg.WithHTML = withHTML
	}

	if !cfg.WithHealth {
		withHealth, err := promptWithHealth()
		if err != nil {
			return err
		}
		cfg.WithHealth = withHealth
	}

	return nil
}

func promptModulePath() (string, error) {
	prompt := promptui.Prompt{
		Label: "模块路径",
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("模块路径不能为空")
			}
			return nil
		},
		Default: "github.com/yourusername/myproject",
	}

	result, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("输入取消: %w", err)
	}
	return result, nil
}

func promptProjectName(defaultName string) (string, error) {
	prompt := promptui.Prompt{
		Label:   "项目名称",
		Default: defaultName,
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("项目名称不能为空")
			}
			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("输入取消: %w", err)
	}
	return result, nil
}

func promptOutputDir(defaultName string) (string, error) {
	prompt := promptui.Prompt{
		Label:   "输出目录",
		Default: "./" + defaultName,
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("输出目录不能为空")
			}
			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("输入取消: %w", err)
	}
	return result, nil
}

func promptTemplateType() (TemplateType, error) {
	prompt := promptui.Select{
		Label: "选择模板类型",
		Items: []TemplateType{
			TemplateTypeBasic,
			TemplateTypeStandard,
			TemplateTypeFull,
		},
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . | green }} ?",
			Active:   `{{ "▸" | green }} {{ . | cyan }} {{ templateTypeDesc . | faint }}`,
			Inactive: `  {{ . | faint }} {{ templateTypeDesc . | faint }}`,
			Selected: `{{ "✓" | green }} {{ . | bold }}`,
		},
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("选择取消: %w", err)
	}

	return TemplateType(result), nil
}

func templateTypeDesc(t TemplateType) string {
	switch t {
	case TemplateTypeBasic:
		return " - 基础模板：目录结构 + go.mod + README"
	case TemplateTypeStandard:
		return " - 标准模板：基础 + 配置文件 + 基础中间件"
	case TemplateTypeFull:
		return " - 完整模板：标准 + 完整示例代码"
	default:
		return ""
	}
}

func confirmOverwrite(path string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("目录 %s 已存在，是否覆盖", path),
		IsConfirm: true,
	}

	_, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrAbort {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func promptGenerateCode() (bool, error) {
	prompt := promptui.Prompt{
		Label:     "是否立即生成容器代码",
		IsConfirm: true,
		Default:   "y",
	}

	_, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrAbort {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func extractProjectName(modulePath string) string {
	parts := []rune(modulePath)
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == '/' {
			return string(parts[i+1:])
		}
	}
	return modulePath
}

func confirmPrompt(message string) bool {
	prompt := promptui.Prompt{
		Label:     message,
		IsConfirm: true,
	}

	_, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrAbort {
			return false
		}
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		return false
	}
	return true
}

func promptWithStatic() (bool, error) {
	prompt := promptui.Prompt{
		Label:     "是否生成静态文件服务 (CSS/JS)",
		IsConfirm: true,
		Default:   "n",
	}

	_, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrAbort {
			return false, nil
		}
		return false, fmt.Errorf("输入取消: %w", err)
	}
	return true, nil
}

func promptWithHTML() (bool, error) {
	prompt := promptui.Prompt{
		Label:     "是否生成 HTML 模板服务",
		IsConfirm: true,
		Default:   "n",
	}

	_, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrAbort {
			return false, nil
		}
		return false, fmt.Errorf("输入取消: %w", err)
	}
	return true, nil
}

func promptWithHealth() (bool, error) {
	prompt := promptui.Prompt{
		Label:     "是否生成健康检查控制器",
		IsConfirm: true,
		Default:   "n",
	}

	_, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrAbort {
			return false, nil
		}
		return false, fmt.Errorf("输入取消: %w", err)
	}
	return true, nil
}
