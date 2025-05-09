package gdao

import (
	"reflect"
)

type DaoExport struct {
	Table                  string
	ColumnsWithComma       string
	Columns                []string
	ColumnToFieldIndexMap  map[string]int
	AutoIncrementColumns   []string
	AutoIncrementStep      int64
	AutoIncrementConvertor func(id int64) reflect.Value
}

func ExportDao[T any](dao *Dao[T]) DaoExport {
	return DaoExport{
		Table:                  dao.p.Table,
		ColumnsWithComma:       dao.p.CommaColumns,
		Columns:                dao.p.Columns,
		ColumnToFieldIndexMap:  dao.p.ColumnToFieldIndexMap,
		AutoIncrementColumns:   dao.p.AutoIncrementColumns,
		AutoIncrementStep:      dao.p.AutoIncrementStep,
		AutoIncrementConvertor: dao.p.AutoIncrementConvertor,
	}
}

var LastInsertIdConvertors = lastInsertIdConvertors

var PrintSql = printSql
var PrintArgs = printArgs
var PrintAffected = printAffected
var PrintWarn = printWarn
var PrintError = printError
