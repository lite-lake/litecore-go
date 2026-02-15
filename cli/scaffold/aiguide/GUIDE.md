# 完整示例：Message 模块

> 从 Entity 到 Controller 的端到端示例。

## Entity

`internal/entities/message.go`

```go
package entities

import "github.com/lite-lake/litecore-go/common"

type Message struct {
	common.BaseEntityWithTimestamps
	Nickname string `gorm:"type:varchar(50);not null" json:"nickname"`
	Content  string `gorm:"type:varchar(500);not null" json:"content"`
}

func (e *Message) EntityName() string    { return "Message" }
func (Message) TableName() string        { return "messages" }
func (e *Message) GetId() string         { return e.ID }
```

## Repository

`internal/repositories/message_repository.go`

```go
package repositories

type IMessageRepository interface {
	common.IBaseRepository
	Create(entity *entities.Message) error
	GetByID(id string) (*entities.Message, error)
	GetAll() ([]*entities.Message, error)
	Update(entity *entities.Message) error
	Delete(id string) error
}

type messageRepositoryImpl struct {
	DBManager databasemgr.IDatabaseManager `inject:""`
}

func NewMessageRepository() IMessageRepository { return &messageRepositoryImpl{} }

func (r *messageRepositoryImpl) RepositoryName() string { return "MessageRepository" }

func (r *messageRepositoryImpl) Create(entity *entities.Message) error {
	return r.DBManager.DB().Create(entity).Error
}

func (r *messageRepositoryImpl) GetByID(id string) (*entities.Message, error) {
	var entity entities.Message
	err := r.DBManager.DB().Where("id = ?", id).First(&entity).Error
	return &entity, err
}

func (r *messageRepositoryImpl) GetAll() ([]*entities.Message, error) {
	var entities []*Message
	err := r.DBManager.DB().Find(&entities).Error
	return entities, err
}

func (r *messageRepositoryImpl) Delete(id string) error {
	return r.DBManager.DB().Where("id = ?", id).Delete(&entities.Message{}).Error
}
```

## Service

`internal/services/message_service.go`

```go
package services

type IMessageService interface {
	common.IBaseService
	Create(nickname, content string) (*entities.Message, error)
	GetByID(id string) (*entities.Message, error)
	GetAll() ([]*entities.Message, error)
	Delete(id string) error
}

type messageServiceImpl struct {
	Repository repositories.IMessageRepository `inject:""`
	LoggerMgr  loggermgr.ILoggerManager        `inject:""`
}

func NewMessageService() IMessageService { return &messageServiceImpl{} }

func (s *messageServiceImpl) ServiceName() string { return "MessageService" }

func (s *messageServiceImpl) Create(nickname, content string) (*entities.Message, error) {
	entity := &entities.Message{Nickname: nickname, Content: content}
	if err := s.Repository.Create(entity); err != nil {
		s.LoggerMgr.Ins().Error("创建失败", "error", err)
		return nil, fmt.Errorf("创建失败: %w", err)
	}
	return entity, nil
}

func (s *messageServiceImpl) GetByID(id string) (*entities.Message, error) {
	return s.Repository.GetByID(id)
}

func (s *messageServiceImpl) Delete(id string) error {
	return s.Repository.Delete(id)
}
```

## Controller

`internal/controllers/message_controller.go`

```go
package controllers

type IMessageController interface { common.IBaseController }

type messageControllerImpl struct {
	MessageService services.IMessageService `inject:""`
	LoggerMgr      loggermgr.ILoggerManager `inject:""`
}

func NewMessageController() IMessageController { return &messageControllerImpl{} }

func (c *messageControllerImpl) ControllerName() string { return "MessageController" }
func (c *messageControllerImpl) GetRouter() string {
	return "/api/messages [POST],/api/messages [GET],/api/messages/:id [GET],/api/messages/:id [DELETE]"
}

func (c *messageControllerImpl) Handle(ctx *gin.Context) {
	switch ctx.Request.Method {
	case "POST":
		c.handleCreate(ctx)
	case "GET":
		if ctx.Param("id") != "" { c.handleGet(ctx) } else { c.handleList(ctx) }
	case "DELETE":
		c.handleDelete(ctx)
	}
}

func (c *messageControllerImpl) handleCreate(ctx *gin.Context) {
	var req struct{ Nickname, Content string }
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}
	entity, err := c.MessageService.Create(req.Nickname, req.Content)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"data": entity})
}

func (c *messageControllerImpl) handleGet(ctx *gin.Context) {
	entity, err := c.MessageService.GetByID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(404, gin.H{"error": "未找到"})
		return
	}
	ctx.JSON(200, gin.H{"data": entity})
}

func (c *messageControllerImpl) handleDelete(ctx *gin.Context) {
	if err := c.MessageService.Delete(ctx.Param("id")); err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"message": "删除成功"})
}
```

## 完成后

```bash
go run ./cmd/generate && go test ./... && go fmt ./... && go vet ./...
```
