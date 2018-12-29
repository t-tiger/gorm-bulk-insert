# Gorm Bulk Insert

Gorm Bulk Insert is a library to implement bulk insert using [gorm](https://github.com/jinzhu/gorm). Execute bulk insert just by passing a struct, as if you were using a regular gorm method.

## Purpose

Saving a large number of records in the database, inserting at once, instead of executing insert every time leads to significant performance improvement. This is widely known as bulk insert.

On the other hand, gorm is one of the most popular ORMapper and provides very developers-friendly features, but bulk insert is not contained.

This library is aimed to solve the bulk insert problem faced by developers using gorm.

## Installation

`go get github.com/casbin/gorm-adapter`

This library depends on gorm, following command is also necessary unless you've installed.

`go get github.com/jinzhu/gorm`


## Usage

```go
gormbulk.BulkInsert(db, sliceValue, 3000)
```

Third argument specifies the maximum number of records to BulkInsert at once. This is because inserting a large number of records and embedding variable at once will exceed the limit of prepared statement.

Depending on the number of variables included, 2000 to 3000 is recommended.

```go
gormbulk.BulkInsert(db, sliceValue, 3000, "Name", "Email")
```

Basically, necessary values are automatically chosen. However if you want to exclude some columns explicitly , you can specify this as argument.

In the above pattern `Name` and` Email` fields are excluded.

### Feature

- Just pass a slice of struct as use gorm normally, records will be created accordingly.
    - **NOTE: passing must be a struct. Map values are not compatible.**
- `CreatedAt` and `UpdatedAt` are automatically set to the current time.
- Fields of relation such as `belongsTo` and `hasMany` are automatically excluded, but foreignKey is subject to Insert.

## Example

```go
package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/t-tiger/gorm-bulk-insert"
	"log"
	"time"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type fakeTable struct {
	ID        int
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func main() {
	db, err := gorm.Open("sqlite3", "mydb")
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

	err := gormbulk.BulkInsert(db, insertRecords, 3000)
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

## License

This project is under Apache 2.0 License. See the [LICENSE](https://github.com/kabukikeiji/gorm-bulk-insert/blob/master/LICENSE.txt) file for the full license text.
