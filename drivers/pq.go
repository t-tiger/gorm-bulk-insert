package drivers

import (
	"database/sql"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"

	"github.com/t-tiger/gorm-bulk-insert/internal"
)

type Pq struct {
	stmt *sql.Stmt
}

	func  (d *Pq) Init(mainScope *gorm.Scope, dbColumns []string) error {
		var err error
		db := mainScope.DB().CommonDB()
		d.stmt, err = db.Prepare(pq.CopyIn(mainScope.TableName(), dbColumns...))
		return err
	}

	func  (d *Pq) PrepareRow(mainScope *gorm.Scope, attrSize int, objAttrs map[string]interface{}, obj interface{}) error {
		variables := make([]interface{}, 0, attrSize)
		for _, key := range internal.SortedKeys(objAttrs) {
			// scope.AddToVars(objAttrs[key])
			variables = append(variables, objAttrs[key])
		}
		_, err := d.stmt.Exec(variables...)
		if err != nil {
			return err
		}
		return nil
	}
	func  (d *Pq) Execute(mainScope *gorm.Scope, dbColumns []string) error {
		_, err := d.stmt.Exec()
		if err != nil {
			return err
		}
		return nil
	}
func (d *Pq) Cleanup() error {
	if d.stmt != nil {
		return d.stmt.Close()
	}
	return nil
}

