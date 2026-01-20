package hash

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"io"
)

// HashOutputFormat 哈希输出格式枚举
type HashOutputFormat int

const (
	// FormatBytes 原始字节数组格式
	FormatBytes HashOutputFormat = iota
	// FormatHexShort 16位十六进制字符串（通常用于MD5短格式）
	FormatHexShort
	// FormatHexMedium 32位十六进制字符串（通常用于MD5、SHA256短格式）
	FormatHexMedium
	// FormatHexFull 完整长度十六进制字符串
	FormatHexFull
)

// hashEngine 哈希操作引擎（内部实现）
type hashEngine struct{}

// Hash 默认的哈希操作实例
var Hash = &hashEngine{}

// HashAlgorithm 哈希算法接口
type HashAlgorithm interface {
	Hash() hash.Hash
}

// MD5Algorithm MD5算法实现
type MD5Algorithm struct{}

func (MD5Algorithm) Hash() hash.Hash { return md5.New() }

// SHA1Algorithm SHA1算法实现
type SHA1Algorithm struct{}

func (SHA1Algorithm) Hash() hash.Hash { return sha1.New() }

// SHA256Algorithm SHA256算法实现
type SHA256Algorithm struct{}

func (SHA256Algorithm) Hash() hash.Hash { return sha256.New() }

// SHA512Algorithm SHA512算法实现
type SHA512Algorithm struct{}

func (SHA512Algorithm) Hash() hash.Hash { return sha512.New() }

// =========================================
// 核心工具函数
// =========================================

// formatHash 根据指定格式格式化哈希值
func formatHash(hashBytes []byte, format HashOutputFormat) string {
	hexStr := hex.EncodeToString(hashBytes)

	switch format {
	case FormatBytes:
		return string(hashBytes)
	case FormatHexShort:
		if len(hexStr) >= 16 {
			return hexStr[:16]
		}
		return hexStr
	case FormatHexMedium:
		if len(hexStr) >= 32 {
			return hexStr[:32]
		}
		return hexStr
	case FormatHexFull:
		return hexStr
	default:
		return hexStr
	}
}

// =========================================
// 泛型哈希函数
// =========================================

// HashGeneric 计算任意哈希算法的值
func HashGeneric[T HashAlgorithm](data string, algorithm T) []byte {
	hasher := algorithm.Hash()
	hasher.Write([]byte(data))
	return hasher.Sum(nil)
}

// HashBytesGeneric 计算字节数组的哈希值
func HashBytesGeneric[T HashAlgorithm](data []byte, algorithm T) []byte {
	hasher := algorithm.Hash()
	hasher.Write(data)
	return hasher.Sum(nil)
}

// HashReaderGeneric 从io.Reader计算哈希值
func HashReaderGeneric[T HashAlgorithm](r io.Reader, algorithm T) ([]byte, error) {
	hasher := algorithm.Hash()
	_, err := io.Copy(hasher, r)
	if err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}

// HashHexGeneric 计算哈希值并返回指定格式的十六进制字符串
func HashHexGeneric[T HashAlgorithm](data string, algorithm T, format HashOutputFormat) string {
	hashBytes := HashGeneric(data, algorithm)
	return formatHash(hashBytes, format)
}

// HashBytesHexGeneric 计算字节数组的哈希值并返回指定格式的十六进制字符串
func HashBytesHexGeneric[T HashAlgorithm](data []byte, algorithm T, format HashOutputFormat) string {
	hashBytes := HashBytesGeneric(data, algorithm)
	return formatHash(hashBytes, format)
}

// HashReaderHexGeneric 从io.Reader计算哈希值并返回指定格式的十六进制字符串
func HashReaderHexGeneric[T HashAlgorithm](r io.Reader, algorithm T, format HashOutputFormat) (string, error) {
	hashBytes, err := HashReaderGeneric(r, algorithm)
	if err != nil {
		return "", err
	}
	return formatHash(hashBytes, format), nil
}

// HashStringGeneric 计算哈希并返回完整十六进制字符串
func HashStringGeneric[T HashAlgorithm](data string, algorithm T) string {
	return HashHexGeneric(data, algorithm, FormatHexFull)
}

// HashReaderStringGeneric 从io.Reader计算哈希并返回完整十六进制字符串
func HashReaderStringGeneric[T HashAlgorithm](r io.Reader, algorithm T) (string, error) {
	return HashReaderHexGeneric(r, algorithm, FormatHexFull)
}

// =========================================
// 泛型HMAC函数
// =========================================

// HMACGeneric 计算HMAC哈希值
func HMACGeneric[T HashAlgorithm](data string, key string, algorithm T) []byte {
	hasher := hmac.New(algorithm.Hash, []byte(key))
	hasher.Write([]byte(data))
	return hasher.Sum(nil)
}

// HMACBytesGeneric 计算字节数组的HMAC哈希值
func HMACBytesGeneric[T HashAlgorithm](data []byte, key []byte, algorithm T) []byte {
	hasher := hmac.New(algorithm.Hash, key)
	hasher.Write(data)
	return hasher.Sum(nil)
}

// HMACReaderGeneric 从io.Reader计算HMAC哈希值
func HMACReaderGeneric[T HashAlgorithm](r io.Reader, key []byte, algorithm T) ([]byte, error) {
	hasher := hmac.New(algorithm.Hash, key)
	_, err := io.Copy(hasher, r)
	if err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}

// HMACHexGeneric 计算HMAC哈希值并返回指定格式的十六进制字符串
func HMACHexGeneric[T HashAlgorithm](data string, key string, algorithm T, format HashOutputFormat) string {
	hashBytes := HMACGeneric(data, key, algorithm)
	return formatHash(hashBytes, format)
}

// HMACBytesHexGeneric 计算字节数组的HMAC哈希值并返回指定格式的十六进制字符串
func HMACBytesHexGeneric[T HashAlgorithm](data []byte, key []byte, algorithm T, format HashOutputFormat) string {
	hashBytes := HMACBytesGeneric(data, key, algorithm)
	return formatHash(hashBytes, format)
}

// HMACReaderHexGeneric 从io.Reader计算HMAC哈希值并返回指定格式的十六进制字符串
func HMACReaderHexGeneric[T HashAlgorithm](r io.Reader, key []byte, algorithm T,
	format HashOutputFormat) (string, error) {
	hashBytes, err := HMACReaderGeneric(r, key, algorithm)
	if err != nil {
		return "", err
	}
	return formatHash(hashBytes, format), nil
}

// HMACStringGeneric 计算HMAC并返回完整十六进制字符串
func HMACStringGeneric[T HashAlgorithm](data string, key string, algorithm T) string {
	return HMACHexGeneric(data, key, algorithm, FormatHexFull)
}

// HMACReaderStringGeneric 从io.Reader计算HMAC并返回完整十六进制字符串
func HMACReaderStringGeneric[T HashAlgorithm](r io.Reader, key []byte, algorithm T) (string, error) {
	return HMACReaderHexGeneric(r, key, algorithm, FormatHexFull)
}

// =========================================
// 便捷方法 - MD5
// =========================================

// MD5 计算MD5哈希值
func (h *hashEngine) MD5(data string) []byte {
	return HashGeneric(data, MD5Algorithm{})
}

// MD5String 计算MD5并返回完整十六进制字符串
func (h *hashEngine) MD5String(data string) string {
	return HashStringGeneric(data, MD5Algorithm{})
}

// MD5String16 计算MD5并返回16位十六进制字符串
func (h *hashEngine) MD5String16(data string) string {
	return HashHexGeneric(data, MD5Algorithm{}, FormatHexShort)
}

// MD5String32 计算MD5并返回32位十六进制字符串
func (h *hashEngine) MD5String32(data string) string {
	return HashHexGeneric(data, MD5Algorithm{}, FormatHexMedium)
}

// =========================================
// 便捷方法 - SHA1
// =========================================

// SHA1 计算SHA1哈希值
func (h *hashEngine) SHA1(data string) []byte {
	return HashGeneric(data, SHA1Algorithm{})
}

// SHA1String 计算SHA1并返回完整十六进制字符串
func (h *hashEngine) SHA1String(data string) string {
	return HashStringGeneric(data, SHA1Algorithm{})
}

// =========================================
// 便捷方法 - SHA256
// =========================================

// SHA256 计算SHA256哈希值
func (h *hashEngine) SHA256(data string) []byte {
	return HashGeneric(data, SHA256Algorithm{})
}

// SHA256String 计算SHA256并返回完整十六进制字符串
func (h *hashEngine) SHA256String(data string) string {
	return HashStringGeneric(data, SHA256Algorithm{})
}

// =========================================
// 便捷方法 - SHA512
// =========================================

// SHA512 计算SHA512哈希值
func (h *hashEngine) SHA512(data string) []byte {
	return HashGeneric(data, SHA512Algorithm{})
}

// SHA512String 计算SHA512并返回完整十六进制字符串
func (h *hashEngine) SHA512String(data string) string {
	return HashStringGeneric(data, SHA512Algorithm{})
}

// =========================================
// 便捷方法 - HMAC
// =========================================

// HMACMD5 计算HMAC-MD5值
func (h *hashEngine) HMACMD5(data string, key string) []byte {
	return HMACGeneric(data, key, MD5Algorithm{})
}

// HMACMD5String 计算HMAC-MD5并返回完整十六进制字符串
func (h *hashEngine) HMACMD5String(data string, key string) string {
	return HMACStringGeneric(data, key, MD5Algorithm{})
}

// HMACSHA1 计算HMAC-SHA1值
func (h *hashEngine) HMACSHA1(data string, key string) []byte {
	return HMACGeneric(data, key, SHA1Algorithm{})
}

// HMACSHA1String 计算HMAC-SHA1并返回完整十六进制字符串
func (h *hashEngine) HMACSHA1String(data string, key string) string {
	return HMACStringGeneric(data, key, SHA1Algorithm{})
}

// HMACSHA256 计算HMAC-SHA256值
func (h *hashEngine) HMACSHA256(data string, key string) []byte {
	return HMACGeneric(data, key, SHA256Algorithm{})
}

// HMACSHA256String 计算HMAC-SHA256并返回完整十六进制字符串
func (h *hashEngine) HMACSHA256String(data string, key string) string {
	return HMACStringGeneric(data, key, SHA256Algorithm{})
}

// HMACSHA512 计算HMAC-SHA512值
func (h *hashEngine) HMACSHA512(data string, key string) []byte {
	return HMACGeneric(data, key, SHA512Algorithm{})
}

// HMACSHA512String 计算HMAC-SHA512并返回完整十六进制字符串
func (h *hashEngine) HMACSHA512String(data string, key string) string {
	return HMACStringGeneric(data, key, SHA512Algorithm{})
}
