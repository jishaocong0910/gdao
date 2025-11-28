/*
Copyright 2024-present jishaocong0910

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gen

import (
	_ "embed"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"strings"
)

//go:embed mysql_base_dao.tpl
var mysqlBaseDaoTpl string

type mySqlGenerator struct {
	*generator__
	database string
}

func (this *mySqlGenerator) getDriverName() string {
	return "mysql"
}

func (this *mySqlGenerator) getBaseDaoTemplate() string {
	return mysqlBaseDaoTpl
}

func (this *mySqlGenerator) getTableInfo(table string) ([]fieldTplParam, string, error) {
	var (
		exists       bool
		fields       []fieldTplParam
		tableComment string
	)

	rows := mustReturn(this.db.Query("SELECT COLUMN_NAME, DATA_TYPE, COLUMN_TYPE, EXTRA = 'auto_increment', IS_NULLABLE = 'NO', COLUMN_DEFAULT IS NOT NULL, COLUMN_COMMENT FROM information_schema.columns WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION", this.database, table))
	defer rows.Close()
	for rows.Next() {
		exists = true
		var (
			fieldType string

			column          string
			dataType        string
			columnType      string
			isAutoIncrement bool
			isNotNull       bool
			hasDefaultValue bool
			comment         string
		)
		must(rows.Scan(&column, &dataType, &columnType, &isAutoIncrement, &isNotNull, &hasDefaultValue, &comment))

		dataType = strings.ToLower(dataType)
		columnType = strings.ToLower(columnType)
		switch dataType {
		case "bit":
			fieldType = "[]uint8"
		case "tinyint":
			if strings.Contains(columnType, "unsigned") {
				fieldType = "*uint8"
			} else if strings.Contains(columnType, "tinyint(1)") {
				fieldType = "*bool"
			} else {
				fieldType = "*int8"
			}
		case "smallint":
			if strings.Contains(columnType, "unsigned") {
				fieldType = "*uint16"
			} else {
				fieldType = "*int16"
			}
		case "mediumint":
			if strings.Contains(columnType, "unsigned") {
				fieldType = "*uint32"
			} else {
				fieldType = "*int32"
			}
		case "int":
			if strings.Contains(columnType, "unsigned") {
				fieldType = "*uint32"
			} else {
				fieldType = "*int32"
			}
		case "bigint":
			if strings.Contains(columnType, "unsigned") {
				fieldType = "*uint64"
			} else {
				fieldType = "*int64"
			}
		case "double", "decimal":
			fieldType = "*float64"
		case "float":
			fieldType = "*float32"
		case "varchar", "char", "text", "tinytext", "mediumtext", "longtext", "enum", "json", "set", "time":
			fieldType = "*string"
		case "datetime", "timestamp", "date":
			fieldType = "*time.Time"
		case "year":
			fieldType = "*int64"
		case "binary", "varbinary", "geometry", "blob", "tinyblob", "mediumblob", "longblob":
			fieldType = "[]byte"
		}

		f := fieldTplParam{
			Column:          column,
			FieldName:       fieldNameMapper.Convert(column),
			FieldType:       fieldType,
			IsAutoIncrement: isAutoIncrement,
			IsNotNull:       isNotNull,
			HasDefaultValue: hasDefaultValue,
			Comment:         comment,
			Valid:           fieldType != "any",
		}
		fields = append(fields, f)
	}

	rows = mustReturn(this.db.Query("SELECT TABLE_COMMENT FROM information_schema.tables WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?", this.database, table))
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&tableComment)
	}
	if !exists {
		return nil, "", errors.New("\"" + table + "\" is not exists")
	}
	return fields, tableComment, nil
}

func newMySqlGenerator(c GenCfg) *mySqlGenerator {
	this := &mySqlGenerator{}
	this.generator__ = extendGenerator_(this, c)

	if this.db != nil {
		database := ""
		rows := mustReturn(this.db.Query("SELECT DATABASE()"))
		if rows.Next() {
			rows.Scan(&database)
		}
		rows.Close()
		this.database = database
	}
	return this
}
