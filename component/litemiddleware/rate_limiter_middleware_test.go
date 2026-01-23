package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/lite-lake/litecore-go/common"
)

type MockLimiterManager struct {
	mock.Mock
}

func (m *MockLimiterManager) ManagerName() string {
	return "mockLimiter"
}

func (m *MockLimiterManager) Health() error {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockLimiterManager) OnStart() error {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockLimiterManager) OnStop() error {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockLimiterManager) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	args := m.Called(ctx, key, limit, window)
	return args.Bool(0), args.Error(1)
}

func (m *MockLimiterManager) GetRemaining(ctx context.Context, key string, limit int, window time.Duration) (int, error) {
	args := m.Called(ctx, key, limit, window)
	return args.Int(0), args.Error(1)
}

func setupRouter(middleware common.IBaseMiddleware) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Wrapper())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	return router
}

func TestNewRateLimiter(t *testing.T) {
	t.Run("默认配置", func(t *testing.T) {
		mw := NewRateLimiterMiddleware(nil)
		rlmw, ok := mw.(*rateLimiterMiddleware)
		assert.True(t, ok)
		assert.NotNil(t, rlmw.config)
		assert.Equal(t, 100, rlmw.config.Limit)
		assert.Equal(t, time.Minute, rlmw.config.Window)
		assert.Equal(t, "rate_limit", rlmw.config.KeyPrefix)
	})

	t.Run("自定义配置", func(t *testing.T) {
		customConfig := &RateLimiterConfig{
			Limit:     10,
			Window:    time.Hour,
			KeyPrefix: "custom",
		}
		mw := NewRateLimiterMiddleware(customConfig)
		rlmw, ok := mw.(*rateLimiterMiddleware)
		assert.True(t, ok)
		assert.Equal(t, 10, rlmw.config.Limit)
		assert.Equal(t, time.Hour, rlmw.config.Window)
		assert.Equal(t, "custom", rlmw.config.KeyPrefix)
	})

	t.Run("按IP限流", func(t *testing.T) {
		mw := NewRateLimiterMiddleware(&RateLimiterConfig{
			Limit:     50,
			Window:    time.Minute,
			KeyPrefix: "ip",
		})
		rlmw, ok := mw.(*rateLimiterMiddleware)
		assert.True(t, ok)
		assert.Equal(t, 50, rlmw.config.Limit)
		assert.Equal(t, time.Minute, rlmw.config.Window)
		assert.Equal(t, "ip", rlmw.config.KeyPrefix)
	})

	t.Run("按路径限流", func(t *testing.T) {
		mw := NewRateLimiterMiddleware(&RateLimiterConfig{
			Limit:     20,
			Window:    time.Hour,
			KeyPrefix: "path",
		})
		rlmw, ok := mw.(*rateLimiterMiddleware)
		assert.True(t, ok)
		assert.Equal(t, 20, rlmw.config.Limit)
		assert.Equal(t, time.Hour, rlmw.config.Window)
		assert.Equal(t, "path", rlmw.config.KeyPrefix)
	})
}

func TestRateLimiterMiddleware_Allow(t *testing.T) {
	t.Run("允许请求", func(t *testing.T) {
		mockLimiter := new(MockLimiterManager)
		mockLimiter.On("Allow", mock.Anything, mock.AnythingOfType("string"), 100, time.Minute).Return(true, nil)
		mockLimiter.On("GetRemaining", mock.Anything, mock.AnythingOfType("string"), 100, time.Minute).Return(99, nil)

		mw := NewRateLimiterMiddleware(&RateLimiterConfig{
			Limit:     100,
			Window:    time.Minute,
			KeyPrefix: "ip",
		})
		rlmw, ok := mw.(*rateLimiterMiddleware)
		assert.True(t, ok)
		rlmw.LimiterMgr = mockLimiter

		router := setupRouter(mw)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
		assert.Equal(t, "100", w.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "99", w.Header().Get("X-RateLimit-Remaining"))
		mockLimiter.AssertExpectations(t)
	})
}

func TestRateLimiterMiddleware_Reject(t *testing.T) {
	t.Run("拒绝请求", func(t *testing.T) {
		mockLimiter := new(MockLimiterManager)
		mockLimiter.On("Allow", mock.Anything, mock.AnythingOfType("string"), 100, time.Minute).Return(false, nil)
		mockLimiter.On("GetRemaining", mock.Anything, mock.AnythingOfType("string"), 100, time.Minute).Return(0, nil)

		mw := NewRateLimiterMiddleware(&RateLimiterConfig{
			Limit:     100,
			Window:    time.Minute,
			KeyPrefix: "ip",
		})
		rlmw, ok := mw.(*rateLimiterMiddleware)
		assert.True(t, ok)
		rlmw.LimiterMgr = mockLimiter

		router := setupRouter(mw)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Contains(t, w.Body.String(), "RATE_LIMIT_EXCEEDED")
		assert.Equal(t, "60", w.Header().Get("Retry-After"))
		mockLimiter.AssertExpectations(t)
	})
}

func TestRateLimiterMiddleware_Error(t *testing.T) {
	t.Run("限流服务错误", func(t *testing.T) {
		mockLimiter := new(MockLimiterManager)
		mockLimiter.On("Allow", mock.Anything, mock.AnythingOfType("string"), 100, time.Minute).Return(false, errors.New("limiter error"))

		mw := NewRateLimiterMiddleware(&RateLimiterConfig{
			Limit:     100,
			Window:    time.Minute,
			KeyPrefix: "ip",
		})
		rlmw, ok := mw.(*rateLimiterMiddleware)
		assert.True(t, ok)
		rlmw.LimiterMgr = mockLimiter

		router := setupRouter(mw)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "限流服务异常")
		mockLimiter.AssertExpectations(t)
	})
}

func TestRateLimiterMiddleware_Skip(t *testing.T) {
	t.Run("跳过限流", func(t *testing.T) {
		mockLimiter := new(MockLimiterManager)

		mw := NewRateLimiterMiddleware(&RateLimiterConfig{
			Limit:  100,
			Window: time.Minute,
			SkipFunc: func(c *gin.Context) bool {
				return c.GetHeader("X-Skip-Limit") == "true"
			},
		})
		rlmw, ok := mw.(*rateLimiterMiddleware)
		assert.True(t, ok)
		rlmw.LimiterMgr = mockLimiter

		router := setupRouter(mw)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Skip-Limit", "true")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockLimiter.AssertNotCalled(t, "Allow")
	})
}

func TestRateLimiterMiddleware_NilLimiter(t *testing.T) {
	t.Run("限流管理器未初始化", func(t *testing.T) {
		mw := NewRateLimiterMiddleware(&RateLimiterConfig{
			Limit:     100,
			Window:    time.Minute,
			KeyPrefix: "ip",
		})
		router := setupRouter(mw)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestRateLimiterMiddleware_CustomKeyFunc(t *testing.T) {
	t.Run("自定义key函数", func(t *testing.T) {
		mockLimiter := new(MockLimiterManager)
		mockLimiter.On("Allow", mock.Anything, "custom:test-user", 50, time.Minute).Return(true, nil)
		mockLimiter.On("GetRemaining", mock.Anything, "custom:test-user", 50, time.Minute).Return(49, nil)

		mw := NewRateLimiterMiddleware(&RateLimiterConfig{
			Limit:     50,
			Window:    time.Minute,
			KeyPrefix: "custom",
			KeyFunc: func(c *gin.Context) string {
				return c.GetHeader("X-User-ID")
			},
		})
		rlmw, ok := mw.(*rateLimiterMiddleware)
		assert.True(t, ok)
		rlmw.LimiterMgr = mockLimiter

		router := setupRouter(mw)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-User-ID", "test-user")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockLimiter.AssertExpectations(t)
	})
}

func TestRateLimiterMiddleware_BasicMethods(t *testing.T) {
	mw := NewRateLimiterMiddleware(&RateLimiterConfig{
		Limit:     100,
		Window:    time.Minute,
		KeyPrefix: "ip",
	})
	rlmw, ok := mw.(*rateLimiterMiddleware)
	assert.True(t, ok)

	assert.Equal(t, "RateLimiterMiddleware", rlmw.MiddlewareName())
	assert.Equal(t, OrderRateLimiter, rlmw.Order())
	assert.NoError(t, rlmw.OnStart())
	assert.NoError(t, rlmw.OnStop())
}
