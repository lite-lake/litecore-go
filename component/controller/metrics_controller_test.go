package controller

import (
	"github.com/lite-lake/litecore-go/common"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lite-lake/litecore-go/util/logger"
	"github.com/stretchr/testify/assert"
)

type mockService struct {
	name     string
	startErr error
	stopErr  error
}

func (s *mockService) ServiceName() string {
	return s.name
}

func (s *mockService) OnStart() error {
	return s.startErr
}

func (s *mockService) OnStop() error {
	return s.stopErr
}

func (s *mockService) Logger() common.ILogger {
	return nil
}

func (s *mockService) SetLoggerManager(mgr logger.ILoggerManager) {
}

func TestNewMetricsController(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"成功创建MetricsController"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewMetricsController()
			assert.NotNil(t, controller)
			assert.IsType(t, &MetricsController{}, controller)
		})
	}
}

func TestMetricsController_ControllerName(t *testing.T) {
	controller := NewMetricsController()
	assert.Equal(t, "MetricsController", controller.ControllerName())
}

func TestMetricsController_GetRouter(t *testing.T) {
	controller := NewMetricsController()
	assert.Equal(t, "/metrics [GET]", controller.GetRouter())
}

func TestMetricsController_Handle_无依赖(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	controller := NewMetricsController()
	engine.GET("/metrics", controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"server":"litecore-go"`)
	assert.Contains(t, w.Body.String(), `"status":"running"`)
	assert.Contains(t, w.Body.String(), `"version":"1.0.0"`)
}

func TestMetricsController_Handle_有Manager(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	controller := NewMetricsController()
	metricsCtrl := controller.(*MetricsController)
	metricsCtrl.ManagerContainer = &mockManager{name: "TestManager", healthy: true}

	engine.GET("/metrics", controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"managers":1`)
}

func TestMetricsController_Handle_有Service(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	controller := NewMetricsController()
	metricsCtrl := controller.(*MetricsController)
	metricsCtrl.ServiceContainer = &mockService{name: "TestService"}

	engine.GET("/metrics", controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"services":1`)
}

func TestMetricsController_Handle_有Manager和Service(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	controller := NewMetricsController()
	metricsCtrl := controller.(*MetricsController)
	metricsCtrl.ManagerContainer = &mockManager{name: "TestManager", healthy: true}
	metricsCtrl.ServiceContainer = &mockService{name: "TestService"}

	engine.GET("/metrics", controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"managers":1`)
	assert.Contains(t, w.Body.String(), `"services":1`)
}
