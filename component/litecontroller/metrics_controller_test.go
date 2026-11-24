package litecontroller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/lite-lake/litecore-go/common/metrics"
)

// TestMetricsControllerMeta 路由与控制器名保持稳定。
func TestMetricsControllerMeta(t *testing.T) {
	controller := NewMetricsController()
	assert.NotNil(t, controller)
	assert.Equal(t, "MetricsController", controller.ControllerName())
	assert.Equal(t, "/metrics [GET]", controller.GetRouter())
}

// TestMetricsControllerHandle 输出 Prometheus 文本与正确的 Content-Type。
func TestMetricsControllerHandle(t *testing.T) {
	metrics.ResetForTest()
	t.Cleanup(metrics.ResetForTest)

	metrics.Register("demo_requests_total", "演示请求总数", metrics.Counter)
	metrics.Add("demo_requests_total", 3, metrics.Labels{"tier": "free"})

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/metrics", NewMetricsController().Handle)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain; version=0.0.4; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "# HELP demo_requests_total 演示请求总数\n")
	assert.Contains(t, w.Body.String(), "# TYPE demo_requests_total counter\n")
	assert.Contains(t, w.Body.String(), `demo_requests_total{tier="free"} 3`)
}

// TestMetricsControllerHandleEmpty 空注册表返回 200 与空响应体。
func TestMetricsControllerHandleEmpty(t *testing.T) {
	metrics.ResetForTest()
	t.Cleanup(metrics.ResetForTest)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/metrics", NewMetricsController().Handle)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain; version=0.0.4; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Empty(t, w.Body.String())
}
