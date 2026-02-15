# Entity 模板

> 位置: `internal/entities/YOUR_entity.go`

## 基类选择

| 基类 | 字段 | 场景 |
|-----|------|------|
| `BaseEntityOnlyID` | ID | 配置表 |
| `BaseEntityWithCreatedAt` | ID, CreatedAt | 日志、审计 |
| `BaseEntityWithTimestamps` | ID, CreatedAt, UpdatedAt | 业务实体 |

## 模板

```go
package entities

import "github.com/lite-lake/litecore-go/common"

type YOUR_ENTITY struct {
	common.BaseEntityWithTimestamps
	Field1 string `gorm:"type:varchar(50);not null" json:"field1"`
	Field2 string `gorm:"type:varchar(500)" json:"field2"`
}

func (e *YOUR_ENTITY) EntityName() string    { return "YOUR_ENTITY" }
func (YOUR_ENTITY) TableName() string        { return "YOUR_TABLE" }
func (e *YOUR_ENTITY) GetId() string         { return e.ID }
```
