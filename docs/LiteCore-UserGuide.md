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
  - [5.4 Controller 层（控制器层）](#54-controller-层控制器层)
  - [5.5 Middleware 层（中间件层）](#55-middleware-层中间件层)
- [6. 内置组件](#6-内置组件)
- [7. 代码生成器使用](#7-代码生成器使用)
- [8. 依赖注入机制](#8-依赖注入机制)
- [9. 配置管理](#9-配置管理)
- [10. 实用工具（util 包）](#10-实用工具util-包)
- [11. 最佳实践](#11-最佳实践)
- [12. 常见问题](#12-常见问题)
- [附录](#附录)

---

## 1. 简介

LiteCore 是一个基于 Go 的轻量级企业级应用框架，旨在提供标准化、可扩展的微服务开发能力。框架采用 5 层分层架构，内置依赖注入容器、配置管理、数据库管理、缓存管理、日志管理等功能，帮助开发者快速构建业务系统。

### 为什么要使用 LiteCore？

- **标准化架构**：统一的 5 层架构规范，降低团队协作成本
- **内置组件**：Config 和 Manager 作为服务器内置组件，自动初始化和注入
- **依赖注入**：自动化的依赖注入容器，简化组件管理
- **开箱即用**：内置数据库、缓存、日志等基础组件
- **代码生成**：自动生成容器代码，减少重复劳动
- **可观测性**：内置日志、指标、链路追踪支持
- **配置驱动**：通过配置文件管理应用行为，无需修改代码

### 适用场景

- Web 应用和 API 服务
- 微服务架构
- 企业级业务系统
- 需要快速原型开发的项目

---

## 2. 核心特性

### 2.1 框架核心功能

| 功能 | 说明 | 实现方式 |
|------|------|----------|
| **5 层架构** | Entity → Repository → Service → Controller/Middleware | 接口定义 + 依赖注入 |
| **内置组件** | Config 和 Manager 自动初始化和注入 | server/builtin 包 |
| **依赖注入** | 自动扫描、自动注入、生命周期管理 | reflect + inject 标签 |
| **代码生成** | 自动生成容器代码和引擎代码 | CLI 工具 |
| **配置管理** | 支持 YAML/JSON 配置文件 | config 包 |
| **数据库管理** | 支持 MySQL/PostgreSQL/SQLite | GORM + Manager 封装 |
| **缓存管理** | 支持 Redis/Memory 缓存 | cache 包 |
| **日志管理** | 基于 Zap 的高性能日志 | logger 包 |
| **遥测支持** | OpenTelemetry 集成 | telemetry 包 |

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
│  Entity    (实体层)     │    │  Manager   (内置组件) │
│  - 数据模型定义          │    │  - DatabaseManager   │
│  - 表结构定义            │    │  - CacheManager      │
└─────────────────────────┘    │  - LoggerManager     │
                              │  - TelemetryManager  │
                              └──────────────────────┘
                                        ↑ 依赖
                              ┌──────────────────────┐
                              │  Config    (内置配置) │
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
Config 和 Manager（内置组件）
```

**规则说明**：
- 上层可以依赖下层
- 下层不能依赖上层
- 同层之间可以相互依赖（例如 Service 可以依赖另一个 Service）
- Controller 不能直接依赖 Repository，必须通过 Service
- Config 和 Manager 作为内置组件，由引擎自动初始化和注入

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
mkdir -p internal/{application,entities,repositories,services,controllers,middlewares,dtos,infras/{configproviders,managers}}

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
│   ├── dtos/                    # 数据传输对象
│   └── infras/                  # 基础设施（Manager 封装）
│       └── managers/            # 管理器封装
│           ├── database_manager.go
│           ├── cache_manager.go
│           └── logger_manager.go
└── go.mod
```

### 4.4 创建配置文件（configs/config.yaml）

```yaml
app:
  name: "myapp"
  version: "1.0.0"

server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"

database:
  driver: "sqlite"              # mysql, postgresql, sqlite, none
  sqlite_config:
    dsn: "./data/myapp.db"

cache:
  driver: "memory"              # redis, memory, none

logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "info"
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
    "github.com/lite-lake/litecore-go/samples/myapp/internal/entities"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/infras/managers"
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
    Config  configmgr.IConfigManager     `inject:""`
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
    "errors"
    "fmt"

    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/entities"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/repositories"
    "github.com/lite-lake/litecore-go/server/builtin/manager/configmgr"
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
    Config     configmgr.IConfigManager      `inject:""`
    Repository repositories.IUserRepository  `inject:""`
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
        return nil, fmt.Errorf("创建用户失败: %w", err)
    }

    return user, nil
}

func (s *userService) GetByID(id uint) (*entities.User, error) {
    user, err := s.Repository.GetByID(id)
    if err != nil {
        return nil, fmt.Errorf("获取用户失败: %w", err)
    }
    if user == nil {
        return nil, errors.New("用户不存在")
    }
    return user, nil
}

func (s *userService) UpdateProfile(id uint, name string) error {
    // 验证输入
    if len(name) < 2 || len(name) > 50 {
        return errors.New("用户名长度必须在 2-50 个字符之间")
    }

    // 获取用户
    user, err := s.Repository.GetByID(id)
    if err != nil {
        return fmt.Errorf("获取用户失败: %w", err)
    }
    if user == nil {
        return errors.New("用户不存在")
    }

    // 更新用户
    user.Name = name
    if err := s.Repository.Update(user); err != nil {
        return fmt.Errorf("更新用户失败: %w", err)
    }

    return nil
}

func (s *userService) DeleteUser(id uint) error {
    // 检查用户是否存在
    user, err := s.Repository.GetByID(id)
    if err != nil {
        return fmt.Errorf("获取用户失败: %w", err)
    }
    if user == nil {
        return errors.New("用户不存在")
    }

    // 删除用户
    if err := s.Repository.Delete(id); err != nil {
        return fmt.Errorf("删除用户失败: %w", err)
    }

    return nil
}

func (s *userService) ListUsers(page, pageSize int) ([]*entities.User, int64, error) {
    if page < 1 {
        page = 1
    }
    if pageSize < 1 || pageSize > 100 {
        pageSize = 20
    }

    offset := (page - 1) * pageSize
    users, total, err := s.Repository.List(offset, pageSize)
    if err != nil {
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
    "github.com/lite-lake/litecore-go/samples/myapp/internal/dtos"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/services"
    "github.com/lite-lake/litecore-go/server/builtin/manager/configmgr"

    "github.com/gin-gonic/gin"
)

type IUserController interface {
    common.IBaseController
}

type userController struct {
    Config      configmgr.IConfigManager `inject:""`
    UserService services.IUserService    `inject:""`
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
        ctx.JSON(http.StatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
        return
    }

    user, err := c.UserService.Register(req.Name, req.Email, req.Age)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, dtos.ErrorResponse(common.HTTPStatusBadRequest, err.Error()))
        return
    }

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

#### 5.5.1 限流器中间件

限流器中间件提供基于 Gin 框架的 HTTP 请求限流功能，支持多种限流策略。

##### 基本用法

```go
package middlewares

import (
    "time"
    "github.com/lite-lake/litecore-go/component/middleware"
)

// 创建按 IP 限流的中间件
// 每分钟最多 100 个请求
func NewRateLimiterMiddleware() middleware.IBaseMiddleware {
    return middleware.NewRateLimiterByIP(100, time.Minute)
}
```

##### 自定义限流策略

```go
// 自定义 Key 生成函数（基于用户ID）
func NewUserRateLimiterMiddleware() middleware.IBaseMiddleware {
    return middleware.NewRateLimiter(&middleware.RateLimiterConfig{
        Limit:     60,
        Window:    time.Minute,
        KeyPrefix: "user_rate_limit",
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
func NewRateLimiterWithSkip() middleware.IBaseMiddleware {
    return middleware.NewRateLimiter(&middleware.RateLimiterConfig{
        Limit:     100,
        Window:    time.Minute,
        KeyPrefix: "api_rate_limit",
        SkipFunc: func(c *gin.Context) bool {
            return c.Request.URL.Path == "/api/health" ||
                   c.Request.URL.Path == "/api/public"
        },
    })
}
```

##### 内置限流器工厂

```go
// 按 IP 限流
middleware.NewRateLimiterByIP(100, time.Minute)

// 按路径限流
middleware.NewRateLimiterByPath(200, time.Minute)

// 按 Header 限流（如 API Key）
middleware.NewRateLimiterByHeader(50, time.Minute, "X-API-Key")

// 按用户 ID 限流（从上下文中获取）
middleware.NewRateLimiterByUserID(60, time.Minute)
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

#### 5.5.2 认证中间件示例

```go
package middlewares

import (
    "strings"

    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/samples/myapp/internal/services"
    "github.com/lite-lake/litecore-go/server/builtin/manager/configmgr"

    "github.com/gin-gonic/gin"
)

type IAuthMiddleware interface {
    common.IBaseMiddleware
}

type authMiddleware struct {
    Config      configmgr.IConfigManager `inject:""`
    AuthService services.IAuthService    `inject:""`
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

#### 5.5.3 中间件执行顺序

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

#### 5.5.4 中间件设计规范

- **横切关注点**：中间件处理认证、日志、CORS 等横切关注点
- **顺序控制**：使用 `Order()` 方法定义执行顺序
- **上下文存储**：使用 `c.Set()` 和 `c.Get()` 存储上下文信息
- **提前终止**：使用 `c.Abort()` 提前终止请求处理

---

## 6. 内置组件

### 6.1 Config（配置）

Config 作为服务器内置组件，由引擎自动初始化。在创建引擎时通过 `builtin.Config` 指定配置文件：

```go
// 引擎自动生成的代码会创建 Config
func NewEngine() (*server.Engine, error) {
    // ...
    return server.NewEngine(
        &builtin.Config{
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

Manager 组件也作为服务器内置组件，由引擎自动初始化。在 `Initialize()` 时自动初始化所有 Manager：

```go
// 框架自动初始化的 Manager（按顺序）
// 1. ConfigManager - 配置管理
// 2. TelemetryManager - 遥测管理
// 3. LoggerManager - 日志管理
// 4. DatabaseManager - 数据库管理
// 5. CacheManager - 缓存管理
// 6. LockManager - 锁管理
// 7. LimiterManager - 限流管理
```

无需手动创建 Manager，只需在代码中通过依赖注入使用：

```go
type userRepository struct {
    Manager databasemgr.IDatabaseManager `inject:""`
}
```

### 6.3 LockMgr（锁管理器）

LockMgr 提供分布式锁功能，支持 Redis 和内存两种实现，用于解决并发访问和资源竞争问题。

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
    
    "github.com/lite-lake/litecore-go/server/builtin/manager/configmgr"
    "github.com/lite-lake/litecore-go/server/builtin/manager/lockmgr"
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

LimiterMgr 提供限流功能，支持 Redis 和内存两种实现，用于保护系统免受过量请求的影响。

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
    "time"
    
    "github.com/lite-lake/litecore-go/server/builtin/manager/configmgr"
    "github.com/lite-lake/litecore-go/server/builtin/manager/limitermgr"
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

### 6.5 可用的内置 Manager

| Manager | 功能 | 支持驱动 |
|---------|------|----------|
| `ConfigManager` | 配置管理 | YAML, JSON |
| `TelemetryManager` | 遥测管理 | OpenTelemetry, None |
| `LoggerManager` | 日志管理 | Zap, None |
| `DatabaseManager` | 数据库管理 | MySQL, PostgreSQL, SQLite, None |
| `CacheManager` | 缓存管理 | Redis, Memory, None |
| `LockManager` | 锁管理 | Redis, Memory, None |
| `LimiterManager` | 限流管理 | Redis, Memory, None |

### 6.6 使用内置组件

在任何层中，都可以通过 `inject:""` 标签自动注入 Config 和 Manager：

```go
type UserServiceImpl struct {
    // 内置组件（由引擎自动注入）
    Config     configmgr.IConfigManager     `inject:""`
    DBManager  databasemgr.IDatabaseManager `inject:""`
    CacheMgr   cachemgr.ICacheManager      `inject:""`
    LockMgr    lockmgr.ILockManager        `inject:""`
    LimiterMgr limitermgr.ILimiterManager  `inject:""`

    // 业务依赖
    UserRepo  IUserRepository             `inject:""`
}
```

---

## 7. 代码生成器使用

LiteCore 提供代码生成器，自动扫描项目中的组件并生成容器代码。

### 6.1 代码生成器配置

位置：`cmd/generate/main.go`

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

    // 自定义配置
    cfg.OutputDir = "internal/application"
    cfg.PackageName = "application"
    cfg.ConfigPath = "configs/config.yaml"

    if err := generator.Run(cfg); err != nil {
        fmt.Fprintf(os.Stderr, "错误: %v\n", err)
        os.Exit(1)
    }
}
```

### 6.2 运行代码生成器

```bash
# 使用默认配置生成
go run ./cmd/generate

# 使用命令行参数生成
go run ./cmd/generate -o internal/application -pkg application -c configs/config.yaml
```

### 6.3 生成时机

需要运行代码生成器的场景：

1. **首次创建项目**：初始化容器代码
2. **新增 Entity**：添加新的实体后
3. **新增 Repository**：添加新的仓储后
4. **新增 Service**：添加新的服务后
5. **新增 Controller**：添加新的控制器后
6. **新增 Middleware**：添加新的中间件后
7. **修改依赖**：修改组件的 `inject` 标签后

### 6.4 生成的文件

代码生成器会自动扫描并生成以下文件：

| 文件 | 说明 |
|------|------|
| `entity_container.go` | 实体容器 |
| `repository_container.go` | 仓储容器 |
| `service_container.go` | 服务容器 |
| `controller_container.go` | 控制器容器 |
| `middleware_container.go` | 中间件容器 |
| `engine.go` | 引擎创建函数 |

**重要**：生成的文件头部标记 `// Code generated by litecore/cli. DO NOT EDIT.`，请勿手动修改。

---

## 8. 依赖注入机制

LiteCore 提供自动化的依赖注入容器，简化组件管理。

### 8.1 注入语法

使用 `inject:""` 标签声明依赖，Config 和 Manager 由引擎自动注入：

```go
type userService struct {
    // 内置组件（引擎自动注入）
    Config     configmgr.IConfigManager      `inject:""`
    DBManager  databasemgr.IDatabaseManager  `inject:""`
    CacheMgr   cachemgr.ICacheManager        `inject:""`

    // 业务依赖
    Repository repositories.IUserRepository  `inject:""`
}
```

### 8.2 依赖规则

| 层级 | 可注入的依赖 |
|------|-------------|
| Repository | Entity, Config, Manager（内置） |
| Service | Repository, Config, Manager（内置）, Service |
| Controller | Service, Config, Manager（内置） |
| Middleware | Service, Config, Manager（内置） |

### 7.3 注入示例

#### Repository 层注入

```go
type userRepository struct {
    // 内置组件（引擎自动注入）
    Config  configmgr.IConfigManager     `inject:""`
    Manager databasemgr.IDatabaseManager `inject:""`

    // 业务依赖
}
```

#### Service 层注入

```go
type userService struct {
    // 内置组件（引擎自动注入）
    Config     configmgr.IConfigManager      `inject:""`
    DBManager  databasemgr.IDatabaseManager  `inject:""`
    CacheMgr   cachemgr.ICacheManager       `inject:""`

    // 业务依赖
    Repository   repositories.IUserRepository  `inject:""`
    OtherService services.IOtherService        `inject:""`
}
```

#### Controller 层注入

```go
type userController struct {
    // 内置组件（引擎自动注入）
    Config    configmgr.IConfigManager `inject:""`

    // 业务依赖
    UserService services.IUserService `inject:""`
}
```

#### Middleware 层注入

```go
type authMiddleware struct {
    // 内置组件（引擎自动注入）
    Config    configmgr.IConfigManager `inject:""`

    // 业务依赖
    AuthService services.IAuthService `inject:""`
}
```

### 8.4 注意事项

1. **不要跨层注入**：Controller 不能直接注入 Repository
2. **接口注入**：优先注入接口，而非具体实现
3. **空标签**：`inject:""` 表示自动注入，无需指定名称
4. **循环依赖**：LiteCore 不支持循环依赖，需要重构代码
5. **类型匹配**：依赖类型必须与声明的接口类型匹配
6. **内置组件**：Config 和 Manager 由引擎自动初始化和注入，无需手动创建

---

## 9. 配置管理

LiteCore 提供统一的配置管理功能，支持 YAML 和 JSON 格式。

### 8.1 配置文件结构

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

# 缓存配置
cache:
  driver: "memory"              # redis, memory, none

# 锁管理配置
lock:
  driver: "memory"              # redis, memory, none
  redis_config:
    host: "localhost"
    port: 6379
    password: ""
    db: 2
    max_idle_conns: 10
    max_open_conns: 100
    conn_max_lifetime: "30s"
  memory_config:
    max_backups: 1000

# 限流管理配置
limiter:
  driver: "memory"              # redis, memory, none
  redis_config:
    host: "localhost"
    port: 6379
    password: ""
    db: 3
    max_idle_conns: 10
    max_open_conns: 100
    conn_max_lifetime: "30s"
  memory_config:
    max_backups: 1000

# 日志配置
logger:
  driver: "zap"                 # zap, none
  zap_config:
    console_enabled: true
    console_config:
      level: "info"

# 遥测配置
telemetry:
  driver: "none"                # none, otel
```

### 8.2 使用配置

```go
import "github.com/lite-lake/litecore-go/server/builtin/manager/configmgr"

// 获取配置值
appName, _ := configProvider.Get("app.name")
port, _ := configProvider.Get("server.port")
enabled, _ := configProvider.Get("logger.console_enabled")

// 检查配置是否存在
if configProvider.Has("database.mysql_config.dsn") {
    // 处理配置
}
```

### 8.3 配置项路径

使用点分隔的路径访问嵌套配置：

```yaml
database:
  driver: "mysql"
  mysql_config:
    dsn: "root:pass@tcp(localhost:3306)/db"
    pool_config:
      max_open_conns: 100
```

访问方式：
```go
configProvider.Get("database.driver")
configProvider.Get("database.mysql_config.dsn")
configProvider.Get("database.mysql_config.pool_config.max_open_conns")
```

---

## 10. 实用工具（util 包）

LiteCore 提供了一系列实用工具包，帮助开发者处理常见的开发任务。

### 9.1 JWT 工具（util/jwt）

JWT 令牌生成、解析和验证。

```go
import (
    "time"
    "github.com/lite-lake/litecore-go/util/jwt"
)

// 生成 JWT Token
secretKey := []byte("your-secret-key")
claims := jwt.MapClaims{
    "user_id": float64(12345),
    "exp":     float64(time.Now().Add(24 * time.Hour).Unix()),
    "iat":     float64(time.Now().Unix()),
}

token, err := jwt.JWT.GenerateHS256Token(claims, secretKey)

// 解析 JWT Token
parsedClaims, err := jwt.JWT.ParseHS256Token(token, secretKey)

// 验证 Claims
err = jwt.JWT.ValidateClaims(parsedClaims)
```

### 9.2 哈希工具（util/hash）

常见哈希算法。

```go
import "github.com/lite-lake/litecore-go/util/hash"

// MD5
md5Hash := hash.Hash.MD5String("hello")

// SHA256
sha256Hash := hash.Hash.SHA256String("hello")
```

### 9.3 加密工具（util/crypt）

密码加密、AES 加密。

```go
import "github.com/lite-lake/litecore-go/util/crypt"

// 密码加密
hashedPassword, err := crypt.BcryptHash("password123")

// 密码验证
err = crypt.BcryptVerify("password123", hashedPassword)

// AES 加密
encrypted, err := crypt.AESEncrypt("plaintext", "key")

// AES 解密
decrypted, err := crypt.AESDecrypt(encrypted, "key")
```

### 9.4 ID 生成工具（util/id）

唯一 ID 生成。

```go
import "github.com/lite-lake/litecore-go/util/id"

// 雪花算法 ID
snowflakeID := id.Snowflake.Generate()

// UUID
uuid := id.UUID.Generate()
```

---

## 11. 最佳实践

### 10.1 目录组织

```
internal/
├── application/         # 自动生成，不要手动修改
├── entities/           # 纯数据实体，无业务逻辑
├── repositories/       # 数据访问层，仅 CRUD
├── services/           # 业务逻辑层，验证、事务、业务规则
├── controllers/        # HTTP 层，仅请求响应处理
├── middlewares/        # 中间件，横切关注点
├── dtos/               # 请求/响应对象
└── infras/             # 基础设施，封装框架组件
    ├── configproviders/ # 配置提供者
    └── managers/        # 管理器封装
```

### 10.2 错误处理

```go
// 在 Service 层包装错误
return nil, fmt.Errorf("failed to create user: %w", err)

// 在 Controller 层返回 HTTP 响应
ctx.JSON(500, gin.H{"error": err.Error()})
```

### 10.3 日志记录

在业务层组件中通过依赖注入使用日志：

```go
import "github.com/lite-lake/litecore-go/server/builtin/manager/loggermgr"

type MyService struct {
    LoggerMgr loggermgr.ILoggerManager `inject:""`
    logger     logger.ILogger
}

func (s *MyService) initLogger() {
    if s.LoggerMgr != nil {
        s.logger = s.LoggerMgr.Ins()
    }
}

func (s *MyService) SomeMethod(userID string) {
    s.initLogger()
    s.logger.Info("操作开始", "user_id", userID)
}

// 注意：main函数中不要使用logger，直接使用fmt和os处理错误即可
// 因为LoggerMgr需要通过引擎初始化后才能使用
```

### 10.4 数据库事务

```go
// 使用 Transaction 方法自动处理提交和回滚
db := r.Manager.DB()
err := db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(user).Error; err != nil {
        return err
    }
    return nil
})
```

### 10.5 缓存使用

```go
// 在 Service 层使用缓存
ctx := context.Background()
cacheKey := fmt.Sprintf("user:%d", id)

var user entities.User
if err := s.CacheMgr.Get(ctx, cacheKey, &user); err == nil {
    return &user, nil
}

// 从数据库查询
user, err := s.Repository.GetByID(id)
if err != nil {
    return nil, err
}

// 写入缓存
s.CacheMgr.Set(ctx, cacheKey, user, time.Hour)
```

### 10.6 测试建议

```go
// 单元测试：使用 Mock 依赖
type mockUserRepository struct {
    repositories.IUserRepository
    // mock 方法
}

// 集成测试：使用 SQLite 内存数据库
sqlite_config:
  dsn: ":memory:"
```

---

## 12. 常见问题

### Q: 如何添加新的 Manager？

在 `internal/infras/managers/` 创建新的 Manager 文件，然后运行 `go run ./cmd/generate`。

### Q: 如何自定义路由？

Controller 的 `GetRouter()` 支持完整的路由语法：
```go
return "/api/users/:id [GET]"
return "/api/users [POST]"
return "/api/files/*filepath [GET]"
```

### Q: 如何支持多种数据库？

在 `configs/config.yaml` 中切换 `database.driver`，无需修改代码：
```yaml
database:
  driver: "mysql"  # 或 "postgresql", "sqlite"
  mysql_config:
    dsn: "user:pass@tcp(localhost:3306)/dbname"
```

### Q: 如何处理循环依赖？

LiteCore 的依赖注入不支持循环依赖。解决方法：
- 重构代码，消除循环依赖
- 使用事件驱动架构
- 将共享逻辑提取到独立的服务

### Q: 如何热重载开发？

```bash
# 安装 air
go install github.com/cosmtrek/air@latest

# 初始化配置
air init

# 运行
air
```

---

## 附录

### A. 完整示例项目

参考 `samples/messageboard` 目录下的留言板示例项目，了解完整的项目结构和代码实现。

### B. 相关文档

- [SOP - 快速实现业务服务](./SOP-build-business-application.md)
- [SOP - 功能包文档撰写](./SOP-package-document.md)
- [Common - 公共基础接口](../common/README.md)
- [Database Manager - 数据库管理器](../server/builtin/manager/databasemgr/README.md)
- [JWT 工具包](../util/jwt/README.md)

### C. 技术支持

如有问题，请提交 Issue 或联系维护团队。
