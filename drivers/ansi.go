package drivers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/t-tiger/gorm-bulk-insert/internal"
)

type Ansi struct {
	placeholders []string

}
func (d *Ansi) Init(mainScope *gorm.Scope, dbColumns []string) error {
	// NOOP
	return nil
}
func (d *Ansi) PrepareRow(mainScope *gorm.Scope, attrSize int, objAttrs map[string]interface{}, obj interface{}) error {
	if d.placeholders == nil {
		d.placeholders = make([]string, 0, attrSize)
	}
	scope := mainScope.DB().NewScope(obj)

	// Append variables
	variables := make([]string, 0, attrSize)
	for _, key := range internal.SortedKeys(objAttrs) {
		scope.AddToVars(objAttrs[key])
		variables = append(variables, "?")
	}

	valueQuery := "(" + strings.Join(variables, ", ") + ")"
	d.placeholders = append(d.placeholders, valueQuery)

	// Also append variables to mainScope
	mainScope.SQLVars = append(mainScope.SQLVars, scope.SQLVars...)
	return nil
}
func (d *Ansi) Execute(scope *gorm.Scope, dbColumns []string) error {
	db := scope.DB()
	insertOption := ""
	if val, ok := db.Get("gorm:insert_option"); ok {
		strVal, ok := val.(string)
		if !ok {
			return errors.New("gorm:insert_option should be a string")
		}
		insertOption = strVal
	}
	for i := range dbColumns {
		dbColumns[i] = scope.Quote(dbColumns[i])
	}
	scope.Raw(fmt.Sprintf("INSERT INTO %s (%s) VALUES %s %s",
		scope.QuotedTableName(),
		strings.Join(dbColumns, ", "),
		strings.Join(d.placeholders, ", "),
		insertOption,
	))
	return db.Exec(scope.SQL, scope.SQLVars...).Error
}
func (d *Ansi) Cleanup() error {
	// NOOP
	return nil
}
