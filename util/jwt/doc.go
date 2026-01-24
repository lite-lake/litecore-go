// Package jwt 提供JWT令牌的生成、解析和验证功能，支持多种签名算法。
//
// 核心特性：
//   - 支持多种签名算法：HS256/HS384/HS512、RS256、ES256等
//   - 提供简洁的API：Generate系列函数生成令牌，Parse系列函数解析令牌
//   - 内置Claims验证：支持过期时间、签发时间等标准声明验证
//   - 灵活的扩展性：支持自定义Claims结构
//   - 完善的错误处理：提供详细的错误信息便于调试
//
// 基本用法：
//
//	import (
//	    loggermgr "github.com/lite-lake/litecore-go/component/manager/loggermgr"
//	)
//
//	loggerMgr := loggermgr.GetLoggerManager()
//	logger := loggerMgr.Logger("main")
//
//	// 生成HS256令牌
//	claims := &jwt.StandardClaims{
//	    UserId:   "123456",
//	    Username: "john.doe",
//	    ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
//	}
//	token, err := util.jwt.GenerateHS256Token(claims, "your-secret-key")
//	if err != nil {
//	    logger.Fatal("Failed to generate token", "error", err)
//	}
//
//	// 解析HS256令牌
//	parsedClaims, err := util.jwt.ParseHS256Token(token, "your-secret-key")
//	if err != nil {
//	    logger.Fatal("Failed to parse token", "error", err)
//	}
//	fmt.Printf("用户ID: %s, 用户名: %s\n", parsedClaims.UserId, parsedClaims.Username)
//
//	// 生成RS256令牌（使用私钥签名）
//	privateKey := []byte("-----BEGIN RSA PRIVATE KEY-----\n...")
//	token, err = util.jwt.GenerateRS256Token(claims, privateKey)
//	if err != nil {
//	    logger.Fatal("Failed to generate token", "error", err)
//	}
//
//	// 解析RS256令牌（使用公钥验证）
//	publicKey := []byte("-----BEGIN PUBLIC KEY-----\n...")
//	parsedClaims, err = util.jwt.ParseRS256Token(token, publicKey)
//	if err != nil {
//	    logger.Fatal("Failed to parse token", "error", err)
//	}
//
//	// 验证Claims
//	if err := util.jwt.ValidateClaims(parsedClaims); err != nil {
//	    logger.Fatal("Failed to validate claims", "error", err)
//	}
//
// 支持的算法：
//
// HMAC算法（对称加密）：
//   - HS256: HMAC using SHA-256
//   - HS384: HMAC using SHA-384
//   - HS512: HMAC using SHA-512
//
// 非对称算法：
//   - RS256: RSASSA-PKCS1-v1_5 using SHA-256
//   - ES256: ECDSA using P-256 and SHA-256
//
// Claims结构：
//
// 标准Claims字段：
//   - Issuer (iss): 签发者
//   - Subject (sub): 主题
//   - Audience (aud): 接收方
//   - ExpiresAt (exp): 过期时间
//   - NotBefore (nbf): 生效时间
//   - IssuedAt (iat): 签发时间
//   - ID (jti): 令牌ID
//
// 注意事项：
//   - 密钥应妥善保管，建议从配置文件或环境变量读取
//   - 生产环境使用RS256等非对称算法更安全
//   - 令牌过期时间不宜过长，建议24小时内
//   - 密钥应足够复杂，避免使用弱密钥
package jwt
