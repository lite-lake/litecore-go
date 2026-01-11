// Package time 提供时间处理工具，支持Java风格的时间格式化、解析、计算和转换
package time

/*
time 包提供了丰富的时间处理功能，兼容 Java 日期格式语法，简化时间操作。

核心特性：
  - Java风格格式化：支持 yyyy-MM-dd HH:mm:ss 等 Java 常用格式
  - 智能解析：自动识别多种日期格式，支持多格式备选解析
  - 时间计算：提供日期增减、时间差计算等常用操作
  - 快速转换：提供常用格式的快捷转换方法
  - 时区支持：支持时区转换和时间戳操作
  - 时间边界：快速获取天/周/月/年的时间边界

基本用法：

	import "litecore-go/util/time"

	// 获取当前时间并格式化
	now := time.Time.Now()
	formatted := time.Time.FormatWithJava(now, "yyyy-MM-dd HH:mm:ss")
	// 输出: 2024-01-15 14:30:45

	// 解析日期字符串
	parsed, err := time.Time.ParseWithJava("2024-01-15", "yyyy-MM-dd")
	if err != nil {
	    // 处理解析错误
	}

	// 快速格式化（常用格式）
	dateStr := time.Time.ToYYYYMMDD(now)            // "20240115"
	datetimeStr := time.Time.ToYYYY_MM_DD_HH_MM_SS(now) // "2024-01-15 14:30:45"

	// 快速解析
	parsed, err := time.Time.FromYYYYMMDD("20240115")
	if err != nil {
	    // 处理解析错误
	}

	// 时间计算
	tomorrow := time.Time.AddDays(now, 1)
	nextMonth := time.Time.AddMonths(now, 1)
	days := time.Time.DaysBetween(now, parsed)

	// 获取时间边界
	startOfDay := time.Time.StartOfDay(now)
	endOfMonth := time.Time.EndOfMonth(now)

	// 自动识别格式解析
	autoParsed, err := time.Time.ParseAuto("2024-01-15 14:30:45")
	if err != nil {
	    // 处理解析错误
	}

Java 格式化语法：

	time 包支持以下 Java 风格的日期格式模式：

	  yyyy - 四位年份          yy - 两位年份
	  MM   - 两位月份          M  - 一位月份
	  dd   - 两位日期          d  - 一位日期
	  HH   - 24小时制小时      H  - 一位小时
	  mm   - 两位分钟          m  - 一位分钟
	  ss   - 两位秒            s  - 一位秒
	  SSS  - 三位毫秒          SS - 两位毫秒  S - 一位毫秒

常用格式示例：

	  "yyyy-MM-dd"           → "2024-01-15"
	  "yyyy/MM/dd"           → "2024/01/15"
	  "yyyy-MM-dd HH:mm:ss"  → "2024-01-15 14:30:45"
	  "yyyyMMdd"             → "20240115"
	  "yyyyMMddHHmmss"       → "20240115143045"
	  "HH:mm:ss"             → "14:30:45"

错误处理：

	time 包提供两种解析方式：
	  - ParseWithJava/ParseAuto：返回错误，适合需要明确错误处理的场景
	  - TryParseWithJava/TryParseAuto：失败返回零值，适合快速处理

多格式解析：

	// 尝试多种格式解析，直到成功
	formats := []string{"yyyy-MM-dd", "yyyy/MM/dd", "yyyyMMdd"}
	parsed, err := time.Time.ParseWithMultipleFormats(dateStr, formats)
	if err != nil {
	    // 所有格式都解析失败
	}
*/
