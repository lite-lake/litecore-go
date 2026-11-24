package litecontroller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewResourceRobotsController(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"成功创建默认ResourceRobotsController"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewResourceRobotsController()
			assert.NotNil(t, controller)
			assert.Equal(t, DefaultRobotsContent, controller.GetConfig().Content)
			assert.Equal(t, "/robots.txt", controller.GetConfig().URLPath)
		})
	}
}

func TestNewResourceRobotsControllerWithContent(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expected    string
		description string
	}{
		{"自定义内容", "User-agent: *\nAllow: /\n", "User-agent: *\nAllow: /\n", "应使用自定义内容"},
		{"空内容回退默认", "", DefaultRobotsContent, "空内容应回退到默认拒绝所有爬虫"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewResourceRobotsControllerWithContent(tt.content)
			assert.NotNil(t, controller)
			assert.Equal(t, tt.expected, controller.GetConfig().Content, tt.description)
			assert.Equal(t, "/robots.txt", controller.GetConfig().URLPath)
		})
	}
}

func TestResourceRobotsController_ControllerName(t *testing.T) {
	controller := NewResourceRobotsController()
	assert.Equal(t, "ResourceRobotsController", controller.ControllerName())
}

func TestResourceRobotsController_GetRouter(t *testing.T) {
	controller := NewResourceRobotsController()
	assert.Equal(t, "/robots.txt [GET]", controller.GetRouter())
}

func TestResourceRobotsController_Handle_默认拒绝所有爬虫(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller := NewResourceRobotsController()
	engine := gin.New()
	engine.GET(controller.GetConfig().URLPath, controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Equal(t, "public, max-age=3600", w.Header().Get("Cache-Control"))
	assert.Equal(t, DefaultRobotsContent, w.Body.String())
}

func TestResourceRobotsController_Handle_自定义内容(t *testing.T) {
	gin.SetMode(gin.TestMode)

	customContent := "User-agent: Googlebot\nAllow: /\n"
	controller := NewResourceRobotsControllerWithContent(customContent)
	engine := gin.New()
	engine.GET(controller.GetConfig().URLPath, controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, customContent, w.Body.String())
}

func TestResourceRobotsController_GetConfig(t *testing.T) {
	controller := NewResourceRobotsController()
	config := controller.GetConfig()

	assert.NotNil(t, config)
	assert.Equal(t, DefaultRobotsContent, config.Content)
	assert.Equal(t, "/robots.txt", config.URLPath)
}
