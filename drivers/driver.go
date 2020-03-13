package drivers

import "github.com/jinzhu/gorm"

type Driver interface {
	Init(mainScope *gorm.Scope, dbColumns []string) error
	PrepareRow(mainScope *gorm.Scope, attrSize int, objAttrs map[string]interface{}, obj interface{}) error
	Execute(mainScope *gorm.Scope, dbColumns []string) error
	Cleanup() error
}
