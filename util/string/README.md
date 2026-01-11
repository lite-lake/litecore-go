# util/string

字符串处理工具集，基于 lancet 库提供丰富的字符串操作功能。

## 特性

- **命名转换**：支持驼峰命名(CamelCase)、蛇形命名(SnakeCase)、短横线命名(KebabCase)之间的相互转换
- **分割与连接**：提供灵活的字符串分割与连接操作，支持去空格、移除空字符串等选项
- **字符串修剪**：去除字符串首尾的空白字符（空格、制表符、换行符等）
- **查找与验证**：检查子串存在性、前缀后缀匹配、字符类型验证（数字、字母、字母数字）
- **替换与模板**：支持字符串替换、模板替换、映射表批量替换
- **截取与提取**：按位置或长度截取子串，提取标记之间的内容
- **格式化操作**：字符串填充、对齐、大小写转换、反转等
- **高级功能**：汉明距离计算、字符串打乱、内容隐藏、省略号截断等

## 快速开始

### 导入包

```go
import "litecore-go/util/string"
```

### 基本使用

```go
// 使用默认实例
result := string.String.ToCamelCase("hello_world")
fmt.Println(result) // 输出: helloWorld

// 字符串分割与连接
parts := string.String.Split("a,b,c", ",")
fmt.Println(parts) // 输出: [a b c]

joined := string.String.Join([]string{"a", "b", "c"}, ",")
fmt.Println(joined) // 输出: a,b,c

// 字符串修剪
trimmed := string.String.Trim("  hello  ")
fmt.Println(trimmed) // 输出: hello
```

## 功能说明

### 命名转换

将不同命名风格的字符串进行相互转换。

```go
// 转换为驼峰命名 (CamelCase)
camel := string.String.ToCamelCase("hello_world")
fmt.Println(camel) // 输出: helloWorld

camel = string.String.ToCamelCase("hello-world")
fmt.Println(camel) // 输出: helloWorld

// 转换为蛇形命名 (SnakeCase)
snake := string.String.ToSnakeCase("HelloWorld")
fmt.Println(snake) // 输出: hello_world

// 转换为大写蛇形命名
upperSnake := string.String.UpperSnakeCase("HelloWorld")
fmt.Println(upperSnake) // 输出: HELLO_WORLD

// 转换为短横线命名 (KebabCase)
kebab := string.String.ToKebabCase("HelloWorld")
fmt.Println(kebab) // 输出: hello-world

// 转换为大写短横线命名
upperKebab := string.String.UpperKebabCase("HelloWorld")
fmt.Println(upperKebab) // 输出: HELLO-WORLD
```

### 分割与连接

对字符串进行分割和连接操作。

```go
// 基本分割
parts := string.String.Split("a,b,c", ",")
fmt.Println(parts) // 输出: [a b c]

// 基本连接
joined := string.String.Join([]string{"a", "b", "c"}, ",")
fmt.Println(joined) // 输出: a,b,c

// 分割并去除每个元素的首尾空白
trimmedParts := string.String.SplitAndTrim("a, b , c", ",")
fmt.Println(trimmedParts) // 输出: [a b c]

// 高级分割（可选择是否移除空字符串）
partsNoEmpty := string.String.SplitEx("a,,b,c", ",", true)
fmt.Println(partsNoEmpty) // 输出: [a b c]

// 按单词分割
words := string.String.SplitWords("hello  world test")
fmt.Println(words) // 输出: [hello world test]

// 连接字符串（length 参数会被忽略）
result := string.String.Concat(3, "a", "b", "c")
fmt.Println(result) // 输出: abc
```

### 字符串修剪

去除字符串首尾的空白字符。

```go
// 去除两侧空白
trimmed := string.String.Trim("  hello  ")
fmt.Println(trimmed) // 输出: hello

// 去除左侧空白
trimmedLeft := string.String.TrimLeft("  hello  ")
fmt.Println(trimmedLeft) // 输出: hello

// 去除右侧空白
trimmedRight := string.String.TrimRight("  hello  ")
fmt.Println(trimmedRight) // 输出:   hello
```

### 查找与验证

检查字符串的属性和内容。

```go
// 检查是否包含子串
contains := string.String.Contains("hello world", "world")
fmt.Println(contains) // 输出: true

// 检查是否包含任意一个子串
containsAny := string.String.ContainsAny("hello world", []string{"foo", "world", "bar"})
fmt.Println(containsAny) // 输出: true

// 检查是否包含所有子串
containsAll := string.String.ContainsAll("hello world", []string{"hello", "world"})
fmt.Println(containsAll) // 输出: true

// 检查前缀
hasPrefix := string.String.HasPrefix("hello world", "hello")
fmt.Println(hasPrefix) // 输出: true

// 检查后缀
hasSuffix := string.String.HasSuffix("hello world", "world")
fmt.Println(hasSuffix) // 输出: true

// 检查是否以任意一个前缀开头
hasPrefixAny := string.String.HasPrefixAny("hello.world", []string{"foo", "hello", "bar"})
fmt.Println(hasPrefixAny) // 输出: true

// 检查是否以任意一个后缀结尾
hasSuffixAny := string.String.HasSuffixAny("hello.world", []string{".com", ".world", ".org"})
fmt.Println(hasSuffixAny) // 输出: true

// 检查字符串是否为空
isEmpty := string.String.IsEmpty("")
fmt.Println(isEmpty) // 输出: true

// 检查字符串是否为空白（只包含空格、制表符等）
isBlank := string.String.IsBlank("   \t\n")
fmt.Println(isBlank) // 输出: true

// 检查是否只包含数字
isNumeric := string.String.IsNumeric("12345")
fmt.Println(isNumeric) // 输出: true

// 检查是否只包含字母
isAlpha := string.String.IsAlpha("hello")
fmt.Println(isAlpha) // 输出: true

// 检查是否只包含字母和数字
isAlphaNumeric := string.String.IsAlphaNumeric("hello123")
fmt.Println(isAlphaNumeric) // 输出: true
```

### 截取与提取

从字符串中提取特定部分。

```go
// 按位置和长度截取子串
sub := string.String.SubString("hello world", 6, 5)
fmt.Println(sub) // 输出: world

// 提取两个标记之间的内容
between := string.String.SubBetween("hello [world] test", "[", "]")
fmt.Println(between) // 输出: world

// 获取子串第一次出现之前的部分
before := string.String.Before("hello.world.test", ".")
fmt.Println(before) // 输出: hello

// 获取子串第一次出现之后的部分
after := string.String.After("hello.world.test", ".")
fmt.Println(after) // 输出: world.test

// 获取子串最后一次出现之前的部分
beforeLast := string.String.BeforeLast("hello.world.test", ".")
fmt.Println(beforeLast) // 输出: hello.world

// 获取子串最后一次出现之后的部分
afterLast := string.String.AfterLast("hello.world.test", ".")
fmt.Println(afterLast) // 输出: test

// 提取所有匹配的内容
contents := string.String.ExtractContent("hello [world] test [example]", "[", "]")
fmt.Println(contents) // 输出: [world example]

// 查找所有出现位置
positions := string.String.FindAllOccurrences("hello world hello", "hello")
fmt.Println(positions) // 输出: [0 12]

// 从指定位置开始查找
index := string.String.IndexOffset("hello world hello", "hello", 6)
fmt.Println(index) // 输出: 12
```

### 大小写转换

转换字符串的大小写。

```go
// 转换为大写
upper := string.String.Uppercase("hello")
fmt.Println(upper) // 输出: HELLO

// 转换为小写
lower := string.String.Lowercase("HELLO")
fmt.Println(lower) // 输出: hello

// 首字母大写（其余小写）
capitalized := string.String.Capitalize("hello")
fmt.Println(capitalized) // 输出: Hello

// 首字母大写（其余不变）
upperFirst := string.String.UpperFirst("hello world")
fmt.Println(upperFirst) // 输出: Hello world

// 首字母小写
lowerFirst := string.String.LowerFirst("Hello World")
fmt.Println(lowerFirst) // 输出: hello World
```

### 填充与对齐

对字符串进行填充和对齐操作。

```go
// 左侧填充
paddedStart := string.String.PadStart("hello", 10, "*")
fmt.Println(paddedStart) // 输出: *****hello

// 右侧填充
paddedEnd := string.String.PadEnd("hello", 10, "*")
fmt.Println(paddedEnd) // 输出: hello*****

// 两侧填充
padded := string.String.Pad("hello", 11, "*")
fmt.Println(padded) // 输出: ***hello***
```

### 替换与模板

对字符串进行替换操作。

```go
// 使用映射表替换
replaced := string.String.ReplaceWithMap("hello world", map[string]string{
    "hello": "hi",
    "world": "earth",
})
fmt.Println(replaced) // 输出: hi earth

// 模板替换（使用 {key} 作为占位符）
template := "Hello {name}, you are {age} years old"
data := map[string]string{
    "name": "John",
    "age":  "30",
}
result := string.String.TemplateReplace(template, data)
fmt.Println(result) // 输出: Hello {John}, you are {30} years old
```

### 格式化操作

对字符串进行各种格式化操作。

```go
// 反转字符串
reversed := string.String.Reverse("hello")
fmt.Println(reversed) // 输出: olleh

// 统计单词数量
count := string.String.WordCount("hello world test")
fmt.Println(count) // 输出: 3

// 打乱字符串
shuffled := string.String.ShuffleString("hello")
fmt.Println(shuffled) // 输出: 随机排列，如 "olleh"

// 旋转字符串
rotated := string.String.Rotate("hello", 2)
fmt.Println(rotated) // 输出: lohel

// 移除空白字符（全部移除）
noSpace := string.String.RemoveWhiteSpace("hello world", true)
fmt.Println(noSpace) // 输出: helloworld

// 移除空白字符（合并为单个空格）
mergedSpace := string.String.RemoveWhiteSpace("hello   world", false)
fmt.Println(mergedSpace) // 输出: hello world

// 移除非打印字符
clean := string.String.RemoveNonPrintable("hello\x00world")
fmt.Println(clean) // 输出: helloworld

// 隐藏字符串部分内容
hidden := string.String.HideString("1234567890", 3, 6, "*")
fmt.Println(hidden) // 输出: 123***7890

// 添加省略号
ellipsis := string.String.Ellipsis("hello world", 5)
fmt.Println(ellipsis) // 输出: hello...

// 包装字符串
wrapped := string.String.Wrap("hello", "*")
fmt.Println(wrapped) // 输出: *hello*

// 解包字符串
unwrapped := string.String.Unwrap("*hello*", "*")
fmt.Println(unwrapped) // 输出: hello
```

### 高级功能

使用字符串的高级操作功能。

```go
// 计算汉明距离（两个等长字符串的不同字符数）
distance, err := string.String.HammingDistance("hello", "hallo")
if err == nil {
    fmt.Println(distance) // 输出: 1
}

// 检查是否为字符串类型
isString := string.String.IsString("hello")
fmt.Println(isString) // 输出: true

isString = string.String.IsString(123)
fmt.Println(isString) // 输出: false

// 字符串转字节（零拷贝）
bytes := string.String.StringToBytes("hello")
fmt.Println(bytes) // 输出: [104 101 108 108 111]

// 字节转字符串（零拷贝）
str := string.String.BytesToString([]byte{104, 101, 108, 108, 111})
fmt.Println(str) // 输出: hello
```

## API 参考

### 基础检查

| 函数 | 说明 | 返回值 |
|------|------|--------|
| `IsEmpty(str string) bool` | 检查字符串是否为空 | 是否为空 |
| `IsNotEmpty(str string) bool` | 检查字符串是否不为空 | 是否不为空 |
| `IsBlank(str string) bool` | 检查字符串是否为空白 | 是否为空白 |
| `IsNotBlank(str string) bool` | 检查字符串是否不为空白 | 是否不为空白 |

### 修剪和分割

| 函数 | 说明 | 返回值 |
|------|------|--------|
| `Trim(str string) string` | 去除字符串两端的空白字符 | 修剪后的字符串 |
| `TrimLeft(str string) string` | 去除字符串左端的空白字符 | 修剪后的字符串 |
| `TrimRight(str string) string` | 去除字符串右端的空白字符 | 修剪后的字符串 |
| `Split(str, sep string) []string` | 分割字符串 | 字符串数组 |
| `Join(elements []string, sep string) string` | 连接字符串数组 | 连接后的字符串 |
| `SplitAndTrim(str, delimiter string, characterMask ...string) []string` | 分割并去除每个元素的空白 | 字符串数组 |

### 子串操作

| 函数 | 说明 | 返回值 |
|------|------|--------|
| `SubString(str string, offset int, length uint) string` | 截取子字符串 | 子字符串 |
| `SubBetween(str, start, end string) string` | 提取两个字符串之间的内容 | 提取的内容 |
| `Contains(str, substr string) bool` | 检查是否包含子字符串 | 是否包含 |
| `ContainsAny(str string, chars []string) bool` | 检查是否包含任意一个字符 | 是否包含 |
| `HasPrefix(str, prefix string) bool` | 检查是否以指定前缀开头 | 是否有前缀 |
| `HasSuffix(str, suffix string) bool` | 检查是否以指定后缀结尾 | 是否有后缀 |

### 大小写转换

| 函数 | 说明 | 返回值 |
|------|------|--------|
| `Uppercase(str string) string` | 转换为大写 | 大写字符串 |
| `Lowercase(str string) string` | 转换为小写 | 小写字符串 |
| `Capitalize(str string) string` | 首字母大写，其余小写 | 首字母大写的字符串 |
| `UpperFirst(str string) string` | 首字母大写 | 首字母大写的字符串 |
| `LowerFirst(str string) string` | 首字母小写 | 首字母小写的字符串 |

### 命名转换

| 函数 | 说明 | 返回值 |
|------|------|--------|
| `ToCamelCase(str string) string` | 转换为驼峰命名 | 驼峰命名字符串 |
| `ToKebabCase(str string) string` | 转换为短横线命名 | 短横线命名字符串 |
| `ToSnakeCase(str string) string` | 转换为下划线命名 | 下划线命名字符串 |
| `UpperSnakeCase(str string) string` | 转换为大写下划线命名 | 大写下划线命名字符串 |
| `UpperKebabCase(str string) string` | 转换为大写短横线命名 | 大写短横线命名字符串 |

### 填充和对齐

| 函数 | 说明 | 返回值 |
|------|------|--------|
| `PadStart(str string, size int, padStr string) string` | 左侧填充字符 | 填充后的字符串 |
| `PadEnd(str string, size int, padStr string) string` | 右侧填充字符 | 填充后的字符串 |
| `Pad(str string, size int, padStr string) string` | 两侧填充字符 | 填充后的字符串 |

### 高级操作

| 函数 | 说明 | 返回值 |
|------|------|--------|
| `Reverse(str string) string` | 反转字符串 | 反转后的字符串 |
| `WordCount(str string) int` | 统计单词数量 | 单词数 |
| `ShuffleString(str string) string` | 打乱字符串 | 打乱后的字符串 |
| `HammingDistance(a, b string) (int, error)` | 计算汉明距离 | 汉明距离和错误 |

### 位置和提取

| 函数 | 说明 | 返回值 |
|------|------|--------|
| `Before(str, substr string) string` | 获取子串第一次出现之前的部分 | 之前的字符串 |
| `After(str, substr string) string` | 获取子串第一次出现之后的部分 | 之后的字符串 |
| `BeforeLast(str, substr string) string` | 获取子串最后一次出现之前的部分 | 之前的字符串 |
| `AfterLast(str, substr string) string` | 获取子串最后一次出现之后的部分 | 之后的字符串 |

### 验证

| 函数 | 说明 | 返回值 |
|------|------|--------|
| `IsNumeric(str string) bool` | 检查是否只包含数字 | 是否为数字 |
| `IsAlpha(str string) bool` | 检查是否只包含字母 | 是否为字母 |
| `IsAlphaNumeric(str string) bool` | 检查是否只包含字母和数字 | 是否为字母数字 |
| `IsString(v any) bool` | 检查是否为字符串类型 | 是否为字符串 |

### 前缀后缀

| 函数 | 说明 | 返回值 |
|------|------|--------|
| `HasPrefixAny(str string, prefixes []string) bool` | 检查是否以任意一个前缀开头 | 是否有前缀 |
| `HasSuffixAny(str string, suffixes []string) bool` | 检查是否以任意一个后缀结尾 | 是否有后缀 |
| `ContainsAll(str string, substrs []string) bool` | 检查是否包含所有子字符串 | 是否全部包含 |

### 分割和处理

| 函数 | 说明 | 返回值 |
|------|------|--------|
| `SplitEx(str, sep string, removeEmptyString bool) []string` | 分割字符串，可选择是否移除空字符串 | 字符串数组 |
| `SplitWords(str string) []string` | 按单词分割字符串 | 单词数组 |

### 包装和替换

| 函数 | 说明 | 返回值 |
|------|------|--------|
| `Wrap(str string, wrapWith string) string` | 用指定字符串包裹原字符串 | 包裹后的字符串 |
| `Unwrap(str string, wrapToken string) string` | 移除指定包裹字符串 | 解包后的字符串 |
| `Rotate(str string, shift int) string` | 旋转字符串 | 旋转后的字符串 |
| `RemoveWhiteSpace(str string, removeAll bool) string` | 移除空白字符 | 处理后的字符串 |
| `RemoveNonPrintable(str string) string` | 移除非打印字符 | 清理后的字符串 |
| `HideString(origin string, start, end int, replaceChar string) string` | 隐藏字符串部分内容 | 隐藏后的字符串 |
| `Ellipsis(str string, length int) string` | 添加省略号 | 处理后的字符串 |
| `TemplateReplace(template string, data map[string]string) string` | 模板替换 | 替换后的字符串 |
| `ReplaceWithMap(str string, replaces map[string]string) string` | 使用映射表替换 | 替换后的字符串 |

### 查找和提取

| 函数 | 说明 | 返回值 |
|------|------|--------|
| `ExtractContent(str, start, end string) []string` | 提取所有标记之间的内容 | 提取的内容数组 |
| `FindAllOccurrences(str, substr string) []int` | 查找所有出现位置 | 位置数组 |
| `IndexOffset(str string, substr string, idxFrom int) int` | 从指定偏移量开始查找 | 位置（-1表示未找到） |

### 连接和转换

| 函数 | 说明 | 返回值 |
|------|------|--------|
| `Concat(length int, str ...string) string` | 连接字符串 | 连接后的字符串 |
| `StringToBytes(str string) []byte` | 字符串转字节（零拷贝） | 字节数组 |
| `BytesToString(bytes []byte) string` | 字节转字符串（零拷贝） | 字符串 |

## 注意事项

1. **错误处理**：大部分函数对于非法输入会返回空字符串或零值，不会返回错误。建议在使用前验证输入参数的有效性。
2. **字符编码**：所有函数都支持 UTF-8 编码，可以正确处理中文字符。
3. **性能考虑**：`StringToBytes` 和 `BytesToString` 函数使用零拷贝技术，性能优于标准库的转换方法。
4. **模板语法**：`TemplateReplace` 函数使用 `{key}` 作为占位符，而非标准的 `{{key}}`。
5. **填充行为**：填充函数在字符串长度已达到或超过指定长度时，不会进行截断，而是返回原字符串。
6. **汉明距离**：`HammingDistance` 函数要求两个字符串长度相同，否则会返回错误。

## 测试

运行测试：

```bash
go test ./util/string
```

运行测试并查看覆盖率：

```bash
go test ./util/string -cover
```

## 依赖

- [github.com/duke-git/lancet/v2](https://github.com/duke-git/lancet) - 提供底层字符串操作实现

## 许可证

本项目采用 MIT 许可证。
