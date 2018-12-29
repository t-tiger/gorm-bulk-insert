package gormbulk

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
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

func TestBulkInsert(t *testing.T) {
	var db *gorm.DB
	_, mock, err := sqlmock.NewWithDSN("mock_db")
	if err != nil {
		panic("Got an unexpected error.")
	}
	db, err = gorm.Open("sqlmock", "mock_db")
	if err != nil {
		panic("Got an unexpected error.")
	}

	// error should occur when passing non struct values
	stringValues := []interface{}{"hoge", "fuga"}
	err = BulkInsert(db, stringValues, 1)
	assert.Error(t, err)

	intValues := []interface{}{0, 1}
	err = BulkInsert(db, intValues, 1)
	assert.Error(t, err)

	boolValues := []interface{}{true, false}
	err = BulkInsert(db, boolValues, 1)
	assert.Error(t, err)

	// passing struct
	fakeValues := []interface{}{
		fakeDB{
			Name: "name1", Email: "", Relation: &fakeRelationDB{},
			Message: sql.NullString{String: "message1", Valid: true}, Publish: false,
		},
		fakeDB{
			Name: "name2", Email: "test2@test.com", Relation: &fakeRelationDB{},
			Message: sql.NullString{String: "", Valid: false}, Publish: true,
		},
	}

	err = BulkInsert(db, fakeValues, 1)
	assert.NoError(t, err)

	_ = mock
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
		Name: "name1", Email: "test1@test.com", Relation: &fakeRelationDB{},
		Message: sql.NullString{String: "message1", Valid: true}, Publish: false,
	}

	// test without excluding columns
	fullKeys := []string{"name", "email", "message", "publish", "created_at", "updated_at"}
	sort.Strings(fullKeys)

	mapVal, err := extractMapValue(value, []string{})
	assert.NoError(t, err)

	mapKeys := collectKeys(mapVal)
	assert.Equal(t, fullKeys, mapKeys)

	// test with excluding columns
	excludedVal, err := extractMapValue(value, []string{"Email", "CreatdAt"})
	assert.NoError(t, err)

	excludedKeys := collectKeys(excludedVal)
	assert.NotContains(t, excludedKeys, "email")
	assert.NotContains(t, excludedKeys, "createdAt")
}
