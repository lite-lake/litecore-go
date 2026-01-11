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
go run cmd/main.go
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
│   └── main.go                 # 应用入口
├── configs/
│   └── config.yaml             # 配置文件
├── internal/
│   ├── application/            # 应用容器
│   │   └── container.go
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

### 添加新功能

1. **添加实体**: 在 `internal/entities/` 创建实体类
2. **添加仓储**: 在 `internal/repositories/` 创建仓储类
3. **添加服务**: 在 `internal/services/` 创建服务类
4. **添加控制器**: 在 `internal/controllers/` 创建控制器类
5. **注册组件**: 在 `internal/application/container.go` 中注册

### 依赖注入

使用 `inject:"` 标签自动注入依赖：

```go
type MyService struct {
    Config     common.BaseConfigProvider  `inject:""`
    DBMgr      databasemgr.DatabaseManager `inject:""`
    Repository *repositories.MyRepository `inject:""`
}
```

## 许可证

MIT License
