# 依赖管理维度代码审查报告

## 一、审查概述
- 审查维度：依赖管理
- 审查日期：2026-01-25
- 审查范围：全项目
- Go 版本：1.25.3
- 项目模块：github.com/lite-lake/litecore-go

## 二、依赖亮点

### 2.1 版本管理优秀
- 主要依赖均为最新版本：Gin v1.11.0 (2025-09)、GORM v1.31.1 (2025-11)、OpenTelemetry v1.39.0 (2025-12)
- 核心框架版本一致，无版本冲突
- 所有依赖均已通过 `go mod tidy` 检查

### 2.2 依赖选择合理
- 使用业界主流框架：Gin、GORM、Redis、OpenTelemetry、Zap
- 数据库驱动全面：MySQL、PostgreSQL、SQLite
- 可观测性完善：OpenTelemetry 全链路支持
- 消息队列支持：RabbitMQ

### 2.3 模块化设计
- 5 层分层架构清晰（Entity → Repository → Service → 交互层）
- 内置管理器自动初始化
- 依赖注入机制完善

### 2.4 依赖统计
- Go 文件总数：300
- 依赖关系数量：570
- 直接依赖：20 个
- 间接依赖：125 个

## 三、发现的问题

### 3.1 高优先级问题

| 序号 | 问题描述 | 依赖包 | 严重程度 | 建议 |
|------|---------|--------|---------|------|
| 1 | 未使用的间接依赖 | cel.dev/expr v0.24.0 | 高 | 通过 grpc 间接引入，主模块不需要该包，可移除 |
| 2 | 未使用的间接依赖 | github.com/yuin/goldmark v1.4.13 | 高 | 通过 go.uber.org/mock 和 golang.org/x/tools 引入，主模块不需要，可移除 |
| 3 | 未使用的间接依赖 | gonum.org/v1/gonum v0.16.0 | 高 | 主模块不需要，可移除 |
| 4 | 未使用的间接依赖 | rsc.io/pdf v0.1.1 | 高 | 主模块不需要，可移除 |
| 5 | 长期未维护依赖 | github.com/manifoldco/promptui v0.9.0 | 高 | 3 年未更新 (2021-10-30)，考虑替换为活跃的 CLI 交互库 |
| 6 | 极旧依赖 | github.com/chzyer/readline v0.0.0-20180603132655 | 高 | 7 年未更新 (2018-06-03)，通过 promptui 引入，移除 promptui 后可自动移除 |
| 7 | 长期未维护依赖 | github.com/modern-go/concurrent v0.0.0-20180228061459 | 高 | 7 年未更新 (2018-02-28)，通过 GORM 引入，建议关注 GORM 是否能升级到新版本移除此依赖 |
| 8 | 长期未维护依赖 | github.com/jinzhu/inflection v1.0.0 | 高 | 6 年未更新 (2019-06-03)，通过 GORM 引入，建议关注 GORM 是否能升级到新版本移除此依赖 |

### 3.2 中优先级问题

| 序号 | 问题描述 | 依赖包 | 严重程度 | 建议 |
|------|---------|--------|---------|------|
| 1 | 较旧版本 | github.com/go-sql-driver/mysql v1.7.0 | 中 | 2 年未更新 (2022-12-02)，但作为间接依赖且稳定，建议保持 |
| 2 | 较旧版本 | golang.org/x/exp v0.0.0-20221208152030-732eee02a75a | 中 | 3 年未更新 (2022-12-08)，建议检查使用场景 |
| 3 | 较旧版本 | github.com/google/uuid v1.6.0 | 中 | 1 年未更新 (2024-01-23)，建议升级到 v1.7.0 |
| 4 | 间接依赖过多 | 125 个间接依赖 | 中 | 考虑精简依赖树，减少间接依赖 |
| 5 | 测试框架未直接使用 | github.com/bsm/ginkgo/v2 v2.12.0 | 中 | 通过 redis 测试引入，项目未使用 Ginkgo 框架，可考虑移除 |

### 3.3 低优先级问题

| 序号 | 问题描述 | 依赖包 | 严重程度 | 建议 |
|------|---------|--------|---------|------|
| 1 | Go 版本过新 | go 1.25.0 | 低 | 考虑使用更稳定的 1.24 版本 |
| 2 | 未使用的间接依赖 | github.com/chzyer/logex v1.1.10 | 低 | 通过 readline 引入，移除 promptui 后可自动移除 |
| 3 | 未使用的间接依赖 | github.com/chzyer/test v0.0.0-20180213035817 | 低 | 通过 readline 引入，移除 promptui 后可自动移除 |

## 四、依赖分析

### 4.1 直接依赖

| 包路径 | 版本 | 用途 | 评价 |
|--------|------|------|------|
| github.com/dgraph-io/ristretto/v2 | v2.4.0 | 内存缓存实现 | ✅ 活跃，用于 cachemgr/memory_impl.go |
| github.com/duke-git/lancet/v2 | v2.3.8 | 工具函数库 | ✅ 活跃，用于字符串处理和转换 |
| github.com/gin-gonic/gin | v1.11.0 | HTTP 框架 | ✅ 最新，业界主流 |
| github.com/go-playground/validator/v10 | v10.27.0 | 数据验证 | ✅ 最新，Gin 默认验证器 |
| github.com/google/uuid | v1.6.0 | UUID 生成 | ⚠️ 建议升级到 v1.7.0 |
| github.com/manifoldco/promptui | v0.9.0 | CLI 交互 | ❌ 3 年未更新，建议替换 |
| github.com/mattn/go-sqlite3 | v1.14.22 | SQLite 驱动 | ✅ 活跃，最新版本 |
| github.com/rabbitmq/amqp091-go | v1.10.0 | RabbitMQ 客户端 | ✅ 活跃，用于 mqmgr/rabbitmq_impl.go |
| github.com/redis/go-redis/v9 | v9.17.2 | Redis 客户端 | ✅ 活跃，用于 cachemgr/redis_impl.go |
| github.com/stretchr/testify | v1.11.1 | 测试框架 | ✅ 最新，业界主流 |
| github.com/urfave/cli/v3 | v3.6.2 | CLI 框架 | ✅ 活跃，用于 CLI 工具 |
| go.opentelemetry.io/otel/* | v1.39.0 | OpenTelemetry | ✅ 最新，用于可观测性 |
| go.uber.org/zap | v1.27.1 | 日志库 | ✅ 最新，高性能日志 |
| golang.org/x/crypto | v0.44.0 | 加密算法 | ✅ 活跃，用于密码哈希等 |
| gopkg.in/natefinch/lumberjack.v2 | v2.2.1 | 日志轮转 | ✅ 活跃，用于日志文件切割 |
| gopkg.in/yaml.v3 | v3.0.1 | YAML 解析 | ✅ 活跃，用于配置文件 |
| gorm.io/gorm | v1.31.1 | ORM 框架 | ✅ 最新，业界主流 |
| gorm.io/driver/* | v1.5.x | 数据库驱动 | ✅ 活跃，支持 MySQL/PostgreSQL/SQLite |

### 4.2 间接依赖

#### 核心间接依赖（稳定且活跃）
- github.com/bytedance/sonic v1.14.0 - 高性能 JSON 库
- github.com/goccy/go-json v0.10.2 - JSON 序列化
- github.com/google.golang.org/grpc v1.77.0 - gRPC 框架
- github.com/prometheus/client_golang v1.19.1 - Prometheus 监控
- golang.org/x/* - Go 扩展库（持续更新）

#### 需要关注的间接依赖
- github.com/jinzhu/inflection v1.0.0 (2019) - 6 年未更新
- github.com/modern-go/concurrent v0.0.0-20180228061459 (2018) - 7 年未更新
- github.com/chzyer/readline v0.0.0-20180603132655 (2018) - 7 年未更新

### 4.3 未使用依赖

| 包路径 | 版本 | 来源 | 建议 |
|--------|------|------|------|
| cel.dev/expr | v0.24.0 | google.golang.org/grpc | 主模块不需要，可移除 |
| github.com/yuin/goldmark | v1.4.13 | go.uber.org/mock, golang.org/x/tools | 主模块不需要，可移除 |
| gonum.org/v1/gonum | v0.16.0 | 未知 | 主模块不需要，可移除 |
| rsc.io/pdf | v0.1.1 | 未知 | 主模块不需要，可移除 |

### 4.4 过时依赖

| 包路径 | 版本 | 更新时间 | 来源 | 建议 |
|--------|------|----------|------|------|
| github.com/manifoldco/promptui | v0.9.0 | 2021-10-30 | 直接依赖 | 替换为 survey 或其他活跃库 |
| github.com/chzyer/readline | v0.0.0-20180603132655 | 2018-06-03 | promptui 间接依赖 | 移除 promptui 后自动移除 |
| github.com/modern-go/concurrent | v0.0.0-20180228061459 | 2018-02-28 | GORM 间接依赖 | 关注 GORM 升级 |
| github.com/jinzhu/inflection | v1.0.0 | 2019-06-03 | GORM 间接依赖 | 关注 GORM 升级 |
| github.com/go-sql-driver/mysql | v1.7.0 | 2022-12-02 | GORM 间接依赖 | 建议保持（稳定） |
| golang.org/x/exp | v0.0.0-20221208152030-732eee02a75a | 2022-12-08 | 间接依赖 | 检查使用场景 |

## 五、改进建议

### 5.1 高优先级改进

#### 1. 移除未使用的间接依赖
```bash
# 执行以下命令清理未使用的间接依赖
go mod tidy
go mod graph | grep "cel.dev/expr"
go mod graph | grep "goldmark"
go mod graph | grep "gonum"
go mod graph | grep "rsc.io/pdf"
```

**建议：**
- 对于通过 grpc 引入的 cel.dev/expr，考虑使用 `go mod edit -droprequire=cel.dev/expr`
- 对于其他未使用的间接依赖，通过 `go mod tidy` 清理

#### 2. 替换 promptui
**当前问题：**
- promptui v0.9.0 自 2021 年以来未更新
- 引入极旧的 readline 依赖（2018 年）

**建议替代方案：**
```go
// 方案 1：使用 survey（活跃维护）
import "github.com/AlecAivazis/survey/v2"

// 方案 2：使用 bubbletea（终端 UI 框架）
import "github.com/charmbracelet/bubbletea"

// 方案 3：使用 liner（活跃的 readline 替代）
import "github.com/peterh/liner"
```

**实施步骤：**
1. 选择替代方案（推荐 survey）
2. 替换 cli/scaffold/interactive.go 中的 promptui 使用
3. 运行 `go mod tidy` 清理旧依赖

#### 3. 关注 GORM 升级
**当前问题：**
- GORM 依赖的 modern-go/concurrent 和 jinzhu/inflection 长期未更新

**建议：**
- 关注 GORM 官方是否计划移除这些依赖
- 跟踪 GORM v1.32+ 版本更新情况

### 5.2 中优先级改进

#### 1. 升级 google/uuid
```bash
go get -u github.com/google/uuid@v1.7.0
go mod tidy
```

#### 2. 检查 golang.org/x/exp 使用
```bash
# 查找使用位置
grep -r "golang.org/x/exp" --include="*.go" .
# 如果未使用，移除
```

#### 3. 精简间接依赖
- 分析 125 个间接依赖，识别可移除的依赖
- 考虑使用更精简的库替代功能丰富的库

### 5.3 低优先级改进

#### 1. 调整 Go 版本
考虑从 go 1.25.0 降级到 go 1.24.0，使用更稳定的版本

#### 2. 定期依赖审计
```bash
# 每月执行一次
go list -u -m all
go get -u ./...
go mod tidy
```

#### 3. 添加依赖安全扫描
```bash
# 使用 govulncheck 检查已知漏洞
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## 六、依赖评分

| 评分项 | 得分 | 说明 |
|--------|------|------|
| 版本管理 | 8/10 | 主要依赖均为最新版本，但存在少量过时间接依赖 |
| 依赖安全 | 7/10 | 核心依赖安全，但存在 3 年未更新的 promptui 和 7 年未更新的 readline |
| 未使用依赖 | 6/10 | 存在 4 个未使用的间接依赖，可清理 |
| 模块内聚 | 9/10 | 5 层分层架构清晰，依赖注入机制完善，无循环依赖 |
| 第三方库选择 | 9/10 | 使用业界主流框架，选择合理 |

**总分：39/50**

## 七、总结

### 7.1 整体评价
litecore-go 项目的依赖管理整体表现良好，核心依赖均为最新版本，框架选择合理，架构设计清晰。主要问题集中在少量未使用的间接依赖和长期未维护的 promptui 相关依赖。

### 7.2 关键行动项
1. **立即执行：** 替换 promptui 为活跃的 CLI 交互库（如 survey）
2. **本周内：** 清理未使用的间接依赖（cel.dev/expr、goldmark、gonum、rsc.io/pdf）
3. **本月内：** 升级 google/uuid 到 v1.7.0，检查 golang.org/x/exp 使用情况
4. **长期关注：** 跟踪 GORM 升级情况，移除 modern-go/concurrent 和 jinzhu/inflection

### 7.3 最佳实践建议
1. 建立 CI/CD 流水线，定期扫描依赖漏洞
2. 每月执行依赖更新和清理
3. 对新引入的依赖进行活跃度评估
4. 使用依赖锁定机制（go.sum）确保构建一致性
5. 记录依赖使用说明，便于后续维护

---

**审查人：** 依赖管理专家
**审查日期：** 2026-01-25
**下次审查建议：** 2026-02-25
