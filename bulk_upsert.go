package gormbulk

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
)

// BulkUpsert multiple records at once
// [objects]        Must be a slice of struct
// [chunkSize]      Number of records to upsert at once.
//                  Embedding a large number of variables at once will raise an error beyond the limit of prepared statement.
//                  Larger size will normally lead the better performance, but 2000 to 3000 is reasonable.
// [excludeColumns] Columns you want to exclude from insert. You can omit if there is no column you want to exclude.
func BulkUpsert(db *gorm.DB, objects []interface{}, chunkSize int, excludeColumns ...string) error {
	// Split records with specified size not to exceed Database parameter limit
	for _, objSet := range splitObjects(objects, chunkSize) {
		if err := upsertObjSet(db, objSet, excludeColumns...); err != nil {
			return err
		}
	}
	return nil
}

func upsertObjSet(db *gorm.DB, objects []interface{}, excludeColumns ...string) error {
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
	updateValues := make([]string, 0, attrSize)

	for _, key := range sortedKeys(firstAttrs) {
		dbColumns = append(dbColumns, mainScope.Quote(gorm.ToColumnName(key)))
	}
	for _, column := range dbColumns {
		// ignore createAt if update rows
		if strings.ToLower(column) == "`CreatedAt`" {
			continue
		}
		updateValues = append(updateValues, fmt.Sprintf("%s = Values(%s)", column, column))
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

	mainScope.Raw(fmt.Sprintf("INSERT INTO %s (%s) VALUES %s ON DUPLICATE KEY UPDATE %s",
		mainScope.QuotedTableName(),
		strings.Join(dbColumns, ", "),
		strings.Join(placeholders, ", "),
		strings.Join(updateValues, ", "),
	))

	return db.Exec(mainScope.SQL, mainScope.SQLVars...).Error
}
