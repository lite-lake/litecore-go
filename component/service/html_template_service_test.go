package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"com.litelake.litecore/common"
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
