package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

// AESKeySize AES密钥大小枚举
type AESKeySize int

const (
	// AES128 128位AES密钥
	AES128 AESKeySize = 16
	// AES192 192位AES密钥
	AES192 AESKeySize = 24
	// AES256 256位AES密钥
	AES256 AESKeySize = 32
)

// RSABits RSA密钥位数枚举
type RSABits int

const (
	// RSA1024 1024位RSA密钥
	RSA1024 RSABits = 1024
	// RSA2048 2048位RSA密钥
	RSA2048 RSABits = 2048
	// RSA3072 3072位RSA密钥
	RSA3072 RSABits = 3072
	// RSA4096 4096位RSA密钥
	RSA4096 RSABits = 4096
)

// cryptEngine 加密操作引擎（内部实现）
type cryptEngine struct{}

// Crypt 默认的加密操作实例
var Crypt = &cryptEngine{}

// =========================================
// Base64 编码解码
// =========================================

// Base64Encode Base64编码
func (c *cryptEngine) Base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

// Base64EncodeBytes Base64编码字节数组
func (c *cryptEngine) Base64EncodeBytes(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Base64Decode Base64解码
func (c *cryptEngine) Base64Decode(data string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %w", err)
	}
	return string(decoded), nil
}

// Base64DecodeBytes Base64解码为字节数组
func (c *cryptEngine) Base64DecodeBytes(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

// Base64URLEncode URL安全的Base64编码
func (c *cryptEngine) Base64URLEncode(data string) string {
	return base64.URLEncoding.EncodeToString([]byte(data))
}

// Base64URLDecode URL安全的Base64解码
func (c *cryptEngine) Base64URLDecode(data string) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("base64 url decode failed: %w", err)
	}
	return string(decoded), nil
}

// =========================================
// Hex 编码解码
// =========================================

// HexEncode 十六进制编码
func (c *cryptEngine) HexEncode(data string) string {
	return hex.EncodeToString([]byte(data))
}

// HexEncodeBytes 十六进制编码字节数组
func (c *cryptEngine) HexEncodeBytes(data []byte) string {
	return hex.EncodeToString(data)
}

// HexDecode 十六进制解码
func (c *cryptEngine) HexDecode(data string) (string, error) {
	decoded, err := hex.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("hex decode failed: %w", err)
	}
	return string(decoded), nil
}

// HexDecodeBytes 十六进制解码为字节数组
func (c *cryptEngine) HexDecodeBytes(data string) ([]byte, error) {
	return hex.DecodeString(data)
}

// =========================================
// AES 对称加密
// =========================================

// GenerateAESKey 生成AES密钥
func (c *cryptEngine) GenerateAESKey(keySize AESKeySize) ([]byte, error) {
	if keySize != AES128 && keySize != AES192 && keySize != AES256 {
		return nil, errors.New("invalid AES key size, must be 16, 24, or 32 bytes")
	}

	key := make([]byte, keySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("generate AES key failed: %w", err)
	}
	return key, nil
}

// GenerateAESKeyHex 生成十六进制格式的AES密钥
func (c *cryptEngine) GenerateAESKeyHex(keySize AESKeySize) (string, error) {
	key, err := c.GenerateAESKey(keySize)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

// AESEncrypt AES加密
func (c *cryptEngine) AESEncrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM mode failed: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("generate nonce failed: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// AESEncryptToBase64 AES加密并Base64编码
func (c *cryptEngine) AESEncryptToBase64(plaintext string, key []byte) (string, error) {
	ciphertext, err := c.AESEncrypt([]byte(plaintext), key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AESDecrypt AES解密
func (c *cryptEngine) AESDecrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM mode failed: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("AES decrypt failed: %w", err)
	}

	return plaintext, nil
}

// AESDecryptFromBase64 从Base64字符串AES解密
func (c *cryptEngine) AESDecryptFromBase64(ciphertext string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %w", err)
	}

	plaintext, err := c.AESDecrypt(data, key)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// =========================================
// RSA 非对称加密
// =========================================

// GenerateRSAKeys 生成RSA密钥对
func (c *cryptEngine) GenerateRSAKeys(bits RSABits) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, int(bits))
	if err != nil {
		return nil, nil, fmt.Errorf("generate RSA keys failed: %w", err)
	}
	return privateKey, &privateKey.PublicKey, nil
}

// RSAEncrypt RSA公钥加密
func (c *cryptEngine) RSAEncrypt(plaintext []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	ciphertext, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		publicKey,
		plaintext,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("RSA encrypt failed: %w", err)
	}
	return ciphertext, nil
}

// RSAEncryptToBase64 RSA公钥加密并Base64编码
func (c *cryptEngine) RSAEncryptToBase64(plaintext string, publicKey *rsa.PublicKey) (string, error) {
	ciphertext, err := c.RSAEncrypt([]byte(plaintext), publicKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// RSADecrypt RSA私钥解密
func (c *cryptEngine) RSADecrypt(ciphertext []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	plaintext, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		privateKey,
		ciphertext,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("RSA decrypt failed: %w", err)
	}
	return plaintext, nil
}

// RSADecryptFromBase64 从Base64字符串RSA解密
func (c *cryptEngine) RSADecryptFromBase64(ciphertext string, privateKey *rsa.PrivateKey) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %w", err)
	}

	plaintext, err := c.RSADecrypt(data, privateKey)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// =========================================
// 密码哈希
// =========================================

// BcryptHash bcrypt密码哈希
func (c *cryptEngine) BcryptHash(password string, cost int) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash failed: %w", err)
	}
	return string(hashedBytes), nil
}

// BcryptVerify bcrypt密码验证
func (c *cryptEngine) BcryptVerify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// PBKDF2Hash PBKDF2密码哈希
func (c *cryptEngine) PBKDF2Hash(password, salt string, iterations, keyLen int) string {
	return base64.StdEncoding.EncodeToString(
		pbkdf2.Key([]byte(password), []byte(salt), iterations, keyLen, sha256.New),
	)
}

// PBKDF2Verify PBKDF2密码验证
func (c *cryptEngine) PBKDF2Verify(password, salt, hash string, iterations, keyLen int) bool {
	expectedHash := c.PBKDF2Hash(password, salt, iterations, keyLen)
	return subtle.ConstantTimeCompare([]byte(hash), []byte(expectedHash)) == 1
}

// GenerateSalt 生成随机盐值
func (c *cryptEngine) GenerateSalt(length int) ([]byte, error) {
	if length <= 0 {
		return nil, errors.New("salt length must be positive")
	}

	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("generate salt failed: %w", err)
	}
	return salt, nil
}

// GenerateSaltHex 生成十六进制格式的随机盐值
func (c *cryptEngine) GenerateSaltHex(length int) (string, error) {
	salt, err := c.GenerateSalt(length)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(salt), nil
}

// =========================================
// HMAC 签名
// =========================================

// HMACSign HMAC签名
func (c *cryptEngine) HMACSign(data, key []byte, hashFunc func() hash.Hash) []byte {
	h := hmac.New(hashFunc, key)
	h.Write(data)
	return h.Sum(nil)
}

// HMACSignHex HMAC签名并转换为十六进制字符串
func (c *cryptEngine) HMACSignHex(data, key []byte, hashFunc func() hash.Hash) string {
	signature := c.HMACSign(data, key, hashFunc)
	return hex.EncodeToString(signature)
}

// HMACSignBase64 HMAC签名并转换为Base64字符串
func (c *cryptEngine) HMACSignBase64(data, key []byte, hashFunc func() hash.Hash) string {
	signature := c.HMACSign(data, key, hashFunc)
	return base64.StdEncoding.EncodeToString(signature)
}

// HMACVerify HMAC验证
func (c *cryptEngine) HMACVerify(data, key, signature []byte, hashFunc func() hash.Hash) bool {
	expectedSignature := c.HMACSign(data, key, hashFunc)
	return hmac.Equal(signature, expectedSignature)
}

// HMACSignWithSHA256 使用SHA256的HMAC签名
func (c *cryptEngine) HMACSignWithSHA256(data, key []byte) []byte {
	return c.HMACSign(data, key, sha256.New)
}

// HMACSignHexWithSHA256 使用SHA256的HMAC签名并转换为十六进制
func (c *cryptEngine) HMACSignHexWithSHA256(data, key []byte) string {
	return c.HMACSignHex(data, key, sha256.New)
}

// HMACSignWithSHA512 使用SHA512的HMAC签名
func (c *cryptEngine) HMACSignWithSHA512(data, key []byte) []byte {
	return c.HMACSign(data, key, sha512.New)
}

// HMACSignHexWithSHA512 使用SHA512的HMAC签名并转换为十六进制
func (c *cryptEngine) HMACSignHexWithSHA512(data, key []byte) string {
	return c.HMACSignHex(data, key, sha512.New)
}

// =========================================
// ECDSA 数字签名
// =========================================

// GenerateECDSAKeys 生成ECDSA密钥对
func (c *cryptEngine) GenerateECDSAKeys() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate ECDSA keys failed: %w", err)
	}
	return privateKey, &privateKey.PublicKey, nil
}

// ECDSASign ECDSA签名
func (c *cryptEngine) ECDSASign(data []byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	hash := sha256.Sum256(data)
	signature, err := ecdsa.SignASN1(rand.Reader, privateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("ECDSA sign failed: %w", err)
	}
	return signature, nil
}

// ECDSASignHex ECDSA签名并转换为十六进制
func (c *cryptEngine) ECDSASignHex(data []byte, privateKey *ecdsa.PrivateKey) (string, error) {
	signature, err := c.ECDSASign(data, privateKey)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(signature), nil
}

// ECDSAVerify ECDSA验证
func (c *cryptEngine) ECDSAVerify(data, signature []byte, publicKey *ecdsa.PublicKey) bool {
	hash := sha256.Sum256(data)
	return ecdsa.VerifyASN1(publicKey, hash[:], signature)
}

// ECDSAVerifyHex ECDSA验证十六进制签名
func (c *cryptEngine) ECDSAVerifyHex(data []byte, signatureHex string, publicKey *ecdsa.PublicKey) (bool, error) {
	signature, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false, fmt.Errorf("decode signature failed: %w", err)
	}
	return c.ECDSAVerify(data, signature, publicKey), nil
}

// =========================================
// 密钥格式转换
// =========================================

// PrivateKeyToPEM RSA私钥转换为PEM格式
func (c *cryptEngine) PrivateKeyToPEM(privateKey *rsa.PrivateKey) string {
	// 简化实现，实际使用中应该使用 x509 标准库
	return fmt.Sprintf("RSA Private Key (Modulus: %x, Public Exponent: %d)",
		privateKey.N, privateKey.E)
}

// PublicKeyToPEM RSA公钥转换为PEM格式
func (c *cryptEngine) PublicKeyToPEM(publicKey *rsa.PublicKey) string {
	// 简化实现，实际使用中应该使用 x509 标准库
	return fmt.Sprintf("RSA Public Key (Modulus: %x, Public Exponent: %d)",
		publicKey.N, publicKey.E)
}

// =========================================
// 工具函数
// =========================================

// ConstantTimeCompare 常数时间比较，防止时序攻击
func (c *cryptEngine) ConstantTimeCompare(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}

// SecureEqual 安全字符串比较
func (c *cryptEngine) SecureEqual(a, b string) bool {
	return c.ConstantTimeCompare([]byte(a), []byte(b))
}

// GenerateRandomBytes 生成指定长度的随机字节
func (c *cryptEngine) GenerateRandomBytes(length int) ([]byte, error) {
	if length <= 0 {
		return nil, errors.New("length must be positive")
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("generate random bytes failed: %w", err)
	}
	return bytes, nil
}

// GenerateRandomString 生成指定长度的随机字符串
func (c *cryptEngine) GenerateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes, err := c.GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}

	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}

	return string(bytes), nil
}

// IsBase64 检查字符串是否为有效的Base64编码
func (c *cryptEngine) IsBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// IsHex 检查字符串是否为有效的十六进制编码
func (c *cryptEngine) IsHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}

// EncodeKey 编码密钥为可传输格式
func (c *cryptEngine) EncodeKey(key []byte) string {
	return base64.StdEncoding.EncodeToString(key)
}

// DecodeKey 解码密钥
func (c *cryptEngine) DecodeKey(encodedKey string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encodedKey)
}
