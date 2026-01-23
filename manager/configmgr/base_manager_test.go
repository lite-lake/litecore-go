package configmgr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBaseConfigManager(t *testing.T) {
	t.Run("成功创建配置管理器", func(t *testing.T) {
		handler := func() (map[string]any, error) {
			return map[string]any{"key": "value"}, nil
		}
		mgr, err := newBaseConfigManager("TestManager", handler)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.Equal(t, "TestManager", mgr.ManagerName())
	})

	t.Run("handler 返回错误", func(t *testing.T) {
		handler := func() (map[string]any, error) {
			return nil, assert.AnError
		}
		mgr, err := newBaseConfigManager("TestManager", handler)
		assert.Error(t, err)
		assert.Nil(t, mgr)
	})
}

func TestBaseConfigManager_ManagerName(t *testing.T) {
	handler := func() (map[string]any, error) {
		return map[string]any{}, nil
	}
	mgr, _ := newBaseConfigManager("MyConfig", handler)
	assert.Equal(t, "MyConfig", mgr.ManagerName())
}

func TestBaseConfigManager_Health(t *testing.T) {
	handler := func() (map[string]any, error) {
		return map[string]any{}, nil
	}
	mgr, _ := newBaseConfigManager("Test", handler)
	assert.NoError(t, mgr.Health())
}

func TestBaseConfigManager_OnStart(t *testing.T) {
	handler := func() (map[string]any, error) {
		return map[string]any{}, nil
	}
	mgr, _ := newBaseConfigManager("Test", handler)
	assert.NoError(t, mgr.OnStart())
}

func TestBaseConfigManager_OnStop(t *testing.T) {
	handler := func() (map[string]any, error) {
		return map[string]any{}, nil
	}
	mgr, _ := newBaseConfigManager("Test", handler)
	assert.NoError(t, mgr.OnStop())
}

func TestBaseConfigManager_Get(t *testing.T) {
	t.Run("空路径返回全部配置", func(t *testing.T) {
		data := map[string]any{"name": "test", "port": 8080}
		handler := func() (map[string]any, error) { return data, nil }
		mgr, _ := newBaseConfigManager("Test", handler)

		result, err := mgr.Get("")
		assert.NoError(t, err)
		assert.Equal(t, data, result)
	})

	t.Run("简单键获取", func(t *testing.T) {
		data := map[string]any{"name": "test", "port": 8080}
		handler := func() (map[string]any, error) { return data, nil }
		mgr, _ := newBaseConfigManager("Test", handler)

		name, err := mgr.Get("name")
		assert.NoError(t, err)
		assert.Equal(t, "test", name)

		port, err := mgr.Get("port")
		assert.NoError(t, err)
		assert.Equal(t, 8080, port)
	})

	t.Run("嵌套路径查询", func(t *testing.T) {
		data := map[string]any{
			"server": map[string]any{
				"host": "localhost",
				"port": 8080,
			},
		}
		handler := func() (map[string]any, error) { return data, nil }
		mgr, _ := newBaseConfigManager("Test", handler)

		host, err := mgr.Get("server.host")
		assert.NoError(t, err)
		assert.Equal(t, "localhost", host)

		port, err := mgr.Get("server.port")
		assert.NoError(t, err)
		assert.Equal(t, 8080, port)
	})

	t.Run("深层嵌套路径查询", func(t *testing.T) {
		data := map[string]any{
			"database": map[string]any{
				"mysql": map[string]any{
					"host": "127.0.0.1",
					"port": 3306,
				},
			},
		}
		handler := func() (map[string]any, error) { return data, nil }
		mgr, _ := newBaseConfigManager("Test", handler)

		host, err := mgr.Get("database.mysql.host")
		assert.NoError(t, err)
		assert.Equal(t, "127.0.0.1", host)
	})

	t.Run("键不存在", func(t *testing.T) {
		data := map[string]any{"name": "test"}
		handler := func() (map[string]any, error) { return data, nil }
		mgr, _ := newBaseConfigManager("Test", handler)

		_, err := mgr.Get("notexist")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("嵌套路径键不存在", func(t *testing.T) {
		data := map[string]any{"server": map[string]any{"port": 8080}}
		handler := func() (map[string]any, error) { return data, nil }
		mgr, _ := newBaseConfigManager("Test", handler)

		_, err := mgr.Get("server.host")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestBaseConfigManager_Get_Array(t *testing.T) {
	t.Run("数组元素访问", func(t *testing.T) {
		data := map[string]any{
			"servers": []any{"s1", "s2", "s3"},
		}
		handler := func() (map[string]any, error) { return data, nil }
		mgr, _ := newBaseConfigManager("Test", handler)

		s1, err := mgr.Get("servers[0]")
		assert.NoError(t, err)
		assert.Equal(t, "s1", s1)

		s2, err := mgr.Get("servers[1]")
		assert.NoError(t, err)
		assert.Equal(t, "s2", s2)
	})

	t.Run("嵌套对象数组访问", func(t *testing.T) {
		data := map[string]any{
			"servers": []any{
				map[string]any{"host": "s1", "port": 8080},
				map[string]any{"host": "s2", "port": 8081},
			},
		}
		handler := func() (map[string]any, error) { return data, nil }
		mgr, _ := newBaseConfigManager("Test", handler)

		host, err := mgr.Get("servers[0].host")
		assert.NoError(t, err)
		assert.Equal(t, "s1", host)

		port, err := mgr.Get("servers[1].port")
		assert.NoError(t, err)
		assert.Equal(t, 8081, port)
	})

	t.Run("数组索引越界", func(t *testing.T) {
		data := map[string]any{
			"servers": []any{"s1", "s2"},
		}
		handler := func() (map[string]any, error) { return data, nil }
		mgr, _ := newBaseConfigManager("Test", handler)

		_, err := mgr.Get("servers[5]")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "out of bounds")
	})

	t.Run("数组索引负数", func(t *testing.T) {
		data := map[string]any{
			"servers": []any{"s1", "s2"},
		}
		handler := func() (map[string]any, error) { return data, nil }
		mgr, _ := newBaseConfigManager("Test", handler)

		_, err := mgr.Get("servers[-1]")
		assert.Error(t, err)
		assert.Error(t, err)
	})

	t.Run("负数数组索引越界", func(t *testing.T) {
		data := map[string]any{
			"items": []any{1, 2, 3},
		}
		handler := func() (map[string]any, error) { return data, nil }
		mgr, _ := newBaseConfigManager("Test", handler)

		_, err := mgr.Get("items[5]")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "out of bounds")
	})

	t.Run("对非数组使用索引", func(t *testing.T) {
		data := map[string]any{
			"name": "test",
		}
		handler := func() (map[string]any, error) { return data, nil }
		mgr, _ := newBaseConfigManager("Test", handler)

		_, err := mgr.Get("name[0]")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not an array")
	})

	t.Run("中间路径包含数组", func(t *testing.T) {
		data := map[string]any{
			"items": []any{
				map[string]any{"name": "item1", "value": 100},
				map[string]any{"name": "item2", "value": 200},
			},
		}
		handler := func() (map[string]any, error) { return data, nil }
		mgr, _ := newBaseConfigManager("Test", handler)

		name, err := mgr.Get("items[1].name")
		assert.NoError(t, err)
		assert.Equal(t, "item2", name)

		value, err := mgr.Get("items[0].value")
		assert.NoError(t, err)
		assert.Equal(t, 100, value)
	})
}

func TestBaseConfigManager_ParsePath(t *testing.T) {
	t.Run("简单路径", func(t *testing.T) {
		handler := func() (map[string]any, error) { return map[string]any{}, nil }
		mgr, _ := newBaseConfigManager("Test", handler)

		parts, err := mgr.(*baseConfigManager).parsePath("server.host")
		assert.NoError(t, err)
		assert.Len(t, parts, 2)
		assert.Equal(t, "server", parts[0].key)
		assert.Equal(t, "host", parts[1].key)
	})

	t.Run("带数组索引的路径", func(t *testing.T) {
		handler := func() (map[string]any, error) { return map[string]any{}, nil }
		mgr, _ := newBaseConfigManager("Test", handler)

		parts, err := mgr.(*baseConfigManager).parsePath("servers[0].port")
		assert.NoError(t, err)
		assert.Len(t, parts, 2)
		assert.Equal(t, "servers", parts[0].key)
		assert.True(t, parts[0].hasIndex)
		assert.Equal(t, 0, parts[0].index)
		assert.Equal(t, "port", parts[1].key)
		assert.False(t, parts[1].hasIndex)
	})

	t.Run("多个数组索引", func(t *testing.T) {
		handler := func() (map[string]any, error) { return map[string]any{}, nil }
		mgr, _ := newBaseConfigManager("Test", handler)

		parts, err := mgr.(*baseConfigManager).parsePath("items[0].subitems[1]")
		assert.NoError(t, err)
		assert.Len(t, parts, 2)
		assert.True(t, parts[0].hasIndex)
		assert.Equal(t, 0, parts[0].index)
		assert.True(t, parts[1].hasIndex)
		assert.Equal(t, 1, parts[1].index)
	})

	t.Run("无效路径语法", func(t *testing.T) {
		handler := func() (map[string]any, error) { return map[string]any{}, nil }
		mgr, _ := newBaseConfigManager("Test", handler)

		_, err := mgr.(*baseConfigManager).parsePath("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid path syntax")
	})

}

func TestBaseConfigManager_Has(t *testing.T) {
	t.Run("存在的键", func(t *testing.T) {
		data := map[string]any{"name": "test", "port": 8080}
		handler := func() (map[string]any, error) { return data, nil }
		mgr, _ := newBaseConfigManager("Test", handler)

		assert.True(t, mgr.Has("name"))
		assert.True(t, mgr.Has("port"))
		assert.True(t, mgr.Has(""))
	})

	t.Run("不存在的键", func(t *testing.T) {
		data := map[string]any{"name": "test"}
		handler := func() (map[string]any, error) { return data, nil }
		mgr, _ := newBaseConfigManager("Test", handler)

		assert.False(t, mgr.Has("notexist"))
		assert.False(t, mgr.Has("server.host"))
	})

	t.Run("嵌套路径存在性", func(t *testing.T) {
		data := map[string]any{
			"server": map[string]any{"host": "localhost"},
		}
		handler := func() (map[string]any, error) { return data, nil }
		mgr, _ := newBaseConfigManager("Test", handler)

		assert.True(t, mgr.Has("server"))
		assert.True(t, mgr.Has("server.host"))
		assert.False(t, mgr.Has("server.port"))
	})

	t.Run("数组路径存在性", func(t *testing.T) {
		data := map[string]any{
			"items": []any{"a", "b", "c"},
		}
		handler := func() (map[string]any, error) { return data, nil }
		mgr, _ := newBaseConfigManager("Test", handler)

		assert.True(t, mgr.Has("items[0]"))
		assert.True(t, mgr.Has("items[2]"))
		assert.False(t, mgr.Has("items[5]"))
	})
}
