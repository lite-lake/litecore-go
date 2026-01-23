package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestWrapResponseWriter(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"成功包装ResponseWriter"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			engine := gin.New()

			engine.GET("/test", func(c *gin.Context) {
				wrapped := wrapResponseWriter(c.Writer)
				assert.NotNil(t, wrapped)
				assert.IsType(t, &responseWriterWrapper{}, wrapped)
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			engine.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestResponseWriterWrapper_WriteHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.GET("/test", func(c *gin.Context) {
		wrapped := wrapResponseWriter(c.Writer)
		wrapper := wrapped.(*responseWriterWrapper)

		wrapper.WriteHeader(http.StatusCreated)
		c.JSON(http.StatusCreated, gin.H{"created": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestResponseWriterWrapper_嵌入接口(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.GET("/test", func(c *gin.Context) {
		wrapped := wrapResponseWriter(c.Writer)
		wrapper := wrapped.(*responseWriterWrapper)

		assert.NotNil(t, wrapper.ResponseWriter)

		testData := []byte("test content")
		_, err := wrapper.Write(testData)
		assert.NoError(t, err)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.Contains(t, w.Body.String(), "test content")
}

func BenchmarkWrapResponseWriter(b *testing.B) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.GET("/test", func(c *gin.Context) {
		wrapped := wrapResponseWriter(c.Writer)
		_ = wrapped
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
	}
}

func BenchmarkResponseWriterWrapper_Write(b *testing.B) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	testData := []byte("test content")

	engine.GET("/test", func(c *gin.Context) {
		wrapped := wrapResponseWriter(c.Writer)
		wrapper := wrapped.(*responseWriterWrapper)
		_, _ = wrapper.Write(testData)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
	}
}
