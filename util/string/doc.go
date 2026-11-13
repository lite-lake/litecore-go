// Package string 字符串处理工具集，基于lancet库提供丰富的字符串操作功能
//
// 核心特性：
//   - 字符串转换：支持驼峰命名(CamelCase)与蛇形命名(SnakeCase)互转
//   - 字符串分割与连接：提供灵活的字符串拆分与合并操作
//   - 字符串裁剪：去除首尾空格或指定字符
//   - 字符串查找：检查子串是否存在，支持大小写敏感/不敏感
//   - 字符串替换：支持全局或指定次数的字符串替换
//   - 字符串截取：按位置或长度截取子串
//
// 基本用法：
//
//	// 驼峰命名与蛇形命名转换
//	camel := util.string.ToCamelCase("hello_world")  // "HelloWorld"
//	snake := util.string.ToSnakeCase("HelloWorld")   // "hello_world"
//
//	// 字符串分割与连接
//	parts := util.string.Split("a,b,c", ",")        // []string{"a", "b", "c"}
//	joined := util.string.Join([]string{"a", "b"}, ",")  // "a,b"
//
//	// 字符串裁剪
//	trimmed := util.string.Trim("  hello  ")         // "hello"
//
//	// 字符串查找
//	exists := util.string.Contains("hello", "ell")   // true
//
//	// 字符串替换
//	replaced := util.string.Replace("hello", "l", "x") // "hexxo"
//
//	// 字符串截取
//	sub := util.string.SubString("hello", 1, 3)      // "el"
//
// 错误处理：
// 该包中的函数通常不会返回错误，对于非法输入会返回空字符串或零值。
// 建议在使用前验证输入参数的有效性。
package string
