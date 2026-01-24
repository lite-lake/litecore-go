package scaffold

import (
	"fmt"

	"github.com/lite-lake/litecore-go/cli/internal/version"
)

const LiteCoreGoVersion = version.Version

type TemplateType string

const (
	TemplateTypeBasic    TemplateType = "basic"
	TemplateTypeStandard TemplateType = "standard"
	TemplateTypeFull     TemplateType = "full"
)

func (t TemplateType) String() string {
	return string(t)
}

func (t TemplateType) Validate() error {
	switch t {
	case TemplateTypeBasic, TemplateTypeStandard, TemplateTypeFull:
		return nil
	default:
		return fmt.Errorf("无效的模板类型: %s", t)
	}
}

type Config struct {
	ModulePath    string       // 模块路径，如 github.com/user/app
	ProjectName   string       // 项目名称
	OutputDir     string       // 输出目录
	TemplateType  TemplateType // 模板类型
	Interactive   bool         // 是否交互式
	LitecoreGoVer string       // LiteCore Go 版本
	WithStatic    bool         // 是否生成静态文件
	WithHTML      bool         // 是否生成 HTML 模板
	WithHealth    bool         // 是否生成健康检查控制器
}

func DefaultConfig() *Config {
	return &Config{
		OutputDir:     ".",
		TemplateType:  TemplateTypeStandard,
		Interactive:   false,
		LitecoreGoVer: version.Version,
		WithStatic:    true,
		WithHTML:      true,
		WithHealth:    true,
	}
}
