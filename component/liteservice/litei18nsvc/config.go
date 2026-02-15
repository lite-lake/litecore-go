package litei18nsvc

// Config 国际化服务配置
type Config struct {
	// DefaultLanguage 默认语言，当请求的语言不支持时回退到此语言
	DefaultLanguage *string

	// SupportedLanguages 支持的语言列表
	SupportedLanguages []string

	// LocalesPath 本地化文件目录路径
	LocalesPath *string
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	defaultLang := "zh-CN"
	return &Config{
		DefaultLanguage:    &defaultLang,
		SupportedLanguages: []string{"zh-CN"},
	}
}

func strPtr(s string) *string {
	return &s
}
