package litehtmltemplatesvc

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/lite-lake/litecore-go/common"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewService(t *testing.T) {
	t.Run("创建默认服务", func(t *testing.T) {
		svc := NewLiteHTMLTemplateService()
		assert.NotNil(t, svc)
		assert.Equal(t, "HTMLTemplateService", svc.ServiceName())
	})

	t.Run("使用配置创建服务", func(t *testing.T) {
		path := "templates/*.html"
		cfg := &Config{
			TemplatePath: &path,
		}
		svc := NewLiteHTMLTemplateServiceWithConfig(cfg)
		assert.NotNil(t, svc)
	})
}

func TestConfig(t *testing.T) {
	t.Run("默认配置", func(t *testing.T) {
		cfg := DefaultConfig()
		assert.Equal(t, "templates/*", cfg.getTemplatePath())
		assert.False(t, cfg.getHotReload())
		assert.Equal(t, "{{", cfg.getLeftDelim())
		assert.Equal(t, "}}", cfg.getRightDelim())
		assert.Equal(t, time.Second, cfg.getReloadInterval())
	})

	t.Run("自定义配置", func(t *testing.T) {
		path := "views/**/*.html"
		hotReload := true
		leftDelim := "{%"
		rightDelim := "%}"

		cfg := &Config{
			TemplatePath: &path,
			HotReload:    &hotReload,
			LeftDelim:    &leftDelim,
			RightDelim:   &rightDelim,
		}

		assert.Equal(t, path, cfg.getTemplatePath())
		assert.True(t, cfg.getHotReload())
		assert.Equal(t, leftDelim, cfg.getLeftDelim())
		assert.Equal(t, rightDelim, cfg.getRightDelim())
	})
}

func TestBuiltinFuncs(t *testing.T) {
	svc := NewLiteHTMLTemplateService().(*liteHTMLTemplateServiceImpl)

	t.Run("lower", func(t *testing.T) {
		fn, ok := svc.funcMap["lower"]
		assert.True(t, ok)
		result := fn.(func(string) string)("HELLO")
		assert.Equal(t, "hello", result)
	})

	t.Run("upper", func(t *testing.T) {
		fn, ok := svc.funcMap["upper"]
		assert.True(t, ok)
		result := fn.(func(string) string)("hello")
		assert.Equal(t, "HELLO", result)
	})

	t.Run("trim", func(t *testing.T) {
		fn, ok := svc.funcMap["trim"]
		assert.True(t, ok)
		result := fn.(func(string) string)("  hello  ")
		assert.Equal(t, "hello", result)
	})

	t.Run("safe", func(t *testing.T) {
		fn, ok := svc.funcMap["safe"]
		assert.True(t, ok)
		result := fn.(func(string) template.HTML)("<b>bold</b>")
		assert.Equal(t, template.HTML("<b>bold</b>"), result)
	})

	t.Run("formatDate", func(t *testing.T) {
		fn, ok := svc.funcMap["formatDate"]
		assert.True(t, ok)
		result := fn.(func(time.Time, string) string)(
			time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			"2006-01-02",
		)
		assert.Equal(t, "2024-01-15", result)
	})

	t.Run("dict", func(t *testing.T) {
		fn, ok := svc.funcMap["dict"]
		assert.True(t, ok)
		result, err := fn.(func(...interface{}) (map[string]interface{}, error))("name", "John", "age", 30)
		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{"name": "John", "age": 30}, result)
	})

	t.Run("json", func(t *testing.T) {
		fn, ok := svc.funcMap["json"]
		assert.True(t, ok)
		result, err := fn.(func(interface{}) (string, error))(map[string]string{"key": "value"})
		assert.NoError(t, err)
		assert.Equal(t, `{"key":"value"}`, result)
	})

	t.Run("default", func(t *testing.T) {
		fn, ok := svc.funcMap["default"]
		assert.True(t, ok)
		defaultFn := fn.(func(interface{}, interface{}) interface{})

		assert.Equal(t, "default", defaultFn("default", nil))
		assert.Equal(t, "default", defaultFn("default", ""))
		assert.Equal(t, "value", defaultFn("default", "value"))
	})
}

func TestAddFunc(t *testing.T) {
	svc := NewLiteHTMLTemplateService()

	t.Run("添加单个函数", func(t *testing.T) {
		svc.AddFunc("customFunc", func(s string) string {
			return "prefix_" + s
		})

		impl := svc.(*liteHTMLTemplateServiceImpl)
		fn, ok := impl.funcMap["customFunc"]
		assert.True(t, ok)
		result := fn.(func(string) string)("test")
		assert.Equal(t, "prefix_test", result)
	})

	t.Run("添加函数映射", func(t *testing.T) {
		svc.AddFuncMap(template.FuncMap{
			"func1": func() string { return "func1" },
			"func2": func() string { return "func2" },
		})

		impl := svc.(*liteHTMLTemplateServiceImpl)
		_, ok1 := impl.funcMap["func1"]
		_, ok2 := impl.funcMap["func2"]
		assert.True(t, ok1)
		assert.True(t, ok2)
	})
}

func TestRenderWithoutEngine(t *testing.T) {
	svc := NewLiteHTMLTemplateService()

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	svc.Render(ctx, "test.html", nil)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "gin engine not set")
}

func TestRenderWithoutTemplatesLoaded(t *testing.T) {
	svc := NewLiteHTMLTemplateService()
	impl := svc.(*liteHTMLTemplateServiceImpl)
	impl.ginEngine = gin.New()

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	svc.Render(ctx, "test.html", nil)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "templates not loaded")
}

func TestFullWorkflow(t *testing.T) {
	tmpDir := t.TempDir()

	tmplContent := `<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>
	<h1>{{.Title}}</h1>
	<p>Lower: {{.Name | lower}}</p>
	<p>Upper: {{.Name | upper}}</p>
</body>
</html>`
	tmplPath := filepath.Join(tmpDir, "index.html")
	err := os.WriteFile(tmplPath, []byte(tmplContent), 0644)
	assert.NoError(t, err)

	cfg := &Config{
		TemplatePath: strPtr(filepath.Join(tmpDir, "*.html")),
	}
	svc := NewLiteHTMLTemplateServiceWithConfig(cfg)

	engine := gin.New()
	svc.SetGinEngine(engine)

	err = svc.OnStart()
	assert.NoError(t, err)

	t.Run("渲染模板", func(t *testing.T) {
		engine.GET("/test", func(ctx *gin.Context) {
			svc.Render(ctx, "index.html", gin.H{
				"Title": "测试页面",
				"Name":  "TestUser",
			})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		body := w.Body.String()
		assert.Contains(t, body, "测试页面")
		assert.Contains(t, body, "testuser")
		assert.Contains(t, body, "TESTUSER")
	})

	t.Run("使用状态码渲染", func(t *testing.T) {
		engine.GET("/test2", func(ctx *gin.Context) {
			svc.RenderWithCode(ctx, http.StatusCreated, "index.html", gin.H{
				"Title": "创建成功",
				"Name":  "New",
			})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test2", nil)
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("获取模板名称", func(t *testing.T) {
		names := svc.GetTemplateNames()
		assert.Contains(t, names, "index.html")
	})

	t.Run("检查模板存在", func(t *testing.T) {
		assert.True(t, svc.HasTemplate("index.html"))
		assert.False(t, svc.HasTemplate("notexist.html"))
	})

	t.Run("停止服务", func(t *testing.T) {
		err := svc.OnStop()
		assert.NoError(t, err)

		impl := svc.(*liteHTMLTemplateServiceImpl)
		assert.False(t, impl.loaded)
	})
}

func TestCustomDelimiters(t *testing.T) {
	tmpDir := t.TempDir()

	tmplContent := `<!DOCTYPE html>
<html>
<body>
	<p><%= .Title %></p>
</body>
</html>`
	tmplPath := filepath.Join(tmpDir, "custom.html")
	err := os.WriteFile(tmplPath, []byte(tmplContent), 0644)
	assert.NoError(t, err)

	leftDelim := "<%="
	rightDelim := "%>"
	cfg := &Config{
		TemplatePath: strPtr(filepath.Join(tmpDir, "*.html")),
		LeftDelim:    &leftDelim,
		RightDelim:   &rightDelim,
	}
	svc := NewLiteHTMLTemplateServiceWithConfig(cfg)

	engine := gin.New()
	svc.SetGinEngine(engine)

	err = svc.OnStart()
	assert.NoError(t, err)

	engine.GET("/custom", func(ctx *gin.Context) {
		svc.Render(ctx, "custom.html", gin.H{
			"Title": "自定义分隔符",
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/custom", nil)
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "自定义分隔符")
}

func TestInterfaceCompliance(t *testing.T) {
	svc := NewLiteHTMLTemplateService()

	t.Run("实现 ILiteHTMLTemplateService 接口", func(t *testing.T) {
		var _ ILiteHTMLTemplateService = svc
	})

	t.Run("实现 IBaseService 接口", func(t *testing.T) {
		var _ common.IBaseService = svc
	})
}

func TestReloadTemplatesError(t *testing.T) {
	t.Run("未设置 Gin 引擎", func(t *testing.T) {
		svc := NewLiteHTMLTemplateService()
		err := svc.ReloadTemplates()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "gin engine not set")
	})

	t.Run("模板路径不存在", func(t *testing.T) {
		cfg := &Config{
			TemplatePath: strPtr("/nonexistent/path/*.html"),
		}
		svc := NewLiteHTMLTemplateServiceWithConfig(cfg)

		engine := gin.New()
		svc.SetGinEngine(engine)

		err := svc.ReloadTemplates()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "解析模板失败")
	})
}
