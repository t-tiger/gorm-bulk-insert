package gormbulk

import (
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/t-tiger/gorm-bulk-insert/drivers"
	"github.com/t-tiger/gorm-bulk-insert/internal"
)

// Insert multiple records at once
// [objects]        Must be a slice of struct
// [chunkSize]      Number of records to insert at once.
//                  Embedding a large number of variables at once will raise an error beyond the limit of prepared statement.
//                  Larger size will normally lead the better performance, but 2000 to 3000 is reasonable.
// [excludeColumns] Columns you want to exclude from insert. You can omit if there is no column you want to exclude.
func BulkInsert(db *gorm.DB, objects []interface{}, chunkSize int, excludeColumns ...string) error {
	driver, err := loadDriver(db)
	if err != nil {
		return err
	}
	// Split records with specified size not to exceed Database parameter limit
	for _, objSet := range internal.SplitObjects(objects, chunkSize) {
		if err := insertObjSet(db, driver, objSet, excludeColumns...); err != nil {
			return err
		}
	}
	return nil
}

func loadDriver(db *gorm.DB) (drivers.Driver, error) {
	driver := (drivers.Driver)(new(drivers.Ansi))
	if val, ok := db.Get("gorm:bulk_insert_driver"); ok {
		driver, ok = val.(drivers.Driver)
		if !ok {
			return nil, errors.New("gorm:bulk_insert_driver needs to implement the gormbulk Driver interface")
		}
	}
	return driver, nil
}

func insertObjSet(db *gorm.DB, driver drivers.Driver, objects []interface{}, excludeColumns ...string) (err error) {
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

	// Replace with database column name
	dbColumns := make([]string, 0, attrSize)
	for _, key := range internal.SortedKeys(firstAttrs) {
		dbColumns = append(dbColumns, gorm.ToColumnName(key))
	}

	err = driver.Init(mainScope, dbColumns)
	defer func() {
		var original error
		if err != nil {
			original = err
		}
		err = driver.Cleanup()
		if original != nil {
			err = original
		}
	}()
	if err != nil {
		return err
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
		err = driver.PrepareRow(mainScope, attrSize, objAttrs, obj)
		if err != nil {
			return err
		}
	}
	return driver.Execute(mainScope, dbColumns)
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

		if !internal.ContainString(excludeColumns, field.Struct.Name) && field.StructField.Relationship == nil && !hasForeignKey &&
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
