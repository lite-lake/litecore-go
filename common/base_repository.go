package common

// BaseRepository 存储库基类接口
// 所有 Repository 类必须继承此接口并实现 GetRepositoryName 方法
// 系统通过此接口判断是否符合标准存储层定义
type BaseRepository interface {
	// GetRepositoryName 返回当前存储库实现的类名
	// 用于标识和调试存储库实例
	GetRepositoryName() string
}
