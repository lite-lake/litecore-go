# util/time - 时间处理工具包

提供强大的时间处理功能，支持 Java 风格的时间格式化、解析、计算和转换。

## 特性

- **Java 风格格式化** - 完全兼容 Java 日期格式语法，支持 `yyyy-MM-dd HH:mm:ss` 等常用格式
- **智能解析** - 自动识别多种日期格式，支持多格式备选解析
- **时间计算** - 提供日期增减、时间差计算等常用操作
- **快速转换** - 提供常用格式的快捷方法，如 `ToYYYYMMDD`、`FromYYYYMMDD`
- **时区支持** - 支持时区转换和时间戳操作
- **时间边界** - 快速获取天/周/月/年的时间边界

## 快速开始

```go
package main

import (
    "fmt"
    stdtime "time"

    "litecore-go/util/time"
)

func main() {
    // 获取当前时间并格式化
    now := time.Time.Now()
    formatted := time.Time.FormatWithJava(now, "yyyy-MM-dd HH:mm:ss")
    fmt.Println("当前时间:", formatted)
    // 输出: 当前时间: 2024-01-15 14:30:45

    // 解析日期字符串
    parsed, err := time.Time.ParseWithJava("2024-01-15", "yyyy-MM-dd")
    if err != nil {
        panic(err)
    }
    fmt.Println("解析结果:", parsed)
    // 输出: 解析结果: 2024-01-15 00:00:00 +0000 UTC

    // 快速格式化（常用格式）
    dateStr := time.Time.ToYYYYMMDD(now)                    // "20240115"
    datetimeStr := time.Time.ToYYYY_MM_DD_HH_MM_SS(now)     // "2024-01-15 14:30:45"
    fmt.Println("日期:", dateStr)
    fmt.Println("日期时间:", datetimeStr)

    // 快速解析
    parsed, err = time.Time.FromYYYYMMDD("20240115")
    if err != nil {
        panic(err)
    }
    fmt.Println("快速解析:", parsed)

    // 时间计算
    tomorrow := time.Time.AddDays(now, 1)
    nextMonth := time.Time.AddMonths(now, 1)
    days := time.Time.DaysBetween(now, parsed)
    fmt.Println("明天:", time.Time.ToYYYY_MM_DD(tomorrow))
    fmt.Println("下个月:", time.Time.ToYYYY_MM_DD(nextMonth))
    fmt.Println("相差天数:", days)

    // 获取时间边界
    startOfDay := time.Time.StartOfDay(now)
    endOfMonth := time.Time.EndOfMonth(now)
    fmt.Println("今天开始:", time.Time.ToYYYY_MM_DD_HH_MM_SS(startOfDay))
    fmt.Println("月末结束:", time.Time.ToYYYY_MM_DD_HH_MM_SS(endOfMonth))

    // 自动识别格式解析
    autoParsed, err := time.Time.ParseAuto("2024-01-15 14:30:45")
    if err != nil {
        panic(err)
    }
    fmt.Println("自动解析:", time.Time.ToYYYY_MM_DD_HH_MM_SS(autoParsed))
}
```

## 功能详解

### 时间格式化

支持 Java 风格的日期格式语法，让时间格式化更加直观。

```go
import (
    stdtime "time"
    "litecore-go/util/time"
)

// 基本格式化
now := time.Time.Now()
formatted := time.Time.FormatWithJava(now, "yyyy-MM-dd HH:mm:ss")
fmt.Println(formatted)  // 2024-01-15 14:30:45

// 不同格式示例
formats := []string{
    "yyyy-MM-dd",              // 2024-01-15
    "yyyy/MM/dd",              // 2024/01/15
    "yyyy年MM月dd日",          // 2024年01月15日
    "yyyyMMdd",                // 20240115
    "yyyy-MM-dd HH:mm:ss",     // 2024-01-15 14:30:45
    "yyyy-MM-dd HH:mm:ss.SSS", // 2024-01-15 14:30:45.123
    "HH:mm:ss",                // 14:30:45
    "HH:mm",                   // 14:30
}

for _, format := range formats {
    result := time.Time.FormatWithJava(now, format)
    fmt.Printf("%s -> %s\n", format, result)
}

// 格式化时处理零值
var zeroTime stdtime.Time
defaultValue := time.Time.FormatWithJavaOrDefault(zeroTime, "yyyy-MM-dd", "未设置时间")
fmt.Println(defaultValue)  // 未设置时间

actualValue := time.Time.FormatWithJavaOrDefault(now, "yyyy-MM-dd", "未设置时间")
fmt.Println(actualValue)   // 2024-01-15
```

### 时间解析

提供灵活的时间解析方法，支持多种格式和错误处理策略。

```go
import "litecore-go/util/time"

// Java 风格解析
parsed, err := time.Time.ParseWithJava("2024-01-15", "yyyy-MM-dd")
if err != nil {
    // 处理解析错误
    fmt.Println("解析失败:", err)
}
fmt.Println(time.Time.ToYYYY_MM_DD(parsed))  // 2024-01-15

// 尝试解析（失败返回零值，不报错）
result := time.Time.TryParseWithJava("2024-01-15", "yyyy-MM-dd")
if time.Time.IsNotZero(result) {
    fmt.Println("解析成功:", result)
} else {
    fmt.Println("解析失败")
}

// 多格式备选解析
formats := []string{
    "yyyy-MM-dd",
    "yyyy/MM/dd",
    "yyyyMMdd",
}
dateStr := "2024/01/15"
parsed, err = time.Time.ParseWithMultipleFormats(dateStr, formats)
if err != nil {
    fmt.Println("所有格式都解析失败:", err)
} else {
    fmt.Println("解析成功:", time.Time.ToYYYY_MM_DD(parsed))
}

// 自动识别格式解析
// 支持常见格式：yyyy-MM-dd, yyyy/MM/dd, yyyyMMdd, yyyy-MM-dd HH:mm:ss 等
autoParsed, err := time.Time.ParseAuto("2024-01-15 14:30:45")
if err != nil {
    fmt.Println("自动解析失败:", err)
} else {
    fmt.Println("自动解析成功:", time.Time.ToYYYY_MM_DD_HH_MM_SS(autoParsed))
}

// 尝试自动解析（失败返回零值）
result = time.Time.TryParseAuto("2024-01-15")
if time.Time.IsNotZero(result) {
    fmt.Println("自动解析成功:", result)
}
```

### 时间计算

提供丰富的时间计算方法，包括日期增减和时间差计算。

```go
import stdtime "time"

now := time.Time.Now()

// 增加时间
tomorrow := time.Time.AddDays(now, 1)
yesterday := time.Time.AddDays(now, -1)
nextWeek := time.Time.AddDays(now, 7)

nextMonth := time.Time.AddMonths(now, 1)
nextYear := time.Time.AddYears(now, 1)

// 使用 AddDuration 一次性增加年月日
future := time.Time.AddDuration(now, 1, 2, 3)  // 1年2个月3天后
fmt.Println("未来时间:", time.Time.ToYYYY_MM_DD(future))

// 使用 Add 增加任意时长
oneHourLater := time.Time.Add(now, stdtime.Hour)
twoHoursLater := time.Time.Add(now, 2*stdtime.Hour)

// 计算时间差
otherTime := time.Time.AddDays(now, 5)
duration := time.Time.Sub(otherTime, now)  // 返回 time.Duration
fmt.Println("时间差:", duration.Hours(), "小时")

// 计算毫秒数差值
millis := time.Time.DurationBetween(otherTime, now)
fmt.Println("毫秒差:", millis)

// 计算天数差
days := time.Time.DaysBetween(otherTime, now)
fmt.Println("天数差:", days)  // 5

// 计算日期到现在的年龄
birthDate, _ := time.Time.ParseWithJava("1990-05-20", "yyyy-MM-dd")
age := time.Time.Age(birthDate)
fmt.Println("年龄:", age)
```

### 时间转换

快速获取时间边界，如一天的开始/结束、一周的开始/结束等。

```go
import stdtime "time"

now := time.Time.Now()

// 获取一天的开始和结束
startOfDay := time.Time.StartOfDay(now)   // 2024-01-15 00:00:00
endOfDay := time.Time.EndOfDay(now)       // 2024-01-15 23:59:59

// 获取一周的开始和结束
startOfWeek := time.Time.StartOfWeek(now) // 周一 00:00:00
endOfWeek := time.Time.EndOfWeek(now)     // 周日 23:59:59

// 获取一个月的开始和结束
startOfMonth := time.Time.StartOfMonth(now) // 2024-01-01 00:00:00
endOfMonth := time.Time.EndOfMonth(now)     // 2024-01-31 23:59:59

// 获取一年的开始和结束
startOfYear := time.Time.StartOfYear(now)   // 2024-01-01 00:00:00
endOfYear := time.Time.EndOfYear(now)       // 2024-12-31 23:59:59

// 使用场景：查询今天的数据
todayStart := time.Time.StartOfDay(time.Time.Now())
todayEnd := time.Time.EndOfDay(time.Time.Now())
fmt.Printf("查询今天数据: %s - %s\n",
    time.Time.ToYYYY_MM_DD_HH_MM_SS(todayStart),
    time.Time.ToYYYY_MM_DD_HH_MM_SS(todayEnd))

// 使用场景：查询本月的数据
monthStart := time.Time.StartOfMonth(time.Time.Now())
monthEnd := time.Time.EndOfMonth(time.Time.Now())
fmt.Printf("查询本月数据: %s - %s\n",
    time.Time.ToYYYY_MM_DD_HH_MM_SS(monthStart),
    time.Time.ToYYYY_MM_DD_HH_MM_SS(monthEnd))
```

### 快速格式化和解析

提供常用格式的快捷方法，简化代码。

```go
import "litecore-go/util/time"

now := time.Time.Now()

// 快速格式化
dateStr := time.Time.ToYYYYMMDD(now)                    // "20240115"
datetimeStr := time.Time.ToYYYYMMDDHHMMSS(now)          // "20240115143045"
formattedDate := time.Time.ToYYYY_MM_DD(now)            // "2024-01-15"
formattedDatetime := time.Time.ToYYYY_MM_DD_HH_MM_SS(now) // "2024-01-15 14:30:45"
timeOnly := time.Time.ToHHMMSS(now)                     // "14:30:45"

fmt.Println(dateStr, formattedDate, formattedDatetime, timeOnly)

// 快速解析
parsed1, err := time.Time.FromYYYYMMDD("20240115")
if err != nil {
    panic(err)
}

parsed2, err := time.Time.FromYYYYMMDDHHMMSS("20240115143045")
if err != nil {
    panic(err)
}

parsed3, err := time.Time.FromYYYY_MM_DD("2024-01-15")
if err != nil {
    panic(err)
}

parsed4, err := time.Time.FromYYYY_MM_DD_HH_MM_SS("2024-01-15 14:30:45")
if err != nil {
    panic(err)
}

fmt.Println(time.Time.ToYYYY_MM_DD(parsed1))
fmt.Println(time.Time.ToYYYY_MM_DD_HH_MM_SS(parsed2))
fmt.Println(time.Time.ToYYYY_MM_DD(parsed3))
fmt.Println(time.Time.ToYYYY_MM_DD_HH_MM_SS(parsed4))
```

### 时区转换

支持时区转换和 UTC 时间操作。

```go
import stdtime "time"

now := time.Time.Now()

// 转换到 UTC
utcTime := time.Time.UTC(now)
fmt.Println("UTC时间:", time.Time.ToYYYY_MM_DD_HH_MM_SS(utcTime))

// 转换到本地时间
localTime := time.Time.Local(now)
fmt.Println("本地时间:", time.Time.ToYYYY_MM_DD_HH_MM_SS(localTime))

// 转换到指定时区
loc, err := time.Time.LoadLocation("America/New_York")
if err != nil {
    panic(err)
}
nyTime := time.Time.InLocation(now, loc)
fmt.Println("纽约时间:", time.Time.ToYYYY_MM_DD_HH_MM_SS(nyTime))

// 其他时区示例
tokyoLoc, _ := time.Time.LoadLocation("Asia/Tokyo")
londonLoc, _ := time.Time.LoadLocation("Europe/London")
tokyoTime := time.Time.InLocation(now, tokyoLoc)
londonTime := time.Time.InLocation(now, londonLoc)
```

### 时间戳操作

提供时间戳与时间的相互转换。

```go
import "litecore-go/util/time"

now := time.Time.Now()

// 转换为 Unix 时间戳（秒）
timestamp := time.Time.ToUnix(now)
fmt.Println("Unix时间戳(秒):", timestamp)

// 转换为 Unix 时间戳（毫秒）
timestampMillis := time.Time.ToUnixMilli(now)
fmt.Println("Unix时间戳(毫秒):", timestampMillis)

// 获取当前时间戳
currentTimestamp := time.Time.NowUnix()          // 秒
currentTimestampMillis := time.Time.NowUnixMilli() // 毫秒
fmt.Println("当前时间戳:", currentTimestamp)

// 从 Unix 时间戳创建时间
fromTimestamp := time.Time.FromUnix(1705310400)
fmt.Println("从时间戳创建:", time.Time.ToYYYY_MM_DD_HH_MM_SS(fromTimestamp))

// 从毫秒时间戳创建时间
fromTimestampMillis := time.Time.FromUnixMilli(1705310400000)
fmt.Println("从毫秒时间戳创建:", time.Time.ToYYYY_MM_DD_HH_MM_SS(fromTimestampMillis))
```

### 时间工具方法

提供各种实用的时间检查和转换方法。

```go
import stdtime "time"

now := time.Time.Now()

// 时间检查
zeroTime := stdtime.Time{}
fmt.Println("是否为零值:", time.Time.IsZero(zeroTime))       // true
fmt.Println("是否非零值:", time.Time.IsNotZero(now))          // true
fmt.Println("是否在之后:", time.Time.After(now, zeroTime))    // true
fmt.Println("是否在之前:", time.Time.Before(zeroTime, now))   // true
fmt.Println("是否相等:", time.Time.Equal(now, now))           // true

// 时间范围检查
start := time.Time.Now()
end := time.Time.AddDays(start, 7)
checkTime := time.Time.AddDays(start, 3)
isBetween := time.Time.Between(checkTime, start, end)
fmt.Println("是否在范围内:", isBetween)  // true

// 闰年判断
isLeap := time.Time.IsLeapYear(2024)
fmt.Println("2024是闰年:", isLeap)  // true

isLeap = time.Time.IsLeapYear(2023)
fmt.Println("2023是闰年:", isLeap)  // false

// 获取月份天数
daysInMonth := time.Time.DaysInMonth(2024, 2)  // 2024年2月
fmt.Println("2024年2月天数:", daysInMonth)  // 29

daysInMonth = time.Time.DaysInMonth(2023, 2)  // 2023年2月
fmt.Println("2023年2月天数:", daysInMonth)  // 28

// 格式化持续时间
duration := 2*stdtime.Hour + 30*stdtime.Minute + 45*stdtime.Second
formatted := time.Time.FormatDuration(duration)
fmt.Println("持续时长:", formatted)  // 02:30:45

shortDuration := 30*stdtime.Minute + 45*stdtime.Second
formatted = time.Time.FormatDuration(shortDuration)
fmt.Println("持续时长:", formatted)  // 30:45

// 时间精度处理
truncated := time.Time.Truncate(now, stdtime.Hour)  // 截断到小时
rounded := time.Time.Round(now, stdtime.Hour)       // 四舍五入到小时
fmt.Println("截断到小时:", time.Time.ToYYYY_MM_DD_HH_MM_SS(truncated))
fmt.Println("四舍五入到小时:", time.Time.ToYYYY_MM_DD_HH_MM_SS(rounded))
```

## API 参考

### 基础时间检查

| 方法 | 说明 |
|------|------|
| `IsZero(tim time.Time) bool` | 检查时间是否为零值 |
| `IsNotZero(tim time.Time) bool` | 检查时间是否不为零值 |
| `After(tim, other time.Time) bool` | 检查时间是否在另一个时间之后 |
| `Before(tim, other time.Time) bool` | 检查时间是否在另一个时间之前 |
| `Equal(tim, other time.Time) bool` | 检查两个时间是否相等 |

### 时间获取

| 方法 | 说明 |
|------|------|
| `Now() time.Time` | 获取当前时间 |
| `NowUnix() int64` | 获取当前时间的Unix时间戳（秒） |
| `NowUnixMilli() int64` | 获取当前时间的Unix时间戳（毫秒） |
| `Unix(sec, nsec int64) time.Time` | 根据Unix时间戳创建时间 |
| `Parse(layout, value string) (time.Time, error)` | 使用Go标准格式解析时间 |

### Java风格格式化

| 方法 | 说明 |
|------|------|
| `ConvertJavaFormatToGo(javaFormat string) string` | 将Java格式转换为Go格式 |
| `FormatWithJava(tim time.Time, javaFormat string) string` | 使用Java风格格式化时间 |
| `FormatWithJavaOrDefault(tim time.Time, javaFormat, defaultValue string) string` | 格式化时间，零值返回默认值 |

### Java风格解析

| 方法 | 说明 |
|------|------|
| `ParseWithJava(value, javaFormat string) (time.Time, error)` | 使用Java风格格式解析时间 |
| `TryParseWithJava(value, javaFormat string) time.Time` | 尝试解析，失败返回零值 |
| `ParseWithMultipleFormats(value string, formats []string) (time.Time, error)` | 使用多种格式解析 |
| `TryParseWithMultipleFormats(value string, formats []string) time.Time` | 尝试多格式解析，失败返回零值 |
| `ParseAuto(value string) (time.Time, error)` | 自动识别并解析时间 |
| `TryParseAuto(value string) time.Time` | 尝试自动解析，失败返回零值 |

### 时间计算

| 方法 | 说明 |
|------|------|
| `Add(tim time.Time, d time.Duration) time.Time` | 增加时间 |
| `AddDuration(tim time.Time, years, months, days int) time.Time` | 增加指定年月日 |
| `AddYears(tim time.Time, years int) time.Time` | 增加年数 |
| `AddMonths(tim time.Time, months int) time.Time` | 增加月数 |
| `AddDays(tim time.Time, days int) time.Time` | 增加天数 |
| `Sub(tim, other time.Time) time.Duration` | 计算时间差 |
| `DurationBetween(tim, other time.Time) int64` | 计算毫秒数差值 |
| `DaysBetween(tim, other time.Time) int` | 计算天数差 |

### 时间转换

| 方法 | 说明 |
|------|------|
| `StartOfDay(tim time.Time) time.Time` | 获取一天的开始时间 |
| `EndOfDay(tim time.Time) time.Time` | 获取一天的结束时间 |
| `StartOfWeek(tim time.Time) time.Time` | 获取一周的开始时间（周一） |
| `EndOfWeek(tim time.Time) time.Time` | 获取一周的结束时间（周日） |
| `StartOfMonth(tim time.Time) time.Time` | 获取一个月的开始时间 |
| `EndOfMonth(tim time.Time) time.Time` | 获取一个月的结束时间 |
| `StartOfYear(tim time.Time) time.Time` | 获取一年的开始时间 |
| `EndOfYear(tim time.Time) time.Time` | 获取一年的结束时间 |

### 时间工具

| 方法 | 说明 |
|------|------|
| `IsLeapYear(year int) bool` | 检查是否为闰年 |
| `DaysInMonth(year, month int) int` | 获取指定月份的天数 |
| `FormatDuration(d time.Duration) string` | 格式化持续时间 |
| `Age(birthDate time.Time) int` | 计算年龄 |
| `Between(tim, start, end time.Time) bool` | 检查时间是否在范围内 |
| `Truncate(tim time.Time, d time.Duration) time.Time` | 截断时间到指定精度 |
| `Round(tim time.Time, d time.Duration) time.Time` | 四舍五入时间到指定精度 |

### 快速格式化

| 方法 | 说明 |
|------|------|
| `ToYYYYMMDD(tim time.Time) string` | 格式化为 yyyyMMdd |
| `ToYYYYMMDDHHMMSS(tim time.Time) string` | 格式化为 yyyyMMddHHmmss |
| `ToYYYY_MM_DD(tim time.Time) string` | 格式化为 yyyy-MM-dd |
| `ToYYYY_MM_DD_HH_MM_SS(tim time.Time) string` | 格式化为 yyyy-MM-dd HH:mm:ss |
| `ToHHMMSS(tim time.Time) string` | 格式化为 HH:mm:ss |

### 快速解析

| 方法 | 说明 |
|------|------|
| `FromYYYYMMDD(value string) (time.Time, error)` | 从 yyyyMMdd 格式解析 |
| `FromYYYYMMDDHHMMSS(value string) (time.Time, error)` | 从 yyyyMMddHHmmss 格式解析 |
| `FromYYYY_MM_DD(value string) (time.Time, error)` | 从 yyyy-MM-dd 格式解析 |
| `FromYYYY_MM_DD_HH_MM_SS(value string) (time.Time, error)` | 从 yyyy-MM-dd HH:mm:ss 格式解析 |

### 时区相关

| 方法 | 说明 |
|------|------|
| `InLocation(tim time.Time, loc *time.Location) time.Time` | 转换到指定时区 |
| `UTC(tim time.Time) time.Time` | 转换到UTC时区 |
| `Local(tim time.Time) time.Time` | 转换到本地时区 |
| `LoadLocation(name string) (*time.Location, error)` | 加载时区 |

### 时间戳相关

| 方法 | 说明 |
|------|------|
| `ToUnix(tim time.Time) int64` | 转换为Unix时间戳（秒） |
| `ToUnixMilli(tim time.Time) int64` | 转换为Unix时间戳（毫秒） |
| `FromUnix(sec int64) time.Time` | 从Unix时间戳（秒）创建时间 |
| `FromUnixMilli(msec int64) time.Time` | 从Unix时间戳（毫秒）创建时间 |

### 验证和辅助

| 方法 | 说明 |
|------|------|
| `IsValidFormat(javaFormat string) bool` | 检查是否为有效的Java日期格式 |
| `GuessFormat(value string) string` | 猜测时间字符串的格式 |

## Java 格式化语法说明

time 包完全兼容 Java 风格的日期格式语法。下表列出了所有支持的格式模式：

### 日期模式

| 模式 | 说明 | 示例 |
|------|------|------|
| `yyyy` | 四位年份 | 2024 |
| `yy` | 两位年份 | 24 |
| `MM` | 两位月份（01-12） | 01, 12 |
| `M` | 一位月份（1-12） | 1, 12 |
| `dd` | 两位日期（01-31） | 01, 31 |
| `d` | 一位日期（1-31） | 1, 31 |

### 时间模式

| 模式 | 说明 | 示例 |
|------|------|------|
| `HH` | 24小时制小时（00-23） | 00, 23 |
| `H` | 24小时制小时（0-23） | 0, 23 |
| `mm` | 两位分钟（00-59） | 00, 59 |
| `m` | 一位分钟（0-59） | 0, 59 |
| `ss` | 两位秒（00-59） | 00, 59 |
| `s` | 一位秒（0-59） | 0, 59 |

### 毫秒模式

| 模式 | 说明 | 示例 |
|------|------|------|
| `SSS` | 三位毫秒（000-999） | 000, 999 |
| `SS` | 两位毫秒（00-99） | 00, 99 |
| `S` | 一位毫秒（0-9） | 0, 9 |

### 常用格式示例

| Java 格式 | 输出示例 | 快捷方法 |
|-----------|----------|----------|
| `yyyy-MM-dd` | 2024-01-15 | `ToYYYY_MM_DD` |
| `yyyy/MM/dd` | 2024/01/15 | - |
| `yyyy年MM月dd日` | 2024年01月15日 | - |
| `yyyyMMdd` | 20240115 | `ToYYYYMMDD` |
| `yyyy-MM-dd HH:mm:ss` | 2024-01-15 14:30:45 | `ToYYYY_MM_DD_HH_MM_SS` |
| `yyyy/MM/dd HH:mm:ss` | 2024/01/15 14:30:45 | - |
| `yyyyMMddHHmmss` | 20240115143045 | `ToYYYYMMDDHHMMSS` |
| `yyyy-MM-dd HH:mm:ss.SSS` | 2024-01-15 14:30:45.123 | - |
| `HH:mm:ss` | 14:30:45 | `ToHHMMSS` |
| `HH:mm` | 14:30 | - |
| `MM-dd` | 01-15 | - |
| `MM/dd` | 01/15 | - |

## 使用建议

### 错误处理

对于关键的时间解析操作，建议使用返回 error 的方法：

```go
parsed, err := time.Time.ParseWithJava(dateStr, "yyyy-MM-dd")
if err != nil {
    // 处理错误
    return fmt.Errorf("日期解析失败: %w", err)
}
```

对于非关键场景，可以使用 Try 方法简化代码：

```go
parsed := time.Time.TryParseWithJava(dateStr, "yyyy-MM-dd")
if time.Time.IsNotZero(parsed) {
    // 解析成功，使用 parsed
}
```

### 格式选择

- **数据库存储**：使用 `yyyy-MM-dd HH:mm:ss` 或 `yyyyMMddHHmmss`
- **API 响应**：使用 `yyyy-MM-dd HH:mm:ss`
- **文件名**：使用 `yyyyMMdd` 或 `yyyyMMddHHmmss`
- **日志记录**：使用 `yyyy-MM-dd HH:mm:ss` 或 `yyyy-MM-dd HH:mm:ss.SSS`
- **用户显示**：根据地区习惯选择格式

### 性能优化

对于频繁使用的固定格式，建议使用快速方法：

```go
// 推荐：使用快速方法
dateStr := time.Time.ToYYYYMMDD(now)

// 不推荐：重复使用相同格式
dateStr := time.Time.FormatWithJava(now, "yyyyMMdd")
```

## 常见问题

### Q: 如何解析不同格式的日期字符串？

使用 `ParseWithMultipleFormats` 或 `ParseAuto`：

```go
// 方法1：指定可能的格式
formats := []string{"yyyy-MM-dd", "yyyy/MM/dd", "yyyyMMdd"}
parsed, err := time.Time.ParseWithMultipleFormats(dateStr, formats)

// 方法2：自动识别格式
parsed, err := time.Time.ParseAuto(dateStr)
```

### Q: 如何获取昨天或明天的日期？

```go
yesterday := time.Time.AddDays(time.Time.Now(), -1)
tomorrow := time.Time.AddDays(time.Time.Now(), 1)
```

### Q: 如何计算两个日期之间相差的天数？

```go
days := time.Time.DaysBetween(date1, date2)
```

### Q: 如何判断一个时间是否在指定范围内？

```go
isInRange := time.Time.Between(checkTime, startTime, endTime)
```

### Q: 如何获取本月的最后一天？

```go
endOfMonth := time.Time.EndOfMonth(time.Time.Now())
lastDay := endOfMonth.Day()
```

## 许可证

本包是 litecore-go 项目的一部分，遵循项目的开源许可证。
