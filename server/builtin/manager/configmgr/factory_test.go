package configmgr

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuild(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("创建 YAML 配置管理器", func(t *testing.T) {
		yamlPath := filepath.Join(tempDir, "config.yaml")
		err := os.WriteFile(yamlPath, []byte("name: test\nport: 8080\n"), 0600)
		require.NoError(t, err)

		mgr, err := Build("yaml", yamlPath)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.Equal(t, "ConfigYamlManager", mgr.ManagerName())

		name, err := mgr.Get("name")
		assert.NoError(t, err)
		assert.Equal(t, "test", name)
	})

	t.Run("创建 JSON 配置管理器", func(t *testing.T) {
		jsonPath := filepath.Join(tempDir, "config.json")
		err := os.WriteFile(jsonPath, []byte(`{"name": "test", "port": 8080}`), 0600)
		require.NoError(t, err)

		mgr, err := Build("json", jsonPath)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.Equal(t, "ConfigJsonManager", mgr.ManagerName())

		name, err := mgr.Get("name")
		assert.NoError(t, err)
		assert.Equal(t, "test", name)
	})

	t.Run("不支持的驱动类型", func(t *testing.T) {
		mgr, err := Build("invalid", "/some/path")
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "unsupported config driver")
	})

	t.Run("YAML 文件不存在", func(t *testing.T) {
		yamlPath := filepath.Join(tempDir, "notexist.yaml")
		mgr, err := Build("yaml", yamlPath)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "failed to read yaml file")
	})

	t.Run("JSON 文件不存在", func(t *testing.T) {
		jsonPath := filepath.Join(tempDir, "notexist.json")
		mgr, err := Build("json", jsonPath)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "failed to read json file")
	})

	t.Run("YAML 格式错误", func(t *testing.T) {
		yamlPath := filepath.Join(tempDir, "invalid.yaml")
		err := os.WriteFile(yamlPath, []byte("name: {{invalid}}\nport: 8080\n"), 0600)
		require.NoError(t, err)

		mgr, err := Build("yaml", yamlPath)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "failed to parse yaml")
	})

	t.Run("JSON 格式错误", func(t *testing.T) {
		jsonPath := filepath.Join(tempDir, "invalid.json")
		err := os.WriteFile(jsonPath, []byte(`{name: test}`), 0600)
		require.NoError(t, err)

		mgr, err := Build("json", jsonPath)
		assert.Error(t, err)
		assert.Nil(t, mgr)
		assert.Contains(t, err.Error(), "failed to parse json")
	})
}
