# 功能包文档撰写 SOP

## 目的

规范项目各功能包的文档撰写，确保文档风格统一、简明扼要、易于理解。

## 适用范围

适用于 `litecore-go` 项目中所有功能包（package）的文档撰写，包括：
- `doc.go` 文件（包级 Go 文档注释）
- `README.md` 文件（功能包说明文档）

## 项目架构概览

### 包结构层次

```
litecore-go/
├── common/              # 公共基础接口定义（Entity、Manager、Repository、Service、Controller、Middleware）
├── container/            # 依赖注入容器（支持 5 层分层架构）
├── manager/              # 管理器组件（基础能力层）
│   ├── cachemgr/        # 缓存管理器
│   ├── configmgr/       # 配置管理器
│   ├── databasemgr/     # 数据库管理器
│   ├── limitermgr/      # 限流管理器
│   ├── lockmgr/         # 锁管理器
│   ├── loggermgr/       # 日志管理器
│   ├── mqmgr/           # 消息队列管理器
│   └── telemetrymgr/    # 可观测性管理器
├── component/            # 业务组件层
│   ├── litecontroller/  # 内置控制器组件
│   ├── litemiddleware/  # 内置中间件组件
│   └── liteservice/     # 内置服务组件
├── logger/               # 日志工具包
├── server/               # 服务器引擎（Engine、生命周期管理）
├── util/                 # 工具函数包（jwt、hash、crypt、id、rand、string、time、json、request、validator）
└── cli/                  # 命令行工具
```

### 5 层依赖注入架构

```
┌─────────────────────────────────────────────────────────┐
│                   Controller Layer                       │
│              (HTTP 请求处理和响应)                        │
├─────────────────────────────────────────────────────────┤
│                  Middleware Layer                        │
│              (请求预处理和后处理)                         │
├─────────────────────────────────────────────────────────┤
│                   Service Layer                          │
│              (业务逻辑和数据处理)                         │
├─────────────────────────────────────────────────────────┤
│                 Repository Layer                         │
│              (数据访问和持久化)                           │
├─────────────────────────────────────────────────────────┤
│                   Entity Layer                           │
│              (数据模型和领域对象)                         │
└─────────────────────────────────────────────────────────┘
           ↑                                              ↑
           └───────────────── Manager Layer ───────────────┘
            (configmgr、loggermgr、databasemgr、cachemgr、
             lockmgr、limitermgr、mqmgr、telemetrymgr)
```

---

## 一、包命名规范

### Manager 组件命名

所有 Manager 组件位于 `manager/` 目录下，包名统一采用 `<功能名>mgr` 格式：

| 包名 | 功能 | 说明 |
|------|------|------|
| `cachemgr` | 缓存管理 | 支持 Ristretto（内存）和 Redis（分布式） |
| `configmgr` | 配置管理 | 支持 YAML 和 JSON 格式 |
| `databasemgr` | 数据库管理 | 基于 GORM，支持 MySQL、PostgreSQL、SQLite |
| `limitermgr` | 限流管理 | 基于时间窗口的请求频率控制 |
| `lockmgr` | 锁管理 | 分布式锁，支持内存和 Redis |
| `loggermgr` | 日志管理 | 基于 Zap，支持 Gin 风格、JSON 格式 |
| `mqmgr` | 消息队列 | 支持 RabbitMQ 和内存队列 |
| `telemetrymgr` | 可观测性 | OpenTelemetry 集成（Traces、Metrics、Logs） |

### Component 组件命名

所有业务组件位于 `component/` 目录下，统一添加 `lite` 前缀：

| 包名 | 功能 | 说明 |
|------|------|------|
| `litecontroller` | 控制器组件 | 内置控制器（Health、Metrics、Pprof、Resource 等） |
| `litemiddleware` | 中间件组件 | 内置中间件（CORS、Recovery、RequestLogger、RateLimiter 等） |
| `liteservice` | 服务组件 | 内置服务（HTMLTemplate 等） |

### 其他包命名

| 目录 | 包名 | 说明 |
|------|------|------|
| `common/` | `common` | 公共基础接口定义 |
| `container/` | `container` | 依赖注入容器 |
| `logger/` | `logger` | 日志工具包 |
| `server/` | `server` | 服务器引擎 |
| `util/` | `crypt`、`hash`、`jwt` 等 | 工具函数包 |
| `cli/` | `cli` | 命令行工具 |

### 包路径规范

所有包在引用时使用完整的模块路径：

```go
import (
    // 公共接口
    "github.com/lite-lake/litecore-go/common"

    // 管理器组件
    "github.com/lite-lake/litecore-go/manager/cachemgr"
    "github.com/lite-lake/litecore-go/manager/configmgr"
    "github.com/lite-lake/litecore-go/manager/databasemgr"
    "github.com/lite-lake/litecore-go/manager/limitermgr"
    "github.com/lite-lake/litecore-go/manager/lockmgr"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
    "github.com/lite-lake/litecore-go/manager/mqmgr"
    "github.com/lite-lake/litecore-go/manager/telemetrymgr"

    // 业务组件
    "github.com/lite-lake/litecore-go/component/litecontroller"
    "github.com/lite-lake/litecore-go/component/litemiddleware"
    "github.com/lite-lake/litecore-go/component/liteservice"

    // 核心模块
    "github.com/lite-lake/litecore-go/container"
    "github.com/lite-lake/litecore-go/server"
)
```

---

## 二、doc.go 撰写规范

### 文件位置

在每个功能包目录下创建 `doc.go` 文件，如 `manager/cachemgr/doc.go`、`manager/configmgr/doc.go`。

### 基本结构

```go
// Package <包名> <一句话功能描述>。
//
// 核心特性：
//   - <特性1>：<简要说明>
//   - <特性2>：<简要说明>
//   - <特性3>：<简要说明>
//
// 基本用法：
//
//	<代码示例>
//
// <可选章节1>：
//
//	<相关说明或示例>
//
// <可选章节2>：
//
//	<相关说明或示例>
package <包名>
```

### 撰写要点

1. **包声明行**
   - 格式：`// Package <包名> <一句话功能描述>。`
   - 描述要简洁明确，说明包的核心功能
   - 示例：`// Package config 提供配置管理功能，支持 JSON 和 YAML 格式。`

2. **核心特性**
   - 使用列表形式列出 3-6 个核心特性
   - 每个特性用简短的描述说明
   - 突出包的优势和特点

3. **基本用法**
   - 提供简单直接的使用示例
   - 从创建/初始化开始，到基本使用
   - 代码使用 Tab 缩进（Go 注释格式）
   - 避免过于复杂的示例

4. **可选章节**
   - 根据包的特性添加必要的说明章节
   - 常见章节：路径语法、配置选项、错误处理、性能考虑等
   - 每个章节都应配有实际示例

5. **代码示例风格**
   - 使用 Tab 缩进（`//` 后跟 Tab）
   - 注释清晰，关键步骤加说明
   - 错误处理要完整（`if err != nil`）

### 示例参考

参考 `config/doc.go` 的完整示例。

---

## 三、README.md 撰写规范

### 文件位置

在每个功能包目录下创建 `README.md` 文件，如 `manager/cachemgr/README.md`、`component/litemiddleware/README.md`。

### 基本结构

```markdown
# <模块名>

<一句话功能描述>。

## 特性

- **<特性1>** - <简要说明>
- **<特性2>** - <简要说明>
- **<特性3>** - <简要说明>

## 快速开始

<完整的代码示例，展示主要用法>

## <核心功能1>

<详细说明和示例>

## <核心功能2>

<详细说明和示例>

## API

### <接口/函数分类>

<接口或函数签名及说明>

## <其他章节>

<如：错误处理、性能、最佳实践等>
```

### 撰写要点

1. **标题和描述**
   - 标题使用模块名（可加中文副标题）
   - 第一段用一句话说明模块功能

2. **特性章节**
   - 列出 3-6 个核心特性
   - 使用粗体突出关键词
   - 每个特性用简短的描述说明

3. **快速开始**
   - 提供完整可运行的代码示例
   - 涵盖最常见的使用场景
   - 包含必要的导入和错误处理

4. **功能章节**
   - 对每个核心功能创建独立章节
   - 提供详细的说明和代码示例
   - 展示不同使用场景

5. **API 章节**
   - 按功能分组列出主要 API
   - 包含函数签名和简要说明
   - 对于复杂类型，说明其方法和用法

6. **其他章节**
   - 根据需要添加：错误处理、性能考虑、线程安全、最佳实践等
   - 提供相关的代码示例和说明

### 示例参考

参考 `config/README.md` 的完整示例。

---

## 四、doc.go 与 README.md 的配合

### 内容分工

| 方面 | doc.go | README.md |
|------|--------|-----------|
| **受众** | IDE 用户、godoc 阅读者 | 代码阅读者、使用者 |
| **详细程度** | 简洁，快速上手 | 详细，覆盖全面 |
| **代码示例** | 基本用法 | 多场景示例 |
| **API 文档** | 不列出完整 API | 列出主要 API |

### 内容一致性

- 核心特性描述保持一致
- 代码示例使用相同的风格
- 术语和命名保持统一

---

## 五、文档撰写检查清单

### doc.go 检查清单

- [ ] 包声明行清晰描述功能
- [ ] 列出 3-6 个核心特性
- [ ] 提供基本用法示例
- [ ] 代码使用 Tab 缩进
- [ ] 示例代码可运行（概念上）
- [ ] 包含必要的错误处理
- [ ] 特性章节说明准确

### README.md 检查清单

- [ ] 标题简洁明确
- [ ] 一句话功能描述
- [ ] 特性章节突出重点
- [ ] 快速开始示例完整
- [ ] 核心功能有独立章节
- [ ] API 按功能分组
- [ ] 包含错误处理说明
- [ ] 必要时包含性能/线程安全说明

### 通用检查清单

- [ ] 无错别字
- [ ] 代码示例格式正确
- [ ] 术语使用一致
- [ ] 中文表达简明流畅
- [ ] 无冗余内容
- [ ] 示例代码符合 Go 最佳实践

---

## 六、最佳实践

1. **保持简洁**
    - 文档不是越长越好，突出重点
    - 避免过度解释显而易见的内容

2. **示例优先**
    - 代码示例比文字描述更直观
    - 每个重要功能都应有示例

3. **面向读者**
    - 从使用者的角度组织内容
    - 先展示常用功能，再展示高级特性

4. **及时更新**
    - 代码变更时同步更新文档
    - 删除过时内容，避免误导

5. **保持一致**
    - 与项目其他包的文档风格保持一致
    - 使用统一的术语和命名

6. **包路径准确**
    - 使用正确的包路径（如 `manager/cachemgr` 而非 `server/builtin/manager/cachemgr`）
    - 组件包名使用 `lite` 前缀（`litecontroller`、`litemiddleware`、`liteservice`）

7. **依赖关系清晰**
    - 明确说明组件之间的依赖关系
    - 对于 Manager 组件，说明依赖的其他 Manager

---

## 七、常见问题

**Q: doc.go 和 README.md 哪个更重要？**

A: 两者同样重要，服务于不同场景。doc.go 是 Go 生态的标准文档方式，会被 godoc 工具提取；README.md 是项目文档的标准方式，更适合在 GitHub 等平台阅读。

**Q: 是否必须在每个包都创建这两个文件？**

A: 建议所有对外暴露的功能包都创建。内部实现包可以根据需要简化。

**Q: 文档应该用中文还是英文？**

A: 根据项目定位确定。litecore-go 目前使用中文文档，但代码和 API 命名应使用英文。

**Q: 如何处理复杂的示例代码？**

A: 在 README.md 中可以包含较长的示例，doc.go 中保持简洁。复杂示例可以放在单独的 `_test.go` 文件或 `examples/` 目录中。

**Q: Manager 组件的包路径是什么？**

A: Manager 组件位于 `manager/` 目录下，完整路径为 `github.com/lite-lake/litecore-go/manager/<功能名>mgr`。例如：`github.com/lite-lake/litecore-go/manager/cachemgr`。

**Q: Component 组件的包命名有什么规范？**

A: Component 组件统一使用 `lite` 前缀：`litecontroller`、`litemiddleware`、`liteservice`。完整路径为 `github.com/lite-lake/litecore-go/component/<包名>`。

**Q: 如何在文档中引用 Manager 组件？**

A: 在代码示例中，使用完整的包路径引用：
```go
import (
    "github.com/lite-lake/litecore-go/manager/cachemgr"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
)

type UserService struct {
    CacheMgr  cachemgr.ICacheManager   `inject:""`
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}
```

**Q: 5 层依赖注入架构包括哪些层？**

A: 包括 Entity、Repository、Service、Controller、Middleware 五层。Manager 层作为基础设施，由 Engine 自动初始化并注入到其他层。

---

## 八、文档示例模板

### Manager 组件 doc.go 模板

```go
// Package <功能名>mgr 提供<功能描述>，支持<支持的驱动类型>驱动。
//
// 核心特性：
//   - 多驱动支持：支持 <驱动1>（<描述1>）、<驱动2>（<描述2>）、<驱动3>（<描述3>）三种驱动
//   - 统一接口：提供统一的 I<功能名>Manager 接口，便于切换实现
//   - 可观测性：内置日志、指标和链路追踪支持
//   - <其他特性1>：<简要说明>
//   - <其他特性2>：<简要说明>
//
// 基本用法：
//
//	// 使用 <默认驱动>
//	mgr := <功能名>mgr.New<功能名>Manager<实现名>Impl(<参数>)
//	defer mgr.Close()
//
//	ctx := context.Background()
//
//	// 基本操作
//	err := mgr.<操作>(ctx, ...)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// 使用 <其他驱动>：
//
//	cfg := &<功能名>mgr.<驱动名>Config{
//	    <配置字段1>: <值1>,
//	    <配置字段2>: <值2>,
//	}
//	mgr, err := <功能名>mgr.New<功能名>Manager<驱动名>Impl(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer mgr.Close()
//
// 配置驱动类型：
//
//	// 通过 Build 函数创建
//	mgr, err := <功能名>mgr.Build("<驱动类型>", map[string]any{
//	    "<配置键>": "<配置值>",
//	})
//
//	// 通过配置提供者创建
//	mgr, err := <功能名>mgr.BuildWithConfigProvider(configProvider)
//
// 高级用法：
//
//	// <高级功能说明>
//	mgr.<高级操作>(ctx, ...)
package <功能名>mgr
```

### Component 组件 doc.go 模板

```go
// Package <litecontroller|litemiddleware|liteservice> 提供<功能描述>组件。
//
// 核心特性：
//   - <特性1>：<简要说明>
//   - <特性2>：<简要说明>
//   - <特性3>：<简要说明>
//   - 依赖注入：支持通过 `inject:""` 标签注入 Manager 和其他组件
//   - 生命周期管理：实现 OnStart/OnStop 钩子方法
//
// 基本用法：
//
//	// 创建组件实例
//	component := <litecontroller|litemiddleware|liteservice>.New<组件名>()
//
//	// 自定义配置（可选）
//	config := &<litecontroller|litemiddleware|liteservice>.<组件名>Config{
//	    <字段>: <值>,
//	}
//	component := <litecontroller|litemiddleware|liteservice>.New<组件名>(config)
//
//	// 注册到容器
//	container.<RegisterController|RegisterMiddleware|RegisterService>(component)
//
// 中间件特有用法：
//
//	// 中间件支持自定义 Name 和 Order
//	name := "CustomMiddleware"
//	order := 300
//	middleware := <litemiddleware>.New<中间件名>(&<litemiddleware>.<中间件名>Config{
//	    Name:  &name,
//	    Order: &order,
//	})
//
// 控制器特有用法：
//
//	// 控制器需要实现 Handle 方法处理请求
//	func (c *<控制器名>) Handle(ctx *gin.Context) {
//	    // 处理逻辑
//	}
//
// 服务特有用法：
//
//	// 服务可以依赖其他服务和 Manager
//	type <服务名> struct {
//	    LoggerMgr loggermgr.ILoggerManager `inject:""`
//	    DBManager databasemgr.IDatabaseManager `inject:""`
//	}
package <litecontroller|litemiddleware|liteservice>
```

### 工具包 doc.go 模板

```go
// Package <包名> 提供<功能描述>功能。
//
// 核心特性：
//   - <特性1>：<简要说明>
//   - <特性2>：<简要说明>
//   - <特性3>：<简要说明>
//
// 基本用法：
//
//	// <功能1>
//	result := <函数名>(<参数>)
//
//	// <功能2>
//	result, err := <函数名>(<参数>)
//	if err != nil {
//	    return err
//	}
//
// 高级用法：
//
//	// <高级功能说明>
//	<代码示例>
package <包名>
```

---

## 九、版本历史

### v2.0.0 (2026-01-24)

- **架构重构**：
  - Manager 组件从 `server/builtin/manager` 迁移至独立的 `manager/` 目录
  - Component 组件重构为 `litecontroller`、`litemiddleware`、`liteservice` 三个子包
  - 5 层依赖注入架构规范化（Entity → Repository → Service → Controller/Middleware）

- **包命名规范更新**：
  - Manager 组件统一使用 `<功能名>mgr` 命名（如 `cachemgr`、`configmgr`）
  - Component 组件统一添加 `lite` 前缀（如 `litecontroller`、`litemiddleware`）
  - 新增 `limitermgr`、`lockmgr`、`mqmgr` 管理器

- **功能增强**：
  - 缓存从 go-cache 替换为 Ristretto（更高性能）
  - 日志格式升级，支持 Gin 风格格式
  - 中间件支持通过配置自定义 Name 和 Order
  - 新增 RateLimiter 中间件
  - 新增启动日志功能

- **依赖注入优化**：
  - Manager 组件由 Engine 自动初始化和注入
  - 统一日志注入机制
  - 支持通过 DI 注入 Manager

- **文档规范**：
  - 更新本文档以反映最新的架构和包结构
  - 所有 Manager 和 Component 组件更新 doc.go 和 README.md

### v1.0.0 (2025-xx-xx)

- 初始版本
- 基础功能包文档规范

---

## 十、相关文档

- [AGENTS.md](../AGENTS.md) - 项目开发指南
- [README.md](../README.md) - 项目主文档
- [manager/README.md](../manager/README.md) - Manager 组件说明
- [component/README.md](../component/README.md) - Component 组件说明
- [container/README.md](../container/README.md) - 依赖注入容器说明
- [common/README.md](../common/README.md) - 公共接口说明
