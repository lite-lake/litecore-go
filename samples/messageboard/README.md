# LiteCore MessageBoard

基于 litecore-go 框架开发的留言板示例应用，演示框架的完整开发流程和核心功能。

## 项目特性

- ✅ 清晰的分层架构（Entity → Repository → Service → Controller）
- ✅ 依赖注入容器（自动注入）
- ✅ 留言审核机制（待审核/已通过/已拒绝）
- ✅ 管理员认证与会话管理
- ✅ MUJI 风格的前端界面
- ✅ SQLite 数据库存储
- ✅ 内存缓存支持

## 技术栈

- **框架**: litecore-go
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: SQLite
- **缓存**: go-cache (内存缓存)
- **日志**: Zap
- **前端**: Bootstrap 5 + jQuery 3

## 快速开始

### 1. 运行应用

```bash
cd samples/messageboard
go run cmd/server/main.go
```

### 2. 访问应用

- 用户首页: http://localhost:8080/
- 管理页面: http://localhost:8080/admin.html

### 3. 管理员登录

- 默认密码: `admin123`

## 项目结构

```
samples/messageboard/
├── cmd/
│   ├── generate/               # 代码生成入口
│   │   └── main.go
│   └── server/                 # 应用入口
│       └── main.go
├── configs/
│   └── config.yaml             # 配置文件
├── internal/
│   ├── application/            # 应用容器（CLI工具自动生成）
│   │   ├── config_container.go
│   │   ├── entity_container.go
│   │   ├── manager_container.go
│   │   ├── repository_container.go
│   │   ├── service_container.go
│   │   ├── controller_container.go
│   │   ├── middleware_container.go
│   │   └── engine.go
│   ├── controllers/            # 控制器层
│   ├── middlewares/            # 中间件层
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

## API 接口

### 用户端 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/messages | 获取已审核留言列表 |
| POST | /api/messages | 提交留言 |

### 管理端 API（需要认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/admin/login | 管理员登录 |
| GET | /api/admin/messages | 获取所有留言 |
| POST | /api/admin/messages/:id/status | 更新留言状态 |
| POST | /api/admin/messages/:id/delete | 删除留言 |

## 配置说明

配置文件位于 `configs/config.yaml`：

```yaml
app:
  admin:
    password: "admin123"        # 管理员密码
    session_timeout: 3600       # 会话超时（秒）

server:
  port: 8080                    # 服务端口
  mode: "debug"                 # 运行模式

database:
  driver: "sqlite"
  sqlite_config:
    dsn: "./data/messageboard.db"

cache:
  driver: "memory"

logger:
  driver: "zap"
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

使用 `inject:"` 标签自动注入依赖：

```go
type MyService struct {
    Config     common.BaseConfigProvider  `inject:""`
    DBMgr      databasemgr.DatabaseManager `inject:""`
    Repository *repositories.MyRepository `inject:""`
}
```

详细的 CLI 工具使用说明请参考 `cli/README.md`

## 许可证

MIT License
