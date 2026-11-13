package configmgr

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsConfigKeyNotFound(t *testing.T) {
	t.Run("nil 错误", func(t *testing.T) {
		assert.False(t, IsConfigKeyNotFound(nil))
	})

	t.Run("包含 'not found' 的错误", func(t *testing.T) {
		err := errors.New("configmgr key 'server.host' not found")
		assert.True(t, IsConfigKeyNotFound(err))
	})

	t.Run("不包含 'not found' 的错误", func(t *testing.T) {
		err := errors.New("some other error")
		assert.False(t, IsConfigKeyNotFound(err))
	})

	t.Run("使用 errors.Is", func(t *testing.T) {
		assert.True(t, IsConfigKeyNotFound(ErrKeyNotFound))
	})
}

func TestGet(t *testing.T) {
	setupManager := func() IConfigManager {
		handler := func() (map[string]any, error) {
			return map[string]any{
				"name":     "test",
				"port":     8080,
				"enabled":  true,
				"timeout":  30.5,
				"database": map[string]any{"host": "localhost"},
			}, nil
		}
		mgr, _ := newBaseConfigManager("Test", handler)
		return mgr
	}

	t.Run("获取字符串类型", func(t *testing.T) {
		mgr := setupManager()
		name, err := Get[string](mgr, "name")
		assert.NoError(t, err)
		assert.Equal(t, "test", name)
	})

	t.Run("获取 int 类型", func(t *testing.T) {
		mgr := setupManager()
		port, err := Get[int](mgr, "port")
		assert.NoError(t, err)
		assert.Equal(t, 8080, port)
	})

	t.Run("获取 float64 类型", func(t *testing.T) {
		mgr := setupManager()
		timeout, err := Get[float64](mgr, "timeout")
		assert.NoError(t, err)
		assert.Equal(t, 30.5, timeout)
	})

	t.Run("获取 bool 类型", func(t *testing.T) {
		mgr := setupManager()
		enabled, err := Get[bool](mgr, "enabled")
		assert.NoError(t, err)
		assert.True(t, enabled)
	})

	t.Run("获取 map 类型", func(t *testing.T) {
		mgr := setupManager()
		db, err := Get[map[string]any](mgr, "database")
		assert.NoError(t, err)
		assert.Equal(t, "localhost", db["host"])
	})

	t.Run("键不存在", func(t *testing.T) {
		mgr := setupManager()
		_, err := Get[string](mgr, "notexist")
		assert.Error(t, err)
		assert.True(t, IsConfigKeyNotFound(err))
	})

	t.Run("类型不匹配", func(t *testing.T) {
		mgr := setupManager()
		_, err := Get[int](mgr, "name")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrTypeMismatch))
	})

	t.Run("JSON float64 转 int", func(t *testing.T) {
		handler := func() (map[string]any, error) {
			return map[string]any{"count": 42.0}, nil
		}
		mgr, _ := newBaseConfigManager("Test", handler)

		count, err := Get[int](mgr, "count")
		assert.NoError(t, err)
		assert.Equal(t, 42, count)
	})

	t.Run("JSON float64 转 int64", func(t *testing.T) {
		handler := func() (map[string]any, error) {
			return map[string]any{"count": 42.0}, nil
		}
		mgr, _ := newBaseConfigManager("Test", handler)

		count, err := Get[int64](mgr, "count")
		assert.NoError(t, err)
		assert.Equal(t, int64(42), count)
	})

	t.Run("JSON float64 转 int32", func(t *testing.T) {
		handler := func() (map[string]any, error) {
			return map[string]any{"count": 42.0}, nil
		}
		mgr, _ := newBaseConfigManager("Test", handler)

		count, err := Get[int32](mgr, "count")
		assert.NoError(t, err)
		assert.Equal(t, int32(42), count)
	})

	t.Run("字符串转 bool", func(t *testing.T) {
		handler := func() (map[string]any, error) {
			return map[string]any{"enabled": "true", "disabled": "false"}, nil
		}
		mgr, _ := newBaseConfigManager("Test", handler)

		enabled, err := Get[bool](mgr, "enabled")
		assert.NoError(t, err)
		assert.True(t, enabled)

		disabled, err := Get[bool](mgr, "disabled")
		assert.NoError(t, err)
		assert.False(t, disabled)
	})

	t.Run("字符串转 int", func(t *testing.T) {
		handler := func() (map[string]any, error) {
			return map[string]any{"port": "8080"}, nil
		}
		mgr, _ := newBaseConfigManager("Test", handler)

		port, err := Get[int](mgr, "port")
		assert.NoError(t, err)
		assert.Equal(t, 8080, port)
	})

	t.Run("bool 转 string", func(t *testing.T) {
		handler := func() (map[string]any, error) {
			return map[string]any{"flag": true}, nil
		}
		mgr, _ := newBaseConfigManager("Test", handler)

		flag, err := Get[string](mgr, "flag")
		assert.NoError(t, err)
		assert.Equal(t, "true", flag)
	})

	t.Run("不支持的类型转换", func(t *testing.T) {
		handler := func() (map[string]any, error) {
			return map[string]any{"data": map[string]any{"key": "value"}}, nil
		}
		mgr, _ := newBaseConfigManager("Test", handler)

		_, err := Get[int](mgr, "data")
		assert.Error(t, err)
	})
}

func TestGetWithDefault(t *testing.T) {
	setupManager := func() IConfigManager {
		handler := func() (map[string]any, error) {
			return map[string]any{"name": "test", "port": 8080}, nil
		}
		mgr, _ := newBaseConfigManager("Test", handler)
		return mgr
	}

	t.Run("键存在时返回实际值", func(t *testing.T) {
		mgr := setupManager()
		name := GetWithDefault(mgr, "name", "default")
		assert.Equal(t, "test", name)
	})

	t.Run("键不存在时返回默认值", func(t *testing.T) {
		mgr := setupManager()
		port := GetWithDefault(mgr, "notexist", 3000)
		assert.Equal(t, 3000, port)
	})

	t.Run("类型不匹配时返回默认值", func(t *testing.T) {
		mgr := setupManager()
		port := GetWithDefault(mgr, "name", 3000)
		assert.Equal(t, 3000, port)
	})

	t.Run("不同类型的默认值", func(t *testing.T) {
		mgr := setupManager()

		boolVal := GetWithDefault(mgr, "enabled", false)
		assert.False(t, boolVal)

		intVal := GetWithDefault(mgr, "count", 10)
		assert.Equal(t, 10, intVal)

		stringVal := GetWithDefault(mgr, "title", "default")
		assert.Equal(t, "default", stringVal)
	})
}

func TestConvertType(t *testing.T) {
	t.Run("转 int", func(t *testing.T) {
		result, err := convertType[int](42.0)
		assert.NoError(t, err)
		assert.Equal(t, 42, result)

		result, err = convertType[int]("8080")
		assert.NoError(t, err)
		assert.Equal(t, 8080, result)
	})

	t.Run("转 int64", func(t *testing.T) {
		result, err := convertType[int64](42.0)
		assert.NoError(t, err)
		assert.Equal(t, int64(42), result)
	})

	t.Run("转 int32", func(t *testing.T) {
		result, err := convertType[int32](42.0)
		assert.NoError(t, err)
		assert.Equal(t, int32(42), result)
	})

	t.Run("转 float64", func(t *testing.T) {
		result, err := convertType[float64](42)
		assert.NoError(t, err)
		assert.Equal(t, 42.0, result)

		result, err = convertType[float64]("42.5")
		assert.NoError(t, err)
		assert.Equal(t, 42.5, result)
	})

	t.Run("转 string", func(t *testing.T) {
		result, err := convertType[string](42)
		assert.NoError(t, err)
		assert.Equal(t, "42", result)

		result, err = convertType[string](true)
		assert.NoError(t, err)
		assert.Equal(t, "true", result)
	})

	t.Run("转 bool 从字符串", func(t *testing.T) {
		result, err := convertType[bool]("true")
		assert.NoError(t, err)
		assert.True(t, result)

		result, err = convertType[bool]("false")
		assert.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("转 bool 从布尔", func(t *testing.T) {
		result, err := convertType[bool](true)
		assert.NoError(t, err)
		assert.True(t, result)

		result, err = convertType[bool](false)
		assert.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("不支持的类型转换", func(t *testing.T) {
		_, err := convertType[map[string]any](42)
		assert.Error(t, err)
	})

	t.Run("无法转 bool", func(t *testing.T) {
		_, err := convertType[bool](map[string]any{"key": "value"})
		assert.Error(t, err)
	})
}
