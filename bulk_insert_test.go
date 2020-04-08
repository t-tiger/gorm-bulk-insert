package gormbulk

import (
	"database/sql"
	"sort"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRelationTable struct{}

type fakeTable struct {
	ID        int `gorm:"primary_key;auto_increment"`
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

	notNow := time.Now().Add(-3600 * time.Second)

	value := fakeTable{
		Name:      "name1",
		Email:     "test1@test.com",
		Relation:  &fakeRelationTable{},
		Message:   sql.NullString{String: "message1", Valid: true},
		CreatedAt: notNow,
		Publish:   false,
	}

	// test without excluding columns
	fullKeys := []string{"name", "email", "message", "publish", "created_at", "updated_at"}
	sort.Strings(fullKeys)

	mapVal, err := extractMapValue(value, []string{})
	assert.NoError(t, err)

	// Ensure we kept the CreatedAt time
	createdAt, ok := mapVal["created_at"].(time.Time)
	require.True(t, ok)
	assert.True(t, createdAt.Before(time.Now().Add(-100*time.Second)))

	// Ensure we set default UpdatedAt time
	updatedAt, ok := mapVal["updated_at"].(time.Time)
	require.True(t, ok)
	assert.True(t, updatedAt.After(time.Now().Add(-1*time.Second)))

	mapKeys := collectKeys(mapVal)
	assert.Equal(t, fullKeys, mapKeys)

	// test with excluding columns
	excludedVal, err := extractMapValue(value, []string{"Email", "CreatedAt"})
	assert.NoError(t, err)

	excludedKeys := collectKeys(excludedVal)
	assert.NotContains(t, excludedKeys, "email")
	assert.NotContains(t, excludedKeys, "created_at")
}

func Test_insertObject(t *testing.T) {
	type Table struct {
		RegularColumn string
		Custom        string `gorm:"column:ThisIsCamelCase"`
	}

	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()

	gdb, err := gorm.Open("mysql", db)
	require.NoError(t, err)

	mock.ExpectExec(
		"INSERT INTO `tables` \\(`ThisIsCamelCase`, `regular_column`\\)",
	).WithArgs(
		"first custom", "first regular",
		"second custom", "second regular",
	).WillReturnResult(
		sqlmock.NewResult(1, 1),
	)

	err = insertObjSet(gdb, []interface{}{
		Table{
			RegularColumn: "first regular",
			Custom:        "first custom",
		},
		Table{
			RegularColumn: "second regular",
			Custom:        "second custom",
		},
	})

	require.NoError(t, err)
}

func Test_fieldIsAutoIncrement(t *testing.T) {
	type explicitSetTable struct {
		ID int `gorm:"column:id;auto_increment"`
	}
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
		{explicitSetTable{}, true},
		{notSpecifiedTable{}, false},
		{primaryKeyTable{}, false},
		{autoIncrementTable{}, false},
	}
	for _, c := range cases {
		for _, field := range (&gorm.Scope{Value: c.Value}).Fields() {
			assert.Equal(t, fieldIsAutoIncrement(field), c.Expected)
		}
	}
}

func Test_fieldIsPrimaryAndBlank(t *testing.T) {
	type notPrimaryTable struct {
		Dummy int
	}
	type primaryKeyTable struct {
		ID int `gorm:"column:id;primary_key"`
	}

	cases := []struct {
		Value    interface{}
		Expected bool
	}{
		{notPrimaryTable{Dummy: 0}, false},
		{notPrimaryTable{Dummy: 1}, false},
		{primaryKeyTable{ID: 0}, true},
		{primaryKeyTable{ID: 1}, false},
	}
	for _, c := range cases {
		for _, field := range (&gorm.Scope{Value: c.Value}).Fields() {
			assert.Equal(t, fieldIsPrimaryAndBlank(field), c.Expected)
		}
	}
}
