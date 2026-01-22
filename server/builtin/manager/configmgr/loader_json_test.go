package configmgr

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadJSON(t *testing.T) {
	t.Run("成功加载 JSON 文件", func(t *testing.T) {
		tempDir := t.TempDir()
		jsonPath := filepath.Join(tempDir, "config.json")
		content := `{
			"name": "test",
			"port": 8080,
			"database": {
				"host": "localhost",
				"port": 3306
			},
			"servers": ["s1", "s2", "s3"]
		}`
		err := os.WriteFile(jsonPath, []byte(content), 0600)
		require.NoError(t, err)

		data, err := LoadJSON(jsonPath)
		assert.NoError(t, err)
		assert.NotNil(t, data)

		assert.Equal(t, "test", data["name"])
		assert.Equal(t, 8080.0, data["port"])
	})

	t.Run("文件不存在", func(t *testing.T) {
		jsonPath := "/nonexistent/path/config.json"
		data, err := LoadJSON(jsonPath)
		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "failed to read json file")
	})

	t.Run("JSON 格式错误", func(t *testing.T) {
		tempDir := t.TempDir()
		jsonPath := filepath.Join(tempDir, "invalid.json")
		err := os.WriteFile(jsonPath, []byte(`{name: "test"}`), 0600)
		require.NoError(t, err)

		data, err := LoadJSON(jsonPath)
		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "failed to parse json")
	})

	t.Run("空 JSON 文件", func(t *testing.T) {
		tempDir := t.TempDir()
		jsonPath := filepath.Join(tempDir, "empty.json")
		err := os.WriteFile(jsonPath, []byte(`{}`), 0600)
		require.NoError(t, err)

		data, err := LoadJSON(jsonPath)
		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.Empty(t, data)
	})

	t.Run("嵌套 JSON 结构", func(t *testing.T) {
		tempDir := t.TempDir()
		jsonPath := filepath.Join(tempDir, "nested.json")
		content := `{
			"app": {
				"server": {
					"host": "localhost",
					"port": 8080
				},
				"database": {
					"mysql": {
						"host": "127.0.0.1",
						"port": 3306,
						"name": "testdb"
					}
				}
			}
		}`
		err := os.WriteFile(jsonPath, []byte(content), 0600)
		require.NoError(t, err)

		data, err := LoadJSON(jsonPath)
		assert.NoError(t, err)
		assert.NotNil(t, data)

		app := data["app"].(map[string]any)
		server := app["server"].(map[string]any)
		assert.Equal(t, "localhost", server["host"])
		assert.Equal(t, 8080.0, server["port"])
	})

	t.Run("数组 JSON 结构", func(t *testing.T) {
		tempDir := t.TempDir()
		jsonPath := filepath.Join(tempDir, "array.json")
		content := `{
			"users": [
				{"name": "alice", "age": 30},
				{"name": "bob", "age": 25}
			]
		}`
		err := os.WriteFile(jsonPath, []byte(content), 0600)
		require.NoError(t, err)

		data, err := LoadJSON(jsonPath)
		assert.NoError(t, err)
		assert.NotNil(t, data)

		users := data["users"].([]any)
		assert.Len(t, users, 2)
	})
}
