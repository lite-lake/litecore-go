package liteservice

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/lite-lake/litecore-go/common"
)

func TestNewHTMLTemplateService(t *testing.T) {
	service := NewHTMLTemplateService("templates/*")
	assert.NotNil(t, service)
	assert.Equal(t, "templates/*", service.config.TemplatePath)
}

func TestHTMLTemplateService_ServiceName(t *testing.T) {
	service := NewHTMLTemplateService("templates/*")
	assert.Equal(t, "HTMLTemplateService", service.ServiceName())
}

func TestHTMLTemplateService_OnStart_WithoutGinEngine(t *testing.T) {
	service := NewHTMLTemplateService("templates/*")
	err := service.OnStart()
	assert.NoError(t, err)
}

func TestHTMLTemplateService_GetConfig(t *testing.T) {
	service := NewHTMLTemplateService("test/*")
	config := service.GetConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "test/*", config.TemplatePath)
}

func TestHTMLTemplateService_Render_WithoutGinEngine(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		service := NewHTMLTemplateService("templates/*")
		service.Render(c, "test.html", gin.H{"key": "value"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "HTML templates not loaded")
}

func TestHTMLTemplateService_SetGinEngine(t *testing.T) {
	service := NewHTMLTemplateService("templates/*")
	ginEngine := gin.New()
	service.SetGinEngine(ginEngine)
	assert.Equal(t, ginEngine, service.ginEngine)
}

func TestHTMLTemplateService_OnStop(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := NewHTMLTemplateService("templates/*")
	ginEngine := gin.New()
	service.SetGinEngine(ginEngine)

	err := service.OnStop()
	assert.NoError(t, err)
	assert.Nil(t, service.ginEngine)
}

var _ IHTMLTemplateService = (*HTMLTemplateService)(nil)
var _ common.IBaseService = (*HTMLTemplateService)(nil)

func TestHTMLTemplateService_OnStart_RewritesStaticPaths(t *testing.T) {
	t.Setenv("LITE_BUILD_ID", "20260803_120000_abc1234")

	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "index.html")
	tmplContent := `<html><head><link href="/static/css/style.css" rel="stylesheet"></head><body>hi</body></html>`
	assert.NoError(t, os.WriteFile(tmplPath, []byte(tmplContent), 0644))

	gin.SetMode(gin.TestMode)
	ginEngine := gin.New()
	service := NewHTMLTemplateService(filepath.Join(dir, "*.html"))
	service.SetGinEngine(ginEngine)
	assert.NoError(t, service.OnStart())

	ginEngine.GET("/", func(c *gin.Context) {
		service.Render(c, "index.html", nil)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ginEngine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
	assert.Contains(t, w.Body.String(), `/static/20260803_120000_abc1234/css/style.css`)
}
