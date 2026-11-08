# liteadmin-go

Admin Center（AC）统一管理平台的 Admin 委托认证协议包（模块 `github.com/lite-lake/liteadmin-go`）。

> 依据内部设计记录（Admin Center 统一管理平台设计方案）§6.5、§8.1 实施。与 `liteauth-go` 对称：liteauth-go 解决「用户会话」（UC OIDC），liteadmin-go 解决「管理员委托」（AC 以平台密钥 + 管理员身份头调用上游 `/api/admin/**`）。

## 目录

| 文件 | 说明 |
| ---- | ---- |
| `doc.go` | 包文档 |
| `headers.go` | `X-Lite-Admin-*` 协议头常量与 Id 格式正则（`^[0-9]{1,18}$`）——**全仓库头名的唯一权威**，禁止在其他文件手打头名 |
| `identity.go` | `DelegatedIdentity` 轻量委托身份 + `NewDelegatedSession(adminID, username)` 字段助手（Session 实体与 ContextKey 为应用级定义，由接入服务回调映射） |
| `client.go` | `UpstreamClient`：AC 侧构造/发起上游 `/api/admin/**` 委托调用（注入委托头 + `X-Lite-Gate-Trace-Id`/`traceparent`，默认超时 15s） |
| `keygen.go` | `GenerateKey()`：32 字节密码学安全随机数 → base64url（轮换时新旧并存） |
| `cmd/keygen/main.go` | CLI：`go run ./cmd/keygen -n 3` 一次生成多把 |

## 协议速查

| 请求头 | 必填 | 校验 |
| ---- | ---- | ---- |
| `X-Lite-Admin-Key` | 是 | ∈ 上游服务配置 `admin_delegation.api_keys`（多把并存支持轮换）；`crypto/subtle` 常数时间比较 |
| `X-Lite-Admin-Id` | 是 | `^[0-9]{1,18}$`（仅格式校验，不验存在性——信任边界是 key） |
| `X-Lite-Admin-Username` | 建议 | 仅透写入日志，不参与校验 |

上游服务侧的校验由 `litecore-go/component/litemiddleware.AdminDelegationMiddleware` 实现（本包被其引用以保证头名同源）；接入指引见 `docs/sops/backend/SOP-admin-center-integration.md`。

## 密钥管理约定

- 真实密钥不进 git、不进数据库，经 userdata secrets 注入服务配置；
- 轮换：新 key 追加进 `api_keys` 数组并发布 → AC 切换到新 key → 移除旧 key；
- AC 侧同一服务的 key 与上游 `admin_delegation.api_keys` 中某一把配对。

## 独立编译

```bash
cd packages/liteadmin-go
go build ./... && go vet ./... && go test ./...
```

本包零第三方依赖（仅标准库），无需 `go.work` 即可独立构建；`litecore-go` 通过 `require + replace` 引用本包（头名唯一权威的反向依赖），仓库内编译由 `lite-backend/go.work` 解析。
