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

package gen

import (
	_ "embed"
	"errors"
	"github.com/jishaocong0910/gdao"
	_ "github.com/microsoft/go-mssqldb"
	"strings"
)

//go:embed sqlserver_base_dao.tpl
var sqlserverBaseDaoTpl string

type sqlServerGenerator struct {
	*generator__
}

func (g sqlServerGenerator) getDriverName() string {
	return "mssql"
}

func (g sqlServerGenerator) getBaseDaoTemplate() string {
	return sqlserverBaseDaoTpl
}

func (g sqlServerGenerator) getTableInfo(table string) ([]fieldTplParam, string, error) {
	var (
		exists       bool
		fields       []fieldTplParam
		tableComment string
	)

	rows := mustReturn(g.db.Query("SELECT B.name, C.increment_value, D.data_type, E.value AS comment FROM sys.tables A LEFT JOIN sys.columns B ON A.object_id = B.object_id LEFT JOIN sys.identity_columns C ON A.object_id =C.object_id and B.name=C.name LEFT JOIN information_schema.columns D ON B.name = D.column_name LEFT JOIN sys.extended_properties E ON B.object_id = E.major_id AND B.column_id = E.minor_id WHERE A.name = :1 AND D.table_name = :2 ORDER BY D.ordinal_position", table, table))
	defer rows.Close()
	for rows.Next() {
		exists = true
		var (
			fieldType string

			column         string
			dataType       string
			comment        *string
			incrementValue *int
		)
		must(rows.Scan(&column, &incrementValue, &dataType, &comment))
		dataType = strings.ToLower(dataType)
		if comment == nil {
			comment = gdao.P("")
		}
		if incrementValue == nil {
			incrementValue = gdao.P(0)
		}

		fieldType = "any"
		switch dataType {
		case "tinyint":
			fieldType = "*uint"
		case "smallint":
			fieldType = "*int16"
		case "int":
			fieldType = "*int32"
		case "bigint":
			fieldType = "*int64"
		case "bit":
			fieldType = "*bool"
		case "decimal", "numeric", "money", "smallmoney", "float":
			fieldType = "*float64"
		case "real":
			fieldType = "*float32"
		case "date", "time", "datetime2", "datetimeoffset", "datetime", "smalldatetime":
			fieldType = "*time.Time"
		case "char", "varchar", "text", "nchar", "nvarchar", "ntext", "xml", "json":
			fieldType = "*string"
		case "binary", "varbinary", "image", "geography", "geometry", "hierarchyid", "uniqueidentifier":
			fieldType = "[]byte"
		}

		f := fieldTplParam{
			Column:            column,
			FieldName:         fieldNameMapper.Convert(column),
			FieldType:         fieldType,
			Comment:           *comment,
			Valid:             fieldType != "any",
			IsAutoIncrement:   *incrementValue > 0,
			AutoIncrementStep: *incrementValue,
		}
		fields = append(fields, f)
	}

	rows = mustReturn(g.db.Query("SELECT value FROM sys.extended_properties WHERE major_id = object_id (:1) AND minor_id = 0", table))
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&tableComment)
	}
	if !exists {
		return nil, "", errors.New("\"" + table + "\" is not exists")
	}
	return fields, tableComment, nil
}

func newSqlServerGenerator(c GenCfg) *sqlServerGenerator {
	this := &sqlServerGenerator{}
	this.generator__ = extendGenerator_(this, c)
	return this
}
