package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// Mock managers for health testing

type healthyManager struct {
	name string
}

func (m *healthyManager) ManagerName() string {
	return m.name
}

func (m *healthyManager) OnStart() error {
	return nil
}

func (m *healthyManager) OnStop() error {
	return nil
}

func (m *healthyManager) Health() error {
	return nil
}

func (m *healthyManager) Config() interface{} {
	return nil
}

type unhealthyManager struct {
	name string
	err  error
}

func (m *unhealthyManager) ManagerName() string {
	return m.name
}

func (m *unhealthyManager) OnStart() error {
	return nil
}

func (m *unhealthyManager) OnStop() error {
	return nil
}

func (m *unhealthyManager) Health() error {
	return m.err
}

func (m *unhealthyManager) Config() interface{} {
	return nil
}

func setupHealthEngine(managers ...interface{}) (*Engine, *gin.Engine, error) {
	gin.SetMode(gin.TestMode)

	engine, err := NewEngine(RegisterManagers(managers...))
	if err != nil {
		return nil, nil, err
	}

	if err := engine.Initialize(); err != nil {
		return nil, nil, err
	}

	return engine, engine.ginEngine, nil
}

func TestHealthHandler_AllHealthy(t *testing.T) {
	_, ginEngine, err := setupHealthEngine(&healthyManager{name: "manager1"})
	if err != nil {
		t.Fatalf("setupHealthEngine() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	ginEngine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	var response HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("status = %s, want ok", response.Status)
	}

	if response.Timestamp == "" {
		t.Error("timestamp should not be empty")
	}

	if len(response.Managers) == 0 {
		t.Error("managers should not be empty")
	}

	if response.Managers["manager1"] != "ok" {
		t.Errorf("manager1 status = %s, want ok", response.Managers["manager1"])
	}
}

func TestHealthHandler_OneUnhealthy(t *testing.T) {
	_, ginEngine, err := setupHealthEngine(
		&healthyManager{name: "manager1"},
		&unhealthyManager{name: "manager2", err: errors.New("connection failed")},
	)
	if err != nil {
		t.Fatalf("setupHealthEngine() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	ginEngine.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}

	var response HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Status != "degraded" {
		t.Errorf("status = %s, want degraded", response.Status)
	}

	if response.Managers["manager1"] != "ok" {
		t.Errorf("manager1 status = %s, want ok", response.Managers["manager1"])
	}

	if response.Managers["manager2"] != "unhealthy: connection failed" {
		t.Errorf("manager2 status = %s, want 'unhealthy: connection failed'", response.Managers["manager2"])
	}
}

func TestHealthHandler_MultipleManagers(t *testing.T) {
	_, ginEngine, err := setupHealthEngine(
		&healthyManager{name: "db"},
		&healthyManager{name: "cache"},
		&healthyManager{name: "queue"},
	)
	if err != nil {
		t.Fatalf("setupHealthEngine() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	ginEngine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	var response HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(response.Managers) != 3 {
		t.Errorf("managers count = %d, want 3", len(response.Managers))
	}
}

func TestHealthHandler_NoManagers(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	if err := engine.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	engine.ginEngine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	var response HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("status = %s, want ok", response.Status)
	}
}

func TestHealthHandler_HealthzEndpoint(t *testing.T) {
	_, ginEngine, err := setupHealthEngine(&healthyManager{name: "manager1"})
	if err != nil {
		t.Fatalf("setupHealthEngine() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	ginEngine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestLivenessHandler(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	if err := engine.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/live", nil)
	w := httptest.NewRecorder()
	engine.ginEngine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["status"] != "alive" {
		t.Errorf("status = %v, want alive", response["status"])
	}

	if response["timestamp"] == nil {
		t.Error("timestamp should not be nil")
	}
}

func TestReadinessHandler_AllReady(t *testing.T) {
	_, ginEngine, err := setupHealthEngine(&healthyManager{name: "manager1"})
	if err != nil {
		t.Fatalf("setupHealthEngine() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()
	ginEngine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["status"] != "ready" {
		t.Errorf("status = %v, want ready", response["status"])
	}
}

func TestReadinessHandler_NotReady(t *testing.T) {
	_, ginEngine, err := setupHealthEngine(
		&unhealthyManager{name: "manager1", err: errors.New("not ready")},
	)
	if err != nil {
		t.Fatalf("setupHealthEngine() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()
	ginEngine.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["status"] != "not_ready" {
		t.Errorf("status = %v, want not_ready", response["status"])
	}
}

func TestDetailedHealthHandler_AllHealthy(t *testing.T) {
	engine, err := NewEngine(
		RegisterManagers(
			&healthyManager{name: "db"},
			&healthyManager{name: "cache"},
		),
	)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	if err := engine.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Register the detailed health route
	engine.RegisterDetailedHealthRoute("/health/detail")

	req := httptest.NewRequest("GET", "/health/detail", nil)
	w := httptest.NewRecorder()
	engine.ginEngine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	var response DetailedHealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("status = %s, want healthy", response.Status)
	}

	if len(response.Managers) != 2 {
		t.Errorf("managers count = %d, want 2", len(response.Managers))
	}

	// Check each manager status
	managerMap := make(map[string]ManagerHealthStatus)
	for _, mgr := range response.Managers {
		managerMap[mgr.Name] = mgr
	}

	if managerMap["db"].Status != "healthy" {
		t.Errorf("db status = %s, want healthy", managerMap["db"].Status)
	}

	if managerMap["cache"].Status != "healthy" {
		t.Errorf("cache status = %s, want healthy", managerMap["cache"].Status)
	}
}

func TestDetailedHealthHandler_SomeUnhealthy(t *testing.T) {
	engine, err := NewEngine(
		RegisterManagers(
			&healthyManager{name: "db"},
			&unhealthyManager{name: "cache", err: errors.New("connection timeout")},
		),
	)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	if err := engine.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	engine.RegisterDetailedHealthRoute("/health/detail")

	req := httptest.NewRequest("GET", "/health/detail", nil)
	w := httptest.NewRecorder()
	engine.ginEngine.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}

	var response DetailedHealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Status != "unhealthy" {
		t.Errorf("status = %s, want unhealthy", response.Status)
	}

	managerMap := make(map[string]ManagerHealthStatus)
	for _, mgr := range response.Managers {
		managerMap[mgr.Name] = mgr
	}

	if managerMap["db"].Status != "healthy" {
		t.Errorf("db status = %s, want healthy", managerMap["db"].Status)
	}

	if managerMap["cache"].Status != "unhealthy" {
		t.Errorf("cache status = %s, want unhealthy", managerMap["cache"].Status)
	}

	if managerMap["cache"].Error != "connection timeout" {
		t.Errorf("cache error = %s, want 'connection timeout'", managerMap["cache"].Error)
	}
}

func TestRegisterDetailedHealthRoute(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	if err := engine.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	customPath := "/custom/health"
	engine.RegisterDetailedHealthRoute(customPath)

	req := httptest.NewRequest("GET", customPath, nil)
	w := httptest.NewRecorder()
	engine.ginEngine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHealthResponse_Structure(t *testing.T) {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: "2024-01-01T00:00:00Z",
		Managers: map[string]string{
			"db":    "ok",
			"cache": "ok",
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var decoded HealthResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if decoded.Status != response.Status {
		t.Errorf("status = %s, want %s", decoded.Status, response.Status)
	}

	if len(decoded.Managers) != 2 {
		t.Errorf("managers count = %d, want 2", len(decoded.Managers))
	}
}

func TestManagerHealthStatus_Structure(t *testing.T) {
	status := ManagerHealthStatus{
		Name:   "test-manager",
		Status: "healthy",
	}

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("failed to marshal status: %v", err)
	}

	var decoded ManagerHealthStatus
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal status: %v", err)
	}

	if decoded.Name != status.Name {
		t.Errorf("name = %s, want %s", decoded.Name, status.Name)
	}

	if decoded.Status != status.Status {
		t.Errorf("status = %s, want %s", decoded.Status, status.Status)
	}
}

func TestDetailedHealthResponse_Structure(t *testing.T) {
	response := DetailedHealthResponse{
		Status:    "healthy",
		Timestamp: "2024-01-01T00:00:00Z",
		Managers: []ManagerHealthStatus{
			{Name: "db", Status: "healthy"},
			{Name: "cache", Status: "unhealthy", Error: "timeout"},
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var decoded DetailedHealthResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if decoded.Status != response.Status {
		t.Errorf("status = %s, want %s", decoded.Status, response.Status)
	}

	if len(decoded.Managers) != 2 {
		t.Errorf("managers count = %d, want 2", len(decoded.Managers))
	}

	if decoded.Managers[1].Error != "timeout" {
		t.Errorf("error = %s, want timeout", decoded.Managers[1].Error)
	}
}
