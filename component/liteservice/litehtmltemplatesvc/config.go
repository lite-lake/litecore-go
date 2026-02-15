package litehtmltemplatesvc

import (
	"html/template"
	"time"
)

type Config struct {
	TemplatePath   *string           // 模板文件路径模式，如 templates/*
	HotReload      *bool             // 是否启用热重载（开发模式）
	LeftDelim      *string           // 左分隔符，默认 {{
	RightDelim     *string           // 右分隔符，默认 }}
	FuncMap        *template.FuncMap // 自定义模板函数
	ReloadInterval *time.Duration    // 热重载检查间隔
}

const (
	defaultTemplatePath   = "templates/*"
	defaultLeftDelim      = "{{"
	defaultRightDelim     = "}}"
	defaultReloadInterval = time.Second
)

func DefaultConfig() *Config {
	return &Config{}
}

func (c *Config) getTemplatePath() string {
	if c != nil && c.TemplatePath != nil && *c.TemplatePath != "" {
		return *c.TemplatePath
	}
	return defaultTemplatePath
}

func (c *Config) getHotReload() bool {
	if c != nil && c.HotReload != nil {
		return *c.HotReload
	}
	return false
}

func (c *Config) getLeftDelim() string {
	if c != nil && c.LeftDelim != nil && *c.LeftDelim != "" {
		return *c.LeftDelim
	}
	return defaultLeftDelim
}

func (c *Config) getRightDelim() string {
	if c != nil && c.RightDelim != nil && *c.RightDelim != "" {
		return *c.RightDelim
	}
	return defaultRightDelim
}

func (c *Config) getReloadInterval() time.Duration {
	if c != nil && c.ReloadInterval != nil && *c.ReloadInterval > 0 {
		return *c.ReloadInterval
	}
	return defaultReloadInterval
}

func (c *Config) getFuncMap() template.FuncMap {
	if c != nil && c.FuncMap != nil {
		return *c.FuncMap
	}
	return nil
}

func strPtr(s string) *string                    { return &s }
func boolPtr(b bool) *bool                       { return &b }
func durationPtr(d time.Duration) *time.Duration { return &d }
