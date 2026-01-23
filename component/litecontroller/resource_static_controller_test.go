package litecontroller

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

func TestNewResourceStaticController(t *testing.T) {
	tests := []struct {
		name     string
		urlPath  string
		filePath string
	}{
		{"成功创建ResourceStaticController", "/static", "./static"},
		{"空路径", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewResourceStaticController(tt.urlPath, tt.filePath)
			assert.NotNil(t, controller)
			assert.Equal(t, tt.urlPath, controller.GetConfig().URLPath)
			assert.Equal(t, tt.filePath, controller.GetConfig().FilePath)
		})
	}
}

func TestResourceStaticController_ControllerName(t *testing.T) {
	controller := NewResourceStaticController("/static", "./static")
	assert.Equal(t, "ResourceStaticController", controller.ControllerName())
}

func TestResourceStaticController_GetRouter(t *testing.T) {
	tests := []struct {
		name     string
		urlPath  string
		expected string
	}{
		{"标准路径", "/static", "/static/*filepath [GET]"},
		{"嵌套路径", "/assets/css", "/assets/css/*filepath [GET]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewResourceStaticController(tt.urlPath, "./static")
			assert.Equal(t, tt.expected, controller.GetRouter())
		})
	}
}

func TestResourceStaticController_Handle_不存在的文件(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tmpDir := t.TempDir()
	staticDir := filepath.Join(tmpDir, "static")
	require.NoError(t, os.MkdirAll(staticDir, 0755))

	controller := NewResourceStaticController("/static", staticDir)
	engine := gin.New()
	engine.GET("/static/*filepath", controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/static/nonexistent.txt", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestResourceStaticController_GetConfig(t *testing.T) {
	urlPath := "/assets"
	filePath := "./public"
	controller := NewResourceStaticController(urlPath, filePath)
	config := controller.GetConfig()

	assert.NotNil(t, config)
	assert.Equal(t, urlPath, config.URLPath)
	assert.Equal(t, filePath, config.FilePath)
}

func BenchmarkResourceStaticController_Handle(b *testing.B) {
	gin.SetMode(gin.TestMode)

	tmpDir := b.TempDir()
	staticDir := filepath.Join(tmpDir, "static")
	b.Helper()
	require.NoError(b, os.MkdirAll(staticDir, 0755))

	testFile := filepath.Join(staticDir, "test.txt")
	require.NoError(b, os.WriteFile(testFile, []byte("test content"), 0644))

	controller := NewResourceStaticController("/static", staticDir)
	engine := gin.New()
	engine.GET("/static/*filepath", controller.Handle)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/static/test.txt", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
	}
}
