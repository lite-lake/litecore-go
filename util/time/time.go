package time

import (
	"fmt"
	"regexp"
	"strings"
	stdtime "time"
)

// ILiteUtilTime 时间工具接口
type ILiteUtilTime interface {
	// 基础时间检查
	IsZero(tim stdtime.Time) bool
	IsNotZero(tim stdtime.Time) bool
	After(tim, other stdtime.Time) bool
	Before(tim, other stdtime.Time) bool
	Equal(tim, other stdtime.Time) bool

	// 时间获取
	Now() stdtime.Time
	NowUnix() int64
	NowUnixMilli() int64
	Unix(sec int64, nsec int64) stdtime.Time
	Parse(layout, value string) (stdtime.Time, error)

	// Java风格格式化
	ConvertJavaFormatToGo(javaFormat string) string
	FormatWithJava(tim stdtime.Time, javaFormat string) string
	FormatWithJavaOrDefault(tim stdtime.Time, javaFormat, defaultValue string) string

	// Java风格解析
	ParseWithJava(value, javaFormat string) (stdtime.Time, error)
	TryParseWithJava(value, javaFormat string) stdtime.Time
	ParseWithMultipleFormats(value string, formats []string) (stdtime.Time, error)
	TryParseWithMultipleFormats(value string, formats []string) stdtime.Time
	ParseAuto(value string) (stdtime.Time, error)
	TryParseAuto(value string) stdtime.Time

	// 时间计算
	Add(tim stdtime.Time, d stdtime.Duration) stdtime.Time
	AddDuration(tim stdtime.Time, years, months, days int) stdtime.Time
	AddYears(tim stdtime.Time, years int) stdtime.Time
	AddMonths(tim stdtime.Time, months int) stdtime.Time
	AddDays(tim stdtime.Time, days int) stdtime.Time
	Sub(tim, other stdtime.Time) stdtime.Duration
	DurationBetween(tim, other stdtime.Time) int64
	DaysBetween(tim, other stdtime.Time) int

	// 时间转换
	StartOfDay(tim stdtime.Time) stdtime.Time
	EndOfDay(tim stdtime.Time) stdtime.Time
	StartOfWeek(tim stdtime.Time) stdtime.Time
	EndOfWeek(tim stdtime.Time) stdtime.Time
	StartOfMonth(tim stdtime.Time) stdtime.Time
	EndOfMonth(tim stdtime.Time) stdtime.Time
	StartOfYear(tim stdtime.Time) stdtime.Time
	EndOfYear(tim stdtime.Time) stdtime.Time

	// 时间工具
	IsLeapYear(year int) bool
	DaysInMonth(year, month int) int
	FormatDuration(d stdtime.Duration) string
	Age(birthDate stdtime.Time) int
	Between(tim, start, end stdtime.Time) bool
	Truncate(tim stdtime.Time, d stdtime.Duration) stdtime.Time
	Round(tim stdtime.Time, d stdtime.Duration) stdtime.Time

	// 快速格式化
	ToYYYYMMDD(tim stdtime.Time) string
	ToYYYYMMDDHHMMSS(tim stdtime.Time) string
	ToYYYY_MM_DD(tim stdtime.Time) string
	ToYYYY_MM_DD_HH_MM_SS(tim stdtime.Time) string
	ToHHMMSS(tim stdtime.Time) string

	// 快速解析
	FromYYYYMMDD(value string) (stdtime.Time, error)
	FromYYYYMMDDHHMMSS(value string) (stdtime.Time, error)
	FromYYYY_MM_DD(value string) (stdtime.Time, error)
	FromYYYY_MM_DD_HH_MM_SS(value string) (stdtime.Time, error)

	// 时区相关
	InLocation(tim stdtime.Time, loc *stdtime.Location) stdtime.Time
	UTC(tim stdtime.Time) stdtime.Time
	Local(tim stdtime.Time) stdtime.Time
	LoadLocation(name string) (*stdtime.Location, error)

	// 时间戳相关
	ToUnix(tim stdtime.Time) int64
	ToUnixMilli(tim stdtime.Time) int64
	FromUnix(sec int64) stdtime.Time
	FromUnixMilli(msec int64) stdtime.Time

	// 验证和辅助
	IsValidFormat(javaFormat string) bool
	GuessFormat(value string) string
}

// timeEngine 时间操作工具类（私有）
type timeEngine struct{}

var (
	// 默认时间操作实例
	defaultTimeOp = newTimeEngine()
	Time          = defaultTimeOp
)

// Default 返回默认的时间工具实例（单例模式）
// Deprecated: 请使用 liteutil.LiteUtil().Time() 来获取时间工具实例

// New 创建一个新的时间工具实例
// Deprecated: 请使用 liteutil.LiteUtil().NewTimeOperation() 来创建新的时间工具实例

// newTimeOperation 创建时间操作工具实例（私有）
func newTimeEngine() ILiteUtilTime {
	return &timeEngine{}
}

// =========================================
// 基础时间检查方法
// =========================================

// IsZero 检查时间是否为零值
func (t *timeEngine) IsZero(tim stdtime.Time) bool {
	return tim.IsZero()
}

// IsNotZero 检查时间是否不为零值
func (t *timeEngine) IsNotZero(tim stdtime.Time) bool {
	return !tim.IsZero()
}

// After 检查时间是否在另一个时间之后
func (t *timeEngine) After(tim, other stdtime.Time) bool {
	return tim.After(other)
}

// Before 检查时间是否在另一个时间之前
func (t *timeEngine) Before(tim, other stdtime.Time) bool {
	return tim.Before(other)
}

// Equal 检查两个时间是否相等
func (t *timeEngine) Equal(tim, other stdtime.Time) bool {
	return tim.Equal(other)
}

// =========================================
// 时间获取方法
// =========================================

// Now 获取当前时间
func (t *timeEngine) Now() stdtime.Time {
	return stdtime.Now()
}

// NowUnix 获取当前时间的Unix时间戳（秒）
func (t *timeEngine) NowUnix() int64 {
	return stdtime.Now().Unix()
}

// NowUnixMilli 获取当前时间的Unix时间戳（毫秒）
func (t *timeEngine) NowUnixMilli() int64 {
	return stdtime.Now().UnixMilli()
}

// Unix 根据Unix时间戳创建时间
func (t *timeEngine) Unix(sec int64, nsec int64) stdtime.Time {
	return stdtime.Unix(sec, nsec)
}

// Parse 从字符串解析时间（使用Go标准格式）
func (t *timeEngine) Parse(layout, value string) (stdtime.Time, error) {
	return stdtime.Parse(layout, value)
}

// =========================================
// Java风格日期格式化方法
// =========================================

// Java风格日期格式模式映射
var javaFormatPattern = map[string]string{
	"yyyy": "2006",
	"yy":   "06",
	"MM":   "01",
	"M":    "1",
	"dd":   "02",
	"d":    "2",
	"HH":   "15",
	"H":    "15",
	"mm":   "04",
	"m":    "4",
	"ss":   "05",
	"s":    "5",
	"SSS":  "000",
	"SS":   "00",
	"S":    "0",
}

// commonJavaFormats 常用的Java日期格式
var commonJavaFormats = []string{
	"yyyyMMdd",
	"yyyy-MM-dd",
	"yyyy/MM/dd",
	"yyyy年MM月dd日",
	"yyyyMMddHHmmss",
	"yyyy-MM-dd HH:mm:ss",
	"yyyy/MM/dd HH:mm:ss",
	"yyyy年MM月dd日 HH:mm:ss",
	"yyyyMMddHHmmssSSS",
	"yyyy-MM-dd HH:mm:ss.SSS",
	"HH:mm:ss",
	"HH:mm",
	"MM-dd",
	"MM/dd",
}

// ConvertJavaFormatToGo 将Java风格的日期格式转换为Go格式
func (t *timeEngine) ConvertJavaFormatToGo(javaFormat string) string {
	// 首先检查是否为常用格式
	for _, common := range commonJavaFormats {
		if javaFormat == common {
			switch javaFormat {
			case "yyyyMMdd":
				return "20060102"
			case "yyyy-MM-dd":
				return "2006-01-02"
			case "yyyy/MM/dd":
				return "2006/01/02"
			case "yyyy年MM月dd日":
				return "2006年01月02日"
			case "yyyyMMddHHmmss":
				return "20060102150405"
			case "yyyy-MM-dd HH:mm:ss":
				return "2006-01-02 15:04:05"
			case "yyyy/MM/dd HH:mm:ss":
				return "2006/01/02 15:04:05"
			case "yyyy年MM月dd日 HH:mm:ss":
				return "2006年01月02日 15:04:05"
			case "yyyyMMddHHmmssSSS":
				return "20060102150405.000"
			case "yyyy-MM-dd HH:mm:ss.SSS":
				return "2006-01-02 15:04:05.000"
			case "HH:mm:ss":
				return "15:04:05"
			case "HH:mm":
				return "15:04"
			case "MM-dd":
				return "01-02"
			case "MM/dd":
				return "01/02"
			}
		}
	}

	// 通用转换逻辑
	result := javaFormat
	// 按照长度从长到短排序，避免部分匹配
	patterns := []string{"yyyy", "yy", "MM", "M", "dd", "d", "HH", "H", "mm", "m", "ss", "s", "SSS", "SS", "S"}

	for _, pattern := range patterns {
		if goPattern, exists := javaFormatPattern[pattern]; exists {
			result = strings.ReplaceAll(result, pattern, goPattern)
		}
	}

	return result
}

// FormatWithJava 使用Java风格格式化时间
func (t *timeEngine) FormatWithJava(tim stdtime.Time, javaFormat string) string {
	goFormat := t.ConvertJavaFormatToGo(javaFormat)
	return tim.Format(goFormat)
}

// FormatWithJavaOrDefault 使用Java风格格式化时间，如果时间为零值则返回默认值
func (t *timeEngine) FormatWithJavaOrDefault(tim stdtime.Time, javaFormat, defaultValue string) string {
	if t.IsZero(tim) {
		return defaultValue
	}
	return t.FormatWithJava(tim, javaFormat)
}

// =========================================
// Java风格日期解析方法
// =========================================

// ParseWithJava 使用Java风格格式解析时间字符串
func (t *timeEngine) ParseWithJava(value, javaFormat string) (stdtime.Time, error) {
	goFormat := t.ConvertJavaFormatToGo(javaFormat)
	return stdtime.Parse(goFormat, value)
}

// TryParseWithJava 尝试使用Java风格格式解析时间字符串，失败返回零值
func (t *timeEngine) TryParseWithJava(value, javaFormat string) stdtime.Time {
	parsed, err := t.ParseWithJava(value, javaFormat)
	if err != nil {
		return stdtime.Time{}
	}
	return parsed
}

// ParseWithMultipleFormats 尝试使用多种Java风格格式解析时间字符串
func (t *timeEngine) ParseWithMultipleFormats(value string, formats []string) (stdtime.Time, error) {
	for _, format := range formats {
		if parsed, err := t.ParseWithJava(value, format); err == nil {
			return parsed, nil
		}
	}
	return stdtime.Time{}, fmt.Errorf("unable to parse time with any of the provided formats: %v", formats)
}

// TryParseWithMultipleFormats 尝试使用多种Java风格格式解析时间字符串
func (t *timeEngine) TryParseWithMultipleFormats(value string, formats []string) stdtime.Time {
	parsed, err := t.ParseWithMultipleFormats(value, formats)
	if err != nil {
		return stdtime.Time{}
	}
	return parsed
}

// ParseAuto 自动识别并解析常见格式的时间字符串
func (t *timeEngine) ParseAuto(value string) (stdtime.Time, error) {
	// 首先尝试Go标准格式
	if parsed, err := stdtime.Parse(stdtime.RFC3339, value); err == nil {
		return parsed, nil
	}
	if parsed, err := stdtime.Parse("2006-01-02 15:04:05", value); err == nil {
		return parsed, nil
	}
	if parsed, err := stdtime.Parse("2006-01-02", value); err == nil {
		return parsed, nil
	}

	// 然后尝试Java常用格式
	return t.ParseWithMultipleFormats(value, commonJavaFormats)
}

// TryParseAuto 自动识别并解析常见格式的时间字符串，失败返回零值
func (t *timeEngine) TryParseAuto(value string) stdtime.Time {
	parsed, err := t.ParseAuto(value)
	if err != nil {
		return stdtime.Time{}
	}
	return parsed
}

// =========================================
// 时间计算方法
// =========================================

// Add 增加时间
func (t *timeEngine) Add(tim stdtime.Time, d stdtime.Duration) stdtime.Time {
	return tim.Add(d)
}

// AddDuration 增加指定时长的时间
func (t *timeEngine) AddDuration(tim stdtime.Time, years, months, days int) stdtime.Time {
	return tim.AddDate(years, months, days)
}

// AddYears 增加指定年数
func (t *timeEngine) AddYears(tim stdtime.Time, years int) stdtime.Time {
	return tim.AddDate(years, 0, 0)
}

// AddMonths 增加指定月数
func (t *timeEngine) AddMonths(tim stdtime.Time, months int) stdtime.Time {
	return tim.AddDate(0, months, 0)
}

// AddDays 增加指定天数
func (t *timeEngine) AddDays(tim stdtime.Time, days int) stdtime.Time {
	return tim.AddDate(0, 0, days)
}

// Sub 计算时间差
func (t *timeEngine) Sub(tim, other stdtime.Time) stdtime.Duration {
	return tim.Sub(other)
}

// DurationBetween 计算两个时间之间的持续时间的毫秒数
func (t *timeEngine) DurationBetween(tim, other stdtime.Time) int64 {
	return tim.Sub(other).Milliseconds()
}

// DaysBetween 计算两个时间之间的天数
func (t *timeEngine) DaysBetween(tim, other stdtime.Time) int {
	duration := tim.Sub(other)
	if duration < 0 {
		duration = -duration
	}
	return int(duration.Hours() / 24)
}

// =========================================
// 时间转换方法
// =========================================

// StartOfDay 获取一天的开始时间（00:00:00）
func (t *timeEngine) StartOfDay(tim stdtime.Time) stdtime.Time {
	year, month, day := tim.Date()
	return stdtime.Date(year, month, day, 0, 0, 0, 0, tim.Location())
}

// EndOfDay 获取一天的结束时间（23:59:59.999999999）
func (t *timeEngine) EndOfDay(tim stdtime.Time) stdtime.Time {
	year, month, day := tim.Date()
	return stdtime.Date(year, month, day, 23, 59, 59, 999999999, tim.Location())
}

// StartOfWeek 获取一周的开始时间（周一 00:00:00）
func (t *timeEngine) StartOfWeek(tim stdtime.Time) stdtime.Time {
	weekday := tim.Weekday()
	if weekday == stdtime.Sunday {
		weekday = 7
	}
	return t.AddDays(t.StartOfDay(tim), -int(weekday)+1)
}

// EndOfWeek 获取一周的结束时间（周日 23:59:59.999999999）
func (t *timeEngine) EndOfWeek(tim stdtime.Time) stdtime.Time {
	return t.AddDays(t.EndOfDay(t.StartOfWeek(tim)), 6)
}

// StartOfMonth 获取一个月的开始时间（1日 00:00:00）
func (t *timeEngine) StartOfMonth(tim stdtime.Time) stdtime.Time {
	year, month, _ := tim.Date()
	return stdtime.Date(year, month, 1, 0, 0, 0, 0, tim.Location())
}

// EndOfMonth 获取一个月的结束时间（最后一天 23:59:59.999999999）
func (t *timeEngine) EndOfMonth(tim stdtime.Time) stdtime.Time {
	startOfMonth := t.StartOfMonth(tim)
	return t.AddDays(startOfMonth.AddDate(0, 1, 0), -1).
		Add(23*stdtime.Hour + 59*stdtime.Minute + 59*stdtime.Second + 999999999*stdtime.Nanosecond)
}

// StartOfYear 获取一年的开始时间（1月1日 00:00:00）
func (t *timeEngine) StartOfYear(tim stdtime.Time) stdtime.Time {
	year, _, _ := tim.Date()
	return stdtime.Date(year, 1, 1, 0, 0, 0, 0, tim.Location())
}

// EndOfYear 获取一年的结束时间（12月31日 23:59:59.999999999）
func (t *timeEngine) EndOfYear(tim stdtime.Time) stdtime.Time {
	startOfYear := t.StartOfYear(tim)
	return startOfYear.AddDate(1, 0, 0).Add(-stdtime.Nanosecond)
}

// =========================================
// 时间工具方法
// =========================================

// IsLeapYear 检查是否为闰年
func (t *timeEngine) IsLeapYear(year int) bool {
	if year%4 != 0 {
		return false
	} else if year%100 != 0 {
		return true
	} else {
		return year%400 == 0
	}
}

// DaysInMonth 获取指定月份的天数
func (t *timeEngine) DaysInMonth(year, month int) int {
	if month == 2 {
		if t.IsLeapYear(year) {
			return 29
		}
		return 28
	}

	// 4, 6, 9, 11 月有30天
	if month == 4 || month == 6 || month == 9 || month == 11 {
		return 30
	}

	// 其他月份有31天
	return 31
}

// FormatDuration 格式化持续时间
func (t *timeEngine) FormatDuration(d stdtime.Duration) string {
	d = d.Abs()
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// Age 计算指定日期到现在的年龄
func (t *timeEngine) Age(birthDate stdtime.Time) int {
	now := t.Now()
	years := now.Year() - birthDate.Year()

	// 如果生日还没到，年龄减1
	if now.Month() < birthDate.Month() || (now.Month() == birthDate.Month() && now.Day() < birthDate.Day()) {
		years--
	}

	return years
}

// Between 检查时间是否在指定范围内
func (t *timeEngine) Between(tim, start, end stdtime.Time) bool {
	return (tim.After(start) || tim.Equal(start)) && (tim.Before(end) || tim.Equal(end))
}

// Truncate 截断时间到指定精度
func (t *timeEngine) Truncate(tim stdtime.Time, d stdtime.Duration) stdtime.Time {
	return tim.Truncate(d)
}

// Round 四舍五入时间到指定精度
func (t *timeEngine) Round(tim stdtime.Time, d stdtime.Duration) stdtime.Time {
	return tim.Round(d)
}

// =========================================
// 时间字符串快速格式化
// =========================================

// ToYYYYMMDD 格式化为 yyyyMMdd
func (t *timeEngine) ToYYYYMMDD(tim stdtime.Time) string {
	return t.FormatWithJava(tim, "yyyyMMdd")
}

// ToYYYYMMDDHHMMSS 格式化为 yyyyMMddHHmmss
func (t *timeEngine) ToYYYYMMDDHHMMSS(tim stdtime.Time) string {
	return t.FormatWithJava(tim, "yyyyMMddHHmmss")
}

// ToYYYY_MM_DD 格式化为 yyyy-MM-dd
func (t *timeEngine) ToYYYY_MM_DD(tim stdtime.Time) string {
	return t.FormatWithJava(tim, "yyyy-MM-dd")
}

// ToYYYY_MM_DD_HH_MM_SS 格式化为 yyyy-MM-dd HH:mm:ss
func (t *timeEngine) ToYYYY_MM_DD_HH_MM_SS(tim stdtime.Time) string {
	return t.FormatWithJava(tim, "yyyy-MM-dd HH:mm:ss")
}

// ToHHMMSS 格式化为 HH:mm:ss
func (t *timeEngine) ToHHMMSS(tim stdtime.Time) string {
	return t.FormatWithJava(tim, "HH:mm:ss")
}

// =========================================
// 时间字符串快速解析
// =========================================

// FromYYYYMMDD 从 yyyyMMdd 格式字符串解析时间
func (t *timeEngine) FromYYYYMMDD(value string) (stdtime.Time, error) {
	return t.ParseWithJava(value, "yyyyMMdd")
}

// FromYYYYMMDDHHMMSS 从 yyyyMMddHHmmss 格式字符串解析时间
func (t *timeEngine) FromYYYYMMDDHHMMSS(value string) (stdtime.Time, error) {
	return t.ParseWithJava(value, "yyyyMMddHHmmss")
}

// FromYYYY_MM_DD 从 yyyy-MM-dd 格式字符串解析时间
func (t *timeEngine) FromYYYY_MM_DD(value string) (stdtime.Time, error) {
	return t.ParseWithJava(value, "yyyy-MM-dd")
}

// FromYYYY_MM_DD_HH_MM_SS 从 yyyy-MM-dd HH:mm:ss 格式字符串解析时间
func (t *timeEngine) FromYYYY_MM_DD_HH_MM_SS(value string) (stdtime.Time, error) {
	return t.ParseWithJava(value, "yyyy-MM-dd HH:mm:ss")
}

// =========================================
// 时区相关方法
// =========================================

// InLocation 将时间转换到指定时区
func (t *timeEngine) InLocation(tim stdtime.Time, loc *stdtime.Location) stdtime.Time {
	return tim.In(loc)
}

// UTC 将时间转换到UTC时区
func (t *timeEngine) UTC(tim stdtime.Time) stdtime.Time {
	return tim.UTC()
}

// Local 将时间转换到本地时区
func (t *timeEngine) Local(tim stdtime.Time) stdtime.Time {
	return tim.Local()
}

// LoadLocation 加载时区
func (t *timeEngine) LoadLocation(name string) (*stdtime.Location, error) {
	return stdtime.LoadLocation(name)
}

// =========================================
// 时间戳相关方法
// =========================================

// ToUnix 转换为Unix时间戳（秒）
func (t *timeEngine) ToUnix(tim stdtime.Time) int64 {
	return tim.Unix()
}

// ToUnixMilli 转换为Unix时间戳（毫秒）
func (t *timeEngine) ToUnixMilli(tim stdtime.Time) int64 {
	return tim.UnixMilli()
}

// FromUnix 从Unix时间戳（秒）创建时间
func (t *timeEngine) FromUnix(sec int64) stdtime.Time {
	return stdtime.Unix(sec, 0)
}

// FromUnixMilli 从Unix时间戳（毫秒）创建时间
func (t *timeEngine) FromUnixMilli(msec int64) stdtime.Time {
	return stdtime.Unix(msec/1000, (msec%1000)*1000000)
}

// =========================================
// 验证和辅助方法
// =========================================

// IsValidFormat 检查是否为有效的Java日期格式
func (t *timeEngine) IsValidFormat(javaFormat string) bool {
	// 检查是否包含有效的格式模式
	validPatterns := []string{"yyyy", "yy", "MM", "M", "dd", "d", "HH", "H", "mm", "m", "ss", "s", "SSS", "SS", "S"}
	hasValidPattern := false

	for _, pattern := range validPatterns {
		if strings.Contains(javaFormat, pattern) {
			hasValidPattern = true
			break
		}
	}

	// 如果不包含有效模式，则为无效格式
	if !hasValidPattern {
		return false
	}

	// 检查是否为完全无效的字符串
	if javaFormat == "invalid" || javaFormat == "" {
		return false
	}

	return true
}

// GuessFormat 尝试猜测时间字符串的格式
func (t *timeEngine) GuessFormat(value string) string {
	// 首先检查包含分隔符的格式
	switch {
	case strings.Contains(value, "-") && strings.Contains(value, ":"):
		if strings.Contains(value, " ") {
			return "yyyy-MM-dd HH:mm:ss"
		}
		return "yyyy-MM-dd"
	case strings.Contains(value, "/") && strings.Contains(value, ":"):
		if strings.Contains(value, " ") {
			return "yyyy/MM/dd HH:mm:ss"
		}
		return "yyyy/MM/dd"
	case strings.Contains(value, ":"):
		if len(strings.Split(value, ":")) == 3 {
			return "HH:mm:ss"
		}
		return "HH:mm"
	case strings.Contains(value, "-"):
		return "yyyy-MM-dd"
	case strings.Contains(value, "/"):
		return "yyyy/MM/dd"
	}

	// 然后检查纯数字格式
	cleanValue := regexp.MustCompile(`[^0-9A-Za-z]`).ReplaceAllString(value, "")
	length := len(cleanValue)

	switch {
	case length == 8 && regexp.MustCompile(`^\d{8}$`).MatchString(cleanValue):
		return "yyyyMMdd"
	case length == 14 && regexp.MustCompile(`^\d{14}$`).MatchString(cleanValue):
		return "yyyyMMddHHmmss"
	case length == 17 && regexp.MustCompile(`^\d{17}$`).MatchString(cleanValue):
		return "yyyyMMddHHmmssSSS"
	}

	return "yyyy-MM-dd" // 默认格式
}
