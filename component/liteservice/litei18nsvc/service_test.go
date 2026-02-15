package litei18nsvc

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNewService(t *testing.T) {
	t.Run("默认配置创建", func(t *testing.T) {
		svc := NewLiteI18nService()
		if svc == nil {
			t.Fatal("服务不应为 nil")
		}
		if svc.GetDefaultLanguage() != "zh-CN" {
			t.Errorf("默认语言应为 zh-CN，实际为 %s", svc.GetDefaultLanguage())
		}
	})

	t.Run("自定义配置创建", func(t *testing.T) {
		defaultLang := "en-US"
		config := &Config{
			DefaultLanguage:    &defaultLang,
			SupportedLanguages: []string{"en-US", "zh-CN"},
		}
		svc := NewLiteI18nServiceWithConfig(config)
		if svc.GetDefaultLanguage() != "en-US" {
			t.Errorf("默认语言应为 en-US，实际为 %s", svc.GetDefaultLanguage())
		}
		langs := svc.GetSupportedLanguages()
		if len(langs) != 2 {
			t.Errorf("支持语言数量应为 2，实际为 %d", len(langs))
		}
	})
}

func TestLoadLocale(t *testing.T) {
	t.Run("加载翻译数据", func(t *testing.T) {
		svc := NewLiteI18nService()
		data := map[string]string{
			"welcome": "欢迎使用",
			"hello":   "你好",
		}
		err := svc.LoadLocale("zh-CN", data)
		if err != nil {
			t.Fatalf("加载失败: %v", err)
		}

		if svc.T("zh-CN", "welcome") != "欢迎使用" {
			t.Errorf("翻译结果应为 '欢迎使用'")
		}
		if svc.T("zh-CN", "hello") != "你好" {
			t.Errorf("翻译结果应为 '你好'")
		}
	})

	t.Run("空语言代码", func(t *testing.T) {
		svc := NewLiteI18nService()
		err := svc.LoadLocale("", map[string]string{"key": "value"})
		if err == nil {
			t.Error("应返回错误")
		}
	})

	t.Run("空数据", func(t *testing.T) {
		svc := NewLiteI18nService()
		err := svc.LoadLocale("zh-CN", nil)
		if err == nil {
			t.Error("应返回错误")
		}
	})

	t.Run("合并翻译数据", func(t *testing.T) {
		svc := NewLiteI18nService()
		svc.LoadLocale("zh-CN", map[string]string{"key1": "值1"})
		svc.LoadLocale("zh-CN", map[string]string{"key2": "值2"})

		if svc.T("zh-CN", "key1") != "值1" {
			t.Error("key1 翻译失败")
		}
		if svc.T("zh-CN", "key2") != "值2" {
			t.Error("key2 翻译失败")
		}
	})
}

func TestT(t *testing.T) {
	t.Run("简单翻译", func(t *testing.T) {
		svc := NewLiteI18nService()
		svc.LoadLocale("zh-CN", map[string]string{"greeting": "你好世界"})

		result := svc.T("zh-CN", "greeting")
		if result != "你好世界" {
			t.Errorf("期望 '你好世界'，实际 '%s'", result)
		}
	})

	t.Run("不存在的键返回键名", func(t *testing.T) {
		svc := NewLiteI18nService()
		result := svc.T("zh-CN", "nonexistent.key")
		if result != "nonexistent.key" {
			t.Errorf("应返回键名，实际 '%s'", result)
		}
	})

	t.Run("语言回退到默认语言", func(t *testing.T) {
		defaultLang := "zh-CN"
		config := &Config{
			DefaultLanguage:    &defaultLang,
			SupportedLanguages: []string{"zh-CN", "en-US"},
		}
		svc := NewLiteI18nServiceWithConfig(config)
		svc.LoadLocale("zh-CN", map[string]string{"only.zh": "仅中文"})

		// 请求日语，应回退到中文
		result := svc.T("ja-JP", "only.zh")
		if result != "仅中文" {
			t.Errorf("应回退到默认语言，实际 '%s'", result)
		}
	})
}

func TestTWithData(t *testing.T) {
	t.Run("变量替换", func(t *testing.T) {
		svc := NewLiteI18nService()
		svc.LoadLocale("zh-CN", map[string]string{
			"hello": "你好，{{.name}}！",
		})

		result := svc.TWithData("zh-CN", "hello", map[string]interface{}{
			"name": "张三",
		})
		if result != "你好，张三！" {
			t.Errorf("期望 '你好，张三！'，实际 '%s'", result)
		}
	})

	t.Run("多个变量替换", func(t *testing.T) {
		svc := NewLiteI18nService()
		svc.LoadLocale("zh-CN", map[string]string{
			"order": "{{.name}} 订购了 {{.count}} 件商品",
		})

		result := svc.TWithData("zh-CN", "order", map[string]interface{}{
			"name":  "李四",
			"count": 3,
		})
		if result != "李四 订购了 3 件商品" {
			t.Errorf("期望 '李四 订购了 3 件商品'，实际 '%s'", result)
		}
	})

	t.Run("无变量模板直接返回", func(t *testing.T) {
		svc := NewLiteI18nService()
		svc.LoadLocale("zh-CN", map[string]string{"simple": "简单文本"})

		result := svc.TWithData("zh-CN", "simple", map[string]interface{}{
			"unused": "value",
		})
		if result != "简单文本" {
			t.Errorf("期望 '简单文本'，实际 '%s'", result)
		}
	})

	t.Run("空数据直接返回", func(t *testing.T) {
		svc := NewLiteI18nService()
		svc.LoadLocale("zh-CN", map[string]string{"greeting": "你好"})

		result := svc.TWithData("zh-CN", "greeting", nil)
		if result != "你好" {
			t.Errorf("期望 '你好'，实际 '%s'", result)
		}
	})
}

func TestLoadLocaleFromFile(t *testing.T) {
	t.Run("从文件加载", func(t *testing.T) {
		// 创建临时文件
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "zh-CN.json")
		content := `{"welcome": "欢迎", "goodbye": "再见"}`
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("创建临时文件失败: %v", err)
		}

		svc := NewLiteI18nService()
		err := svc.LoadLocaleFromFile("zh-CN", tmpFile)
		if err != nil {
			t.Fatalf("加载失败: %v", err)
		}

		if svc.T("zh-CN", "welcome") != "欢迎" {
			t.Error("翻译结果不正确")
		}
		if svc.T("zh-CN", "goodbye") != "再见" {
			t.Error("翻译结果不正确")
		}
	})

	t.Run("文件不存在", func(t *testing.T) {
		svc := NewLiteI18nService()
		err := svc.LoadLocaleFromFile("zh-CN", "/nonexistent/file.json")
		if err == nil {
			t.Error("应返回错误")
		}
	})

	t.Run("无效 JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "invalid.json")
		if err := os.WriteFile(tmpFile, []byte(`{invalid json}`), 0644); err != nil {
			t.Fatalf("创建临时文件失败: %v", err)
		}

		svc := NewLiteI18nService()
		err := svc.LoadLocaleFromFile("zh-CN", tmpFile)
		if err == nil {
			t.Error("应返回错误")
		}
	})
}

func TestLoadLocalesFromDir(t *testing.T) {
	t.Run("从目录加载多个文件", func(t *testing.T) {
		tmpDir := t.TempDir()

		// 创建多个语言文件
		zhFile := filepath.Join(tmpDir, "zh-CN.json")
		enFile := filepath.Join(tmpDir, "en-US.json")

		os.WriteFile(zhFile, []byte(`{"hello": "你好"}`), 0644)
		os.WriteFile(enFile, []byte(`{"hello": "Hello"}`), 0644)

		svc := NewLiteI18nService()
		err := svc.LoadLocalesFromDir(tmpDir)
		if err != nil {
			t.Fatalf("加载失败: %v", err)
		}

		if svc.T("zh-CN", "hello") != "你好" {
			t.Error("中文翻译不正确")
		}
		if svc.T("en-US", "hello") != "Hello" {
			t.Error("英文翻译不正确")
		}
	})

	t.Run("目录不存在", func(t *testing.T) {
		svc := NewLiteI18nService()
		err := svc.LoadLocalesFromDir("/nonexistent/directory")
		if err == nil {
			t.Error("应返回错误")
		}
	})

	t.Run("路径是文件不是目录", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "file.txt")
		os.WriteFile(tmpFile, []byte("content"), 0644)

		svc := NewLiteI18nService()
		err := svc.LoadLocalesFromDir(tmpFile)
		if err == nil {
			t.Error("应返回错误")
		}
	})

	t.Run("空目录", func(t *testing.T) {
		tmpDir := t.TempDir()

		svc := NewLiteI18nService()
		err := svc.LoadLocalesFromDir(tmpDir)
		if err != nil {
			t.Errorf("空目录不应返回错误: %v", err)
		}
	})
}

func TestIsSupportedLanguage(t *testing.T) {
	t.Run("检查支持的语言", func(t *testing.T) {
		defaultLang := "zh-CN"
		config := &Config{
			DefaultLanguage:    &defaultLang,
			SupportedLanguages: []string{"zh-CN", "en-US", "ja-JP"},
		}
		svc := NewLiteI18nServiceWithConfig(config)

		if !svc.IsSupportedLanguage("zh-CN") {
			t.Error("zh-CN 应该支持")
		}
		if !svc.IsSupportedLanguage("en-US") {
			t.Error("en-US 应该支持")
		}
		if svc.IsSupportedLanguage("fr-FR") {
			t.Error("fr-FR 不应该支持")
		}
	})
}

func TestGetSupportedLanguages(t *testing.T) {
	t.Run("获取支持的语言列表", func(t *testing.T) {
		defaultLang := "zh-CN"
		config := &Config{
			DefaultLanguage:    &defaultLang,
			SupportedLanguages: []string{"zh-CN", "en-US"},
		}
		svc := NewLiteI18nServiceWithConfig(config)

		langs := svc.GetSupportedLanguages()
		if len(langs) != 2 {
			t.Errorf("应返回 2 种语言，实际 %d", len(langs))
		}
	})
}

func TestReloadLocales(t *testing.T) {
	t.Run("清空现有数据", func(t *testing.T) {
		svc := NewLiteI18nService()
		svc.LoadLocale("zh-CN", map[string]string{"key": "值"})

		if svc.T("zh-CN", "key") != "值" {
			t.Fatal("初始数据未加载")
		}

		err := svc.ReloadLocales()
		if err != nil {
			t.Fatalf("重载失败: %v", err)
		}

		// 重载后应返回键名
		if svc.T("zh-CN", "key") != "key" {
			t.Error("重载后数据应被清空")
		}
	})

	t.Run("从配置目录重载", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "zh-CN.json")
		os.WriteFile(tmpFile, []byte(`{"reloaded": "重载后的值"}`), 0644)

		config := &Config{
			DefaultLanguage: strPtr("zh-CN"),
			LocalesPath:     &tmpDir,
		}
		svc := NewLiteI18nServiceWithConfig(config)
		svc.LoadLocale("zh-CN", map[string]string{"old": "旧值"})

		err := svc.ReloadLocales()
		if err != nil {
			t.Fatalf("重载失败: %v", err)
		}

		// 旧数据应被清空
		if svc.T("zh-CN", "old") != "old" {
			t.Error("旧数据应被清空")
		}
		// 新数据应被加载
		if svc.T("zh-CN", "reloaded") != "重载后的值" {
			t.Error("应加载新数据")
		}
	})
}

func TestConcurrency(t *testing.T) {
	t.Run("并发读写", func(t *testing.T) {
		svc := NewLiteI18nService()

		done := make(chan bool)

		// 并发写入
		for i := 0; i < 10; i++ {
			go func(n int) {
				data := map[string]string{fmt.Sprintf("key%d", n): fmt.Sprintf("值%d", n)}
				svc.LoadLocale("zh-CN", data)
				done <- true
			}(i)
		}

		// 并发读取
		for i := 0; i < 10; i++ {
			go func(n int) {
				svc.T("zh-CN", fmt.Sprintf("key%d", n))
				done <- true
			}(i)
		}

		// 等待所有操作完成
		for i := 0; i < 20; i++ {
			<-done
		}
	})
}
