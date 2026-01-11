package rand

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// ILiteUtilRand 随机数工具接口
type ILiteUtilRand interface {
	// 随机数生成
	RandomInt(min, max int) int
	RandomInt64(min, max int64) int64
	RandomFloat(min, max float64) float64
	RandomBool() bool

	// 随机字符串生成
	RandomStringFromCharset(length int, charset string) string
	RandomString(length int) string
	RandomLetters(length int) string
	RandomDigits(length int) string
	RandomLowercase(length int) string
	RandomUppercase(length int) string

	// UUID 生成
	RandomUUID() string
}

// randEngine 随机操作工具类
type randEngine struct{}

// 预定义字符集常量
const (
	CharsetAlphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	CharsetLetters      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharsetDigits       = "0123456789"
	CharsetLowercase    = "abcdefghijklmnopqrstuvwxyz"
	CharsetUppercase    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// Rand 默认随机操作实例（私有）
var defaultRandOp = &randEngine{}

// Default 获取默认的随机数操作实例（单例模式）
// Deprecated: 请使用 liteutil.LiteUtil().Rand() 来获取随机数工具实例

// New 创建新的随机数操作实例
// Deprecated: 请使用 liteutil.LiteUtil().NewRandOperation() 来创建新的随机数工具实例
func newRandEngine() ILiteUtilRand {
	return &randEngine{}
}

// RandomInt 生成指定范围内的随机整数 [min, max]
func (r *randEngine) RandomInt(min, max int) int {
	if min > max {
		min, max = max, min
	}
	if min == max {
		return min
	}

	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		// 如果加密随机数失败，回退到简单的伪随机数
		return min + int(float64(max-min)*0.5) // 简单的回退实现
	}
	return min + int(nBig.Int64())
}

// RandomInt64 生成指定范围内的随机int64整数 [min, max]
func (r *randEngine) RandomInt64(min, max int64) int64 {
	if min > max {
		min, max = max, min
	}
	if min == max {
		return min
	}

	nBig, err := rand.Int(rand.Reader, big.NewInt(max-min+1))
	if err != nil {
		return min
	}
	return min + nBig.Int64()
}

// RandomFloat 生成指定范围内的随机浮点数 [min, max)
func (r *randEngine) RandomFloat(min, max float64) float64 {
	if min > max {
		min, max = max, min
	}
	if min == max {
		return min
	}

	nBig, err := rand.Int(rand.Reader, big.NewInt(1<<53))
	if err != nil {
		return min
	}

	// 使用 [0, 1) 范围的随机数
	random := float64(nBig.Int64()) / float64(1<<53)
	return min + random*(max-min)
}

// RandomBool 生成随机布尔值
func (r *randEngine) RandomBool() bool {
	nBig, err := rand.Int(rand.Reader, big.NewInt(2))
	if err != nil {
		// 回退方案
		return r.RandomInt(0, 1) == 1
	}
	return nBig.Int64() == 1
}

// RandomStringFromCharset 从指定字符集生成随机字符串
func (r *randEngine) RandomStringFromCharset(length int, charset string) string {
	if length <= 0 || charset == "" {
		return ""
	}

	result := make([]byte, length)
	charsetSize := big.NewInt(int64(len(charset)))

	for i := range result {
		nBig, err := rand.Int(rand.Reader, charsetSize)
		if err != nil {
			// 回退方案
			result[i] = charset[i%len(charset)]
			continue
		}
		result[i] = charset[nBig.Int64()]
	}

	return string(result)
}

// RandomString 生成指定长度的随机字母数字字符串
func (r *randEngine) RandomString(length int) string {
	return r.RandomStringFromCharset(length, CharsetAlphanumeric)
}

// RandomLetters 生成指定长度的随机字母字符串
func (r *randEngine) RandomLetters(length int) string {
	return r.RandomStringFromCharset(length, CharsetLetters)
}

// RandomDigits 生成指定长度的随机数字字符串
func (r *randEngine) RandomDigits(length int) string {
	return r.RandomStringFromCharset(length, CharsetDigits)
}

// RandomLowercase 生成指定长度的随机小写字母字符串
func (r *randEngine) RandomLowercase(length int) string {
	return r.RandomStringFromCharset(length, CharsetLowercase)
}

// RandomUppercase 生成指定长度的随机大写字母字符串
func (r *randEngine) RandomUppercase(length int) string {
	return r.RandomStringFromCharset(length, CharsetUppercase)
}

// RandomUUID 生成简单的UUID格式字符串 (版本4)
func (r *randEngine) RandomUUID() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		// 回退方案：生成一个伪随机UUID
		for i := range bytes {
			bytes[i] = byte(r.RandomInt(0, 255))
		}
	}

	// 设置版本号和变体
	bytes[6] = (bytes[6] & 0x0f) | 0x40 // Version 4
	bytes[8] = (bytes[8] & 0x3f) | 0x80 // Variant 10

	return fmt.Sprintf("%x-%x-%x-%x-%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}

// RandomChoice 泛型函数：从给定的选项中随机选择一个
func RandomChoice[T any](options []T) T {
	var zero T
	if len(options) == 0 {
		return zero
	}

	randOp := defaultRandOp
	index := randOp.RandomInt(0, len(options)-1)
	return options[index]
}

// RandomChoices 泛型函数：从给定的选项中随机选择指定数量的元素（不重复）
func RandomChoices[T any](options []T, count int) []T {
	if len(options) == 0 || count <= 0 {
		return nil
	}

	randOp := defaultRandOp

	if count >= len(options) {
		// 复制整个切片并打乱顺序
		result := make([]T, len(options))
		copy(result, options)

		// Fisher-Yates 洗牌算法
		for i := len(result) - 1; i > 0; i-- {
			j := randOp.RandomInt(0, i)
			result[i], result[j] = result[j], result[i]
		}

		return result
	}

	// 使用map来避免重复
	selected := make(map[int]struct{})
	result := make([]T, 0, count)

	for len(selected) < count {
		index := randOp.RandomInt(0, len(options)-1)
		if _, exists := selected[index]; !exists {
			selected[index] = struct{}{}
			result = append(result, options[index])
		}
	}

	return result
}

// Rand 默认随机数操作实例（公开）
var Rand = &randEngine{}
