package scaffold

import (
	"testing"
)

func TestTemplateType_String(t *testing.T) {
	tests := []struct {
		name string
		t    TemplateType
		want string
	}{
		{
			name: "Basic 模板",
			t:    TemplateTypeBasic,
			want: "basic",
		},
		{
			name: "Standard 模板",
			t:    TemplateTypeStandard,
			want: "standard",
		},
		{
			name: "Full 模板",
			t:    TemplateTypeFull,
			want: "full",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.String(); got != tt.want {
				t.Errorf("TemplateType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplateType_Validate(t *testing.T) {
	tests := []struct {
		name    string
		t       TemplateType
		wantErr bool
	}{
		{
			name:    "Basic 有效",
			t:       TemplateTypeBasic,
			wantErr: false,
		},
		{
			name:    "Standard 有效",
			t:       TemplateTypeStandard,
			wantErr: false,
		},
		{
			name:    "Full 有效",
			t:       TemplateTypeFull,
			wantErr: false,
		},
		{
			name:    "无效模板类型",
			t:       TemplateType("invalid"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.t.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("TemplateType.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	t.Run("默认配置", func(t *testing.T) {
		cfg := DefaultConfig()

		if cfg == nil {
			t.Fatal("默认配置不能为 nil")
		}

		if cfg.OutputDir != "." {
			t.Errorf("期望输出目录为 '.', 实际: %s", cfg.OutputDir)
		}

		if cfg.TemplateType != TemplateTypeStandard {
			t.Errorf("期望模板类型为 Standard, 实际: %s", cfg.TemplateType)
		}

		if cfg.Interactive {
			t.Error("期望交互式模式为 false")
		}

		if cfg.LitecoreGoVer == "" {
			t.Error("LiteCore Go 版本不能为空")
		}

		if !cfg.WithStatic {
			t.Error("期望默认生成静态文件")
		}

		if !cfg.WithHTML {
			t.Error("期望默认生成 HTML 模板")
		}

		if !cfg.WithHealth {
			t.Error("期望默认生成健康检查")
		}
	})
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "有效配置",
			cfg: &Config{
				ModulePath:   "github.com/test/app",
				ProjectName:  "test-app",
				TemplateType: TemplateTypeBasic,
			},
			wantErr: false,
		},
		{
			name: "缺少模块路径",
			cfg: &Config{
				ProjectName:  "test-app",
				TemplateType: TemplateTypeBasic,
			},
			wantErr: true,
		},
		{
			name: "缺少项目名称",
			cfg: &Config{
				ModulePath:   "github.com/test/app",
				TemplateType: TemplateTypeBasic,
			},
			wantErr: true,
		},
		{
			name: "缺少模板类型",
			cfg: &Config{
				ModulePath:  "github.com/test/app",
				ProjectName: "test-app",
			},
			wantErr: true,
		},
		{
			name: "空模块路径",
			cfg: &Config{
				ModulePath:   "",
				ProjectName:  "test-app",
				TemplateType: TemplateTypeBasic,
			},
			wantErr: true,
		},
		{
			name: "空项目名称",
			cfg: &Config{
				ModulePath:   "github.com/test/app",
				ProjectName:  "",
				TemplateType: TemplateTypeBasic,
			},
			wantErr: true,
		},
		{
			name: "无效模板类型",
			cfg: &Config{
				ModulePath:   "github.com/test/app",
				ProjectName:  "test-app",
				TemplateType: TemplateType("invalid"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cfg.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_Validate_ErrorMessages(t *testing.T) {
	tests := []struct {
		name         string
		cfg          *Config
		expectSubstr string
	}{
		{
			name: "模块路径错误提示",
			cfg: &Config{
				ModulePath:   "",
				ProjectName:  "test",
				TemplateType: TemplateTypeBasic,
			},
			expectSubstr: "模块路径",
		},
		{
			name: "项目名称错误提示",
			cfg: &Config{
				ModulePath:   "github.com/test/app",
				ProjectName:  "",
				TemplateType: TemplateTypeBasic,
			},
			expectSubstr: "项目名称",
		},
		{
			name: "模板类型错误提示",
			cfg: &Config{
				ModulePath:   "github.com/test/app",
				ProjectName:  "test",
				TemplateType: TemplateType("invalid"),
			},
			expectSubstr: "模板类型",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if err == nil {
				t.Error("期望返回错误，但实际为 nil")
				return
			}

			if !contains(err.Error(), tt.expectSubstr) {
				t.Errorf("期望错误信息包含 %q, 实际: %v", tt.expectSubstr, err)
			}
		})
	}
}

func TestExtractProjectName(t *testing.T) {
	tests := []struct {
		name       string
		modulePath string
		want       string
	}{
		{
			name:       "标准模块路径",
			modulePath: "github.com/user/project",
			want:       "project",
		},
		{
			name:       "带版本号的路径",
			modulePath: "github.com/user/project/v2",
			want:       "v2",
		},
		{
			name:       "简单路径",
			modulePath: "project",
			want:       "project",
		},
		{
			name:       "多层路径",
			modulePath: "github.com/user/org/project",
			want:       "project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractProjectName(tt.modulePath); got != tt.want {
				t.Errorf("extractProjectName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
