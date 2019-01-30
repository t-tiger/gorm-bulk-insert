package gormbulk

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
	"time"
)

type fakeRelationDB struct{}

type fakeDB struct {
	ID        int
	Name      string
	Email     string           `gorm:"default:default@mail.com"`
	Relation  *fakeRelationDB  `gorm:"foreignkey:RelationID"`
	Relations []fakeRelationDB `gorm:"foreignkey:UserRefer"`
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

	value := fakeDB{
		Name:     "name1",
		Email:    "test1@test.com",
		Relation: &fakeRelationDB{},
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
