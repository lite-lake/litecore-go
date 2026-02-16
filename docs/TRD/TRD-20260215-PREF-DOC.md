# 文档重构任务：面向框架使用者的用户指南

## 任务标识

- **任务编号**: PREF-DOC
- **创建日期**: 2026-02-15
- **任务类型**: 文档重构
- **优先级**: 中

---

## 1. 背景与问题

### 1.1 现状

LiteCore-Go 框架当前已有较完善的文档，但存在以下问题：

1. **文档分散**：用户文档分布在多处
   - `README.md` - 项目入口文档
   - `docs/GUIDE-lite-core-framework-usage.md` - 详细使用指南（约1500行）
   - `docs/SOP-build-business-application.md` - 业务开发规范
   - `docs/SOP-middleware.md` - 中间件规范
   - 各子目录的 README.md（manager/、container/、common/、server/、component/、cli/）

2. **受众混杂**：
   - `README.md` 和 `docs/GUIDE-*.md` 同时面向使用者和开发者
   - `AGENTS.md` 面向 AI 编码助手
   - 缺乏专门面向"框架使用者"的独立文档体系

3. **查找困难**：
   - 用户需要阅读多个文档才能找到所需信息
   - 组件信息分散在各子目录 README 中
   - 缺乏统一的索引和导航

### 1.2 目标用户

本次文档重构的目标用户是**框架使用者**，即：
- 使用 LiteCore-Go 框架开发业务系统的开发者
- 需要快速了解框架功能、学会使用框架的工程师
- 查找特定组件用法的中级用户

**不包括**：
- 框架贡献者（面向 AI 助手和贡献者的文档保留在 AGENTS.md）
- 框架设计者（TRD 文档保留在 docs/TRD/）

---

## 2. 目标与边界

### 2.1 目标

创建独立、完整的用户指南文档体系，放在 `docs/user-guides/` 目录下：

1. **快速入门**：5 分钟内完成安装和第一个接口
2. **架构理解**：清晰理解 5 层架构和依赖规则
3. **开发指南**：按步骤完成业务开发
4. **组件索引**：快速查找内置组件的用法
5. **最佳实践**：避免常见问题

### 2.2 边界

**包含**：
- 创建新的 `docs/user-guides/` 目录及所有文档
- 新文档内容整合自现有文档，但重新组织和精简
- 每个文档控制在 200-400 行

**不包含**：
- 不删除现有文档（README.md、AGENTS.md、docs/GUIDE-*.md 等）
- 不修改现有代码
- 不创建英文版本文档

### 2.3 与现有文档的关系

| 现有文档 | 处理方式 | 新文档关系 |
|---------|---------|-----------|
| `README.md` | 保留，作为项目入口 | 添加指向 user-guides 的链接 |
| `AGENTS.md` | 保留 | 面向 AI 和贡献者，用户指南不覆盖 |
| `docs/GUIDE-*.md` | 保留 | 内容整合到 user-guides，作为详细参考 |
| `docs/SOP-*.md` | 保留 | 内容整合到 user-guides |
| `docs/TRD/` | 保留 | 技术设计文档，用户指南不涉及 |
| 子目录 README | 保留 | 作为详细参考，user-guides 做索引摘要 |

---

## 3. 目标文档结构

```
docs/user-guides/
├── README.md                      # 文档索引（入口导航）
├── 01-quick-start.md              # 快速开始（5分钟入门）
├── 02-architecture-overview.md    # 架构概览（5层架构、依赖规则）
├── 03-development-sop.md          # 业务开发 SOP（分步骤指南）
├── 04-components/                 # 组件索引目录
│   ├── README.md                  # 组件总览
│   ├── managers.md                # Manager 组件索引（9个内置管理器）
│   ├── middlewares.md             # 中间件组件索引（6个内置中间件）
│   ├── controllers.md             # 控制器组件索引（内置控制器）
│   ├── services.md                # 服务组件索引（内置服务）
│   ├── utils.md                   # 工具包索引（util/*）
│   └── cli.md                     # CLI 工具使用
└── 05-best-practices.md           # 最佳实践与常见问题
```

### 3.1 各文档内容规划

#### `README.md` - 文档索引
- 文档体系说明
- 快速导航（按用户场景）
- 版本信息

#### `01-quick-start.md` - 快速开始
目标：5 分钟内完成从安装到运行

内容：
1. 环境要求（Go 1.25+）
2. 安装 CLI 工具
3. 使用脚手架创建项目
4. 运行应用
5. 添加第一个接口（完整示例）

来源：`docs/GUIDE-lite-core-framework-usage.md` 第 2 章

#### `02-architecture-overview.md` - 架构概览
目标：理解框架设计思想

内容：
1. 5 层架构图（ASCII 图）
2. 依赖规则（谁依赖谁）
3. 各层职责边界
4. 生命周期管理（OnStart/OnStop）
5. 依赖注入机制

来源：`docs/GUIDE-lite-core-framework-usage.md` 第 4 章

#### `03-development-sop.md` - 业务开发 SOP
目标：按步骤完成业务开发

内容：
1. 创建实体（Entity）- 使用基类
2. 创建仓储（Repository）
3. 创建服务（Service）
4. 创建控制器（Controller）
5. 创建中间件（Middleware）
6. 创建监听器（Listener）
7. 创建定时器（Scheduler）
8. 运行代码生成器
9. 启动应用

每个步骤包含：
- 代码模板
- 命名规范
- 注意事项

来源：`docs/SOP-build-business-application.md` + `docs/GUIDE-lite-core-framework-usage.md` 第 5 章

#### `04-components/managers.md` - Manager 组件索引
目标：快速查找 Manager 用法

内容表格：
| Manager | 包路径 | 接口 | 功能 | 支持驱动 | 配置示例 |
|---------|--------|------|------|----------|---------|

每个 Manager 包含：
- 核心接口定义
- 配置说明
- 使用示例（代码片段）
- 初始化顺序说明

来源：`manager/README.md`

覆盖的 9 个 Manager：
1. ConfigManager (`manager/configmgr`)
2. TelemetryManager (`manager/telemetrymgr`)
3. LoggerManager (`manager/loggermgr`)
4. DatabaseManager (`manager/databasemgr`)
5. CacheManager (`manager/cachemgr`)
6. LockManager (`manager/lockmgr`)
7. LimiterManager (`manager/limitermgr`)
8. MQManager (`manager/mqmgr`)
9. SchedulerManager (`manager/schedulermgr`)

#### `04-components/middlewares.md` - 中间件组件索引
目标：快速查找内置中间件用法

内容表格：
| 中间件 | 默认 Order | 功能 | 配置项 | 依赖 |

每个中间件包含：
- 功能说明
- 配置结构
- 使用示例（默认配置 + 自定义配置）
- Order 说明

来源：`component/litemiddleware/README.md` + `component/README.md`

覆盖的 6 个中间件：
1. RecoveryMiddleware (Order: 0) - Panic 恢复
2. RequestLoggerMiddleware (Order: 50) - 请求日志
3. CORSMiddleware (Order: 100) - 跨域处理
4. SecurityHeadersMiddleware (Order: 150) - 安全头
5. RateLimiterMiddleware (Order: 200) - 限流
6. TelemetryMiddleware (Order: 250) - 遥测

#### `04-components/controllers.md` - 控制器组件索引
目标：快速查找内置控制器用法

内容表格：
| 控制器 | 路由 | 功能 | 依赖 |

每个控制器包含：
- 功能说明
- 路由定义
- 使用示例

来源：`component/litecontroller/README.md`

覆盖的控制器：
1. HealthController - `/health [GET]` - 健康检查
2. MetricsController - `/metrics [GET]` - 运行指标
3. PprofIndexController - `/debug/pprof [GET]`
4. PprofHeapController - `/debug/pprof/heap [GET]`
5. PprofGoroutineController - `/debug/pprof/goroutine [GET]`
6. PprofAllocsController - `/debug/pprof/allocs [GET]`
7. PprofBlockController - `/debug/pprof/block [GET]`
8. PprofMutexController - `/debug/pprof/mutex [GET]`
9. PprofProfileController - `/debug/pprof/profile [GET]`
10. PprofTraceController - `/debug/pprof/trace [GET]`
11. ResourceStaticController - 自定义 - 静态文件服务
12. ResourceHTMLController - 自定义 - HTML 模板渲染

#### `04-components/services.md` - 服务组件索引
目标：快速查找内置服务用法

内容表格：
| 服务 | 功能 | 依赖 |

每个服务包含：
- 功能说明
- 使用示例

来源：`component/liteservice/README.md`

覆盖的服务：
1. HTMLTemplateService - HTML 模板渲染
2. BloomFilterService - 布隆过滤器
3. I18nService - 国际化

#### `04-components/utils.md` - 工具包索引
目标：快速查找工具函数

内容表格：
| 工具包 | 功能 | 常用函数 |

每个工具包包含：
- 功能说明
- 常用函数示例

来源：各 `util/*/README.md` 或代码

覆盖的工具包：
- `util/jwt` - JWT 令牌
- `util/hash` - 哈希算法
- `util/crypt` - 加密解密
- `util/id` - ID 生成（CUID2）
- `util/validator` - 数据验证
- `util/string` - 字符串处理
- `util/time` - 时间处理
- `util/json` - JSON 处理
- `util/rand` - 随机数

#### `04-components/cli.md` - CLI 工具使用
目标：掌握 CLI 工具命令

内容：
1. 安装方式
2. 命令概览
   - `litecore-cli generate` - 代码生成
   - `litecore-cli scaffold` - 项目脚手架
   - `litecore-cli version` - 版本信息
3. 参数说明
4. 模板类型（basic/standard/full）

来源：`cli/README.md`

#### `05-best-practices.md` - 最佳实践与常见问题
目标：避免常见问题

内容：
1. 命名规范
2. 错误处理
3. 日志使用规范
4. 事务管理
5. 性能优化建议
6. 常见问题 FAQ（Q&A 形式）

来源：`AGENTS.md` + `docs/GUIDE-lite-core-framework-usage.md` 第 11-12 章

---

## 4. 设计原则

### 4.1 文档风格

1. **简洁优先**：每个文档控制在 200-400 行
2. **代码示例**：每个功能点配有可直接运行的代码
3. **表格索引**：组件特性用表格快速呈现
4. **渐进式**：从快速开始 → 架构理解 → 深入组件 → 最佳实践
5. **中文**：所有文档使用中文

### 4.2 代码示例规范

```go
// 包名和导入
package example

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/databasemgr"
)

// 示例代码要完整可运行
// 注释使用中文
```

### 4.3 格式规范

- 使用 GitHub 风格 Markdown
- 代码块指定语言
- 表格对齐
- 标题层级不超过 4 级

---

## 5. 执行步骤

### 5.1 准备工作

1. 创建 `docs/user-guides/` 目录
2. 创建 `docs/user-guides/04-components/` 目录

### 5.2 创建文档（按顺序）

1. `docs/user-guides/README.md` - 文档索引
2. `docs/user-guides/01-quick-start.md` - 快速开始
3. `docs/user-guides/02-architecture-overview.md` - 架构概览
4. `docs/user-guides/03-development-sop.md` - 开发 SOP
5. `docs/user-guides/04-components/README.md` - 组件总览
6. `docs/user-guides/04-components/managers.md` - Manager 索引
7. `docs/user-guides/04-components/middlewares.md` - 中间件索引
8. `docs/user-guides/04-components/controllers.md` - 控制器索引
9. `docs/user-guides/04-components/services.md` - 服务索引
10. `docs/user-guides/04-components/utils.md` - 工具包索引
11. `docs/user-guides/04-components/cli.md` - CLI 工具
12. `docs/user-guides/05-best-practices.md` - 最佳实践

### 5.3 验收标准

1. 所有文档创建完成
2. 每个文档不超过 400 行
3. 代码示例可直接运行
4. 表格信息完整准确
5. 文档间链接正确

---

## 6. 关键参考文档

执行时需要参考的现有文档：

| 文档 | 路径 | 用途 |
|------|------|------|
| 主 README | `README.md` | 项目概述、快速开始参考 |
| AI 指南 | `AGENTS.md` | 命名规范、代码风格参考 |
| 详细使用指南 | `docs/GUIDE-lite-core-framework-usage.md` | 主要内容来源 |
| 业务开发 SOP | `docs/SOP-build-business-application.md` | 开发流程参考 |
| 中间件 SOP | `docs/SOP-middleware.md` | 中间件详细说明 |
| Manager README | `manager/README.md` | Manager 组件详情 |
| Server README | `server/README.md` | 服务引擎详情 |
| Container README | `container/README.md` | 依赖注入详情 |
| Common README | `common/README.md` | 基础接口详情 |
| Component README | `component/README.md` | 内置组件总览 |
| litemiddleware README | `component/litemiddleware/README.md` | 中间件详情 |
| litecontroller README | `component/litecontroller/README.md` | 控制器详情 |
| liteservice README | `component/liteservice/README.md` | 服务详情 |
| CLI README | `cli/README.md` | CLI 工具详情 |

---

## 7. 注意事项

1. **不要修改现有文档**：本任务只创建新文档，不修改或删除现有文档
2. **内容整合而非复制**：从现有文档提取精华，重新组织，不是简单复制
3. **保持一致性**：与 AGENTS.md 中的命名规范、代码风格保持一致
4. **示例项目引用**：适当引用 `samples/messageboard` 作为完整示例
5. **实体基类**：强调使用 `BaseEntityWithTimestamps` 等基类，CUID2 ID 类型为 string

---

## 8. 附录：框架核心概念速查

### 8.1 5 层架构

```
交互层 (Controller/Middleware/Listener/Scheduler)
         ↓ 依赖
    Service 层
         ↓ 依赖
   Repository 层
         ↓ 依赖
     Entity 层
         ↑
    Manager 层（内置，自动注入）
```

### 8.2 9 个内置 Manager

1. ConfigManager - 配置管理
2. TelemetryManager - 遥测管理
3. LoggerManager - 日志管理
4. DatabaseManager - 数据库管理
5. CacheManager - 缓存管理
6. LockManager - 锁管理
7. LimiterManager - 限流管理
8. MQManager - 消息队列管理
9. SchedulerManager - 定时任务管理

### 8.3 6 个内置中间件

| 中间件 | Order | 功能 |
|--------|-------|------|
| Recovery | 0 | Panic 恢复 |
| RequestLogger | 50 | 请求日志 |
| CORS | 100 | 跨域处理 |
| SecurityHeaders | 150 | 安全头 |
| RateLimiter | 200 | 限流 |
| Telemetry | 250 | 遥测 |

### 8.4 依赖注入

使用 `inject:""` 标签声明依赖：

```go
type MessageService struct {
    Config    configmgr.IConfigManager    `inject:""`
    DBManager databasemgr.IDatabaseManager `inject:""`
    LoggerMgr loggermgr.ILoggerManager   `inject:""`
    Repo      IMessageRepository          `inject:""`
}
```

### 8.5 实体基类

```go
type Message struct {
    common.BaseEntityWithTimestamps  // ID + CreatedAt + UpdatedAt
    Nickname string `gorm:"type:varchar(20);not null"`
    Content  string `gorm:"type:varchar(500);not null"`
}
```

- ID 类型：string（CUID2 25位）
- 数据库存储：varchar(32)
- 时间戳：由 GORM Hook 自动填充
