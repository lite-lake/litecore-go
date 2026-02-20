package upgrade

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCommand(t *testing.T) {
	t.Run("创建升级命令", func(t *testing.T) {
		cmd := GetCommand()

		assert.Equal(t, "upgrade", cmd.Name)
		assert.Equal(t, "升级 CLI 到最新版本", cmd.Usage)
		assert.NotEmpty(t, cmd.Description)
		assert.NotNil(t, cmd.Action)

		assert.Len(t, cmd.Flags, 2)

		flagNames := make(map[string]bool)
		for _, flag := range cmd.Flags {
			flagNames[flag.Names()[0]] = true
		}

		assert.True(t, flagNames["force"])
		assert.True(t, flagNames["check"])
	})
}

func TestGetLatestRelease(t *testing.T) {
	t.Run("获取最新版本", func(t *testing.T) {
		release, err := getLatestRelease()
		if err != nil {
			t.Skipf("跳过测试: 无法连接 GitHub API: %v", err)
		}

		assert.NotEmpty(t, release.TagName)
		assert.NotEmpty(t, release.HTMLURL)
		assert.Contains(t, release.TagName, "v")
	})
}

func TestIsWindowsFileLocked(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{"空错误", "", false},
		{"普通错误", "some other error", false},
		{"Access denied", "Access is denied", true},
		{"小写 access denied", "access is denied", true},
		{"Used by another process", "The process cannot access the file because it is being used by another process", true},
		{"Being used", "file is being used", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &testError{msg: tt.errMsg}
			result := isWindowsFileLocked(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetBinaryName(t *testing.T) {
	name := getBinaryName()
	assert.NotEmpty(t, name)
	assert.Contains(t, name, "litecore-cli")
}

func TestGetTargetBinaryPath(t *testing.T) {
	path := getTargetBinaryPath()
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "litecore-cli")
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
