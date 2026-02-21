# 业务开发入口

> AI Agent 开发业务系统的唯一入口。

## 硬规则

| # | 规则 | 违反表现 |
|---|------|----------|
| 1 | ID 类型为 `string` | `cannot use xxx as type uint` |
| 2 | 查询用 `Where("id = ?", id)` | `record not found` |
| 3 | 禁止手动设置时间戳 | ID为空 / 时间不更新 |
| 4 | 依赖字段必须 `inject:""` 标签 | `nil pointer dereference` |
| 5 | 路由格式 `/path [METHOD]` | `404 page not found` |
| 6 | 改动后运行 `go run ./cmd/generate` | 新组件不生效 |

## 按需读取

| 任务 | 文档 |
|------|------|
| 不确定命名/依赖规则 | [QUICK-REF.md](QUICK-REF.md) |
| 第一次开发/看完整示例 | [GUIDE.md](GUIDE.md) |
| 写 Entity | [snippets/entity.md](snippets/entity.md) |
| 写 Repository | [snippets/repository.md](snippets/repository.md) |
| 写 Service | [snippets/service.md](snippets/service.md) |
| 写 Controller | [snippets/controller.md](snippets/controller.md) |
| 写 Middleware/Listener/Scheduler | [snippets/other.md](snippets/other.md) |
| i18n 多语言开发（--i18n 项目） | [snippets/i18n.md](snippets/i18n.md) |
| 查 Manager 接口 | [reference/managers.md](reference/managers.md) |
| 遇到报错 | [reference/errors.md](reference/errors.md) |

## 开发流程

```
Entity → Repository → Service → Controller → go run ./cmd/generate → go run ./cmd/server
```

## 验证命令

```bash
go run ./cmd/generate && go test ./... && go fmt ./... && go vet ./...
```
