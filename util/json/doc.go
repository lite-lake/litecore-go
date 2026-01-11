// Package json 提供全面的 JSON 操作工具集，包括验证、格式化、数据转换、路径查询等功能
package json

/*
核心特性：
	• JSON 验证与格式化：快速验证 JSON 有效性，支持美化和压缩输出
	• 数据类型转换：在 JSON 字符串、Map 和结构体之间灵活转换
	• 路径查询：使用点号语法访问嵌套字段，支持对象和数组索引
	• 高级操作：支持 JSON 对象合并、差异比较等复杂操作
	• 实用工具：类型检测、键值查询、大小获取等辅助功能

基本用法：

	import "litecore-go/util/json"

	// 获取 JSON 工具实例（推荐使用 liteutil.LiteUtil().Json()）
	j := json.Default()

	// 验证 JSON 有效性
	if j.IsValid(jsonStr) {
		// JSON 有效
	}

	// 格式化 JSON（美化输出）
	formatted, err := j.PrettyPrint(jsonStr, "  ")
	if err != nil {
		// 处理错误
	}

	// 压缩 JSON（移除空白）
	compacted, err := j.Compact(jsonStr)
	if err != nil {
		// 处理错误
	}

	// JSON 转换为 Map
	data, err := j.ToMap(jsonStr)
	if err != nil {
		// 处理错误
	}

	// JSON 转换为结构体
	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	var user User
	err = j.ToStruct(jsonStr, &user)
	if err != nil {
		// 处理错误
	}

	// 结构体转换为 JSON
	jsonBytes, err := j.FromStruct(user)
	if err != nil {
		// 处理错误
	}

路径操作语法：

	// 使用点号访问嵌套字段
	value, err := j.GetValue(jsonStr, "user.profile.age")

	// 使用数组索引（从 0 开始）
	firstItem, err := j.GetValue(jsonStr, "items.0.name")

	// 获取指定类型的值
	name, err := j.GetString(jsonStr, "user.name")
	price, err := j.GetFloat64(jsonStr, "product.price")
	isActive, err := j.GetBool(jsonStr, "user.active")

高级操作：

	// 合并两个 JSON 对象（后者覆盖前者）
	merged, err := j.Merge(jsonStr1, jsonStr2)
	if err != nil {
		// 处理错误
	}

	// 比较两个 JSON 是否不同
	diff, err := j.Diff(jsonStr1, jsonStr2)
	if err != nil {
		// 处理错误
	}
	if diff {
		// JSON 存在差异
	}

实用工具：

	// 检查 JSON 类型
	if j.IsObject(jsonStr) {
		// 是 JSON 对象
	}
	if j.IsArray(jsonStr) {
		// 是 JSON 数组
	}

	// 获取字段类型
	typ, err := j.GetType(jsonStr, "user.age")

	// 获取对象的所有键
	keys, err := j.GetKeys(jsonStr, "user")

	// 获取数组或对象的大小
	size, err := j.GetSize(jsonStr, "items")

	// 检查对象是否包含指定键
	exists, err := j.Contains(jsonStr, "user", "email")

注意事项：

	• 所有路径操作使用点号（.）分隔符，支持对象字段和数组索引
	• 数组索引使用非负整数，从 0 开始计数
	• 路径查询时如字段不存在或索引越界会返回错误
	• Merge 操作仅支持 JSON 对象类型，不合并数组
	• 类型转换函数（GetString、GetFloat64、GetBool）会尝试自动类型转换
*/
