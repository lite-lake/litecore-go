package liteadmin

import (
	"encoding/base64"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateKey(t *testing.T) {
	t.Run("长度与字符集符合 base64url 无填充", func(t *testing.T) {
		key, err := GenerateKey()
		require.NoError(t, err)
		// 32 字节 → base64url 无填充 43 字符
		assert.Len(t, key, 43)
		assert.Regexp(t, regexp.MustCompile(`^[A-Za-z0-9_-]+$`), key)
		_, err = base64.RawURLEncoding.DecodeString(key)
		assert.NoError(t, err, "应为合法 base64url 无填充编码")
	})

	t.Run("多次生成互不相同", func(t *testing.T) {
		seen := make(map[string]struct{})
		for i := 0; i < 100; i++ {
			key, err := GenerateKey()
			require.NoError(t, err)
			_, dup := seen[key]
			assert.False(t, dup, "随机密钥不应重复")
			seen[key] = struct{}{}
		}
	})

	t.Run("解码后恰为 32 字节", func(t *testing.T) {
		key, err := GenerateKey()
		require.NoError(t, err)
		raw, err := base64.RawURLEncoding.DecodeString(key)
		require.NoError(t, err)
		assert.Len(t, raw, KeyBytes)
		assert.Len(t, raw, 32)
	})
}
