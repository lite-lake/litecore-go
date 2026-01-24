# CLI - 代码生成器

CLI 是 LiteCore 框架的配套命令行工具，提供代码生成和项目脚手架功能。

## 特性

- **智能扫描**：自动扫描项目中的 Entity、Repository、Service、Controller、Middleware、Listener、Scheduler 组件
- **7 层架构支持**：完整的 7 层依赖注入架构（内置管理器层 → Entity → Repository → Service → 交互层）
- **交互层生成**：自动生成 Controller/Middleware/Listener/Scheduler 四种容器初始化代码
- **项目脚手架**：支持快速创建符合 LiteCore 架构的新项目
- **灵活配置**：支持自定义项目路径、输出目录、包名和配置文件路径
- **双模式使用**：既可作为命令行工具，也可作为库导入使用

## 快速开始

### 安装

#### 方法 1：go install（推荐）
```bash
go install github.com/lite-lake/litecore-go/cli@latest
```
安装后直接使用 `cli` 命令

#### 方法 2：本地构建
```bash
# 构建可执行文件
go build -o litecore-cli ./cli

# 或直接使用 go run
go run ./cli/main.go
```

### 命令概览

```bash
# 查看帮助
litecore-cli --help

# 查看版本
litecore-cli version

# 生成容器代码
litecore-cli generate

# 创建新项目
litecore-cli scaffold

# 生成 Shell 补全
litecore-cli completion bash
```

## 代码生成

### 基本用法

```bash
# 使用默认配置生成（当前目录，输出到 internal/application）
litecore-cli generate

# 指定项目路径
litecore-cli generate --project /path/to/project

# 指定输出目录
litecore-cli generate --output internal/container

# 指定包名
litecore-cli generate --package container

# 指定配置文件路径
litecore-cli generate --config configs/app.yaml

# 完整示例
litecore-cli generate --project . --output internal/application --package application --config configs/config.yaml
```

### 参数说明

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--project` | `-p` | `.` | 项目路径 |
| `--output` | `-o` | `internal/application` | 输出目录 |
| `--package` | - | `application` | 包名 |
| `--config` | `-c` | `configs/config.yaml` | 配置文件路径 |

### 作为库使用

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

## 项目脚手架

### 基本用法

```bash
# 交互式模式（推荐）
litecore-cli scaffold

# 使用 basic 模板
litecore-cli scaffold --module github.com/user/app --project myapp --template basic

# 使用 standard 模板
litecore-cli scaffold --module github.com/user/app --project myapp --template standard

# 使用 full 模板
litecore-cli scaffold --module github.com/user/app --project myapp --template full
```

### 参数说明

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--module` | `-m` | - | 模块路径（如 github.com/user/app） |
| `--project` | `-n` | - | 项目名称 |
| `--output` | `-o` | - | 输出目录 |
| `--template` | `-t` | - | 模板类型（basic/standard/full） |
| `--interactive` | `-i` | - | 交互式模式 |

### 模板类型

| 模板 | 包含内容 |
|------|---------|
| `basic` | 目录结构 + go.mod + README |
| `standard` | basic + 配置文件 + 基础中间件 + 自动生成容器代码 |
| `full` | standard + 完整示例代码（entity/repository/service/controller/listener/scheduler） |

## 代码约定

### 目录结构

```
项目根目录/
├── internal/
│   ├── entities/        # 实体层：导出的结构体类型
│   ├── repositories/    # 仓储层：I 前缀接口 + New 前缀工厂函数
│   ├── services/        # 服务层：I 前缀接口 + New 前缀工厂函数
│   ├── controllers/     # 交互层 - 控制器：I 前缀接口 + New 前缀工厂函数
│   ├── middlewares/     # 交互层 - 中间件：I 前缀接口 + New 前缀工厂函数
│   ├── listeners/       # 交互层 - 监听器：I 前缀接口 + New 前缀工厂函数
│   ├── schedulers/      # 交互层 - 定时器：I 前缀接口 + New 前缀工厂函数
│   └── application/     # 生成的容器代码
└── configs/
    └── config.yaml      # 配置文件
```

### 命名规范

| 层级 | 接口命名 | 工厂函数命名 | 实现结构体 |
|------|----------|--------------|-----------|
| Entity | `MessageEntity` | 无需工厂函数 | `Message` |
| Repository | `IMessageRepository` | `NewMessageRepository()` | `messageRepositoryImpl` |
| Service | `IMessageService` | `NewMessageService()` | `messageServiceImpl` |
| Controller | `IMessageController` | `NewMessageController()` | `msgControllerImpl` |
| Middleware | `IAuthMiddleware` | `NewAuthMiddleware()` | `authMiddlewareImpl` |
| Listener | `IMessageListener` | `NewMessageListener()` | `messageListenerImpl` |
| Scheduler | `ICleanupScheduler` | `NewCleanupScheduler()` | `cleanupSchedulerImpl` |

### 组件示例

#### Repository 示例

```go
package repositories

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/manager/databasemgr"
)

type IMessageRepository interface {
    common.IBaseRepository
}

type messageRepositoryImpl struct {
    DBManager databasemgr.IDatabaseManager `inject:""`
}

func NewMessageRepository() IMessageRepository {
    return &messageRepositoryImpl{}
}

var _ IMessageRepository = (*messageRepositoryImpl)(nil)
```

#### Service 示例

```go
package services

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/internal/repositories"
)

type IMessageService interface {
    common.IBaseService
}

type messageServiceImpl struct {
    MessageRepo repositories.IMessageRepository `inject:""`
}

func NewMessageService() IMessageService {
    return &messageServiceImpl{}
}

var _ IMessageService = (*messageServiceImpl)(nil)
```

#### Controller 示例

```go
package controllers

import (
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/internal/services"
    "github.com/lite-lake/litecore-go/server/http"
)

type IMessageController interface {
    common.IBaseController
}

type messageControllerImpl struct {
    MessageService services.IMessageService `inject:""`
}

func NewMessageController() IMessageController {
    return &messageControllerImpl{}
}

func (c *messageControllerImpl) GetRouter() http.Router {
    return nil
}

var _ IMessageController = (*messageControllerImpl)(nil)
```

## 生成的文件

```
internal/application/
├── entity_container.go      # 实体容器初始化
├── repository_container.go  # 仓储容器初始化
├── service_container.go     # 服务容器初始化
├── controller_container.go  # 控制器容器初始化
├── middleware_container.go  # 中间件容器初始化
├── listener_container.go    # 监听器容器初始化
├── scheduler_container.go   # 定时器容器初始化
└── engine.go                # 引擎创建函数
```

## 使用生成的代码

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

## API 说明

### generator.Config

```go
type Config struct {
    ProjectPath string // 项目路径
    OutputDir   string // 输出目录
    PackageName string // 包名
    ConfigPath  string // 配置文件路径
}
```

### 主要函数

```go
// 运行代码生成器
func Run(cfg *Config) error

// 运行代码生成器，失败时 panic
func MustRun(cfg *Config)

// 返回默认配置
func DefaultConfig() *Config

// 查找项目模块名
func FindModuleName(projectPath string) (string, error)
```

### scaffold.Config

```go
type Config struct {
    ModulePath    string       // 模块路径
    ProjectName   string       // 项目名称
    OutputDir     string       // 输出目录
    TemplateType  TemplateType  // 模板类型
    Interactive   bool         // 是否交互式
    LitecoreGoVer string       // LiteCore Go 版本
}

type TemplateType string

const (
    TemplateTypeBasic    TemplateType = "basic"
    TemplateTypeStandard TemplateType = "standard"
    TemplateTypeFull     TemplateType = "full"
)
```

## 工作流程

### 代码生成流程

1. **扫描项目**：解析器扫描 `internal/` 目录下的各层组件
2. **提取组件**：提取接口定义、工厂函数
3. **生成代码**：构建器使用模板生成容器初始化代码
4. **写入文件**：将生成的代码写入输出目录

### 脚手架创建流程

1. **创建目录结构**：根据模板类型创建项目目录
2. **生成配置文件**：生成 go.mod、config.yaml 等
3. **生成示例代码**：根据模板类型生成示例组件
4. **调用生成器**：自动运行代码生成器

## 注意事项

1. **不要手动修改生成的代码**：所有生成的文件头部包含 `// Code generated by litecore/cli. DO NOT EDIT.` 注释
2. **接口命名**：组件接口必须使用 `I` 前缀（如 `IMessageService`）
3. **工厂函数**：工厂函数必须使用 `New` 前缀（如 `NewMessageService`）
4. **配置文件**：配置文件必须存在且格式正确
5. **模块路径**：项目必须有 go.mod 文件，且能正确解析模块名
6. **依赖注入**：需要注入的依赖必须使用 `inject:""` 标签，工厂函数不接受参数

## 完整示例

参考 `samples/messageboard` 项目，展示了完整的代码生成和使用流程：

```bash
# 进入示例项目
cd samples/messageboard

# 运行代码生成器
go run ../cli/main.go generate

# 查看生成的代码
ls internal/application/

# 运行应用
go run cmd/server/main.go
```

## 常见问题

**Q: 为什么需要代码生成器？**

A: LiteCore 采用 7 层依赖注入架构，手动编写容器初始化代码容易出错且维护困难。代码生成器自动化这一过程，确保代码的一致性和正确性。

**Q: 生成的代码可以手动修改吗？**

A: 不可以。所有生成的代码头部都包含 `// Code generated by litecore/cli. DO NOT EDIT.` 注释。如果需要修改，应该更新业务代码后重新生成。

**Q: 如何更新生成的代码？**

A: 修改业务代码（添加/删除组件）后，重新运行代码生成器即可。

**Q: 支持哪些层级？**

A: 支持 7 层架构：Entity、Repository、Service、Controller、Middleware、Listener、Scheduler。

**Q: basic/standard/full 模板有什么区别？**

A: basic 只包含基础结构，standard 包含配置文件和基础中间件，full 包含完整的示例代码。

**Q: 代码生成器会影响性能吗？**

A: 代码生成器仅在开发阶段运行，不影响生产环境的性能。

**Q: 如何自定义输出目录？**

A: 使用 `--output` 参数或修改 Config 的 `OutputDir` 字段。

**Q: factory 函数为什么不接受参数？**

A: LiteCore 的依赖注入机制会自动注入 `inject:""` 标记的依赖，因此工厂函数不需要显式传入参数。
