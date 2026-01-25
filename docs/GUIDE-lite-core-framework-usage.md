# LiteCore 使用指南

 ## 目录

 - [1. 简介](#1-简介)
 - [2. 核心特性](#2-核心特性)
 - [3. 架构概述](#3-架构概述)
 - [4. 快速开始](#4-快速开始)
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

LiteCore 是一个基于 Go 的轻量级应用框架，旨在提供标准化、可扩展的微服务开发能力。框架采用 5 层分层架构，内置依赖注入容器、配置管理、数据库管理、缓存管理、日志管理、锁管理、限流管理、消息队列等功能，帮助开发者快速构建业务系统。

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

## 2. 核心特性

### 2.1 框架核心功能

| 功能 | 说明 | 实现方式 |
|------|------|----------|
| **5 层架构** | Entity → Repository → Service → Controller/Middleware | 接口定义 + 依赖注入 |
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

### 2.2 实用工具（util 包）

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

## 3. 架构概述

### 3.1 5 层架构图

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
└─────────────────────────┘    │  - CacheManager      │
                                │  - LoggerManager     │
                                │  - LockManager       │
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
```

### 3.2 依赖规则

```
Entity 层（无外部依赖）
    ↓
Repository 层（依赖 Entity、Config、Manager）
    ↓
Service 层（依赖 Repository、Config、Manager、Service）
    ↓
Controller 层（依赖 Service、Config、Manager）
Middleware 层（依赖 Service、Config、Manager）
    ↑ 依赖（由引擎自动注入）
Config 和 Manager（独立包，由引擎自动初始化和注入）
```

**规则说明**：
- 上层可以依赖下层
- 下层不能依赖上层
- 同层之间可以相互依赖（例如 Service 可以依赖另一个 Service）
- Controller 不能直接依赖 Repository，必须通过 Service
- Config 和 Manager 作为独立包，由引擎自动初始化和注入
- Manager 包位于 `manager/` 目录，包括：configmgr, databasemgr, cachemgr, loggermgr, lockmgr, limitermgr, mqmgr, telemetrymgr
- 内置组件位于 `component/` 目录，包括：litecontroller, litemiddleware, liteservice

### 3.3 生命周期管理

所有实现了生命周期接口的组件都会在以下时机被调用：

| 方法 | 调用时机 | 用途 |
|------|----------|------|
| `OnStart()` | 服务器启动时 | 初始化资源（连接数据库、加载缓存等） |
| `OnStop()` | 服务器停止时 | 清理资源（关闭连接、保存数据等） |
| `Health()` | 健康检查时 | 检查组件健康状态（内置 Manager 组件） |

---

## 4. 快速开始

### 4.1 引用私有仓库的 LiteCore

#### 方式一：配置 GOPRIVATE（推荐）

适用于生产环境和团队协作：

```bash
# 1. 设置私有模块前缀
export GOPRIVATE=github.com/lite-lake/litecore-go

# 2. 在新项目中引用指定版本
go mod init com.litelake.myapp
go get github.com/lite-lake/litecore-go@v0.0.1

# 3. 或使用最新版本
go get github.com/lite-lake/litecore-go@latest
```

#### 方式二：使用 replace 指令

适用于本地开发和调试：

```bash
# 1. 初始化项目
go mod init com.litelake.myapp

# 2. 在 go.mod 中添加 replace 指令
# replace github.com/lite-lake/litecore-go => /Users/kentzhu/Projects/lite-lake/litecore-go

# 3. 执行依赖整理
go mod tidy

# 4. 运行应用
go run ./cmd/server
```

### 4.2 初始化项目

```bash
# 创建项目目录
mkdir myapp && cd myapp

# 初始化 Go 模块
go mod init github.com/lite-lake/litecore-go/samples/myapp

# 引用 LiteCore
go get github.com/lite-lake/litecore-go@latest

# 创建项目结构
mkdir -p cmd/server cmd/generate configs data
mkdir -p internal/{entities,repositories,services,controllers,middlewares,dtos}

# 创建配置文件
touch configs/config.yaml
```

### 4.3 项目结构

```
myapp/
├── cmd/
│   ├── server/main.go          # 应用入口
│   └── generate/main.go         # 代码生成器
├── configs/config.yaml          # 配置文件
├── internal/
│   ├── application/             # 自动生成的容器（DO NOT EDIT）
│   │   ├── entity_container.go
│   │   ├── repository_container.go
│   │   ├── service_container.go
│   │   ├── controller_container.go
│   │   ├── middleware_container.go
│   │   └── engine.go
│   ├── entities/                # 实体层（无依赖）
│   ├── repositories/            # 仓储层（依赖 Manager）
│   ├── services/                # 服务层（依赖 Repository）
│   ├── controllers/             # 控制器层（依赖 Service）
│   ├── middlewares/             # 中间件层（依赖 Service）
│   └── dtos/                    # 数据传输对象
└── go.mod
```

**框架目录结构（LiteCore）**：

```
litecore-go/
├── manager/                    # 管理器组件（独立包）
│   ├── configmgr/              # 配置管理器
│   ├── databasemgr/            # 数据库管理器
│   ├── cachemgr/               # 缓存管理器
│   ├── loggermgr/              # 日志管理器
│   ├── lockmgr/                # 锁管理器
│   ├── limitermgr/             # 限流管理器
│   ├── mqmgr/                  # 消息队列管理器
│   └── telemetrymgr/           # 遥测管理器
├── component/                   # 内置组件
│   ├── litecontroller/         # 内置控制器
│   ├── litemiddleware/         # 内置中间件
│   └── liteservice/            # 内置服务
├── container/                   # 依赖注入容器
├── server/                      # 服务器引擎
├── logger/                      # 日志工具
├── util/                        # 实用工具
└── cli/                        # CLI 工具
```

### 4.4 创建配置文件（configs/config.yaml）

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
  sqlite_config:
    dsn: "./data/myapp.db"
    pool_config:
      max_open_conns: 1
      max_idle_conns: 1
  observability_config:
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

# 日志配置
logger:
  driver: "zap"                 # zap, default, none
  zap_config:
    telemetry_enabled: false    # 是否启用观测日志
    telemetry_config:
      level: "info"             # 日志级别
    console_enabled: true       # 是否启用控制台日志
    console_config:
      level: "info"             # 日志级别
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

# 遥测配置
telemetry:
  driver: "none"                # none, otel
```

### 4.5 创建应用入口（cmd/server/main.go）

```go
package main

import (
    "fmt"
    "os"
    app "github.com/lite-lake/litecore-go/samples/myapp/internal/application"
)

func main() {
    engine, err := app.NewEngine()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create engine: %v\n", err)
        os.Exit(1)
    }

    if err := engine.Initialize(); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to initialize engine: %v\n", err)
        os.Exit(1)
    }

    if err := engine.Start(); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to start engine: %v\n", err)
        os.Exit(1)
    }

    engine.WaitForShutdown()
}
```

### 4.6 配置代码生成器（cmd/generate/main.go）

```go
package main

import (
    "flag"
    "fmt"
    "os"

    "github.com/lite-lake/litecore-go/cli/generator"
)

func main() {
    cfg := generator.DefaultConfig()
    cfg.OutputDir = "internal/application"
    cfg.PackageName = "application"
    cfg.ConfigPath = "configs/config.yaml"

    if err := generator.Run(cfg); err != nil {
        fmt.Fprintf(os.Stderr, "错误: %v\n", err)
        os.Exit(1)
    }
}
```

### 4.7 初始化应用

```bash
# 首次生成容器代码
go run ./cmd/generate

# 运行应用
go run ./cmd/server/main.go
```

---

## 5. 5 层架构详解

### 5.1 Entity 层（实体层）

Entity 层定义数据实体，映射到数据库表结构。实体层无外部依赖，只包含纯数据定义。

#### 5.1.1 实体示例

```go
package entities

import (
    "fmt"
    "time"

    "github.com/lite-lake/litecore-go/common"
)

type User struct {
    ID        uint      `gorm:"primarykey" json:"id"`
    Name      string    `gorm:"type:varchar(50);not null" json:"name"`
    Email     string    `gorm:"type:varchar(100);uniqueIndex" json:"email"`
    Age       int       `gorm:"not null" json:"age"`
    Status    string    `gorm:"type:varchar(20);default:'active'" json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// EntityName 返回实体名称
func (u *User) EntityName() string {
    return "User"
}

// TableName 返回数据库表名
func (u *User) TableName() string {
    return "users"
}

// GetId 返回实体的唯一标识
func (u *User) GetId() string {
    return fmt.Sprintf("%d", u.ID)
}

// IsActive 检查用户是否激活
func (u *User) IsActive() bool {
    return u.Status == "active"
}

var _ common.IBaseEntity = (*User)(nil)
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

#### 5.1.4 实体设计规范

- **纯数据模型**：实体只包含数据，不包含业务逻辑
- **GORM 标签**：使用 GORM 标签定义表结构
- **接口实现**：必须实现 `common.IBaseEntity` 接口
- **辅助方法**：可以添加简单的辅助方法（如 `IsActive()`）
- **无依赖**：实体层不依赖任何其他层

---

### 5.2 Repository 层（仓储层）

Repository 层负责数据访问，提供 CRUD 操作和数据库交互。

#### 5.2.1 Repository 示例

```go
package repositories

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/databasemgr"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/entities"
)

type IUserRepository interface {
    common.IBaseRepository
    Create(user *entities.User) error
    GetByID(id uint) (*entities.User, error)
    GetByEmail(email string) (*entities.User, error)
    Update(user *entities.User) error
    Delete(id uint) error
    List(offset, limit int) ([]*entities.User, int64, error)
}

type userRepository struct {
    Manager databasemgr.IDatabaseManager `inject:""`
}

func NewUserRepository() IUserRepository {
    return &userRepository{}
}

func (r *userRepository) RepositoryName() string {
    return "UserRepository"
}

func (r *userRepository) OnStart() error {
    // 自动迁移表结构
    return r.Manager.AutoMigrate(&entities.User{})
}

func (r *userRepository) OnStop() error {
    return nil
}

func (r *userRepository) Create(user *entities.User) error {
    return r.Manager.DB().Create(user).Error
}

func (r *userRepository) GetByID(id uint) (*entities.User, error) {
    var user entities.User
    err := r.Manager.DB().First(&user, id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*entities.User, error) {
    var user entities.User
    err := r.Manager.DB().Where("email = ?", email).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) Update(user *entities.User) error {
    return r.Manager.DB().Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
    return r.Manager.DB().Delete(&entities.User{}, id).Error
}

func (r *userRepository) List(offset, limit int) ([]*entities.User, int64, error) {
    var users []*entities.User
    var total int64

    db := r.Manager.DB().Model(&entities.User{})
    if err := db.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    if err := db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
        return nil, 0, err
    }

    return users, total, nil
}

var _ IUserRepository = (*userRepository)(nil)
```

#### 5.2.2 Repository 设计规范

- **接口定义**：定义接口 `IXxxRepository`
- **依赖注入**：使用 `inject:""` 标签注入依赖
- **接口实现**：结构体命名为小写 `xxxRepository`
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
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/cachemgr"
    "github.com/lite-lake/litecore-go/manager/configmgr"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/entities"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/repositories"
)

type IUserService interface {
    common.IBaseService
    Register(name, email string, age int) (*entities.User, error)
    GetByID(id uint) (*entities.User, error)
    UpdateProfile(id uint, name string) error
    DeleteUser(id uint) error
    ListUsers(page, pageSize int) ([]*entities.User, int64, error)
}

type userService struct {
    Config     configmgr.IConfigManager     `inject:""`
    CacheMgr   cachemgr.ICacheManager      `inject:""`
    LoggerMgr  loggermgr.ILoggerManager    `inject:""`
    Repository repositories.IUserRepository `inject:""`
    logger     loggermgr.ILogger
}

func NewUserService() IUserService {
    return &userService{}
}

func (s *userService) ServiceName() string {
    return "UserService"
}

func (s *userService) OnStart() error {
    return nil
}

func (s *userService) OnStop() error {
    return nil
}

func (s *userService) Register(name, email string, age int) (*entities.User, error) {
    s.initLogger()
    // 验证输入
    if len(name) < 2 || len(name) > 50 {
        return nil, errors.New("用户名长度必须在 2-50 个字符之间")
    }
    if age < 0 || age > 150 {
        return nil, errors.New("年龄必须在 0-150 之间")
    }

    // 检查邮箱是否已存在
    existing, err := s.Repository.GetByEmail(email)
    if err == nil && existing != nil {
        return nil, errors.New("邮箱已被注册")
    }

    // 创建用户
    user := &entities.User{
        Name:   name,
        Email:  email,
        Age:    age,
        Status: "active",
    }

    if err := s.Repository.Create(user); err != nil {
        s.logger.Error("创建用户失败", "error", err, "email", email)
        return nil, fmt.Errorf("创建用户失败: %w", err)
    }

    s.logger.Info("用户注册成功", "user_id", user.ID, "email", email)
    return user, nil
}

func (s *userService) GetByID(id uint) (*entities.User, error) {
    s.initLogger()

    // 尝试从缓存获取
    cacheKey := fmt.Sprintf("user:%d", id)
    var user entities.User
    if err := s.CacheMgr.Get(context.Background(), cacheKey, &user); err == nil {
        return &user, nil
    }

    // 从数据库查询
    user, err := s.Repository.GetByID(id)
    if err != nil {
        s.logger.Error("获取用户失败", "error", err, "user_id", id)
        return nil, fmt.Errorf("获取用户失败: %w", err)
    }
    if user == nil {
        return nil, errors.New("用户不存在")
    }

    // 写入缓存
    s.CacheMgr.Set(context.Background(), cacheKey, user, time.Hour)

    return user, nil
}

func (s *userService) UpdateProfile(id uint, name string) error {
    s.initLogger()
    // 验证输入
    if len(name) < 2 || len(name) > 50 {
        return errors.New("用户名长度必须在 2-50 个字符之间")
    }

    // 获取用户
    user, err := s.Repository.GetByID(id)
    if err != nil {
        s.logger.Error("获取用户失败", "error", err, "user_id", id)
        return fmt.Errorf("获取用户失败: %w", err)
    }
    if user == nil {
        return errors.New("用户不存在")
    }

    // 更新用户
    user.Name = name
    if err := s.Repository.Update(user); err != nil {
        s.logger.Error("更新用户失败", "error", err, "user_id", id)
        return fmt.Errorf("更新用户失败: %w", err)
    }

    // 清除缓存
    cacheKey := fmt.Sprintf("user:%d", id)
    s.CacheMgr.Delete(context.Background(), cacheKey)

    s.logger.Info("用户信息更新成功", "user_id", id)
    return nil
}

func (s *userService) DeleteUser(id uint) error {
    s.initLogger()
    // 检查用户是否存在
    user, err := s.Repository.GetByID(id)
    if err != nil {
        s.logger.Error("获取用户失败", "error", err, "user_id", id)
        return fmt.Errorf("获取用户失败: %w", err)
    }
    if user == nil {
        return errors.New("用户不存在")
    }

    // 删除用户
    if err := s.Repository.Delete(id); err != nil {
        s.logger.Error("删除用户失败", "error", err, "user_id", id)
        return fmt.Errorf("删除用户失败: %w", err)
    }

    // 清除缓存
    cacheKey := fmt.Sprintf("user:%d", id)
    s.CacheMgr.Delete(context.Background(), cacheKey)

    s.logger.Info("用户删除成功", "user_id", id)
    return nil
}

func (s *userService) ListUsers(page, pageSize int) ([]*entities.User, int64, error) {
    s.initLogger()
    if page < 1 {
        page = 1
    }
    if pageSize < 1 || pageSize > 100 {
        pageSize = 20
    }

    offset := (page - 1) * pageSize
    users, total, err := s.Repository.List(offset, pageSize)
    if err != nil {
        s.logger.Error("获取用户列表失败", "error", err)
        return nil, 0, fmt.Errorf("获取用户列表失败: %w", err)
    }

    return users, total, nil
}

var _ IUserService = (*userService)(nil)
```

#### 5.3.2 Service 设计规范

- **业务逻辑**：在 Service 层实现所有业务逻辑
- **数据验证**：在 Service 层进行输入验证
- **错误包装**：使用 `fmt.Errorf()` 包装错误信息
- **事务管理**：在 Service 层管理数据库事务
- **依赖注入**：可以依赖 Repository、Manager、其他 Service

---

### 5.4 Controller 层（控制器层）

Controller 层负责 HTTP 请求处理，包括参数验证、调用 Service、响应封装。

#### 5.4.1 Controller 示例

```go
package controllers

import (
    "net/http"

    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/configmgr"
    "github.com/lite-lake/litecore-go/manager/loggermgr"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/dtos"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/services"

    "github.com/gin-gonic/gin"
)

type IUserController interface {
    common.IBaseController
}

type userController struct {
    Config      configmgr.IConfigManager  `inject:""`
    LoggerMgr   loggermgr.ILoggerManager  `inject:""`
    UserService services.IUserService      `inject:""`
    logger      loggermgr.ILogger
}

func NewUserController() IUserController {
    return &userController{}
}

func (c *userController) ControllerName() string {
    return "userController"
}

// RegisterUser 注册用户
// @Router /api/users/register [POST]
func (c *userController) GetRouter() string {
    return "/api/users/register [POST]"
}

func (c *userController) Handle(ctx *gin.Context) {

    var req dtos.RegisterUserRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        c.logger.Warn("参数验证失败", "error", err)
        ctx.JSON(http.StatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
        return
    }

    user, err := c.UserService.Register(req.Name, req.Email, req.Age)
    if err != nil {
        c.logger.Warn("注册用户失败", "error", err, "email", req.Email)
        ctx.JSON(http.StatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
        return
    }

    c.logger.Info("注册用户成功", "user_id", user.ID)
    ctx.JSON(http.StatusOK, dtos.SuccessResponse("注册成功", dtos.UserResponse{
        ID:    user.ID,
        Name:  user.Name,
        Email: user.Email,
    }))
}

var _ IUserController = (*userController)(nil)
```

#### 5.4.2 DTO 示例

```go
package dtos

import "time"

// RegisterUserRequest 注册用户请求
type RegisterUserRequest struct {
    Name  string `json:"name" binding:"required,min=2,max=50"`
    Email string `json:"email" binding:"required,email"`
    Age   int    `json:"age" binding:"required,min=0,max=150"`
}

// UserResponse 用户响应
type UserResponse struct {
    ID        uint      `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Age       int       `json:"age"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
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

#### 5.4.4 路由定义格式

Controller 的 `GetRouter()` 方法支持完整的路由语法：

```go
// 基本 CRUD
return "/api/users [GET]"              // 获取列表
return "/api/users [POST]"             // 创建
return "/api/users/:id [GET]"          // 获取详情
return "/api/users/:id [PUT]"          // 更新
return "/api/users/:id [DELETE]"       // 删除

// 路径参数
return "/api/files/*filepath [GET]"    // 通配符

// 路由分组
return "/api/admin/users [GET]"        // 管理端路由
return "/api/v1/users [GET]"           // 版本化路由
```

---

### 5.5 Middleware 层（中间件层）

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
    logger      loggermgr.ILogger
}

func NewAuthMiddleware() IAuthMiddleware {
    return &authMiddleware{}
}

func (m *authMiddleware) MiddlewareName() string {
    return "AuthMiddleware"
}

func (m *authMiddleware) Order() int {
    return 100
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
            m.logger.Warn("未提供认证令牌", "path", c.Request.URL.Path)
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
            m.logger.Warn("认证令牌格式错误", "path", c.Request.URL.Path)
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
            m.logger.Warn("认证令牌无效", "path", c.Request.URL.Path, "error", err)
            c.JSON(common.HTTPStatusUnauthorized, gin.H{
                "code":    common.HTTPStatusUnauthorized,
                "message": "认证令牌无效或已过期",
            })
            c.Abort()
            return
        }

        // 将用户信息存入上下文
        c.Set("user_session", session)
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
func (m *AuthMiddleware) Order() int              { return 100 }  // 认证中间件
func (m *LoggerMiddleware) Order() int           { return 200 }  // 日志中间件
func (m *TelemetryMiddleware) Order() int        { return 300 }  // 遥测中间件
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
        "timestamp": time.Now().Unix(),
    }

    data, err := json.Marshal(notification)
    if err != nil {
        return err
    }

    return s.MQMgr.Publish(ctx, "notifications", data)
}
```

**订阅消息**

```go
// 在 Service 的 OnStart 方法中启动消费者
func (s *NotificationService) OnStart() error {
    ctx := context.Background()
    return s.MQMgr.SubscribeWithCallback(ctx, "notifications", s.handleNotification)
}

func (s *NotificationService) handleNotification(ctx context.Context, msg mqmgr.Message) error {
    var notification map[string]interface{}
    if err := json.Unmarshal(msg.Body(), &notification); err != nil {
        return err
    }

    // 处理通知
    fmt.Printf("Received notification: %+v\n", notification)

    // 确认消息已处理
    return s.MQMgr.Ack(ctx, msg)
}
```

#### 6.5.3 使用场景

- **异步任务**：邮件发送、短信发送等耗时操作
- **事件驱动**：用户注册、订单创建等事件处理
- **系统解耦**：模块间通过消息队列解耦
- **流量削峰**：高并发场景下的缓冲处理

### 6.6 可用的内置 Manager

| Manager | 接口 | 功能 | 驱动 |
|---------|------|------|------|
| ConfigManager | IConfigManager | 配置管理 | YAML/JSON |
| LoggerManager | ILoggerManager | 日志管理 | Zap/Default |
| DatabaseManager | IDatabaseManager | 数据库管理 | MySQL/PostgreSQL/SQLite |
| CacheManager | ICacheManager | 缓存管理 | Redis/Memory |
| LockManager | ILockManager | 锁管理 | Redis/Memory |
| LimiterManager | ILimiterManager | 限流管理 | Redis/Memory |
| MQManager | IMQManager | 消息队列管理 | RabbitMQ/Memory |
| TelemetryManager | ITelemetryManager | 遥测管理 | OpenTelemetry |

### 6.7 使用内置组件

所有内置 Manager 都通过依赖注入自动注入到你的组件中，无需手动创建：

```go
type userService struct {
    Config    configmgr.IConfigManager    `inject:""`
    Logger    loggermgr.ILoggerManager   `inject:""`
    Database  databasemgr.IDatabaseManager `inject:""`
    Cache     cachemgr.ICacheManager      `inject:""`
    Lock      lockmgr.ILockManager       `inject:""`
    Limiter   limitermgr.ILimiterManager `inject:""`
    MQ        mqmgr.IMQManager           `inject:""`
    Telemetry telemetrymgr.ITelemetryManager `inject:""`
}
```

### 6.8 日志配置（Gin 格式）

LiteCore 支持 Gin 风格、JSON 和 Default 三种日志格式。

#### 日志格式类型

| 格式 | 说明 | 适用场景 |
|------|------|----------|
| `gin` | Gin 风格，竖线分隔符，彩色输出 | 控制台输出，开发环境 |
| `json` | JSON 格式，结构化日志 | 日志收集系统，生产环境 |
| `default` | Zap 默认 ConsoleEncoder 格式 | 默认日志格式 |

#### Gin 格式特点

- 统一格式：`{时间} | {级别} | {消息} | {字段1}={值1} {字段2}={值2} ...`
- 时间固定宽度 23 字符：`2006-01-02 15:04:05.000`
- 级别固定宽度 5 字符，右对齐，带颜色
- 字段格式：`key=value`，字符串值用引号包裹

#### 配置示例

```yaml
logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "info"               # 日志级别：debug, info, warn, error, fatal
      format: "gin"                # 格式：gin | json | default
      color: true                  # 是否启用颜色
      time_format: "2006-01-02 15:04:05.000"
```

#### 颜色配置

- **color**: 控制是否启用彩色输出（默认 true，由终端自动检测）
- 支持在配置文件中手动关闭颜色：`color: false`

#### 日志级别颜色

| 级别 | ANSI 颜色 | 说明 |
|------|-----------|------|
| DEBUG | 灰色 | 开发调试信息 |
| INFO | 绿色 | 正常业务流程 |
| WARN | 黄色 | 降级处理、慢查询 |
| ERROR | 红色 | 业务错误、操作失败 |
| FATAL | 红色+粗体 | 致命错误 |

#### HTTP 状态码颜色

| 状态码范围 | 颜色 | 说明 |
|-----------|------|------|
| 2xx | 绿色 | 成功 |
| 3xx | 黄色 | 重定向 |
| 4xx | 橙色 | 客户端错误 |
| 5xx | 红色 | 服务器错误 |

#### 日志格式示例

```
# Gin 格式输出示例
2026-01-24 15:04:05.123 | INFO  | 开始依赖注入 | count=23
2026-01-24 15:04:05.456 | WARN  | 慢查询检测 | duration=1.2s
2026-01-24 15:04:05.789 | ERROR | 数据库连接失败 | error="connection refused"

# 请求日志（Gin 格式）
2026-01-24 15:04:05.123 | 200   | 1.234ms | 127.0.0.1 | GET | /api/messages
2026-01-24 15:04:05.456 | 404   | 0.456ms | 192.168.1.1 | POST | /api/unknown
```

### 6.9 启动日志

框架提供详细的启动日志，记录应用启动过程中的各个阶段和耗时。

#### 启动日志配置

```yaml
server:
  startup_log:
    enabled: true      # 是否启用启动日志
    async: true        # 是否异步日志（提高性能）
    buffer: 100        # 日志缓冲区大小
```

#### 启动阶段

框架将启动过程分为以下几个阶段：

| 阶段 | 说明 |
|------|------|
| PhaseConfig | 配置加载阶段 |
| PhaseManagers | 管理器初始化阶段 |
| PhaseInjection | 依赖注入阶段 |
| PhaseRouter | 路由注册阶段 |
| PhaseStartup | 服务启动阶段 |

#### 启动日志示例

```
2026-01-24 15:04:05.001 | INFO  | 配置文件: configs/config.yaml
2026-01-24 15:04:05.002 | INFO  | 配置驱动: yaml
2026-01-24 15:04:05.003 | INFO  | 初始化完成: ConfigManager
2026-01-24 15:04:05.010 | INFO  | 初始化完成: TelemetryManager
2026-01-24 15:04:05.015 | INFO  | 初始化完成: LoggerManager
2026-01-24 15:04:05.020 | INFO  | 初始化完成: DatabaseManager
2026-01-24 15:04:05.025 | INFO  | 初始化完成: CacheManager
2026-01-24 15:04:05.030 | INFO  | 初始化完成: LockManager
2026-01-24 15:04:05.035 | INFO  | 初始化完成: LimiterManager
2026-01-24 15:04:05.040 | INFO  | 初始化完成: MQManager
2026-01-24 15:04:05.041 | INFO  | 管理器初始化完成 | count=8
2026-01-24 15:04:05.042 | INFO  | 开始依赖注入
2026-01-24 15:04:05.043 | INFO  | [Repository 层] MessageRepository: 注入完成
2026-01-24 15:04:05.044 | INFO  | [Service 层] MessageService: 注入完成
2026-01-24 15:04:05.045 | INFO  | [Controller 层] MsgCreateController: 注入完成
2026-01-24 15:04:05.046 | INFO  | [Middleware 层] AuthMiddleware: 注入完成
2026-01-24 15:04:05.047 | INFO  | 依赖注入完成 | count=4
2026-01-24 15:04:05.048 | INFO  | 开始注册路由
2026-01-24 15:04:05.049 | INFO  | 注册路由 | method=POST path=/api/messages controller=msgCreateControllerImpl
2026-01-24 15:04:05.050 | INFO  | 路由注册完成 | route_count=1 controller_count=1
2026-01-24 15:04:05.051 | INFO  | HTTP server listening | addr=0.0.0.0:8080
2026-01-24 15:04:05.052 | INFO  | 服务启动完成，开始对外提供服务 | total_duration=51ms
```

---

## 7. 代码生成器使用

### 7.1 代码生成器概述

LiteCore 提供了自动代码生成器，可以自动生成依赖注入容器代码和引擎代码，减少重复劳动。

### 7.2 代码生成器配置

```go
package main

import (
    "github.com/lite-lake/litecore-go/cli/generator"
)

func main() {
    cfg := generator.DefaultConfig()
    cfg.OutputDir = "internal/application"
    cfg.PackageName = "application"
    cfg.ConfigPath = "configs/config.yaml"

    if err := generator.Run(cfg); err != nil {
        panic(err)
    }
}
```

### 7.3 代码生成器选项

| 选项 | 默认值 | 说明 |
|------|--------|------|
| ProjectPath | "." | 项目路径 |
| OutputDir | "internal/application" | 输出目录 |
| PackageName | "application" | 包名 |
| ConfigPath | "configs/config.yaml" | 配置文件路径 |

### 7.4 生成的文件

代码生成器会自动生成以下文件：

| 文件 | 说明 |
|------|------|
| entity_container.go | 实体容器初始化代码 |
| repository_container.go | 仓储容器初始化代码 |
| service_container.go | 服务容器初始化代码 |
| controller_container.go | 控制器容器初始化代码 |
| middleware_container.go | 中间件容器初始化代码 |
| engine.go | 引擎初始化代码 |

**注意**：所有生成的文件头部都有 `// Code generated by litecore/cli. DO NOT EDIT.` 注释，不要手动修改这些文件。

### 7.5 代码生成流程

1. **扫描代码**：扫描 `internal/entities`、`internal/repositories`、`internal/services`、`internal/controllers`、`internal/middlewares` 目录
2. **解析组件**：解析实现了相应接口的组件
3. **生成容器**：生成各个容器的初始化代码
4. **生成引擎**：生成引擎初始化代码

### 7.6 自定义代码生成

你可以通过修改代码生成器的模板来定制生成的代码风格。模板文件位于 `cli/generator/template.go`。

---

## 8. 依赖注入机制

### 8.1 依赖注入概述

LiteCore 使用标签驱动的依赖注入机制，通过 `inject:""` 标签自动注入依赖。

### 8.2 依赖注入标签

```go
type userService struct {
    Config    configmgr.IConfigManager    `inject:""`
    Logger    loggermgr.ILoggerManager   `inject:""`
    Repository repositories.IUserRepository `inject:""`
}
```

### 8.3 依赖注入顺序

依赖注入按以下顺序进行：

1. **Manager 注入**：注入内置管理器
2. **Repository 注入**：注入仓储依赖
3. **Service 注入**：注入服务依赖
4. **Controller 注入**：注入控制器依赖
5. **Middleware 注入**：注入中间件依赖

### 8.4 依赖注入规则

- **标签识别**：只有带有 `inject:""` 标签的字段才会被注入
- **类型匹配**：依赖项的类型必须与容器中注册的类型匹配
- **单向依赖**：下层不能依赖上层，同层可以相互依赖
- **自动注入**：框架会自动识别并注入所有依赖项

### 8.5 依赖注入示例

```go
// Service 依赖 Repository
type userService struct {
    Repository repositories.IUserRepository `inject:""`
}

// Controller 依赖 Service
type userController struct {
    UserService services.IUserService `inject:""`
}

// Middleware 依赖 Service
type authMiddleware struct {
    AuthService services.IAuthService `inject:""`
}

// Service 依赖 Manager
type cacheService struct {
    CacheMgr cachemgr.ICacheManager `inject:""`
}
```

### 8.6 依赖注入错误处理

如果依赖注入失败，框架会返回详细的错误信息，包括：

- **依赖未注册**：依赖的类型未在容器中注册
- **循环依赖**：检测到循环依赖关系
- **类型不匹配**：依赖的类型与注册的类型不匹配

---

## 9. 配置管理

### 9.1 配置文件格式

LiteCore 支持 YAML 和 JSON 两种配置文件格式。

### 9.2 配置加载

框架会在引擎初始化时自动加载配置文件，无需手动加载。

### 9.3 配置访问

通过依赖注入的 `ConfigManager` 访问配置：

```go
type userService struct {
    Config configmgr.IConfigManager `inject:""`
}

func (s *userService) GetConfig() {
    // 获取配置值
    appName := s.Config.GetString("app.name")
    appVersion := s.Config.GetString("app.version")
}
```

### 9.4 配置结构

```yaml
app:
  name: "myapp"
  version: "1.0.0"

server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"
  read_timeout: "10s"
  write_timeout: "10s"
  idle_timeout: "60s"
  enable_recovery: true
  shutdown_timeout: "30s"
  startup_log:
    enabled: true
    async: true
    buffer: 100

database:
  driver: "sqlite"
  sqlite_config:
    dsn: "./data/myapp.db"
    pool_config:
      max_open_conns: 1
      max_idle_conns: 1
  observability_config:
    slow_query_threshold: "1s"
    log_sql: false

cache:
  driver: "memory"
  memory_config:
    max_size: 100
    max_age: "720h"
    max_backups: 1000
    compress: false

lock:
  driver: "memory"
  memory_config:
    max_backups: 1000

limiter:
  driver: "memory"
  memory_config:
    max_backups: 1000

mq:
  driver: "memory"
  memory_config:
    max_queue_size: 10000
    channel_buffer: 100

logger:
  driver: "zap"
  zap_config:
    telemetry_enabled: false
    telemetry_config:
      level: "info"
    console_enabled: true
    console_config:
      level: "info"
      format: "gin"
      color: true
      time_format: "2006-01-02 15:04:05.000"
    file_enabled: false
    file_config:
      level: "info"
      path: "./logs/myapp.log"
      rotation:
        max_size: 100
        max_age: 30
        max_backups: 10
        compress: true

telemetry:
  driver: "none"
```

---

## 10. 实用工具（util 包）

### 10.1 JWT 工具

```go
import "github.com/lite-lake/litecore-go/util/jwt"

// 生成 JWT token
token, err := jwt.GenerateHS256Token(claims, "your-secret-key")

// 解析 JWT token
claims, err := jwt.ParseHS256Token(token, "your-secret-key")

// 验证 JWT token
valid, err := jwt.ValidateHS256Token(token, "your-secret-key")
```

### 10.2 哈希工具

```go
import "github.com/lite-lake/litecore-go/util/hash"

// MD5 哈希
hash := hash.MD5("your-data")

// SHA1 哈希
hash := hash.SHA1("your-data")

// SHA256 哈希
hash := hash.SHA256("your-data")
```

### 10.3 加密工具

```go
import "github.com/lite-lake/litecore-go/util/crypt"

// 密码加密
hashedPassword, err := crypt.HashPassword("your-password")

// 验证密码
valid, err := crypt.VerifyPassword("your-password", hashedPassword)

// AES 加密
encrypted, err := crypt.AESGCMEncrypt(key, plaintext)

// AES 解密
decrypted, err := crypt.AESGCMDecrypt(key, encrypted)
```

### 10.4 ID 生成工具

```go
import "github.com/lite-lake/litecore-go/util/id"

// 生成雪花算法 ID
snowflakeID := id.NewSnowflakeID()

// 生成 UUID
uuid := id.NewUUID()
```

### 10.5 字符串工具

```go
import "github.com/lite-lake/litecore-go/util/stringutil"

// 生成随机字符串
randomStr := stringutil.RandomString(16)

// 驼峰转下划线
snakeStr := stringutil.CamelToSnake("MyString")

// 下划线转驼峰
camelStr := stringutil.SnakeToCamel("my_string")
```

---

## 11. 最佳实践

### 11.1 项目组织

- **分层清晰**：严格按照 5 层架构组织代码
- **职责明确**：每层只负责自己的职责，不越界
- **依赖单向**：上层依赖下层，下层不依赖上层

### 11.2 错误处理

- **错误包装**：使用 `fmt.Errorf()` 包装错误，保留原始错误信息
- **日志记录**：关键错误必须记录日志
- **用户友好**：向用户返回友好的错误信息

```go
func (s *userService) CreateUser(name, email string) error {
    if err := s.Repository.Create(user); err != nil {
        s.logger.Error("创建用户失败", "error", err, "email", email)
        return fmt.Errorf("创建用户失败: %w", err)
    }
    return nil
}
```

### 11.3 日志记录

- **结构化日志**：使用结构化日志，便于查询和分析
- **日志级别**：合理使用不同的日志级别
- **敏感信息**：不要记录敏感信息（密码、token等）

```go
s.logger.Info("用户注册成功", "user_id", user.ID, "email", email)
s.logger.Warn("登录失败", "email", email, "reason", "invalid_password")
s.logger.Error("数据库连接失败", "error", err)
```

### 11.4 性能优化

- **使用缓存**：热点数据使用缓存
- **批量操作**：尽量使用批量操作减少数据库访问
- **异步处理**：耗时操作使用异步处理（消息队列）

### 11.5 安全实践

- **输入验证**：所有用户输入都必须验证
- **SQL 注入**：使用参数化查询，防止 SQL 注入
- **XSS 防护**：对用户输出进行转义，防止 XSS 攻击
- **认证授权**：严格实施认证和授权机制

### 11.6 测试实践

- **单元测试**：为每个函数编写单元测试
- **集成测试**：编写集成测试验证组件协作
- **覆盖率**：保持测试覆盖率在合理水平

### 11.7 代码风格

- **遵循规范**：遵循 Go 语言官方代码规范
- **命名清晰**：使用有意义的命名
- **注释完整**：导出的函数和类型必须添加注释

---

## 12. 常见问题

### Q1: 如何切换数据库？

修改配置文件中的 `database.driver` 和对应的配置：

```yaml
database:
  driver: "mysql"  # 从 sqlite 切换到 mysql
  mysql_config:
    dsn: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
```

### Q2: 如何启用 Redis？

修改配置文件中的 `cache.driver` 和 `lock.driver`：

```yaml
cache:
  driver: "redis"  # 从 memory 切换到 redis
  redis_config:
    host: "localhost"
    port: 6379
```

### Q3: 如何自定义日志格式？

修改配置文件中的 `logger.zap_config.console_config.format`：

```yaml
logger:
  zap_config:
    console_config:
      format: "json"  # 从 gin 切换到 json
```

### Q4: 如何添加新的 Manager？

1. 在 `manager/` 目录下创建新的包
2. 实现 `common.IBaseManager` 接口
3. 实现 `BuildWithConfigProvider()` 函数
4. 在 `server/builtin.go` 中注册

### Q5: 如何调试依赖注入问题？

框架会在启动时输出详细的依赖注入日志，包括：

- 每个组件的注入状态
- 依赖关系的解析过程
- 错误信息的详细描述

---

## 附录

### A. HTTP 状态码

| 状态码 | 名称 | 常量 |
|--------|------|------|
| 200 | OK | `HTTPStatusOK` |
| 201 | Created | `HTTPStatusCreated` |
| 204 | No Content | `HTTPStatusNoContent` |
| 400 | Bad Request | `HTTPStatusBadRequest` |
| 401 | Unauthorized | `HTTPStatusUnauthorized` |
| 403 | Forbidden | `HTTPStatusForbidden` |
| 404 | Not Found | `HTTPStatusNotFound` |
| 429 | Too Many Requests | `HTTPStatusTooManyRequests` |
| 500 | Internal Server Error | `HTTPStatusInternalServerError` |
| 502 | Bad Gateway | `HTTPStatusBadGateway` |
| 503 | Service Unavailable | `HTTPStatusServiceUnavailable` |

### B. 日志级别

| 级别 | 值 | 说明 |
|------|----|----|
| Debug | 0 | 开发调试信息 |
| Info | 1 | 正常业务流程 |
| Warn | 2 | 降级处理、慢查询 |
| Error | 3 | 业务错误、操作失败 |
| Fatal | 4 | 致命错误 |

### C. 配置驱动类型

| 驱动 | 说明 |
|------|------|
| `yaml` | YAML 格式配置文件 |
| `json` | JSON 格式配置文件 |

### D. 数据库驱动类型

| 驱动 | 说明 |
|------|------|
| `mysql` | MySQL 数据库 |
| `postgresql` | PostgreSQL 数据库 |
| `sqlite` | SQLite 数据库 |
| `none` | 不使用数据库 |

### E. 缓存驱动类型

| 驱动 | 说明 |
|------|------|
| `redis` | Redis 缓存 |
| `memory` | 内存缓存（基于 Ristretto） |
| `none` | 不使用缓存 |

### F. 锁驱动类型

| 驱动 | 说明 |
|------|------|
| `redis` | Redis 分布式锁 |
| `memory` | 内存锁 |
| `none` | 不使用锁 |

### G. 限流驱动类型

| 驱动 | 说明 |
|------|------|
| `redis` | Redis 限流 |
| `memory` | 内存限流 |
| `none` | 不使用限流 |

### H. 消息队列驱动类型

| 驱动 | 说明 |
|------|------|
| `rabbitmq` | RabbitMQ 消息队列 |
| `memory` | 内存消息队列 |
| `none` | 不使用消息队列 |

### I. 日志驱动类型

| 驱动 | 说明 |
|------|------|
| `zap` | Zap 高性能日志库 |
| `default` | Go 标准库日志 |
| `none` | 不使用日志 |

### J. 遥测驱动类型

| 驱动 | 说明 |
|------|------|
| `otel` | OpenTelemetry |
| `none` | 不使用遥测 |

---

**文档版本**：v1.0.0
**最后更新**：2026-01-24
