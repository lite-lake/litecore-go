# 容器机制设计

## 设计目标

本容器机制的核心目标是 **严格约束架构分层，禁止不规范调用关系**。通过分层容器确保：

1. **单向依赖**：上层可以依赖下层，下层不能依赖上层
2. **隔离性**：各层容器独立管理，防止跨层直接访问
3. **可维护性**：清晰的依赖边界，便于代码审查和重构

## 架构分层

系统定义以下层级（接口定义在 common 包）：

```
┌─────────────────────────────────────────────────────────────┐
│  Controller Layer   (BaseController)                        │
│  Middleware Layer   (BaseMiddleware)                        │
├─────────────────────────────────────────────────────────────┤
│  Service Layer      (BaseService)                           │
├─────────────────────────────────────────────────────────────┤
│  Repository Layer   (BaseRepository)                        │
├─────────────────────────────────────────────────────────────┤
│  Entity Layer       (BaseEntity)                            │
│  Manager Layer      (BaseManager)                           │
├─────────────────────────────────────────────────────────────┤
│  Config Layer       (BaseConfigProvider)                    │
└─────────────────────────────────────────────────────────────┘
```

**依赖关系说明**：

| 层级       | 接口               | 可依赖的层级                              |
| ---------- | ------------------ | ----------------------------------------- |
| Config     | BaseConfigProvider | 无依赖                                    |
| Entity     | BaseEntity         | 无依赖                                    |
| Manager    | BaseManager        | Config, 其他 Manager                      |
| Repository | BaseRepository     | Config, Manager, Entity                   |
| Service    | BaseService        | Config, Manager, Repository, 其他 Service |
| Controller | BaseController     | Config, Manager, Service                  |
| Middleware | BaseMiddleware     | Config, Manager, Service                  |

所有接口均定义 `XXName() string` 方法，用于获取对象实例名称。

## 容器定义

### ConfigContainer

```go
type ConfigContainer struct {
    mu    sync.RWMutex
    items map[string]BaseConfigProvider
}

func New() *ConfigContainer
func (c *ConfigContainer) Register(ins BaseConfigProvider) error
func (c *ConfigContainer) InjectAll() error
func (c *ConfigContainer) GetAll() []BaseConfigProvider
func (c *ConfigContainer) GetByName(name string) (BaseConfigProvider, error)
func (c *ConfigContainer) GetByType(typ reflect.Type) ([]BaseConfigProvider, error)
```

**说明**：Config 层无依赖，`InjectAll` 为空操作。

### ManagerContainer

```go
type ManagerContainer struct {
    mu               sync.RWMutex
    items            map[string]BaseManager
    configContainer  *ConfigContainer
    injected         bool // 标记是否已执行注入
}

func New(config *ConfigContainer) *ManagerContainer
func (m *ManagerContainer) Register(ins BaseManager) error
func (m *ManagerContainer) InjectAll() error
func (m *ManagerContainer) GetAll() []BaseManager
func (m *ManagerContainer) GetByName(name string) (BaseManager, error)
func (m *ManagerContainer) GetByType(typ reflect.Type) ([]BaseManager, error)
```

**InjectAll 行为**：

1. 注入 `BaseConfigProvider`（从 ConfigContainer 获取）
2. 注入其他 `BaseManager`（支持同层依赖，按拓扑顺序注入）

### EntityContainer

```go
type EntityContainer struct {
    mu    sync.RWMutex
    items map[string]BaseEntity
}

func New() *EntityContainer
func (e *EntityContainer) Register(ins BaseEntity) error
func (e *EntityContainer) InjectAll() error
func (e *EntityContainer) GetAll() []BaseEntity
func (e *EntityContainer) GetByName(name string) (BaseEntity, error)
func (e *EntityContainer) GetByType(typ reflect.Type) ([]BaseEntity, error)
```

**说明**：Entity 层无依赖，`InjectAll` 为空操作。

### RepositoryContainer

```go
type RepositoryContainer struct {
    mu                 sync.RWMutex
    items              map[string]BaseRepository
    configContainer    *ConfigContainer
    managerContainer   *ManagerContainer
    entityContainer    *EntityContainer
    injected           bool
}

func New(config *ConfigContainer, manager *ManagerContainer, entity *EntityContainer) *RepositoryContainer
func (r *RepositoryContainer) Register(ins BaseRepository) error
func (r *RepositoryContainer) InjectAll() error
func (r *RepositoryContainer) GetAll() []BaseRepository
func (r *RepositoryContainer) GetByName(name string) (BaseRepository, error)
func (r *RepositoryContainer) GetByType(typ reflect.Type) ([]BaseRepository, error)
```

**InjectAll 行为**：

1. 注入 `BaseConfigProvider`
2. 注入 `BaseManager`
3. 注入 `BaseEntity`

### ServiceContainer

```go
type ServiceContainer struct {
    mu                   sync.RWMutex
    items                map[string]BaseService
    configContainer      *ConfigContainer
    managerContainer     *ManagerContainer
    repositoryContainer  *RepositoryContainer
    injected             bool
}

func New(config *ConfigContainer, manager *ManagerContainer, repository *RepositoryContainer) *ServiceContainer
func (s *ServiceContainer) Register(ins BaseService) error
func (s *ServiceContainer) InjectAll() error
func (s *ServiceContainer) GetAll() []BaseService
func (s *ServiceContainer) GetByName(name string) (BaseService, error)
func (s *ServiceContainer) GetByType(typ reflect.Type) ([]BaseService, error)
```

**InjectAll 行为**：

1. 注入 `BaseConfigProvider`
2. 注入 `BaseManager`
3. 注入 `BaseRepository`
4. 注入其他 `BaseService`（支持同层依赖，按拓扑顺序注入）

### ControllerContainer

```go
type ControllerContainer struct {
    mu                sync.RWMutex
    items             map[string]BaseController
    configContainer   *ConfigContainer
    managerContainer  *ManagerContainer
    serviceContainer  *ServiceContainer
    injected          bool
}

func New(config *ConfigContainer, manager *ManagerContainer, service *ServiceContainer) *ControllerContainer
func (c *ControllerContainer) Register(ins BaseController) error
func (c *ControllerContainer) InjectAll() error
func (c *ControllerContainer) GetAll() []BaseController
func (c *ControllerContainer) GetByName(name string) (BaseController, error)
func (c *ControllerContainer) GetByType(typ reflect.Type) ([]BaseController, error)
```

**InjectAll 行为**：

1. 注入 `BaseConfigProvider`
2. 注入 `BaseManager`
3. 注入 `BaseService`

### MiddlewareContainer

```go
type MiddlewareContainer struct {
    mu                sync.RWMutex
    items             map[string]BaseMiddleware
    configContainer   *ConfigContainer
    managerContainer  *ManagerContainer
    serviceContainer  *ServiceContainer
    injected          bool
}

func New(config *ConfigContainer, manager *ManagerContainer, service *ServiceContainer) *MiddlewareContainer
func (m *MiddlewareContainer) Register(ins BaseMiddleware) error
func (m *MiddlewareContainer) InjectAll() error
func (m *MiddlewareContainer) GetAll() []BaseMiddleware
func (m *MiddlewareContainer) GetByName(name string) (BaseMiddleware, error)
func (m *MiddlewareContainer) GetByType(typ reflect.Type) ([]BaseMiddleware, error)
```

**InjectAll 行为**：

1. 注入 `BaseConfigProvider`
2. 注入 `BaseManager`
3. 注入 `BaseService`

## 依赖注入机制

### 两阶段注入

容器采用 **注册-注入分离** 的两阶段模式：

**阶段 1：注册阶段 (Register)**

- 仅将实例加入容器的 items map
- 不执行任何依赖注入操作
- 可按任意顺序注册
- 注册时检查名称唯一性

**阶段 2：注入阶段 (InjectAll)**

- 遍历容器内所有已注册实例
- 反射解析实例字段，执行依赖注入
- 对于同层依赖（如 Service 依赖 Service），按拓扑顺序注入
- 检测循环依赖和缺失依赖，失败时报错

### 注入规则

依赖注入基于 **接口实现匹配**：

1. **结构体标签**：使用 `inject` 标记需要注入的字段
2. **匹配方式**：根据字段类型查找对应容器中已注册的实例
3. **匹配策略**：
   - 精确匹配：字段类型与注册实例类型完全一致
   - 接口匹配：字段类型是接口，注册实例实现了该接口
   - 唯一性要求：匹配结果必须唯一，否则报错

### 同层依赖与拓扑排序

对于 **Manager 依赖 Manager** 和 **Service 依赖 Service** 的场景，容器使用拓扑排序确定注入顺序。

#### 依赖图构建

```go
// 示例：Service 之间的依赖关系
UserService    → OrderService, PaymentService
OrderService   → InventoryService, PaymentService
PaymentService → (无同层依赖)
InventoryService → (无同层依赖)

// 依赖图：
// UserService ──────> OrderService ────> InventoryService
//     │                    │
//     └────────────────────> PaymentService
```

#### 拓扑排序算法

```go
func (c *ServiceContainer) buildDependencyGraph() (map[string][]string, error) {
    graph := make(map[string][]string) // name -> dependencies

    for name, svc := range c.items {
        var deps []string
        // 通过反射获取 Service 的字段
        val := reflect.ValueOf(svc).Elem()
        typ := val.Type()

        for i := 0; i < val.NumField(); i++ {
            field := typ.Field(i)
            // 只处理标记了 inject 的字段
            if tag := field.Tag.Get("inject"); tag == "" {
                continue
            }

            fieldType := field.Type
            // 判断是否为 BaseService 类型（同层依赖）
            if implementsBaseService(fieldType) {
                // 查找该字段依赖的 Service 实例名称
                depName := findDependencyByName(fieldType, c.items)
                if depName == "" {
                    return nil, &DependencyNotFoundError{
                        ServiceName: name,
                        FieldName:   field.Name,
                        FieldType:   fieldType,
                    }
                }
                deps = append(deps, depName)
            }
        }
        graph[name] = deps
    }

    return graph, nil
}

func (c *ServiceContainer) topologicalSort(graph map[string][]string) ([]string, error) {
    // Kahn 算法
    inDegree := make(map[string]int)
    adjList := make(map[string][]string)

    // 初始化
    for node := range graph {
        inDegree[node] = 0
    }

    // 构建邻接表和入度
    for node, deps := range graph {
        for _, dep := range deps {
            adjList[dep] = append(adjList[dep], node)
            inDegree[node]++
        }
    }

    // 找到所有入度为 0 的节点
    var queue []string
    for node, degree := range inDegree {
        if degree == 0 {
            queue = append(queue, node)
        }
    }

    var result []string
    for len(queue) > 0 {
        // 取出一个节点
        node := queue[0]
        queue = queue[1:]
        result = append(result, node)

        // 减少邻接节点的入度
        for _, neighbor := range adjList[node] {
            inDegree[neighbor]--
            if inDegree[neighbor] == 0 {
                queue = append(queue, neighbor)
            }
        }
    }

    // 检测循环依赖
    if len(result) != len(graph) {
        return nil, &CircularDependencyError{
            RemainingNodes: getRemainingNodes(inDegree),
        }
    }

    return result, nil
}
```

#### 注入执行流程

```go
func (c *ServiceContainer) InjectAll() error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if c.injected {
        return nil // 已注入，跳过
    }

    // 1. 构建依赖图
    graph, err := c.buildDependencyGraph()
    if err != nil {
        return fmt.Errorf("build dependency graph failed: %w", err)
    }

    // 2. 拓扑排序
    order, err := c.topologicalSort(graph)
    if err != nil {
        return fmt.Errorf("topological sort failed: %w", err)
    }

    // 3. 按顺序注入
    for _, name := range order {
        svc := c.items[name]
        if err := c.injectDependencies(svc); err != nil {
            return fmt.Errorf("inject %s failed: %w", name, err)
        }
    }

    c.injected = true
    return nil
}

func (c *ServiceContainer) injectDependencies(svc BaseService) error {
    val := reflect.ValueOf(svc).Elem()
    typ := val.Type()

    for i := 0; i < val.NumField(); i++ {
        field := typ.Field(i)
        fieldVal := val.Field(i)

        if field.Tag.Get("inject") == "" {
            continue
        }

        fieldType := field.Type

        // 根据字段类型从对应容器获取依赖
        dependency, err := c.resolveDependency(fieldType)
        if err != nil {
            return err
        }

        if fieldVal.CanSet() {
            fieldVal.Set(reflect.ValueOf(dependency))
        }
    }

    return nil
}
```

### 代码结构规范

业务模块应遵循以下文件组织：

```
user/
├── entity.go          // 实现 BaseEntity
├── repository.go      // 定义 UserRepository 接口
├── repository_impl.go // 实现 UserRepository 和 BaseRepository
├── service.go         // 定义 UserService 接口
├── service_impl.go    // 实现 UserService 和 BaseService
└── controller.go      // 实现 UserController 和 BaseController
```

### 注入示例

```go
// service_impl.go
type UserServiceImpl struct {
    // 通过标签声明依赖注入
    Config     BaseConfigProvider `inject:""`
    DBManager  DatabaseManager    `inject:""`  // BaseManager 的具体实现
    UserRepo   UserRepository     `inject:""`  // BaseRepository 的子接口
    OrderSvc   OrderService       `inject:""`  // 依赖其他 Service（同层依赖）
}

// 注入执行顺序（假设拓扑排序结果）：
// 1. OrderService (无同层依赖)
// 2. UserService (依赖 OrderService)
```

### 完整注入流程

```
Register 阶段：
┌─────────────────────────────────────┐
│ Register(OrderService)              │
│ Register(UserService)               │
│ Register(PaymentService)            │
│ ← 任意顺序，仅加入 map               │
└─────────────────────────────────────┘
            ↓
InjectAll 阶段：
┌─────────────────────────────────────┐
│ 1. 构建依赖图                        │
│    UserService → OrderService       │
│    OrderService → PaymentService    │
│                                     │
│ 2. 拓扑排序                          │
│    [PaymentService, OrderService,   │
│     UserService]                    │
│                                     │
│ 3. 按顺序注入                        │
│    a. PaymentService (注入跨层依赖)  │
│    b. OrderService (注入跨层 + 同层) │
│    c. UserService (注入跨层 + 同层)  │
│                                     │
│ 4. 检测错误                          │
│    - 循环依赖：报错                  │
│    - 缺失依赖：报错                  │
│    - 多重匹配：报错                  │
└─────────────────────────────────────┘
```

## 错误处理

### 错误类型

```go
// 依赖缺失错误
type DependencyNotFoundError struct {
    ServiceName   string  // 当前服务名称
    FieldName     string  // 缺失依赖的字段名
    FieldType     reflect.Type  // 期望的依赖类型
    ContainerType string  // 应该从哪个容器查找
}

func (e *DependencyNotFoundError) Error() string {
    return fmt.Sprintf("dependency not found for %s.%s: need type %s from %s container",
        e.ServiceName, e.FieldName, e.FieldType, e.ContainerType)
}

// 循环依赖错误
type CircularDependencyError struct {
    Cycle []string  // 循环依赖链
}

func (e *CircularDependencyError) Error() string {
    return fmt.Sprintf("circular dependency detected: %s → %s",
        strings.Join(e.Cycle, " → "), e.Cycle[0])
}

// 多重匹配错误
type AmbiguousMatchError struct {
    ServiceName string
    FieldName   string
    FieldType   reflect.Type
    Candidates  []string  // 匹配的候选实例名称
}

func (e *AmbiguousMatchError) Error() string {
    return fmt.Sprintf("ambiguous match for %s.%s: type %s matches multiple instances: %s",
        e.ServiceName, e.FieldName, e.FieldType, strings.Join(e.Candidates, ", "))
}

// 重复注册错误
type DuplicateRegistrationError struct {
    Name      string
    Existing  interface{}
    New       interface{}
}

func (e *DuplicateRegistrationError) Error() string {
    return fmt.Sprintf("duplicate registration: name '%s' already exists", e.Name)
}
```

### 错误场景与检测

#### 1. 依赖缺失

**场景**：标记了 `inject` 的字段无法在对应容器中找到匹配的实例。

**示例**：

```go
type UserServiceImpl struct {
    PaymentSvc PaymentService `inject:""`  // PaymentService 未注册
}

// 错误信息：
// dependency not found for UserServiceImpl.PaymentSvc: need type PaymentService from ServiceContainer
```

**检测时机**：`InjectAll` 阶段，构建依赖图时

#### 2. 循环依赖

**场景**：同层实例之间形成环形依赖链。

**示例**：

```go
// UserService → OrderService → PaymentService → UserService
type UserServiceImpl struct {
    PaymentSvc PaymentService `inject:""`
}

type OrderServiceImpl struct {
    UserSvc UserService `inject:""`
}

type PaymentServiceImpl struct {
    OrderSvc OrderService `inject:""`
}

// 错误信息：
// circular dependency detected: UserService → OrderService → PaymentService → UserService
```

**检测时机**：`InjectAll` 阶段，拓扑排序时（Kahn 算法无法处理所有节点）

#### 3. 多重匹配

**场景**：字段类型匹配了多个已注册实例。

**示例**：

```go
// 容器中注册了多个实现同一接口的实例
serviceContainer.Register(&OrderServiceImpl{})  // 实现 IOrderService
serviceContainer.Register(&MockOrderService{})  // 也实现 IOrderService

type UserServiceImpl struct {
    OrderSvc IOrderService `inject:""`  // 不知道该注入哪个
}

// 错误信息：
// ambiguous match for UserServiceImpl.OrderSvc: type IOrderService matches multiple instances: OrderService, MockOrderService
```

**检测时机**：`InjectAll` 阶段，解析依赖时

#### 4. 重复注册

**场景**：同一容器中注册了相同名称的实例。

**示例**：

```go
serviceContainer.Register(&UserServiceImpl{})  // name: "user"
serviceContainer.Register(&AnotherUserServiceImpl{})  // name: "user"

// 错误信息：
// duplicate registration: name 'user' already exists
```

**检测时机**：`Register` 阶段

### 处理策略

1. **立即失败**：检测到任何错误立即返回，不继续注入
2. **详细错误信息**：包含服务名、字段名、类型、循环链等完整上下文
3. **修复建议**：错误消息中明确指出问题原因和修复方向
4. **幂等性**：`InjectAll` 可多次调用，已注入的容器跳过

### 错误恢复

```go
if err := serviceContainer.InjectAll(); err != nil {
    var depErr *DependencyNotFoundError
    if errors.As(err, &depErr) {
        log.Fatalf("Missing dependency: %v\n"+
            "Hint: Register %s instance before calling InjectAll", depErr, depErr.FieldType)
    }

    var circErr *CircularDependencyError
    if errors.As(err, &circErr) {
        log.Fatalf("Circular dependency detected: %v\n"+
            "Hint: Break the cycle by refactoring one of the services", circErr)
    }

    log.Fatal(err)
}
```

## 并发安全性

### 访问模式

- **写入阶段**：应用启动时单线程顺序注册，无并发写入
- **读取阶段**：服务运行期间多线程并发读取

### 实现要求

容器内部使用 `sync.RWMutex` 保护：

```go
type Container struct {
    mu    sync.RWMutex
    items map[string]BaseEntity
}

func (c *Container) Register(ins BaseEntity) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    // 注册逻辑
}

func (c *Container) GetByName(name string) (BaseEntity, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    // 查询逻辑
}
```

## 容器生命周期

```
┌─────────────────────────────────────────────────────────────┐
│  阶段 1: 创建容器                                            │
│  - 按依赖顺序创建各层容器                                     │
│  - 容器间建立引用关系                                         │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  阶段 2: 注册实例 (Register)                                 │
│  - 按任意顺序注册实例到各容器                                 │
│  - 仅加入 items map，不执行注入                               │
│  - 检查名称唯一性                                             │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  阶段 3: 依赖注入 (InjectAll)                                │
│  - 按层次从下到上执行注入                                     │
│  - Config/Entity: 跳过（无依赖）                             │
│  - Manager: 注入 Config + 同层拓扑排序                       │
│  - Repository: 注入 Config/Manager/Entity                   │
│  - Service: 注入 Config/Manager/Repository + 同层拓扑排序    │
│  - Controller/Middleware: 注入下层依赖                       │
│  - 检测循环依赖、缺失依赖、多重匹配                           │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  阶段 4: 运行时                                              │
│  - 容器只读，并发安全                                         │
│  - 通过 GetByName/GetByType 获取实例                         │
│  - 业务代码使用已注入的实例                                   │
└─────────────────────────────────────────────────────────────┘
```

### 关键约束

1. **注册顺序无关**：Register 可按任意顺序调用
2. **注入顺序固定**：InjectAll 必须按层次从下到上执行
3. **单次注入**：InjectAll 执行后容器标记为已注入，重复调用跳过
4. **不可变性**：注入完成后容器状态固定，不再变更

## 使用示例

### 完整初始化流程

```go
func main() {
    // 1. 创建容器（按依赖顺序）
    configContainer := NewConfigContainer()
    managerContainer := NewManagerContainer(configContainer)
    entityContainer := NewEntityContainer()
    repositoryContainer := NewRepositoryContainer(configContainer, managerContainer, entityContainer)
    serviceContainer := NewServiceContainer(configContainer, managerContainer, repositoryContainer)
    controllerContainer := NewControllerContainer(configContainer, managerContainer, serviceContainer)
    middlewareContainer := NewMiddlewareContainer(configContainer, managerContainer, serviceContainer)

    // 2. 注册实例（可按任意顺序）
    // Config 层
    configContainer.Register(&AppConfig{})

    // Manager 层
    managerContainer.Register(&DatabaseManager{})
    managerContainer.Register(&CacheManager{})

    // Entity 层
    entityContainer.Register(&User{})
    entityContainer.Register(&Order{})

    // Repository 层
    repositoryContainer.Register(&UserRepositoryImpl{})
    repositoryContainer.Register(&OrderRepositoryImpl{})

    // Service 层（注意：OrderService 依赖 PaymentService，但注册顺序不限）
    serviceContainer.Register(&UserServiceImpl{})      // 依赖 OrderService
    serviceContainer.Register(&OrderServiceImpl{})     // 依赖 PaymentService, InventoryService
    serviceContainer.Register(&PaymentServiceImpl{})   // 无同层依赖
    serviceContainer.Register(&InventoryServiceImpl{}) // 无同层依赖

    // Controller 层
    controllerContainer.Register(&UserControllerImpl{})
    controllerContainer.Register(&OrderControllerImpl{})

    // Middleware 层
    middlewareContainer.Register(&AuthMiddleware{})
    middlewareContainer.Register(&LoggingMiddleware{})

    // 3. 执行依赖注入（按层次从下到上）
    // Config 和 Entity 无需注入
    if err := managerContainer.InjectAll(); err != nil {
        log.Fatalf("Manager injection failed: %v", err)
    }

    if err := repositoryContainer.InjectAll(); err != nil {
        log.Fatalf("Repository injection failed: %v", err)
    }

    if err := serviceContainer.InjectAll(); err != nil {
        log.Fatalf("Service injection failed: %v", err)
    }

    if err := controllerContainer.InjectAll(); err != nil {
        log.Fatalf("Controller injection failed: %v", err)
    }

    if err := middlewareContainer.InjectAll(); err != nil {
        log.Fatalf("Middleware injection failed: %v", err)
    }

    // 4. 获取实例使用
    userCtrl, _ := controllerContainer.GetByName("user")
    userCtrl.Handle()
}
```

### 同层依赖示例

```go
// 场景：UserService 依赖 OrderService，OrderService 依赖 PaymentService

// 1. 注册（顺序不限）
serviceContainer.Register(&UserServiceImpl{})      // 依赖 OrderService
serviceContainer.Register(&OrderServiceImpl{})     // 依赖 PaymentService
serviceContainer.Register(&PaymentServiceImpl{})   // 无依赖

// 2. InjectAll 自动处理依赖顺序
// 内部执行流程：
// - 构建依赖图：UserService → OrderService → PaymentService
// - 拓扑排序：[PaymentService, OrderService, UserService]
// - 按顺序注入：先 PaymentService，再 OrderService，最后 UserService
serviceContainer.InjectAll()
```

### 错误处理示例

```go
func initializeContainers() error {
    // ... 创建和注册容器 ...

    // 注入 Manager
    if err := managerContainer.InjectAll(); err != nil {
        return fmt.Errorf("manager injection: %w", err)
    }

    // 注入 Repository
    if err := repositoryContainer.InjectAll(); err != nil {
        return fmt.Errorf("repository injection: %w", err)
    }

    // 注入 Service（可能发生循环依赖或依赖缺失）
    if err := serviceContainer.InjectAll(); err != nil {
        // 详细错误处理
        var depErr *DependencyNotFoundError
        if errors.As(err, &depErr) {
            return fmt.Errorf("service %s missing dependency %s: %w",
                depErr.ServiceName, depErr.FieldName, err)
        }

        var circErr *CircularDependencyError
        if errors.As(err, &circErr) {
            return fmt.Errorf("circular dependency in services: %s → ...",
                strings.Join(circErr.Cycle, " → "))
        }

        return fmt.Errorf("service injection: %w", err)
    }

    return nil
}
```

## 设计约束与最佳实践

### 架构约束

1. **禁止跨层访问**
   - Service 只能通过 ServiceContainer 访问其他 Service，不能直接访问 Controller
   - 下层不能依赖上层，Repository 不能依赖 Service

2. **单向依赖**
   - 依赖方向必须单向：上层 → 下层
   - 同层依赖允许（Service ↔ Service，Manager ↔ Manager），但必须无环

3. **接口隔离**
   - 通过容器隔离，强制使用接口而非具体实现
   - 依赖字段应该声明为接口类型，而非具体实现

4. **全局唯一**
   - 同一容器内实例名称必须唯一（通过 XXName() 方法获取）
   - 类型匹配必须唯一，避免歧义

### 最佳实践

#### 1. 依赖声明

```go
// ✅ 推荐：声明为接口类型
type UserServiceImpl struct {
    OrderSvc   OrderService    `inject:""`  // 接口类型
    UserRepo   UserRepository  `inject:""`  // 接口类型
}

// ❌ 避免：声明为具体实现类型
type UserServiceImpl struct {
    OrderSvc   *OrderServiceImpl `inject:""`  // 具体实现
}
```

#### 2. 同层依赖设计

```go
// ✅ 推荐：无环的依赖关系
UserService → OrderService → PaymentService (无同层依赖)
UserService → NotificationService (无同层依赖)

// ❌ 避免：循环依赖
UserService → OrderService → UserService (循环!)
```

#### 3. 可选依赖处理

```go
// 对于可选依赖，使用指针类型
type UserServiceImpl struct {
    Config     BaseConfigProvider `inject:""`
    CacheMgr   CacheManager       `inject:"optional"` // 可选依赖
}

// InjectAll 时找不到 CacheManager 不会报错，字段保持 nil
```

#### 4. 避免过度依赖

```go
// ✅ 推荐：依赖聚焦
type UserServiceImpl struct {
    UserRepo   UserRepository  `inject:""`
}

// ❌ 避免：依赖过多
type UserServiceImpl struct {
    Config     BaseConfigProvider `inject:""`
    DBManager  DatabaseManager    `inject:""`
    CacheMgr   CacheManager       `inject:""`
    UserRepo   UserRepository     `inject:""`
    OrderRepo  OrderRepository    `inject:""`
    OrderSvc   OrderService       `inject:""`
    PaymentSvc PaymentService     `inject:""`
    // ... 10+ 依赖
}
```

## 实现注意事项

### 性能考虑

1. **反射开销**：InjectAll 使用反射解析字段，仅在启动时执行一次，可接受
2. **拓扑排序复杂度**：O(V + E)，V 为服务数量，E 为依赖边数
3. **并发读取**：注入完成后使用 RWMutex 保护，读取性能高

### 可测试性

```go
// 测试时可以手动注册 Mock 实现
func TestUserService(t *testing.T) {
    mockContainer := NewServiceContainer(nil, nil, nil)

    // 注册 Mock 实现
    mockContainer.Register(&MockOrderService{})
    mockContainer.Register(&UserServiceImpl{})

    // 注入
    mockContainer.InjectAll()

    // 测试
    userSvc, _ := mockContainer.GetByName("user")
    // ...
}
```

### 扩展性

如需支持更复杂的依赖注入场景，可考虑：

1. **限定符注入**：通过标签区分同一接口的多个实现

   ```go
   type UserServiceImpl struct {
       PrimaryDB   DatabaseManager `inject:"primary"`
       ReplicaDB   DatabaseManager `inject:"replica"`
   }
   ```

2. **生命周期管理**：添加 Shutdown 方法支持资源清理

   ```go
   func (c *ServiceContainer) Shutdown() error {
       // 按注入逆序清理资源
   }
   ```

3. **延迟初始化**：支持懒加载，首次使用时才注入
   ```go
   func (c *ServiceContainer) GetLazy(name string) (BaseService, error)
   ```
