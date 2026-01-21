// Package hash 提供多种哈希算法和HMAC计算功能，支持泛型编程
//
// 核心特性：
//   - 支持多种哈希算法：MD5、SHA1、SHA256、SHA512
//   - 支持HMAC（基于哈希的消息认证码）计算
//   - 支持Bcrypt密码哈希（用于安全的密码存储和验证）
//   - 提供泛型函数，可扩展支持自定义哈希算法
//   - 支持多种输出格式：原始字节、16位/32位/完整长度十六进制字符串
//   - 提供便捷方法，通过 util.Hash 实例快速调用
//   - 支持 io.Reader 数据源，可处理流式数据
//
// 基本用法：
//
//	// 计算SHA256哈希值（返回十六进制字符串）
//	hashStr := util.Hash.SHA256String("hello world")
//	fmt.Println(hashStr) // b94d27b9...
//
//	// 计算HMAC-SHA256（需要密钥）
//	hmacStr := util.Hash.HMACSHA256String("data", "secret-key")
//	fmt.Println(hmacStr)
//
//	// 计算MD5并返回16位短格式
//	md5Short := util.Hash.MD5String16("filename")
//
//	// 使用 Bcrypt 哈希密码（安全存储）
//	hashedPassword, err := util.Hash.BcryptHash("mypassword")
//	if err != nil {
//	    logger.Fatal("密码哈希失败", "error", err)
//	}
//	// 验证密码
//	isValid := util.Hash.BcryptVerify("mypassword", hashedPassword)
//
//	// 使用泛型函数计算哈希值
//	hashBytes := hash.HashGeneric("data", hash.SHA256Algorithm{})
//
//	// 从文件计算哈希值
//	file, _ := os.Open("file.txt")
//	defer file.Close()
//	fileHash, err := hash.HashReaderStringGeneric(file, hash.SHA256Algorithm{})
//	if err != nil {
//	    logger.Fatal("文件哈希计算失败", "error", err)
//	}
//
// 泛型支持：
//
//	通过泛型函数 HashGeneric 和 HMACGeneric，可以扩展支持自定义哈希算法。
//	只需实现 HashAlgorithm 接口即可：
//
//	type CustomAlgorithm struct{}
//	func (CustomAlgorithm) Hash() hash.Hash {
//	    return sha3.New256() // 或其他哈希算法
//	}
//
//	result := hash.HashGeneric("data", CustomAlgorithm{})
//
// 输出格式：
//
//	Package hash 支持以下输出格式：
//	  - FormatBytes: 原始字节数组
//	  - FormatHexShort: 16位十六进制字符串（常用于MD5短格式）
//	  - FormatHexMedium: 32位十六进制字符串
//	  - FormatHexFull: 完整长度十六进制字符串（默认）
package hash
