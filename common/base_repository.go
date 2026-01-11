package common

// BaseRepository 存储库基类接口
// 所有 Repository 类必须继承此接口并实现 RepositoryName 方法
// 系统通过此接口判断是否符合标准存储层定义
type BaseRepository interface {
	// RepositoryName 返回当前存储库实现的类名
	// 用于标识和调试存储库实例
	RepositoryName() string

	// OnStart 在服务器启动时触发
	OnStart() error
	// OnStop 在服务器停止时触发
	OnStop() error
}
