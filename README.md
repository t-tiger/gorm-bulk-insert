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

## Example


## License

This project is under Apache 2.0 License. See the [LICENSE](https://github.com/kabukikeiji/gorm-bulk-insert/blob/master/LICENSE.txt) file for the full license text.
