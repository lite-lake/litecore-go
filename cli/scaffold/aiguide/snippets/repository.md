# Repository 模板

> 位置: `internal/repositories/YOUR_entity_repository.go`

## 模板

```go
package repositories

type IYOUR_ENTITYRepository interface {
	common.IBaseRepository
	Create(entity *entities.YOUR_ENTITY) error
	GetByID(id string) (*entities.YOUR_ENTITY, error)
	GetAll() ([]*entities.YOUR_ENTITY, error)
	Update(entity *entities.YOUR_ENTITY) error
	Delete(id string) error
}

type yourEntityRepositoryImpl struct {
	DBManager databasemgr.IDatabaseManager `inject:""`
}

func NewYOUR_ENTITYRepository() IYOUR_ENTITYRepository {
	return &yourEntityRepositoryImpl{}
}

func (r *yourEntityRepositoryImpl) RepositoryName() string { return "YOUR_ENTITYRepository" }

func (r *yourEntityRepositoryImpl) Create(entity *entities.YOUR_ENTITY) error {
	return r.DBManager.DB().Create(entity).Error
}

func (r *yourEntityRepositoryImpl) GetByID(id string) (*entities.YOUR_ENTITY, error) {
	var entity entities.YOUR_ENTITY
	err := r.DBManager.DB().Where("id = ?", id).First(&entity).Error
	return &entity, err
}

func (r *yourEntityRepositoryImpl) GetAll() ([]*entities.YOUR_ENTITY, error) {
	var entities []*entities.YOUR_ENTITY
	err := r.DBManager.DB().Find(&entities).Error
	return entities, err
}

func (r *yourEntityRepositoryImpl) Update(entity *entities.YOUR_ENTITY) error {
	return r.DBManager.DB().Save(entity).Error
}

func (r *yourEntityRepositoryImpl) Delete(id string) error {
	return r.DBManager.DB().Where("id = ?", id).Delete(&entities.YOUR_ENTITY{}).Error
}
```
