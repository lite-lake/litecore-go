// Package crypt 提供全面的加密解密功能，包括对称加密、非对称加密、哈希、签名、Base64编码等
package crypt

/*
Crypt 包提供了常用的加密解密功能，支持多种加密算法和编码方式。

核心特性：
  - 对称加密：支持 AES 加密解密，适用于数据加密存储和传输
  - 非对称加密：支持 RSA 加密解密，适用于敏感数据保护和密钥交换
  - 哈希算法：支持 Bcrypt 密码哈希，适用于安全的密码存储
  - 数字签名：支持 HMAC 和 ECDSA 签名验证，确保数据完整性和真实性
  - 编码转换：支持 Base64 编码解码，便于数据传输和存储

基本用法：

	import "your-module-path/util/crypt"

	// Base64 编码解码
	encoded := crypt.Base64Encode("Hello, World!")
	// 输出: "SGVsbG8sIFdvcmxkIQ=="
	decoded, err := crypt.Base64Decode(encoded)
	if err != nil {
		log.Fatal(err)
	}

	// AES 对称加密解密
	key := []byte("32-byte-long-secret-key-1234567890")
	ciphertext, err := crypt.AESEncrypt("sensitive data", key)
	if err != nil {
		log.Fatal(err)
	}
	plaintext, err := crypt.AESDecrypt(ciphertext, key)
	if err != nil {
		log.Fatal(err)
	}

	// Bcrypt 密码哈希
	hashedPassword, err := crypt.BcryptHash("my-password", 10)
	if err != nil {
		log.Fatal(err)
	}
	err = crypt.BcryptVerify("my-password", hashedPassword)
	if err != nil {
		log.Println("密码验证失败")
	}

	// HMAC 签名验证
	secret := []byte("hmac-secret-key")
	message := "important message"
	signature := crypt.HMACSign(message, secret)
	valid := crypt.HMACVerify(message, signature, secret)

错误处理：
  所有加密解密函数均返回 error 类型，调用时应检查错误处理
  密钥长度、参数错误等会返回相应的错误信息，便于问题定位

性能考虑：
  - Bcrypt 的 cost 参数影响计算时间，建议值 10-12，可根据安全需求和性能权衡调整
  - RSA 操作相对较慢，大量数据加密建议使用 AES+RSA 混合加密方案
  - 密钥应妥善保管，避免硬编码在代码中，建议使用环境变量或配置管理服务
*/
