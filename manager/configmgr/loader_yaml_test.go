package configmgr

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadYAML(t *testing.T) {
	t.Run("成功加载 YAML 文件", func(t *testing.T) {
		tempDir := t.TempDir()
		yamlPath := filepath.Join(tempDir, "config.yaml")
		content := `name: test
port: 8080
database:
  host: localhost
  port: 3306
servers:
  - s1
  - s2
  - s3
`
		err := os.WriteFile(yamlPath, []byte(content), 0600)
		require.NoError(t, err)

		data, err := LoadYAML(yamlPath)
		assert.NoError(t, err)
		assert.NotNil(t, data)

		assert.Equal(t, "test", data["name"])
		assert.Equal(t, 8080, data["port"])
	})

	t.Run("文件不存在", func(t *testing.T) {
		yamlPath := "/nonexistent/path/config.yaml"
		data, err := LoadYAML(yamlPath)
		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "failed to read yaml file")
	})

	t.Run("YAML 格式错误", func(t *testing.T) {
		tempDir := t.TempDir()
		yamlPath := filepath.Join(tempDir, "invalid.yaml")
		err := os.WriteFile(yamlPath, []byte("name: {{invalid}}\nport: 8080\n"), 0600)
		require.NoError(t, err)

		data, err := LoadYAML(yamlPath)
		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "failed to parse yaml")
	})

	t.Run("空 YAML 文件", func(t *testing.T) {
		tempDir := t.TempDir()
		yamlPath := filepath.Join(tempDir, "empty.yaml")
		err := os.WriteFile(yamlPath, []byte("{}"), 0600)
		require.NoError(t, err)

		data, err := LoadYAML(yamlPath)
		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.Empty(t, data)
	})

	t.Run("嵌套 YAML 结构", func(t *testing.T) {
		tempDir := t.TempDir()
		yamlPath := filepath.Join(tempDir, "nested.yaml")
		content := `app:
  server:
    host: localhost
    port: 8080
  database:
    mysql:
      host: 127.0.0.1
      port: 3306
      name: testdb
`
		err := os.WriteFile(yamlPath, []byte(content), 0600)
		require.NoError(t, err)

		data, err := LoadYAML(yamlPath)
		assert.NoError(t, err)
		assert.NotNil(t, data)

		app := data["app"].(map[string]any)
		server := app["server"].(map[string]any)
		assert.Equal(t, "localhost", server["host"])
		assert.Equal(t, 8080, server["port"])
	})

	t.Run("数组 YAML 结构", func(t *testing.T) {
		tempDir := t.TempDir()
		yamlPath := filepath.Join(tempDir, "array.yaml")
		content := `users:
  - name: alice
    age: 30
  - name: bob
    age: 25
`
		err := os.WriteFile(yamlPath, []byte(content), 0600)
		require.NoError(t, err)

		data, err := LoadYAML(yamlPath)
		assert.NoError(t, err)
		assert.NotNil(t, data)

		users := data["users"].([]any)
		assert.Len(t, users, 2)
	})

	t.Run("YAML 布尔值", func(t *testing.T) {
		tempDir := t.TempDir()
		yamlPath := filepath.Join(tempDir, "bool.yaml")
		content := `enabled: true
disabled: false
`
		err := os.WriteFile(yamlPath, []byte(content), 0600)
		require.NoError(t, err)

		data, err := LoadYAML(yamlPath)
		assert.NoError(t, err)
		assert.NotNil(t, data)

		assert.True(t, data["enabled"].(bool))
		assert.False(t, data["disabled"].(bool))
	})
}
