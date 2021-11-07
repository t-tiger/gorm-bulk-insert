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

func TestBulkInsertWithReturningValues(t *testing.T) {
	type Table struct {
		ID            uint `gorm:"primary_key;auto_increment"`
		RegularColumn string
		Custom        string `gorm:"column:ThisIsCamelCase"`
	}

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open("mysql", db)
	require.NoError(t, err)

	mock.ExpectQuery(
		"INSERT INTO `tables` \\(`ThisIsCamelCase`, `regular_column`\\)",
	).WithArgs(
		"first custom", "first regular",
		"second custom", "second regular",
	).WillReturnRows(
		sqlmock.NewRows([]string{"id", "ThisIsCamelCase", "regular_column"}).
			AddRow(1, "first custom", "first regular").
			AddRow(2, "second custom", "second regular"),
	)

	var returnedVals []Table
	obj := []interface{}{
		Table{
			RegularColumn: "first regular",
			Custom:        "first custom",
		},
		Table{
			RegularColumn: "second regular",
			Custom:        "second custom",
		},
	}

	gdb = gdb.Set("gorm_insert_option", "RETURNING id, ThisIsCamelCase, regular_column")
	err = BulkInsertWithReturningValues(gdb, obj, &returnedVals, 1000)
	require.NoError(t, err)

	expected := []Table{
		{ID: 1, RegularColumn: "first regular", Custom: "first custom"},
		{ID: 2, RegularColumn: "second regular", Custom: "second custom"},
	}
	assert.Equal(t, expected, returnedVals)
}

func TestBulkInsertWithReturningValues_InvalidTypeOfReturnedVals(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open("mysql", db)
	require.NoError(t, err)

	tests := []struct {
		name string
		vals interface{}
	} {
		{name: "not a pointer", vals: []struct{Name string}{{Name: "1"}}},
		{name: "element is not a slice", vals: &struct{Name string}{Name: "1"}},
		{name: "slice element is not a struct", vals: &[]string{"1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := BulkInsertWithReturningValues(gdb, []interface{}{}, tt.vals, 1000)
			assert.EqualError(t, err, "returnedVals must be a pointer to a slice of struct")
		})
	}
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

	_, err = insertObjSet(gdb, []interface{}{
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
