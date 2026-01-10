package common

// BaseEntity 实体基类接口
// 所有 Entity 类必须继承此接口并实现 GetEntityName 和 GetId 方法
// 系统通过此接口判断是否符合标准实体定义
type BaseEntity interface {
	// GetEntityName 返回当前实体实现的类名
	// 用于标识和调试实体实例
	GetEntityName() string

	// TableName 返回当前实体的表名
	// 用于数据库操作
	TableName() string

	// GetId 返回实体的唯一标识
	// 用于实体的索引和检索
	GetId() string
}
