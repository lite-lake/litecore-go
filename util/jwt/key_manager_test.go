package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 生成测试用的RSA密钥对文件
func generateTestRSAKeyFiles(t *testing.T) (string, string, func()) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	// 创建私钥文件
	privateKeyFile, err := os.CreateTemp("", "test-rsa-*.key")
	assert.NoError(t, err)
	defer privateKeyFile.Close()

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privatePEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	err = pem.Encode(privateKeyFile, privatePEM)
	assert.NoError(t, err)

	// 创建公钥文件
	publicKeyFile, err := os.CreateTemp("", "test-rsa-*.pub")
	assert.NoError(t, err)
	defer publicKeyFile.Close()

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	assert.NoError(t, err)
	publicPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	err = pem.Encode(publicKeyFile, publicPEM)
	assert.NoError(t, err)

	cleanup := func() {
		os.Remove(privateKeyFile.Name())
		os.Remove(publicKeyFile.Name())
	}

	return privateKeyFile.Name(), publicKeyFile.Name(), cleanup
}

func TestKeyManager_LoadKeys(t *testing.T) {
	privateKeyPath, publicKeyPath, cleanup := generateTestRSAKeyFiles(t)
	defer cleanup()

	km := NewKeyManager("test-key-001")
	err := km.LoadKeys(privateKeyPath, publicKeyPath)
	assert.NoError(t, err)

	assert.NotNil(t, km.GetPrivateKey())
	assert.NotNil(t, km.GetPublicKey())
	assert.Equal(t, "test-key-001", km.GetKeyID())
}

func TestKeyManager_GetPublicKeyJWK(t *testing.T) {
	privateKeyPath, publicKeyPath, cleanup := generateTestRSAKeyFiles(t)
	defer cleanup()

	km := NewKeyManager("test-key-001")
	err := km.LoadKeys(privateKeyPath, publicKeyPath)
	assert.NoError(t, err)

	jwk := km.GetPublicKeyJWK()
	assert.NotNil(t, jwk)
	assert.Equal(t, "RSA", jwk.Kty)
	assert.Equal(t, "sig", jwk.Use)
	assert.Equal(t, "RS256", jwk.Alg)
	assert.Equal(t, "test-key-001", jwk.Kid)
	assert.NotEmpty(t, jwk.N)
	assert.NotEmpty(t, jwk.E)
}

func TestDefaultKeyManager_GenerateAndParseRS256Token(t *testing.T) {
	privateKeyPath, publicKeyPath, cleanup := generateTestRSAKeyFiles(t)
	defer cleanup()

	err := DefaultKeyManager().LoadKeys(privateKeyPath, publicKeyPath)
	assert.NoError(t, err)

	claims := &StandardClaims{
		Issuer:    "test-issuer",
		Subject:   "test-subject",
		ExpiresAt: 0, // 不过期
	}

	token, err := GenerateRS256Token(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	parsedClaims, err := ParseRS256Token(token)
	assert.NoError(t, err)
	assert.Equal(t, "test-issuer", parsedClaims.GetIssuer())
	assert.Equal(t, "test-subject", parsedClaims.GetSubject())
}
