# 错误诊断表

## 数据库

| 错误 | 根因 | 修正 |
|------|------|------|
| `record not found` | `First(&e, id)` 查询 string ID | `Where("id = ?", id).First(&e)` |
| `ID is empty` | 未用基类或手动设置 ID | 用 `BaseEntityWithTimestamps`，删除手动赋值 |
| `table doesn't exist` | 未实现 `TableName()` | 添加 `func (Entity) TableName() string` |
| `connection refused` | 数据库未启动 | 检查服务状态和 DSN 配置 |

## 依赖注入

| 错误 | 根因 | 修正 |
|------|------|------|
| `inject: field not found` | 缺少 `inject:""` 标签 | 添加标签，运行 `go run ./cmd/generate` |
| `circular dependency` | A 依赖 B，B 依赖 A | 提取公共逻辑到 C |
| `nil pointer dereference` | 注入失败 | 检查 `inject:""` 标签 |

## 路由

| 错误 | 根因 | 修正 |
|------|------|------|
| `404 page not found` | 路由未注册或格式错误 | 检查 `GetRouter()`，运行 `go run ./cmd/generate` |

路由格式：`/api/messages [POST]`，不是 `POST /api/messages` 或 `/api/messages`

## 配置

| 错误 | 根因 | 修正 |
|------|------|------|
| `config file not found` | 配置文件不存在 | 确保 `configs/config.yaml` 存在 |
| `yaml: unmarshal errors` | YAML 格式错误 | 检查缩进（用空格） |
| `driver not supported` | 驱动名错误 | 数据库: mysql/postgresql/sqlite/none，缓存: redis/memory/none |

## 运行时

| 错误 | 根因 | 修正 |
|------|------|------|
| `context deadline exceeded` | 连接超时 | 检查服务可用性，增加超时配置 |
| `too many connections` | 连接池耗尽 | 调整 `max_open_conns` |

## 编译

| 错误 | 根因 | 修正 |
|------|------|------|
| `undefined: xxx` | 包未导入 | 检查 import，运行 `go mod tidy` |
| `cannot use xxx as type yyy` | 类型不匹配 | 确保 ID 类型为 `string` |
| `interface missing method` | 未实现接口 | 添加 `var _ IXXX = (*xxxImpl)(nil)` 验证 |

## 快速排查

```bash
go run ./cmd/generate && go test ./... && go fmt ./... && go vet ./...
```
