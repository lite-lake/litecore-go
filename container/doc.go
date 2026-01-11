// Package container 提供分层依赖注入容器，严格约束架构分层并管理组件生命周期。
//
// 核心特性：
//   - 分层架构：定义 Config/Entity/Manager/Repository/Service/Controller/Middleware 七层容器
//   - 单向依赖：上层可依赖下层，下层不能依赖上层，禁止跨层访问
//   - 依赖注入：通过 inject 标签自动注入依赖，支持接口匹配
//   - 同层依赖：Manager 和 Service 层支持同层依赖，自动拓扑排序确定注入顺序
//   - 错误检测：自动检测循环依赖、依赖缺失、多重匹配、重复注册等错误
//   - 并发安全：容器内部使用 RWMutex 保护，支持多线程并发读取
//
// 基本用法：
//
//	// 1. 创建容器（按依赖顺序）
//	configContainer := container.NewConfigContainer()
//	managerContainer := container.NewManagerContainer(configContainer)
//	entityContainer := container.NewEntityContainer()
//	repositoryContainer := container.NewRepositoryContainer(configContainer, managerContainer, entityContainer)
//	serviceContainer := container.NewServiceContainer(configContainer, managerContainer, repositoryContainer)
//	controllerContainer := container.NewControllerContainer(configContainer, managerContainer, serviceContainer)
//
//	// 2. 注册实例（可按任意顺序）
//	configContainer.Register(&AppConfig{})
//	managerContainer.Register(&DatabaseManager{})
//	repositoryContainer.Register(&UserRepositoryImpl{})
//	serviceContainer.Register(&UserServiceImpl{})
//	controllerContainer.Register(&UserControllerImpl{})
//
//	// 3. 执行依赖注入（按层次从下到上）
//	managerContainer.InjectAll()
//	repositoryContainer.InjectAll()
//	serviceContainer.InjectAll()
//	controllerContainer.InjectAll()
//
//	// 4. 获取实例使用
//	userCtrl, _ := controllerContainer.GetByName("user")
//
// 依赖声明：
//
//	// 在结构体中使用 inject 标签声明依赖
//	type UserServiceImpl struct {
//		Config     BaseConfigProvider `inject:""`
//		DBManager  DatabaseManager    `inject:""`
//		UserRepo   UserRepository     `inject:""`
//		OrderSvc   OrderService       `inject:""`  // 同层依赖
//	}
//
// 同层依赖：
//
//	Service 和 Manager 层支持同层依赖，容器会自动构建依赖图并进行拓扑排序。
//
// 例如：UserService 依赖 OrderService，OrderService 依赖 PaymentService，
// 容器会按 [PaymentService, OrderService, UserService] 的顺序注入。
//
// 错误处理：
//
//	InjectAll 可能返回以下错误：
//	- DependencyNotFoundError: 标记了 inject 的字段无法找到匹配的依赖
//	- CircularDependencyError: 同层实例之间存在循环依赖
//	- AmbiguousMatchError: 字段类型匹配了多个实例
//	- DuplicateRegistrationError: 尝试注册相同名称的实例
//
// 容器层次：
//
//	Config     (无依赖)
//	Entity     (无依赖)
//	Manager    → Config, 其他 Manager
//	Repository → Config, Manager, Entity
//	Service    → Config, Manager, Repository, 其他 Service
//	Controller → Config, Manager, Service
//	Middleware → Config, Manager, Service
package container
