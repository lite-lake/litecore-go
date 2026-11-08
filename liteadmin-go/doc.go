// Package liteadmin 提供 Admin Center（AC）统一管理平台的 Admin 委托认证协议实现，
// 供 AC 侧（app-ac-api）与上游业务服务共同引用，保证协议两端零漂移。
//
// 本包与 liteauth-go 对称：liteauth-go 解决「用户会话」（UC OIDC），liteadmin-go 解决
// 「管理员委托」（AC 以平台密钥 + 管理员身份头调用上游 /api/admin/**）。
//
// 组成（详见 PRD《Admin Center 统一管理平台设计方案》§6.5、§8.1）：
//
//   - headers.go：X-Lite-Admin-* 协议头常量与 Id 格式正则 —— 全仓库头名的唯一权威，
//     禁止在其他文件手打头名字符串（litecore-go/component/litemiddleware 亦从此引用）。
//   - identity.go：DelegatedIdentity 轻量委托身份结构与 NewDelegatedSession 字段助手
//     —— Session 实体与 ContextKey 均为应用级定义，本包不引用任何应用实体，
//     由接入服务的中间件回调把 DelegatedIdentity 映射为自己的会话类型。
//   - client.go：AC 侧上游客户端 UpstreamClient，构造发往上游 /api/admin/** 的
//     委托调用请求（注入委托头 + trace 头）。
//   - keygen.go：32 字节随机密钥生成工具（base64url，CLI 可用：go run ./cmd/keygen）。
//
// 本包刻意保持零第三方依赖（仅标准库），以便被 litecore-go 与各应用双向安全引用。
//
// 模块路径：github.com/lite-lake/liteadmin-go
package liteadmin
