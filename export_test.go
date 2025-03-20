package gdao

import (
	"reflect"
)

type DaoExport struct {
	Table                  string
	ColumnsWithComma       string
	Columns                []string
	ColumnToFieldIndexMap  map[string]int
	AutoIncrementColumn    string
	AutoIncrementOffset    int64
	AutoIncrementConvertor func(id int64) reflect.Value
}

func ExportDao[T any](dao Dao[T]) DaoExport {
	return DaoExport{
		Table:                  dao.table,
		ColumnsWithComma:       dao.columnsWithComma,
		Columns:                dao.columns,
		ColumnToFieldIndexMap:  dao.columnToFieldIndexMap,
		AutoIncrementColumn:    dao.autoIncrementColumn,
		AutoIncrementOffset:    dao.autoIncrementOffset,
		AutoIncrementConvertor: dao.autoIncrementConvertor,
	}
}

var LastInsertIdConvertors = lastInsertIdConvertors

var PrintSql = printSql
var PrintWarn = printWarn
var PrintError = printError
