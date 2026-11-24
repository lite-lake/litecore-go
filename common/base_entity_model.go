package common

import (
	"time"

	"github.com/lite-lake/litecore-go/util/id"
	"gorm.io/gorm"
)

type BaseEntityOnlyID struct {
	ID string `gorm:"type:varchar(32);primarykey" json:"id"`
}

func (b *BaseEntityOnlyID) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		newID, err := id.NewCUID2()
		if err != nil {
			return err
		}
		b.ID = newID
	}
	return nil
}

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
