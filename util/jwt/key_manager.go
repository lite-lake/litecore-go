package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"sync"
)

// IKeyManager RSA密钥管理器接口
type IKeyManager interface {
	// LoadKeys 加载公钥私钥
	LoadKeys(privateKeyPath, publicKeyPath string) error
	// GetPrivateKey 获取私钥
	GetPrivateKey() *rsa.PrivateKey
	// GetPublicKey 获取公钥
	GetPublicKey() *rsa.PublicKey
	// GetPublicKeyJWK 获取公钥的JWK格式
	GetPublicKeyJWK() *JWK
	// GetKeyID 获取密钥ID
	GetKeyID() string
}

// JWK JSON Web Key 结构
type JWK struct {
	Kty string `json:"kty"` // 密钥类型
	Use string `json:"use"` // 公钥用途
	Alg string `json:"alg"` // 算法
	Kid string `json:"kid"` // 密钥ID
	N   string `json:"n"`   // RSA公钥模数
	E   string `json:"e"`   // RSA公钥指数
}

// JWKS JSON Web Key Set 结构
type JWKS struct {
	Keys []*JWK `json:"keys"`
}

// keyManager RSA密钥管理器实现
type keyManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	keyID      string
	mu         sync.RWMutex
}

var defaultKeyManager = &keyManager{
	keyID: "default-rsa-key-001",
}

// DefaultKeyManager 获取默认密钥管理器实例
func DefaultKeyManager() IKeyManager {
	return defaultKeyManager
}

// NewKeyManager 创建新的密钥管理器实例
func NewKeyManager(keyID string) IKeyManager {
	return &keyManager{
		keyID: keyID,
	}
}

// base64URLEncode URL安全的Base64编码
func base64URLEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

// LoadKeys 从文件加载公钥和私钥
func (k *keyManager) LoadKeys(privateKeyPath, publicKeyPath string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// 加载私钥
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("读取私钥文件失败: %w", err)
	}

	privateBlock, _ := pem.Decode(privateKeyBytes)
	if privateBlock == nil {
		return errors.New("无效的私钥PEM格式")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(privateBlock.Bytes)
	if err != nil {
		// 尝试PKCS1格式
		privateKey, err = x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
		if err != nil {
			return fmt.Errorf("解析私钥失败: %w", err)
		}
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return errors.New("私钥不是RSA类型")
	}

	// 加载公钥
	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("读取公钥文件失败: %w", err)
	}

	publicBlock, _ := pem.Decode(publicKeyBytes)
	if publicBlock == nil {
		return errors.New("无效的公钥PEM格式")
	}

	publicKey, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	if err != nil {
		// 尝试PKCS1格式
		publicKey, err = x509.ParsePKCS1PublicKey(publicBlock.Bytes)
		if err != nil {
			return fmt.Errorf("解析公钥失败: %w", err)
		}
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return errors.New("公钥不是RSA类型")
	}

	k.privateKey = rsaPrivateKey
	k.publicKey = rsaPublicKey

	return nil
}

// GetPrivateKey 获取私钥
func (k *keyManager) GetPrivateKey() *rsa.PrivateKey {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.privateKey
}

// GetPublicKey 获取公钥
func (k *keyManager) GetPublicKey() *rsa.PublicKey {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.publicKey
}

// GetKeyID 获取密钥ID
func (k *keyManager) GetKeyID() string {
	return k.keyID
}

// GetPublicKeyJWK 获取公钥的JWK格式
func (k *keyManager) GetPublicKeyJWK() *JWK {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.publicKey == nil {
		return nil
	}

	// 编码模数和指数为Base64URL
	n := base64URLEncode(k.publicKey.N.Bytes())
	e := base64URLEncode(big.NewInt(int64(k.publicKey.E)).Bytes())

	return &JWK{
		Kty: "RSA",
		Use: "sig",
		Alg: "RS256",
		Kid: k.keyID,
		N:   n,
		E:   e,
	}
}

// GenerateRS256Token 使用默认密钥管理器的私钥生成RS256令牌
func GenerateRS256Token(claims ILiteUtilJWTClaims) (string, error) {
	privateKey := defaultKeyManager.GetPrivateKey()
	if privateKey == nil {
		return "", errors.New("RSA私钥未加载")
	}
	return defaultJWT.GenerateRS256Token(claims, privateKey)
}

// ParseRS256Token 使用默认密钥管理器的公钥解析RS256令牌
func ParseRS256Token(token string) (MapClaims, error) {
	publicKey := defaultKeyManager.GetPublicKey()
	if publicKey == nil {
		return nil, errors.New("RSA公钥未加载")
	}
	return defaultJWT.ParseRS256Token(token, publicKey)
}
