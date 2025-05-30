package gdao

import (
	"reflect"
)

type DaoExport struct {
	ColumnsWithComma       string
	Columns                []string
	ColumnToFieldIndexMap  map[string]int
	AutoIncrementColumns   []string
	AutoIncrementStep      int64
	AutoIncrementConvertor func(id int64) reflect.Value
}

func ExportDao[T any](dao *Dao[T]) DaoExport {
	return DaoExport{
		ColumnsWithComma:       dao.commaColumns,
		Columns:                dao.columns,
		ColumnToFieldIndexMap:  dao.columnToFieldIndexMap,
		AutoIncrementColumns:   dao.autoIncrementColumns,
		AutoIncrementStep:      dao.autoIncrementStep,
		AutoIncrementConvertor: dao.autoIncrementConvertor,
	}
}

var LastInsertIdConvertors = lastInsertIdConvertors

var PrintSql = printSql
var PrintWarn = printWarn
var PrintError = printError
