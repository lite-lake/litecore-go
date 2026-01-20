package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockEntity struct {
	id   string
	name string
}

func (m *mockEntity) EntityName() string {
	return m.name
}

func (m *mockEntity) TableName() string {
	return "test_table"
}

func (m *mockEntity) GetId() string {
	return m.id
}

type incompleteEntity struct{}

func (i *incompleteEntity) EntityName() string {
	return "incomplete"
}

func TestIBaseEntity_基础接口实现(t *testing.T) {
	entity := &mockEntity{
		id:   "123",
		name: "TestEntity",
	}

	assert.Equal(t, "TestEntity", entity.EntityName())
	assert.Equal(t, "test_table", entity.TableName())
	assert.Equal(t, "123", entity.GetId())
}

func TestIBaseEntity_空实现(t *testing.T) {
	tests := []struct {
		name   string
		entity IBaseEntity
	}{
		{
			name:   "空实体实例",
			entity: &mockEntity{},
		},
		{
			name:   "带有空字符串ID的实体",
			entity: &mockEntity{id: "", name: "EmptyId"},
		},
		{
			name:   "带有空名称的实体",
			entity: &mockEntity{id: "456", name: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.entity.EntityName())
			assert.NotNil(t, tt.entity.TableName())
			assert.NotNil(t, tt.entity.GetId())
		})
	}
}

func TestIBaseEntity_接口组合(t *testing.T) {
	entity := &mockEntity{
		id:   "789",
		name: "CombinedEntity",
	}

	var iface IBaseEntity = entity
	assert.Equal(t, "CombinedEntity", iface.EntityName())
	assert.Equal(t, "789", iface.GetId())
}
