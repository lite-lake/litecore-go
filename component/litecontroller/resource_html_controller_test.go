package controller

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResourceHTMLController(t *testing.T) {
	tests := []struct {
		name         string
		templatePath string
	}{
		{"成功创建ResourceHTMLController", "templates/*"},
		{"空模板路径", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewResourceHTMLController(tt.templatePath)
			assert.NotNil(t, controller)
			assert.Equal(t, tt.templatePath, controller.GetConfig().TemplatePath)
		})
	}
}

func TestResourceHTMLController_ControllerName(t *testing.T) {
	controller := NewResourceHTMLController("templates/*")
	assert.Equal(t, "ResourceHTMLController", controller.ControllerName())
}

func TestResourceHTMLController_GetRouter(t *testing.T) {
	controller := NewResourceHTMLController("templates/*")
	assert.Equal(t, "", controller.GetRouter())
}

func TestResourceHTMLController_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	controller := NewResourceHTMLController("templates/*")
	engine.GET("/test", controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "ResourceHTMLController should not be registered as a route")
}

func TestResourceHTMLController_LoadTemplates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		templatePath string
	}{
		{"成功加载模板", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewResourceHTMLController(tt.templatePath)
			engine := gin.New()

			if tt.templatePath == "" {
				t.Skip("跳过空路径测试")
			}

			controller.LoadTemplates(engine)
		})
	}
}

func TestResourceHTMLController_LoadTemplates_实际模板(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tmpDir := t.TempDir()
	templateContent := `<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>{{.Content}}</body>
</html>`
	templateFile := filepath.Join(tmpDir, "test.html")
	require.NoError(t, os.WriteFile(templateFile, []byte(templateContent), 0644))

	controller := NewResourceHTMLController(filepath.Join(tmpDir, "*.html"))
	engine := gin.New()

	controller.LoadTemplates(engine)
	assert.NotNil(t, controller)
}

func TestResourceHTMLController_Render_未加载模板(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	controller := NewResourceHTMLController("templates/*")
	engine.GET("/render", func(c *gin.Context) {
		controller.Render(c, "test.html", map[string]string{"Title": "Test"})
	})

	req := httptest.NewRequest(http.MethodGet, "/render", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "HTML templates not loaded")
}

func TestResourceHTMLController_Render_已加载模板(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tmpDir := t.TempDir()
	templateContent := `<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>{{.Content}}</body>
</html>`
	templateFile := filepath.Join(tmpDir, "test.html")
	require.NoError(t, os.WriteFile(templateFile, []byte(templateContent), 0644))

	controller := NewResourceHTMLController(filepath.Join(tmpDir, "*.html"))
	engine := gin.New()

	controller.LoadTemplates(engine)
	engine.GET("/render", func(c *gin.Context) {
		controller.Render(c, "test.html", map[string]string{
			"Title":   "Test Title",
			"Content": "Test Content",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/render", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Test Title")
	assert.Contains(t, w.Body.String(), "Test Content")
}

func TestResourceHTMLController_GetConfig(t *testing.T) {
	templatePath := "templates/*"
	controller := NewResourceHTMLController(templatePath)
	config := controller.GetConfig()

	assert.NotNil(t, config)
	assert.Equal(t, templatePath, config.TemplatePath)
}

func BenchmarkResourceHTMLController_Render(b *testing.B) {
	gin.SetMode(gin.TestMode)

	tmpDir := b.TempDir()
	templateContent := `<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>{{.Content}}</body>
</html>`
	templateFile := filepath.Join(tmpDir, "test.html")
	b.Helper()
	require.NoError(b, os.WriteFile(templateFile, []byte(templateContent), 0644))

	controller := NewResourceHTMLController(filepath.Join(tmpDir, "*.html"))
	engine := gin.New()
	controller.LoadTemplates(engine)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/render", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		controller.Render(c, "test.html", map[string]string{
			"Title":   "Test Title",
			"Content": "Test Content",
		})
	}
}
