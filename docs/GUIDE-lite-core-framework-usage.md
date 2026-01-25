# LiteCore 使用指南

## 目录

- [1. 简介](#1-简介)
- [2. 快速开始](#2-快速开始)
  - [2.1 安装 CLI 工具](#21-安装-cli-工具)
  - [2.2 使用脚手架创建项目](#22-使用脚手架创建项目)
  - [2.3 项目结构](#23-项目结构)
  - [2.4 运行应用](#24-运行应用)
  - [2.5 手动创建项目（可选）](#25-手动创建项目可选)
- [3. 核心特性](#3-核心特性)
- [4. 架构概述](#4-架构概述)
  - [4.1 5 层架构图](#41-5-层架构图)
  - [4.2 依赖规则](#42-依赖规则)
  - [4.3 生命周期管理](#43-生命周期管理)
- [5. 5 层架构详解](#5-5-层架构详解)
  - [5.1 Entity 层（实体层）](#51-entity-层实体层)
  - [5.2 Repository 层（仓储层）](#52-repository-层仓储层)
  - [5.3 Service 层（服务层）](#53-service-层服务层)
  - [5.4 交互层 - Controller（控制器层）](#54-交互层---controller控制器层)
  - [5.5 交互层 - Middleware（中间件层）](#55-交互层---middleware中间件层)
    - [5.5.1 内置中间件](#551-内置中间件)
    - [5.5.2 限流器中间件详解](#552-限流器中间件详解)
    - [5.5.3 认证中间件示例](#553-认证中间件示例)
    - [5.5.4 中间件执行顺序](#554-中间件执行顺序)
    - [5.5.5 中间件设计规范](#555-中间件设计规范)
  - [5.6 交互层 - Listener（监听器层）](#56-交互层---listener监听器层)
  - [5.7 交互层 - Scheduler（调度器层）](#57-交互层---scheduler调度器层)
- [6. 内置组件](#6-内置组件)
  - [6.1 Config（配置）](#61-config配置)
  - [6.2 Manager（管理器）](#62-manager管理器)
  - [6.3 LockMgr（锁管理器）](#63-lockmgr锁管理器)
  - [6.4 LimiterMgr（限流管理器）](#64-limitermgr限流管理器)
  - [6.5 MQMgr（消息队列管理器）](#65-mqmgr消息队列管理器)
  - [6.6 可用的内置 Manager](#66-可用的内置-manager)
  - [6.7 使用内置组件](#67-使用内置组件)
  - [6.8 日志配置（Gin 格式）](#68-日志配置gin-格式)
  - [6.9 启动日志](#69-启动日志)
- [7. 代码生成器使用](#7-代码生成器使用)
- [8. 依赖注入机制](#8-依赖注入机制)
- [9. 配置管理](#9-配置管理)
- [10. 实用工具（util 包）](#10-实用工具util-包)
- [11. 最佳实践](#11-最佳实践)
- [12. 常见问题](#12-常见问题)
- [附录](#附录)

---

## 1. 简介

LiteCore 是一个基于 Go 的轻量级应用框架，旨在提供标准化、可扩展的微服务开发能力。框架采用 **5 层分层架构**，内置依赖注入容器、配置管理、数据库管理、缓存管理、日志管理、锁管理、限流管理、消息队列等功能，帮助开发者快速构建业务系统。

### 为什么要使用 LiteCore？

- **标准化架构**：统一的 5 层架构规范，降低团队协作成本
- **独立管理器**：Manager 组件作为独立包，易于扩展和维护
- **内置组件**：提供丰富的内置中间件和控制器，开箱即用
- **依赖注入**：自动化的依赖注入容器，简化组件管理
- **高性能缓存**：基于 Ristretto 的内存缓存，性能优异
- **分布式支持**：内置分布式锁、限流和消息队列
- **灵活日志**：支持 Gin 风格、JSON、Default 等多种日志格式
- **代码生成**：自动生成容器代码，减少重复劳动
- **可观测性**：内置日志、指标、链路追踪支持
- **配置驱动**：通过配置文件管理应用行为，无需修改代码

### 适用场景

- Web 应用和 API 服务
- 微服务架构
- 标准业务系统
- 需要快速原型开发的项目

---

## 2. 快速开始

### 2.1 安装 CLI 工具

#### 方法 1：go install（推荐）

```bash
go install github.com/lite-lake/litecore-go/cli@latest
```

安装后直接使用 `litecore-cli` 命令。

#### 方法 2：本地构建

```bash
# 构建可执行文件
go build -o litecore-cli ./cli

# 或直接使用 go run
go run ./cli/main.go
```

### 2.2 使用脚手架创建项目（推荐）

使用 CLI 脚手架可以快速创建符合 LiteCore 架构的项目骨架。

#### 交互式模式（最简单）

```bash
# 交互式创建项目（推荐新手使用）
litecore-cli scaffold
```

系统会提示您输入：
- 模块路径（如：`github.com/yourname/myapp`）
- 项目名称（如：`myapp`）
- 输出目录（默认：当前目录）
- 模板类型（basic/standard/full）
- 是否生成静态文件服务
- 是否生成 HTML 模板服务
- 是否生成健康检查控制器

#### 命令行模式

```bash
# 使用 basic 模板（仅基础结构）
litecore-cli scaffold --module github.com/yourname/myapp --project myapp --template basic

# 使用 standard 模板（基础 + 配置文件 + 中间件）
litecore-cli scaffold --module github.com/yourname/myapp --project myapp --template standard

# 使用 full 模板（standard + 完整示例代码）
litecore-cli scaffold --module github.com/yourname/myapp --project myapp --template full

# 添加扩展选项
litecore-cli scaffold \
  --module github.com/yourname/myapp \
  --project myapp \
  --template full \
  --static \
  --html \
  --health
```

#### 模板说明

| 模板 | 包含内容 | 适用场景 |
|------|---------|---------|
| `basic` | 目录结构 + go.mod + README | 了解框架基础结构 |
| `standard` | basic + 配置文件 + 基础中间件 + 自动生成容器代码 | 快速开始新项目 |
| `full` | standard + 完整示例代码（entity/repository/service/controller/listener/scheduler） | 查看完整实现 |

#### 扩展选项

| 选项 | 说明 |
|------|------|
| `--static` | 生成静态文件服务（CSS/JS） |
| `--html` | 生成 HTML 模板服务 |
| `--health` | 生成健康检查控制器 |

### 2.3 项目结构

创建后的项目结构（以 full 模板为例）：

```
myapp/
├── cmd/
│   ├── generate/               # 代码生成器
│   │   └── main.go
│   └── server/                 # 应用入口
│       └── main.go
├── configs/
│   └── config.yaml             # 配置文件
├── internal/
│   ├── application/            # 生成的容器代码（DO NOT EDIT）
│   │   ├── entity_container.go
│   │   ├── repository_container.go
│   │   ├── service_container.go
│   │   ├── controller_container.go
│   │   ├── middleware_container.go
│   │   ├── listener_container.go
│   │   ├── scheduler_container.go
│   │   └── engine.go
│   ├── entities/               # 实体层
│   │   └── example_entity.go   # 实体示例（full 模板）
│   ├── repositories/           # 仓储层
│   │   └── example_repository.go
│   ├── services/               # 服务层
│   │   └── example_service.go
│   ├── controllers/            # 控制器层
│   │   └── example_controller.go
│   ├── middlewares/            # 中间件层
│   │   └── recovery_middleware.go
│   ├── listeners/              # 监听器层
│   │   └── example_listener.go
│   ├── schedulers/             # 调度器层
│   │   └── example_scheduler.go
│   └── dtos/                   # 数据传输对象
├── static/                     # 静态资源（--static 选项）
│   ├── css/
│   │   └── style.css
│   └── js/
│       └── app.js
├── templates/                  # HTML 模板（--html 选项）
│   └── index.html
├── data/                       # 数据目录
├── logs/                       # 日志目录
├── go.mod
├── go.sum
└── README.md
```

### 2.4 运行应用

```bash
# 进入项目目录
cd myapp

# 生成容器代码（标准/ full 模板已自动生成）
go run ./cmd/generate

# 运行应用
go run ./cmd/server
```

应用启动后，访问：
- 首页：http://localhost:8080/
- 健康检查：http://localhost:8080/api/health

### 2.5 手动创建项目（可选）

如果需要完全手动创建项目，可以按照以下步骤：

```bash
# 创建项目目录
mkdir myapp && cd myapp

# 初始化 Go 模块
go mod init github.com/yourname/myapp

# 引用 LiteCore
go get github.com/lite-lake/litecore-go@latest

# 创建项目结构
mkdir -p cmd/server cmd/generate configs data
mkdir -p internal/{entities,repositories,services,controllers,middlewares,listeners,schedulers,dtos}

# 创建配置文件
touch configs/config.yaml
```

然后手动创建配置文件和代码，最后运行代码生成器。

**推荐使用 CLI 脚手架，可以快速创建项目骨架。**

---

## 3. 核心特性

### 3.1 框架核心功能

| 功能 | 说明 | 实现方式 |
|------|------|----------|
| **5 层架构** | 内置管理器层 → Entity → Repository → Service → 交互层（Controller/Middleware/Listener/Scheduler） | 接口定义 + 依赖注入 |
| **内置组件** | Config 和 Manager 自动初始化和注入 | server 包 + manager 独立包 |
| **依赖注入** | 自动扫描、自动注入、生命周期管理 | reflect + inject 标签 |
| **代码生成** | 自动生成容器代码和引擎代码 | CLI 工具 |
| **配置管理** | 支持 YAML/JSON 配置文件 | manager/configmgr 包 |
| **数据库管理** | 支持 MySQL/PostgreSQL/SQLite | GORM + manager/databasemgr |
| **缓存管理** | 支持 Redis/Memory 缓存（基于 Ristretto） | manager/cachemgr |
| **日志管理** | 基于 Zap 的高性能日志，支持 Gin 格式 | manager/loggermgr |
| **锁管理** | 支持 Redis/Memory 分布式锁 | manager/lockmgr |
| **限流管理** | 支持 Redis/Memory 限流 | manager/limitermgr |
| **消息队列** | 支持 RabbitMQ/Memory 消息队列 | manager/mqmgr |
| **遥测支持** | OpenTelemetry 集成 | manager/telemetrymgr |
| **启动日志** | 支持异步启动日志记录 | server 包 |
| **中间件组件** | 提供内置中间件，支持配置化 | component/litemiddleware |
| **实体基类** | 提供 BaseEntityOnlyID、BaseEntityWithCreatedAt、BaseEntityWithTimestamps 三种基类 | common 包 + GORM Hook |

### 3.2 实用工具（util 包）

LiteCore 提供了一系列实用的工具包，帮助开发者处理常见的开发任务：

| 工具包 | 功能 |
|--------|------|
| `util/jwt` | JWT 令牌生成、解析和验证 |
| `util/hash` | 常见哈希算法（MD5、SHA1、SHA256） |
| `util/crypt` | 密码加密、AES 加密 |
| `util/id` | 唯一 ID 生成（雪花算法、UUID） |
| `util/validator` | 数据验证工具 |
| `util/string` | 字符串处理工具 |
| `util/time` | 时间处理工具 |
| `util/json` | JSON 处理工具 |
| `util/rand` | 随机数生成工具 |

---

## 4. 架构概述

### 4.1 5 层架构图

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP Request                         │
└─────────────────────────┬───────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────┐
│  Middleware 层（中间件）                                │
│  - Recovery - CORS - Auth - Logger - Telemetry        │
│  - RateLimiter - SecurityHeaders                       │
└─────────────────────────┬───────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────┐
│  Controller 层（控制器）                                │
│  - 请求参数验证                                          │
│  - 调用 Service                                          │
│  - 响应封装                                              │
└─────────────────────────┬───────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────┐
│  Service 层（服务）                                      │
│  - 业务逻辑                                              │
│  - 数据验证                                              │
│  - 事务管理                                              │
│  - 缓存、锁、限流、消息队列                              │
└─────────────────────────┬───────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────┐
│  Repository 层（仓储）                                    │
│  - 数据访问                                              │
│  - ORM 操作                                              │
│  - 数据库迁移                                            │
└─────────────────────────┬───────────────────────────────┘
            ↑ 依赖              ↓ 使用
┌─────────────────────────┐    ┌──────────────────────┐
│  Entity    (实体层)     │    │  Manager   (独立包)  │
│  - 数据模型定义          │    │  - ConfigManager     │
│  - 表结构定义            │    │  - DatabaseManager   │
│  - CUID2 ID             │    │  - CacheManager      │
│  - 自动时间戳            │    │  - LoggerManager     │
└─────────────────────────┘    │  - LockManager       │
                                 │  - LimiterManager    │
                                 │  - MQManager         │
                                 │  - TelemetryManager  │
                                 └──────────────────────┘
                                           ↑ 依赖
                                 ┌──────────────────────┐
                                 │  Config    (独立包)  │
                                 │  - 配置文件加载       │
                                 │  - 配置项访问         │
                                 └──────────────────────┘

┌─────────────────────────────────────────────────────────┐
│            交互层 - Listener（监听器层）                  │
│  - 异步处理消息队列事件                                   │
│  - 事件驱动架构                                           │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│            交互层 - Scheduler（调度器层）                 │
│  - 定时任务执行                                           │
│  - 周期性后台任务                                         │
└─────────────────────────────────────────────────────────┘
```

### 4.2 依赖规则

```
内置管理器层（无外部依赖，由引擎自动初始化）
    ↓
Entity 层（无外部依赖）
    ↓
Repository 层（依赖 Entity、Config、Manager）
    ↓
Service 层（依赖 Repository、Config、Manager、Service）
    ↓
交互层（Controller/Middleware/Listener/Scheduler）（依赖 Service、Config、Manager）
```

**规则说明**：
- 上层可以依赖下层
- 下层不能依赖上层
- 同层之间可以相互依赖（例如 Service 可以依赖另一个 Service）
- Controller 不能直接依赖 Repository，必须通过 Service
- Config 和 Manager 作为独立包，由引擎自动初始化和注入
- Manager 包位于 `manager/` 目录，包括：configmgr, databasemgr, cachemgr, loggermgr, lockmgr, limitermgr, mqmgr, telemetrymgr
- 内置组件位于 `component/` 目录，包括：litecontroller, litemiddleware, liteservice

### 4.3 生命周期管理

所有实现了生命周期接口的组件都会在以下时机被调用：

| 方法 | 调用时机 | 用途 |
|------|----------|------|
| `OnStart()` | 服务器启动时 | 初始化资源（连接数据库、加载缓存等） |
| `OnStop()` | 服务器停止时 | 清理资源（关闭连接、保存数据等） |
| `Health()` | 健康检查时 | 检查组件健康状态（内置 Manager 组件） |

---

## 5. 5 层架构详解

### 5.1 Entity 层（实体层）

Entity 层定义数据实体，映射到数据库表结构。实体层无外部依赖，只包含纯数据定义。

#### 5.1.1 实体基类选择

框架提供 3 种预定义的实体基类，使用 CUID2 ID 和 GORM Hook 自动填充：

| 基类 | 字段 | 适用场景 |
|-----|------|---------|
| `BaseEntityOnlyID` | ID | 配置表、字典表（无需时间戳） |
| `BaseEntityWithCreatedAt` | ID, CreatedAt | 日志、审计记录（只需创建时间） |
| `BaseEntityWithTimestamps` | ID, CreatedAt, UpdatedAt | 业务实体（最常用） |

#### 5.1.2 实体示例（使用基类）

```go
package entities

import (
    "github.com/lite-lake/litecore-go/common"
)

// 使用 BaseEntityWithTimestamps 基类（最常用）
// 基类提供：
// - ID（25 位 CUID2 字符串）
// - CreatedAt（创建时间）
// - UpdatedAt（更新时间）
// - GORM Hook 自动填充
type Message struct {
    common.BaseEntityWithTimestamps
    Nickname string `gorm:"type:varchar(20);not null" json:"nickname"`
    Content  string `gorm:"type:varchar(500);not null" json:"content"`
    Status   string `gorm:"type:varchar(20);default:'pending'" json:"status"`
}

func (m *Message) EntityName() string {
    return "Message"
}

func (m *Message) TableName() string {
    return "messages"
}

func (m *Message) GetId() string {
    return m.ID
}

// IsApproved 检查留言是否已审核通过
func (m *Message) IsApproved() bool {
    return m.Status == "approved"
}

var _ common.IBaseEntity = (*Message)(nil)
```

**其他基类示例**：

```go
// BaseEntityOnlyID - 仅 ID（适合配置表、字典表）
type Config struct {
    common.BaseEntityOnlyID
    Key   string `gorm:"type:varchar(50);uniqueIndex;not null" json:"key"`
    Value string `gorm:"type:text;not null" json:"value"`
}

func (c *Config) EntityName() string { return "Config" }
func (c *Config) TableName() string { return "configs" }
func (c *Config) GetId() string { return c.ID }
var _ common.IBaseEntity = (*Config)(nil)

// BaseEntityWithCreatedAt - ID + 创建时间（适合日志、审计记录）
type AuditLog struct {
    common.BaseEntityWithCreatedAt
    Action  string `gorm:"type:varchar(50);not null" json:"action"`
    Details string `gorm:"type:text" json:"details"`
}

func (a *AuditLog) EntityName() string { return "AuditLog" }
func (a *AuditLog) TableName() string { return "audit_logs" }
func (a *AuditLog) GetId() string { return a.ID }
var _ common.IBaseEntity = (*AuditLog)(nil)
```

#### 5.1.3 GORM 标签说明

| 标签 | 说明 | 示例 |
|------|------|------|
| `primarykey` | 主键 | `gorm:"primarykey"` |
| `type` | 字段类型 | `gorm:"type:varchar(50)"` |
| `not null` | 非空约束 | `gorm:"not null"` |
| `uniqueIndex` | 唯一索引 | `gorm:"uniqueIndex"` |
| `index` | 普通索引 | `gorm:"index"` |
| `default` | 默认值 | `gorm:"default:'active'"` |
| `size` | 字段大小 | `gorm:"size:100"` |
| `column` | 列名 | `gorm:"column:user_name"` |

#### 5.1.4 实体命名规范

| 字段 | 类型 | 说明 |
|-----|------|------|
| ID | string | CUID2 25位字符串，由基类自动生成 |
| ID 存储类型 | varchar(32) | 数据库字段类型，预留兼容空间 |
| CreatedAt | time.Time | 创建时间，由基类 Hook 自动填充 |
| UpdatedAt | time.Time | 更新时间，由基类 Hook 自动填充 |

#### 5.1.5 实体设计规范

- **使用基类**：推荐使用 `BaseEntityWithTimestamps` 基类，自动生成 ID 和时间戳
- **ID 类型**：始终使用 string 类型（CUID2 25位）
- **纯数据模型**：实体只包含数据，不包含业务逻辑
- **GORM 标签**：使用 GORM 标签定义表结构
- **接口实现**：必须实现 `common.IBaseEntity` 接口
- **辅助方法**：可以添加简单的辅助方法（如 `IsApproved()`）
- **无依赖**：实体层不依赖任何其他层

---

### 5.2 Repository 层（仓储层）

Repository 层负责数据访问，提供 CRUD 操作和数据库交互。

#### 5.2.1 Repository 示例

```go
package repositories

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/configmgr"
    "github.com/lite-lake/litecore-go/manager/databasemgr"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/entities"
)

type IMessageRepository interface {
    common.IBaseRepository
    Create(message *entities.Message) error
    GetByID(id string) (*entities.Message, error)  // ID 类型为 string
    GetApprovedMessages() ([]*entities.Message, error)
    GetAllMessages() ([]*entities.Message, error)
    UpdateStatus(id string, status string) error   // ID 类型为 string
    Delete(id string) error                        // ID 类型为 string
}

type messageRepositoryImpl struct {
    Config  configmgr.IConfigManager     `inject:""`
    Manager databasemgr.IDatabaseManager `inject:""`
}

func NewMessageRepository() IMessageRepository {
    return &messageRepositoryImpl{}
}

func (r *messageRepositoryImpl) RepositoryName() string {
    return "MessageRepository"
}

func (r *messageRepositoryImpl) OnStart() error {
    return nil
}

func (r *messageRepositoryImpl) OnStop() error {
    return nil
}

func (r *messageRepositoryImpl) Create(message *entities.Message) error {
    return r.Manager.DB().Create(message).Error
}

func (r *messageRepositoryImpl) GetByID(id string) (*entities.Message, error) {
    var message entities.Message
    err := r.Manager.DB().Where("id = ?", id).First(&message).Error  // 使用 Where 查询
    if err != nil {
        return nil, err
    }
    return &message, nil
}

func (r *messageRepositoryImpl) GetApprovedMessages() ([]*entities.Message, error) {
    var messages []*entities.Message
    err := r.Manager.DB().Where("status = ?", "approved").
        Order("created_at DESC").
        Find(&messages).Error
    return messages, err
}

func (r *messageRepositoryImpl) UpdateStatus(id string, status string) error {
    return r.Manager.DB().Model(&entities.Message{}).
        Where("id = ?", id).
        Update("status", status).Error
}

func (r *messageRepositoryImpl) Delete(id string) error {
    return r.Manager.DB().Where("id = ?", id).Delete(&entities.Message{}).Error
}

var _ IMessageRepository = (*messageRepositoryImpl)(nil)
```

#### 5.2.2 Repository 设计规范

- **接口定义**：定义接口 `IXxxRepository`
- **依赖注入**：使用 `inject:""` 标签注入依赖
- **接口实现**：结构体命名为小写 `xxxRepositoryImpl`
- **ID 类型**：ID 参数类型为 string
- **查询方法**：使用 `Where("id = ?", id)` 而非 `First(entity, id)`
- **生命周期**：在 `OnStart()` 中自动迁移表结构
- **错误处理**：直接返回 GORM 错误，不在 Repository 层包装
- **事务管理**：在 Service 层管理事务

#### 5.2.3 使用事务

Repository 层只提供数据库访问方法，事务管理在 Service 层进行：

```go
// Service 层
func (s *userService) CreateUserWithProfile(user *entities.User, profile *entities.Profile) error {
    return s.Manager.DB().Transaction(func(tx *gorm.DB) error {
        // 创建用户
        if err := tx.Create(user).Error; err != nil {
            return err
        }

        // 创建用户档案
        profile.UserID = user.ID
        if err := tx.Create(profile).Error; err != nil {
            return err
        }

        return nil
    })
}
```

---

### 5.3 Service 层（服务层）

Service 层实现业务逻辑，负责数据验证、事务管理、业务规则等。

#### 5.3.1 Service 示例

```go
package services

import (
    "errors"
    "fmt"

    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/configmgr"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/entities"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/repositories"
)

type IMessageService interface {
    common.IBaseService
    CreateMessage(nickname, content string) (*entities.Message, error)
    GetApprovedMessages() ([]*entities.Message, error)
    UpdateMessageStatus(id string, status string) error  // ID 类型为 string
    DeleteMessage(id string) error                        // ID 类型为 string
}

type messageServiceImpl struct {
    Config     configmgr.IConfigManager        `inject:""`
    Repository repositories.IMessageRepository `inject:""`
    LoggerMgr  loggermgr.ILoggerManager       `inject:""`
}

func NewMessageService() IMessageService {
    return &messageServiceImpl{}
}

func (s *messageServiceImpl) ServiceName() string {
    return "MessageService"
}

func (s *messageServiceImpl) OnStart() error {
    return nil
}

func (s *messageServiceImpl) OnStop() error {
    return nil
}

func (s *messageServiceImpl) CreateMessage(nickname, content string) (*entities.Message, error) {
    // 验证输入
    if len(nickname) < 2 || len(nickname) > 20 {
        return nil, errors.New("昵称长度必须在 2-20 个字符之间")
    }
    if len(content) < 5 || len(content) > 500 {
        return nil, errors.New("留言内容长度必须在 5-500 个字符之间")
    }

    // 创建消息（ID、CreatedAt、UpdatedAt 由 Hook 自动填充）
    message := &entities.Message{
        Nickname: nickname,
        Content:  content,
        Status:   "pending",
    }

    if err := s.Repository.Create(message); err != nil {
        s.LoggerMgr.Ins().Error("创建留言失败", "error", err, "nickname", nickname)
        return nil, fmt.Errorf("创建留言失败: %w", err)
    }

    s.LoggerMgr.Ins().Info("留言创建成功", "id", message.ID, "nickname", message.Nickname)
    return message, nil
}

func (s *messageServiceImpl) UpdateMessageStatus(id string, status string) error {
    if status != "pending" && status != "approved" && status != "rejected" {
        return errors.New("无效的状态值")
    }

    message, err := s.Repository.GetByID(id)
    if err != nil {
        s.LoggerMgr.Ins().Error("获取留言失败", "error", err, "id", id)
        return fmt.Errorf("获取留言失败: %w", err)
    }
    if message == nil {
        return errors.New("留言不存在")
    }

    if err := s.Repository.UpdateStatus(id, status); err != nil {
        s.LoggerMgr.Ins().Error("更新留言状态失败", "error", err, "id", id)
        return fmt.Errorf("更新留言状态失败: %w", err)
    }

    s.LoggerMgr.Ins().Info("留言状态更新成功", "id", id, "old_status", message.Status, "new_status", status)
    return nil
}

func (s *messageServiceImpl) DeleteMessage(id string) error {
    message, err := s.Repository.GetByID(id)
    if err != nil {
        s.LoggerMgr.Ins().Error("获取留言失败", "error", err, "id", id)
        return fmt.Errorf("获取留言失败: %w", err)
    }
    if message == nil {
        return errors.New("留言不存在")
    }

    if err := s.Repository.Delete(id); err != nil {
        s.LoggerMgr.Ins().Error("删除留言失败", "error", err, "id", id)
        return fmt.Errorf("删除留言失败: %w", err)
    }

    s.LoggerMgr.Ins().Info("留言删除成功", "id", id, "nickname", message.Nickname)
    return nil
}

func (s *messageServiceImpl) GetApprovedMessages() ([]*entities.Message, error) {
    s.LoggerMgr.Ins().Debug("获取已审核留言列表")

    messages, err := s.Repository.GetApprovedMessages()
    if err != nil {
        s.LoggerMgr.Ins().Error("获取已审核留言失败", "error", err)
        return nil, fmt.Errorf("获取已审核留言失败: %w", err)
    }

    s.LoggerMgr.Ins().Debug("已审核留言列表获取成功", "count", len(messages))
    return messages, nil
}

var _ IMessageService = (*messageServiceImpl)(nil)
```

#### 5.3.2 Service 设计规范

- **业务逻辑**：在 Service 层实现所有业务逻辑
- **数据验证**：在 Service 层进行输入验证
- **错误包装**：使用 `fmt.Errorf()` 包装错误信息
- **事务管理**：在 Service 层管理数据库事务
- **依赖注入**：可以依赖 Repository、Manager、其他 Service
- **ID 处理**：ID 参数类型为 string，直接传递给 Repository
- **时间戳**：无需手动设置 CreatedAt、UpdatedAt，由 Hook 自动填充

---

### 5.4 交互层 - Controller（控制器层）

Controller 层负责 HTTP 请求处理，包括参数验证、调用 Service、响应封装。

#### 5.4.1 Controller 示例

```go
package controllers

import (
    "net/http"

    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/dtos"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/services"

    "github.com/gin-gonic/gin"
)

type IMessageController interface {
    common.IBaseController
}

type messageControllerImpl struct {
    MessageService services.IMessageService `inject:""`
    LoggerMgr      loggermgr.ILoggerManager `inject:""`
}

func NewMessageController() IMessageController {
    return &messageControllerImpl{}
}

func (c *messageControllerImpl) ControllerName() string {
    return "MessageController"
}

func (c *messageControllerImpl) GetRouter() string {
    return "/api/messages [POST],/api/messages [GET]"
}

func (c *messageControllerImpl) Handle(ctx *gin.Context) {
    method := ctx.Request.Method

    if method == "POST" {
        c.handleCreate(ctx)
    } else if method == "GET" {
        c.handleList(ctx)
    }
}

func (c *messageControllerImpl) handleCreate(ctx *gin.Context) {
    var req dtos.CreateMessageRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        c.LoggerMgr.Ins().Warn("参数验证失败", "error", err)
        ctx.JSON(http.StatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
        return
    }

    message, err := c.MessageService.CreateMessage(req.Nickname, req.Content)
    if err != nil {
        c.LoggerMgr.Ins().Warn("创建留言失败", "error", err, "nickname", req.Nickname)
        ctx.JSON(http.StatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
        return
    }

    c.LoggerMgr.Ins().Info("留言创建成功", "id", message.ID, "nickname", message.Nickname)
    ctx.JSON(http.StatusOK, dtos.SuccessResponse("创建成功", dtos.MessageResponse{
        ID:        message.ID,
        Nickname:  message.Nickname,
        Content:   message.Content,
        Status:    message.Status,
        CreatedAt: message.CreatedAt,
    }))
}

func (c *messageControllerImpl) handleList(ctx *gin.Context) {
    messages, err := c.MessageService.GetApprovedMessages()
    if err != nil {
        c.LoggerMgr.Ins().Error("获取留言列表失败", "error", err)
        ctx.JSON(http.StatusInternalServerError, dtos.ErrorResponse(common.HTTPStatusInternalServerError, err.Error()))
        return
    }

    ctx.JSON(http.StatusOK, dtos.SuccessResponse("获取成功", messages))
}

var _ IMessageController = (*messageControllerImpl)(nil)
```

#### 5.4.2 DTO 示例

```go
package dtos

import "time"

// CreateMessageRequest 创建留言请求
type CreateMessageRequest struct {
    Nickname string `json:"nickname" binding:"required,min=2,max=20"`
    Content  string `json:"content" binding:"required,min=5,max=500"`
}

// MessageResponse 留言响应
type MessageResponse struct {
    ID        string    `json:"id"`
    Nickname  string    `json:"nickname"`
    Content   string    `json:"content"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}

func SuccessResponse(message string, data interface{}) SuccessResponse {
    return SuccessResponse{
        Code:    common.HTTPStatusOK,
        Message: message,
        Data:    data,
    }
}

// ErrorResponse 错误响应
type ErrorResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

func ErrorResponse(code int, message string) ErrorResponse {
    return ErrorResponse{
        Code:    code,
        Message: message,
    }
}
```

#### 5.4.3 Controller 设计规范

- **路由定义**：在 `GetRouter()` 中定义路由
- **参数验证**：使用 Gin 的 `ShouldBindJSON()` 验证参数
- **调用 Service**：通过依赖注入调用 Service 层方法
- **响应封装**：使用统一的响应格式
- **错误处理**：将 Service 层错误转换为 HTTP 响应
- **ID 处理**：ID 直接从 `ctx.Param("id")` 获取，类型为 string，无需转换

#### 5.4.4 路由定义格式

Controller 的 `GetRouter()` 方法支持完整的路由语法：

```go
// 基本 CRUD
return "/api/messages [GET]"              // 获取列表
return "/api/messages [POST]"             // 创建
return "/api/messages/:id [GET]"          // 获取详情
return "/api/messages/:id [PUT]"          // 更新
return "/api/messages/:id [DELETE]"       // 删除

// 路径参数
return "/api/files/*filepath [GET]"    // 通配符

// 路由分组
return "/api/admin/messages [GET]"        // 管理端路由
return "/api/v1/messages [GET]"           // 版本化路由
```

---

### 5.5 交互层 - Middleware（中间件层）

Middleware 层负责横切关注点的处理，如认证、授权、日志、CORS、限流等。

#### 5.5.1 内置中间件

LiteCore 提供了多个内置中间件，位于 `component/litemiddleware` 包中。

##### 可用的内置中间件

| 中间件 | 功能 |
|--------|------|
| `RecoveryMiddleware` | Panic 恢复 |
| `CORSMiddleware` | CORS 跨域处理 |
| `RequestLoggerMiddleware` | 请求日志记录 |
| `SecurityHeadersMiddleware` | 安全头设置 |
| `RateLimiterMiddleware` | 限流保护 |
| `TelemetryMiddleware` | 遥测数据采集 |

##### 使用内置中间件

所有内置中间件都支持可选配置，使用指针类型实现部分配置：

```go
package middlewares

import (
    "time"
    "github.com/lite-lake/litecore-go/component/litemiddleware"
)

// 使用默认配置
func NewCorsMiddleware() common.IBaseMiddleware {
    return litemiddleware.NewCorsMiddleware(nil)
}

// 自定义 CORS 配置
func NewCustomCorsMiddleware() common.IBaseMiddleware {
    allowOrigins := []string{"https://example.com"}
    allowCredentials := true
    return litemiddleware.NewCorsMiddleware(&litemiddleware.CorsConfig{
        AllowOrigins:     &allowOrigins,
        AllowCredentials: &allowCredentials,
    })
}

// 自定义限流中间件
func NewRateLimiterMiddleware() common.IBaseMiddleware {
    limit := 100
    window := time.Minute
    return litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
        Limit:     &limit,
        Window:    &window,
        KeyPrefix: strPtr("api_rate_limit"),
    })
}

func strPtr(s string) *string {
    return &s
}
```

##### 中间件配置结构

所有中间件配置都支持 `Name` 和 `Order` 字段，用于自定义中间件名称和执行顺序：

```go
type CorsConfig struct {
    Name             *string       // 中间件名称
    Order            *int          // 执行顺序
    AllowOrigins     *[]string     // 允许的源
    AllowMethods     *[]string     // 允许的 HTTP 方法
    AllowHeaders     *[]string     // 允许的请求头
    AllowCredentials *bool         // 是否允许携带凭证
    MaxAge           *time.Duration // 预检请求缓存时间
}

type RateLimiterConfig struct {
    Name      *string       // 中间件名称
    Order     *int          // 执行顺序
    Limit     *int          // 时间窗口内最大请求数
    Window    *time.Duration // 时间窗口大小
    KeyFunc   KeyFunc       // 自定义key生成函数
    SkipFunc  SkipFunc      // 跳过限流的条件
    KeyPrefix *string       // key前缀
}
```

##### 中间件执行顺序

预定义的中间件执行顺序（按 Order 值从小到大）：

| 中间件 | 默认 Order | 说明 |
|--------|-----------|------|
| Recovery | 0 | panic 恢复（最先执行） |
| RequestLogger | 50 | 请求日志 |
| CORS | 100 | 跨域处理 |
| SecurityHeaders | 150 | 安全头 |
| RateLimiter | 200 | 限流 |
| Telemetry | 250 | 遥测 |

业务自定义中间件建议从 Order 350 开始。可通过配置覆盖默认顺序：

```go
order := 150
limit := 100
window := time.Minute
name := "MyRateLimiter"
litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
    Name:      &name,
    Order:     &order,
    Limit:     &limit,
    Window:    &window,
})
```

#### 5.5.2 限流器中间件详解

限流器中间件提供基于 Gin 框架的 HTTP 请求限流功能，支持多种限流策略。

##### 基本用法

```go
package middlewares

import (
    "time"
    "github.com/lite-lake/litecore-go/component/litemiddleware"
)

// 创建按 IP 限流的中间件（默认配置）
func NewRateLimiterMiddleware() common.IBaseMiddleware {
    return litemiddleware.NewRateLimiterMiddleware(nil)
}
```

##### 自定义限流策略

```go
// 自定义 Key 生成函数（基于用户ID）
func NewUserRateLimiterMiddleware() common.IBaseMiddleware {
    limit := 60
    window := time.Minute
    keyPrefix := "user_rate_limit"
    return litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
        Limit:     &limit,
        Window:    &window,
        KeyPrefix: &keyPrefix,
        KeyFunc: func(c *gin.Context) string {
            if userID, exists := c.Get("user_id"); exists {
                if uid, ok := userID.(string); ok {
                    return uid
                }
            }
            return c.ClientIP()
        },
    })
}

// 添加跳过逻辑（公开接口不限流）
func NewRateLimiterWithSkip() common.IBaseMiddleware {
    limit := 100
    window := time.Minute
    return litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
        Limit:  &limit,
        Window: &window,
        SkipFunc: func(c *gin.Context) bool {
            return c.Request.URL.Path == "/api/health" ||
                   c.Request.URL.Path == "/api/public"
        },
    })
}
```

##### 响应头说明

限流器会在响应中添加以下头信息：

| 响应头 | 说明 |
|--------|------|
| `X-RateLimit-Limit` | 时间窗口内的最大请求数 |
| `X-RateLimit-Remaining` | 当前窗口剩余的请求数 |
| `Retry-After` | 被限流时，建议的等待时间（秒） |

##### 限流响应示例

```json
// 请求成功
Status: 200 OK
Headers:
  X-RateLimit-Limit: 100
  X-RateLimit-Remaining: 99

// 被限流
Status: 429 Too Many Requests
Body: {
  "error": "请求过于频繁，请 1m0s 后再试",
  "code": "RATE_LIMIT_EXCEEDED"
}
```

#### 5.5.3 认证中间件示例

```go
package middlewares

import (
    "strings"

    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/configmgr"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/services"

    "github.com/gin-gonic/gin"
)

type IAuthMiddleware interface {
    common.IBaseMiddleware
}

type authMiddleware struct {
    Config      configmgr.IConfigManager `inject:""`
    LoggerMgr   loggermgr.ILoggerManager `inject:""`
    AuthService services.IAuthService    `inject:""`
}

func NewAuthMiddleware() IAuthMiddleware {
    return &authMiddleware{}
}

func (m *authMiddleware) MiddlewareName() string {
    return "AuthMiddleware"
}

func (m *authMiddleware) Order() int {
    return 300
}

func (m *authMiddleware) Wrapper() gin.HandlerFunc {
    return func(c *gin.Context) {

        // 跳过公开路由
        if strings.HasPrefix(c.Request.URL.Path, "/api/public") {
            c.Next()
            return
        }

        // 获取 Authorization header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            m.LoggerMgr.Ins().Warn("未提供认证令牌", "path", c.Request.URL.Path)
            c.JSON(common.HTTPStatusUnauthorized, gin.H{
                "code":    common.HTTPStatusUnauthorized,
                "message": "未提供认证令牌",
            })
            c.Abort()
            return
        }

        // 解析 Bearer token
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            m.LoggerMgr.Ins().Warn("认证令牌格式错误", "path", c.Request.URL.Path)
            c.JSON(common.HTTPStatusUnauthorized, gin.H{
                "code":    common.HTTPStatusUnauthorized,
                "message": "认证令牌格式错误",
            })
            c.Abort()
            return
        }

        token := parts[1]

        // 验证 token
        session, err := m.AuthService.ValidateToken(token)
        if err != nil {
            m.LoggerMgr.Ins().Warn("认证令牌无效", "path", c.Request.URL.Path, "error", err)
            c.JSON(common.HTTPStatusUnauthorized, gin.H{
                "code":    common.HTTPStatusUnauthorized,
                "message": "认证令牌无效或已过期",
            })
            c.Abort()
            return
        }

        // 将用户信息存入上下文
        c.Set("session", session)
        c.Next()
    }
}

func (m *authMiddleware) OnStart() error {
    return nil
}

func (m *authMiddleware) OnStop() error {
    return nil
}

var _ IAuthMiddleware = (*authMiddleware)(nil)
```

#### 5.5.4 中间件执行顺序

中间件按照 `Order()` 返回的值从小到大执行：

```go
// 推荐的中间件顺序
func (m *RecoveryMiddleware) Order() int         { return 10 }   // 恢复中间件
func (m *CORSMiddleware) Order() int              { return 20 }   // CORS 中间件
func (m *RateLimiterMiddleware) Order() int       { return 30 }   // 限流中间件
func (m *AuthMiddleware) Order() int              { return 300 }  // 认证中间件
func (m *LoggerMiddleware) Order() int           { return 200 }  // 日志中间件
func (m *TelemetryMiddleware) Order() int        { return 250 }  // 遥测中间件
```

**说明**：
- 限流中间件应放在 CORS 之后、认证之前，这样可以：
  1. 正常处理跨域请求
  2. 对所有请求（包括未认证请求）进行限流保护
  3. 在认证之前拦截恶意请求，减少认证服务压力
- 内置中间件默认 Order 值：
  - Recovery: 0
  - RequestLogger: 50
  - CORS: 100
  - SecurityHeaders: 150
  - RateLimiter: 200
  - Telemetry: 250

#### 5.5.5 中间件设计规范

- **横切关注点**：中间件处理认证、日志、CORS 等横切关注点
- **顺序控制**：使用 `Order()` 方法定义执行顺序，或通过配置覆盖
- **上下文存储**：使用 `c.Set()` 和 `c.Get()` 存储上下文信息
- **提前终止**：使用 `c.Abort()` 提前终止请求处理
- **依赖注入**：使用 `inject:""` 标签注入所需 Manager
- **配置化**：内置中间件支持可选配置，提供灵活的定制能力

---

### 5.6 交互层 - Listener（监听器层）

Listener 层负责异步处理消息队列事件，实现事件驱动架构。

#### 5.6.1 Listener 接口

所有监听器需要实现 `common.IBaseListener` 接口：

```go
type IBaseListener interface {
    ListenerName() string                     // 监听器名称
    GetQueue() string                          // 监听的队列名称
    GetSubscribeOptions() []ISubscribeOption  // 订阅配置选项
    OnStart() error                            // 启动回调
    OnStop() error                             // 停止回调
    Handle(ctx context.Context, msg IMessageListener) error  // 处理消息
}
```

#### 5.6.2 Listener 示例

```go
package listeners

import (
    "context"

    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
)

type IMessageCreatedListener interface {
    common.IBaseListener
}

type messageCreatedListenerImpl struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func NewMessageCreatedListener() IMessageCreatedListener {
    return &messageCreatedListenerImpl{}
}

func (l *messageCreatedListenerImpl) ListenerName() string {
    return "MessageCreatedListener"
}

func (l *messageCreatedListenerImpl) GetQueue() string {
    return "message.created"  // 监听的队列名称
}

func (l *messageCreatedListenerImpl) GetSubscribeOptions() []common.ISubscribeOption {
    return []common.ISubscribeOption{}
}

func (l *messageCreatedListenerImpl) OnStart() error {
    l.LoggerMgr.Ins().Info("留言创建监听器启动")
    return nil
}

func (l *messageCreatedListenerImpl) OnStop() error {
    l.LoggerMgr.Ins().Info("留言创建监听器停止")
    return nil
}

func (l *messageCreatedListenerImpl) Handle(ctx context.Context, msg common.IMessageListener) error {
    l.LoggerMgr.Ins().Info("收到留言创建事件",
        "message_id", msg.ID(),
        "body", string(msg.Body()),
        "headers", msg.Headers())
    return nil
}

var _ IMessageCreatedListener = (*messageCreatedListenerImpl)(nil)
var _ common.IBaseListener = (*messageCreatedListenerImpl)(nil)
```

#### 5.6.3 发送消息

在 Service 中通过 `MQManager` 发送消息到指定队列：

```go
package services

import (
    "encoding/json"

    "github.com/lite-lake/litecore-go/manager/mqmgr"
)

type messageServiceImpl struct {
    MQManager mqmgr.IMQManager `inject:""`
}

func (s *messageServiceImpl) CreateMessage(nickname, content string) (*entities.Message, error) {
    // ... 创建留言逻辑 ...

    // 发送消息到队列
    if s.MQManager != nil {
        messageBody, _ := json.Marshal(map[string]interface{}{
            "id":       message.ID,
            "nickname": message.Nickname,
            "content":  message.Content,
        })
        s.MQManager.Publish(context.Background(), "message.created", messageBody)
    }

    return message, nil
}
```

#### 5.6.4 Listener 设计规范

- **队列名称**：通过 `GetQueue()` 指定监听的队列名称
- **消息处理**：在 `Handle()` 方法中处理消息
- **错误处理**：消息处理失败会自动重试（取决于 MQ 配置）
- **异步处理**：Listener 不会阻塞主请求流程
- **事件驱动**：使用 Listener 实现松耦合的事件驱动架构

---

### 5.7 交互层 - Scheduler（调度器层）

Scheduler 层负责执行定时任务，如数据统计、清理等后台任务。

#### 5.7.1 Scheduler 接口

所有调度器需要实现 `common.IBaseScheduler` 接口：

```go
type IBaseScheduler interface {
    SchedulerName() string        // 调度器名称
    GetRule() string             // Cron 表达式
    GetTimezone() string         // 时区
    OnTick(tickID int64) error   // 定时任务执行回调
    OnStart() error              // 启动回调
    OnStop() error               // 停止回调
}
```

#### 5.7.2 Scheduler 示例

```go
package schedulers

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/services"
)

type IStatisticsScheduler interface {
    common.IBaseScheduler
}

type statisticsSchedulerImpl struct {
    MessageService services.IMessageService `inject:""`
    LoggerMgr      loggermgr.ILoggerManager `inject:""`
}

func NewStatisticsScheduler() IStatisticsScheduler {
    return &statisticsSchedulerImpl{}
}

func (s *statisticsSchedulerImpl) SchedulerName() string {
    return "statisticsScheduler"
}

func (s *statisticsSchedulerImpl) GetRule() string {
    return "0 0 * * * *"  // 每小时执行一次
}

func (s *statisticsSchedulerImpl) GetTimezone() string {
    return "Asia/Shanghai"
}

func (s *statisticsSchedulerImpl) OnTick(tickID int64) error {
    s.LoggerMgr.Ins().Info("开始执行统计任务", "tick_id", tickID)

    stats, err := s.MessageService.GetStatistics()
    if err != nil {
        s.LoggerMgr.Ins().Error("获取统计信息失败", "error", err)
        return err
    }

    s.LoggerMgr.Ins().Info("统计任务完成",
        "tick_id", tickID,
        "pending", stats["pending"],
        "approved", stats["approved"],
        "rejected", stats["rejected"],
        "total", stats["total"])
    return nil
}

func (s *statisticsSchedulerImpl) OnStart() error {
    s.LoggerMgr.Ins().Info("统计调度器启动")
    return nil
}

func (s *statisticsSchedulerImpl) OnStop() error {
    s.LoggerMgr.Ins().Info("统计调度器停止")
    return nil
}

var _ IStatisticsScheduler = (*statisticsSchedulerImpl)(nil)
var _ common.IBaseScheduler = (*statisticsSchedulerImpl)(nil)
```

#### 5.7.3 Cron 表达式

项目使用标准 Cron 表达式格式：`秒 分 时 日 月 周`

| 表达式 | 说明 |
|--------|------|
| `0 * * * * *` | 每分钟执行 |
| `0 0 * * * *` | 每小时执行 |
| `0 0 2 * * *` | 每天凌晨 2 点执行 |
| `0 0 0 * * 1` | 每周一凌晨执行 |
| `0 0 0 1 * *` | 每月 1 号凌晨执行 |

#### 5.7.4 Scheduler 设计规范

- **Cron 表达式**：使用标准 Cron 表达式定义执行规则
- **时区**：通过 `GetTimezone()` 指定时区（推荐：`Asia/Shanghai`）
- **错误处理**：任务执行失败会记录日志，不会影响其他任务
- **独立执行**：每个调度器独立执行，互不影响
- **周期性任务**：适合执行周期性后台任务（统计、清理等）

---

## 6. 内置组件

### 6.1 Config（配置）

Config 作为服务器内置组件，由引擎自动初始化。在创建引擎时通过 `server.BuiltinConfig` 指定配置文件：

```go
// 引擎自动生成的代码会创建 Config
func NewEngine() (*server.Engine, error) {
    // ...
    return server.NewEngine(
        &server.BuiltinConfig{
            Driver:   "yaml",
            FilePath: "configs/config.yaml",
        },
        entityContainer,
        repositoryContainer,
        serviceContainer,
        controllerContainer,
        middlewareContainer,
        listenerContainer,
        schedulerContainer,
    ), nil
}
```

### 6.2 Manager（管理器）

Manager 组件作为独立包位于 `manager` 目录，由引擎自动初始化。在 `Initialize()` 时自动初始化所有 Manager：

```go
// 框架自动初始化的 Manager（按顺序）
// 1. ConfigManager - 配置管理 (manager/configmgr)
// 2. TelemetryManager - 遥测管理 (manager/telemetrymgr)
// 3. LoggerManager - 日志管理 (manager/loggermgr)
// 4. DatabaseManager - 数据库管理 (manager/databasemgr)
// 5. CacheManager - 缓存管理 (manager/cachemgr)
// 6. LockManager - 锁管理 (manager/lockmgr)
// 7. LimiterManager - 限流管理 (manager/limitermgr)
// 8. MQManager - 消息队列管理 (manager/mqmgr)
```

无需手动创建 Manager，只需在代码中通过依赖注入使用：

```go
type userRepository struct {
    Manager databasemgr.IDatabaseManager `inject:""`
}
```

### 6.3 LockMgr（锁管理器）

LockMgr 位于 `manager/lockmgr` 包，提供分布式锁功能，支持 Redis 和内存两种实现，用于解决并发访问和资源竞争问题。

#### 6.3.1 接口定义

```go
type ILockManager interface {
    // Lock 获取锁（阻塞直到成功或超时）
    Lock(ctx context.Context, key string, ttl time.Duration) error

    // Unlock 释放锁
    Unlock(ctx context.Context, key string) error

    // TryLock 尝试获取锁（非阻塞）
    TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
}
```

#### 6.3.2 使用示例

**在 Service 层使用锁**

```go
package services

import (
    "context"
    "time"

    "github.com/lite-lake/litecore-go/manager/configmgr"
    "github.com/lite-lake/litecore-go/manager/lockmgr"
)

type OrderService struct {
    Config  configmgr.IConfigManager `inject:""`
    LockMgr lockmgr.ILockManager    `inject:""`
}

// ProcessOrder 处理订单（使用分布式锁防止重复处理）
func (s *OrderService) ProcessOrder(ctx context.Context, orderID string) error {
    lockKey := "order:process:" + orderID

    // 尝试获取锁，30秒后过期
    acquired, err := s.LockMgr.TryLock(ctx, lockKey, 30*time.Second)
    if err != nil {
        return err
    }
    if !acquired {
        return errors.New("订单正在处理中，请稍后重试")
    }
    defer s.LockMgr.Unlock(ctx, lockKey)

    // 执行业务逻辑
    return s.processOrderInternal(ctx, orderID)
}

// UpdateInventory 更新库存（使用阻塞锁）
func (s *OrderService) UpdateInventory(ctx context.Context, productID string, quantity int) error {
    lockKey := "inventory:update:" + productID

    // 获取锁，最多等待10秒，锁自动过期30秒
    err := s.LockMgr.Lock(ctx, lockKey, 30*time.Second)
    if err != nil {
        return err
    }
    defer s.LockMgr.Unlock(ctx, lockKey)

    // 执行库存更新
    return s.updateInventoryInternal(ctx, productID, quantity)
}
```

#### 6.3.3 使用场景

- **防止重复处理**：订单、任务等幂等性控制
- **资源竞争**：库存扣减、余额更新等并发场景
- **定时任务**：防止多个实例同时执行同一任务
- **缓存重建**：防止缓存击穿时的并发重建

### 6.4 LimiterMgr（限流管理器）

LimiterMgr 位于 `manager/limitermgr` 包，提供限流功能，支持 Redis 和内存两种实现，用于保护系统免受过量请求的影响。

#### 6.4.1 接口定义

```go
type ILimiterManager interface {
    // Allow 检查是否允许通过限流
    Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)

    // GetRemaining 获取剩余可访问次数
    GetRemaining(ctx context.Context, key string, limit int, window time.Duration) (int, error)
}
```

#### 6.4.2 使用示例

**在 Service 层使用限流**

```go
package services

import (
    "context"
    "fmt"
    "time"

    "github.com/lite-lake/litecore-go/manager/configmgr"
    "github.com/lite-lake/litecore-go/manager/limitermgr"
)

type SMSService struct {
    Config     configmgr.IConfigManager    `inject:""`
    LimiterMgr limitermgr.ILimiterManager `inject:""`
}

// SendSMS 发送短信（按手机号限流）
func (s *SMSService) SendSMS(ctx context.Context, phone string) error {
    limitKey := "sms:send:" + phone

    // 每个手机号每分钟最多发送 5 条短信
    allowed, err := s.LimiterMgr.Allow(ctx, limitKey, 5, time.Minute)
    if err != nil {
        return err
    }

    if !allowed {
        return errors.New("发送频率过高，请稍后重试")
    }

    // 获取剩余次数
    remaining, _ := s.LimiterMgr.GetRemaining(ctx, limitKey, 5, time.Minute)
    fmt.Printf("剩余可发送次数: %d\n", remaining)

    // 发送短信逻辑
    return s.sendSMSInternal(ctx, phone)
}

// CreateOrder 创建订单（按用户限流）
func (s *OrderService) CreateOrder(ctx context.Context, userID string) error {
    limitKey := "order:create:" + userID

    // 每个用户每分钟最多创建 10 个订单
    allowed, err := s.LimiterMgr.Allow(ctx, limitKey, 10, time.Minute)
    if err != nil {
        return err
    }

    if !allowed {
        return errors.New("操作过于频繁，请稍后重试")
    }

    // 创建订单逻辑
    return s.createOrderInternal(ctx, userID)
}
```

#### 6.4.3 使用场景

- **API 限流**：保护 API 免受 DDoS 攻击
- **用户限流**：防止恶意用户频繁操作
- **资源保护**：限制短信、邮件等高成本资源的使用
- **服务降级**：在系统负载过高时进行限流

### 6.5 MQMgr（消息队列管理器）

MQMgr 位于 `manager/mqmgr` 包，提供消息队列功能，支持 RabbitMQ 和内存两种实现。

#### 6.5.1 接口定义

```go
type IMQManager interface {
    // Publish 发布消息到指定队列
    Publish(ctx context.Context, queue string, message []byte, options ...PublishOption) error

    // Subscribe 订阅指定队列，返回消息通道
    Subscribe(ctx context.Context, queue string, options ...SubscribeOption) (<-chan Message, error)

    // SubscribeWithCallback 使用回调函数订阅指定队列
    SubscribeWithCallback(ctx context.Context, queue string, handler MessageHandler, options ...SubscribeOption) error

    // Ack 确认消息已处理
    Ack(ctx context.Context, message Message) error

    // Nack 拒绝消息，可选择是否重新入队
    Nack(ctx context.Context, message Message, requeue bool) error

    // QueueLength 获取队列长度
    QueueLength(ctx context.Context, queue string) (int64, error)

    // Purge 清空队列
    Purge(ctx context.Context, queue string) error

    // Close 关闭管理器
    Close() error
}
```

#### 6.5.2 使用示例

**发布消息**

```go
package services

import (
    "context"
    "encoding/json"

    "github.com/lite-lake/litecore-go/manager/mqmgr"
)

type NotificationService struct {
    MQMgr mqmgr.IMQManager `inject:""`
}

// SendNotification 发送通知消息
func (s *NotificationService) SendNotification(ctx context.Context, userID string, message string) error {
    notification := map[string]interface{}{
        "user_id": userID,
        "message": message,
    }

    body, err := json.Marshal(notification)
    if err != nil {
        return err
    }

    return s.MQMgr.Publish(ctx, "notifications.send", body)
}
```

**订阅消息（Listener 层）**

监听器（Listener）层会自动订阅指定的队列，处理消息。详见 [5.6 Listener 层](#56-交互层---listener监听器层)。

#### 6.5.3 使用场景

- **异步处理**：将耗时操作放入队列异步处理
- **事件驱动**：实现松耦合的事件驱动架构
- **系统解耦**：服务之间通过消息队列通信
- **流量削峰**：通过队列缓冲突发流量

### 6.6 可用的内置 Manager

| Manager | 接口 | 说明 |
|---------|------|------|
| Config Manager | `configmgr.IConfigManager` | 配置管理器，支持 YAML 配置 |
| Database Manager | `databasemgr.IDatabaseManager` | 数据库管理器，支持 MySQL、PostgreSQL、SQLite |
| Cache Manager | `cachemgr.ICacheManager` | 缓存管理器，支持 Redis、Memory（Ristretto）、None |
| Logger Manager | `loggermgr.ILoggerManager` | 日志管理器，支持 Gin、JSON、Default 格式 |
| Telemetry Manager | `telemetrymgr.ITelemetryManager` | 遥测管理器，支持 OTel |
| Limiter Manager | `limitermgr.ILimiterManager` | 限流管理器，支持 Redis、Memory |
| Lock Manager | `lockmgr.ILockManager` | 锁管理器，支持 Redis、Memory |
| MQ Manager | `mqmgr.IMQManager` | 消息队列管理器，支持 RabbitMQ、Memory |

### 6.7 使用内置组件

所有 Manager 组件都可以通过依赖注入使用：

```go
type userServiceImpl struct {
    Config     configmgr.IConfigManager     `inject:""`
    DBManager  databasemgr.IDatabaseManager `inject:""`
    CacheMgr   cachemgr.ICacheManager      `inject:""`
    LoggerMgr  loggermgr.ILoggerManager    `inject:""`
    LockMgr    lockmgr.ILockManager       `inject:""`
    LimiterMgr limitermgr.ILimiterManager `inject:""`
    MQManager  mqmgr.IMQManager           `inject:""`
}
```

### 6.8 日志配置（Gin 格式）

日志配置位于 `configs/config.yaml`：

```yaml
logger:
  driver: "zap"                 # zap, default, none
  zap_config:
    telemetry_enabled: false    # 是否启用观测日志
    telemetry_config:
      level: "info"             # 日志级别
    console_enabled: true       # 是否启用控制台日志
    console_config:
      level: "info"             # 日志级别：debug, info, warn, error, fatal
      format: "gin"             # 格式：gin | json | default
      color: true               # 是否启用颜色
      time_format: "2006-01-02 15:04:05.000"  # 时间格式
    file_enabled: false         # 是否启用文件日志
    file_config:
      level: "info"             # 日志级别
      path: "./logs/myapp.log"
      rotation:
        max_size: 100           # 单个日志文件最大大小（MB）
        max_age: 30             # 日志文件保留天数
        max_backups: 10         # 保留的旧日志文件最大数量
        compress: true          # 是否压缩旧日志文件
```

#### 日志格式说明

| 格式 | 说明 |
|------|------|
| `gin` | Gin 风格，竖线分隔符，适合控制台输出（默认格式） |
| `json` | JSON 格式，适合日志分析和监控 |
| `default` | 默认 ConsoleEncoder 格式 |

#### Gin 格式特点

- 统一格式：`{时间} | {级别} | {消息} | {字段1}={值1} {字段2}={值2} ...`
- 时间固定宽度 23 字符：`2006-01-02 15:04:05.000`
- 级别固定宽度 5 字符，右对齐，带颜色
- 字段格式：`key=value`，字符串值用引号包裹

#### 日志级别颜色

| 级别 | ANSI 颜色 | 说明 |
|------|-----------|------|
| DEBUG | 灰色 | 开发调试信息 |
| INFO | 绿色 | 正常业务流程 |
| WARN | 黄色 | 降级处理、慢查询 |
| ERROR | 红色 | 业务错误、操作失败 |
| FATAL | 红色+粗体 | 致命错误 |

#### 日志使用示例

```go
// 在 Service 层使用日志
type MyService struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func (s *MyService) SomeMethod() error {
    // Debug 级别：开发调试信息
    s.LoggerMgr.Ins().Debug("开始处理请求", "param", value)

    // Info 级别：正常业务流程
    s.LoggerMgr.Ins().Info("操作成功", "id", id)

    // Warn 级别：降级处理、慢查询
    s.LoggerMgr.Ins().Warn("慢查询检测", "duration", "1.2s")

    // Error 级别：业务错误、操作失败
    s.LoggerMgr.Ins().Error("操作失败", "error", err)

    // Fatal 级别：致命错误
    s.LoggerMgr.Ins().Fatal("系统崩溃", "error", err)

    return nil
}
```

### 6.9 启动日志

启动日志配置位于 `configs/config.yaml`：

```yaml
server:
  startup_log:                  # 启动日志配置
    enabled: true               # 是否启用启动日志
    async: true                 # 是否异步日志
    buffer: 100                 # 日志缓冲区大小
```

启动日志会在应用启动时记录各组件的初始化情况，便于排查启动问题。

---

## 7. 代码生成器使用

### 7.1 命令行使用

```bash
# 使用默认配置生成
litecore-cli generate

# 指定项目路径
litecore-cli generate --project /path/to/project

# 指定输出目录
litecore-cli generate --output internal/application

# 指定包名
litecore-cli generate --package application

# 指定配置文件路径
litecore-cli generate --config configs/config.yaml

# 完整示例
litecore-cli generate \
  --project . \
  --output internal/application \
  --package application \
  --config configs/config.yaml
```

### 7.2 作为库使用

```go
package main

import (
    "os"

    "github.com/lite-lake/litecore-go/cli/generator"
)

func main() {
    cfg := generator.DefaultConfig()

    // 修改配置
    cfg.ProjectPath = "."
    cfg.OutputDir = "internal/container"
    cfg.PackageName = "container"
    cfg.ConfigPath = "configs/config.yaml"

    if err := generator.Run(cfg); err != nil {
        os.Exit(1)
    }
}
```

### 7.3 生成的文件

```
internal/application/
├── entity_container.go      # 实体容器初始化
├── repository_container.go  # 仓储容器初始化
├── service_container.go     # 服务容器初始化
├── controller_container.go  # 控制器容器初始化
├── middleware_container.go  # 中间件容器初始化
├── listener_container.go    # 监听器容器初始化
├── scheduler_container.go   # 调度器容器初始化
└── engine.go                # 引擎创建函数
```

### 7.4 使用生成的代码

```go
package main

import (
    "log"

    app "github.com/example/project/internal/application"
)

func main() {
    engine, err := app.NewEngine()
    if err != nil {
        log.Fatalf("Failed to create engine: %v", err)
    }

    if err := engine.Initialize(); err != nil {
        log.Fatalf("Failed to initialize engine: %v", err)
    }

    if err := engine.Start(); err != nil {
        log.Fatalf("Failed to start engine: %v", err)
    }

    engine.WaitForShutdown()
}
```

---

## 8. 依赖注入机制

### 8.1 声明式依赖注入

LiteCore 使用声明式依赖注入，通过 `inject:""` 标签自动注入依赖：

```go
type messageServiceImpl struct {
    Config     configmgr.IConfigManager        `inject:""`
    Repository repositories.IMessageRepository `inject:""`
    LoggerMgr  loggermgr.ILoggerManager       `inject:""`
    MQManager  mqmgr.IMQManager               `inject:""`
}
```

### 8.2 注入规则

- **内置组件**：Config、DatabaseManager、CacheManager、LoggerManager、TelemetryManager、LimiterManager、LockManager、MQManager 由引擎自动注入
- **应用组件**：Repository、Service、Controller、Middleware、Listener、Scheduler 由容器自动注入
- **工厂函数**：所有工厂函数不接受参数（`NewXxx()`），所有依赖通过 `inject:""` 标签注入
- **初始化顺序**：Entity → Repository → Service → Controller/Middleware/Listener/Scheduler

### 8.3 依赖注入示例

```go
// Repository 层：可以注入内置 Manager
type messageRepositoryImpl struct {
    Config  configmgr.IConfigManager     `inject:""`
    Manager databasemgr.IDatabaseManager `inject:""`
}

// Service 层：可以注入内置 Manager 和 Repository
type messageServiceImpl struct {
    Config     configmgr.IConfigManager        `inject:""`
    Repository repositories.IMessageRepository `inject:""`
    LoggerMgr  loggermgr.ILoggerManager       `inject:""`
    MQManager  mqmgr.IMQManager               `inject:""`
}

// Controller 层：可以注入内置 Manager 和 Service
type messageControllerImpl struct {
    MessageService services.IMessageService `inject:""`
    LoggerMgr      loggermgr.ILoggerManager `inject:""`
}

// Listener 层：可以注入内置 Manager
type messageCreatedListenerImpl struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

// Scheduler 层：可以注入内置 Manager 和 Service
type statisticsSchedulerImpl struct {
    MessageService services.IMessageService `inject:""`
    LoggerMgr      loggermgr.ILoggerManager `inject:""`
}
```

---

## 9. 配置管理

### 9.1 配置文件结构

配置文件位于 `configs/config.yaml`，包含以下主要配置项：

```yaml
# 应用配置
app:
  name: "myapp"
  version: "1.0.0"

# 服务器配置
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"                 # debug, release, test
  read_timeout: "10s"
  write_timeout: "10s"
  idle_timeout: "60s"
  enable_recovery: true
  shutdown_timeout: "30s"
  startup_log:                  # 启动日志配置
    enabled: true               # 是否启用启动日志
    async: true                 # 是否异步日志
    buffer: 100                 # 日志缓冲区大小

# 数据库配置
database:
  driver: "sqlite"              # mysql, postgresql, sqlite, none
  auto_migrate: true            # 是否自动迁移数据库表结构
  sqlite_config:
    dsn: "./data/myapp.db"
    pool_config:
      max_open_conns: 1
      max_idle_conns: 1
  observability_config:         # 可观测性配置
    slow_query_threshold: "1s"
    log_sql: false

# 缓存配置（基于 Ristretto）
cache:
  driver: "memory"              # redis, memory, none
  memory_config:
    max_size: 100               # 最大缓存大小（MB）
    max_age: "720h"             # 最大缓存时间
    max_backups: 1000           # 最大备份项数
    compress: false             # 是否压缩

# 日志配置
logger:
  driver: "zap"                 # zap, default, none
  zap_config:
    console_enabled: true       # 是否启用控制台日志
    console_config:
      level: "info"             # 日志级别
      format: "gin"             # 格式：gin | json | default
      color: true               # 是否启用颜色
      time_format: "2006-01-02 15:04:05.000"
    file_enabled: false         # 是否启用文件日志

# 限流配置
limiter:
  driver: "memory"              # redis, memory, none
  memory_config:
    max_backups: 1000           # 最大备份项数

# 锁配置
lock:
  driver: "memory"              # redis, memory, none
  memory_config:
    max_backups: 1000           # 最大备份项数

# 消息队列配置
mq:
  driver: "memory"              # rabbitmq, memory, none
  memory_config:
    max_queue_size: 10000       # 最大队列大小
    channel_buffer: 100          # 通道缓冲区大小

# 遥测配置
telemetry:
  driver: "none"                # none, otel

# 定时任务配置
scheduler:
  driver: "cron"               # cron
  cron_config:
    validate_on_startup: true  # 启动时是否检查所有 Scheduler 配置
```

### 9.2 配置文件路径

配置文件路径在 `cmd/generate/main.go` 中指定：

```go
cfg := generator.DefaultConfig()
cfg.ConfigPath = "configs/config.yaml"
```

或在命令行中指定：

```bash
litecore-cli generate --config configs/config.yaml
```

### 9.3 切换数据库驱动

```yaml
# 切换到 MySQL
database:
  driver: "mysql"
  mysql_config:
    dsn: "root:password@tcp(localhost:3306)/myapp?charset=utf8mb4&parseTime=True&loc=Local"

# 切换到 PostgreSQL
database:
  driver: "postgresql"
  postgresql_config:
    dsn: "host=localhost port=5432 user=postgres password= dbname=myapp sslmode=disable"
```

### 9.4 切换缓存/限流/锁驱动

```yaml
# 切换到 Redis
cache:
  driver: "redis"
  redis_config:
    host: "localhost"
    port: 6379
    password: ""
    db: 0

limiter:
  driver: "redis"
  redis_config:
    host: "localhost"
    port: 6379
    password: ""
    db: 0

lock:
  driver: "redis"
  redis_config:
    host: "localhost"
    port: 6379
    password: ""
    db: 0
```

---

## 10. 实用工具（util 包）

LiteCore 提供了一系列实用的工具包，帮助开发者处理常见的开发任务。

### 10.1 JWT 工具

```go
import "github.com/lite-lake/litecore-go/util/jwt"

// 生成 HS256 Token
token, err := jwt.GenerateHS256Token(claims, secret)

// 解析和验证 Token
claims, err := jwt.ParseHS256Token(token, secret)
```

### 10.2 哈希工具

```go
import "github.com/lite-lake/litecore-go/util/hash"

// MD5
md5Hash := hash.MD5("data")

// SHA1
sha1Hash := hash.SHA1("data")

// SHA256
sha256Hash := hash.SHA256("data")
```

### 10.3 加密工具

```go
import "github.com/lite-lake/litecore-go/util/crypt"

// 密码加密（bcrypt）
hashedPassword, err := crypt.HashPassword("password")

// 验证密码
err := crypt.CheckPassword("password", hashedPassword)

// AES 加密
encrypted, err := crypt.AESEncrypt("plaintext", "key")

// AES 解密
decrypted, err := crypt.AESDecrypt(encrypted, "key")
```

### 10.4 ID 生成工具

```go
import "github.com/lite-lake/litecore-go/util/id"

// 生成雪花算法 ID
snowflakeID := id.NewSnowflakeID()

// 生成 UUID
uuidV4 := id.NewUUIDV4()
```

---

## 11. 最佳实践

### 11.1 使用实体基类

- **推荐使用 `BaseEntityWithTimestamps`**：自动生成 CUID2 ID 和时间戳
- **ID 类型**：始终使用 string 类型（CUID2 25位）
- **Repository 查询**：使用 `Where("id = ?", id)` 而非 `First(entity, id)`
- **Service 层**：无需手动设置 ID、CreatedAt、UpdatedAt
- **Controller 层**：ID 直接从 `ctx.Param("id")` 获取，无需转换

### 11.2 错误处理

- **Repository 层**：直接返回 GORM 错误
- **Service 层**：使用 `fmt.Errorf()` 包装错误信息
- **Controller 层**：将 Service 层错误转换为 HTTP 响应
- **日志记录**：使用 `LoggerMgr.Ins()` 记录错误信息

### 11.3 日志使用

- **使用 LoggerMgr**：依赖注入 `loggermgr.ILoggerManager`
- **结构化日志**：使用 `logger.Info("msg", "key", value)` 格式
- **日志级别**：Debug（开发调试）、Info（正常业务）、Warn（降级处理）、Error（业务错误）、Fatal（致命错误）
- **敏感信息**：密码、token 等必须脱敏

### 11.4 事务管理

- **Service 层管理事务**：使用 `Manager.DB().Transaction()`
- **Repository 层只提供方法**：不包含事务逻辑
- **错误回滚**：Transaction 函数返回错误会自动回滚

### 11.5 中间件顺序

- **推荐顺序**：Recovery → CORS → SecurityHeaders → RateLimiter → Telemetry → Auth → 其他
- **默认 Order**：Recovery(0) → RequestLogger(50) → CORS(100) → SecurityHeaders(150) → RateLimiter(200) → Telemetry(250)
- **自定义中间件**：建议从 Order 350 开始

### 11.6 代码生成

- **不要手动修改生成的代码**：所有生成的文件头部包含 `// Code generated by litecore/cli. DO NOT EDIT.` 注释
- **添加新组件后重新生成**：运行 `go run ./cmd/generate`
- **接口命名**：组件接口必须使用 `I` 前缀（如 `IMessageService`）
- **工厂函数**：工厂函数必须使用 `New` 前缀（如 `NewMessageService`）

---

## 12. 常见问题

### Q: 为什么要使用实体基类？

A: 实体基类提供以下优势：
- **自动生成 ID**：CUID2 25位字符串，时间有序、高唯一性、分布式安全
- **自动填充时间戳**：通过 GORM Hook 自动设置 CreatedAt、UpdatedAt
- **类型安全**：ID 类型为 string，Repository 和 Service 层无需类型转换
- **简化代码**：无需在 Service 层手动设置 ID 和时间戳

### Q: 如何切换数据库？

A: 修改 `configs/config.yaml` 中的 `database.driver` 和对应的配置：

```yaml
database:
  driver: "mysql"
  mysql_config:
    dsn: "root:password@tcp(localhost:3306)/myapp?charset=utf8mb4"
```

### Q: 如何添加新的中间件？

A: 在 `internal/middlewares/` 目录创建中间件文件：

```go
package middlewares

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
)

type ICustomMiddleware interface {
    common.IBaseMiddleware
}

type customMiddlewareImpl struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func NewCustomMiddleware() ICustomMiddleware {
    return &customMiddlewareImpl{}
}

func (m *customMiddlewareImpl) MiddlewareName() string {
    return "CustomMiddleware"
}

func (m *customMiddlewareImpl) Order() int {
    return 350
}

func (m *customMiddlewareImpl) Wrapper() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 中间件逻辑
        c.Next()
    }
}

var _ ICustomMiddleware = (*customMiddlewareImpl)(nil)
```

然后运行 `go run ./cmd/generate` 重新生成容器代码。

### Q: 如何添加新的监听器？

A: 在 `internal/listeners/` 目录创建监听器文件：

```go
package listeners

import (
    "context"
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
)

type ICustomListener interface {
    common.IBaseListener
}

type customListenerImpl struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func NewCustomListener() ICustomListener {
    return &customListenerImpl{}
}

func (l *customListenerImpl) ListenerName() string {
    return "CustomListener"
}

func (l *customListenerImpl) GetQueue() string {
    return "custom.queue"
}

func (l *customListenerImpl) Handle(ctx context.Context, msg common.IMessageListener) error {
    // 处理消息
    return nil
}

var _ ICustomListener = (*customListenerImpl)(nil)
```

然后运行 `go run ./cmd/generate` 重新生成容器代码。

### Q: 如何添加新的调度器？

A: 在 `internal/schedulers/` 目录创建调度器文件：

```go
package schedulers

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
)

type ICustomScheduler interface {
    common.IBaseScheduler
}

type customSchedulerImpl struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
}

func NewCustomScheduler() ICustomScheduler {
    return &customSchedulerImpl{}
}

func (s *customSchedulerImpl) SchedulerName() string {
    return "customScheduler"
}

func (s *customSchedulerImpl) GetRule() string {
    return "0 0 * * * *"  // 每小时执行
}

func (s *customSchedulerImpl) GetTimezone() string {
    return "Asia/Shanghai"
}

func (s *customSchedulerImpl) OnTick(tickID int64) error {
    // 执行定时任务
    return nil
}

var _ ICustomScheduler = (*customSchedulerImpl)(nil)
```

然后运行 `go run ./cmd/generate` 重新生成容器代码。

### Q: 生成的代码可以手动修改吗？

A: 不可以。所有生成的代码头部都包含 `// Code generated by litecore/cli. DO NOT EDIT.` 注释。如果需要修改，应该更新业务代码后重新生成。

### Q: 如何更新生成的代码？

A: 修改业务代码（添加/删除组件）后，重新运行代码生成器：

```bash
go run ./cmd/generate
```

### Q: 如何调试启动问题？

A: 启用启动日志和 Debug 日志级别：

```yaml
server:
  startup_log:
    enabled: true
    async: false

logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "debug"
      format: "gin"
```

### Q: 如何禁用某个 Manager？

A: 在 `configs/config.yaml` 中将对应 Manager 的 `driver` 设置为 `"none"`：

```yaml
database:
  driver: "none"

cache:
  driver: "none"

logger:
  driver: "none"
```

---

## 附录

### 示例项目

完整的示例项目请参考 `samples/messageboard`，展示了：
- 完整的 5 层架构实现
- 实体基类的使用
- 监听器和调度器的实现
- 中间件的使用
- 认证与会话管理
- 限流和安全保护
- HTML 模板和静态资源服务

### CLI 工具命令

```bash
# 查看帮助
litecore-cli --help

# 查看版本
litecore-cli version

# 生成容器代码
litecore-cli generate

# 创建新项目（交互式）
litecore-cli scaffold

# 创建新项目（命令行）
litecore-cli scaffold --module github.com/user/app --project myapp --template full

# 生成 Shell 补全
litecore-cli completion bash
```

### 参考资源

- [LiteCore CLI 文档](../cli/README.md)
- [Messageboard 示例](../samples/messageboard/README.md)
- [CLI 使用示例](../cli/EXAMPLES.md)

---

**最后更新：2026-01-25**
