package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/container"
	"github.com/lite-lake/litecore-go/manager/databasemgr"
)

// unhealthyManager 用于验证聚合逻辑未被放宽：任意真实 Manager 不健康仍须导致 503。
type unhealthyManager struct{}

func (m *unhealthyManager) ManagerName() string { return "unhealthyTestManager" }
func (m *unhealthyManager) Health() error       { return errors.New("dependency unavailable") }
func (m *unhealthyManager) OnStart() error      { return nil }
func (m *unhealthyManager) OnStop() error       { return nil }

// newReadinessTestEngine 构造仅含系统路由的最小 Engine，不执行 Initialize。
func newReadinessTestEngine(t *testing.T, managers ...common.IBaseManager) *Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	mgrContainer := container.NewManagerContainer()
	for _, mgr := range managers {
		if err := container.RegisterManager[common.IBaseManager](mgrContainer, mgr); err != nil {
			t.Fatalf("register manager %s: %v", mgr.ManagerName(), err)
		}
	}

	engine := &Engine{
		ginEngine: gin.New(),
		Manager:   mgrContainer,
		ready:     true,
	}
	engine.registerSystemRoutes()
	return engine
}

func requestReadiness(t *testing.T, engine *Engine, path string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	engine.ginEngine.ServeHTTP(w, req)

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body %q: %v", w.Body.String(), err)
	}
	return w, body
}

// TestReadiness_NoneDriverHealthy 显式 database.driver: none 应视为无依赖、就绪健康。
func TestReadiness_NoneDriverHealthy(t *testing.T) {
	engine := newReadinessTestEngine(t, databasemgr.NewDatabaseManagerNoneImpl(nil, nil))

	w, body := requestReadiness(t, engine, "/api/ready")

	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/ready status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	if body["status"] != "ok" {
		t.Errorf("status = %v, want ok", body["status"])
	}

	managers, ok := body["managers"].(map[string]any)
	if !ok {
		t.Fatalf("managers field missing or wrong type: %v", body["managers"])
	}
	if got := managers["databaseManagerNoneImpl"]; got != "ok" {
		t.Errorf("managers.databaseManagerNoneImpl = %v, want ok", got)
	}
}

// TestReadiness_UnhealthyManagerStillDegraded 真实 Manager 不健康时仍须 503，防止聚合逻辑被放宽。
func TestReadiness_UnhealthyManagerStillDegraded(t *testing.T) {
	engine := newReadinessTestEngine(t, &unhealthyManager{})

	w, body := requestReadiness(t, engine, "/api/ready")

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("GET /api/ready status = %d, want 503; body=%s", w.Code, w.Body.String())
	}
	if body["status"] != "degraded" {
		t.Errorf("status = %v, want degraded", body["status"])
	}

	managers, ok := body["managers"].(map[string]any)
	if !ok {
		t.Fatalf("managers field missing or wrong type: %v", body["managers"])
	}
	got, _ := managers["unhealthyTestManager"].(string)
	if got != "unhealthy: dependency unavailable" {
		t.Errorf("managers.unhealthyTestManager = %v, want unhealthy detail", managers["unhealthyTestManager"])
	}
}

// TestReadiness_NotReadyStill503 引擎未完成启动时仍为 503 not_ready，即便依赖全部健康。
func TestReadiness_NotReadyStill503(t *testing.T) {
	engine := newReadinessTestEngine(t, databasemgr.NewDatabaseManagerNoneImpl(nil, nil))
	engine.ready = false

	w, body := requestReadiness(t, engine, "/api/ready")

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("GET /api/ready status = %d, want 503; body=%s", w.Code, w.Body.String())
	}
	if body["status"] != "not_ready" {
		t.Errorf("status = %v, want not_ready", body["status"])
	}
}

// TestHealth_UnchangedForNoneDriver liveness 语义不变：none 驱动下 /api/health 仍为 200 ok。
func TestHealth_UnchangedForNoneDriver(t *testing.T) {
	engine := newReadinessTestEngine(t, databasemgr.NewDatabaseManagerNoneImpl(nil, nil))

	w, body := requestReadiness(t, engine, "/api/health")

	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/health status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	if body["status"] != "ok" {
		t.Errorf("status = %v, want ok", body["status"])
	}
}
