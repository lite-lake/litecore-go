# util/string

字符串处理工具集，基于 lancet 库提供丰富的字符串操作功能。

## 特性

- **命名转换**：支持 CamelCase、SnakeCase、KebabCase 及其大写形式之间的相互转换
- **基础检查**：提供字符串为空、空白、数字、字母、字母数字等验证功能
- **修剪与分割**：支持去除空白字符、灵活分割与连接字符串
- **子串操作**：提供截取、查找、前缀后缀匹配等子串处理功能
- **大小写转换**：支持大小写转换、首字母大小写调整
- **填充与对齐**：支持字符串左右两侧填充、对齐操作
- **高级操作**：提供反转、打乱、汉明距离计算、隐藏内容等高级功能
- **零拷贝转换**：提供 StringToBytes 和 BytesToString 零拷贝转换

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

## API

### 基础检查

| 函数 | 说明 |
|------|------|
| `IsEmpty(str string) bool` | 检查字符串是否为空 |
| `IsNotEmpty(str string) bool` | 检查字符串是否不为空 |
| `IsBlank(str string) bool` | 检查字符串是否为空白（只包含空格、制表符等） |
| `IsNotBlank(str string) bool` | 检查字符串是否不为空白 |
| `IsNumeric(str string) bool` | 检查是否只包含数字 |
| `IsAlpha(str string) bool` | 检查是否只包含字母 |
| `IsAlphaNumeric(str string) bool` | 检查是否只包含字母和数字 |
| `IsString(v any) bool` | 检查是否为字符串类型 |

### 修剪与分割

| 函数 | 说明 |
|------|------|
| `Trim(str string) string` | 去除字符串两端的空白字符 |
| `TrimLeft(str string) string` | 去除字符串左端的空白字符 |
| `TrimRight(str string) string` | 去除字符串右端的空白字符 |
| `Split(str, sep string) []string` | 分割字符串 |
| `Join(elements []string, sep string) string` | 连接字符串数组 |
| `SplitAndTrim(str, delimiter string, characterMask ...string) []string` | 分割并去除每个元素的空白字符 |
| `SplitEx(str, sep string, removeEmptyString bool) []string` | 分割字符串，可选择是否移除空字符串 |
| `SplitWords(str string) []string` | 按单词分割字符串 |

### 子串操作

| 函数 | 说明 |
|------|------|
| `SubString(str string, offset int, length uint) string` | 截取子字符串 |
| `SubBetween(str, start, end string) string` | 提取两个字符串之间的内容 |
| `Contains(str, substr string) bool` | 检查是否包含子字符串 |
| `ContainsAny(str string, chars []string) bool` | 检查是否包含任意一个字符 |
| `ContainsAll(str string, substrs []string) bool` | 检查是否包含所有子字符串 |
| `HasPrefix(str, prefix string) bool` | 检查是否以指定前缀开头 |
| `HasSuffix(str, suffix string) bool` | 检查是否以指定后缀结尾 |
| `HasPrefixAny(str string, prefixes []string) bool` | 检查是否以任意一个前缀开头 |
| `HasSuffixAny(str string, suffixes []string) bool` | 检查是否以任意一个后缀结尾 |
| `Before(str, substr string) string` | 获取子字符串第一次出现之前的部分 |
| `After(str, substr string) string` | 获取子字符串第一次出现之后的部分 |
| `BeforeLast(str, substr string) string` | 获取子字符串最后一次出现之前的部分 |
| `AfterLast(str, substr string) string` | 获取子字符串最后一次出现之后的部分 |
| `ExtractContent(str, start, end string) []string` | 提取所有标记之间的内容 |
| `FindAllOccurrences(str, substr string) []int` | 查找所有出现位置 |
| `IndexOffset(str string, substr string, idxFrom int) int` | 从指定偏移量开始查找 |

### 大小写转换

| 函数 | 说明 |
|------|------|
| `Uppercase(str string) string` | 转换为大写 |
| `Lowercase(str string) string` | 转换为小写 |
| `Capitalize(str string) string` | 首字母大写，其余小写 |
| `UpperFirst(str string) string` | 首字母大写 |
| `LowerFirst(str string) string` | 首字母小写 |

### 命名转换

| 函数 | 说明 |
|------|------|
| `ToCamelCase(str string) string` | 转换为驼峰命名 |
| `ToKebabCase(str string) string` | 转换为短横线命名 |
| `ToSnakeCase(str string) string` | 转换为下划线命名 |
| `UpperSnakeCase(str string) string` | 转换为大写下划线命名 |
| `UpperKebabCase(str string) string` | 转换为大写短横线命名 |

### 填充与对齐

| 函数 | 说明 |
|------|------|
| `PadStart(str string, size int, padStr string) string` | 左侧填充字符 |
| `PadEnd(str string, size int, padStr string) string` | 右侧填充字符 |
| `Pad(str string, size int, padStr string) string` | 两侧填充字符 |

### 高级操作

| 函数 | 说明 |
|------|------|
| `Reverse(str string) string` | 反转字符串 |
| `WordCount(str string) int` | 统计单词数量 |
| `ShuffleString(str string) string` | 打乱字符串 |
| `HammingDistance(a, b string) (int, error)` | 计算汉明距离 |
| `Rotate(str string, shift int) string` | 旋转字符串 |
| `RemoveWhiteSpace(str string, removeAll bool) string` | 移除空白字符 |
| `RemoveNonPrintable(str string) string` | 移除非打印字符 |
| `HideString(origin string, start, end int, replaceChar string) string` | 隐藏字符串部分内容 |
| `Ellipsis(str string, length int) string` | 添加省略号 |
| `Wrap(str string, wrapWith string) string` | 用指定字符串包裹原字符串 |
| `Unwrap(str string, wrapToken string) string` | 移除指定包裹字符串 |
| `TemplateReplace(template string, data map[string]string) string` | 模板替换 |
| `ReplaceWithMap(str string, replaces map[string]string) string` | 使用映射表替换 |

### 连接与转换

| 函数 | 说明 |
|------|------|
| `Concat(length int, str ...string) string` | 连接字符串 |
| `StringToBytes(str string) []byte` | 字符串转字节（零拷贝） |
| `BytesToString(bytes []byte) string` | 字节转字符串（零拷贝） |

## 测试

```bash
go test ./util/string
go test ./util/string -cover
```

## 依赖

- [github.com/duke-git/lancet/v2](https://github.com/duke-git/lancet) - 提供底层字符串操作实现
