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

	"github.com/lite-lake/litecore-go/common/cachebust"
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

func TestResourceStaticController_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tmpDir := t.TempDir()
	staticDir := filepath.Join(tmpDir, "static")
	require.NoError(t, os.MkdirAll(filepath.Join(staticDir, "css"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(staticDir, "css", "style.css"), []byte("body{}"), 0644))

	controller := NewResourceStaticController("/static", staticDir)
	engine := gin.New()
	engine.GET("/static/*filepath", controller.Handle)

	version := cachebust.Version()

	tests := []struct {
		name          string
		url           string
		wantCode      int
		wantLocation  string
		wantCacheCtrl string
	}{
		{
			name:          "未版本化路径302收敛",
			url:           "/static/css/style.css",
			wantCode:      http.StatusFound,
			wantLocation:  cachebust.VersionPrefix() + "/css/style.css",
			wantCacheCtrl: cachebust.HeaderNoCache,
		},
		{
			name:          "版本化路径返回文件",
			url:           "/static/" + version + "/css/style.css",
			wantCode:      http.StatusOK,
			wantCacheCtrl: cachebust.CacheControlForStatic(),
		},
		{
			name:          "过期版本路径302收敛",
			url:           "/static/20260715_100000_abc1234/css/style.css",
			wantCode:      http.StatusFound,
			wantLocation:  cachebust.VersionPrefix() + "/20260715_100000_abc1234/css/style.css",
			wantCacheCtrl: cachebust.HeaderNoCache,
		},
		{
			name:     "版本化文件不存在返回404",
			url:      "/static/" + version + "/nonexistent.txt",
			wantCode: http.StatusNotFound,
		},
		{
			name:          "locales JSON走302版本化",
			url:           "/static/locales/zhs.json",
			wantCode:      http.StatusFound,
			wantLocation:  cachebust.VersionPrefix() + "/locales/zhs.json",
			wantCacheCtrl: cachebust.HeaderNoCache,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			engine.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantLocation != "" {
				assert.Equal(t, tt.wantLocation, w.Header().Get("Location"))
			}
			if tt.wantCacheCtrl != "" {
				assert.Equal(t, tt.wantCacheCtrl, w.Header().Get("Cache-Control"))
			}
		})
	}
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
		// 版本化路径是生产实际服务文件的路径（首跳 302 后浏览器命中）
		req := httptest.NewRequest(http.MethodGet, "/static/"+cachebust.Version()+"/test.txt", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
	}
}
