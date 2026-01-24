package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFile(t *testing.T) {
	t.Run("写入文件", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "test.txt")
		content := "测试内容"

		err := writeFile(filePath, content)
		if err != nil {
			t.Fatalf("writeFile() error = %v", err)
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("读取文件失败: %v", err)
		}

		if string(data) != content {
			t.Errorf("期望内容为 %q, 实际: %q", content, string(data))
		}
	})
}

func TestWriteFile_CreateDirectory(t *testing.T) {
	t.Run("创建不存在的目录", func(t *testing.T) {
		tempDir := t.TempDir()
		subDir := filepath.Join(tempDir, "subdir")
		filePath := filepath.Join(subDir, "test.txt")
		content := "测试内容"

		err := os.MkdirAll(subDir, 0755)
		if err != nil {
			t.Fatalf("创建目录失败: %v", err)
		}

		err = writeFile(filePath, content)
		if err != nil {
			t.Fatalf("writeFile() error = %v", err)
		}

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Error("文件应该存在")
		}
	})
}

func TestWriteFile_Overwrite(t *testing.T) {
	t.Run("覆盖已有文件", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "test.txt")

		initialContent := "初始内容"
		err := os.WriteFile(filePath, []byte(initialContent), 0644)
		if err != nil {
			t.Fatalf("创建初始文件失败: %v", err)
		}

		newContent := "新内容"
		err = writeFile(filePath, newContent)
		if err != nil {
			t.Fatalf("writeFile() error = %v", err)
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("读取文件失败: %v", err)
		}

		if string(data) != newContent {
			t.Errorf("期望内容为 %q, 实际: %q", newContent, string(data))
		}
	})
}

func TestTemplateTypeDesc(t *testing.T) {
	tests := []struct {
		name string
		t    TemplateType
		want string
	}{
		{
			name: "Basic 描述",
			t:    TemplateTypeBasic,
			want: " - 基础模板：目录结构 + go.mod + README",
		},
		{
			name: "Standard 描述",
			t:    TemplateTypeStandard,
			want: " - 标准模板：基础 + 配置文件 + 基础中间件",
		},
		{
			name: "Full 描述",
			t:    TemplateTypeFull,
			want: " - 完整模板：标准 + 完整示例代码",
		},
		{
			name: "未知类型",
			t:    TemplateType("unknown"),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := templateTypeDesc(tt.t); got != tt.want {
				t.Errorf("templateTypeDesc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateIndexHTML(t *testing.T) {
	t.Run("生成 HTML 内容", func(t *testing.T) {
		projectName := "TestProject"
		html := generateIndexHTML(projectName)

		if html == "" {
			t.Error("HTML 内容不能为空")
		}

		if !strings.Contains(html, "<!DOCTYPE html>") {
			t.Error("HTML 应该包含 DOCTYPE 声明")
		}

		if !strings.Contains(html, "</html>") {
			t.Error("HTML 应该包含结束标签")
		}

		if !strings.Contains(html, "Bootstrap") {
			t.Error("HTML 应该包含 Bootstrap")
		}
	})
}

func TestGenerateIndexHTML_TemplateVariables(t *testing.T) {
	t.Run("检查模板变量", func(t *testing.T) {
		projectName := "TestProject"
		html := generateIndexHTML(projectName)

		expectedVars := []string{
			"{{.title}}",
			"nickname",
			"content",
		}

		for _, expectedVar := range expectedVars {
			if !strings.Contains(html, expectedVar) {
				t.Errorf("HTML 应该包含变量 %s", expectedVar)
			}
		}
	})
}

func TestTemplateData_Fields(t *testing.T) {
	t.Run("验证模板数据字段", func(t *testing.T) {
		data := &TemplateData{
			ModulePath:  "github.com/test/app",
			ProjectName: "test-app",
			LitecoreVer: "v1.0.0",
		}

		if data.ModulePath != "github.com/test/app" {
			t.Errorf("ModulePath 错误")
		}

		if data.ProjectName != "test-app" {
			t.Errorf("ProjectName 错误")
		}

		if data.LitecoreVer != "v1.0.0" {
			t.Errorf("LitecoreVer 错误")
		}
	})
}

func TestConfig_DefaultValues(t *testing.T) {
	t.Run("配置字段验证", func(t *testing.T) {
		cfg := DefaultConfig()

		if cfg.ModulePath != "" {
			t.Errorf("默认 ModulePath 应该为空")
		}

		if cfg.ProjectName != "" {
			t.Errorf("默认 ProjectName 应该为空")
		}

		if cfg.OutputDir != "." {
			t.Errorf("期望 OutputDir 为 '.', 实际: %s", cfg.OutputDir)
		}

		if cfg.TemplateType != TemplateTypeStandard {
			t.Errorf("期望默认模板为 Standard, 实际: %s", cfg.TemplateType)
		}

		if cfg.LitecoreGoVer == "" {
			t.Error("LitecoreGoVer 不能为空")
		}
	})
}

func TestConfig_AllFields(t *testing.T) {
	t.Run("完整配置设置", func(t *testing.T) {
		cfg := &Config{
			ModulePath:    "github.com/test/app",
			ProjectName:   "test-app",
			OutputDir:     "./output",
			TemplateType:  TemplateTypeFull,
			Interactive:   true,
			LitecoreGoVer: "v1.0.0",
			WithStatic:    true,
			WithHTML:      true,
			WithHealth:    true,
		}

		if cfg.ModulePath != "github.com/test/app" {
			t.Errorf("ModulePath 设置失败")
		}

		if cfg.TemplateType != TemplateTypeFull {
			t.Errorf("TemplateType 设置失败")
		}

		if !cfg.WithStatic || !cfg.WithHTML || !cfg.WithHealth {
			t.Error("扩展选项设置失败")
		}
	})
}
