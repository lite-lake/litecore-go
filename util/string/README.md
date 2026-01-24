# util/string

字符串处理工具集，基于 lancet 库提供丰富的字符串操作功能。

## 特性

- **命名转换** - 支持 CamelCase、SnakeCase、KebabCase 及其大写形式之间的相互转换
- **基础检查** - 提供字符串为空、空白、数字、字母、字母数字等验证功能
- **修剪与分割** - 支持去除空白字符、灵活分割与连接字符串
- **子串操作** - 提供截取、查找、前缀后缀匹配等子串处理功能
- **大小写转换** - 支持大小写转换、首字母大小写调整
- **填充与对齐** - 支持字符串左右两侧填充、对齐操作
- **高级操作** - 提供反转、打乱、汉明距离计算、隐藏内容等高级功能
- **零拷贝转换** - 提供 StringToBytes 和 BytesToString 零拷贝转换

## 快速开始

```go
import "github.com/lite-lake/litecore-go/util/string"

// 命名转换
camel := string.String.ToCamelCase("hello_world")
fmt.Println(camel) // 输出: helloWorld

// 基础检查
isEmpty := string.String.IsEmpty("")
fmt.Println(isEmpty) // 输出: true

// 修剪与分割
trimmed := string.String.Trim("  hello  ")
fmt.Println(trimmed) // 输出: hello
```

## 命名转换

支持常见的命名风格转换，适用于变量名、函数名、配置键等场景。

```go
// 下划线转驼峰
camel := string.String.ToCamelCase("hello_world")
// 输出: helloWorld

// 驼峰转短横线
kebab := string.String.ToKebabCase("helloWorld")
// 输出: hello-world

// 驼峰转下划线
snake := string.String.ToSnakeCase("helloWorld")
// 输出: hello_world

// 大写形式
upperSnake := string.String.UpperSnakeCase("helloWorld")
// 输出: HELLO_WORLD
```

## 字符串验证

提供多种字符串验证方法，用于数据校验和输入过滤。

```go
// 基础验证
empty := string.String.IsEmpty("")           // true
notEmpty := string.String.IsNotEmpty("abc")  // true
blank := string.String.IsBlank("   ")        // true

// 字符内容验证
numeric := string.String.IsNumeric("12345")     // true
alpha := string.String.IsAlpha("hello")         // true
alphaNum := string.String.IsAlphaNumeric("abc123")  // true
```

## 修剪与分割

去除空白字符和灵活分割字符串，适用于数据清洗和解析。

```go
// 修剪空白
trimmed := string.String.Trim("  hello  ")         // "hello"
trimmedLeft := string.String.TrimLeft("  hello")    // "hello"
trimmedRight := string.String.TrimRight("hello  ")  // "hello"

// 分割字符串
parts := string.String.Split("a,b,c", ",")         // []string{"a", "b", "c"}
joined := string.String.Join([]string{"a", "b"}, ",")  // "a,b"

// 高级分割
trimmedParts := string.String.SplitAndTrim("a, b , c", ",")  // []string{"a", "b", "c"}
words := string.String.SplitWords("hello  world")            // []string{"hello", "world"}
```

## 子串操作

查找、截取、提取子字符串，满足各种字符串处理需求。

```go
// 包含检查
contains := string.String.Contains("hello", "ell")          // true
containsAny := string.String.ContainsAny("test", []string{"a", "e"})  // true
containsAll := string.String.ContainsAll("test", []string{"te", "st"}) // true

// 前缀后缀
hasPrefix := string.String.HasPrefix("hello", "he")         // true
hasSuffix := string.String.HasSuffix("hello", "lo")         // true
hasPrefixAny := string.String.HasPrefixAny("hello", []string{"he", "hi"})  // true

// 截取子串
sub := string.String.SubString("hello", 1, 3)               // "ell"
between := string.String.SubBetween("<a>content</a>", "<a>", "</a>")  // "content"

// 位置提取
before := string.String.Before("key:value", ":")            // "key"
after := string.String.After("key:value", ":")              // "value"
```

## API

### 接口定义

```go
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
```

### 使用方式

通过默认实例 `string.String` 调用方法：

```go
result := string.String.MethodName(params)
```

## 测试

```bash
go test ./util/string
go test ./util/string -cover
```

## 依赖

- [github.com/duke-git/lancet/v2](https://github.com/duke-git/lancet) - 提供底层字符串操作实现
