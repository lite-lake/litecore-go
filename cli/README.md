# LiteCore CLI

LiteCore CLI 是一个代码生成工具，用于自动生成 LiteCore 框架的容器初始化代码。

## 功能

 - 自动扫描项目中的实体、仓储、服务、控制器、中间件等组件
 - 按照 LiteCore 的分层架构生成容器初始化代码
 - 每个容器生成一个独立的 Go 文件
 - 支持自定义配置路径、输出目录和包名

## 项目结构说明

LiteCore 采用 5 层分层架构：

```
项目根目录/
├── entity/           # 实体层（无依赖）
├── repository/       # 仓储层（依赖 Entity + Config + Manager）
├── service/          # 服务层（依赖 Repository + Config + Manager）
├── controller/       # 控制器层（依赖 Service + Config + Manager）
└── middleware/       # 中间件层（依赖 Service + Config + Manager）
```

Manager 组件位置：
- 所有 Manager 位于 `manager/` 目录下
- 包括：`configmgr`, `databasemgr`, `loggermgr`, `cachemgr`, `lockmgr`, `limitermgr`, `mqmgr`, `telemetrymgr`

内置组件位置：
- 控制器：`component/litecontroller/`
- 中间件：`component/litemiddleware/`
- 服务：`component/liteservice/`

## 安装

### 方式一：作为独立命令行工具

```bash
go build -o litecore-generate ./cli
```

### 方式二：在业务项目中导入使用

业务项目可以导入 `github.com/lite-lake/litecore-go/cli/generator` 包，调用简化的 API：

```go
import "github.com/lite-lake/litecore-go/cli/generator"

func main() {
    // 使用默认配置
    cfg := generator.DefaultConfig()
    generator.Run(cfg)
}

// 或者自定义配置
func main() {
    cfg := &generator.Config{
        ProjectPath: ".",
        OutputDir:   "internal/application",
        PackageName: "application",
        ConfigPath:  "configs/config.yaml",
    }
    generator.Run(cfg)
}
```

## 使用方法

### 方式一：作为独立命令行工具使用

```bash
# 在项目根目录执行
./litecore-generate

# 或指定参数
./litecore-generate -project . -output internal/application -package application -config configs/config.yaml
```

### 方式二：在业务项目中自定义生成器入口

在业务项目中创建自定义的生成器入口，例如 `cmd/generate/main.go`：

```go
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lite-lake/litecore-go/cli/generator"
)

func main() {
    // 使用默认配置
    cfg := generator.DefaultConfig()

    // 支持命令行参数覆盖
    outputDir := flag.String("o", cfg.OutputDir, "输出目录")
    packageName := flag.String("pkg", cfg.PackageName, "包名")
    configPath := flag.String("c", cfg.ConfigPath, "配置文件路径")
    flag.Parse()

    // 更新配置
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

然后运行：

```bash
go run ./cmd/generate
```

### 命令行参数（独立工具）

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `-project` | `-p` | `.` | 项目路径 |
| `-output` | `-o` | `internal/application` | 输出目录 |
| `-package` | `-pkg` | `application` | 包名 |
| `-config` | `-c` | `configs/config.yaml` | 配置文件路径 |
| `-version` | `-v` | - | 显示版本信息 |

## 配置文件

生成的代码使用配置文件来初始化内置 Manager 组件。配置文件示例（`configs/config.yaml`）：

```yaml
# 日志配置
logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "info"                               # 日志级别：debug, info, warn, error, fatal
      format: "gin"                                # 格式：gin | json | default
      color: true                                  # 是否启用颜色
      time_format: "2006-01-24 15:04:05.000"     # 时间格式

# 数据库配置
database:
  driver: "sqlite"
  sqlite_config:
    dsn: "./data/app.db"

# 缓存配置
cache:
  driver: "memory"

# 其他 Manager 配置...
```

## 示例

### 独立工具使用

```bash
# 在任意目录生成
./litecore-generate -project /path/to/project -output internal/application -package application -config configs/config.yaml
```

### 业务项目入口使用

参考 `samples/messageboard/cmd/generate/main.go`：

```bash
cd samples/messageboard
go run ./cmd/generate
```

## 生成的文件

```
internal/application/
├── entity_container.go        # 实体容器初始化
├── repository_container.go    # 仓储容器初始化
├── service_container.go       # 服务容器初始化
├── controller_container.go    # 控制器容器初始化
├── middleware_container.go    # 中间件容器初始化
└── engine.go                   # 引擎创建函数
```

## 使用生成的代码

在 `main.go` 中使用生成的代码：

```go
package main

import (
    "log"

    messageboardapp "github.com/lite-lake/litecore-go/samples/messageboard/internal/application"
)

func main() {
    engine, err := messageboardapp.NewEngine()
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

## 生成的代码说明

### 引擎初始化

生成的 `NewEngine` 函数会按照以下顺序初始化组件：

1. 创建各层容器（Entity → Repository → Service → Controller → Middleware）
2. 使用配置文件初始化内置 Manager 组件（通过 `server.BuiltinConfig`）
3. 内置 Manager 包括：
   - ConfigManager（配置管理）
   - TelemetryManager（遥测）
   - LoggerManager（日志）
   - DatabaseManager（数据库）
   - CacheManager（缓存）
   - LockManager（分布式锁）
   - LimiterManager（限流）
   - MQManager（消息队列）

### 依赖注入

容器会自动识别并注入以下标记的依赖：

```go
type MyService struct {
    Config    configmgr.IConfigManager    `inject:""`
    DBManager databasemgr.IDatabaseManager `inject:""`
    LoggerMgr loggermgr.ILoggerManager    `inject:""`
    Repo      repository.IMyRepository    `inject:""`
}
```

## 注意事项

1. 生成的代码不应手动修改，因为每次运行生成器都会覆盖这些文件
2. 如果需要自定义容器的初始化逻辑，可以在生成后手动修改相关文件
3. 配置文件必须存在且格式正确，否则无法启动
4. Manager 组件会根据配置文件中的 `driver` 字段自动选择实现

## 工作原理

1. **扫描阶段**: 扫描项目目录，识别符合 LiteCore 分层架构的组件
2. **解析阶段**: 使用 Go 的 AST 解析器解析源代码，提取接口和工厂函数信息
3. **生成阶段**: 使用 Go 模板引擎生成容器初始化代码
4. **写入阶段**: 将生成的代码写入到指定目录

## API 使用示例

### 简单调用

```go
import "github.com/lite-lake/litecore-go/cli/generator"

// 使用默认配置
err := generator.Run(generator.DefaultConfig())
```

### 自定义配置

```go
import "github.com/lite-lake/litecore-go/cli/generator"

cfg := &generator.Config{
    ProjectPath: ".",
    OutputDir:   "internal/application",
    PackageName: "application",
    ConfigPath:  "configs/config.yaml",
}
err := generator.Run(cfg)
```

### MustRun（失败时 panic）

```go
import "github.com/lite-lake/litecore-go/cli/generator"

// 失败时 panic，适合 main 包
generator.MustRun(generator.DefaultConfig())
```

## 日志配置

LiteCore 支持三种日志格式，可在配置文件中设置：

### Gin 格式（推荐用于控制台）

```yaml
logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "info"
      format: "gin"           # Gin 风格：竖线分隔，带颜色
      color: true
      time_format: "2006-01-24 15:04:05.000"
```

Gin 格式输出示例：
```
2026-01-24 15:04:05.123 | INFO  | 开始依赖注入 | count=23
2026-01-24 15:04:05.456 | WARN  | 慢查询检测 | duration=1.2s
2026-01-24 15:04:05.789 | ERROR | 数据库连接失败 | error="connection refused"
```

### JSON 格式（适合日志分析）

```yaml
logger:
  zap_config:
    console_config:
      format: "json"          # JSON 格式
```

### Default 格式

```yaml
logger:
  zap_config:
    console_config:
      format: "default"       # 默认 ConsoleEncoder 格式
```

## 版本

当前版本: 1.0.0

## 许可证

与 LiteCore 框架相同
