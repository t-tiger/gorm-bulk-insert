/*
Package gormbulk provides a bulk-insert method using a DB instance of gorm.
This aims to shorten the overhead caused by inserting a large number of records.
*/
package gormbulk

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

// BulkInsert executes the query to insert multiple records at once.
//
// [objects] must be a slice of struct.
//
// [chunkSize] is a number of variables embedded in query. To prevent the error which occurs embedding a large number of variables at once
// and exceeds the limit of prepared statement. Larger size normally leads to better performance, in most cases 2000 to 3000 is reasonable.
//
// [excludeColumns] is column names to exclude from insert.
func BulkInsert(db *gorm.DB, objects []interface{}, chunkSize int, excludeColumns ...string) error {
	// Split records with specified size not to exceed Database parameter limit
	for _, objSet := range splitObjects(objects, chunkSize) {
		if err := insertObjSet(db, objSet, excludeColumns...); err != nil {
			return err
		}
	}
	return nil
}

// BulkInsertWithAssigningIDs executes the query to insert multiple records at once.
// it will scan the result of `returning id` or `returning *` to [returnedValue] after every insert.
// it's necessary to set "gorm:insert_option"="returning id" in *gorm.DB
//
// [returnedValue] slice of primary_key or model, must be a *[]uint(for integer), *[]string(for uuid), *[]struct(for `returning *`)
//
// [objects] must be a slice of struct.
//
// [chunkSize] is a number of variables embedded in query. To prevent the error which occurs embedding a large number of variables at once
// and exceeds the limit of prepared statement. Larger size normally leads to better performance, in most cases 2000 to 3000 is reasonable.
//
// [excludeColumns] is column names to exclude from insert.
func BulkInsertWithAssigningIDs(db *gorm.DB, returnedValue interface{}, objects []interface{}, chunkSize int, excludeColumns ...string) error {
	typ := reflect.TypeOf(returnedValue)
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Slice {
		return errors.New("returningId must be a slice ptr")
	}

	allIds := reflect.Indirect(reflect.ValueOf(returnedValue))
	typ = allIds.Type()

	// Deference value of slice
	valueTyp := typ.Elem()
	for valueTyp.Kind() == reflect.Ptr {
		valueTyp = valueTyp.Elem()
	}

	// Split records with specified size not to exceed Database parameter limit
	for _, objSet := range splitObjects(objects, chunkSize) {
		returnValueSlice := reflect.New(typ)
		var scanReturningId func(*gorm.DB) error
		switch valueTyp.Kind() {
		case reflect.Struct:
			// If user want to scan `returning *` with returnedValue=[]struct{...}
			scanReturningId = func(db *gorm.DB) error {
				return db.Scan(returnValueSlice.Interface()).Error
			}
		default:
			// If user want to scan primary key `returning pk` with returnedValue=[]struct{...}
			pk := db.NewScope(objects[0]).PrimaryKey()
			scanReturningId = func(db *gorm.DB) error {
				return db.Pluck(pk, returnValueSlice.Interface()).Error
			}
		}

		if err := insertObjSetWithCallback(db, objSet, scanReturningId, excludeColumns...); err != nil {
			return err
		}

		allIds.Set(reflect.AppendSlice(allIds, returnValueSlice.Elem()))
	}
	return nil
}

func insertObjSet(db *gorm.DB, objects []interface{}, excludeColumns ...string) error {
	return insertObjSetWithCallback(db, objects, nil, excludeColumns...)
}

func insertObjSetWithCallback(db *gorm.DB, objects []interface{}, postInsert func(*gorm.DB) error, excludeColumns ...string) error {
	if len(objects) == 0 {
		return nil
	}

	firstAttrs, err := extractMapValue(objects[0], excludeColumns)
	if err != nil {
		return err
	}

	attrSize := len(firstAttrs)

	// Scope to eventually run SQL
	mainScope := db.NewScope(objects[0])
	// Store placeholders for embedding variables
	placeholders := make([]string, 0, attrSize)

	// Replace with database column name
	dbColumns := make([]string, 0, attrSize)
	for _, key := range sortedKeys(firstAttrs) {
		dbColumns = append(dbColumns, mainScope.Quote(key))
	}

	for _, obj := range objects {
		objAttrs, err := extractMapValue(obj, excludeColumns)
		if err != nil {
			return err
		}

		// If object sizes are different, SQL statement loses consistency
		if len(objAttrs) != attrSize {
			return errors.New("attribute sizes are inconsistent")
		}

		scope := db.NewScope(obj)

		// Append variables
		variables := make([]string, 0, attrSize)
		for _, key := range sortedKeys(objAttrs) {
			scope.AddToVars(objAttrs[key])
			variables = append(variables, "?")
		}

		valueQuery := "(" + strings.Join(variables, ", ") + ")"
		placeholders = append(placeholders, valueQuery)

		// Also append variables to mainScope
		mainScope.SQLVars = append(mainScope.SQLVars, scope.SQLVars...)
	}

	insertOption := ""
	if val, ok := db.Get("gorm:insert_option"); ok {
		strVal, ok := val.(string)
		if !ok {
			return errors.New("gorm:insert_option should be a string")
		}
		insertOption = strVal
	}

	mainScope.Raw(fmt.Sprintf("INSERT INTO %s (%s) VALUES %s %s",
		mainScope.QuotedTableName(),
		strings.Join(dbColumns, ", "),
		strings.Join(placeholders, ", "),
		insertOption,
	))

	db = db.Raw(mainScope.SQL, mainScope.SQLVars...)

	if err := db.Error; err != nil {
		return err
	}

	if postInsert != nil {
		if err := postInsert(db); err != nil {
			return err
		}
	}

	return nil
}

// Obtain columns and values required for insert from interface
func extractMapValue(value interface{}, excludeColumns []string) (map[string]interface{}, error) {
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
		value = rv.Interface()
	}
	if rv.Kind() != reflect.Struct {
		return nil, errors.New("value must be kind of Struct")
	}

	var attrs = map[string]interface{}{}

	for _, field := range (&gorm.Scope{Value: value}).Fields() {
		// Exclude relational record because it's not directly contained in database columns
		_, hasForeignKey := field.TagSettingsGet("FOREIGNKEY")

		if !containString(excludeColumns, field.Struct.Name) && field.StructField.Relationship == nil && !hasForeignKey &&
			!field.IsIgnored && !fieldIsAutoIncrement(field) && !fieldIsPrimaryAndBlank(field) {
			if (field.Struct.Name == "CreatedAt" || field.Struct.Name == "UpdatedAt") && field.IsBlank {
				attrs[field.DBName] = time.Now()
			} else if field.StructField.HasDefaultValue && field.IsBlank {
				// If default value presents and field is empty, assign a default value
				if val, ok := field.TagSettingsGet("DEFAULT"); ok {
					attrs[field.DBName] = val
				} else {
					attrs[field.DBName] = field.Field.Interface()
				}
			} else {
				attrs[field.DBName] = field.Field.Interface()
			}
		}
	}
	return attrs, nil
}

func fieldIsAutoIncrement(field *gorm.Field) bool {
	if value, ok := field.TagSettingsGet("AUTO_INCREMENT"); ok {
		return strings.ToLower(value) != "false"
	}
	return false
}

func fieldIsPrimaryAndBlank(field *gorm.Field) bool {
	return field.IsPrimaryKey && field.IsBlank
}
