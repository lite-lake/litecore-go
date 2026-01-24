# LiteCore MessageBoard

基于 litecore-go 框架开发的留言板示例应用，演示框架的完整开发流程和核心功能。

## 项目特性

- ✅ 清晰的 5 层分层架构（Entity → Repository → Service → Controller → Middleware）
- ✅ 内置组件自动初始化（Config、Database、Cache、Logger、Telemetry、Limiter、Lock、MQ）
- ✅ 声明式依赖注入容器（通过 `inject:""` 标签自动注入）
- ✅ 留言审核机制（待审核/已通过/已拒绝）
- ✅ 管理员认证与会话管理（Session 存储在缓存中）
- ✅ 基于限流器的请求限流保护（每 IP 每分钟 100 次请求）
- ✅ 完整的中间件链（恢复、日志、CORS、安全头、限流、遥测、认证）
- ✅ SQLite 数据库存储 + GORM ORM
- ✅ Ristretto 高性能内存缓存
- ✅ Gin 格式化日志输出
- ✅ Bootstrap 5 + jQuery 3 前端界面

## 技术栈

- **框架**: litecore-go
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: SQLite
- **缓存**: Ristretto (高性能内存缓存)
- **日志**: Zap (Gin 格式)
- **前端**: Bootstrap 5 + jQuery 3
- **密码加密**: bcrypt

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
请输入管理员密码: ********
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
│   ├── generate/               # 代码生成入口（CLI 工具）
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
│   ├── controllers/            # 控制器层（处理 HTTP 请求）
│   ├── middlewares/            # 中间件层（封装框架中间件）
│   ├── dtos/                   # 数据传输对象
│   ├── entities/               # 实体层（数据模型）
│   ├── repositories/           # 仓储层（数据访问）
│   └── services/               # 服务层（业务逻辑）
├── static/                     # 静态资源
│   ├── css/
│   └── js/
├── templates/                  # HTML 模板
├── data/                       # 数据目录
└── README.md
```

## 核心架构

### 5 层分层架构

```
┌─────────────────────────────────────────────────────────────┐
│                    Middlewares（中间件层）                     │
│    Recovery → Logger → CORS → Security → RateLimit → Auth    │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                  Controllers（控制器层）                      │
│           接收 HTTP 请求，参数验证，调用 Service               │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                   Services（服务层）                          │
│              实现业务逻辑，协调多个 Repository                 │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                Repositories（仓储层）                         │
│              数据访问层，封装数据库操作（GORM）                 │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                   Entities（实体层）                          │
│                  数据模型定义，实现 BaseEntity 接口             │
└─────────────────────────────────────────────────────────────┘
```

### 依赖注入容器

项目使用声明式依赖注入，通过 `inject:""` 标签自动注入依赖：

```go
type MessageRepository struct {
    // 内置组件（引擎自动注入）
    Config  configmgr.IConfigManager     `inject:""`  // 配置管理器
    Manager databasemgr.IDatabaseManager `inject:""`  // 数据库管理器
}

type MessageService struct {
    Config     configmgr.IConfigManager        `inject:""`  // 配置管理器
    Repository repositories.IMessageRepository `inject:""`  // 留言仓储
    LoggerMgr  loggermgr.ILoggerManager        `inject:""`  // 日志管理器
}
```

### 框架内置组件

由引擎自动初始化和注入的内置组件：

| 组件 | 接口 | 说明 |
|------|------|------|
| Config Manager | `configmgr.IConfigManager` | 配置管理器，支持 YAML 配置 |
| Database Manager | `databasemgr.IDatabaseManager` | 数据库管理器，支持 MySQL、PostgreSQL、SQLite |
| Cache Manager | `cachemgr.ICacheManager` | 缓存管理器，支持 Redis、Memory（Ristretto）、None |
| Logger Manager | `loggermgr.ILoggerManager` | 日志管理器，支持 Gin、JSON、Default 格式 |
| Telemetry Manager | `telemetrymgr.ITelemetryManager` | 遥测管理器，支持 OTel |
| Limiter Manager | `limitermgr.ILimiterManager` | 限流管理器，支持 Redis、Memory |
| Lock Manager | `lockmgr.ILockManager` | 锁管理器，支持 Redis、Memory |
| MQ Manager | `mqmgr.IMQManager` | 消息队列管理器，支持 RabbitMQ、Memory |

### 中间件链

所有中间件按 Order 值顺序执行（值越小越先执行）：

| 中间件 | Order | 说明 |
|--------|-------|------|
| RecoveryMiddleware | 0 | panic 恢复中间件 |
| RequestLoggerMiddleware | 50 | 请求日志中间件 |
| CORSMiddleware | 100 | CORS 跨域中间件 |
| SecurityHeadersMiddleware | 150 | 安全头中间件 |
| RateLimiterMiddleware | 200 | 限流中间件（基于 IP，每分钟 100 次） |
| TelemetryMiddleware | 250 | 遥测中间件 |
| AuthMiddleware | 300 | 认证中间件（自定义，保护 /api/admin 路径） |

## 功能模块

### 1. 留言管理（用户端）

**提交留言**
- 昵称：2-20 个字符
- 内容：5-500 个字符
- 初始状态：pending（待审核）

**查看留言**
- 只显示已审核通过（approved）的留言
- 按创建时间倒序排列
- 隐藏状态信息

### 2. 留言管理（管理端）

**查看所有留言**
- 显示所有状态的留言（pending、approved、rejected）
- 按创建时间倒序排列
- 显示状态信息

**审核留言**
- 通过审核：状态变更为 approved
- 拒绝审核：状态变更为 rejected
- 重新审核：状态变更为 pending

**删除留言**
- 永久删除指定留言

**统计信息**
- 待审核数量
- 已通过数量
- 已拒绝数量
- 总留言数

### 3. 管理员认证

**密码加密**
- 使用 bcrypt 算法加密
- 成本因子：10
- 加密密码存储在配置文件中

**会话管理**
- 登录成功后生成 UUID token
- token 存储在缓存中（session:{token}）
- 默认超时时间：3600 秒（1小时）
- 支持自动续期（每次验证会话时延长）

**认证方式**
- Bearer Token 认证
- Authorization: Bearer {token}
- 保护 /api/admin 路径（登录接口除外）

### 4. 安全特性

**请求限流**
- 默认策略：每 IP 每分钟最多 100 次请求
- 使用内存存储（可切换为 Redis）
- 超限返回 429 状态码
- 响应头包含限流信息：
  - X-RateLimit-Limit: 时间窗口最大请求数
  - X-RateLimit-Remaining: 剩余请求数
  - X-RateLimit-Reset: 时间窗口重置时间

**安全头**
- X-Content-Type-Options: nosniff
- X-Frame-Options: DENY
- X-XSS-Protection: 1; mode=block
- Referrer-Policy: strict-origin-when-cross-origin
- Content-Security-Policy: default-src 'self'

**参数验证**
- 使用 Gin Binding 验证请求参数
- 验证失败返回 400 状态码
- 服务层二次验证业务规则

## API 接口

### 用户端 API

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | /api/messages | 获取已审核留言列表 | ❌ |
| POST | /api/messages | 提交留言 | ❌ |
| GET | / | 首页 | ❌ |
| GET | /static/* | 静态资源 | ❌ |

**注**: 所有用户端 API 均受限流保护（默认每 IP 每分钟 100 次请求）。

### 管理端 API（需要认证）

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /api/admin/login | 管理员登录 | ❌ |
| GET | /api/admin/messages | 获取所有留言 | ✅ |
| POST | /api/admin/messages/:id/status | 更新留言状态 | ✅ |
| DELETE | /api/admin/messages/:id | 删除留言 | ✅ |
| GET | /admin.html | 管理页面 | ✅ |

**认证方式**: Bearer Token（Authorization: Bearer {token}）

### 系统接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/health | 健康检查接口 |
| GET | /api/metrics | 系统指标 |

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

**支持的其他数据库**：
- **MySQL**: 设置 `driver: "mysql"`，配置 `mysql_config.dsn`
- **PostgreSQL**: 设置 `driver: "postgresql"`，配置 `postgresql_config.dsn`

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

**切换到 Redis**：
```yaml
cache:
  driver: "redis"
  redis_config:
    host: "localhost"
    port: 6379
    password: ""
    db: 0
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

**日志格式说明**：
- **gin**: Gin 风格，竖线分隔符，适合控制台输出（默认格式）
- **json**: JSON 格式，适合日志分析和监控
- **default**: 默认 ConsoleEncoder 格式

### 限流配置
```yaml
limiter:
  driver: "memory"                        # 驱动类型：redis, memory
  memory_config:
    max_backups: 1000                     # 最大备份项数
```

**切换到 Redis**：
```yaml
limiter:
  driver: "redis"
  redis_config:
    host: "localhost"
    port: 6379
    password: ""
    db: 0
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
    Config  configmgr.IConfigManager     `inject:""`  // 配置管理器
    Manager databasemgr.IDatabaseManager `inject:""`  // 数据库管理器
}

type MessageService struct {
    Config     configmgr.IConfigManager        `inject:""`  // 配置管理器
    Repository repositories.IMessageRepository `inject:""`  // 留言仓储（应用组件，自动注入）
    LoggerMgr  loggermgr.ILoggerManager        `inject:""`  // 日志管理器
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

### 中间件配置

项目内置的中间件位于 `internal/middlewares/`，通过封装框架提供的中间件实现。

**自定义限流中间件**：

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

**自定义认证中间件**：

```go
// internal/middlewares/auth_middleware.go
type authMiddleware struct {
    AuthService services.IAuthService `inject:""`  // 认证服务
}

func (m *authMiddleware) Wrapper() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 检查 token
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }
        // 验证 token
        session, err := m.AuthService.ValidateToken(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }
        // 存入上下文
        c.Set("session", session)
        c.Next()
    }
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

**切换数据库驱动**：
```yaml
database:
  driver: "mysql"  # 从 "sqlite" 切换到 "mysql"
  mysql_config:
    dsn: "root:password@tcp(localhost:3306)/messageboard?charset=utf8mb4&parseTime=True&loc=Local"
```

框架会自动根据配置创建对应的驱动实现，无需修改代码。

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
- 支持自动续期

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

### 参数验证

- 使用 Gin Binding 验证请求参数
- 验证失败返回 400 状态码
- 服务层二次验证业务规则

## 命令行工具

### genpasswd - 密码生成工具

生成 bcrypt 加密的管理员密码：

```bash
go run cmd/genpasswd/main.go
```

输出示例：
```
请输入管理员密码: ********
加密后的密码: $2a$10$OzRRxaA.5Njv.o0d6VuHdec2190L0zSD5OA11oUfEjJruMfXhYkVK
```

### generate - 代码生成工具

自动生成容器初始化代码：

```bash
go run cmd/generate/main.go
```

生成的代码：
- `internal/application/entity_container.go`
- `internal/application/repository_container.go`
- `internal/application/service_container.go`
- `internal/application/controller_container.go`
- `internal/application/middleware_container.go`
- `internal/application/engine.go`

## 测试

运行所有测试：

```bash
go test ./...
```

运行特定包的测试：

```bash
go test ./internal/services
```

运行带覆盖率的测试：

```bash
go test -cover ./...
```

## 前端说明

### 静态资源

- CSS: `static/css/style.css`
- JavaScript: `static/js/app.js`（用户端）、`static/js/admin.js`（管理端）
- 模板: `templates/index.html`（首页）、`templates/admin.html`（管理页）

### 前端技术栈

- Bootstrap 5.3 - CSS 框架
- jQuery 3.7 - JavaScript 库
- AJAX - 异步数据交互

### 页面功能

**首页（index.html）**
- 显示已审核通过的留言列表
- 提交新留言表单
- 实时刷新留言列表

**管理页（admin.html）**
- 管理员登录表单
- 显示所有留言（包含状态）
- 审核操作（通过/拒绝）
- 删除留言
- 统计信息显示

## 许可证

MIT License
