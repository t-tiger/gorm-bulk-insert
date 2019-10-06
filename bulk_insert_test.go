package gormbulk

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
	"time"
)

type fakeRelationTable struct{}

type fakeTable struct {
	ID        int
	Name      string
	Email     string              `gorm:"default:default@mail.com"`
	Relation  *fakeRelationTable  `gorm:"foreignkey:RelationID"`
	Relations []fakeRelationTable `gorm:"foreignkey:UserRefer"`
	Message   sql.NullString
	Publish   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func Test_extractMapValue(t *testing.T) {
	collectKeys := func(val map[string]interface{}) []string {
		keys := make([]string, 0, len(val))
		for key := range val {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		return keys
	}

	value := fakeTable{
		Name:     "name1",
		Email:    "test1@test.com",
		Relation: &fakeRelationTable{},
		Message:  sql.NullString{String: "message1", Valid: true},
		Publish:  false,
	}

	// test without excluding columns
	fullKeys := []string{"name", "email", "message", "publish", "created_at", "updated_at"}
	sort.Strings(fullKeys)

	mapVal, err := extractMapValue(value, []string{})
	assert.NoError(t, err)

	mapKeys := collectKeys(mapVal)
	assert.Equal(t, fullKeys, mapKeys)

	// test with excluding columns
	excludedVal, err := extractMapValue(value, []string{"Email", "CreatedAt"})
	assert.NoError(t, err)

	excludedKeys := collectKeys(excludedVal)
	assert.NotContains(t, excludedKeys, "email")
	assert.NotContains(t, excludedKeys, "created_at")
}

func Test_fieldIsAutoIncrement(t *testing.T) {
	type notSpecifiedTable struct {
		ID int `gorm:"column:id"`
	}
	type primaryKeyTable struct {
		ID int `gorm:"column:id;primary_key"`
	}
	type autoIncrementTable struct {
		ID int `gorm:"column:id;primary_key;auto_increment:false"`
	}

	cases := []struct {
		Value    interface{}
		Expected bool
	}{
		{notSpecifiedTable{}, true},
		{primaryKeyTable{}, true},
		{autoIncrementTable{}, false},
	}
	for _, c := range cases {
		for _, field := range (&gorm.Scope{Value: c.Value}).Fields() {
			assert.Equal(t, fieldIsAutoIncrement(field), c.Expected)
		}
	}
}
