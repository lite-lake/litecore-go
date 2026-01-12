// Package container 提供分层依赖注入容器，严格约束架构分层并管理组件生命周期。
//
// 核心特性：
//   - 分层架构：定义 Config/Entity/Manager/Repository/Service/Controller/Middleware 七层容器
//   - 单向依赖：上层可依赖下层，下层不能依赖上层，禁止跨层访问
//   - 依赖注入：通过 inject 标签自动注入依赖，支持接口类型匹配
//   - 同层依赖：Manager 和 Service 层支持同层依赖，自动拓扑排序确定注入顺序
//   - 按类型注册：使用接口类型作为索引，每个接口类型只能注册一个实现
//   - 错误检测：自动检测循环依赖、依赖缺失、接口未实现等错误
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
//	// 2. 注册实例（按接口类型注册）
//	var appConfig *AppConfig = &AppConfig{}
//	configContainer.RegisterByType(reflect.TypeOf((*BaseConfigProvider)(nil)).Elem(), appConfig)
//
//	var dbManager *DatabaseManager = &DatabaseManager{}
//	managerContainer.RegisterByType(reflect.TypeOf((*DatabaseManager)(nil)).Elem(), dbManager)
//
//	// 3. 执行依赖注入（按层次从下到上）
//	managerContainer.InjectAll()
//	repositoryContainer.InjectAll()
//	serviceContainer.InjectAll()
//	controllerContainer.InjectAll()
//
//	// 4. 获取实例使用
//	userService, err := serviceContainer.GetByType(reflect.TypeOf((*UserService)(nil)).Elem())
//
// 依赖声明：
//
//	// 在结构体中使用 inject 标签声明依赖
//	type UserServiceImpl struct {
//		Config     BaseConfigProvider `inject:""`
//		DBManager  DatabaseManager    `inject:""`
//		UserRepo   UserRepository     `inject:""`
//		OrderSvc   OrderService       `inject:""`  // 同层依赖
//		CacheMgr   CacheManager       `inject:"optional"` // 可选依赖
//	}
//
// 同层依赖：
//
//	Service 和 Manager 层支持同层依赖，容器会自动构建依赖图并进行拓扑排序。
//
// 例如：UserService 依赖 OrderService，OrderService 依赖 PaymentService，
// 容器会按 [PaymentService, OrderService, UserService] 的顺序注入。
//
// 注册规则：
//
//	1. 按接口类型注册：使用 RegisterByType(ifaceType, impl) 注册实例
//	2. 接口唯一性：每个接口类型只能注册一个实现
//	3. 实现校验：注册时会检查实现是否真正实现了接口
//	4. 并发安全：RegisterByType 使用写锁，GetByType 使用读锁
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
//
// 错误处理：
//
//	InjectAll 可能返回以下错误：
//	- DependencyNotFoundError: 标记了 inject 的字段无法找到匹配的依赖
//	- CircularDependencyError: 同层实例之间存在循环依赖
//	- InterfaceAlreadyRegisteredError: 尝试重复注册相同接口
//	- ImplementationDoesNotImplementInterfaceError: 实现未实现指定接口
package container
