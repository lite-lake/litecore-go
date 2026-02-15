// Package litei18nsvc 提供国际化（i18n）服务，支持多语言翻译和本地化管理。
//
// 核心特性：
//   - 多语言翻译：支持键值对方式的翻译文本管理
//   - 变量替换：支持模板变量替换，如 "hello {{.name}}"
//   - 默认语言回退：当请求的语言不支持时，自动回退到默认语言
//   - 动态加载：支持运行时动态加载、重载本地化文件
//   - 目录扫描：支持从目录批量加载本地化文件
//   - 线程安全：使用读写锁保证并发安全
//   - 依赖注入：通过 inject 标签注入 LoggerMgr
//
// 基本用法：
//
//	// 方式一：使用默认配置
//	service := litei18nsvc.NewService()
//
//	// 方式二：使用自定义配置
//	config := &litei18nsvc.Config{
//	    DefaultLanguage: strPtr("zh-CN"),
//	    SupportedLanguages: []string{"zh-CN", "en-US"},
//	}
//	service := litei18nsvc.NewServiceWithConfig(config)
//
// 加载翻译：
//
//	// 从字符串加载
//	service.LoadLocale("zh-CN", "welcome", "欢迎使用")
//
//	// 从文件加载
//	service.LoadLocaleFromFile("zh-CN", "locales/zh-CN.json")
//
//	// 从目录加载
//	service.LoadLocalesFromDir("locales")
//
// 翻译使用：
//
//	// 简单翻译
//	msg := service.T("zh-CN", "welcome")
//
//	// 带变量的翻译
//	msg := service.TWithData("zh-CN", "hello", map[string]interface{}{
//	    "name": "张三",
//	})
//
// 使用依赖注入：
//
//	type MyService struct {
//	    I18nService litei18nsvc.IService `inject:""`
//	}
package litei18nsvc
