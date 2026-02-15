# Service 模板

> 位置: `internal/services/YOUR_entity_service.go`

## 模板

```go
package services

type IYOUR_ENTITYService interface {
	common.IBaseService
	Create(field1, field2 string) (*entities.YOUR_ENTITY, error)
	GetByID(id string) (*entities.YOUR_ENTITY, error)
	GetAll() ([]*entities.YOUR_ENTITY, error)
	Delete(id string) error
}

type yourEntityServiceImpl struct {
	Repository repositories.IYOUR_ENTITYRepository `inject:""`
	LoggerMgr  loggermgr.ILoggerManager            `inject:""`
}

func NewYOUR_ENTITYService() IYOUR_ENTITYService {
	return &yourEntityServiceImpl{}
}

func (s *yourEntityServiceImpl) ServiceName() string { return "YOUR_ENTITYService" }

func (s *yourEntityServiceImpl) Create(field1, field2 string) (*entities.YOUR_ENTITY, error) {
	entity := &entities.YOUR_ENTITY{Field1: field1, Field2: field2}
	if err := s.Repository.Create(entity); err != nil {
		s.LoggerMgr.Ins().Error("创建失败", "error", err)
		return nil, fmt.Errorf("创建失败: %w", err)
	}
	return entity, nil
}

func (s *yourEntityServiceImpl) GetByID(id string) (*entities.YOUR_ENTITY, error) {
	return s.Repository.GetByID(id)
}

func (s *yourEntityServiceImpl) GetAll() ([]*entities.YOUR_ENTITY, error) {
	return s.Repository.GetAll()
}

func (s *yourEntityServiceImpl) Delete(id string) error {
	return s.Repository.Delete(id)
}
```
