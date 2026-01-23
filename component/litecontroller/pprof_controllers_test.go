package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewPprofIndexController(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"成功创建PprofIndexController"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewPprofIndexController()
			assert.NotNil(t, controller)
			assert.Equal(t, "PprofIndexController", controller.ControllerName())
			assert.Equal(t, "/debug/pprof [GET]", controller.GetRouter())
		})
	}
}

func TestPprofIndexController_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	controller := NewPprofIndexController()
	engine.GET("/debug/pprof", controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "/debug/pprof/")
}

func TestNewPprofHeapController(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"成功创建PprofHeapController"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewPprofHeapController()
			assert.NotNil(t, controller)
			assert.Equal(t, "PprofHeapController", controller.ControllerName())
			assert.Equal(t, "/debug/pprof/heap [GET]", controller.GetRouter())
		})
	}
}

func TestPprofHeapController_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	controller := NewPprofHeapController()
	engine.GET("/debug/pprof/heap", controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/heap", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewPprofGoroutineController(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"成功创建PprofGoroutineController"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewPprofGoroutineController()
			assert.NotNil(t, controller)
			assert.Equal(t, "PprofGoroutineController", controller.ControllerName())
			assert.Equal(t, "/debug/pprof/goroutine [GET]", controller.GetRouter())
		})
	}
}

func TestPprofGoroutineController_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	controller := NewPprofGoroutineController()
	engine.GET("/debug/pprof/goroutine", controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/goroutine", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewPprofAllocsController(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"成功创建PprofAllocsController"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewPprofAllocsController()
			assert.NotNil(t, controller)
			assert.Equal(t, "PprofAllocsController", controller.ControllerName())
			assert.Equal(t, "/debug/pprof/allocs [GET]", controller.GetRouter())
		})
	}
}

func TestPprofAllocsController_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	controller := NewPprofAllocsController()
	engine.GET("/debug/pprof/allocs", controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/allocs", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewPprofBlockController(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"成功创建PprofBlockController"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewPprofBlockController()
			assert.NotNil(t, controller)
			assert.Equal(t, "PprofBlockController", controller.ControllerName())
			assert.Equal(t, "/debug/pprof/block [GET]", controller.GetRouter())
		})
	}
}

func TestPprofBlockController_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	controller := NewPprofBlockController()
	engine.GET("/debug/pprof/block", controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/block", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewPprofMutexController(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"成功创建PprofMutexController"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewPprofMutexController()
			assert.NotNil(t, controller)
			assert.Equal(t, "PprofMutexController", controller.ControllerName())
			assert.Equal(t, "/debug/pprof/mutex [GET]", controller.GetRouter())
		})
	}
}

func TestPprofMutexController_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	controller := NewPprofMutexController()
	engine.GET("/debug/pprof/mutex", controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/mutex", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewPprofProfileController(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"成功创建PprofProfileController"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewPprofProfileController()
			assert.NotNil(t, controller)
			assert.Equal(t, "PprofProfileController", controller.ControllerName())
			assert.Equal(t, "/debug/pprof/profile [GET]", controller.GetRouter())
		})
	}
}

func TestPprofProfileController_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	controller := NewPprofProfileController()
	engine.GET("/debug/pprof/profile", controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/profile", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewPprofTraceController(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"成功创建PprofTraceController"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewPprofTraceController()
			assert.NotNil(t, controller)
			assert.Equal(t, "PprofTraceController", controller.ControllerName())
			assert.Equal(t, "/debug/pprof/trace [GET]", controller.GetRouter())
		})
	}
}

func TestPprofTraceController_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	controller := NewPprofTraceController()
	engine.GET("/debug/pprof/trace", controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/trace", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewPprofSymbolController(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"成功创建PprofSymbolController"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewPprofSymbolController()
			assert.NotNil(t, controller)
			assert.Equal(t, "PprofSymbolController", controller.ControllerName())
			assert.Equal(t, "/debug/pprof/symbol [GET]", controller.GetRouter())
		})
	}
}

func TestPprofSymbolController_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	controller := NewPprofSymbolController()
	engine.GET("/debug/pprof/symbol", controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/symbol", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewPprofSymbolPostController(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"成功创建PprofSymbolPostController"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewPprofSymbolPostController()
			assert.NotNil(t, controller)
			assert.Equal(t, "PprofSymbolPostController", controller.ControllerName())
			assert.Equal(t, "/debug/pprof/symbol [POST]", controller.GetRouter())
		})
	}
}

func TestPprofSymbolPostController_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	controller := NewPprofSymbolPostController()
	engine.POST("/debug/pprof/symbol", controller.Handle)

	req := httptest.NewRequest(http.MethodPost, "/debug/pprof/symbol", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewPprofCmdlineController(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"成功创建PprofCmdlineController"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewPprofCmdlineController()
			assert.NotNil(t, controller)
			assert.Equal(t, "PprofCmdlineController", controller.ControllerName())
			assert.Equal(t, "/debug/pprof/cmdline [GET]", controller.GetRouter())
		})
	}
}

func TestPprofCmdlineController_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	controller := NewPprofCmdlineController()
	engine.GET("/debug/pprof/cmdline", controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/cmdline", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewPprofThreadcreateController(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"成功创建PprofThreadcreateController"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewPprofThreadcreateController()
			assert.NotNil(t, controller)
			assert.Equal(t, "PprofThreadcreateController", controller.ControllerName())
			assert.Equal(t, "/debug/pprof/threadcreate [GET]", controller.GetRouter())
		})
	}
}

func TestPprofThreadcreateController_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	controller := NewPprofThreadcreateController()
	engine.GET("/debug/pprof/threadcreate", controller.Handle)

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/threadcreate", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func BenchmarkPprofIndexController_Handle(b *testing.B) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	controller := NewPprofIndexController()
	engine.GET("/debug/pprof", controller.Handle)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/debug/pprof", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
	}
}
