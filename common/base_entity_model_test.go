package common

import (
	"testing"
	"time"

	"github.com/lite-lake/litecore-go/util/id"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T, dst interface{}) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(dst)
	assert.NoError(t, err)

	return db
}

type testEntityOnlyID struct {
	ID   string `gorm:"type:varchar(32);primarykey" json:"id"`
	Name string `gorm:"type:varchar(100)"`
}

func (b *testEntityOnlyID) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		newID, err := id.NewCUID2()
		if err != nil {
			return err
		}
		b.ID = newID
	}
	return nil
}

func (t *testEntityOnlyID) TableName() string {
	return "test_entities"
}

type testEntityWithCreatedAt struct {
	ID        string    `gorm:"type:varchar(32);primarykey" json:"id"`
	CreatedAt time.Time `gorm:"type:timestamp;not null" json:"created_at"`
	Name      string    `gorm:"type:varchar(100)"`
}

func (b *testEntityWithCreatedAt) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		newID, err := id.NewCUID2()
		if err != nil {
			return err
		}
		b.ID = newID
	}
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now()
	}
	return nil
}

func (t *testEntityWithCreatedAt) TableName() string {
	return "test_entities"
}

type testEntityWithTimestamps struct {
	ID        string    `gorm:"type:varchar(32);primarykey" json:"id"`
	CreatedAt time.Time `gorm:"type:timestamp;not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;not null" json:"updated_at"`
	Name      string    `gorm:"type:varchar(100)"`
}

func (b *testEntityWithTimestamps) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		newID, err := id.NewCUID2()
		if err != nil {
			return err
		}
		b.ID = newID
	}
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now()
	}
	if b.UpdatedAt.IsZero() {
		b.UpdatedAt = time.Now()
	}
	return nil
}

func (b *testEntityWithTimestamps) BeforeUpdate(tx *gorm.DB) error {
	b.UpdatedAt = time.Now()
	return nil
}

func (t *testEntityWithTimestamps) TableName() string {
	return "test_entities"
}

func TestBaseEntityOnlyID_BeforeCreate(t *testing.T) {
	entity := &testEntityOnlyID{Name: "test"}
	db := setupTestDB(t, &testEntityOnlyID{})

	err := db.Create(entity).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, entity.ID)
	assert.Len(t, entity.ID, 25)
}

func TestBaseEntityOnlyID_BeforeCreate_WithExistingID(t *testing.T) {
	entity := &testEntityOnlyID{ID: "custom-id-123", Name: "test"}
	db := setupTestDB(t, &testEntityOnlyID{})

	err := db.Create(entity).Error
	assert.NoError(t, err)
	assert.Equal(t, "custom-id-123", entity.ID)
}

func TestBaseEntityWithCreatedAt_BeforeCreate(t *testing.T) {
	entity := &testEntityWithCreatedAt{Name: "test"}
	db := setupTestDB(t, &testEntityWithCreatedAt{})

	beforeCreate := time.Now()
	err := db.Create(entity).Error
	assert.NoError(t, err)

	assert.NotEmpty(t, entity.ID)
	assert.Len(t, entity.ID, 25)
	assert.False(t, entity.CreatedAt.IsZero())
	assert.True(t, entity.CreatedAt.After(beforeCreate) || entity.CreatedAt.Equal(beforeCreate))
}

func TestBaseEntityWithCreatedAt_BeforeCreate_WithExistingTime(t *testing.T) {
	existingTime := time.Now().Add(-1 * time.Hour)
	entity := &testEntityWithCreatedAt{
		ID:        "custom-id-123",
		CreatedAt: existingTime,
		Name:      "test",
	}
	db := setupTestDB(t, &testEntityWithCreatedAt{})

	err := db.Create(entity).Error
	assert.NoError(t, err)

	assert.Equal(t, "custom-id-123", entity.ID)
	assert.True(t, entity.CreatedAt.Equal(existingTime))
}

func TestBaseEntityWithTimestamps_BeforeCreate(t *testing.T) {
	entity := &testEntityWithTimestamps{Name: "test"}
	db := setupTestDB(t, &testEntityWithTimestamps{})

	beforeCreate := time.Now()
	err := db.Create(entity).Error
	assert.NoError(t, err)

	assert.NotEmpty(t, entity.ID)
	assert.Len(t, entity.ID, 25)
	assert.False(t, entity.CreatedAt.IsZero())
	assert.False(t, entity.UpdatedAt.IsZero())
	assert.True(t, entity.CreatedAt.After(beforeCreate) || entity.CreatedAt.Equal(beforeCreate))
	assert.True(t, entity.UpdatedAt.After(beforeCreate) || entity.UpdatedAt.Equal(beforeCreate))
}

func TestBaseEntityWithTimestamps_BeforeUpdate(t *testing.T) {
	entity := &testEntityWithTimestamps{Name: "test"}
	db := setupTestDB(t, &testEntityWithTimestamps{})

	err := db.Create(entity).Error
	assert.NoError(t, err)

	originalUpdatedAt := entity.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	entity.Name = "updated"
	err = db.Save(entity).Error
	assert.NoError(t, err)

	assert.True(t, entity.UpdatedAt.After(originalUpdatedAt))
}

func TestBaseEntityInTransaction(t *testing.T) {
	db := setupTestDB(t, &testEntityWithTimestamps{})

	err := db.Transaction(func(tx *gorm.DB) error {
		entity := &testEntityWithTimestamps{Name: "test"}
		if err := tx.Create(entity).Error; err != nil {
			return err
		}
		assert.NotEmpty(t, entity.ID)
		return nil
	})
	assert.NoError(t, err)
}

func TestBaseEntityBatchInsert(t *testing.T) {
	db := setupTestDB(t, &testEntityWithTimestamps{})

	var entities []*testEntityWithTimestamps
	for i := 0; i < 100; i++ {
		entities = append(entities, &testEntityWithTimestamps{Name: "test"})
	}

	err := db.Create(&entities).Error
	assert.NoError(t, err)

	idMap := make(map[string]bool)
	for _, entity := range entities {
		assert.NotEmpty(t, entity.ID)
		assert.Len(t, entity.ID, 25)
		assert.False(t, idMap[entity.ID])
		idMap[entity.ID] = true
		assert.False(t, entity.CreatedAt.IsZero())
		assert.False(t, entity.UpdatedAt.IsZero())
	}
}

func TestBaseEntityConcurrentCreate(t *testing.T) {
	t.Skip("跳过并发测试：SQLite :memory: 在并发场景下存在问题")
}

func BenchmarkBaseEntityCreate(b *testing.B) {
	db := setupTestDB(&testing.T{}, &testEntityWithTimestamps{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entity := &testEntityWithTimestamps{Name: "test"}
		if err := db.Create(entity).Error; err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBaseEntityBatchInsert(b *testing.B) {
	db := setupTestDB(&testing.T{}, &testEntityWithTimestamps{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var entities []*testEntityWithTimestamps
		for j := 0; j < 1000; j++ {
			entities = append(entities, &testEntityWithTimestamps{Name: "test"})
		}
		if err := db.Create(&entities).Error; err != nil {
			b.Fatal(err)
		}
	}
}
