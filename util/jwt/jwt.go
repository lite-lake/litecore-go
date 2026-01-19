package jwt

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// JWTAlgorithm JWT签名算法类型
type JWTAlgorithm string

const (
	// HS256 HMAC使用SHA-256
	HS256 JWTAlgorithm = "HS256"
	// HS384 HMAC使用SHA-384
	HS384 JWTAlgorithm = "HS384"
	// HS512 HMAC使用SHA-512
	HS512 JWTAlgorithm = "HS512"
	// RS256 RSASSA-PKCS1-v1_5使用SHA-256
	RS256 JWTAlgorithm = "RS256"
	// RS384 RSASSA-PKCS1-v1_5使用SHA-384
	RS384 JWTAlgorithm = "RS384"
	// RS512 RSASSA-PKCS1-v1_5使用SHA-512
	RS512 JWTAlgorithm = "RS512"
	// ES256 ECDSA使用P-256和SHA-256
	ES256 JWTAlgorithm = "ES256"
	// ES384 ECDSA使用P-384和SHA-384
	ES384 JWTAlgorithm = "ES384"
	// ES512 ECDSA使用P-521和SHA-512
	ES512 JWTAlgorithm = "ES512"
)

var (
	// claimsMapPool 重用标准Claims的map对象，减少内存分配
	claimsMapPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]interface{}, 7)
		},
	}
)

// ILiteUtilJWTClaims JWT声明接口
type ILiteUtilJWTClaims interface {
	// GetExpiresAt 获取过期时间
	GetExpiresAt() *time.Time
	// GetIssuedAt 获取签发时间
	GetIssuedAt() *time.Time
	// GetNotBefore 获取生效时间
	GetNotBefore() *time.Time
	// GetIssuer 获取签发者
	GetIssuer() string
	// GetSubject 获取主题
	GetSubject() string
	// GetAudience 获取受众
	GetAudience() []string
	// GetCustomClaims 获取自定义声明
	GetCustomClaims() map[string]interface{}
	// SetCustomClaims 设置自定义声明
	SetCustomClaims(claims map[string]interface{})
}

// StandardClaims 标准JWT声明
type StandardClaims struct {
	Audience  []string `json:"aud,omitempty"`
	ExpiresAt int64    `json:"exp,omitempty"`
	ID        string   `json:"jti,omitempty"`
	IssuedAt  int64    `json:"iat,omitempty"`
	Issuer    string   `json:"iss,omitempty"`
	NotBefore int64    `json:"nbf,omitempty"`
	Subject   string   `json:"sub,omitempty"`
}

// MapClaims 映射形式的JWT声明，支持自定义字段
type MapClaims map[string]interface{}

// ILiteUtilJWT JWT 工具接口
type ILiteUtilJWT interface {
	// JWT 生成方法
	GenerateHS256Token(claims ILiteUtilJWTClaims, secretKey []byte) (string, error)
	GenerateHS512Token(claims ILiteUtilJWTClaims, secretKey []byte) (string, error)
	GenerateRS256Token(claims ILiteUtilJWTClaims, privateKey *rsa.PrivateKey) (string, error)
	GenerateES256Token(claims ILiteUtilJWTClaims, privateKey *ecdsa.PrivateKey) (string, error)
	GenerateToken(claims ILiteUtilJWTClaims, algorithm JWTAlgorithm, secretKey []byte,
		rsaPrivateKey *rsa.PrivateKey, ecdsaPrivateKey *ecdsa.PrivateKey) (string, error)

	// JWT 解析方法
	ParseHS256Token(token string, secretKey []byte) (MapClaims, error)
	ParseHS512Token(token string, secretKey []byte) (MapClaims, error)
	ParseRS256Token(token string, publicKey *rsa.PublicKey) (MapClaims, error)
	ParseES256Token(token string, publicKey *ecdsa.PublicKey) (MapClaims, error)
	ParseToken(token string, algorithm JWTAlgorithm, secretKey []byte,
		rsaPublicKey *rsa.PublicKey, ecdsaPublicKey *ecdsa.PublicKey) (MapClaims, error)

	// JWT 验证方法
	ValidateClaims(claims ILiteUtilJWTClaims, options ...ValidateOption) error

	// 便捷方法
	NewStandardClaims() *StandardClaims
	NewMapClaims() MapClaims
	SetExpiration(claims ILiteUtilJWTClaims, duration time.Duration)
	SetIssuedAt(claims ILiteUtilJWTClaims, t time.Time)
	SetNotBefore(claims ILiteUtilJWTClaims, t time.Time)
	SetIssuer(claims ILiteUtilJWTClaims, issuer string)
	SetSubject(claims ILiteUtilJWTClaims, subject string)
	SetAudience(claims ILiteUtilJWTClaims, audience ...string)
	AddCustomClaim(claims ILiteUtilJWTClaims, key string, value interface{})
}

// jwtEngine JWT操作工具类（私有结构体）
type jwtEngine struct{}

// 默认JWT操作实例
var defaultJWT = newJWTEngine()
var JWT = defaultJWT

// New 创建新的JWT操作实例
// Deprecated: 请使用 liteutil.LiteUtil().NewJwtOperation() 来创建新的 JWT 工具实例
func newJWTEngine() ILiteUtilJWT {
	return &jwtEngine{}
}

// Default 返回默认的JWT操作实例（单例模式）
// Deprecated: 请使用 liteutil.LiteUtil().JWT() 来获取 JWT 工具实例

// =========================================
// JWT Header 结构和工具方法
// =========================================

// jwtHeader JWT头部结构
type jwtHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
	KeyID     string `json:"kid,omitempty"`
}

// newJWTHeader 创建JWT头部
func newJWTHeader(alg JWTAlgorithm, keyID ...string) jwtHeader {
	header := jwtHeader{
		Algorithm: string(alg),
		Type:      "JWT",
	}

	if len(keyID) > 0 && keyID[0] != "" {
		header.KeyID = keyID[0]
	}

	return header
}

// encodeHeader 编码JWT头部
func (j *jwtEngine) encodeHeader(header jwtHeader) string {
	headerBytes, _ := json.Marshal(header)
	return j.base64URLEncode(headerBytes)
}

// =========================================
// Claims 实现方法
// =========================================

// GetExpiresAt 获取过期时间
func (c StandardClaims) GetExpiresAt() *time.Time {
	if c.ExpiresAt == 0 {
		return nil
	}
	t := time.Unix(c.ExpiresAt, 0)
	return &t
}

// GetIssuedAt 获取签发时间
func (c StandardClaims) GetIssuedAt() *time.Time {
	if c.IssuedAt == 0 {
		return nil
	}
	t := time.Unix(c.IssuedAt, 0)
	return &t
}

// GetNotBefore 获取生效时间
func (c StandardClaims) GetNotBefore() *time.Time {
	if c.NotBefore == 0 {
		return nil
	}
	t := time.Unix(c.NotBefore, 0)
	return &t
}

// GetIssuer 获取签发者
func (c StandardClaims) GetIssuer() string {
	return c.Issuer
}

// GetSubject 获取主题
func (c StandardClaims) GetSubject() string {
	return c.Subject
}

// GetAudience 获取受众
func (c StandardClaims) GetAudience() []string {
	return c.Audience
}

// GetCustomClaims 获取自定义声明（StandardClaims无自定义字段）
func (c StandardClaims) GetCustomClaims() map[string]interface{} {
	return make(map[string]interface{})
}

// SetCustomClaims 设置自定义声明（StandardClaims不支持自定义字段）
func (c *StandardClaims) SetCustomClaims(claims map[string]interface{}) {
	// StandardClaims不支持自定义字段，此方法为空实现
	// 保留方法以实现接口，但不执行任何操作
}

// GetExpiresAt 获取过期时间
func (c MapClaims) GetExpiresAt() *time.Time {
	if exp, ok := c["exp"].(float64); ok {
		t := time.Unix(int64(exp), 0)
		return &t
	}
	return nil
}

// GetIssuedAt 获取签发时间
func (c MapClaims) GetIssuedAt() *time.Time {
	if iat, ok := c["iat"].(float64); ok {
		t := time.Unix(int64(iat), 0)
		return &t
	}
	return nil
}

// GetNotBefore 获取生效时间
func (c MapClaims) GetNotBefore() *time.Time {
	if nbf, ok := c["nbf"].(float64); ok {
		t := time.Unix(int64(nbf), 0)
		return &t
	}
	return nil
}

// GetIssuer 获取签发者
func (c MapClaims) GetIssuer() string {
	if iss, ok := c["iss"].(string); ok {
		return iss
	}
	return ""
}

// GetSubject 获取主题
func (c MapClaims) GetSubject() string {
	if sub, ok := c["sub"].(string); ok {
		return sub
	}
	return ""
}

// GetAudience 获取受众
func (c MapClaims) GetAudience() []string {
	switch aud := c["aud"].(type) {
	case string:
		return []string{aud}
	case []interface{}:
		var result []string
		for _, v := range aud {
			if s, ok := v.(string); ok {
				result = append(result, s)
			}
		}
		return result
	case []string:
		return aud
	}
	return []string{}
}

// GetCustomClaims 获取自定义声明
func (c MapClaims) GetCustomClaims() map[string]interface{} {
	customClaims := make(map[string]interface{})

	// 排除标准字段
	standardFields := map[string]bool{
		"iss": true, "sub": true, "aud": true,
		"exp": true, "nbf": true, "iat": true, "jti": true,
	}

	for k, v := range c {
		if !standardFields[k] {
			customClaims[k] = v
		}
	}

	return customClaims
}

// SetCustomClaims 设置自定义声明
func (c MapClaims) SetCustomClaims(claims map[string]interface{}) {
	for k, v := range claims {
		c[k] = v
	}
}

// =========================================
// JWT 生成方法
// =========================================

// GenerateHS256Token 使用HMAC SHA-256算法生成JWT
func (j *jwtEngine) GenerateHS256Token(claims ILiteUtilJWTClaims, secretKey []byte) (string, error) {
	return j.GenerateToken(claims, HS256, secretKey, nil, nil)
}

// GenerateHS512Token 使用HMAC SHA-512算法生成JWT
func (j *jwtEngine) GenerateHS512Token(claims ILiteUtilJWTClaims, secretKey []byte) (string, error) {
	return j.GenerateToken(claims, HS512, secretKey, nil, nil)
}

// GenerateRS256Token 使用RSA SHA-256算法生成JWT
func (j *jwtEngine) GenerateRS256Token(claims ILiteUtilJWTClaims, privateKey *rsa.PrivateKey) (string, error) {
	return j.GenerateToken(claims, RS256, nil, privateKey, nil)
}

// GenerateES256Token 使用ECDSA P-256算法生成JWT
func (j *jwtEngine) GenerateES256Token(claims ILiteUtilJWTClaims, privateKey *ecdsa.PrivateKey) (string, error) {
	return j.GenerateToken(claims, ES256, nil, nil, privateKey)
}

// GenerateToken 通用JWT生成方法
func (j *jwtEngine) GenerateToken(claims ILiteUtilJWTClaims, algorithm JWTAlgorithm,
	secretKey []byte, rsaPrivateKey *rsa.PrivateKey, ecdsaPrivateKey *ecdsa.PrivateKey) (string, error) {

	// 创建头部
	header := newJWTHeader(algorithm)
	encodedHeader := j.encodeHeader(header)

	// 编码Claims
	encodedPayload, err := j.encodeClaims(claims)
	if err != nil {
		return "", fmt.Errorf("encode claims failed: %w", err)
	}

	// 创建签名消息
	message := encodedHeader + "." + encodedPayload

	// 生成签名
	signature, err := j.signMessage(message, algorithm, secretKey, rsaPrivateKey, ecdsaPrivateKey)
	if err != nil {
		return "", fmt.Errorf("sign message failed: %w", err)
	}

	return message + "." + signature, nil
}

// =========================================
// JWT 解析方法
// =========================================

// ParseHS256Token 解析使用HMAC SHA-256算法签名的JWT
func (j *jwtEngine) ParseHS256Token(token string, secretKey []byte) (MapClaims, error) {
	return j.ParseToken(token, HS256, secretKey, nil, nil)
}

// ParseHS512Token 解析使用HMAC SHA-512算法签名的JWT
func (j *jwtEngine) ParseHS512Token(token string, secretKey []byte) (MapClaims, error) {
	return j.ParseToken(token, HS512, secretKey, nil, nil)
}

// ParseRS256Token 解析使用RSA SHA-256算法签名的JWT
func (j *jwtEngine) ParseRS256Token(token string, publicKey *rsa.PublicKey) (MapClaims, error) {
	return j.ParseToken(token, RS256, nil, publicKey, nil)
}

// ParseES256Token 解析使用ECDSA P-256算法签名的JWT
func (j *jwtEngine) ParseES256Token(token string, publicKey *ecdsa.PublicKey) (MapClaims, error) {
	return j.ParseToken(token, ES256, nil, nil, publicKey)
}

// ParseToken 通用JWT解析方法
func (j *jwtEngine) ParseToken(token string, algorithm JWTAlgorithm,
	secretKey []byte, rsaPublicKey *rsa.PublicKey, ecdsaPublicKey *ecdsa.PublicKey) (MapClaims, error) {

	// 分割JWT
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid JWT format, must have 3 parts")
	}

	encodedHeader, encodedPayload, encodedSignature := parts[0], parts[1], parts[2]

	// 验证签名
	message := encodedHeader + "." + encodedPayload
	if err := j.verifySignature(message, encodedSignature, algorithm, secretKey, rsaPublicKey, ecdsaPublicKey); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	// 解析Claims
	claims, err := j.decodeClaims(encodedPayload)
	if err != nil {
		return nil, fmt.Errorf("decode claims failed: %w", err)
	}

	return claims, nil
}

// =========================================
// JWT 验证方法
// =========================================

// ValidateClaims 验证Claims的有效性
func (j *jwtEngine) ValidateClaims(claims ILiteUtilJWTClaims, options ...ValidateOption) error {
	opts := &ValidateOptions{
		CurrentTime: time.Now(),
	}

	for _, option := range options {
		option(opts)
	}

	// 验证过期时间
	if exp := claims.GetExpiresAt(); exp != nil {
		if opts.CurrentTime.After(*exp) {
			return errors.New("token is expired")
		}
	}

	// 验证生效时间
	if nbf := claims.GetNotBefore(); nbf != nil {
		if opts.CurrentTime.Before(*nbf) {
			return errors.New("token is not valid yet")
		}
	}

	// 验证签发者
	if opts.Issuer != "" && claims.GetIssuer() != opts.Issuer {
		return fmt.Errorf("invalid issuer, expected %s, got %s", opts.Issuer, claims.GetIssuer())
	}

	// 验证主题
	if opts.Subject != "" && claims.GetSubject() != opts.Subject {
		return fmt.Errorf("invalid subject, expected %s, got %s", opts.Subject, claims.GetSubject())
	}

	// 验证受众
	if len(opts.Audience) > 0 {
		claimAudience := claims.GetAudience()
		if !j.containsAny(claimAudience, opts.Audience) {
			return fmt.Errorf("invalid audience, expected one of %v, got %v", opts.Audience, claimAudience)
		}
	}

	return nil
}

// ValidateOptions JWT验证选项
type ValidateOptions struct {
	CurrentTime time.Time
	Issuer      string
	Subject     string
	Audience    []string
}

// ValidateOption JWT验证选项函数
type ValidateOption func(*ValidateOptions)

// WithIssuer 设置验证签发者
func WithIssuer(issuer string) ValidateOption {
	return func(opts *ValidateOptions) {
		opts.Issuer = issuer
	}
}

// WithSubject 设置验证主题
func WithSubject(subject string) ValidateOption {
	return func(opts *ValidateOptions) {
		opts.Subject = subject
	}
}

// WithAudience 设置验证受众
func WithAudience(audience ...string) ValidateOption {
	return func(opts *ValidateOptions) {
		opts.Audience = audience
	}
}

// WithCurrentTime 设置当前时间（用于测试）
func WithCurrentTime(t time.Time) ValidateOption {
	return func(opts *ValidateOptions) {
		opts.CurrentTime = t
	}
}

// =========================================
// 工具方法
// =========================================

// base64URLEncode URL安全的Base64编码
func (j *jwtEngine) base64URLEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

// base64URLDecode URL安全的Base64解码
func (j *jwtEngine) base64URLDecode(data string) ([]byte, error) {
	// 补充填充字符
	switch len(data) % 4 {
	case 2:
		data += "=="
	case 3:
		data += "="
	}

	return base64.URLEncoding.DecodeString(data)
}

// encodeClaims 编码Claims
func (j *jwtEngine) encodeClaims(claims ILiteUtilJWTClaims) (string, error) {
	var claimsMap map[string]interface{}
	var isFromPool bool

	// 根据Claims类型处理
	switch c := claims.(type) {
	case MapClaims:
		claimsMap = c
	case *StandardClaims:
		claimsMap = j.standardClaimsToMap(*c)
		isFromPool = true
	default:
		// 其他类型转换为MapClaims
		customClaims := claims.GetCustomClaims()
		claimsMap = make(map[string]interface{}, len(customClaims)+7)
		for k, v := range customClaims {
			claimsMap[k] = v
		}

		// 添加标准字段
		if exp := claims.GetExpiresAt(); exp != nil {
			claimsMap["exp"] = float64(exp.Unix())
		}
		if iat := claims.GetIssuedAt(); iat != nil {
			claimsMap["iat"] = float64(iat.Unix())
		}
		if nbf := claims.GetNotBefore(); nbf != nil {
			claimsMap["nbf"] = float64(nbf.Unix())
		}
		if iss := claims.GetIssuer(); iss != "" {
			claimsMap["iss"] = iss
		}
		if sub := claims.GetSubject(); sub != "" {
			claimsMap["sub"] = sub
		}
		if aud := claims.GetAudience(); len(aud) > 0 {
			if len(aud) == 1 {
				claimsMap["aud"] = aud[0]
			} else {
				claimsMap["aud"] = aud
			}
		}
	}

	claimsBytes, err := json.Marshal(claimsMap)
	if err != nil {
		if isFromPool {
			claimsMapPool.Put(claimsMap)
		}
		return "", err
	}

	result := j.base64URLEncode(claimsBytes)

	// 回收pool中的对象
	if isFromPool {
		claimsMapPool.Put(claimsMap)
	}

	return result, nil
}

// decodeClaims 解码Claims
func (j *jwtEngine) decodeClaims(encodedClaims string) (MapClaims, error) {
	claimsBytes, err := j.base64URLDecode(encodedClaims)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %w", err)
	}

	var claims MapClaims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}

	return claims, nil
}

// standardClaimsToMap StandardClaims转换为map
func (j *jwtEngine) standardClaimsToMap(claims StandardClaims) map[string]interface{} {
	result := claimsMapPool.Get().(map[string]interface{})

	for k := range result {
		delete(result, k)
	}

	if claims.Audience != nil {
		if len(claims.Audience) == 1 {
			result["aud"] = claims.Audience[0]
		} else {
			result["aud"] = claims.Audience
		}
	}
	if claims.ExpiresAt != 0 {
		result["exp"] = claims.ExpiresAt
	}
	if claims.ID != "" {
		result["jti"] = claims.ID
	}
	if claims.IssuedAt != 0 {
		result["iat"] = claims.IssuedAt
	}
	if claims.Issuer != "" {
		result["iss"] = claims.Issuer
	}
	if claims.NotBefore != 0 {
		result["nbf"] = claims.NotBefore
	}
	if claims.Subject != "" {
		result["sub"] = claims.Subject
	}

	return result
}

// signMessage 对消息进行签名
func (j *jwtEngine) signMessage(message string, algorithm JWTAlgorithm,
	secretKey []byte, rsaPrivateKey *rsa.PrivateKey, ecdsaPrivateKey *ecdsa.PrivateKey) (string, error) {

	switch algorithm {
	case HS256:
		return j.hmacSign(message, secretKey, crypto.SHA256)
	case HS384:
		return j.hmacSign(message, secretKey, crypto.SHA384)
	case HS512:
		return j.hmacSign(message, secretKey, crypto.SHA512)
	case RS256:
		return j.rsaSign(message, rsaPrivateKey, crypto.SHA256)
	case RS384:
		return j.rsaSign(message, rsaPrivateKey, crypto.SHA384)
	case RS512:
		return j.rsaSign(message, rsaPrivateKey, crypto.SHA512)
	case ES256:
		return j.ecdsaSign(message, ecdsaPrivateKey, crypto.SHA256)
	case ES384:
		return j.ecdsaSign(message, ecdsaPrivateKey, crypto.SHA384)
	case ES512:
		return j.ecdsaSign(message, ecdsaPrivateKey, crypto.SHA512)
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// hmacSign HMAC签名
func (j *jwtEngine) hmacSign(message string, key []byte, hash crypto.Hash) (string, error) {
	if !hash.Available() {
		return "", fmt.Errorf("hash algorithm not available: %v", hash)
	}

	h := hmac.New(hash.New, key)
	h.Write([]byte(message))
	signature := h.Sum(nil)

	return j.base64URLEncode(signature), nil
}

// rsaSign RSA签名
func (j *jwtEngine) rsaSign(message string, privateKey *rsa.PrivateKey, hash crypto.Hash) (string, error) {
	if privateKey == nil {
		return "", errors.New("RSA private key is required for RSA signing")
	}

	if !hash.Available() {
		return "", fmt.Errorf("hash algorithm not available: %v", hash)
	}

	hasher := hash.New()
	hasher.Write([]byte(message))
	hashed := hasher.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, hash, hashed)
	if err != nil {
		return "", fmt.Errorf("RSA sign failed: %w", err)
	}

	return j.base64URLEncode(signature), nil
}

// ecdsaSign ECDSA签名
func (j *jwtEngine) ecdsaSign(message string, privateKey *ecdsa.PrivateKey, hash crypto.Hash) (string, error) {
	if privateKey == nil {
		return "", errors.New("ECDSA private key is required for ECDSA signing")
	}

	if !hash.Available() {
		return "", fmt.Errorf("hash algorithm not available: %v", hash)
	}

	hasher := hash.New()
	hasher.Write([]byte(message))
	hashed := hasher.Sum(nil)

	signature, err := ecdsa.SignASN1(rand.Reader, privateKey, hashed)
	if err != nil {
		return "", fmt.Errorf("ECDSA sign failed: %w", err)
	}

	return j.base64URLEncode(signature), nil
}

// verifySignature 验证签名
func (j *jwtEngine) verifySignature(message, encodedSignature string, algorithm JWTAlgorithm,
	secretKey []byte, rsaPublicKey *rsa.PublicKey, ecdsaPublicKey *ecdsa.PublicKey) error {

	signature, err := j.base64URLDecode(encodedSignature)
	if err != nil {
		return fmt.Errorf("decode signature failed: %w", err)
	}

	switch algorithm {
	case HS256:
		return j.hmacVerify(message, signature, secretKey, crypto.SHA256)
	case HS384:
		return j.hmacVerify(message, signature, secretKey, crypto.SHA384)
	case HS512:
		return j.hmacVerify(message, signature, secretKey, crypto.SHA512)
	case RS256:
		return j.rsaVerify(message, signature, rsaPublicKey, crypto.SHA256)
	case RS384:
		return j.rsaVerify(message, signature, rsaPublicKey, crypto.SHA384)
	case RS512:
		return j.rsaVerify(message, signature, rsaPublicKey, crypto.SHA512)
	case ES256:
		return j.ecdsaVerify(message, signature, ecdsaPublicKey, crypto.SHA256)
	case ES384:
		return j.ecdsaVerify(message, signature, ecdsaPublicKey, crypto.SHA384)
	case ES512:
		return j.ecdsaVerify(message, signature, ecdsaPublicKey, crypto.SHA512)
	default:
		return fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// hmacVerify HMAC验证
func (j *jwtEngine) hmacVerify(message string, signature, key []byte, hash crypto.Hash) error {
	if !hash.Available() {
		return fmt.Errorf("hash algorithm not available: %v", hash)
	}

	h := hmac.New(hash.New, key)
	h.Write([]byte(message))
	expectedSignature := h.Sum(nil)

	if !hmac.Equal(signature, expectedSignature) {
		return errors.New("HMAC signature verification failed")
	}

	return nil
}

// rsaVerify RSA验证
func (j *jwtEngine) rsaVerify(message string, signature []byte, publicKey *rsa.PublicKey, hash crypto.Hash) error {
	if publicKey == nil {
		return errors.New("RSA public key is required for RSA verification")
	}

	if !hash.Available() {
		return fmt.Errorf("hash algorithm not available: %v", hash)
	}

	hasher := hash.New()
	hasher.Write([]byte(message))
	hashed := hasher.Sum(nil)

	err := rsa.VerifyPKCS1v15(publicKey, hash, hashed, signature)
	if err != nil {
		return fmt.Errorf("RSA signature verification failed: %w", err)
	}

	return nil
}

// ecdsaVerify ECDSA验证
func (j *jwtEngine) ecdsaVerify(message string, signature []byte, publicKey *ecdsa.PublicKey, hash crypto.Hash) error {
	if publicKey == nil {
		return errors.New("ECDSA public key is required for ECDSA verification")
	}

	if !hash.Available() {
		return fmt.Errorf("hash algorithm not available: %v", hash)
	}

	hasher := hash.New()
	hasher.Write([]byte(message))
	hashed := hasher.Sum(nil)

	if !ecdsa.VerifyASN1(publicKey, hashed, signature) {
		return errors.New("ECDSA signature verification failed")
	}

	return nil
}

// containsAny 检查切片是否包含目标中的任意元素
func (j *jwtEngine) containsAny(slice, targets []string) bool {
	targetMap := make(map[string]bool)
	for _, t := range targets {
		targetMap[t] = true
	}

	for _, item := range slice {
		if targetMap[item] {
			return true
		}
	}

	return false
}

// =========================================
// 便捷方法
// =========================================

// NewStandardClaims 创建标准Claims
func (j *jwtEngine) NewStandardClaims() *StandardClaims {
	return &StandardClaims{}
}

// NewMapClaims 创建映射Claims
func (j *jwtEngine) NewMapClaims() MapClaims {
	return make(MapClaims)
}

// SetExpiration 设置Claims过期时间
func (j *jwtEngine) SetExpiration(claims ILiteUtilJWTClaims, duration time.Duration) {
	if duration <= 0 {
		return
	}

	expiration := time.Now().Add(duration)

	switch c := claims.(type) {
	case *StandardClaims:
		c.ExpiresAt = expiration.Unix()
	case MapClaims:
		c["exp"] = float64(expiration.Unix())
	}
}

// SetIssuedAt 设置Claims签发时间
func (j *jwtEngine) SetIssuedAt(claims ILiteUtilJWTClaims, t time.Time) {
	switch c := claims.(type) {
	case *StandardClaims:
		c.IssuedAt = t.Unix()
	case MapClaims:
		c["iat"] = float64(t.Unix())
	}
}

// SetNotBefore 设置Claims生效时间
func (j *jwtEngine) SetNotBefore(claims ILiteUtilJWTClaims, t time.Time) {
	switch c := claims.(type) {
	case *StandardClaims:
		c.NotBefore = t.Unix()
	case MapClaims:
		c["nbf"] = float64(t.Unix())
	}
}

// SetIssuer 设置Claims签发者
func (j *jwtEngine) SetIssuer(claims ILiteUtilJWTClaims, issuer string) {
	switch c := claims.(type) {
	case *StandardClaims:
		c.Issuer = issuer
	case MapClaims:
		c["iss"] = issuer
	}
}

// SetSubject 设置Claims主题
func (j *jwtEngine) SetSubject(claims ILiteUtilJWTClaims, subject string) {
	switch c := claims.(type) {
	case *StandardClaims:
		c.Subject = subject
	case MapClaims:
		c["sub"] = subject
	}
}

// SetAudience 设置Claims受众
func (j *jwtEngine) SetAudience(claims ILiteUtilJWTClaims, audience ...string) {
	switch c := claims.(type) {
	case *StandardClaims:
		c.Audience = audience
	case MapClaims:
		if len(audience) == 1 {
			c["aud"] = audience[0]
		} else {
			c["aud"] = audience
		}
	}
}

// AddCustomClaim 添加自定义声明
func (j *jwtEngine) AddCustomClaim(claims ILiteUtilJWTClaims, key string, value interface{}) {
	switch c := claims.(type) {
	case MapClaims:
		c[key] = value
	default:
		// 对于其他类型，使用SetCustomClaims方法
		customClaims := claims.GetCustomClaims()
		customClaims[key] = value
		claims.SetCustomClaims(customClaims)
	}
}
