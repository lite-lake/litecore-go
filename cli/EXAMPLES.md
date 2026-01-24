# CLI 工具使用示例

## 概述

`litecore-cli` 是 litecore 框架提供的代码生成器命令行工具，用于自动生成依赖注入容器代码。

## 安装

```bash
# 从源码运行
go run ./cli/main.go [参数]

# 编译为可执行文件
go build -o litecore-generate ./cli/main.go

# 安装到 $GOPATH/bin
go install ./cli/main.go
```

## 命令行参数

| 参数 | 简写 | 说明 | 默认值 |
|------|------|------|--------|
| `version` | `v` | 显示版本信息 | - |
| `project` | `p` | 项目路径 | `.` |
| `output` | `o` | 输出目录 | `internal/application` |
| `package` | `pkg` | 包名 | `application` |
| `configmgr` | `c` | 配置文件路径 | `configs/config.yaml` |

## 基本使用

### 查看版本

```bash
go run ./cli/main.go -version
# 或
go run ./cli/main.go -v
```

### 使用默认配置生成

```bash
# 在当前项目目录执行
go run ./cli/main.go
```

### 自定义输出目录

```bash
go run ./cli/main.go -o internal/app
```

### 自定义包名

```bash
go run ./cli/main.go -pkg myapp
```

### 自定义配置文件路径

```bash
go run ./cli/main.go -c config/dev.yaml
```

### 完整自定义

```bash
go run ./cli/main.go \
  -p /path/to/project \
  -o internal/application \
  -pkg application \
  -c configs/config.yaml
```

## 项目集成示例

### 方式一：命令行工具调用

在项目根目录执行：

```bash
# 生成默认配置
litecore-generate

# 或使用完整路径
go run github.com/lite-lake/litecore-go/cli/main.go \
  -o internal/application \
  -pkg application \
  -c configs/config.yaml
```

### 方式二：嵌入到项目中

创建 `cmd/generate/main.go`:

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

运行：

```bash
go run ./cmd/generate -o internal/app
```

### 方式三：最简实现

创建 `cmd/generate/main.go`:

```go
package main

import "github.com/lite-lake/litecore-go/cli/generator"

func main() {
	generator.MustRun(generator.DefaultConfig())
}
```

运行：

```bash
go run ./cmd/generate
```

## 实际项目示例：messageboard

messageboard 示例展示了完整的 CLI 工具使用方式：

### 1. 目录结构

```
samples/messageboard/
├── cmd/
│   ├── generate/          # 代码生成器入口
│   │   └── main.go
│   ├── genpasswd/         # 密码生成工具
│   │   └── main.go
│   └── server/            # 应用启动入口
│       └── main.go
├── configs/
│   └── config.yaml        # 应用配置文件
├── internal/
│   ├── application/       # 生成的容器代码
│   │   ├── entity_container.go
│   │   ├── repository_container.go
│   │   ├── service_container.go
│   │   ├── controller_container.go
│   │   ├── middleware_container.go
│   │   └── engine.go
│   ├── entities/
│   ├── repositories/
│   ├── services/
│   ├── controllers/
│   └── middlewares/
└── go.mod
```

### 2. 生成容器代码

在 `samples/messageboard` 目录下执行：

```bash
go run ./cmd/generate
```

这将：
- 扫描 `internal/entities` 目录生成实体容器
- 扫描 `internal/repositories` 目录生成仓储容器
- 扫描 `internal/services` 目录生成服务容器
- 扫描 `internal/controllers` 目录生成控制器容器
- 扫描 `internal/middlewares` 目录生成中间件容器
- 生成 `engine.go` 引擎初始化代码

### 3. 使用生成的代码

在 `cmd/server/main.go` 中使用：

```go
package main

import (
	"fmt"
	"os"

	messageboardapp "github.com/lite-lake/litecore-go/samples/messageboard/internal/application"
)

func main() {
	// 创建应用引擎
	engine, err := messageboardapp.NewEngine()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create engine: %v\n", err)
		os.Exit(1)
	}

	// 初始化引擎（注册路由、依赖注入等）
	if err := engine.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize engine: %v\n", err)
		os.Exit(1)
	}

	// 启动引擎（启动所有 Manager 和 HTTP 服务器）
	if err := engine.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start engine: %v\n", err)
		os.Exit(1)
	}

	// 等待关闭信号
	engine.WaitForShutdown()
}
```

### 4. 配置文件

生成的容器代码会读取 `configs/config.yaml` 来初始化内置 Manager 组件：

```yaml
logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "info"
      format: "gin"

database:
  driver: "sqlite"
  sqlite_config:
    dsn: "./data/messageboard.db"

cache:
  driver: "memory"

limiter:
  driver: "memory"

lock:
  driver: "memory"

mq:
  driver: "memory"

telemetry:
  driver: "none"
```

## 生成的代码说明

生成器会在指定的输出目录创建以下文件：

- `entity_container.go` - 实体容器初始化
- `repository_container.go` - 仓储容器初始化
- `service_container.go` - 服务容器初始化
- `controller_container.go` - 控制器容器初始化
- `middleware_container.go` - 中间件容器初始化
- `engine.go` - 引擎创建函数

### 生成的 Engine 接口

```go
// NewEngine 创建应用引擎
func NewEngine() (*server.Engine, error)
```

返回的 `server.Engine` 提供以下方法：

```go
// Initialize 初始化引擎
func (e *Engine) Initialize() error

// Start 启动引擎
func (e *Engine) Start() error

// Stop 停止引擎
func (e *Engine) Stop() error

// WaitForShutdown 等待关闭信号
func (e *Engine) WaitForShutdown()
```

## API 参考

### Config 配置结构

```go
type Config struct {
    ProjectPath string  // 项目路径（默认: "."）
    OutputDir   string  // 输出目录（默认: "internal/application"）
    PackageName string  // 包名（默认: "application"）
    ConfigPath  string  // 配置文件路径（默认: "configs/config.yaml"）
}
```

### 可用函数

```go
// DefaultConfig 返回默认配置
func DefaultConfig() *Config

// Run 运行生成器
func Run(cfg *Config) error

// MustRun 运行生成器，失败时 panic
func MustRun(cfg *Config)
```

## 注意事项

1. **代码覆盖**：生成的代码会覆盖同名文件，请确保不要手动修改生成的文件
2. **路径规范**：配置文件路径必须是相对于项目根目录的路径
3. **模块依赖**：确保项目已正确初始化 `go mod`，生成器会读取 `go.mod` 文件
4. **目录结构**：确保项目包含标准的分层目录结构（entities, repositories, services, controllers, middlewares）
5. **依赖注入标签**：在需要依赖注入的字段上使用 `inject:""` 标签

## 依赖注入标签示例

```go
type MessageService struct {
    Config    configmgr.IConfigManager    `inject:""`
    DBManager databasemgr.IDatabaseManager `inject:""`
    LoggerMgr loggermgr.ILoggerManager   `inject:""`
    Repo      repository.IMessageRepository `inject:""`
    logger    loggermgr.ILogger
}
```
