# 实体层基类升级技术需求文档 (TRD)

| 文档信息 | 内容 |
|---------|------|
| 文档编号 | TRD-20260125-entity-upgrade |
| 文档标题 | 实体层基类升级 - CUID2 ID 生成与时间戳自动填充 |
| 版本号 | v1.1 |
| 创建日期 | 2026-01-25 |
| 作者 | AI Assistant |
| 状态 | 待评审 |

---

## 1. 背景与目标

### 1.1 背景

当前 `samples/messageboard/` 项目中，实体 ID 使用 `uint` 类型，由 GORM 自增生成。存在以下问题：

- **ID 格式受限**：`uint` 类型不便分布式场景，迁移时容易产生 ID 冲突
- **不可读性**：自增 ID 暴露业务规模信息
- **无时间戳自动填充**：`CreatedAt` 和 `UpdatedAt` 需要在 Service 层手动设置，违反单一职责原则

### 1.2 目标

1. **统一 ID 生成策略**：使用 CUID2 算法生成 25 位字符串 ID，具备以下特性：
   - 时间有序：前缀包含时间戳，保证大致按时间排序
   - 高唯一性：结合时间戳和加密级随机数，碰撞概率极低
   - 可读性：仅包含小写字母和数字，便于人类识别
   - 分布式安全：无需中央协调，各节点独立生成

2. **提供 3 种基类**：满足不同业务场景需求
   - `BaseEntityOnlyID`：仅包含 ID
   - `BaseEntityWithCreatedAt`：包含 ID + 创建时间
   - `BaseEntityWithTimestamps`：包含 ID + 创建时间 + 更新时间（最常用）

3. **自动填充时间戳**：通过 GORM Hook 自动处理，无需业务代码干预

---

## 2. 技术方案

### 2.1 依赖关系

```
common.BaseEntityOnlyID (只有 ID)
    ↓ 嵌入
common.BaseEntityWithCreatedAt (ID + CreatedAt)
    ↓ 嵌入
common.BaseEntityWithTimestamps (ID + CreatedAt + UpdatedAt)
```

### 2.2 基类定义

#### 2.2.1 BaseEntityOnlyID

**适用场景**：无需时间戳的实体（如配置表、字典表）

```go
package common

import (
	"github.com/lite-lake/litecore-go/util/id"
	"gorm.io/gorm"
)

type BaseEntityOnlyID struct {
	ID string `gorm:"type:varchar(25);primarykey" json:"id"`
}

func (b *BaseEntityOnlyID) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		id, err := id.NewCUID2()
		if err != nil {
			return err
		}
		b.ID = id
	}
	return nil
}
```

#### 2.2.2 BaseEntityWithCreatedAt

**适用场景**：只需要记录创建时间的实体（如日志、审计记录）

```go
type BaseEntityWithCreatedAt struct {
	BaseEntityOnlyID
	CreatedAt time.Time `gorm:"type:timestamp;not null" json:"created_at"`
}

func (b *BaseEntityWithCreatedAt) BeforeCreate(tx *gorm.DB) error {
	if err := b.BaseEntityOnlyID.BeforeCreate(tx); err != nil {
		return err
	}
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now()
	}
	return nil
}
```

#### 2.2.3 BaseEntityWithTimestamps

**适用场景**：需要追踪创建和修改时间的实体（最常用）

```go
type BaseEntityWithTimestamps struct {
	BaseEntityWithCreatedAt
	UpdatedAt time.Time `gorm:"type:timestamp;not null" json:"updated_at"`
}

func (b *BaseEntityWithTimestamps) BeforeCreate(tx *gorm.DB) error {
	if err := b.BaseEntityWithCreatedAt.BeforeCreate(tx); err != nil {
		return err
	}
	if b.UpdatedAt.IsZero() {
		b.UpdatedAt = time.Now()
	}
	return nil
}

func (b *BaseEntityWithTimestamps) BeforeUpdate(tx *gorm.DB) error {
	b.UpdatedAt = time.Now()
	return nil
}
```

#### 2.2.4 重要注意事项

**关于 GORM Hook 继承**：
- GORM 不会自动调用嵌入结构体的 Hook
- 必须手动调用父类的 Hook 方法
- 示例：
  ```go
  func (b *BaseEntityWithTimestamps) BeforeCreate(tx *gorm.DB) error {
      // 必须先调用父类的 Hook
      if err := b.BaseEntityWithCreatedAt.BeforeCreate(tx); err != nil {
          return err
      }
      // 然后执行自己的逻辑
      if b.UpdatedAt.IsZero() {
          b.UpdatedAt = time.Now()
      }
      return nil
  }
  ```

**关于批量操作性能**：
- CUID2 生成比自增 ID 慢（约 10μs vs < 1μs）
- 批量插入 1000 条记录可能需要 10ms 的额外开销
- 如果性能是关键因素，考虑以下优化：
  1. 使用 goroutine 并发生成 ID
  2. 缓存一批预生成的 ID
  3. 对关键表保留自增 ID（不使用基类）

**关于并发安全性**：
- `crypto/rand.Read` 是并发安全的
- CUID2 生成器可以在 goroutine 中并发使用
- 已通过 `TestNewCUID2_Concurrency` 测试验证

### 2.3 使用示例

#### 示例 1：使用 BaseEntityWithTimestamps（最常用）

```go
type Message struct {
	common.BaseEntityWithTimestamps
	Nickname  string `gorm:"type:varchar(20);not null" json:"nickname"`
	Content   string `gorm:"type:varchar(500);not null" json:"content"`
	Status    string `gorm:"type:varchar(20);default:'pending'" json:"status"`
}

func (m *Message) EntityName() string {
	return "Message"
}

func (m *Message) TableName() string {
	return "messages"
}

func (m *Message) GetId() string {
	return m.ID
}
```

**Repository 层无需修改**：

```go
func (r *messageRepositoryImpl) Create(message *entities.Message) error {
	db := r.Manager.DB()
	return db.Create(message).Error  // Hook 自动填充 ID、CreatedAt、UpdatedAt
}
```

**Service 层代码简化**：

```go
func (s *messageServiceImpl) CreateMessage(nickname, content string) (*entities.Message, error) {
	// 验证逻辑...

	message := &entities.Message{
		Nickname: nickname,
		Content:  content,
		Status:   "pending",
		// 不再需要手动设置 CreatedAt 和 UpdatedAt
	}

	if err := s.Repository.Create(message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return message, nil
}
```

#### 示例 2：使用 BaseEntityWithCreatedAt（日志实体）

```go
type AuditLog struct {
	common.BaseEntityWithCreatedAt
	Action  string `gorm:"type:varchar(50);not null" json:"action"`
	Details string `gorm:"type:text" json:"details"`
}

func (a *AuditLog) EntityName() string {
	return "AuditLog"
}

func (a *AuditLog) TableName() string {
	return "audit_logs"
}

func (a *AuditLog) GetId() string {
	return a.ID
}
```

#### 示例 3：使用 BaseEntityOnlyID（配置实体）

```go
type SystemConfig struct {
	common.BaseEntityOnlyID
	Key   string `gorm:"type:varchar(100);uniqueIndex;not null" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}

func (s *SystemConfig) EntityName() string {
	return "SystemConfig"
}

func (s *SystemConfig) TableName() string {
	return "system_configs"
}

func (s *SystemConfig) GetId() string {
	return s.ID
}
```

---

## 3. 改造计划

### 3.1 实施步骤

#### 阶段 1：基础基类开发（1 天）

| 任务 | 负责人 | 预计工时 | 产出物 |
|-----|-------|---------|-------|
| 创建 `common/base_entity_model.go` | 开发 | 2h | 3 种基类代码 |
| 编写基类单元测试 | 开发 | 2h | 测试覆盖 100% |
| 补充 common/README.md 文档 | 开发 | 1h | 使用文档 |
| 代码审查与调整 | Team | 2h | 审查通过 |

#### 阶段 2：Sample 项目改造（2 天）

| 任务 | 负责人 | 预计工时 | 产出物 |
|-----|-------|---------|-------|
| 修改 `Message` 实体 | 开发 | 0.5h | 新实体定义（ID: string） |
| 修改 Repository 接口 | 开发 | 0.5h | uint ID 改为 string ID |
| 修改 Repository 实现 | 开发 | 0.5h | 适配新的 ID 类型 |
| 修改 Service 接口 | 开发 | 0.5h | uint ID 改为 string ID |
| 修改 Service 实现 | 开发 | 0.5h | 移除时间戳手动设置，更新 ID 处理 |
| 修改 DTO | 开发 | 0.25h | ID 类型改为 string |
| 更新 Controller | 开发 | 0.25h | ID 解析逻辑适配 string |
| 运行集成测试 | 开发 | 0.5h | 测试通过 |
| 更新 Sample README | 开发 | 0.25h | 文档更新 |

#### 阶段 3：CLI 模板更新（1 天）

| 任务 | 负责人 | 预计工时 | 产出物 |
|-----|-------|---------|-------|
| 更新 entity 模板（支持 3 种基类） | 开发 | 1.5h | 支持 ID 类型选择 |
| 更新 repository 模板 | 开发 | 1h | string ID |
| 更新 service 模板 | 开发 | 0.5h | string ID |
| 更新 controller 模板 | 开发 | 0.5h | string ID 解析 |
| 更新 scaffold 文档 | 开发 | 1h | 使用说明 |
| 验证模板生成 | 开发 | 0.5h | 生成测试 |

#### 阶段 4：回归测试与发布（0.5 天）

| 任务 | 负责人 | 预计工时 | 产出物 |
|-----|-------|---------|-------|
| 运行所有测试 | 开发 | 1h | 无回归 |
| 性能测试 | 开发 | 1h | 性能报告 |
| 版本发布 | Team | 0.5h | Release Note |

**总工时预估**：4.5 天（1 人）

### 3.2 影响范围

| 组件 | 影响程度 | 改造内容 |
|-----|---------|---------|
| common/ | 高 | 新增 `base_entity_model.go`，更新 README |
| samples/messageboard/ | 高 | 修改所有实体、Repository、Service、DTO，ID 类型 uint → string |
| cli/scaffold/ | 高 | 更新 entity、repository、service、controller 模板 |
| 文档 | 中 | 更新所有相关 README |

### 3.3 架构兼容性

- **依赖注入**：容器层完全兼容，`IBaseEntity` 接口 `GetId() string` 已满足要求
- **生命周期管理**：Hook 机制不依赖 Manager，完全兼容
- **GORM 配置**：无需修改 databasemgr，自动支持 Hook
- **数据库迁移**：`AutoMigrate` 自动创建 varchar(25) 类型的 ID 字段

### 3.4 CLI 模板具体改造内容

#### 3.4.1 entityTemplate 改造

**文件位置**：`cli/scaffold/templates.go:397-426`

**修改前**：
```go
const entityTemplate = `package entities

import (
	"fmt"
	"time"

	"github.com/lite-lake/litecore-go/common"
)

type Example struct {
	ID        uint      ` + "`" + `gorm:"primarykey" json:"id"` + "`" + `
	Name      string    ` + "`" + `gorm:"type:varchar(100);not null" json:"name"` + "`" + `
	CreatedAt time.Time ` + "`" + `json:"created_at"` + "`" + `
	UpdatedAt time.Time ` + "`" + `json:"updated_at"` + "`" + `
}

func (e *Example) EntityName() string {
	return "Example"
}

func (Example) TableName() string {
	return "examples"
}

func (e *Example) GetId() string {
	return fmt.Sprintf("%d", e.ID)
}

var _ common.IBaseEntity = (*Example)(nil)
`
```

**修改后**：
```go
const entityTemplate = `package entities

import (
	"github.com/lite-lake/litecore-go/common"
)

type Example struct {
	common.BaseEntityWithTimestamps
	Name string ` + "`" + `gorm:"type:varchar(100);not null" json:"name"` + "`" + `
}

func (e *Example) EntityName() string {
	return "Example"
}

func (Example) TableName() string {
	return "examples"
}

func (e *Example) GetId() string {
	return e.ID
}

var _ common.IBaseEntity = (*Example)(nil)
`
```

**改动说明**：
- 移除 `fmt` 和 `time` 导入（不再需要）
- 嵌入 `common.BaseEntityWithTimestamps`
- 移除手动定义的 `ID`、`CreatedAt`、`UpdatedAt` 字段
- `GetId()` 直接返回 `e.ID`（无需转换）

#### 3.4.2 repositoryTemplate 改造

**文件位置**：`cli/scaffold/templates.go:428-481`

**修改前**：
```go
type IExampleRepository interface {
	common.IBaseRepository
	Create(example *entities.Example) error
	GetByID(id uint) (*entities.Example, error)
}
```

**修改后**：
```go
type IExampleRepository interface {
	common.IBaseRepository
	Create(example *entities.Example) error
	GetByID(id string) (*entities.Example, error)
}
```

**实现方法修改**：
```go
func (r *exampleRepositoryImpl) GetByID(id string) (*entities.Example, error) {
	db := r.Manager.DB()
	var example entities.Example
	err := db.Where("id = ?", id).First(&example).Error
	if err != nil {
		return nil, err
	}
	return &example, nil
}
```

**改动说明**：
- 接口方法参数：`id uint` → `id string`
- 实现方法：`db.First(&example, id)` → `db.Where("id = ?", id).First(&example)`
- 原因：`db.First(&example, id)` 对 uint 有特殊处理（主键查询），string ID 需要使用 `Where` 子句

#### 3.4.3 serviceTemplate 改造

**文件位置**：`cli/scaffold/templates.go:483-547`

**修改前**：
```go
type IExampleService interface {
	common.IBaseService
	CreateExample(name string) (*entities.Example, error)
	GetExample(id uint) (*entities.Example, error)
}
```

**修改后**：
```go
type IExampleService interface {
	common.IBaseService
	CreateExample(name string) (*entities.Example, error)
	GetExample(id string) (*entities.Example, error)
}
```

**实现方法修改**：
```go
func (s *exampleServiceImpl) GetExample(id string) (*entities.Example, error) {
	return s.Repository.GetByID(id)
}
```

**改动说明**：
- 接口方法参数：`id uint` → `id string`
- 实现方法无需修改（直接调用 Repository）
- 日志中 `example.ID` 已经是 string，无需修改

#### 3.4.4 controllerTemplate 改造

**文件位置**：`cli/scaffold/templates.go:549-630`

**修改前**：
```go
func (c *exampleControllerImpl) handleGet(ctx *gin.Context) {
	id := ctx.Param("id")
	var exampleID uint
	if _, err := fmt.Sscanf(id, "%d", &exampleID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	example, err := c.ExampleService.GetExample(exampleID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "示例不存在"})
		return
	}

	ctx.JSON(http.StatusOK, example)
}
```

**修改后**：
```go
func (c *exampleControllerImpl) handleGet(ctx *gin.Context) {
	id := ctx.Param("id")

	example, err := c.ExampleService.GetExample(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "示例不存在"})
		return
	}

	ctx.JSON(http.StatusOK, example)
}
```

**改动说明**：
- 移除 `fmt` 导入
- 移除 `fmt.Sscanf` 解析逻辑
- 直接使用 string 类型的 `id` 参数
- 简化代码，减少不必要的转换

#### 3.4.5 支持多基类选择（可选）

为了支持 3 种基类选择，可以在 `cli/scaffold/interactive.go` 中添加交互式选择：

```go
func selectBaseType() string {
	questions := []*survey.Question{
		{
			Name: "base_type",
			Prompt: &survey.Select{
				Message: "选择基类类型：",
				Options: []string{
					"BaseEntityOnlyID (仅ID)",
					"BaseEntityWithCreatedAt (ID + 创建时间)",
					"BaseEntityWithTimestamps (ID + 创建时间 + 更新时间) [推荐]",
				},
				Default: "BaseEntityWithTimestamps (ID + 创建时间 + 更新时间) [推荐]",
			},
		},
	}

	answers := struct {
		BaseType string `survey:"base_type"`
	}{}

	if err := survey.Ask(questions, &answers); err != nil {
		return "BaseEntityWithTimestamps"
	}

	switch answers.BaseType {
	case "BaseEntityOnlyID (仅ID)":
		return "BaseEntityOnlyID"
	case "BaseEntityWithCreatedAt (ID + 创建时间)":
		return "BaseEntityWithCreatedAt"
	default:
		return "BaseEntityWithTimestamps"
	}
}
```

然后在 `entityTemplate` 中使用变量替换：

```go
const entityTemplate = `package entities

import (
	"github.com/lite-lake/litecore-go/common"
)

type Example struct {
	common.{{.BaseType}}
	Name string ` + "`" + `gorm:"type:varchar(100);not null" json:"name"` + "`" + `
}

func (e *Example) EntityName() string {
	return "Example"
}

func (Example) TableName() string {
	return "examples"
}

func (e *Example) GetId() string {
	return e.ID
}

var _ common.IBaseEntity = (*Example)(nil)
`
```

---

## 4. 测试计划

### 4.1 单元测试

#### 4.1.1 基类测试（`common/base_entity_model_test.go`）

| 测试用例 | 测试内容 | 预期结果 |
|---------|---------|---------|
| TestBaseEntityOnlyID_BeforeCreate | 空字符串 ID 自动生成 CUID2 | ID 不为空，长度 25 |
| TestBaseEntityOnlyID_BeforeCreate_WithExistingID | 已设置 ID 不重复生成 | ID 保持不变 |
| TestBaseEntityWithCreatedAt_BeforeCreate | 创建时间自动填充 | CreatedAt 不为零值 |
| TestBaseEntityWithCreatedAt_BeforeCreate_WithExistingTime | 已设置时间不重复填充 | CreatedAt 保持不变 |
| TestBaseEntityWithTimestamps_BeforeCreate | 创建时间和更新时间自动填充 | 两个字段都不为零值 |
| TestBaseEntityWithTimestamps_BeforeUpdate | 更新时间自动刷新 | UpdatedAt 大于旧值 |
| TestBaseEntityInTransaction | 事务中创建记录，Hook 被调用 | Hook 正常工作，事务提交后数据一致 |
| TestBaseEntityHookError | 模拟 Hook 返回错误 | 操作失败，事务回滚 |
| TestBaseEntityBatchInsert | 批量插入 100 条记录 | 所有 ID 唯一，Hook 都被调用 |
| TestBaseEntityConcurrentCreate | 并发创建记录 | 所有 ID 唯一，无冲突 |

#### 4.1.2 集成测试（`samples/messageboard/`）

| 测试用例 | 测试内容 | 预期结果 |
|---------|---------|---------|
| TestCreateMessage | 创建留言，验证 ID、时间戳自动填充 | ID 为 25 位 CUID2，时间戳正确 |
| TestUpdateMessageStatus | 更新留言状态，验证 UpdatedAt 刷新 | UpdatedAt 大于 CreatedAt |

### 4.2 性能测试

| 测试项 | 指标 | 目标值 |
|-------|------|--------|
| CUID2 生成性能 | 单次生成耗时 | < 10μs |
| BatchInsert | 批量插入 1000 条 | < 100ms |
| BatchInsert (对比自增) | 批量插入 1000 条（自增 ID） | < 10ms |
| Hook 开销 | BeforeCreate 额外耗时 | < 1μs |
| 并发创建 | 100 goroutine 并发创建 | 无冲突，无死锁 |

---

## 5. 风险评估

| 风险项 | 风险等级 | 影响 | 缓解措施 |
|-------|---------|------|---------|
| GORM Hook 调用失败导致数据插入失败 | 中 | 无法创建数据 | 添加错误处理和日志记录 |
| ID 生成器依赖 crypto/rand | 低 | 性能略低 | 监控生成性能，必要时优化 |
| 批量操作性能下降 | 中 | 批量插入变慢 | 提供性能优化建议 |
| 时间戳自动填充逻辑错误 | 低 | 时间不准确 | 完善单元测试覆盖 |

---

## 6. 验收标准

### 6.1 功能验收

- [ ] 3 种基类实现完成，代码符合 Go 规范
- [ ] 所有单元测试通过，覆盖率达到 100%
- [ ] `samples/messageboard/` 改造完成，所有测试通过
- [ ] CLI 模板支持 3 种基类生成

### 6.2 性能验收

- [ ] CUID2 生成性能符合目标值（< 10μs）
- [ ] Hook 开销可忽略不计（< 1μs）
- [ ] 批量插入 1000 条记录 < 100ms

### 6.3 文档验收

- [ ] `common/README.md` 更新完成，包含使用示例和注意事项
- [ ] `samples/messageboard/README.md` 更新完成
- [ ] CLI 模板文档更新完成

---

## 7. 附录

### 7.1 CUID2 特性说明

| 特性 | 说明 | 示例 |
|-----|------|------|
| 长度 | 25 个字符 | `k3y4j5m6n7p8q9r0s1t2u3v4` |
| 字符集 | 0-9 和 a-z（base36） | 小写字母和数字 |
| 时间有序 | 前缀包含时间戳 | 大致按创建时间排序 |
| 唯一性 | 结合时间戳和随机数 | 碰撞概率 < 10^-20 |

### 7.2 数据库字段定义

| 字段 | 类型 | 长度 | 是否必填 | 说明 |
|-----|------|------|---------|------|
| ID | varchar | 25 | 是 | CUID2 字符串 |
| CreatedAt | timestamp | - | 是 | ISO 8601 格式 |
| UpdatedAt | timestamp | - | 是 | ISO 8601 格式 |

### 7.3 代码规范

- 所有注释使用中文
- 遵循 Go 官方代码规范
- 导入顺序：标准库 → 第三方库 → 本地模块
- 使用 Tab 缩进，每行最多 120 字符

---

## 8. 变更记录

| 版本 | 日期 | 作者 | 变更内容 |
|-----|------|------|---------|
| v1.0 | 2026-01-25 | AI Assistant | 初始版本 |
| v1.1 | 2026-01-25 | AI Assistant | 更新改造计划，补充重要注意事项和测试用例 |

---

**审批意见**：

- [ ] 架构师审批
- [ ] 技术负责人审批
- [ ] 产品经理审批
