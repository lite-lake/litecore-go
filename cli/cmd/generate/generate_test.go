package generate

import (
	"testing"

	"github.com/urfave/cli/v3"
)

func TestGetCommand(t *testing.T) {
	t.Run("创建生成命令", func(t *testing.T) {
		cmd := GetCommand()

		if cmd.Name != "generate" {
			t.Errorf("期望命令名为 'generate', 实际: %s", cmd.Name)
		}

		if cmd.Usage != "生成依赖注入容器代码" {
			t.Errorf("期望用法为 '生成依赖注入容器代码', 实际: %s", cmd.Usage)
		}

		if cmd.Description == "" {
			t.Error("Description 不能为空")
		}

		if cmd.Action == nil {
			t.Error("Action 不能为 nil")
		}
	})
}

func TestGetCommand_Flags(t *testing.T) {
	t.Run("验证命令参数", func(t *testing.T) {
		cmd := GetCommand()

		expectedFlags := []string{
			"project",
			"output",
			"package",
			"config",
		}

		flagMap := make(map[string]bool)
		for _, flag := range cmd.Flags {
			flagMap[flag.Names()[0]] = true
		}

		for _, expected := range expectedFlags {
			if !flagMap[expected] {
				t.Errorf("缺少参数: %s", expected)
			}
		}
	})
}

func TestGetCommand_FlagAliases(t *testing.T) {
	t.Run("验证参数别名", func(t *testing.T) {
		cmd := GetCommand()

		tests := []struct {
			name   string
			alias  string
			expect string
		}{
			{"project 短参数", "p", "project"},
			{"output 短参数", "o", "output"},
			{"config 短参数", "c", "config"},
		}

		flagMap := make(map[string]string)
		for _, flag := range cmd.Flags {
			if len(flag.Names()) > 1 {
				flagMap[flag.Names()[1]] = flag.Names()[0]
			}
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if flagMap[tt.alias] != tt.expect {
					t.Errorf("期望别名 %s 对应 %s, 实际: %s", tt.alias, tt.expect, flagMap[tt.alias])
				}
			})
		}
	})
}

func TestGetCommand_DefaultValues(t *testing.T) {
	t.Run("验证默认值", func(t *testing.T) {
		cmd := GetCommand()

		tests := []struct {
			name  string
			flag  string
			value string
		}{
			{"project 默认值", "project", "."},
			{"output 默认值", "output", "internal/application"},
			{"package 默认值", "package", "application"},
			{"config 默认值", "config", "configs/config.yaml"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				for _, flag := range cmd.Flags {
					if flag.Names()[0] == tt.flag {
						strFlag, ok := flag.(*cli.StringFlag)
						if !ok {
							t.Errorf("参数 %s 不是 StringFlag", tt.flag)
							return
						}
						if strFlag.Value != tt.value {
							t.Errorf("期望默认值为 %q, 实际: %q", tt.value, strFlag.Value)
						}
						return
					}
				}
				t.Errorf("未找到参数: %s", tt.flag)
			})
		}
	})
}

func TestGetCommand_FlagDescriptions(t *testing.T) {
	t.Run("验证参数描述", func(t *testing.T) {
		cmd := GetCommand()

		tests := []struct {
			name        string
			flag        string
			description string
		}{
			{"project 描述", "project", "项目路径"},
			{"output 描述", "output", "输出目录"},
			{"package 描述", "package", "包名"},
			{"config 描述", "config", "配置文件路径"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				for _, flag := range cmd.Flags {
					if flag.Names()[0] == tt.flag {
						strFlag, ok := flag.(*cli.StringFlag)
						if !ok {
							t.Errorf("参数 %s 不是 StringFlag", tt.flag)
							return
						}
						if strFlag.Usage == "" {
							t.Errorf("参数 %s 的描述不能为空", tt.flag)
						}
						if strFlag.Usage != tt.description {
							t.Logf("参数 %s 描述为: %s", tt.flag, strFlag.Usage)
						}
						return
					}
				}
				t.Errorf("未找到参数: %s", tt.flag)
			})
		}
	})
}
