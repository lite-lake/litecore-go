package liteadmin

import (
	"crypto/rand"
	"encoding/base64"
)

// KeyBytes 委托密钥的随机字节数（32 字节，PRD §10 安全设计）
const KeyBytes = 32

// GenerateKey 生成一把 Admin 委托密钥：32 字节密码学安全随机数，base64url 无填充编码。
// 生成后由运维经 userdata secrets 注入服务配置（admin_delegation.api_keys），
// 真实密钥不进 git、不进数据库（PRD §10）。
func GenerateKey() (string, error) {
	buf := make([]byte, KeyBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
