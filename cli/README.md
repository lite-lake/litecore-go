# CLI - 代码生成器

 CLI 是 LiteCore 框架的代码生成工具，用于自动生成 5 层依赖注入架构的容器初始化代码，大幅减少手动编写依赖注入配置的工作量。

 ## 特性

 - **智能扫描**：自动扫描项目中的 Entity、Repository、Service、Controller、Middleware、Listener、Scheduler 组件
 - **5 层架构支持**：完整的 5 层依赖注入架构（内置管理器层 → Entity → Repository → Service → 交互层）
 - **交互层生成**：自动生成 Controller/Middleware/Listener/Scheduler 四种容器初始化代码
 - **自动依赖注入**：自动识别并注册 `inject:""` 标签的依赖
 - **灵活配置**：支持自定义项目路径、输出目录、包名和配置文件路径
 - **双模式使用**：既可作为独立命令行工具，也可作为库导入使用
 - **类型安全**：生成的代码遵循 Go 类型系统，编译时检查依赖关系

## 快速开始

### 方式一：使用独立命令行工具

```bash
# 进入项目目录
cd /path/to/your/project

# 构建代码生成器
go build -o litecore-generate ./cli

# 使用默认配置生成
./litecore-generate

# 使用自定义参数
./litecore-generate -project . -output internal/application -package application -config configs/config.yaml
```

### 方式二：作为库导入使用

在业务项目中创建自定义生成器入口：

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

    // 支持命令行参数覆盖
    outputDir := flag.String("o", cfg.OutputDir, "输出目录")
    packageName := flag.String("pkg", cfg.PackageName, "包名")
    configPath := flag.String("c", cfg.ConfigPath, "配置文件路径")
    flag.Parse()

    if outputDir != nil {
        cfg.OutputDir = *outputDir
    }
    if packageName != nil {
        cfg.PackageName = *packageName
    }
    if configPath != nil {
        cfg.ConfigPath = *configPath
    }

    if err := generator.Run(cfg); err != nil {
        fmt.Fprintf(os.Stderr, "错误: %v\n", err)
        os.Exit(1)
    }
}
```

参考示例：`samples/messageboard/cmd/generate/main.go`

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
 │   └── infras/          # 基础设施层：New 前缀工厂函数
 └── configs/
     └── config.yaml      # 配置文件
 ```

 ### 命名规范

 | 层级 | 接口命名 | 工厂函数命名 | 实现结构体 |
 |------|----------|--------------|-----------|
 | Entity | `MessageEntity` | 无需工厂函数 | `Message` |
 | Repository | `IMessageRepository` | `NewMessageRepository()` | `messageRepositoryImpl` |
 | Service | `IMessageService` | `NewMessageService()` | `messageServiceImpl` |
 | 交互层 - Controller | `IMessageController` | `NewMessageController()` | `msgControllerImpl` |
 | 交互层 - Middleware | `IAuthMiddleware` | `NewAuthMiddleware()` | `authMiddlewareImpl` |
 | 交互层 - Listener | `IMessageListener` | `NewMessageListener()` | `messageListenerImpl` |
 | 交互层 - Scheduler | `ICleanupScheduler` | `NewCleanupScheduler()` | `cleanupSchedulerImpl` |

### 依赖注入规范

所有需要依赖注入的组件都必须使用 `inject:""` 标签：

```go
type MessageServiceImpl struct {
    Config    configmgr.IConfigManager    `inject:""`
    DBManager databasemgr.IDatabaseManager `inject:""`
    LoggerMgr loggermgr.ILoggerManager    `inject:""`
    Repo      repositories.IMessageRepository `inject:""`
}

// NewMessageService 创建服务实例（空实现，依赖由容器注入）
func NewMessageService() IMessageService {
    return &MessageServiceImpl{}
}
```

## 代码生成器架构

CLI 代码生成器由以下核心模块组成：

 ### 1. Parser（解析器）
 负责解析项目中的 Go 源代码，提取各层组件信息。

 - 解析 `internal/entities/` 目录，提取导出的结构体类型
 - 解析 `internal/repositories/`、`internal/services/` 目录，提取 `I` 前缀的接口定义和 `New` 前缀的工厂函数
 - 解析 `internal/controllers/`、`internal/middlewares/`、`internal/listeners/`、`internal/schedulers/`（交互层）目录，提取 `I` 前缀的接口定义和 `New` 前缀的工厂函数
 - 解析 `internal/infras/` 目录，提取 `New` 前缀的工厂函数

 ### 2. Builder（构建器）
 根据解析结果，使用模板引擎生成容器初始化代码。

 - `InitEntityContainer()` - 实体容器初始化
 - `InitRepositoryContainer()` - 仓储容器初始化
 - `InitServiceContainer()` - 服务容器初始化
 - `InitControllerContainer()` - 交互层 - 控制器容器初始化
 - `InitMiddlewareContainer()` - 交互层 - 中间件容器初始化
 - `InitListenerContainer()` - 交互层 - 监听器容器初始化
 - `InitSchedulerContainer()` - 交互层 - 定时器容器初始化
 - `NewEngine()` - 引擎创建函数

### 3. Template（模板）
定义生成代码的结构和格式，使用 Go 标准库 `text/template`。

### 4. Analyzer（分析器）
辅助分析器，用于检测代码层级和提取组件信息。

## 命令行参数

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `-project` | `-p` | `.` | 项目路径 |
| `-output` | `-o` | `internal/application` | 输出目录 |
| `-package` | `-pkg` | `application` | 包名 |
| `-config` | `-c` | `configs/config.yaml` | 配置文件路径 |
| `-version` | `-v` | - | 显示版本信息 |

## 配置说明

### Config 结构体

```go
type Config struct {
    ProjectPath string // 项目路径
    OutputDir   string // 输出目录
    PackageName string // 包名
    ConfigPath  string // 配置文件路径
}
```

### 默认配置

```go
func DefaultConfig() *Config {
    return &Config{
        ProjectPath: ".",
        OutputDir:   "internal/application",
        PackageName: "application",
        ConfigPath:  "configs/config.yaml",
    }
}
```

## API 说明

### generator.Run

运行代码生成器。

```go
func Run(cfg *Config) error
```

### generator.MustRun

运行代码生成器，失败时 panic。

```go
func MustRun(cfg *Config)
```

### generator.DefaultConfig

返回默认配置。

```go
func DefaultConfig() *Config
```

### generator.FindModuleName

查找项目模块名（解析 go.mod）。

```go
func FindModuleName(projectPath string) (string, error)
```

## 生成的文件

```
internal/application/
├── entity_container.go      # 实体容器初始化
├── repository_container.go  # 仓储容器初始化
├── service_container.go     # 服务容器初始化
├── controller_container.go  # 控制器容器初始化
├── middleware_container.go # 中间件容器初始化
├── listener_container.go   # 监听器容器初始化
├── scheduler_container.go  # 定时器容器初始化
└── engine.go               # 引擎创建函数
```

## 使用生成的代码

```go
package main

import (
    "log"

    app "github.com/example/project/internal/application"
)

func main() {
    // 创建应用引擎
    engine, err := app.NewEngine()
    if err != nil {
        log.Fatalf("Failed to create engine: %v", err)
    }

    // 初始化引擎（注册路由、依赖注入等）
    if err := engine.Initialize(); err != nil {
        log.Fatalf("Failed to initialize engine: %v", err)
    }

    // 启动引擎（启动所有 Manager 和 HTTP 服务器）
    if err := engine.Start(); err != nil {
        log.Fatalf("Failed to start engine: %v", err)
    }

    // 等待关闭信号
    engine.WaitForShutdown()
}
```

## 工作流程

1. **扫描项目**：解析器扫描 `internal/` 目录下的各层组件
2. **提取组件**：提取接口定义、工厂函数、依赖注入标签
3. **生成代码**：构建器使用模板生成容器初始化代码
4. **写入文件**：将生成的代码写入输出目录

## 注意事项

1. **不要手动修改生成的代码**：所有生成的文件头部包含 `// Code generated by litecore/cli. DO NOT EDIT.` 注释
2. **接口命名**：组件接口必须使用 `I` 前缀（如 `IMessageService`）
3. **工厂函数**：工厂函数必须使用 `New` 前缀（如 `NewMessageService`）
4. **配置文件**：配置文件必须存在且格式正确
5. **模块路径**：项目必须有 go.mod 文件，且能正确解析模块名
6. **依赖注入**：需要注入的依赖必须使用 `inject:""` 标签，且工厂函数不接受参数（依赖由容器自动注入）

## 完整示例

参考 `samples/messageboard` 项目，展示了完整的代码生成和使用流程：

```bash
# 进入示例项目
cd samples/messageboard

# 运行代码生成器
go run cmd/generate/main.go

# 查看生成的代码
ls internal/application/

# 运行应用
go run cmd/server/main.go
```

## 常见问题

 **Q: 为什么需要代码生成器？**

 A: LiteCore 采用 5 层依赖注入架构，手动编写容器初始化代码容易出错且维护困难。代码生成器自动化这一过程，确保代码的一致性和正确性。

**Q: 生成的代码可以手动修改吗？**

A: 不可以。所有生成的代码头部都包含 `// Code generated by litecore/cli. DO NOT EDIT.` 注释。如果需要修改，应该更新业务代码后重新生成。

**Q: 如何更新生成的代码？**

A: 修改业务代码（添加/删除组件）后，重新运行代码生成器即可。

**Q: 为什么工厂函数不接受参数？**

A: LiteCore 的依赖注入机制会自动注入 `inject:""` 标记的依赖，因此工厂函数不需要显式传入参数。

 **Q: 支持哪些层级？**

 A: 支持 5 层架构：内置管理器层、Entity、Repository、Service、交互层（Controller/Middleware/Listener/Scheduler）。

**Q: 如何自定义输出目录？**

A: 使用 `-output` 参数或修改 Config 的 `OutputDir` 字段。

**Q: 代码生成器会影响性能吗？**

A: 代码生成器仅在开发阶段运行，不影响生产环境的性能。
