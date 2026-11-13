// Package rand 提供生成各种类型随机数和随机字符串的工具函数。
//
// 核心特性:
//   - 生成随机整数：支持指定范围内的随机整数
//   - 生成随机浮点数：支持指定精度的小数
//   - 生成随机字符串：支持自定义字符集和长度
//   - 生成随机UUID：生成标准的UUID v4格式字符串
//   - 随机选择：从切片中随机选择一个或多个元素
//   - 随机布尔值：生成随机的true/false值
//
// 基本用法:
//
//	// 生成随机字符串（32位，包含字母和数字）
//	str := rand.RandomString(32)
//
//	// 生成UUID
//	uuid := rand.RandomUUID()
//
//	// 生成指定范围的随机整数（1-100）
//	num := rand.RandomInt(1, 100)
//
//	// 生成指定范围的随机浮点数（0.0-1.0，保留2位小数）
//	f := rand.RandomFloat(0.0, 1.0, 2)
//
//	// 从切片中随机选择一个元素
//	choices := []string{"apple", "banana", "orange"}
//	pick := rand.RandomChoice(choices)
//
//	// 从切片中随机选择多个元素（不重复）
//	picks := rand.RandomChoices(choices, 2)
//
// 随机性说明:
// 本包使用 crypto/rand 作为随机源，确保生成的随机数具有密码学强度的安全性。
// 所有随机数生成操作都可能返回错误，调用方应适当处理错误。
package rand
