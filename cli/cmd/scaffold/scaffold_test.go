package scaffold

import (
	"testing"

	"github.com/urfave/cli/v3"
)

func TestGetCommand(t *testing.T) {
	t.Run("创建脚手架命令", func(t *testing.T) {
		cmd := GetCommand()

		if cmd.Name != "scaffold" {
			t.Errorf("期望命令名为 'scaffold', 实际: %s", cmd.Name)
		}

		if cmd.Usage != "创建新项目" {
			t.Errorf("期望用法为 '创建新项目', 实际: %s", cmd.Usage)
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
			"module",
			"project",
			"output",
			"template",
			"interactive",
			"static",
			"html",
			"health",
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
			{"module 短参数", "m", "module"},
			{"project 短参数", "n", "project"},
			{"output 短参数", "o", "output"},
			{"template 短参数", "t", "template"},
			{"interactive 短参数", "i", "interactive"},
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

func TestGetCommand_FlagTypes(t *testing.T) {
	t.Run("验证参数类型", func(t *testing.T) {
		cmd := GetCommand()

		stringFlags := []string{"module", "project", "output", "template"}
		boolFlags := []string{"interactive", "static", "html", "health"}

		for _, flagName := range stringFlags {
			found := false
			for _, flag := range cmd.Flags {
				if flag.Names()[0] == flagName {
					if _, ok := flag.(*cli.StringFlag); !ok {
						t.Errorf("参数 %s 应该是 StringFlag", flagName)
					}
					found = true
					break
				}
			}
			if !found {
				t.Errorf("未找到字符串参数: %s", flagName)
			}
		}

		for _, flagName := range boolFlags {
			found := false
			for _, flag := range cmd.Flags {
				if flag.Names()[0] == flagName {
					if _, ok := flag.(*cli.BoolFlag); !ok {
						t.Errorf("参数 %s 应该是 BoolFlag", flagName)
					}
					found = true
					break
				}
			}
			if !found {
				t.Errorf("未找到布尔参数: %s", flagName)
			}
		}
	})
}

func TestGetCommand_FlagDescriptions(t *testing.T) {
	t.Run("验证参数描述", func(t *testing.T) {
		cmd := GetCommand()

		tests := []struct {
			name   string
			flag   string
			expect string
		}{
			{"module 描述", "module", "模块路径"},
			{"project 描述", "project", "项目名称"},
			{"output 描述", "output", "输出目录"},
			{"template 描述", "template", "模板类型"},
			{"interactive 描述", "interactive", "交互式模式"},
			{"static 描述", "static", "静态文件"},
			{"html 描述", "html", "HTML 模板"},
			{"health 描述", "health", "健康检查"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				for _, flag := range cmd.Flags {
					if flag.Names()[0] == tt.flag {
						return
					}
				}
				t.Errorf("未找到参数: %s", tt.flag)
			})
		}
	})
}
