# LiteCore MessageBoard

基于 litecore-go 框架开发的留言板示例应用，演示框架的完整开发流程和核心功能。

## 项目特性

- ✅ 清晰的 5 层分层架构（Entity → Repository → Service → Controller）
- ✅ 内置组件（Config 和 Manager 自动初始化）
- ✅ 依赖注入容器（自动注入）
- ✅ 留言审核机制（待审核/已通过/已拒绝）
- ✅ 管理员认证与会话管理
- ✅ MUJI 风格的前端界面
- ✅ SQLite 数据库存储
- ✅ Ristretto 高性能内存缓存
- ✅ 请求限流功能（基于 IP）
- ✅ 完整的中间件链（恢复、日志、CORS、安全头、限流、遥测、认证）
- ✅ Gin 格式化日志输出

## 技术栈

- **框架**: litecore-go
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: SQLite
- **缓存**: Ristretto (高性能内存缓存)
- **日志**: Zap (Gin 格式)
- **前端**: Bootstrap 5 + jQuery 3

## 快速开始

### 1. 生成管理员密码（首次使用必需）

出于安全考虑，管理员密码需要使用 bcrypt 加密后存储在配置文件中。

运行密码生成工具：

```bash
cd samples/messageboard
go run cmd/genpasswd/main.go
```

按照提示输入密码，工具会生成加密后的密码，例如：

```
加密后的密码: $2a$10$OzRRxaA.5Njv.o0d6VuHdec2190L0zSD5OA11oUfEjJruMfXhYkVK
```

将生成的加密密码复制到 `configs/config.yaml` 文件的 `app.admin.password` 字段。

### 2. 运行应用

```bash
cd samples/messageboard
go run cmd/server/main.go
```

### 3. 访问应用

- 用户首页: http://localhost:8080/
- 管理页面: http://localhost:8080/admin.html

### 4. 管理员登录

使用你在步骤1中设置的明文密码登录。

## 项目结构

```
samples/messageboard/
├── cmd/
│   ├── generate/               # 代码生成入口
│   │   └── main.go
│   ├── genpasswd/              # 管理员密码生成工具
│   │   └── main.go
│   └── server/                 # 应用入口
│       └── main.go
├── configs/
│   └── config.yaml             # 配置文件
├── internal/
│   ├── application/            # 应用容器（CLI工具自动生成）
│   │   ├── entity_container.go
│   │   ├── repository_container.go
│   │   ├── service_container.go
│   │   ├── controller_container.go
│   │   ├── middleware_container.go
│   │   └── engine.go
│   ├── controllers/            # 控制器层
│   ├── middlewares/            # 中间件层（封装框架中间件）
│   ├── dtos/                   # 数据传输对象
│   ├── entities/               # 实体层
│   ├── repositories/           # 仓储层
│   └── services/               # 服务层
├── static/                     # 静态资源
│   ├── css/
│   └── js/
├── templates/                  # HTML 模板
├── data/                       # 数据目录
└── README.md
```

### 核心组件说明

**框架内置组件（由引擎自动初始化）**：
- **Config Manager** (`manager/configmgr`): 配置管理器，支持 YAML 配置
- **Database Manager** (`manager/databasemgr`): 数据库管理器，支持 MySQL、PostgreSQL、SQLite
- **Cache Manager** (`manager/cachemgr`): 缓存管理器，支持 Redis、Memory（Ristretto）、None
- **Logger Manager** (`manager/loggermgr`): 日志管理器，支持 Gin、JSON、Default 格式
- **Telemetry Manager** (`manager/telemetrymgr`): 遥测管理器，支持 OTel
- **Limiter Manager** (`manager/limitermgr`): 限流管理器，支持 Redis、Memory
- **Lock Manager** (`manager/lockmgr`): 分布式锁管理器，支持 Redis、Memory
- **MQ Manager** (`manager/mqmgr`): 消息队列管理器，支持 RabbitMQ、Memory

**框架内置中间件**（位于 `component/litemiddleware`）：
- **RecoveryMiddleware**: panic 恢复中间件（Order: 0）
- **RequestLoggerMiddleware**: 请求日志中间件（Order: 50）
- **CORSMiddleware**: CORS 跨域中间件（Order: 100）
- **SecurityHeadersMiddleware**: 安全头中间件（Order: 150）
- **RateLimiterMiddleware**: 限流中间件（Order: 200）
- **TelemetryMiddleware**: 遥测中间件（Order: 250）
- **AuthMiddleware**: 认证中间件（Order: 300，需自定义）

所有中间件均支持通过配置自定义 `Name` 和 `Order` 属性。

## API 接口

### 用户端 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/messages | 获取已审核留言列表 |
| POST | /api/messages | 提交留言 |

**注**: 所有用户端 API 均受限流保护（默认每 IP 每分钟 100 次请求）。

### 管理端 API（需要认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/admin/login | 管理员登录 |
| GET | /api/admin/messages | 获取所有留言 |
| POST | /api/admin/messages/:id/status | 更新留言状态 |
| POST | /api/admin/messages/:id/delete | 删除留言 |

**限流说明**:
- 默认配置：每 IP 每分钟最多 100 次请求
- 限流器使用内存存储（`limiter.memory_config`）
- 可在配置文件中调整 `limiter.memory_config.max_backups` 和中间件配置

### 系统接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/health | 健康检查接口 |

## 配置说明

配置文件位于 `configs/config.yaml`：

### 应用配置
```yaml
app:
  name: "litecore-messageboard"           # 应用名称
  version: "1.0.0"                        # 应用版本
  admin:
    password: "$2a$10$..."                # 管理员密码（bcrypt加密，使用 cmd/genpasswd 生成）
    session_timeout: 3600                 # 会话超时时间（秒）
```

### 服务器配置
```yaml
server:
  host: "0.0.0.0"                         # 监听主机地址
  port: 8080                              # 监听端口
  mode: "debug"                           # 运行模式：debug, release, test
  read_timeout: "10s"                      # 读取超时时间
  write_timeout: "10s"                     # 写入超时时间
  idle_timeout: "60s"                      # 空闲超时时间
  enable_recovery: true                    # 是否启用panic恢复
  shutdown_timeout: "30s"                  # 优雅关闭超时时间
  startup_log:                            # 启动日志配置
    enabled: true                         # 是否启用启动日志
    async: true                           # 是否异步日志
    buffer: 100                           # 日志缓冲区大小
```

### 数据库配置
```yaml
database:
  driver: "sqlite"                        # 驱动类型：mysql, postgresql, sqlite, none
  sqlite_config:
    dsn: "./data/messageboard.db"         # SQLite 数据库文件路径
    pool_config:
      max_open_conns: 1                   # 最大打开连接数
      max_idle_conns: 1                   # 最大空闲连接数
      conn_max_lifetime: "30s"            # 连接最大存活时间
      conn_max_idle_time: "5m"            # 连接最大空闲时间
  observability_config:                    # 可观测性配置
    slow_query_threshold: "1s"            # 慢查询阈值
    log_sql: false                        # 是否记录完整SQL
    sample_rate: 1.0                      # 采样率（0.0-1.0）
```

### 缓存配置（基于 Ristretto）
```yaml
cache:
  driver: "memory"                        # 驱动类型：redis, memory, none
  memory_config:
    max_size: 100                         # 最大缓存大小（MB）
    max_age: "720h"                       # 最大缓存时间（30天）
    max_backups: 1000                     # 最大备份项数
    compress: false                       # 是否压缩
```

### 日志配置（Gin 格式）
```yaml
logger:
  driver: "zap"                           # 驱动类型：zap, default, none
  zap_config:
    telemetry_enabled: false              # 是否启用观测日志
    console_enabled: true                 # 是否启用控制台日志
    console_config:
      level: "info"                       # 日志级别：debug, info, warn, error, fatal
      format: "gin"                       # 格式：gin | json | default
      color: true                         # 是否启用颜色
      time_format: "2006-01-24 15:04:05.000"  # 时间格式
    file_enabled: false                   # 是否启用文件日志
```

### 限流配置
```yaml
limiter:
  driver: "memory"                        # 驱动类型：redis, memory
  memory_config:
    max_backups: 1000                     # 最大备份项数
```

### 锁配置
```yaml
lock:
  driver: "memory"                        # 驱动类型：redis, memory
  memory_config:
    max_backups: 1000                     # 最大备份项数
```

### 消息队列配置
```yaml
mq:
  driver: "memory"                        # 驱动类型：rabbitmq, memory
  memory_config:
    max_queue_size: 10000                 # 最大队列大小
    channel_buffer: 100                   # 通道缓冲区大小
```

### 遥测配置
```yaml
telemetry:
  driver: "none"                          # 驱动类型：none, otel
```

## 安全性

### 密码加密

项目使用 bcrypt 算法加密管理员密码：
- 加密成本因子: 10（默认）
- 算法: bcrypt (基于 Blowfish)

**重要**: 请勿将明文密码直接写入配置文件，必须使用 `cmd/genpasswd` 工具生成加密密码。

### Session 管理

- Session 存储在内存缓存中（基于 Ristretto）
- 默认超时时间: 3600 秒（1小时）
- 配置项: `app.admin.session_timeout`

### 请求限流

- 默认限流策略：每 IP 每分钟最多 100 次请求
- 使用内存存储（可切换为 Redis）
- 支持自定义限流策略（修改 `internal/middlewares/rate_limiter_middleware.go`）
- 限流响应头：`X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`

### 安全头

项目内置安全头中间件，自动添加以下 HTTP 响应头：
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Content-Security-Policy: default-src 'self'`

### CORS 配置

CORS 中间件支持自定义跨域配置，默认允许所有来源。可在 `internal/middlewares/cors_middleware.go` 中修改。

## 开发指南

### 代码生成

项目使用 LiteCore CLI 工具自动生成容器初始化代码：

```bash
# 重新生成容器代码（添加新组件后执行）
go run ./cmd/generate
```

生成的容器代码位于 `internal/application/`，包括各层容器的初始化文件和引擎创建函数。

### 添加新功能

1. **添加实体**: 在 `internal/entities/` 创建实体类
2. **添加仓储**: 在 `internal/repositories/` 创建仓储类
3. **添加服务**: 在 `internal/services/` 创建服务类
4. **添加控制器**: 在 `internal/controllers/` 创建控制器类
5. **生成容器**: 运行 `go run ./cmd/generate` 重新生成容器代码

### 依赖注入

使用 `inject:"` 标签自动注入依赖，Manager 组件由引擎自动注入：

```go
type MessageRepository struct {
    // 内置组件（引擎自动注入）
    Config  configmgr.IConfigManager     `inject:""`
    Manager databasemgr.IDatabaseManager `inject:""`
}
```

**可注入的内置组件**：
- `configmgr.IConfigManager`: 配置管理器
- `databasemgr.IDatabaseManager`: 数据库管理器
- `cachemgr.ICacheManager`: 缓存管理器
- `loggermgr.ILoggerManager`: 日志管理器
- `telemetrymgr.ITelemetryManager`: 遥测管理器
- `limitermgr.ILimiterManager`: 限流管理器
- `lockmgr.ILockManager`: 锁管理器
- `mqmgr.IMQManager`: 消息队列管理器

详细的 CLI 工具使用说明请参考 `cli/README.md`

### 中间件配置

项目内置的中间件位于 `internal/middlewares/`，通过封装框架提供的中间件实现。

**示例：自定义限流中间件**

```go
// internal/middlewares/rate_limiter_middleware.go
package middlewares

import (
    "time"
    "github.com/lite-lake/litecore-go/common"
    "github.com/lite-lake/litecore-go/component/litemiddleware"
)

type IRateLimiterMiddleware interface {
    common.IBaseMiddleware
}

func NewRateLimiterMiddleware() IRateLimiterMiddleware {
    limit := 100
    window := time.Minute
    keyPrefix := "ip"
    return litemiddleware.NewRateLimiterMiddleware(&litemiddleware.RateLimiterConfig{
        Limit:     &limit,      // 时间窗口内最大请求数
        Window:    &window,     // 时间窗口大小
        KeyPrefix: &keyPrefix,  // key前缀
    })
}
```

### 日志格式说明

项目使用 Gin 格式化日志输出，特点：
- 统一格式：`{时间} | {级别} | {消息} | {字段1}={值1} {字段2}={值2} ...`
- 时间固定宽度 23 字符：`2006-01-24 15:04:05.000`
- 级别固定宽度 5 字符，右对齐，带颜色
- 字段格式：`key=value`，字符串值用引号包裹

**Gin 格式输出示例**：
```
2026-01-24 15:04:05.123 | INFO  | 开始依赖注入 | count=23
2026-01-24 15:04:05.456 | WARN  | 慢查询检测 | duration=1.2s
2026-01-24 15:04:05.789 | ERROR | 数据库连接失败 | error="connection refused"
```

**请求日志示例**：
```
2026-01-24 15:04:05.123 | 200   | 1.234ms | 127.0.0.1 | GET | /api/messages
```

### 自定义 Manager 驱动

所有 Manager 组件支持多种驱动，可通过配置文件切换：

**切换缓存驱动**：
```yaml
cache:
  driver: "redis"  # 从 "memory" 切换到 "redis"
  redis_config:
    host: "localhost"
    port: 6379
    password: ""
    db: 0
```

**切换限流驱动**：
```yaml
limiter:
  driver: "redis"  # 从 "memory" 切换到 "redis"
  redis_config:
    host: "localhost"
    port: 6379
    password: ""
    db: 0
```

框架会自动根据配置创建对应的驱动实现，无需修改代码。

## 许可证

MIT License
