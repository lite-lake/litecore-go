# 留言板应用技术需求文档（TRD）

## 1. 项目概述

### 1.1 项目名称

litecore-go MessageBoard（留言板示例应用）

### 1.2 项目背景

基于 litecore-go 框架开发的一个简单前后端分离留言板应用，用于演示框架的核心功能和使用方法。

### 1.3 项目目标

- 演示 litecore-go 框架的完整开发流程
- 提供一个可运行的学习示例
- 验证框架的分层架构设计

## 2. 技术栈

### 2.1 后端技术栈

| 组件     | 技术选型 | 版本要求 |
| -------- | -------- | -------- |
| 编程语言 | Go       | 1.25+    |
| Web框架  | Gin      | 1.11.0   |
| ORM      | GORM     | 1.31.1   |
| 数据库   | SQLite   | 1.6.0    |
| 缓存     | go-cache | 2.1.0+   |
| 配置管理 | YAML     | yaml.v3  |
| 日志管理 | Zap      | 1.27.1   |

### 2.2 前端技术栈

| 组件              | 技术选型 | 用途          |
| ----------------- | -------- | ------------- |
| HTML5             | -        | 页面结构      |
| Bootstrap         | 5.x      | UI框架        |
| jQuery            | 3.x      | DOM操作和AJAX |
| jQuery Validation | -        | 表单验证      |

## 3. 功能需求

### 3.1 用户端功能

#### 3.1.1 查看公开留言

- **功能描述**：访问首页可查看所有已审核通过的公开留言
- **API接口**：`GET /api/messages`
- **响应格式**：

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "nickname": "用户昵称",
      "content": "留言内容",
      "created_at": "2025-01-11T12:00:00Z"
    }
  ]
}
```

#### 3.1.2 提交留言

- **功能描述**：用户填写昵称和留言内容，提交后待审核
- **API接口**：`POST /api/messages`
- **请求参数**：

```json
{
  "nickname": "用户昵称",
  "content": "留言内容"
}
```

- **验证规则**：
  - 昵称：必填，2-20个字符
  - 内容：必填，5-500个字符
- **响应格式**：

```json
{
  "code": 200,
  "message": "留言提交成功，等待审核",
  "data": {
    "id": 1
  }
}
```

### 3.2 管理端功能

#### 3.2.1 管理员登录

- **功能描述**：使用配置文件中的固定密码进行身份验证
- **API接口**：`POST /api/admin/login`
- **请求参数**：

```json
{
  "password": "管理员密码"
}
```

- **响应格式**：

```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "会话令牌"
  }
}
```

#### 3.2.2 查看所有留言

- **功能描述**：管理员可查看所有留言（包括未审核）
- **API接口**：`GET /api/admin/messages`
- **请求头**：

```
Authorization: Bearer {token}
```

- **响应格式**：

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "nickname": "用户昵称",
      "content": "留言内容",
      "status": "pending",
      "created_at": "2025-01-11T12:00:00Z"
    }
  ]
}
```

#### 3.2.3 修改留言审核状态

- **功能描述**：管理员审核通过或拒绝留言
- **API接口**：`POST /api/admin/messages/:id/status`
- **请求头**：

```
Authorization: Bearer {token}
```

- **请求参数**（表单）：

```
status=approved
```

- **状态值**：
  - `pending`：待审核
  - `approved`：已通过
  - `rejected`：已拒绝
- **响应格式**：

```json
{
  "code": 200,
  "message": "状态更新成功"
}
```

#### 3.2.4 删除留言

- **功能描述**：管理员删除不当留言
- **API接口**：`POST /api/admin/messages/:id/delete`
- **请求头**：

```
Authorization: Bearer {token}
```

- **响应格式**：

```json
{
  "code": 200,
  "message": "删除成功"
}
```

## 4. 系统架构

### 4.1 分层架构

```
┌─────────────────────────────────────┐
│         Frontend Layer              │
│    (HTML + jQuery + Bootstrap)      │
└─────────────────────────────────────┘
              ↓ HTTP/JSON
┌─────────────────────────────────────┐
│         Controller Layer            │
│   (处理HTTP请求，参数验证)            │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│          Service Layer              │
│   (业务逻辑，事务管理)                │
└─────────────────────────────────────┘
              ↓           ↘
┌───────────────────┐  ┌──────────────────┐
│ Repository Layer  │  │  Cache Layer     │
│  (数据访问抽象)    │  │  (go-cache)      │
└───────────────────┘  └──────────────────┘
              ↓                    ↑
┌─────────────────────────────────────┐
│         Data Layer                  │
│      (SQLite + GORM)                │
└─────────────────────────────────────┘
```

**缓存使用场景**：

- 管理员会话令牌存储
- 可选：留言列表缓存（减少数据库查询）

### 4.2 目录结构

```
samples/messageboard/
├── cmd/
│   └── main.go                 # 应用入口
├── configs/
│   └── config.yaml             # 配置文件
├── internal/
│   ├── application/            # 应用层（依赖注入容器）
│   │   └── container.go
│   ├── controllers/            # 控制器层
│   │   ├── message_controller.go
│   │   └── admin_controller.go
│   ├── middlewares/            # 中间件层
│   │   └── auth_middleware.go  # 管理员认证中间件
│   ├── dtos/                   # 数据传输对象
│   │   ├── message_dto.go
│   │   └── response_dto.go
│   ├── entities/               # 实体层
│   │   └── message.go
│   ├── repositories/           # 仓储层
│   │   └── message_repository.go
│   └── services/               # 服务层
│       ├── message_service.go
│       ├── auth_service.go
│       └── session_service.go  # 会话管理服务
├── static/                     # 静态资源
│   ├── css/
│   ├── js/
│   └── vendor/
├── templates/                  # HTML模板
│   ├── index.html             # 用户首页
│   └── admin.html             # 管理页面
└── README.md
```

## 5. 数据模型

### 5.1 Message 实体

```go
type Message struct {
    ID        uint      `gorm:"primarykey"`
    Nickname  string    `gorm:"type:varchar(20);not null"`
    Content   string    `gorm:"type:varchar(500);not null"`
    Status    string    `gorm:"type:varchar(20);default:'pending'"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### 5.2 AdminSession 会话实体（内存缓存）

```go
// internal/services/admin_session.go
package services

// AdminSession 管理员会话信息
// 存储在缓存中，不持久化到数据库
type AdminSession struct {
    Token     string    `json:"token"`
    ExpiresAt time.Time `json:"expires_at"`
}
```

**说明**：

- `AdminSession` 不是持久化实体，仅存在于缓存中
- Token 使用 UUID 生成，确保唯一性
- 过期时间由配置文件中的 `admin.session_timeout` 控制

## 6. 配置设计

### 6.1 config.yaml 结构

```yaml
# 应用配置（自定义业务配置）
app:
  name: "litecore-messageboard"
  version: "1.0.0"
  admin:
    password: "admin123" # 管理员密码
    session_timeout: 3600 # 会话超时时间（秒）

# 服务器配置（符合 server.ServerConfig）
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug" # debug, release, test
  read_timeout: "10s"
  write_timeout: "10s"
  idle_timeout: "60s"
  enable_metrics: true
  enable_health: true
  enable_pprof: false
  enable_recovery: true
  shutdown_timeout: "30s"

# 数据库配置（符合 databasemgr.DatabaseConfig）
database:
  driver: "sqlite" # mysql, postgresql, sqlite, none
  sqlite_config:
    dsn: "./data/messageboard.db"
    pool_config:
      max_open_conns: 1 # SQLite 通常设置为 1
      max_idle_conns: 1
      conn_max_lifetime: "30s"
      conn_max_idle_time: "5m"
  observability_config:
    slow_query_threshold: "1s" # 慢查询阈值
    log_sql: false # 是否记录完整 SQL（生产环境建议关闭）
    sample_rate: 1.0 # 采样率（0.0-1.0）

# 缓存配置（符合 cachemgr.CacheConfig）
cache:
  driver: "memory" # redis, memory, none
  memory_config:
    max_size: 100 # 最大缓存大小（MB）
    max_age: "720h" # 最大缓存时间（30天）
    max_backups: 1000 # 最大备份项数
    compress: false # 是否压缩

# 日志配置（符合 loggermgr.LoggerConfig）
logger:
  driver: "zap" # zap, none
  zap_config:
    telemetry_enabled: false
    telemetry_config:
      level: "info"
    console_enabled: true
    console_config:
      level: "info"
    file_enabled: false
    file_config:
      level: "info"
      path: "./logs/messageboard.log"
      rotation:
        max_size: 100 # MB
        max_age: 30 # days
        max_backups: 10
        compress: true
```

### 6.2 配置结构说明

litecore 框架的配置严格遵循 manager 的接口设计，每个 manager 都有对应的配置结构：

**ServerConfig（server.ServerConfig）**：

- `host`: 监听地址
- `port`: 监听端口
- `mode`: 运行模式（debug/release/test）
- 各种超时配置
- 特性开关（metrics、health、pprof、recovery）

**DatabaseConfig（manager/databasemgr.DatabaseConfig）**：

- `driver`: 驱动类型
- `{driver}_config`: 对应驱动的配置
  - `sqlite_config`: SQLite 配置
  - `mysql_config`: MySQL 配置
  - `postgresql_config`: PostgreSQL 配置
- `observability_config`: 可观测性配置（慢查询、SQL日志、采样率）

**CacheConfig（manager/cachemgr.CacheConfig）**：

- `driver`: 驱动类型
- `memory_config`: 内存缓存配置
- `redis_config`: Redis 缓存配置

**LoggerConfig（manager/loggermgr.LoggerConfig）**：

- `driver`: 驱动类型
- `zap_config`: Zap 日志配置
  - `telemetry_enabled/config`: 观测日志
  - `console_enabled/config`: 控制台日志
  - `file_enabled/config`: 文件日志（支持日志轮转）

### 6.3 litecore 框架核心组件使用

**配置加载（config.NewYamlConfigProvider）**：

```go
import "com.litelake.litecore/config"

// 创建 YAML 配置提供者
configProvider, err := config.NewYamlConfigProvider("configs/config.yaml")
if err != nil {
    log.Fatal(err)
}

// 获取配置节点
appConfig, err := configProvider.Get("app")
serverConfig, err := configProvider.Get("server")
databaseConfig, err := configProvider.Get("database")
cacheConfig, err := configProvider.Get("cache")
loggerConfig, err := configProvider.Get("logger")
```

**数据库管理器（databasemgr.Build）**：

```go
import "com.litelake.litecore/manager/databasemgr"

// 解析配置
dbConfig, err := databasemgr.ParseDatabaseConfigFromMap(databaseConfig)
if err != nil {
    log.Fatal(err)
}

// 创建管理器
dbMgr, err := databasemgr.Build(dbConfig.Driver, databaseConfig)
if err != nil {
    log.Fatal(err)
}

// 使用 GORM 数据库
dbMgr.AutoMigrate(&Message{})
db := dbMgr.DB()
```

**缓存管理器（cachemgr.Build）**：

```go
import "com.litelake.litecore/manager/cachemgr"

// 解析配置
cacheConfig, err := cachemgr.ParseCacheConfigFromMap(cacheConfig)
if err != nil {
    log.Fatal(err)
}

// 创建管理器
cacheMgr, err := cachemgr.Build(cacheConfig.Driver, cacheConfig)
if err != nil {
    log.Fatal(err)
}

// 使用缓存（需要根据实际接口调整）
cacheMgr.Set("key", value, expiration)
cacheMgr.Get("key", &target)
cacheMgr.Delete("key")
```

**日志管理器（loggermgr.Build）**：

```go
import "com.litelake.litecore/manager/loggermgr"

// 解析配置
logConfig, err := loggermgr.ParseLoggerConfigFromMap(loggerConfig)
if err != nil {
    log.Fatal(err)
}

// 创建管理器
logMgr, err := loggermgr.Build(logConfig.Driver, loggerConfig)
if err != nil {
    log.Fatal(err)
}

// 使用日志
logger := logMgr.Logger("app")
logger.Info("Application started", "version", "1.0.0")
logger.Error("Database error", "error", err)
```

**服务器引擎（server.NewEngine）**：

```go
import "com.litelake.litecore/server"

// 创建服务器
engine := server.NewEngine(
    server.WithConfig(serverConfig),
    server.WithLogger(logMgr),
)

// 配置路由
engine.Router().GET("/health", func(c *gin.Context) {
    c.JSON(200, gin.H{"status": "ok"})
})

// 启动服务器
if err := engine.Start(); err != nil {
    log.Fatal(err)
}

// 优雅关闭（通常在信号处理中调用）
engine.Shutdown()
```

## 7. API设计规范

### 7.1 统一响应格式

```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

### 7.2 错误码定义

| 错误码 | 说明           |
| ------ | -------------- |
| 200    | 成功           |
| 400    | 请求参数错误   |
| 401    | 未授权         |
| 403    | 禁止访问       |
| 404    | 资源不存在     |
| 500    | 服务器内部错误 |

### 7.3 HTTP方法限制

由于公司防火墙限制，仅使用以下方法：

- `GET`：查询资源
- `POST`：创建、更新、删除资源

## 8. 安全设计

### 8.1 身份验证

- 管理员登录后生成会话令牌
- 令牌存储在 go-cache 内存缓存中，设置过期时间
- 管理接口通过 Bearer Token 验证
- 使用 UUID 生成唯一令牌

### 8.2 输入验证

- 昵称：2-20个字符，过滤特殊字符
- 内容：5-500个字符，过滤XSS
- 密码：固定配置，不存储到数据库

### 8.3 防护措施

- SQL注入防护：使用GORM参数化查询
- XSS防护：输出时转义HTML
- CSRF防护：同源策略验证

## 9. 前端设计

### 9.1 用户首页 (index.html)

**功能区域**：

1. 留言展示区（列表形式）
2. 留言提交表单
3. 页面标题和说明

**页面结构**：

```html
<div class="container">
  <h1>留言板</h1>

  <!-- 留言列表 -->
  <div id="message-list"></div>

  <!-- 提交表单 -->
  <form id="message-form">
    <input name="nickname" placeholder="昵称" />
    <textarea name="content" placeholder="留言内容"></textarea>
    <button type="submit">提交留言</button>
  </form>
</div>
```

### 9.2 管理页面 (admin.html)

**功能区域**：

1. 登录表单
2. 留言管理面板（登录后显示）
3. 批量操作按钮

**页面结构**：

```html
<div class="container">
  <!-- 未登录状态 -->
  <div id="login-panel">
    <form id="login-form">
      <input type="password" name="password" placeholder="管理员密码" />
      <button type="submit">登录</button>
    </form>
  </div>

  <!-- 已登录状态 -->
  <div id="admin-panel" style="display:none">
    <h2>留言管理</h2>
    <div id="admin-message-list"></div>
  </div>
</div>
```

### 9.3 JavaScript交互

- 使用jQuery进行DOM操作
- 使用jQuery Validation进行表单验证
- AJAX请求使用 `$.ajax()` 或 `$.get()`/`$.post()`
- 统一错误处理和提示（Bootstrap Toast）

## 10. 实施步骤

### 10.1 第一阶段：后端基础框架

1. 创建项目目录结构
2. 实现配置加载（使用 config.NewYamlConfigProvider）
3. 初始化服务器配置（server.ServerConfig）
4. 初始化数据库管理器（databasemgr.Build）
5. 初始化缓存管理器（cachemgr.Build）
6. 初始化日志管理器（loggermgr.Build）
7. 创建Message实体和数据库迁移

**配置加载示例**：

```go
// 加载 YAML 配置文件
configProvider, err := config.NewYamlConfigProvider("configs/config.yaml")
if err != nil {
    log.Fatal(err)
}

// 解析服务器配置
serverConfigMap, err := configProvider.Get("server")
if err != nil {
    log.Fatal(err)
}
serverConfig := server.DefaultServerConfig()
// 从 map 填充 serverConfig...

// 解析数据库配置
databaseConfigMap, err := configProvider.Get("database")
if err != nil {
    log.Fatal(err)
}
databaseConfig, err := databasemgr.ParseDatabaseConfigFromMap(databaseConfigMap)
if err != nil {
    log.Fatal(err)
}

// 创建数据库管理器
dbMgr, err := databasemgr.Build(databaseConfig.Driver, databaseConfigMap)
if err != nil {
    log.Fatal(err)
}

// 类似方式创建其他 manager...
```

### 10.2 第二阶段：数据访问层

1. 实现MessageRepository接口
2. 实现CRUD操作
3. 添加事务支持

### 10.3 第三阶段：业务逻辑层

1. 实现MessageService（留言CRUD）
2. 实现AuthService（管理员认证）
3. 添加业务验证规则

### 10.4 第四阶段：中间件层

1. 实现 AuthMiddleware（基于 common.BaseMiddleware）
2. 实现管理员令牌验证逻辑
3. 实现 Bearer Token 解析
4. 实现会话过期检查
5. 编写中间件单元测试

### 10.5 第五阶段：控制器层

1. 实现MessageController（用户端API）
2. 实现AdminController（管理端API）
3. 实现统一响应处理
4. 路由配置和中间件注册

### 10.7 第七阶段：前端开发

1. 创建HTML页面
2. 集成Bootstrap和jQuery
3. 实现AJAX交互
4. 添加表单验证
5. 实现用户体验优化

## 11. 测试策略

### 11.1 单元测试

- Repository层测试（使用SQLite内存数据库）
- Service层测试（Mock Repository）
- Middleware层测试（Mock Cache）
- Controller层测试（使用httptest）

### 11.2 集成测试

- API端到端测试
- 前后端集成测试

### 11.3 测试覆盖率目标

- 核心业务逻辑：≥80%
- 数据访问层：≥70%

## 12. 部署说明

### 12.1 本地运行

```bash
cd samples/messageboard
go run cmd/main.go
```

### 12.2 构建部署

```bash
go build -o messageboard cmd/main.go
./messageboard
```

### 12.3 访问地址

- 用户首页：http://localhost:8080/
- 管理页面：http://localhost:8080/admin.html
- API文档：http://localhost:8080/api/docs

## 13. 扩展性考虑

### 13.1 可扩展功能

- 留言点赞/回复
- 邮件通知
- 敏感词过滤
- 留言分页
- 管理员多用户
- 数据库迁移到MySQL/PostgreSQL

### 13.2 性能优化

- 留言列表分页查询
- 数据库索引优化
- 静态资源CDN
- API响应压缩
- HTTP/2 支持

## 14. 缓存实现

### 14.1 使用 cachemgr 管理器

```go
import "com.litelake.litecore/manager/cachemgr"

// 从配置创建缓存管理器
cacheConfigMap, err := configProvider.Get("cache")
if err != nil {
    log.Fatal(err)
}

// 解析缓存配置
cacheConfig, err := cachemgr.ParseCacheConfigFromMap(cacheConfigMap)
if err != nil {
    log.Fatal(err)
}

// 验证配置
if err := cacheConfig.Validate(); err != nil {
    log.Fatal(err)
}

// 创建缓存管理器
cacheMgr, err := cachemgr.Build(cacheConfig.Driver, cacheConfigMap)
if err != nil {
    log.Fatal(err)
}

// 使用缓存管理器
// 注意：cachemgr 接口提供的是基础缓存操作，对于会话存储需要自行封装
```

### 14.2 会话管理封装

```go
// internal/services/session_service.go
package services

import (
    "com.litelake.litecore/manager/cachemgr"
    "github.com/google/uuid"
    "time"
)

type SessionService struct {
    cacheMgr cachemgr.CacheManager
    timeout  time.Duration
}

func NewSessionService(cacheMgr cachemgr.CacheManager, timeout time.Duration) *SessionService {
    return &SessionService{
        cacheMgr: cacheMgr,
        timeout:  timeout,
    }
}

// CreateSession 创建会话并返回令牌
func (s *SessionService) CreateSession() (string, error) {
    token := uuid.New().String()
    session := &AdminSession{
        Token:     token,
        ExpiresAt: time.Now().Add(s.timeout),
    }

    // 使用缓存管理器存储会话
    // 注意：需要根据 cachemgr.CacheManager 接口调整
    if err := s.cacheMgr.Set("session:"+token, session, s.timeout); err != nil {
        return "", err
    }

    return token, nil
}

// ValidateSession 验证会话
func (s *SessionService) ValidateSession(token string) (*AdminSession, error) {
    var session AdminSession
    if err := s.cacheMgr.Get("session:"+token, &session); err != nil {
        return nil, err
    }

    // 检查是否过期
    if time.Now().After(session.ExpiresAt) {
        s.cacheMgr.Delete("session:" + token)
        return nil, errors.New("session expired")
    }

    return &session, nil
}

// DeleteSession 删除会话
func (s *SessionService) DeleteSession(token string) error {
    return s.cacheMgr.Delete("session:" + token)
}
```

## 15. 中间件实现

### 15.1 基于 BaseMiddleware 的认证中间件

```go
// internal/middlewares/auth_middleware.go
package middlewares

import (
    "com.litelake.litecore/common"
    "messageboard/internal/services"
    "github.com/gin-gonic/gin"
    "strings"
)

// AuthMiddleware 管理员认证中间件
// 基于 common.BaseMiddleware 接口实现
type AuthMiddleware struct {
    sessionService *services.SessionService
}

// NewAuthMiddleware 创建认证中间件实例
func NewAuthMiddleware(sessionService *services.SessionService) *AuthMiddleware {
    return &AuthMiddleware{
        sessionService: sessionService,
    }
}

// MiddlewareName 返回中间件名称（实现 BaseMiddleware 接口）
func (m *AuthMiddleware) MiddlewareName() string {
    return "AuthMiddleware"
}

// Order 返回中间件执行顺序（实现 BaseMiddleware 接口）
// 数值越小越先执行，认证中间件应该在路由之后、控制器之前执行
func (m *AuthMiddleware) Order() int {
    return 100
}

// Wrapper 返回 Gin 中间件函数（实现 BaseMiddleware 接口）
func (m *AuthMiddleware) Wrapper() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 获取 Authorization 请求头
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{
                "code":    401,
                "message": "未提供认证令牌",
            })
            c.Abort()
            return
        }

        // 解析 Bearer Token
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(401, gin.H{
                "code":    401,
                "message": "认证令牌格式错误",
            })
            c.Abort()
            return
        }

        token := parts[1]

        // 使用 SessionService 验证会话
        session, err := m.sessionService.ValidateSession(token)
        if err != nil {
            c.JSON(401, gin.H{
                "code":    401,
                "message": "认证令牌无效或已过期",
            })
            c.Abort()
            return
        }

        // 将会话信息存入上下文，供后续处理器使用
        c.Set("admin_session", session)
        c.Next()
    }
}

// OnStart 服务器启动时调用（实现 BaseMiddleware 接口）
func (m *AuthMiddleware) OnStart() error {
    // 初始化逻辑，如日志记录
    return nil
}

// OnStop 服务器停止时调用（实现 BaseMiddleware 接口）
func (m *AuthMiddleware) OnStop() error {
    // 清理逻辑
    return nil
}
```

### 15.2 中间件注册与使用

```go
// 在路由配置中注册中间件
func SetupRoutes(engine *gin.Engine, authMiddleware *AuthMiddleware) {
    // 用户端 API（无需认证）
    userGroup := engine.Group("/api")
    {
        userGroup.GET("/messages", messageController.GetMessages)
        userGroup.POST("/messages", messageController.CreateMessage)
    }

    // 管理端 API（需要认证）
    adminGroup := engine.Group("/api/admin")
    adminGroup.Use(authMiddleware.Wrapper())  // 应用认证中间件
    {
        adminGroup.POST("/login", adminController.Login)
        adminGroup.GET("/messages", adminController.GetAllMessages)
        adminGroup.POST("/messages/:id/status", adminController.UpdateStatus)
        adminGroup.POST("/messages/:id/delete", adminController.DeleteMessage)
    }

    // 静态文件服务
    engine.Static("/static", "./static")
    engine.LoadHTMLGlob("templates/*")

    // 页面路由
    engine.GET("/", func(c *gin.Context) {
        c.HTML(200, "index.html", nil)
    })

    engine.GET("/admin.html", func(c *gin.Context) {
        c.HTML(200, "admin.html", nil)
    })
}
```

### 15.3 依赖注入容器中的中间件注册

```go
// internal/application/container.go
package application

import (
    "com.litelake.litecore/common"
    "messageboard/internal/middlewares"
    "messageboard/internal/services"
)

type MiddlewareContainer struct {
    authMiddleware  *middlewares.AuthMiddleware
    sessionService  *services.SessionService
}

func NewMiddlewareContainer(sessionService *services.SessionService) *MiddlewareContainer {
    return &MiddlewareContainer{
        sessionService: sessionService,
    }
}

func (c *MiddlewareContainer) Register(middleware common.BaseMiddleware) {
    switch middleware.(type) {
    case *middlewares.AuthMiddleware:
        c.authMiddleware = middleware.(*middlewares.AuthMiddleware)
    }
}

func (c *MiddlewareContainer) AuthMiddleware() *middlewares.AuthMiddleware {
    return c.authMiddleware
}

// InjectAll 执行依赖注入
func (c *MiddlewareContainer) InjectAll() error {
    // 创建认证中间件并注入 SessionService
    authMiddleware := middlewares.NewAuthMiddleware(c.sessionService)
    c.Register(authMiddleware)

    // 调用中间件的 OnStart
    if err := authMiddleware.OnStart(); err != nil {
        return err
    }

    return nil
}
```

### 15.4 中间件执行流程

```
请求 → CORS中间件 → 日志中间件 → 认证中间件 → 控制器
                     (可选)         (管理路由)
```

**中间件职责划分**：

- **认证中间件**：验证管理员令牌，不涉及业务逻辑
- **控制器**：处理请求参数，调用 Service 层
- **Service 层**：实现业务逻辑和事务管理

### 15.5 中间件单元测试

```go
// internal/middlewares/auth_middleware_test.go
package middlewares

import (
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    "messageboard/internal/services"
)

// MockSessionService 是 SessionService 的模拟实现
type MockSessionService struct {
    mock.Mock
}

func (m *MockSessionService) ValidateSession(token string) (*services.AdminSession, error) {
    args := m.Called(token)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*services.AdminSession), args.Error(1)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
    gin.SetMode(gin.TestMode)

    // 创建 mock 服务
    mockSessionService := new(MockSessionService)
    testToken := "test-token-123"
    testSession := &services.AdminSession{
        Token:     testToken,
        ExpiresAt: time.Now().Add(time.Hour),
    }

    // 设置 mock 期望
    mockSessionService.On("ValidateSession", testToken).Return(testSession, nil)

    // 创建中间件
    middleware := NewAuthMiddleware(mockSessionService)

    // 创建测试路由
    router := gin.New()
    router.Use(middleware.Wrapper())
    router.GET("/test", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "success"})
    })

    // 创建测试请求
    req := httptest.NewRequest("GET", "/test", nil)
    req.Header.Set("Authorization", "Bearer "+testToken)
    w := httptest.NewRecorder()

    // 执行请求
    router.ServeHTTP(w, req)

    // 验证结果
    assert.Equal(t, http.StatusOK, w.Code)
    mockSessionService.AssertExpectations(t)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
    gin.SetMode(gin.TestMode)

    mockSessionService := new(MockSessionService)

    // 设置 mock 期望 - 返回错误
    mockSessionService.On("ValidateSession", "invalid-token").Return(nil, errors.New("invalid token"))

    middleware := NewAuthMiddleware(mockSessionService)

    router := gin.New()
    router.Use(middleware.Wrapper())
    router.GET("/test", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "success"})
    })

    req := httptest.NewRequest("GET", "/test", nil)
    req.Header.Set("Authorization", "Bearer invalid-token")
    w := httptest.NewRecorder()

    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusUnauthorized, w.Code)
    mockSessionService.AssertExpectations(t)
}

func TestAuthMiddleware_NoToken(t *testing.T) {
    gin.SetMode(gin.TestMode)

    mockSessionService := new(MockSessionService)
    middleware := NewAuthMiddleware(mockSessionService)

    router := gin.New()
    router.Use(middleware.Wrapper())
    router.GET("/test", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "success"})
    })

    req := httptest.NewRequest("GET", "/test", nil)
    w := httptest.NewRecorder()

    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusUnauthorized, w.Code)
    // 不应该调用 ValidateSession
    mockSessionService.AssertNotCalled(t, "ValidateSession")
}
```

## 16. 依赖管理

### 16.1 Go模块依赖

```go
require (
    github.com/gin-gonic/gin v1.11.0
    gorm.io/driver/sqlite v1.6.0
    gorm.io/gorm v1.31.1
    github.com/patrickmn/go-cache v2.1.0+incompatible
    gopkg.in/yaml.v3 v3.0.1
    go.uber.org/zap v1.27.1
    com.litelake.litecore // 本地框架
)
```

### 16.2 前端依赖

- Bootstrap 5.3 (CDN)
- jQuery 3.7 (CDN)
- jQuery Validation 1.21 (CDN)

## 17. 文档要求

### 17.1 代码文档

- 所有公开接口添加注释
- 复杂业务逻辑添加说明
- 关键配置项添加注释

### 17.2 用户文档

- README.md：快速开始指南
- API.md：API接口文档
- DEPLOY.md：部署指南

---

**文档版本**: 1.0
**创建日期**: 2025-01-11
**最后更新**: 2025-01-11
**作者**: litecore-go Team
