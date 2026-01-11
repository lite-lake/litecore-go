package crypt

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// =========================================
// Base64 编码解码测试
// =========================================

func TestBase64Encode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "简单字符串",
			input:    "Hello, World!",
			expected: "SGVsbG8sIFdvcmxkIQ==",
		},
		{
			name:     "空字符串",
			input:    "",
			expected: "",
		},
		{
			name:     "中文字符串",
			input:    "你好世界",
			expected: "5L2g5aW95LiW55WM",
		},
		{
			name:     "特殊字符",
			input:    "!@#$%^&*()",
			expected: "IUAjJCVeJiooKQ==",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Crypt.Base64Encode(tt.input)
			if result != tt.expected {
				t.Errorf("Base64Encode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBase64EncodeBytes(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "字节数组",
			input: []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f},
		},
		{
			name:  "空字节数组",
			input: []byte{},
		},
		{
			name:  "包含null字节",
			input: []byte{0x00, 0x01, 0x02},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Crypt.Base64EncodeBytes(tt.input)
			// 验证可以正确解码
			decoded, err := Crypt.Base64DecodeBytes(result)
			if err != nil {
				t.Errorf("Base64EncodeBytes() 产生无效的Base64: %v", err)
			}
			if !bytes.Equal(decoded, tt.input) {
				t.Errorf("Base64EncodeBytes() 编码解码不匹配")
			}
		})
	}
}

func TestBase64Decode(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		expectErr bool
	}{
		{
			name:      "有效Base64",
			input:     "SGVsbG8sIFdvcmxkIQ==",
			expected:  "Hello, World!",
			expectErr: false,
		},
		{
			name:      "空字符串",
			input:     "",
			expected:  "",
			expectErr: false,
		},
		{
			name:      "无效Base64",
			input:     "Invalid@Base64!",
			expectErr: true,
		},
		{
			name:      "中文字符",
			input:     "5L2g5aW95LiW55WM",
			expected:  "你好世界",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Crypt.Base64Decode(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Error("Base64Decode() 预期返回错误，但没有")
				}
			} else {
				if err != nil {
					t.Errorf("Base64Decode() 意外错误: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Base64Decode() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestBase64DecodeBytes(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name:      "有效Base64",
			input:     "SGVsbG8sIFdvcmxkIQ==",
			expectErr: false,
		},
		{
			name:      "无效Base64",
			input:     "Invalid!@#",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Crypt.Base64DecodeBytes(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Error("Base64DecodeBytes() 预期返回错误，但没有")
				}
			} else {
				if err != nil {
					t.Errorf("Base64DecodeBytes() 意外错误: %v", err)
				}
				if len(result) == 0 && tt.input != "" {
					t.Error("Base64DecodeBytes() 返回空结果")
				}
			}
		})
	}
}

func TestBase64URLEncode(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "普通字符串",
			input: "Hello, World!",
		},
		{
			name:  "URL特殊字符",
			input: "?&=#+",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Crypt.Base64URLEncode(tt.input)
			// 验证可以解码
			decoded, err := Crypt.Base64URLDecode(result)
			if err != nil {
				t.Errorf("Base64URLEncode() 产生无效的URL Base64: %v", err)
			}
			if decoded != tt.input {
				t.Errorf("Base64URLEncode() 编码解码不匹配: got %v, want %v", decoded, tt.input)
			}
		})
	}
}

func TestBase64URLDecode(t *testing.T) {
	// 先生成一个有效的URL Base64字符串
	validInput := Crypt.Base64URLEncode("Hello, World!")

	tests := []struct {
		name      string
		input     string
		expected  string
		expectErr bool
	}{
		{
			name:      "有效URL Base64",
			input:     validInput,
			expected:  "Hello, World!",
			expectErr: false,
		},
		{
			name:      "无效URL Base64",
			input:     "Invalid!@#",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Crypt.Base64URLDecode(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Error("Base64URLDecode() 预期返回错误，但没有")
				}
			} else {
				if err != nil {
					t.Errorf("Base64URLDecode() 意外错误: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Base64URLDecode() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

// =========================================
// Hex 编码解码测试
// =========================================

func TestHexEncode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "简单字符串",
			input:    "Hello",
			expected: "48656c6c6f",
		},
		{
			name:     "空字符串",
			input:    "",
			expected: "",
		},
		{
			name:     "数字",
			input:    "123",
			expected: "313233",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Crypt.HexEncode(tt.input)
			if result != tt.expected {
				t.Errorf("HexEncode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHexEncodeBytes(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "字节数组",
			input: []byte{0x00, 0xFF, 0xAA, 0x55},
		},
		{
			name:  "空字节数组",
			input: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Crypt.HexEncodeBytes(tt.input)
			// 验证可以解码
			decoded, err := Crypt.HexDecodeBytes(result)
			if err != nil {
				t.Errorf("HexEncodeBytes() 产生无效的Hex: %v", err)
			}
			if !bytes.Equal(decoded, tt.input) {
				t.Errorf("HexEncodeBytes() 编码解码不匹配")
			}
		})
	}
}

func TestHexDecode(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		expectErr bool
	}{
		{
			name:      "有效Hex",
			input:     "48656c6c6f",
			expected:  "Hello",
			expectErr: false,
		},
		{
			name:      "空字符串",
			input:     "",
			expected:  "",
			expectErr: false,
		},
		{
			name:      "无效Hex",
			input:     "GGGG",
			expectErr: true,
		},
		{
			name:      "奇数长度",
			input:     "abc",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Crypt.HexDecode(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Error("HexDecode() 预期返回错误，但没有")
				}
			} else {
				if err != nil {
					t.Errorf("HexDecode() 意外错误: %v", err)
				}
				if result != tt.expected {
					t.Errorf("HexDecode() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestHexDecodeBytes(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name:      "有效Hex",
			input:     "48656c6c6f",
			expectErr: false,
		},
		{
			name:      "无效Hex",
			input:     "xyz",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Crypt.HexDecodeBytes(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Error("HexDecodeBytes() 预期返回错误，但没有")
				}
			} else {
				if err != nil {
					t.Errorf("HexDecodeBytes() 意外错误: %v", err)
				}
				if len(result) == 0 && tt.input != "" {
					t.Error("HexDecodeBytes() 返回空结果")
				}
			}
		})
	}
}

// =========================================
// AES 对称加密测试
// =========================================

func TestGenerateAESKey(t *testing.T) {
	tests := []struct {
		name      string
		keySize   AESKeySize
		expectErr bool
	}{
		{
			name:      "AES128密钥",
			keySize:   AES128,
			expectErr: false,
		},
		{
			name:      "AES192密钥",
			keySize:   AES192,
			expectErr: false,
		},
		{
			name:      "AES256密钥",
			keySize:   AES256,
			expectErr: false,
		},
		{
			name:      "无效密钥大小",
			keySize:   10,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Crypt.GenerateAESKey(tt.keySize)
			if tt.expectErr {
				if err == nil {
					t.Error("GenerateAESKey() 预期返回错误，但没有")
				}
			} else {
				if err != nil {
					t.Errorf("GenerateAESKey() 意外错误: %v", err)
				}
				if len(result) != int(tt.keySize) {
					t.Errorf("GenerateAESKey() 密钥长度错误: got %d, want %d", len(result), tt.keySize)
				}
			}
		})
	}
}

func TestGenerateAESKeyHex(t *testing.T) {
	tests := []struct {
		name      string
		keySize   AESKeySize
		expectErr bool
	}{
		{
			name:      "生成AES128 Hex密钥",
			keySize:   AES128,
			expectErr: false,
		},
		{
			name:      "生成AES256 Hex密钥",
			keySize:   AES256,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Crypt.GenerateAESKeyHex(tt.keySize)
			if tt.expectErr {
				if err == nil {
					t.Error("GenerateAESKeyHex() 预期返回错误，但没有")
				}
			} else {
				if err != nil {
					t.Errorf("GenerateAESKeyHex() 意外错误: %v", err)
				}
				// 验证是有效的Hex字符串
				_, err = Crypt.HexDecode(result)
				if err != nil {
					t.Errorf("GenerateAESKeyHex() 返回无效的Hex: %v", err)
				}
				// 验证长度（Hex编码是原始长度的2倍）
				expectedLen := int(tt.keySize) * 2
				if len(result) != expectedLen {
					t.Errorf("GenerateAESKeyHex() Hex长度错误: got %d, want %d", len(result), expectedLen)
				}
			}
		})
	}
}

func TestAESEncryptDecrypt(t *testing.T) {
	// 生成密钥
	key, err := Crypt.GenerateAESKey(AES256)
	if err != nil {
		t.Fatalf("生成AES密钥失败: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "短文本",
			plaintext: "Hello, World!",
		},
		{
			name:      "长文本",
			plaintext: strings.Repeat("A", 1000),
		},
		{
			name:      "包含特殊字符",
			plaintext: "!@#$%^&*()_+-=[]{}|;':\",./<>?",
		},
		{
			name:      "包含中文字符",
			plaintext: "这是一个测试文本，包含中文字符和特殊符号！@#￥%……&*（）",
		},
		{
			name:      "空字符串",
			plaintext: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 加密
			ciphertext, err := Crypt.AESEncrypt([]byte(tt.plaintext), key)
			if err != nil {
				t.Errorf("AESEncrypt() 失败: %v", err)
				return
			}

			// 验证密文不为空
			if len(ciphertext) == 0 && tt.plaintext != "" {
				t.Error("AESEncrypt() 返回空密文")
			}

			// 验证密文与原文不同
			if bytes.Equal(ciphertext, []byte(tt.plaintext)) {
				t.Error("AESEncrypt() 密文与原文相同")
			}

			// 解密
			decrypted, err := Crypt.AESDecrypt(ciphertext, key)
			if err != nil {
				t.Errorf("AESDecrypt() 失败: %v", err)
				return
			}

			// 验证解密结果与原文相同
			if string(decrypted) != tt.plaintext {
				t.Errorf("AESDecrypt() 解密结果不匹配: got %v, want %v", string(decrypted), tt.plaintext)
			}
		})
	}
}

func TestAESEncryptToBase64_AESDecryptFromBase64(t *testing.T) {
	key, err := Crypt.GenerateAESKey(AES256)
	if err != nil {
		t.Fatalf("生成AES密钥失败: %v", err)
	}

	plaintext := "这是一个敏感信息，需要加密保护！"

	// 加密并Base64编码
	encrypted, err := Crypt.AESEncryptToBase64(plaintext, key)
	if err != nil {
		t.Fatalf("AESEncryptToBase64() 失败: %v", err)
	}

	// 验证是有效的Base64
	if !Crypt.IsBase64(encrypted) {
		t.Error("AESEncryptToBase64() 返回无效的Base64")
	}

	// 从Base64解密
	decrypted, err := Crypt.AESDecryptFromBase64(encrypted, key)
	if err != nil {
		t.Fatalf("AESDecryptFromBase64() 失败: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("解密结果不匹配: got %v, want %v", decrypted, plaintext)
	}
}

func TestAESDecrypt_InvalidCiphertext(t *testing.T) {
	key, err := Crypt.GenerateAESKey(AES256)
	if err != nil {
		t.Fatalf("生成AES密钥失败: %v", err)
	}

	tests := []struct {
		name       string
		ciphertext []byte
		expectErr  bool
	}{
		{
			name:       "空密文",
			ciphertext: []byte{},
			expectErr:  true,
		},
		{
			name:       "太短的密文",
			ciphertext: []byte{0x01, 0x02},
			expectErr:  true,
		},
		{
			name:       "无效的密文",
			ciphertext: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c},
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Crypt.AESDecrypt(tt.ciphertext, key)
			if tt.expectErr && err == nil {
				t.Error("AESDecrypt() 预期返回错误，但没有")
			}
		})
	}
}

// =========================================
// RSA 非对称加密测试
// =========================================

func TestGenerateRSAKeys(t *testing.T) {
	tests := []struct {
		name  string
		bits  RSABits
		check bool
	}{
		{
			name:  "RSA1024密钥对",
			bits:  RSA1024,
			check: true,
		},
		{
			name:  "RSA2048密钥对",
			bits:  RSA2048,
			check: true,
		},
		{
			name:  "RSA4096密钥对",
			bits:  RSA4096,
			check: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privateKey, publicKey, err := Crypt.GenerateRSAKeys(tt.bits)
			if err != nil {
				t.Errorf("GenerateRSAKeys() 失败: %v", err)
				return
			}

			if tt.check {
				if privateKey == nil {
					t.Error("GenerateRSAKeys() 私钥为空")
				}
				if publicKey == nil {
					t.Error("GenerateRSAKeys() 公钥为空")
				}
				// 验证密钥大小
				if privateKey.N.BitLen() != int(tt.bits) {
					t.Errorf("GenerateRSAKeys() 密钥大小错误: got %d, want %d", privateKey.N.BitLen(), tt.bits)
				}
			}
		})
	}
}

func TestRSAEncryptDecrypt(t *testing.T) {
	// 生成RSA密钥对
	privateKey, publicKey, err := Crypt.GenerateRSAKeys(RSA2048)
	if err != nil {
		t.Fatalf("生成RSA密钥失败: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "短文本",
			plaintext: "Hello, RSA!",
		},
		{
			name:      "中文字符",
			plaintext: "RSA加密测试",
		},
		{
			name:      "混合字符",
			plaintext: "Test测试123!@#",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 公钥加密
			ciphertext, err := Crypt.RSAEncrypt([]byte(tt.plaintext), publicKey)
			if err != nil {
				t.Errorf("RSAEncrypt() 失败: %v", err)
				return
			}

			// 验证密文与原文不同
			if bytes.Equal(ciphertext, []byte(tt.plaintext)) {
				t.Error("RSAEncrypt() 密文与原文相同")
			}

			// 私钥解密
			decrypted, err := Crypt.RSADecrypt(ciphertext, privateKey)
			if err != nil {
				t.Errorf("RSADecrypt() 失败: %v", err)
				return
			}

			// 验证解密结果
			if string(decrypted) != tt.plaintext {
				t.Errorf("RSADecrypt() 解密结果不匹配: got %v, want %v", string(decrypted), tt.plaintext)
			}
		})
	}
}

func TestRSAEncryptToBase64_RSADecryptFromBase64(t *testing.T) {
	privateKey, publicKey, err := Crypt.GenerateRSAKeys(RSA2048)
	if err != nil {
		t.Fatalf("生成RSA密钥失败: %v", err)
	}

	plaintext := "RSA加密数据"

	// 加密并Base64编码
	encrypted, err := Crypt.RSAEncryptToBase64(plaintext, publicKey)
	if err != nil {
		t.Fatalf("RSAEncryptToBase64() 失败: %v", err)
	}

	// 验证是有效的Base64
	if !Crypt.IsBase64(encrypted) {
		t.Error("RSAEncryptToBase64() 返回无效的Base64")
	}

	// 从Base64解密
	decrypted, err := Crypt.RSADecryptFromBase64(encrypted, privateKey)
	if err != nil {
		t.Fatalf("RSADecryptFromBase64() 失败: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("解密结果不匹配: got %v, want %v", decrypted, plaintext)
	}
}

// =========================================
// 密码哈希测试
// =========================================

func TestBcryptHash(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		cost      int
		expectErr bool
	}{
		{
			name:      "正常密码",
			password:  "mypassword123",
			cost:      bcrypt.DefaultCost,
			expectErr: false,
		},
		{
			name:      "简单密码",
			password:  "123456",
			cost:      10,
			expectErr: false,
		},
		{
			name:      "包含特殊字符",
			password:  "!@#$%^&*()",
			cost:      bcrypt.DefaultCost,
			expectErr: false,
		},
		{
			name:      "空密码",
			password:  "",
			cost:      bcrypt.DefaultCost,
			expectErr: false,
		},
		{
			name:      "中文密码",
			password:  "密码123",
			cost:      bcrypt.DefaultCost,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := Crypt.BcryptHash(tt.password, tt.cost)
			if tt.expectErr {
				if err == nil {
					t.Error("BcryptHash() 预期返回错误，但没有")
				}
			} else {
				if err != nil {
					t.Errorf("BcryptHash() 失败: %v", err)
				}
				// 验证哈希不为空
				if hash == "" {
					t.Error("BcryptHash() 返回空哈希")
				}
				// 验证哈希以$开头（bcrypt标准格式）
				if !strings.HasPrefix(hash, "$") {
					t.Error("BcryptHash() 返回无效的bcrypt哈希格式")
				}
				// 验证哈希与密码不同
				if hash == tt.password {
					t.Error("BcryptHash() 哈希与密码相同")
				}
			}
		})
	}
}

func TestBcryptVerify(t *testing.T) {
	password := "testpassword123"
	hash, err := Crypt.BcryptHash(password, bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("BcryptHash() 失败: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		expected bool
	}{
		{
			name:     "正确密码",
			password: password,
			hash:     hash,
			expected: true,
		},
		{
			name:     "错误密码",
			password: "wrongpassword",
			hash:     hash,
			expected: false,
		},
		{
			name:     "空密码",
			password: "",
			hash:     hash,
			expected: false,
		},
		{
			name:     "无效哈希",
			password: password,
			hash:     "invalidhash",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Crypt.BcryptVerify(tt.password, tt.hash)
			if result != tt.expected {
				t.Errorf("BcryptVerify() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPBKDF2Hash(t *testing.T) {
	tests := []struct {
		name       string
		password   string
		salt       string
		iterations int
		keyLen     int
	}{
		{
			name:       "标准参数",
			password:   "password123",
			salt:       "randomsalt",
			iterations: 10000,
			keyLen:     32,
		},
		{
			name:       "低迭代次数",
			password:   "test",
			salt:       "salt",
			iterations: 1000,
			keyLen:     16,
		},
		{
			name:       "高迭代次数",
			password:   "securepassword",
			salt:       "securesalt",
			iterations: 100000,
			keyLen:     64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := Crypt.PBKDF2Hash(tt.password, tt.salt, tt.iterations, tt.keyLen)
			// 验证哈希不为空
			if hash == "" {
				t.Error("PBKDF2Hash() 返回空哈希")
			}
			// 验证是有效的Base64
			if !Crypt.IsBase64(hash) {
				t.Error("PBKDF2Hash() 返回无效的Base64")
			}
			// 验证哈希与密码不同
			if hash == tt.password {
				t.Error("PBKDF2Hash() 哈希与密码相同")
			}
		})
	}
}

func TestPBKDF2Verify(t *testing.T) {
	password := "testpassword"
	salt := "randomsalt"
	iterations := 10000
	keyLen := 32

	hash := Crypt.PBKDF2Hash(password, salt, iterations, keyLen)

	tests := []struct {
		name       string
		password   string
		salt       string
		hash       string
		iterations int
		keyLen     int
		expected   bool
	}{
		{
			name:       "正确密码和盐",
			password:   password,
			salt:       salt,
			hash:       hash,
			iterations: iterations,
			keyLen:     keyLen,
			expected:   true,
		},
		{
			name:       "错误密码",
			password:   "wrongpassword",
			salt:       salt,
			hash:       hash,
			iterations: iterations,
			keyLen:     keyLen,
			expected:   false,
		},
		{
			name:       "错误盐",
			password:   password,
			salt:       "wrongsalt",
			hash:       hash,
			iterations: iterations,
			keyLen:     keyLen,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Crypt.PBKDF2Verify(tt.password, tt.salt, tt.hash, tt.iterations, tt.keyLen)
			if result != tt.expected {
				t.Errorf("PBKDF2Verify() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateSalt(t *testing.T) {
	tests := []struct {
		name      string
		length    int
		expectErr bool
	}{
		{
			name:      "正常长度",
			length:    16,
			expectErr: false,
		},
		{
			name:      "短盐",
			length:    8,
			expectErr: false,
		},
		{
			name:      "长盐",
			length:    64,
			expectErr: false,
		},
		{
			name:      "无效长度",
			length:    0,
			expectErr: true,
		},
		{
			name:      "负长度",
			length:    -1,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			salt, err := Crypt.GenerateSalt(tt.length)
			if tt.expectErr {
				if err == nil {
					t.Error("GenerateSalt() 预期返回错误，但没有")
				}
			} else {
				if err != nil {
					t.Errorf("GenerateSalt() 失败: %v", err)
				}
				if len(salt) != tt.length {
					t.Errorf("GenerateSalt() 长度错误: got %d, want %d", len(salt), tt.length)
				}
			}
		})
	}
}

func TestGenerateSaltHex(t *testing.T) {
	tests := []struct {
		name      string
		length    int
		expectErr bool
	}{
		{
			name:      "生成16字节盐",
			length:    16,
			expectErr: false,
		},
		{
			name:      "生成32字节盐",
			length:    32,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saltHex, err := Crypt.GenerateSaltHex(tt.length)
			if tt.expectErr {
				if err == nil {
					t.Error("GenerateSaltHex() 预期返回错误，但没有")
				}
			} else {
				if err != nil {
					t.Errorf("GenerateSaltHex() 失败: %v", err)
				}
				// 验证是有效的Hex
				if !Crypt.IsHex(saltHex) {
					t.Error("GenerateSaltHex() 返回无效的Hex")
				}
				// 验证长度（Hex编码是原始长度的2倍）
				expectedLen := tt.length * 2
				if len(saltHex) != expectedLen {
					t.Errorf("GenerateSaltHex() Hex长度错误: got %d, want %d", len(saltHex), expectedLen)
				}
			}
		})
	}
}

// =========================================
// HMAC 签名测试
// =========================================

func TestHMACSign(t *testing.T) {
	data := []byte("test data")
	key := []byte("secret key")

	tests := []struct {
		name     string
		hashFunc func() hash.Hash
	}{
		{
			name:     "SHA256",
			hashFunc: sha256.New,
		},
		{
			name:     "SHA512",
			hashFunc: sha512.New,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature := Crypt.HMACSign(data, key, tt.hashFunc)
			// 验证签名不为空
			if len(signature) == 0 {
				t.Error("HMACSign() 返回空签名")
			}
			// 验证签名长度（SHA256=32, SHA512=64）
			expectedSize := tt.hashFunc().Size()
			if len(signature) != expectedSize {
				t.Errorf("HMACSign() 签名长度错误: got %d, want %d", len(signature), expectedSize)
			}
		})
	}
}

func TestHMACSignHex(t *testing.T) {
	data := []byte("test data")
	key := []byte("secret key")

	tests := []struct {
		name     string
		hashFunc func() hash.Hash
	}{
		{
			name:     "SHA256 Hex",
			hashFunc: sha256.New,
		},
		{
			name:     "SHA512 Hex",
			hashFunc: sha512.New,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signatureHex := Crypt.HMACSignHex(data, key, tt.hashFunc)
			// 验证是有效的Hex
			if !Crypt.IsHex(signatureHex) {
				t.Error("HMACSignHex() 返回无效的Hex")
			}
			// 验证长度（Hex编码是原始长度的2倍）
			expectedSize := tt.hashFunc().Size() * 2
			if len(signatureHex) != expectedSize {
				t.Errorf("HMACSignHex() Hex长度错误: got %d, want %d", len(signatureHex), expectedSize)
			}
		})
	}
}

func TestHMACSignBase64(t *testing.T) {
	data := []byte("test data")
	key := []byte("secret key")

	signatureBase64 := Crypt.HMACSignBase64(data, key, sha256.New)

	// 验证是有效的Base64
	if !Crypt.IsBase64(signatureBase64) {
		t.Error("HMACSignBase64() 返回无效的Base64")
	}

	// 验证可以解码
	decoded, err := Crypt.Base64DecodeBytes(signatureBase64)
	if err != nil {
		t.Errorf("HMACSignBase64() 无法解码: %v", err)
	}
	if len(decoded) == 0 {
		t.Error("HMACSignBase64() 解码后为空")
	}
}

func TestHMACVerify(t *testing.T) {
	data := []byte("test data")
	key := []byte("secret key")

	tests := []struct {
		name     string
		hashFunc func() hash.Hash
	}{
		{
			name:     "SHA256验证",
			hashFunc: sha256.New,
		},
		{
			name:     "SHA512验证",
			hashFunc: sha512.New,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 生成签名
			signature := Crypt.HMACSign(data, key, tt.hashFunc)

			// 验证正确签名
			if !Crypt.HMACVerify(data, key, signature, tt.hashFunc) {
				t.Error("HMACVerify() 未能验证正确的签名")
			}

			// 验证错误签名
			wrongSignature := make([]byte, len(signature))
			copy(wrongSignature, signature)
			wrongSignature[0] ^= 0xFF // 翻转第一个字节
			if Crypt.HMACVerify(data, key, wrongSignature, tt.hashFunc) {
				t.Error("HMACVerify() 错误地验证了错误的签名")
			}

			// 验证错误数据
			wrongData := []byte("wrong data")
			if Crypt.HMACVerify(wrongData, key, signature, tt.hashFunc) {
				t.Error("HMACVerify() 错误地验证了错误的数据")
			}

			// 验证错误密钥
			wrongKey := []byte("wrong key")
			if Crypt.HMACVerify(data, wrongKey, signature, tt.hashFunc) {
				t.Error("HMACVerify() 错误地验证了错误的密钥")
			}
		})
	}
}

func TestHMACSignWithSHA256(t *testing.T) {
	data := []byte("test data")
	key := []byte("secret key")

	signature := Crypt.HMACSignWithSHA256(data, key)

	// 验证签名长度（SHA256=32字节）
	if len(signature) != 32 {
		t.Errorf("HMACSignWithSHA256() 签名长度错误: got %d, want 32", len(signature))
	}

	// 验证与通用方法结果一致
	expectedSig := Crypt.HMACSign(data, key, sha256.New)
	if !bytes.Equal(signature, expectedSig) {
		t.Error("HMACSignWithSHA256() 与通用方法结果不一致")
	}
}

func TestHMACSignHexWithSHA256(t *testing.T) {
	data := []byte("test data")
	key := []byte("secret key")

	signatureHex := Crypt.HMACSignHexWithSHA256(data, key)

	// 验证是有效的Hex
	if !Crypt.IsHex(signatureHex) {
		t.Error("HMACSignHexWithSHA256() 返回无效的Hex")
	}

	// 验证与通用方法结果一致
	expectedSig := Crypt.HMACSignHex(data, key, sha256.New)
	if signatureHex != expectedSig {
		t.Error("HMACSignHexWithSHA256() 与通用方法结果不一致")
	}
}

func TestHMACSignWithSHA512(t *testing.T) {
	data := []byte("test data")
	key := []byte("secret key")

	signature := Crypt.HMACSignWithSHA512(data, key)

	// 验证签名长度（SHA512=64字节）
	if len(signature) != 64 {
		t.Errorf("HMACSignWithSHA512() 签名长度错误: got %d, want 64", len(signature))
	}

	// 验证与通用方法结果一致
	expectedSig := Crypt.HMACSign(data, key, sha512.New)
	if !bytes.Equal(signature, expectedSig) {
		t.Error("HMACSignWithSHA512() 与通用方法结果不一致")
	}
}

func TestHMACSignHexWithSHA512(t *testing.T) {
	data := []byte("test data")
	key := []byte("secret key")

	signatureHex := Crypt.HMACSignHexWithSHA512(data, key)

	// 验证是有效的Hex
	if !Crypt.IsHex(signatureHex) {
		t.Error("HMACSignHexWithSHA512() 返回无效的Hex")
	}

	// 验证与通用方法结果一致
	expectedSig := Crypt.HMACSignHex(data, key, sha512.New)
	if signatureHex != expectedSig {
		t.Error("HMACSignHexWithSHA512() 与通用方法结果不一致")
	}
}

// =========================================
// ECDSA 数字签名测试
// =========================================

func TestGenerateECDSAKeys(t *testing.T) {
	privateKey, publicKey, err := Crypt.GenerateECDSAKeys()

	if err != nil {
		t.Fatalf("GenerateECDSAKeys() 失败: %v", err)
	}

	if privateKey == nil {
		t.Error("GenerateECDSAKeys() 私钥为空")
	}

	if publicKey == nil {
		t.Error("GenerateECDSAKeys() 公钥为空")
	}

	// 验证密钥对
	if !reflect.DeepEqual(privateKey.PublicKey, *publicKey) {
		t.Error("GenerateECDSAKeys() 公钥不匹配")
	}
}

func TestECDSASign(t *testing.T) {
	privateKey, _, err := Crypt.GenerateECDSAKeys()
	if err != nil {
		t.Fatalf("GenerateECDSAKeys() 失败: %v", err)
	}

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "简单数据",
			data: []byte("test data"),
		},
		{
			name: "空数据",
			data: []byte{},
		},
		{
			name: "长数据",
			data: []byte(strings.Repeat("A", 1000)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature, err := Crypt.ECDSASign(tt.data, privateKey)
			if err != nil {
				t.Errorf("ECDSASign() 失败: %v", err)
				return
			}

			// 验证签名不为空
			if len(signature) == 0 {
				t.Error("ECDSASign() 返回空签名")
			}
		})
	}
}

func TestECDSASignHex(t *testing.T) {
	privateKey, _, err := Crypt.GenerateECDSAKeys()
	if err != nil {
		t.Fatalf("GenerateECDSAKeys() 失败: %v", err)
	}

	data := []byte("test data")

	signatureHex, err := Crypt.ECDSASignHex(data, privateKey)
	if err != nil {
		t.Fatalf("ECDSASignHex() 失败: %v", err)
	}

	// 验证是有效的Hex
	if !Crypt.IsHex(signatureHex) {
		t.Error("ECDSASignHex() 返回无效的Hex")
	}

	// 验证可以解码
	_, err = Crypt.HexDecodeBytes(signatureHex)
	if err != nil {
		t.Errorf("ECDSASignHex() 无法解码: %v", err)
	}
}

func TestECDSAVerify(t *testing.T) {
	privateKey, publicKey, err := Crypt.GenerateECDSAKeys()
	if err != nil {
		t.Fatalf("GenerateECDSAKeys() 失败: %v", err)
	}

	data := []byte("test data")

	// 生成签名
	signature, err := Crypt.ECDSASign(data, privateKey)
	if err != nil {
		t.Fatalf("ECDSASign() 失败: %v", err)
	}

	tests := []struct {
		name      string
		data      []byte
		signature []byte
		expected  bool
	}{
		{
			name:      "正确数据和签名",
			data:      data,
			signature: signature,
			expected:  true,
		},
		{
			name:      "错误数据",
			data:      []byte("wrong data"),
			signature: signature,
			expected:  false,
		},
		{
			name:      "错误签名",
			data:      data,
			signature: []byte("wrong signature"),
			expected:  false,
		},
		{
			name:      "空签名",
			data:      data,
			signature: []byte{},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Crypt.ECDSAVerify(tt.data, tt.signature, publicKey)
			if result != tt.expected {
				t.Errorf("ECDSAVerify() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestECDSAVerifyHex(t *testing.T) {
	privateKey, publicKey, err := Crypt.GenerateECDSAKeys()
	if err != nil {
		t.Fatalf("GenerateECDSAKeys() 失败: %v", err)
	}

	data := []byte("test data")

	// 生成签名
	signatureHex, err := Crypt.ECDSASignHex(data, privateKey)
	if err != nil {
		t.Fatalf("ECDSASignHex() 失败: %v", err)
	}

	// 验证正确的签名
	result, err := Crypt.ECDSAVerifyHex(data, signatureHex, publicKey)
	if err != nil {
		t.Errorf("ECDSAVerifyHex() 意外错误: %v", err)
	}
	if !result {
		t.Error("ECDSAVerifyHex() 未能验证正确的签名")
	}

	// 验证无效的Hex
	_, err = Crypt.ECDSAVerifyHex(data, "invalid-hex", publicKey)
	if err == nil {
		t.Error("ECDSAVerifyHex() 预期返回错误，但没有")
	}

	// 验证错误的签名
	wrongSig := "abcd1234"
	result, _ = Crypt.ECDSAVerifyHex(data, wrongSig, publicKey)
	if result {
		t.Error("ECDSAVerifyHex() 错误地验证了错误的签名")
	}
}

// =========================================
// 密钥格式转换测试
// =========================================

func TestPrivateKeyToPEM(t *testing.T) {
	privateKey, _, err := Crypt.GenerateRSAKeys(RSA2048)
	if err != nil {
		t.Fatalf("GenerateRSAKeys() 失败: %v", err)
	}

	pem := Crypt.PrivateKeyToPEM(privateKey)

	// 验证PEM不为空
	if pem == "" {
		t.Error("PrivateKeyToPEM() 返回空字符串")
	}

	// 验证包含关键信息
	if !strings.Contains(pem, "RSA Private Key") {
		t.Error("PrivateKeyToPEM() 不包含预期内容")
	}
}

func TestPublicKeyToPEM(t *testing.T) {
	_, publicKey, err := Crypt.GenerateRSAKeys(RSA2048)
	if err != nil {
		t.Fatalf("GenerateRSAKeys() 失败: %v", err)
	}

	pem := Crypt.PublicKeyToPEM(publicKey)

	// 验证PEM不为空
	if pem == "" {
		t.Error("PublicKeyToPEM() 返回空字符串")
	}

	// 验证包含关键信息
	if !strings.Contains(pem, "RSA Public Key") {
		t.Error("PublicKeyToPEM() 不包含预期内容")
	}
}

// =========================================
// 工具函数测试
// =========================================

func TestConstantTimeCompare(t *testing.T) {
	tests := []struct {
		name     string
		a        []byte
		b        []byte
		expected bool
	}{
		{
			name:     "相同数据",
			a:        []byte("hello"),
			b:        []byte("hello"),
			expected: true,
		},
		{
			name:     "不同数据",
			a:        []byte("hello"),
			b:        []byte("world"),
			expected: false,
		},
		{
			name:     "不同长度",
			a:        []byte("hello"),
			b:        []byte("hello!"),
			expected: false,
		},
		{
			name:     "空数据",
			a:        []byte{},
			b:        []byte{},
			expected: true,
		},
		{
			name:     "一空一非空",
			a:        []byte{},
			b:        []byte("hello"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Crypt.ConstantTimeCompare(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("ConstantTimeCompare() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSecureEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{
			name:     "相同字符串",
			a:        "hello",
			b:        "hello",
			expected: true,
		},
		{
			name:     "不同字符串",
			a:        "hello",
			b:        "world",
			expected: false,
		},
		{
			name:     "空字符串",
			a:        "",
			b:        "",
			expected: true,
		},
		{
			name:     "一空一非空",
			a:        "",
			b:        "hello",
			expected: false,
		},
		{
			name:     "包含特殊字符",
			a:        "!@#$%",
			b:        "!@#$%",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Crypt.SecureEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("SecureEqual() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	tests := []struct {
		name      string
		length    int
		expectErr bool
	}{
		{
			name:      "正常长度",
			length:    16,
			expectErr: false,
		},
		{
			name:      "大长度",
			length:    1024,
			expectErr: false,
		},
		{
			name:      "无效长度",
			length:    0,
			expectErr: true,
		},
		{
			name:      "负长度",
			length:    -1,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Crypt.GenerateRandomBytes(tt.length)
			if tt.expectErr {
				if err == nil {
					t.Error("GenerateRandomBytes() 预期返回错误，但没有")
				}
			} else {
				if err != nil {
					t.Errorf("GenerateRandomBytes() 失败: %v", err)
				}
				if len(result) != tt.length {
					t.Errorf("GenerateRandomBytes() 长度错误: got %d, want %d", len(result), tt.length)
				}
			}
		})
	}

	// 测试随机性：生成多次，确保结果不同
	t.Run("随机性测试", func(t *testing.T) {
		results := make([][]byte, 10)
		for i := range results {
			var err error
			results[i], err = Crypt.GenerateRandomBytes(32)
			if err != nil {
				t.Fatalf("GenerateRandomBytes() 失败: %v", err)
			}
		}

		// 验证至少有一些结果不同
		allSame := true
		for i := 1; i < len(results); i++ {
			if !bytes.Equal(results[0], results[i]) {
				allSame = false
				break
			}
		}
		if allSame {
			t.Error("GenerateRandomBytes() 生成结果全部相同，随机性不足")
		}
	})
}

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name      string
		length    int
		expectErr bool
	}{
		{
			name:      "短字符串",
			length:    8,
			expectErr: false,
		},
		{
			name:      "长字符串",
			length:    64,
			expectErr: false,
		},
		{
			name:      "无效长度",
			length:    0,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Crypt.GenerateRandomString(tt.length)
			if tt.expectErr {
				if err == nil {
					t.Error("GenerateRandomString() 预期返回错误，但没有")
				}
			} else {
				if err != nil {
					t.Errorf("GenerateRandomString() 失败: %v", err)
				}
				if len(result) != tt.length {
					t.Errorf("GenerateRandomString() 长度错误: got %d, want %d", len(result), tt.length)
				}

				// 验证只包含允许的字符
				const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
				for _, c := range result {
					if !strings.ContainsRune(charset, c) {
						t.Errorf("GenerateRandomString() 包含非法字符: %c", c)
					}
				}
			}
		})
	}

	// 测试随机性
	t.Run("随机性测试", func(t *testing.T) {
		results := make([]string, 10)
		for i := range results {
			var err error
			results[i], err = Crypt.GenerateRandomString(16)
			if err != nil {
				t.Fatalf("GenerateRandomString() 失败: %v", err)
			}
		}

		// 验证至少有一些结果不同
		allSame := true
		for i := 1; i < len(results); i++ {
			if results[0] != results[i] {
				allSame = false
				break
			}
		}
		if allSame {
			t.Error("GenerateRandomString() 生成结果全部相同，随机性不足")
		}
	})
}

func TestIsBase64(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "有效Base64",
			input:    "SGVsbG8sIFdvcmxkIQ==",
			expected: true,
		},
		{
			name:     "无效Base64",
			input:    "Hello!@#",
			expected: false,
		},
		{
			name:     "空字符串",
			input:    "",
			expected: true,
		},
		{
			name:     "包含特殊字符",
			input:    "Invalid@Base64",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Crypt.IsBase64(tt.input)
			if result != tt.expected {
				t.Errorf("IsBase64() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsHex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "有效Hex（小写）",
			input:    "48656c6c6f",
			expected: true,
		},
		{
			name:     "有效Hex（大写）",
			input:    "48654C4C6F",
			expected: true,
		},
		{
			name:     "无效Hex",
			input:    "hello",
			expected: false,
		},
		{
			name:     "空字符串",
			input:    "",
			expected: true,
		},
		{
			name:     "包含非法字符",
			input:    "48656g6c6f",
			expected: false,
		},
		{
			name:     "奇数长度",
			input:    "abc",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Crypt.IsHex(tt.input)
			if result != tt.expected {
				t.Errorf("IsHex() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEncodeKey(t *testing.T) {
	key := []byte{0x01, 0x02, 0x03, 0x04, 0x05}

	encoded := Crypt.EncodeKey(key)

	// 验证是有效的Base64
	if !Crypt.IsBase64(encoded) {
		t.Error("EncodeKey() 返回无效的Base64")
	}

	// 验证可以解码
	decoded, err := Crypt.DecodeKey(encoded)
	if err != nil {
		t.Errorf("DecodeKey() 无法解码: %v", err)
	}
	if !bytes.Equal(decoded, key) {
		t.Error("EncodeKey/DecodeKey 编码解码不匹配")
	}
}

func TestDecodeKey(t *testing.T) {
	tests := []struct {
		name       string
		encodedKey string
		expectErr  bool
	}{
		{
			name:       "有效编码密钥",
			encodedKey: "AQIDBAU=",
			expectErr:  false,
		},
		{
			name:       "无效编码密钥",
			encodedKey: "Invalid@Key",
			expectErr:  true,
		},
		{
			name:       "空字符串",
			encodedKey: "",
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Crypt.DecodeKey(tt.encodedKey)
			if tt.expectErr {
				if err == nil {
					t.Error("DecodeKey() 预期返回错误，但没有")
				}
			} else {
				if err != nil {
					t.Errorf("DecodeKey() 意外错误: %v", err)
				}
				if tt.encodedKey != "" && len(result) == 0 {
					t.Error("DecodeKey() 返回空结果")
				}
			}
		})
	}
}

// =========================================
// 集成测试
// =========================================

func TestCryptIntegration_AES_RSASign(t *testing.T) {
	// 1. AES加密解密
	aesKey, _ := Crypt.GenerateAESKey(AES256)
	plaintext := "敏感信息"

	encrypted, err := Crypt.AESEncryptToBase64(plaintext, aesKey)
	if err != nil {
		t.Fatalf("AES加密失败: %v", err)
	}

	decrypted, err := Crypt.AESDecryptFromBase64(encrypted, aesKey)
	if err != nil {
		t.Fatalf("AES解密失败: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("AES加解密结果不匹配: got %v, want %v", decrypted, plaintext)
	}

	// 2. RSA加密解密
	privateKey, publicKey, _ := Crypt.GenerateRSAKeys(RSA2048)

	rsaEncrypted, err := Crypt.RSAEncryptToBase64(plaintext, publicKey)
	if err != nil {
		t.Fatalf("RSA加密失败: %v", err)
	}

	rsaDecrypted, err := Crypt.RSADecryptFromBase64(rsaEncrypted, privateKey)
	if err != nil {
		t.Fatalf("RSA解密失败: %v", err)
	}

	if rsaDecrypted != plaintext {
		t.Errorf("RSA加解密结果不匹配: got %v, want %v", rsaDecrypted, plaintext)
	}

	// 3. ECDSA签名验证
	ecdsaPrivateKey, ecdsaPublicKey, _ := Crypt.GenerateECDSAKeys()
	data := []byte(plaintext)

	signature, err := Crypt.ECDSASignHex(data, ecdsaPrivateKey)
	if err != nil {
		t.Fatalf("ECDSA签名失败: %v", err)
	}

	verified, err := Crypt.ECDSAVerifyHex(data, signature, ecdsaPublicKey)
	if err != nil {
		t.Fatalf("ECDSA验证失败: %v", err)
	}

	if !verified {
		t.Error("ECDSA签名验证失败")
	}

	// 4. HMAC签名验证
	hmacKey := []byte("hmac-key")
	hmacSig := Crypt.HMACSignHexWithSHA256(data, hmacKey)
	hmacSigBytes, _ := Crypt.HexDecodeBytes(hmacSig)
	if !Crypt.HMACVerify(data, hmacKey, hmacSigBytes, sha256.New) {
		t.Error("HMAC签名验证失败")
	}

	// 5. 密码哈希验证
	password := "testpassword123"
	bcryptHash, _ := Crypt.BcryptHash(password, bcrypt.DefaultCost)
	if !Crypt.BcryptVerify(password, bcryptHash) {
		t.Error("Bcrypt验证失败")
	}

	salt, _ := Crypt.GenerateSaltHex(16)
	pbkdf2Hash := Crypt.PBKDF2Hash(password, salt, 10000, 32)
	if !Crypt.PBKDF2Verify(password, salt, pbkdf2Hash, 10000, 32) {
		t.Error("PBKDF2验证失败")
	}
}

// 基准测试
func BenchmarkBcryptHash(b *testing.B) {
	password := "testpassword123"
	for i := 0; i < b.N; i++ {
		Crypt.BcryptHash(password, bcrypt.DefaultCost)
	}
}

func BenchmarkAESEncrypt(b *testing.B) {
	key, _ := Crypt.GenerateAESKey(AES256)
	plaintext := []byte("This is a test message for benchmarking AES encryption")
	for i := 0; i < b.N; i++ {
		Crypt.AESEncrypt(plaintext, key)
	}
}

func BenchmarkHMACSign(b *testing.B) {
	data := []byte("test data")
	key := []byte("secret key")
	for i := 0; i < b.N; i++ {
		Crypt.HMACSignWithSHA256(data, key)
	}
}
