package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockManager struct {
	name     string
	healthy  bool
	startErr error
	stopErr  error
}

func (m *mockManager) ManagerName() string {
	return m.name
}

func (m *mockManager) Health() error {
	if m.healthy {
		return nil
	}
	return assert.AnError
}

func (m *mockManager) OnStart() error {
	return m.startErr
}

func (m *mockManager) OnStop() error {
	return m.stopErr
}

func TestNewHealthController(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"成功创建HealthController"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewHealthController()
			assert.NotNil(t, controller)
			assert.IsType(t, &HealthController{}, controller)
		})
	}
}

func TestHealthController_ControllerName(t *testing.T) {
	controller := NewHealthController()
	assert.Equal(t, "HealthController", controller.ControllerName())
}

func TestHealthController_GetRouter(t *testing.T) {
	controller := NewHealthController()
	assert.Equal(t, "/health [GET]", controller.GetRouter())
}

func TestHealthController_Handle_无Manager(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	controller := NewHealthController()
	engine.GET("/health", controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"ok"`)
}

func TestHealthController_Handle_健康Manager(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	controller := NewHealthController()
	healthCtrl := controller.(*HealthController)
	healthCtrl.ManagerContainer = &mockManager{name: "TestManager", healthy: true}

	engine.GET("/health", controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"ok"`)
	assert.Contains(t, w.Body.String(), `"TestManager":"ok"`)
}

func TestHealthController_Handle_不健康Manager(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	controller := NewHealthController()
	healthCtrl := controller.(*HealthController)
	healthCtrl.ManagerContainer = &mockManager{name: "UnhealthyManager", healthy: false}

	engine.GET("/health", controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"degraded"`)
	assert.Contains(t, w.Body.String(), `"UnhealthyManager":"unhealthy`)
}
