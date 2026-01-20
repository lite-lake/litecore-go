package hash

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"strings"
	"testing"
)

// =========================================
// 测试辅助函数
// =========================================

// 测试用例结构
type testCase struct {
	name     string
	data     string
	expected string
}

// getExpectedMD5 返回标准的MD5哈希值
func getExpectedMD5(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// getExpectedSHA1 返回标准的SHA1哈希值
func getExpectedSHA1(data string) string {
	hash := sha1.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// getExpectedSHA256 返回标准的SHA256哈希值
func getExpectedSHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// getExpectedSHA512 返回标准的SHA512哈希值
func getExpectedSHA512(data string) string {
	hash := sha512.Sum512([]byte(data))
	return hex.EncodeToString(hash[:])
}

// =========================================
// 测试 MD5 方法
// =========================================

func TestHashEngine_MD5(t *testing.T) {
	tests := []testCase{
		{"空字符串", "", "d41d8cd98f00b204e9800998ecf8427e"},
		{"简单字符串", "hello", "5d41402abc4b2a76b9719d911017c592"},
		{"中文字符串", "你好世界", "65396ee4aad0b4f17aacd1c6112ee364"},
		{"特殊字符", "!@#$%^&*()", "05b28d17a7b6e7024b6e5d8cc43a8bf7"},
		{"长字符串", strings.Repeat("a", 1000), "cabe45dcc9ae5b66ba86600cca6b8ba8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Hash.MD5(tt.data)
			resultHex := hex.EncodeToString(result)

			if resultHex != tt.expected {
				t.Errorf("MD5() = %v, want %v", resultHex, tt.expected)
			}
		})
	}
}

func TestHashEngine_MD5String(t *testing.T) {
	tests := []testCase{
		{"空字符串", "", "d41d8cd98f00b204e9800998ecf8427e"},
		{"简单字符串", "hello", "5d41402abc4b2a76b9719d911017c592"},
		{"中文字符串", "你好世界", "65396ee4aad0b4f17aacd1c6112ee364"},
		{"数字字符串", "123456", "e10adc3949ba59abbe56e057f20f883e"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Hash.MD5String(tt.data)

			if result != tt.expected {
				t.Errorf("MD5String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHashEngine_MD5String16(t *testing.T) {
	tests := []testCase{
		{"简单字符串", "hello", "5d41402abc4b2a76"},
		{"空字符串", "", "d41d8cd98f00b204"},
		{"中文字符串", "你好世界", "65396ee4aad0b4f1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Hash.MD5String16(tt.data)
			expected := tt.expected[:16]

			if result != expected {
				t.Errorf("MD5String16() = %v, want %v", result, expected)
			}

			// 验证长度
			if len(result) != 16 {
				t.Errorf("MD5String16() length = %v, want 16", len(result))
			}
		})
	}
}

func TestHashEngine_MD5String32(t *testing.T) {
	tests := []testCase{
		{"简单字符串", "hello", "5d41402abc4b2a76b9719d911017c592"},
		{"空字符串", "", "d41d8cd98f00b204e9800998ecf8427e"},
		{"中文字符串", "你好世界", "65396ee4aad0b4f17aacd1c6112ee364"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Hash.MD5String32(tt.data)

			if result != tt.expected {
				t.Errorf("MD5String32() = %v, want %v", result, tt.expected)
			}

			// 验证长度
			if len(result) != 32 {
				t.Errorf("MD5String32() length = %v, want 32", len(result))
			}
		})
	}
}

// =========================================
// 测试 SHA1 方法
// =========================================

func TestHashEngine_SHA1(t *testing.T) {
	tests := []testCase{
		{"空字符串", "", "da39a3ee5e6b4b0d3255bfef95601890afd80709"},
		{"简单字符串", "hello", "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"},
		{"中文字符串", "你好世界", "dabaa5fe7c47fb21be902480a13013f16a1ab6eb"},
		{"特殊字符", "!@#$%^&*()", "bf24d65c9bb05b9b814a966940bcfa50767c8a8d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Hash.SHA1(tt.data)
			resultHex := hex.EncodeToString(result)

			if resultHex != tt.expected {
				t.Errorf("SHA1() = %v, want %v", resultHex, tt.expected)
			}
		})
	}
}

func TestHashEngine_SHA1String(t *testing.T) {
	tests := []testCase{
		{"空字符串", "", "da39a3ee5e6b4b0d3255bfef95601890afd80709"},
		{"简单字符串", "hello", "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"},
		{"数字字符串", "123456", "7c4a8d09ca3762af61e59520943dc26494f8941b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Hash.SHA1String(tt.data)

			if result != tt.expected {
				t.Errorf("SHA1String() = %v, want %v", result, tt.expected)
			}

			// 验证长度 (SHA1 输出40个字符)
			if len(result) != 40 {
				t.Errorf("SHA1String() length = %v, want 40", len(result))
			}
		})
	}
}

// =========================================
// 测试 SHA256 方法
// =========================================

func TestHashEngine_SHA256(t *testing.T) {
	tests := []testCase{
		{"空字符串", "", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"简单字符串", "hello", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{"中文字符串", "你好世界", "beca6335b20ff57ccc47403ef4d9e0b8fccb4442b3151c2e7d50050673d43172"},
		{"特殊字符", "!@#$%^&*()", "95ce789c5c9d18490972709838ca3a9719094bca3ac16332cfec0652b0236141"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Hash.SHA256(tt.data)
			resultHex := hex.EncodeToString(result)

			if resultHex != tt.expected {
				t.Errorf("SHA256() = %v, want %v", resultHex, tt.expected)
			}
		})
	}
}

func TestHashEngine_SHA256String(t *testing.T) {
	tests := []testCase{
		{"空字符串", "", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"简单字符串", "hello", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{"数字字符串", "123456", "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Hash.SHA256String(tt.data)

			if result != tt.expected {
				t.Errorf("SHA256String() = %v, want %v", result, tt.expected)
			}

			// 验证长度 (SHA256 输出64个字符)
			if len(result) != 64 {
				t.Errorf("SHA256String() length = %v, want 64", len(result))
			}
		})
	}
}

// =========================================
// 测试 SHA512 方法
// =========================================

func TestHashEngine_SHA512(t *testing.T) {
	tests := []testCase{
		{"空字符串", "",
			"cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce" +
				"47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"},
		{"简单字符串", "hello",
			"9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca723" +
				"23c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"},
		{"中文字符串", "你好世界",
			"4b28a152c8e203ebb52e099301041e3cf704a56190d3097ec8b086a0f9bfb4b9" +
				"d533ce71fc3bcf374359e506dc5f17322ec3911eac8dd8f5b35308d938ba0c26"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Hash.SHA512(tt.data)
			resultHex := hex.EncodeToString(result)

			if resultHex != tt.expected {
				t.Errorf("SHA512() = %v, want %v", resultHex, tt.expected)
			}
		})
	}
}

func TestHashEngine_SHA512String(t *testing.T) {
	tests := []testCase{
		{"空字符串", "",
			"cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce" +
				"47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"},
		{"简单字符串", "hello",
			"9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca723" +
				"23c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"},
		{"数字字符串", "123456",
			"ba3253876aed6bc22d4a6ff53d8406c6ad864195ed144ab5c87621b6c233b548ba" +
				"eae6956df346ec8c17f5ea10f35ee3cbc514797ed7ddd3145464e2a0bab413"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Hash.SHA512String(tt.data)

			if result != tt.expected {
				t.Errorf("SHA512String() = %v, want %v", result, tt.expected)
			}

			// 验证长度 (SHA512 输出128个字符)
			if len(result) != 128 {
				t.Errorf("SHA512String() length = %v, want 128", len(result))
			}
		})
	}
}

// =========================================
// 测试 HMAC 方法
// =========================================

func TestHashEngine_HMACMD5(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		key      string
		expected string
	}{
		{"简单HMAC-MD5", "hello", "key", "04130747afca4d79e32e87cf2104f087"},
		{"空数据", "", "key", "63530468a04e386459855da0063b6596"},
		{"空密钥", "hello", "", "2a566e7a1b0190f15c0e7f523012cdc9"},
		{"中文字符", "你好世界", "密钥", "17857b8115ac41bdacf92b44a92898d5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Hash.HMACMD5(tt.data, tt.key)
			resultHex := hex.EncodeToString(result)

			if resultHex != tt.expected {
				t.Errorf("HMACMD5() = %v, want %v", resultHex, tt.expected)
			}
		})
	}
}

func TestHashEngine_HMACMD5String(t *testing.T) {
	data := "hello"
	key := "secret"
	result := Hash.HMACMD5String(data, key)

	if len(result) != 32 {
		t.Errorf("HMACMD5String() length = %v, want 32", len(result))
	}

	// 验证一致性
	result2 := Hash.HMACMD5String(data, key)
	if result != result2 {
		t.Error("HMACMD5String() 重复调用结果不一致")
	}
}

func TestHashEngine_HMACSHA1(t *testing.T) {
	tests := []struct {
		name string
		data string
		key  string
	}{
		{"简单HMAC-SHA1", "hello", "key"},
		{"空数据", "", "key"},
		{"空密钥", "hello", ""},
		{"中文字符", "你好世界", "密钥"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Hash.HMACSHA1(tt.data, tt.key)
			resultHex := hex.EncodeToString(result)

			// 验证长度 (SHA1 HMAC 输出20字节=40个十六进制字符)
			if len(resultHex) != 40 {
				t.Errorf("HMACSHA1() length = %v, want 40", len(resultHex))
			}

			// 验证一致性
			result2 := Hash.HMACSHA1(tt.data, tt.key)
			result2Hex := hex.EncodeToString(result2)
			if resultHex != result2Hex {
				t.Error("HMACSHA1() 重复调用结果不一致")
			}
		})
	}
}

func TestHashEngine_HMACSHA1String(t *testing.T) {
	data := "hello"
	key := "secret"
	result := Hash.HMACSHA1String(data, key)

	if len(result) != 40 {
		t.Errorf("HMACSHA1String() length = %v, want 40", len(result))
	}
}

func TestHashEngine_HMACSHA256(t *testing.T) {
	tests := []struct {
		name string
		data string
		key  string
	}{
		{"简单HMAC-SHA256", "hello", "key"},
		{"空数据", "", "key"},
		{"空密钥", "hello", ""},
		{"中文字符", "你好世界", "密钥"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Hash.HMACSHA256(tt.data, tt.key)
			resultHex := hex.EncodeToString(result)

			// 验证长度 (SHA256 HMAC 输出32字节=64个十六进制字符)
			if len(resultHex) != 64 {
				t.Errorf("HMACSHA256() length = %v, want 64", len(resultHex))
			}

			// 验证一致性
			result2 := Hash.HMACSHA256(tt.data, tt.key)
			result2Hex := hex.EncodeToString(result2)
			if resultHex != result2Hex {
				t.Error("HMACSHA256() 重复调用结果不一致")
			}
		})
	}
}

func TestHashEngine_HMACSHA256String(t *testing.T) {
	data := "hello"
	key := "secret"
	result := Hash.HMACSHA256String(data, key)

	if len(result) != 64 {
		t.Errorf("HMACSHA256String() length = %v, want 64", len(result))
	}
}

func TestHashEngine_HMACSHA512(t *testing.T) {
	tests := []struct {
		name string
		data string
		key  string
	}{
		{"简单HMAC-SHA512", "hello", "key"},
		{"空数据", "", "key"},
		{"空密钥", "hello", ""},
		{"中文字符", "你好世界", "密钥"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Hash.HMACSHA512(tt.data, tt.key)
			resultHex := hex.EncodeToString(result)

			// 验证长度 (SHA512 HMAC 输出64字节=128个十六进制字符)
			if len(resultHex) != 128 {
				t.Errorf("HMACSHA512() length = %v, want 128", len(resultHex))
			}

			// 验证一致性
			result2 := Hash.HMACSHA512(tt.data, tt.key)
			result2Hex := hex.EncodeToString(result2)
			if resultHex != result2Hex {
				t.Error("HMACSHA512() 重复调用结果不一致")
			}
		})
	}
}

func TestHashEngine_HMACSHA512String(t *testing.T) {
	data := "hello"
	key := "secret"
	result := Hash.HMACSHA512String(data, key)

	if len(result) != 128 {
		t.Errorf("HMACSHA512String() length = %v, want 128", len(result))
	}
}

// =========================================
// 测试泛型哈希函数
// =========================================

func TestHashGeneric(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		algorithm HashAlgorithm
		wantLen   int
	}{
		{"MD5", "hello", MD5Algorithm{}, 16},
		{"SHA1", "hello", SHA1Algorithm{}, 20},
		{"SHA256", "hello", SHA256Algorithm{}, 32},
		{"SHA512", "hello", SHA512Algorithm{}, 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HashGeneric(tt.data, tt.algorithm)

			if len(result) != tt.wantLen {
				t.Errorf("HashGeneric() length = %v, want %v", len(result), tt.wantLen)
			}
		})
	}
}

func TestHashBytesGeneric(t *testing.T) {
	data := []byte("hello world")

	tests := []struct {
		name      string
		algorithm HashAlgorithm
		wantLen   int
	}{
		{"MD5", MD5Algorithm{}, 16},
		{"SHA1", SHA1Algorithm{}, 20},
		{"SHA256", SHA256Algorithm{}, 32},
		{"SHA512", SHA512Algorithm{}, 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HashBytesGeneric(data, tt.algorithm)

			if len(result) != tt.wantLen {
				t.Errorf("HashBytesGeneric() length = %v, want %v", len(result), tt.wantLen)
			}
		})
	}
}

func TestHashReaderGeneric(t *testing.T) {
	data := "hello world"
	reader := strings.NewReader(data)

	tests := []struct {
		name      string
		algorithm HashAlgorithm
		wantLen   int
	}{
		{"MD5", MD5Algorithm{}, 16},
		{"SHA1", SHA1Algorithm{}, 20},
		{"SHA256", SHA256Algorithm{}, 32},
		{"SHA512", SHA512Algorithm{}, 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HashReaderGeneric(reader, tt.algorithm)

			if err != nil {
				t.Errorf("HashReaderGeneric() error = %v", err)
			}

			if len(result) != tt.wantLen {
				t.Errorf("HashReaderGeneric() length = %v, want %v", len(result), tt.wantLen)
			}
		})
	}
}

func TestHashReaderGeneric_Error(t *testing.T) {
	// 创建一个会返回错误的 reader
	errorReader := &errorReader{}

	_, err := HashReaderGeneric(errorReader, MD5Algorithm{})
	if err == nil {
		t.Error("HashReaderGeneric() 应该返回错误但没有")
	}
}

// =========================================
// 测试泛型哈希十六进制函数
// =========================================

func TestHashHexGeneric(t *testing.T) {
	data := "hello"

	tests := []struct {
		name      string
		algorithm HashAlgorithm
		format    HashOutputFormat
		wantLen   int
	}{
		{"MD5-Short", MD5Algorithm{}, FormatHexShort, 16},
		{"MD5-Medium", MD5Algorithm{}, FormatHexMedium, 32},
		{"MD5-Full", MD5Algorithm{}, FormatHexFull, 32},
		{"SHA1-Full", SHA1Algorithm{}, FormatHexFull, 40},
		{"SHA256-Full", SHA256Algorithm{}, FormatHexFull, 64},
		{"SHA512-Full", SHA512Algorithm{}, FormatHexFull, 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HashHexGeneric(data, tt.algorithm, tt.format)

			if len(result) != tt.wantLen {
				t.Errorf("HashHexGeneric() length = %v, want %v", len(result), tt.wantLen)
			}
		})
	}
}

func TestHashBytesHexGeneric(t *testing.T) {
	data := []byte("hello")

	result := HashBytesHexGeneric(data, SHA256Algorithm{}, FormatHexFull)

	if len(result) != 64 {
		t.Errorf("HashBytesHexGeneric() length = %v, want 64", len(result))
	}
}

func TestHashReaderHexGeneric(t *testing.T) {
	reader := strings.NewReader("hello world")

	result, err := HashReaderHexGeneric(reader, SHA256Algorithm{}, FormatHexFull)

	if err != nil {
		t.Errorf("HashReaderHexGeneric() error = %v", err)
	}

	if len(result) != 64 {
		t.Errorf("HashReaderHexGeneric() length = %v, want 64", len(result))
	}
}

func TestHashStringGeneric(t *testing.T) {
	data := "hello world"

	tests := []struct {
		name      string
		algorithm HashAlgorithm
		wantLen   int
	}{
		{"MD5", MD5Algorithm{}, 32},
		{"SHA1", SHA1Algorithm{}, 40},
		{"SHA256", SHA256Algorithm{}, 64},
		{"SHA512", SHA512Algorithm{}, 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HashStringGeneric(data, tt.algorithm)

			if len(result) != tt.wantLen {
				t.Errorf("HashStringGeneric() length = %v, want %v", len(result), tt.wantLen)
			}
		})
	}
}

func TestHashReaderStringGeneric(t *testing.T) {
	reader := strings.NewReader("hello")

	result, err := HashReaderStringGeneric(reader, MD5Algorithm{})

	if err != nil {
		t.Errorf("HashReaderStringGeneric() error = %v", err)
	}

	if len(result) != 32 {
		t.Errorf("HashReaderStringGeneric() length = %v, want 32", len(result))
	}
}

// =========================================
// 测试泛型 HMAC 函数
// =========================================

func TestHMACGeneric(t *testing.T) {
	data := "hello"
	key := "secret"

	tests := []struct {
		name      string
		algorithm HashAlgorithm
		wantLen   int
	}{
		{"HMAC-MD5", MD5Algorithm{}, 16},
		{"HMAC-SHA1", SHA1Algorithm{}, 20},
		{"HMAC-SHA256", SHA256Algorithm{}, 32},
		{"HMAC-SHA512", SHA512Algorithm{}, 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HMACGeneric(data, key, tt.algorithm)

			if len(result) != tt.wantLen {
				t.Errorf("HMACGeneric() length = %v, want %v", len(result), tt.wantLen)
			}
		})
	}
}

func TestHMACBytesGeneric(t *testing.T) {
	data := []byte("hello")
	key := []byte("secret")

	tests := []struct {
		name      string
		algorithm HashAlgorithm
		wantLen   int
	}{
		{"HMAC-MD5", MD5Algorithm{}, 16},
		{"HMAC-SHA1", SHA1Algorithm{}, 20},
		{"HMAC-SHA256", SHA256Algorithm{}, 32},
		{"HMAC-SHA512", SHA512Algorithm{}, 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HMACBytesGeneric(data, key, tt.algorithm)

			if len(result) != tt.wantLen {
				t.Errorf("HMACBytesGeneric() length = %v, want %v", len(result), tt.wantLen)
			}
		})
	}
}

func TestHMACReaderGeneric(t *testing.T) {
	reader := strings.NewReader("hello world")
	key := []byte("secret")

	tests := []struct {
		name      string
		algorithm HashAlgorithm
		wantLen   int
	}{
		{"HMAC-MD5", MD5Algorithm{}, 16},
		{"HMAC-SHA1", SHA1Algorithm{}, 20},
		{"HMAC-SHA256", SHA256Algorithm{}, 32},
		{"HMAC-SHA512", SHA512Algorithm{}, 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HMACReaderGeneric(reader, key, tt.algorithm)

			if err != nil {
				t.Errorf("HMACReaderGeneric() error = %v", err)
			}

			if len(result) != tt.wantLen {
				t.Errorf("HMACReaderGeneric() length = %v, want %v", len(result), tt.wantLen)
			}
		})
	}
}

// =========================================
// 测试泛型 HMAC 十六进制函数
// =========================================

func TestHMACHexGeneric(t *testing.T) {
	data := "hello"
	key := "secret"

	tests := []struct {
		name      string
		algorithm HashAlgorithm
		format    HashOutputFormat
		wantLen   int
	}{
		{"MD5-Short", MD5Algorithm{}, FormatHexShort, 16},
		{"MD5-Medium", MD5Algorithm{}, FormatHexMedium, 32},
		{"MD5-Full", MD5Algorithm{}, FormatHexFull, 32},
		{"SHA1-Full", SHA1Algorithm{}, FormatHexFull, 40},
		{"SHA256-Full", SHA256Algorithm{}, FormatHexFull, 64},
		{"SHA512-Full", SHA512Algorithm{}, FormatHexFull, 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HMACHexGeneric(data, key, tt.algorithm, tt.format)

			if len(result) != tt.wantLen {
				t.Errorf("HMACHexGeneric() length = %v, want %v", len(result), tt.wantLen)
			}
		})
	}
}

func TestHMACBytesHexGeneric(t *testing.T) {
	data := []byte("hello")
	key := []byte("secret")

	result := HMACBytesHexGeneric(data, key, SHA256Algorithm{}, FormatHexFull)

	if len(result) != 64 {
		t.Errorf("HMACBytesHexGeneric() length = %v, want 64", len(result))
	}
}

func TestHMACReaderHexGeneric(t *testing.T) {
	reader := strings.NewReader("hello world")
	key := []byte("secret")

	result, err := HMACReaderHexGeneric(reader, key, SHA256Algorithm{}, FormatHexFull)

	if err != nil {
		t.Errorf("HMACReaderHexGeneric() error = %v", err)
	}

	if len(result) != 64 {
		t.Errorf("HMACReaderHexGeneric() length = %v, want 64", len(result))
	}
}

func TestHMACStringGeneric(t *testing.T) {
	data := "hello"
	key := "secret"

	tests := []struct {
		name      string
		algorithm HashAlgorithm
		wantLen   int
	}{
		{"HMAC-MD5", MD5Algorithm{}, 32},
		{"HMAC-SHA1", SHA1Algorithm{}, 40},
		{"HMAC-SHA256", SHA256Algorithm{}, 64},
		{"HMAC-SHA512", SHA512Algorithm{}, 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HMACStringGeneric(data, key, tt.algorithm)

			if len(result) != tt.wantLen {
				t.Errorf("HMACStringGeneric() length = %v, want %v", len(result), tt.wantLen)
			}
		})
	}
}

func TestHMACReaderStringGeneric(t *testing.T) {
	reader := strings.NewReader("hello")
	key := []byte("secret")

	result, err := HMACReaderStringGeneric(reader, key, MD5Algorithm{})

	if err != nil {
		t.Errorf("HMACReaderStringGeneric() error = %v", err)
	}

	if len(result) != 32 {
		t.Errorf("HMACReaderStringGeneric() length = %v, want 32", len(result))
	}
}

// =========================================
// 测试哈希输出格式
// =========================================

func TestFormatHash(t *testing.T) {
	tests := []struct {
		name         string
		hashBytes    []byte
		format       HashOutputFormat
		wantContains string
		wantLen      int
	}{
		{
			name:         "FormatBytes-16字节",
			hashBytes:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10},
			format:       FormatBytes,
			wantContains: "\x01\x02\x03",
			wantLen:      16,
		},
		{
			name:      "FormatHexShort-短哈希",
			hashBytes: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10},
			format:    FormatHexShort,
			wantLen:   16,
		},
		{
			name:      "FormatHexMedium-中等哈希",
			hashBytes: make([]byte, 32),
			format:    FormatHexMedium,
			wantLen:   32,
		},
		{
			name:      "FormatHexFull-完整哈希",
			hashBytes: make([]byte, 32),
			format:    FormatHexFull,
			wantLen:   64,
		},
		{
			name:      "FormatHexShort-超短哈希",
			hashBytes: []byte{0x01, 0x02},
			format:    FormatHexShort,
			wantLen:   4,
		},
		{
			name:      "FormatHexMedium-超短哈希",
			hashBytes: []byte{0x01, 0x02},
			format:    FormatHexMedium,
			wantLen:   4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatHash(tt.hashBytes, tt.format)

			if len(result) != tt.wantLen {
				t.Errorf("formatHash() length = %v, want %v", len(result), tt.wantLen)
			}

			if tt.wantContains != "" && !strings.Contains(result, tt.wantContains) {
				t.Errorf("formatHash() 应该包含 %q", tt.wantContains)
			}
		})
	}
}

// =========================================
// 边界条件测试
// =========================================

func TestBoundaryConditions(t *testing.T) {
	t.Run("空字符串MD5", func(t *testing.T) {
		result := Hash.MD5String("")
		expected := "d41d8cd98f00b204e9800998ecf8427e"
		if result != expected {
			t.Errorf("空字符串MD5 = %v, want %v", result, expected)
		}
	})

	t.Run("空字符串SHA256", func(t *testing.T) {
		result := Hash.SHA256String("")
		expected := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		if result != expected {
			t.Errorf("空字符串SHA256 = %v, want %v", result, expected)
		}
	})

	t.Run("超长字符串", func(t *testing.T) {
		longData := strings.Repeat("a", 10000)
		result := Hash.MD5String(longData)

		if len(result) != 32 {
			t.Errorf("超长字符串MD5长度 = %v, want 32", len(result))
		}
	})

	t.Run("特殊Unicode字符", func(t *testing.T) {
		special := "🔥💯✨🚀"
		result := Hash.SHA256String(special)

		if len(result) != 64 {
			t.Errorf("特殊Unicode字符SHA256长度 = %v, want 64", len(result))
		}
	})

	t.Run("HMAC空密钥", func(t *testing.T) {
		data := "hello"
		result := Hash.HMACSHA256(data, "")

		if len(result) != 32 {
			t.Errorf("HMAC空密钥长度 = %v, want 32", len(result))
		}
	})

	t.Run("HMAC空数据", func(t *testing.T) {
		key := "secret"
		result := Hash.HMACSHA256("", key)

		if len(result) != 32 {
			t.Errorf("HMAC空数据长度 = %v, want 32", len(result))
		}
	})
}

// =========================================
// 一致性测试
// =========================================

func TestConsistency(t *testing.T) {
	data := "test data"

	t.Run("MD5一致性", func(t *testing.T) {
		result1 := Hash.MD5String(data)
		result2 := Hash.MD5String(data)

		if result1 != result2 {
			t.Error("MD5 多次调用结果不一致")
		}
	})

	t.Run("SHA256一致性", func(t *testing.T) {
		result1 := Hash.SHA256String(data)
		result2 := Hash.SHA256String(data)

		if result1 != result2 {
			t.Error("SHA256 多次调用结果不一致")
		}
	})

	t.Run("HMAC-SHA256一致性", func(t *testing.T) {
		key := "secret"
		result1 := Hash.HMACSHA256String(data, key)
		result2 := Hash.HMACSHA256String(data, key)

		if result1 != result2 {
			t.Error("HMAC-SHA256 多次调用结果不一致")
		}
	})

	t.Run("泛型函数与便捷方法一致性", func(t *testing.T) {
		result1 := Hash.MD5String(data)
		result2 := HashStringGeneric(data, MD5Algorithm{})

		if result1 != result2 {
			t.Error("泛型函数与便捷方法结果不一致")
		}
	})
}

// =========================================
// 性能基准测试
// =========================================

func BenchmarkMD5(b *testing.B) {
	data := strings.Repeat("a", 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Hash.MD5(data)
	}
}

func BenchmarkSHA256(b *testing.B) {
	data := strings.Repeat("a", 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Hash.SHA256(data)
	}
}

func BenchmarkSHA512(b *testing.B) {
	data := strings.Repeat("a", 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Hash.SHA512(data)
	}
}

func BenchmarkHMACSHA256(b *testing.B) {
	data := strings.Repeat("a", 1000)
	key := "secret"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Hash.HMACSHA256(data, key)
	}
}

// =========================================
// 测试辅助类型
// =========================================

// errorReader 用于测试错误处理
type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, &testReadError{}
}

// testReadError 测试用错误类型
type testReadError struct{}

func (e *testReadError) Error() string {
	return "test read error"
}
