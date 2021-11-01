# Gorm Bulk Insert

`Gorm Bulk Insert` is a library to implement bulk insert using [gorm](https://github.com/jinzhu/gorm). Execute bulk insert just by passing a slice of struct, as if you were using a gorm regularly.

## Purpose

When saving a large number of records in database, inserting at once - instead of inserting one by one - leads to significant performance improvement. This is widely known as bulk insert.

Gorm is one of the most popular ORM and contains very developer-friendly features, but bulk insert is not provided.

This library is aimed to solve the bulk insert problem.

## Installation

`$ go get github.com/t-tiger/gorm-bulk-insert/v2`

This library depends on gorm, following command is also necessary unless you've installed gorm.

`$ go get github.com/jinzhu/gorm`

## Usage

```go
gormbulk.BulkInsert(db, sliceValue, 3000)
```

Third argument specifies the maximum number of records to bulk insert at once. This is because inserting a large number
of records and embedding variable at once will exceed the limit of prepared statement.

Depending on the number of variables included, 2000 to 3000 is recommended.

```go
gormbulk.BulkInsert(db, sliceValue, 3000, "Name", "Email")
```

Basically, inserting struct values are automatically chosen. However if you want to exclude some columns explicitly, you can specify as argument.

In the above pattern `Name` and `Email` fields are excluded.

### Feature

- Just pass a slice of struct as using gorm normally, records will be created.
    - **NOTE: passing value must be a slice of struct. Map or other values are not compatible.**
- `CreatedAt` and `UpdatedAt` are automatically set to the current time.
- Fields of relation such as `belongsTo` and `hasMany` are automatically excluded, but foreignKey is subject to Insert.

## Example

### BulkInsert

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	gormbulk "github.com/t-tiger/gorm-bulk-insert/v2"
)

type fakeTable struct {
	ID        int `gorm:"AUTO_INCREMENT"`
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func main() {
	db, err := gorm.Open("mysql", "mydb")
	if err != nil {
		log.Fatal(err)
	}

	var insertRecords []interface{}
	for i := 0; i < 10; i++ {
		insertRecords = append(insertRecords,
			fakeTable{
				Name:  fmt.Sprintf("name%d", i),
				Email: fmt.Sprintf("test%d@test.com", i),
				// you don't need to set CreatedAt, UpdatedAt
			},
		)
	}

	err = gormbulk.BulkInsert(db, insertRecords, 3000)
	if err != nil {
		// do something
	}

	// columns you want to exclude from Insert, specify as an argument
	err = gormbulk.BulkInsert(db, insertRecords, 3000, "Email")
	if err != nil {
		// do something
	}
}
```

### BulkInsertWithAssigningIDs

```go
package main

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	gormbulk "github.com/t-tiger/gorm-bulk-insert/v2"
)

type fakeTable struct {
	IdPK      uint      `gorm:"primary_key"`
	CreatedAt time.Time `gorm:"default:now()"`
	Data      string
}

func main() {
	db, err := gorm.Open("postgres", "host=localhost port=5432 user=cloudwalker dbname=cloudwalker password=cloudwalker sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.SingularTable(true)

	if err := db.AutoMigrate(fakeTable{}).Error; err != nil {
		panic(err)
	}

	models := []interface{}{
		fakeTable{Data: "aaa"},
		fakeTable{Data: "bbb"},
		fakeTable{Data: "ccc"},
	}

	// if you want to scan * back
	var returnModel []fakeTable
	if err := gormbulk.BulkInsertWithAssigningIDs(
		db.Set("gorm:insert_option", "returning *"), &returnModel, models, 1000); err != nil {
		panic(err)
	}
	fmt.Printf("success to insert with returning: %+v\n", returnModel)
	// success to insert with returning: [
	// {IdPK:1 CreatedAt:2021-10-31 16:21:48.019947 +0000 UTC Data:aaa} 
	// {IdPK:2 CreatedAt:2021-10-31 16:21:48.019959 +0000 UTC Data:bbb} 
	// {IdPK:3 CreatedAt:2021-10-31 16:21:48.019965 +0000 UTC Data:ccc}
	// ]

	// if you want to scan primary key
	var returnId []uint
	if err := gormbulk.BulkInsertWithAssigningIDs(
		db.Set("gorm:insert_option", "returning id"), &returnId, models, 1000); err != nil {
		panic(err)
	}
	fmt.Printf("success to insert with returning: %+v\n", returnId)
	// `success to insert with returning: [4 5 6]`
}
```

## License

This project is under Apache 2.0 License. See the [LICENSE](https://github.com/kabukikeiji/gorm-bulk-insert/blob/master/LICENSE.txt) file for the full license text.
