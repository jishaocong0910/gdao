/*
 * Copyright 2024-present jishaocong0910
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gdao

import (
	"reflect"
)

type DaoExport struct {
	ColumnsWithComma       string
	Columns                []string
	ColumnToFieldIndex     map[string]int
	ColumnToFieldConvertor map[string]fieldConvertor
	FieldNameToColumn      map[string]string
	AutoIncrementColumns   []string
	AutoIncrementStep      int64
	AutoIncrementConvertor func(id int64) reflect.Value
}

func ExportDao[T any](dao *Dao[T]) DaoExport {
	return DaoExport{
		ColumnsWithComma:       dao.commaColumns,
		Columns:                dao.columns,
		ColumnToFieldIndex:     dao.columnToFieldIndex,
		ColumnToFieldConvertor: dao.columnToFieldConvertor,
		FieldNameToColumn:      dao.fieldNameToColumn,
		AutoIncrementColumns:   dao.autoIncrementColumns,
		AutoIncrementStep:      dao.autoIncrementStep,
		AutoIncrementConvertor: dao.autoIncrementConvert,
	}
}

func ConvertLastInsertId(typeName string, id int64) any {
	return lastInsertIdConvertor_.OfString(typeName).convert(id).Elem().Interface()
}

var PrintSql = printSql
var PrintWarn = printWarn
