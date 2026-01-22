package middleware

import (
	"github.com/lite-lake/litecore-go/common"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockRequestLogger struct {
	debugMsgs []string
	infoMsgs  []string
	warnMsgs  []string
	errorMsgs []string
	fatalMsgs []string
}

func (m *mockRequestLogger) Debug(msg string, args ...any) {
	m.debugMsgs = append(m.debugMsgs, msg)
}

func (m *mockRequestLogger) Info(msg string, args ...any) {
	m.infoMsgs = append(m.infoMsgs, msg)
}

func (m *mockRequestLogger) Warn(msg string, args ...any) {
	m.warnMsgs = append(m.warnMsgs, msg)
}

func (m *mockRequestLogger) Error(msg string, args ...any) {
	m.errorMsgs = append(m.errorMsgs, msg)
}

func (m *mockRequestLogger) Fatal(msg string, args ...any) {
	m.fatalMsgs = append(m.fatalMsgs, msg)
}

func (m *mockRequestLogger) With(args ...any) common.ILogger {
	return m
}

func (m *mockRequestLogger) SetLevel(level common.LogLevel) {
}

func TestNewRequestLoggerMiddleware(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "创建请求日志中间件",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewRequestLoggerMiddleware()
			assert.NotNil(t, middleware)
			assert.IsType(t, &RequestLoggerMiddleware{}, middleware)
		})
	}
}

func TestRequestLoggerMiddleware_MiddlewareName(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "返回中间件名称",
			expected: "RequestLoggerMiddleware",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewRequestLoggerMiddleware().(*RequestLoggerMiddleware)
			assert.Equal(t, tt.expected, middleware.MiddlewareName())
		})
	}
}

func TestRequestLoggerMiddleware_Order(t *testing.T) {
	tests := []struct {
		name     string
		expected int
	}{
		{
			name:     "返回执行顺序",
			expected: 20,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewRequestLoggerMiddleware().(*RequestLoggerMiddleware)
			assert.Equal(t, tt.expected, middleware.Order())
		})
	}
}

func TestRequestLoggerMiddleware_Wrapper(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		statusCode int
		withLogger bool
		withError  bool
		hasBody    bool
	}{
		{
			name:       "GET请求成功",
			method:     http.MethodGet,
			statusCode: http.StatusOK,
			withLogger: true,
		},
		{
			name:       "POST请求成功",
			method:     http.MethodPost,
			statusCode: http.StatusOK,
			withLogger: true,
			hasBody:    true,
		},
		{
			name:       "PUT请求成功",
			method:     http.MethodPut,
			statusCode: http.StatusOK,
			withLogger: true,
			hasBody:    true,
		},
		{
			name:       "DELETE请求成功",
			method:     http.MethodDelete,
			statusCode: http.StatusOK,
			withLogger: true,
		},
		{
			name:       "请求失败",
			method:     http.MethodGet,
			statusCode: http.StatusBadRequest,
			withLogger: true,
			withError:  true,
		},
		{
			name:       "PATCH请求成功",
			method:     http.MethodPatch,
			statusCode: http.StatusOK,
			withLogger: true,
			hasBody:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			middleware := NewRequestLoggerMiddleware().(*RequestLoggerMiddleware)
			if tt.withLogger {
				middleware.Logger = &mockRequestLogger{}
			}
			router.Use(middleware.Wrapper())

			router.GET("/test", func(c *gin.Context) {
				if tt.withError {
					c.Error(assert.AnError)
				}
				c.JSON(tt.statusCode, gin.H{"message": "ok"})
			})
			router.POST("/test", func(c *gin.Context) {
				if tt.withError {
					c.Error(assert.AnError)
				}
				c.JSON(tt.statusCode, gin.H{"message": "ok"})
			})
			router.PUT("/test", func(c *gin.Context) {
				if tt.withError {
					c.Error(assert.AnError)
				}
				c.JSON(tt.statusCode, gin.H{"message": "ok"})
			})
			router.DELETE("/test", func(c *gin.Context) {
				if tt.withError {
					c.Error(assert.AnError)
				}
				c.JSON(tt.statusCode, gin.H{"message": "ok"})
			})
			router.PATCH("/test", func(c *gin.Context) {
				if tt.withError {
					c.Error(assert.AnError)
				}
				c.JSON(tt.statusCode, gin.H{"message": "ok"})
			})

			var body string
			if tt.hasBody {
				body = `{"test": "data"}`
			}

			req := httptest.NewRequest(tt.method, "/test", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func TestRequestLoggerMiddleware_getRequestID(t *testing.T) {
	tests := []struct {
		name            string
		headerRequestID string
		contextID       string
	}{
		{
			name:            "从X-Request-Id头获取请求ID",
			headerRequestID: "test-id-1",
		},
		{
			name:            "从X-Request-ID头获取请求ID",
			headerRequestID: "test-id-2",
		},
		{
			name:            "无请求ID时生成新ID",
			headerRequestID: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			middleware := NewRequestLoggerMiddleware().(*RequestLoggerMiddleware)

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

			if tt.name == "从X-Request-Id头获取请求ID" {
				c.Request.Header.Set("X-Request-Id", tt.headerRequestID)
			} else if tt.name == "从X-Request-ID头获取请求ID" {
				c.Request.Header.Set("X-Request-ID", tt.headerRequestID)
			}

			requestID := middleware.getRequestID(c)

			if tt.headerRequestID != "" {
				assert.Equal(t, tt.headerRequestID, requestID)
				assert.Equal(t, tt.headerRequestID, c.GetString("request_id"))
			} else {
				assert.NotEmpty(t, requestID)
				assert.Equal(t, requestID, c.GetString("request_id"))
			}
		})
	}
}

func TestRequestLoggerMiddleware_OnStart(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "服务器启动回调",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewRequestLoggerMiddleware().(*RequestLoggerMiddleware)
			err := middleware.OnStart()
			assert.NoError(t, err)
		})
	}
}

func TestRequestLoggerMiddleware_OnStop(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "服务器停止回调",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewRequestLoggerMiddleware().(*RequestLoggerMiddleware)
			err := middleware.OnStop()
			assert.NoError(t, err)
		})
	}
}

func TestRequestLoggerMiddleware_WithError(t *testing.T) {
	tests := []struct {
		name          string
		setErrorCount int
	}{
		{
			name:          "一个错误",
			setErrorCount: 1,
		},
		{
			name:          "多个错误",
			setErrorCount: 3,
		},
		{
			name:          "无错误",
			setErrorCount: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			middleware := NewRequestLoggerMiddleware().(*RequestLoggerMiddleware)
			middleware.Logger = &mockRequestLogger{}
			router.Use(middleware.Wrapper())

			router.GET("/test", func(c *gin.Context) {
				for i := 0; i < tt.setErrorCount; i++ {
					c.Error(assert.AnError)
				}
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestRequestLoggerMiddleware_RequestBodyHandling(t *testing.T) {
	tests := []struct {
		name   string
		method string
		body   string
		empty  bool
	}{
		{
			name:   "POST请求带请求体",
			method: http.MethodPost,
			body:   `{"key": "value"}`,
		},
		{
			name:   "PUT请求带请求体",
			method: http.MethodPut,
			body:   `{"key": "value"}`,
		},
		{
			name:   "GET请求不带请求体",
			method: http.MethodGet,
			body:   "",
		},
		{
			name:   "DELETE请求不带请求体",
			method: http.MethodDelete,
			body:   "",
		},
		{
			name:   "POST请求空请求体",
			method: http.MethodPost,
			body:   "",
			empty:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			middleware := NewRequestLoggerMiddleware().(*RequestLoggerMiddleware)
			middleware.Logger = &mockRequestLogger{}
			router.Use(middleware.Wrapper())

			var handler gin.HandlerFunc
			if tt.name == "POST请求带请求体" {
				handler = func(c *gin.Context) {
					var data map[string]string
					if err := c.ShouldBindJSON(&data); err == nil {
						assert.Equal(t, "value", data["key"])
					}
					c.JSON(http.StatusOK, gin.H{"message": "ok"})
				}
			} else {
				handler = func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "ok"})
				}
			}

			router.POST("/test", handler)
			router.PUT("/test", handler)
			router.GET("/test", handler)
			router.DELETE("/test", handler)

			req := httptest.NewRequest(tt.method, "/test", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}
