# 最简调用示例

## 业务项目中的最简实现

### 最简单的方式（无需任何参数）

```go
// cmd/generate/main.go
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

### 支持命令行参数

```go
// cmd/generate/main.go
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
# 使用默认配置
go run ./cmd/generate

# 自定义输出目录
go run ./cmd/generate -o internal/app

# 自定义包名
go run ./cmd/generate -pkg myapp

# 自定义配置文件
go run ./cmd/generate -c config/dev.yaml
```

### 完全自定义配置

```go
// cmd/generate/main.go
package main

import (
	"fmt"
	"os"

	"github.com/lite-lake/litecore-go/cli/generator"
)

func main() {
	cfg := &generator.Config{
		ProjectPath: ".",
		OutputDir:   "internal/app",
		PackageName: "myapp",
		ConfigPath:  "configs/config.yaml",
	}

	if err := generator.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}
```

## 对比

### 旧方式（复杂）
```go
func run(projectPath, outputDir, packageName, configPath string) error {
	parser := generator.NewParser(projectPath)
	moduleName, err := generator.FindModuleName(projectPath)
	if err != nil {
		return fmt.Errorf("查找模块名失败: %w", err)
	}
	info, err := parser.Parse(moduleName)
	if err != nil {
		return fmt.Errorf("解析项目失败: %w", err)
	}
	builder := generator.NewBuilder(projectPath, outputDir, packageName, moduleName, configPath)
	if err := builder.Generate(info); err != nil {
		return fmt.Errorf("生成代码失败: %w", err)
	}
	fmt.Printf("成功生成容器代码到 %s\n", outputDir)
	return nil
}
```

### 新方式（简单）
```go
generator.Run(&generator.Config{
    ProjectPath: ".",
    OutputDir:   "internal/application",
    PackageName: "application",
    ConfigPath:  "configs/config.yaml",
})
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

## 实际项目示例

参考 `samples/messageboard/cmd/generate/main.go` 的实现。

## 配置文件示例

生成器生成的代码会读取配置文件来初始化内置 Manager 组件。典型的配置文件结构：

```yaml
# 日志配置（支持 Gin 格式）
logger:
  driver: "zap"
  zap_config:
    console_enabled: true
    console_config:
      level: "info"
      format: "gin"                    # gin | json | default
      color: true
      time_format: "2006-01-24 15:04:05.000"

# 数据库配置
database:
  driver: "sqlite"
  sqlite_config:
    dsn: "./data/app.db"

# 缓存配置
cache:
  driver: "memory"

# 限流配置
limiter:
  driver: "memory"

# 锁配置
lock:
  driver: "memory"

# 遥测配置
telemetry:
  driver: "none"

# 消息队列配置
mq:
  driver: "memory"
```

## 生成的代码使用

生成器会在指定的输出目录创建以下文件：

- `entity_container.go` - 实体容器初始化
- `repository_container.go` - 仓储容器初始化
- `service_container.go` - 服务容器初始化
- `controller_container.go` - 控制器容器初始化
- `middleware_container.go` - 中间件容器初始化
- `engine.go` - 引擎创建函数

在 `main.go` 中使用：

```go
package main

import (
    "log"

    myapp "myproject/internal/application"
)

func main() {
    // 创建引擎
    engine, err := myapp.NewEngine()
    if err != nil {
        log.Fatalf("Failed to create engine: %v", err)
    }

    // 初始化
    if err := engine.Initialize(); err != nil {
        log.Fatalf("Failed to initialize engine: %v", err)
    }

    // 启动
    if err := engine.Start(); err != nil {
        log.Fatalf("Failed to start engine: %v", err)
    }

    // 等待关闭信号
    engine.WaitForShutdown()
}
```

## 注意事项

1. 生成的代码会覆盖同名文件，请确保不要手动修改生成的文件
2. 配置文件路径必须是相对于项目根目录的路径
3. Manager 组件会根据配置文件中的 `driver` 字段自动选择对应的实现
