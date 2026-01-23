package logger

// Field 日志字段类型（用于结构化日志）
type Field = any

// F 创建日志字段（便捷函数）
func F(key string, value any) Field {
	return []any{key, value}
}
