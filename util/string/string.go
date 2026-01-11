package string

import (
	"strings"
	"unicode"

	"github.com/duke-git/lancet/v2/strutil"
)

// ILiteUtilString 字符串工具接口
type ILiteUtilString interface {
	// 基础检查
	IsEmpty(str string) bool
	IsNotEmpty(str string) bool
	IsBlank(str string) bool
	IsNotBlank(str string) bool

	// 修剪和分割
	Trim(str string) string
	TrimLeft(str string) string
	TrimRight(str string) string
	Split(str, sep string) []string
	Join(elements []string, sep string) string
	SplitAndTrim(str, delimiter string, characterMask ...string) []string

	// 子串操作
	SubString(str string, offset int, length uint) string
	SubBetween(str, start, end string) string
	Contains(str, substr string) bool
	ContainsAny(str string, chars []string) bool
	HasPrefix(str, prefix string) bool
	HasSuffix(str, suffix string) bool

	// 大小写转换
	Uppercase(str string) string
	Lowercase(str string) string
	Capitalize(str string) string
	UpperFirst(str string) string
	LowerFirst(str string) string

	// 命名转换
	ToCamelCase(str string) string
	ToKebabCase(str string) string
	ToSnakeCase(str string) string
	UpperSnakeCase(str string) string
	UpperKebabCase(str string) string

	// 填充和对齐
	PadStart(str string, size int, padStr string) string
	PadEnd(str string, size int, padStr string) string
	Pad(str string, size int, padStr string) string

	// 高级操作
	Reverse(str string) string
	WordCount(str string) int
	ShuffleString(str string) string
	HammingDistance(a, b string) (int, error)

	// 位置和提取
	Before(str, substr string) string
	After(str, substr string) string
	BeforeLast(str, substr string) string
	AfterLast(str, substr string) string

	// 验证
	IsNumeric(str string) bool
	IsAlpha(str string) bool
	IsAlphaNumeric(str string) bool
	IsString(v any) bool

	// 前缀后缀
	HasPrefixAny(str string, prefixes []string) bool
	HasSuffixAny(str string, suffixes []string) bool
	ContainsAll(str string, substrs []string) bool

	// 分割和处理
	SplitEx(str, sep string, removeEmptyString bool) []string
	SplitWords(str string) []string

	// 包装和替换
	Wrap(str string, wrapWith string) string
	Unwrap(str string, wrapToken string) string
	Rotate(str string, shift int) string
	RemoveWhiteSpace(str string, removeAll bool) string
	RemoveNonPrintable(str string) string
	HideString(origin string, start, end int, replaceChar string) string
	Ellipsis(str string, length int) string
	TemplateReplace(template string, data map[string]string) string
	ReplaceWithMap(str string, replaces map[string]string) string

	// 查找和提取
	ExtractContent(str, start, end string) []string
	FindAllOccurrences(str, substr string) []int
	IndexOffset(str string, substr string, idxFrom int) int

	// 连接和转换
	Concat(length int, str ...string) string
	StringToBytes(str string) []byte
	BytesToString(bytes []byte) string
}

// stringEngine 字符串操作工具实现
type stringEngine struct{}

// String 默认的字符串操作实例
var String = &stringEngine{}

// 默认实例
var defaultInstance ILiteUtilString = newStringEngine()

// Default 返回默认实例
func newStringEngine() ILiteUtilString {
	return &stringEngine{}
}

// Deprecated: 请使用 liteutil.LiteUtil().String() 来获取字符串工具实例

// New 创建新的字符串工具实例
// Deprecated: 请使用 liteutil.LiteUtil().NewStringOperation() 来创建新的字符串工具实例

// IsEmpty 检查字符串是否为空
func (s *stringEngine) IsEmpty(str string) bool {
	return str == ""
}

// IsNotEmpty 检查字符串是否不为空
func (s *stringEngine) IsNotEmpty(str string) bool {
	return str != ""
}

// IsBlank 检查字符串是否为空白（只包含空格、制表符等）
func (s *stringEngine) IsBlank(str string) bool {
	return strutil.IsBlank(str)
}

// IsNotBlank 检查字符串是否不为空白
func (s *stringEngine) IsNotBlank(str string) bool {
	return strutil.IsNotBlank(str)
}

// Trim 去除字符串两端的空白字符
func (s *stringEngine) Trim(str string) string {
	return strutil.Trim(str)
}

// TrimLeft 去除字符串左端的空白字符
func (s *stringEngine) TrimLeft(str string) string {
	return strings.TrimLeft(str, " \t\n\r")
}

// TrimRight 去除字符串右端的空白字符
func (s *stringEngine) TrimRight(str string) string {
	return strings.TrimRight(str, " \t\n\r")
}

// Contains 检查字符串是否包含子字符串
func (s *stringEngine) Contains(str, substr string) bool {
	return strings.Contains(str, substr)
}

// ContainsAny 检查字符串是否包含任意一个字符
func (s *stringEngine) ContainsAny(str string, chars []string) bool {
	return strutil.ContainsAny(str, chars)
}

// HasPrefix 检查字符串是否以指定前缀开头
func (s *stringEngine) HasPrefix(str, prefix string) bool {
	return strings.HasPrefix(str, prefix)
}

// HasSuffix 检查字符串是否以指定后缀结尾
func (s *stringEngine) HasSuffix(str, suffix string) bool {
	return strings.HasSuffix(str, suffix)
}

// Split 分割字符串
func (s *stringEngine) Split(str, sep string) []string {
	return strings.Split(str, sep)
}

// Join 连接字符串数组
func (s *stringEngine) Join(elements []string, sep string) string {
	return strings.Join(elements, sep)
}

// SplitAndTrim 分割并去除每个元素的空白字符
func (s *stringEngine) SplitAndTrim(str, delimiter string, characterMask ...string) []string {
	return strutil.SplitAndTrim(str, delimiter, characterMask...)
}

// SubString 截取子字符串
func (s *stringEngine) SubString(str string, offset int, length uint) string {
	return strutil.Substring(str, offset, length)
}

// SubBetween 提取两个字符串之间的子字符串
func (s *stringEngine) SubBetween(str, start, end string) string {
	return strutil.SubInBetween(str, start, end)
}

// ToCamelCase 转换为驼峰命名
func (s *stringEngine) ToCamelCase(str string) string {
	return strutil.CamelCase(str)
}

// ToKebabCase 转换为短横线命名
func (s *stringEngine) ToKebabCase(str string) string {
	return strutil.KebabCase(str)
}

// ToSnakeCase 转换为下划线命名
func (s *stringEngine) ToSnakeCase(str string) string {
	return strutil.SnakeCase(str)
}

// UpperFirst 首字母大写
func (s *stringEngine) UpperFirst(str string) string {
	return strutil.UpperFirst(str)
}

// LowerFirst 首字母小写
func (s *stringEngine) LowerFirst(str string) string {
	return strutil.LowerFirst(str)
}

// PadStart 左侧填充字符
func (s *stringEngine) PadStart(str string, size int, padStr string) string {
	return strutil.PadStart(str, size, padStr)
}

// PadEnd 右侧填充字符
func (s *stringEngine) PadEnd(str string, size int, padStr string) string {
	return strutil.PadEnd(str, size, padStr)
}

// Pad 两侧填充字符
func (s *stringEngine) Pad(str string, size int, padStr string) string {
	return strutil.Pad(str, size, padStr)
}

// Reverse 反转字符串
func (s *stringEngine) Reverse(str string) string {
	return strutil.Reverse(str)
}

// WordCount 统计单词数量
func (s *stringEngine) WordCount(str string) int {
	return strutil.WordCount(str)
}

// Uppercase 转换为大写
func (s *stringEngine) Uppercase(str string) string {
	return strings.ToUpper(str)
}

// Lowercase 转换为小写
func (s *stringEngine) Lowercase(str string) string {
	return strings.ToLower(str)
}

// ShuffleString 打乱字符串
func (s *stringEngine) ShuffleString(str string) string {
	return strutil.Shuffle(str)
}

// HammingDistance 计算汉明距离
func (s *stringEngine) HammingDistance(a, b string) (int, error) {
	return strutil.HammingDistance(a, b)
}

// Before 获取子字符串第一次出现之前的部分
func (s *stringEngine) Before(str, substr string) string {
	return strutil.Before(str, substr)
}

// After 获取子字符串第一次出现之后的部分
func (s *stringEngine) After(str, substr string) string {
	return strutil.After(str, substr)
}

// BeforeLast 获取子字符串最后一次出现之前的部分
func (s *stringEngine) BeforeLast(str, substr string) string {
	return strutil.BeforeLast(str, substr)
}

// AfterLast 获取子字符串最后一次出现之后的部分
func (s *stringEngine) AfterLast(str, substr string) string {
	return strutil.AfterLast(str, substr)
}

// IsNumeric 检查字符串是否只包含数字
func (s *stringEngine) IsNumeric(str string) bool {
	for _, r := range str {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(str) > 0
}

// IsAlpha 检查字符串是否只包含字母
func (s *stringEngine) IsAlpha(str string) bool {
	for _, r := range str {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return len(str) > 0
}

// IsAlphaNumeric 检查字符串是否只包含字母和数字
func (s *stringEngine) IsAlphaNumeric(str string) bool {
	for _, r := range str {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return len(str) > 0
}

// Capitalize 首字母大写，其余小写
func (s *stringEngine) Capitalize(str string) string {
	return strutil.Capitalize(str)
}

// HasPrefixAny 检查字符串是否以任意一个前缀开头
func (s *stringEngine) HasPrefixAny(str string, prefixes []string) bool {
	return strutil.HasPrefixAny(str, prefixes)
}

// HasSuffixAny 检查字符串是否以任意一个后缀结尾
func (s *stringEngine) HasSuffixAny(str string, suffixes []string) bool {
	return strutil.HasSuffixAny(str, suffixes)
}

// ContainsAll 检查字符串是否包含所有子字符串
func (s *stringEngine) ContainsAll(str string, substrs []string) bool {
	return strutil.ContainsAll(str, substrs)
}

// SplitEx 分割字符串，可选择是否移除空字符串
func (s *stringEngine) SplitEx(str, sep string, removeEmptyString bool) []string {
	return strutil.SplitEx(str, sep, removeEmptyString)
}

// SplitWords 按单词分割字符串
func (s *stringEngine) SplitWords(str string) []string {
	return strutil.SplitWords(str)
}

// Wrap 用指定字符串包裹原字符串
func (s *stringEngine) Wrap(str string, wrapWith string) string {
	return strutil.Wrap(str, wrapWith)
}

// Unwrap 移除指定包裹字符串
func (s *stringEngine) Unwrap(str string, wrapToken string) string {
	return strutil.Unwrap(str, wrapToken)
}

// Rotate 旋转字符串
func (s *stringEngine) Rotate(str string, shift int) string {
	return strutil.Rotate(str, shift)
}

// RemoveWhiteSpace 移除空白字符
func (s *stringEngine) RemoveWhiteSpace(str string, removeAll bool) string {
	return strutil.RemoveWhiteSpace(str, removeAll)
}

// RemoveNonPrintable 移除非打印字符
func (s *stringEngine) RemoveNonPrintable(str string) string {
	return strutil.RemoveNonPrintable(str)
}

// HideString 隐藏字符串部分内容
func (s *stringEngine) HideString(origin string, start, end int, replaceChar string) string {
	return strutil.HideString(origin, start, end, replaceChar)
}

// Ellipsis 添加省略号
func (s *stringEngine) Ellipsis(str string, length int) string {
	return strutil.Ellipsis(str, length)
}

// TemplateReplace 模板替换
func (s *stringEngine) TemplateReplace(template string, data map[string]string) string {
	return strutil.TemplateReplace(template, data)
}

// ReplaceWithMap 使用映射表替换字符串
func (s *stringEngine) ReplaceWithMap(str string, replaces map[string]string) string {
	return strutil.ReplaceWithMap(str, replaces)
}

// ExtractContent 提取内容
func (s *stringEngine) ExtractContent(str, start, end string) []string {
	return strutil.ExtractContent(str, start, end)
}

// FindAllOccurrences 查找所有出现位置
func (s *stringEngine) FindAllOccurrences(str, substr string) []int {
	return strutil.FindAllOccurrences(str, substr)
}

// IndexOffset 从指定偏移量开始查找
func (s *stringEngine) IndexOffset(str string, substr string, idxFrom int) int {
	return strutil.IndexOffset(str, substr, idxFrom)
}

// IsString 检查是否为字符串类型
func (s *stringEngine) IsString(v any) bool {
	return strutil.IsString(v)
}

// Concat 连接字符串
func (s *stringEngine) Concat(length int, str ...string) string {
	return strutil.Concat(length, str...)
}

// UpperSnakeCase 转换为大写下划线命名
func (s *stringEngine) UpperSnakeCase(str string) string {
	return strutil.UpperSnakeCase(str)
}

// UpperKebabCase 转换为大写短横线命名
func (s *stringEngine) UpperKebabCase(str string) string {
	return strutil.UpperKebabCase(str)
}

// StringToBytes 字符串转字节
func (s *stringEngine) StringToBytes(str string) []byte {
	return strutil.StringToBytes(str)
}

// BytesToString 字节转字符串
func (s *stringEngine) BytesToString(bytes []byte) string {
	return strutil.BytesToString(bytes)
}
