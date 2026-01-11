// Package id 提供 ID 生成工具函数
// 支持 CUID2 风格的分布式唯一 ID 生成，具有高可读性和低碰撞概率
package id

import (
	"crypto/rand"
	"encoding/binary"
	"math/big"
	"time"
)

const (
	// cuid2Length 定义 CUID2 标准长度（25 个字符）
	cuid2Length = 25
	// alphabet 用于 base36 编码的字符集（数字 + 小写字母）
	alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
)

var (
	// bigIntAlphabetLen 字符集长度的大整数表示，用于随机数生成
	bigIntAlphabetLen = big.NewInt(int64(len(alphabet)))
)

// NewCUID2 生成 CUID2 风格的唯一标识符
// 返回 25 个字符的小写字母数字字符串，具有以下特性：
//   - 时间有序：前缀包含时间戳，保证大致按时间排序
//   - 高唯一性：结合时间戳和加密级随机数，碰撞概率极低
//   - 可读性：仅包含小写字母和数字，便于人类识别
//   - 分布式安全：无需中央协调，各节点独立生成
//
// 适用场景：数据库主键、分布式追踪 ID、会话 ID 等
func NewCUID2() string {
	// 使用毫秒级时间戳作为前缀，保证时间排序特性
	timestamp := time.Now().UnixMilli()

	// 生成 16 字节加密级随机数，确保唯一性和不可预测性
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		panic(err) // 随机数生成器故障属于系统级错误，应直接终止
	}

	// 编码生成 CUID2 字符串并截取标准长度
	id := encodeCUID2(timestamp, randomBytes)
	return id[:cuid2Length]
}

// encodeCUID2 将时间戳和随机字节编码为 CUID2 格式字符串
// 采用 base36 编码（0-9 和 a-z），使 ID 更短且可读
func encodeCUID2(timestamp int64, randomBytes []byte) string {
	// 将时间戳转换为 base36 字符串作为前缀
	timestampPart := encodeBase36(uint64(timestamp))

	// 将随机字节每 4 字节一组转换为 base36 字符串
	// 使用 BigEndian 保证字节序一致性
	var randomPart string
	for i := 0; i < len(randomBytes); i += 4 {
		if i+4 <= len(randomBytes) {
			num := binary.BigEndian.Uint32(randomBytes[i : i+4])
			randomPart += encodeBase36(uint64(num))
		}
	}

	return timestampPart + randomPart
}

// encodeBase36 将无符号整数转换为 base36 编码字符串
// base36 编码使用 0-9 和 a-z 共 36 个字符，相比 base16（十六进制）更紧凑
// 编码顺序：最高位在左侧，符合人类阅读习惯
func encodeBase36(num uint64) string {
	// 特殊情况：0 直接返回
	if num == 0 {
		return "0"
	}

	// 预分配 13 字节缓冲区（uint64 最大值约为 1.8e19，base36 编码后约 13 位）
	chars := make([]byte, 0, 13)
	for num > 0 {
		remainder := num % 36
		chars = append(chars, alphabet[remainder])
		num /= 36
	}

	// 反转字符数组：先计算出的是低位（在右侧），需要反转使高位在左侧
	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}

	return string(chars)
}
