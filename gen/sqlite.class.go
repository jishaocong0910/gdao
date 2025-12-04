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
	_ "github.com/mattn/go-sqlite3"
	"strings"
)

//go:embed sqlite_base_dao.tpl
var sqliteBaseDaoTpl string

type sqliteGenerator struct {
	*generator__
}

func (this *sqliteGenerator) getDriverName() string {
	return "sqlite3"
}

func (this *sqliteGenerator) getBaseDaoTemplate() string {
	return sqliteBaseDaoTpl
}

func (this *sqliteGenerator) getTableInfo(table string) ([]fieldTplParam, string, error) {
	var (
		exists  bool
		pkCount int
		fields  []fieldTplParam
	)

	rows := mustReturn(this.db.Query("PRAGMA table_info(" + table + ")"))
	defer rows.Close()
	for rows.Next() {
		exists = true
		var (
			fieldType       string
			hasDefaultValue bool
			isAutoIncrement bool
			// 扫描的字段
			column    string
			dataType  string
			isNotNull bool
			dfltValue *any
			pk        int
		)

		must(rows.Scan(new(any), &column, &dataType, &isNotNull, &dfltValue, &pk))
		dataType = strings.ToUpper(dataType)
		hasDefaultValue = dfltValue != nil
		if pk != 0 {
			isAutoIncrement = true
			pkCount++
		}
		if strings.HasPrefix(dataType, "DECIMAL") {
			dataType = "DECIMAL"
		}

		fieldType = "any"
		switch dataType {
		case "TINYINT", "INT8":
			fieldType = "*int8"
		case "SMALLINT", "INT2":
			fieldType = "*int16"
		case "MEDIUM INT", "INTEGER", "INT":
			fieldType = "*int32"
		case "BIGINT":
			fieldType = "*int64"
		case "UNSIGNED INT", "UNSIGNED INTEGER":
			fieldType = "*uint32"
		case "UNSIGNED BIG INT":
			fieldType = "*uint64"
		case "REAL", "NUMERIC", "DOUBLE", "DOUBLE PRECISION", "FLOAT", "DECIMAL":
			fieldType = "*float64"
		case "BOOLEAN":
			fieldType = "*bool"
		case "DATETIME", "TIMESTAMP", "TIMESTAMP WITH TIME ZONE", "TIME WITH TIME ZONE":
			fieldType = "*time.Time"
		case "TIME", "DATE", "TEXT", "CHARACTER", "VARCHAR", "VARYING CHARACTER", "VARYING", "NCHAR", "NATIVE CHARACTER", "NVARCHAR", "CLOB":
			fieldType = "*string"
		case "BLOB":
			fieldType = "[]byte"
		}

		f := fieldTplParam{
			Column:          column,
			FieldName:       fieldNameMapper.Convert(column),
			FieldType:       fieldType,
			IsAutoIncrement: isAutoIncrement,
			IsNotNull:       isNotNull,
			HasDefaultValue: hasDefaultValue,
			Valid:           fieldType != "any",
		}
		fields = append(fields, f)
	}
	if pkCount > 1 {
		for _, field := range fields {
			field.IsAutoIncrement = false
		}
	}
	if !exists {
		return nil, "", errors.New("\"" + table + "\" is not exists")
	}
	return fields, "", nil
}

func newSqliteGenerator(c GenCfg) *sqliteGenerator {
	this := &sqliteGenerator{}
	this.generator__ = extendGenerator_(this, c)
	return this
}
