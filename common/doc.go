// Package common 提供七层架构的基础接口定义，规范 Entity、Manager、Repository、Service、Controller、Middleware 和 ConfigMgr 的行为契约。
//
// 核心特性：
//   - 七层架构基础接口：定义各层的基础接口类型，确保架构一致性
//   - 生命周期管理：提供统一的 OnStart 和 OnStop 钩子方法
//   - 命名规范：每层接口要求实现对应的名称方法，便于调试和日志
//   - 行为契约：通过接口定义各层的核心行为，建立分层依赖关系
//   - 依赖注入支持：为依赖注入容器提供标准接口类型
//
// 基本用法：
//
//	// 实现 Entity 接口
//	type User struct {
//		ID   string `gorm:"primaryKey"`
//		Name string
//	}
//
//	func (u *User) EntityName() string { return "User" }
//	func (u *User) TableName() string { return "users" }
//	func (u *User) GetId() string { return u.ID }
//
//	// 实现 Service 接口
//	type UserService struct {}
//
//	func (s *UserService) ServiceName() string { return "UserService" }
//	func (s *UserService) OnStart() error { return nil }
//	func (s *UserService) OnStop() error { return nil }
//
// 接口层次：
//
//	各层之间有明确的依赖关系，从低到高依次为：
//	ConfigMgr → Entity → Manager → Repository → Service → Controller/Middleware
//	上层可以依赖下层，下层不能依赖上层。
package common
