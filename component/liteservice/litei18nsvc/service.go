package litei18nsvc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/lite-lake/litecore-go/manager/loggermgr"
)

// ILiteI18nService 国际化服务接口
type ILiteI18nService interface {
	// T 翻译指定语言的键
	T(lang, key string) string

	// TWithData 翻译指定语言的键，支持变量替换
	TWithData(lang, key string, data map[string]interface{}) string

	// GetSupportedLanguages 获取支持的语言列表
	GetSupportedLanguages() []string

	// IsSupportedLanguage 检查语言是否支持
	IsSupportedLanguage(lang string) bool

	// GetDefaultLanguage 获取默认语言
	GetDefaultLanguage() string

	// LoadLocale 从内存加载本地化数据
	LoadLocale(lang string, data map[string]string) error

	// LoadLocaleFromFile 从文件加载本地化数据
	LoadLocaleFromFile(lang, filePath string) error

	// LoadLocalesFromDir 从目录加载所有本地化文件
	LoadLocalesFromDir(dirPath string) error

	// ReloadLocales 重新加载所有本地化数据
	ReloadLocales() error
}

// liteI18nServiceImpl 国际化服务实现
type liteI18nServiceImpl struct {
	LoggerMgr loggermgr.ILoggerManager `inject:""`
	config    *Config

	// locales 存储各语言的翻译数据
	// 结构: map[语言]map[键]翻译文本
	locales map[string]map[string]string

	// rwmu 读写锁，保证并发安全
	rwmu sync.RWMutex
}

// NewLiteI18nService 使用默认配置创建国际化服务
func NewLiteI18nService() ILiteI18nService {
	return NewLiteI18nServiceWithConfig(nil)
}

// NewLiteI18nServiceWithConfig 使用自定义配置创建国际化服务
func NewLiteI18nServiceWithConfig(config *Config) ILiteI18nService {
	cfg := config
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// 填充默认值
	if cfg.DefaultLanguage == nil {
		defaultLang := "zh-CN"
		cfg.DefaultLanguage = &defaultLang
	}

	if len(cfg.SupportedLanguages) == 0 {
		cfg.SupportedLanguages = []string{*cfg.DefaultLanguage}
	}

	return &liteI18nServiceImpl{
		config:  cfg,
		locales: make(map[string]map[string]string),
	}
}

// T 翻译指定语言的键
func (s *liteI18nServiceImpl) T(lang, key string) string {
	return s.TWithData(lang, key, nil)
}

// TWithData 翻译指定语言的键，支持变量替换
func (s *liteI18nServiceImpl) TWithData(lang, key string, data map[string]interface{}) string {
	s.rwmu.RLock()
	defer s.rwmu.RUnlock()

	// 获取翻译文本
	text := s.getTranslation(lang, key)

	// 如果没有变量数据，直接返回
	if data == nil || len(data) == 0 {
		return text
	}

	// 执行变量替换
	return s.replaceVariables(text, data)
}

// getTranslation 获取翻译文本，支持语言回退
func (s *liteI18nServiceImpl) getTranslation(lang, key string) string {
	// 尝试获取指定语言的翻译
	if locale, ok := s.locales[lang]; ok {
		if text, ok := locale[key]; ok {
			return text
		}
	}

	// 回退到默认语言
	if lang != *s.config.DefaultLanguage {
		if locale, ok := s.locales[*s.config.DefaultLanguage]; ok {
			if text, ok := locale[key]; ok {
				s.logDebug("翻译回退到默认语言", "lang", lang, "key", key, "defaultLang", *s.config.DefaultLanguage)
				return text
			}
		}
	}

	// 未找到翻译，返回键名
	s.logDebug("未找到翻译", "lang", lang, "key", key)
	return key
}

// replaceVariables 替换模板变量
func (s *liteI18nServiceImpl) replaceVariables(text string, data map[string]interface{}) string {
	// 检查是否包含模板语法
	if !strings.Contains(text, "{{") {
		return text
	}

	tmpl, err := template.New("i18n").Parse(text)
	if err != nil {
		s.logWarn("模板解析失败", "text", text, "error", err.Error())
		return text
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		s.logWarn("模板执行失败", "text", text, "error", err.Error())
		return text
	}

	return buf.String()
}

// GetSupportedLanguages 获取支持的语言列表
func (s *liteI18nServiceImpl) GetSupportedLanguages() []string {
	return s.config.SupportedLanguages
}

// IsSupportedLanguage 检查语言是否支持
func (s *liteI18nServiceImpl) IsSupportedLanguage(lang string) bool {
	for _, supported := range s.config.SupportedLanguages {
		if supported == lang {
			return true
		}
	}
	return false
}

// GetDefaultLanguage 获取默认语言
func (s *liteI18nServiceImpl) GetDefaultLanguage() string {
	return *s.config.DefaultLanguage
}

// LoadLocale 从内存加载本地化数据
func (s *liteI18nServiceImpl) LoadLocale(lang string, data map[string]string) error {
	if lang == "" {
		return fmt.Errorf("语言代码不能为空")
	}

	if data == nil {
		return fmt.Errorf("翻译数据不能为空")
	}

	s.rwmu.Lock()
	defer s.rwmu.Unlock()

	// 初始化语言的 map
	if s.locales[lang] == nil {
		s.locales[lang] = make(map[string]string)
	}

	// 合并翻译数据
	for key, value := range data {
		s.locales[lang][key] = value
	}

	s.logInfo("加载本地化数据", "lang", lang, "count", len(data))
	return nil
}

// LoadLocaleFromFile 从文件加载本地化数据
func (s *liteI18nServiceImpl) LoadLocaleFromFile(lang, filePath string) error {
	if lang == "" {
		return fmt.Errorf("语言代码不能为空")
	}

	if filePath == "" {
		return fmt.Errorf("文件路径不能为空")
	}

	// 读取文件
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	// 解析 JSON
	var data map[string]string
	if err := json.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("解析 JSON 失败: %w", err)
	}

	return s.LoadLocale(lang, data)
}

// LoadLocalesFromDir 从目录加载所有本地化文件
// 文件命名规则: <语言代码>.json，如 zh-CN.json, en-US.json
func (s *liteI18nServiceImpl) LoadLocalesFromDir(dirPath string) error {
	if dirPath == "" {
		return fmt.Errorf("目录路径不能为空")
	}

	// 检查目录是否存在
	info, err := os.Stat(dirPath)
	if err != nil {
		return fmt.Errorf("访问目录失败: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("路径不是目录: %s", dirPath)
	}

	// 遍历目录中的 JSON 文件
	files, err := filepath.Glob(filepath.Join(dirPath, "*.json"))
	if err != nil {
		return fmt.Errorf("扫描目录失败: %w", err)
	}

	if len(files) == 0 {
		s.logWarn("目录中没有找到 JSON 文件", "dir", dirPath)
		return nil
	}

	loadedCount := 0
	for _, file := range files {
		// 从文件名提取语言代码
		baseName := filepath.Base(file)
		lang := strings.TrimSuffix(baseName, ".json")

		if err := s.LoadLocaleFromFile(lang, file); err != nil {
			s.logWarn("加载本地化文件失败", "file", file, "error", err.Error())
			continue
		}
		loadedCount++
	}

	s.logInfo("从目录加载本地化文件完成", "dir", dirPath, "loaded", loadedCount, "total", len(files))
	return nil
}

// ReloadLocales 重新加载所有本地化数据
func (s *liteI18nServiceImpl) ReloadLocales() error {
	s.rwmu.Lock()
	// 清空现有数据
	s.locales = make(map[string]map[string]string)
	s.rwmu.Unlock()

	// 如果配置了本地化目录，重新加载
	if s.config.LocalesPath != nil && *s.config.LocalesPath != "" {
		return s.LoadLocalesFromDir(*s.config.LocalesPath)
	}

	s.logInfo("已清空本地化数据")
	return nil
}

// logDebug 记录调试日志
func (s *liteI18nServiceImpl) logDebug(msg string, args ...interface{}) {
	if s.LoggerMgr != nil && s.LoggerMgr.Ins() != nil {
		s.LoggerMgr.Ins().Debug(msg, args...)
	}
}

// logInfo 记录信息日志
func (s *liteI18nServiceImpl) logInfo(msg string, args ...interface{}) {
	if s.LoggerMgr != nil && s.LoggerMgr.Ins() != nil {
		s.LoggerMgr.Ins().Info(msg, args...)
	}
}

// logWarn 记录警告日志
func (s *liteI18nServiceImpl) logWarn(msg string, args ...interface{}) {
	if s.LoggerMgr != nil && s.LoggerMgr.Ins() != nil {
		s.LoggerMgr.Ins().Warn(msg, args...)
	}
}
