# LiteCore CLI

LiteCore CLI 是一个代码生成工具，用于自动生成 LiteCore 框架的容器初始化代码。

## 功能

 - 自动扫描项目中的实体、仓储、服务、控制器、中间件等组件
 - 按照 LiteCore 的分层架构生成容器初始化代码
 - 每个容器生成一个独立的 Go 文件
 - 支持自定义配置路径、输出目录和包名

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
./litecore-generate -project . -output internal/application -package application -configmgr configs/config.yaml
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

## 示例

### 独立工具使用

```bash
# 在任意目录生成
./litecore-generate -project /path/to/project -output internal/application -package application -configmgr configs/config.yaml
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

## 注意事项

1. 生成的代码不应手动修改，因为每次运行生成器都会覆盖这些文件
2. 如果需要自定义容器的初始化逻辑，可以在生成后手动修改相关文件

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

## 版本

当前版本: 1.0.0

## 许可证

与 LiteCore 框架相同
