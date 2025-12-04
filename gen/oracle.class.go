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
	_ "github.com/sijms/go-ora/v2"
	"strings"
)

//go:embed oracle_base_dao.tpl
var oracleBaseDaoTpl string

type oracleGenerator struct {
	*generator__
}

func (this *oracleGenerator) getDriverName() string {
	return "oracle"
}

func (this *oracleGenerator) getBaseDaoTemplate() string {
	return oracleBaseDaoTpl
}

func (this *oracleGenerator) getTableInfo(table string) ([]fieldTplParam, string, error) {
	var (
		exists       bool
		fields       []fieldTplParam
		tableComment string
	)

	rows := mustReturn(this.db.Query(`SELECT c.column_name, c.data_type, c.data_precision, c.data_scale, c.char_length, c.nullable = 'N', c.data_default IS NOT NULL, c2.comments FROM user_tab_columns c LEFT JOIN user_col_comments c2 ON c.table_name =c2.table_name AND c.COLUMN_NAME =c2.COLUMN_NAME WHERE c.table_name = :1 ORDER BY c.column_id`, strings.ToUpper(table)))
	defer rows.Close()
	for rows.Next() {
		exists = true
		var (
			fieldType string
			// 扫描的字段
			column          string
			dataType        string
			precision       *int
			scale           *int
			charLength      int
			isNotNull       bool
			hasDefaultValue bool
			comment         *string
		)
		must(rows.Scan(&column, &dataType, &precision, &scale, &charLength, &isNotNull, &hasDefaultValue, &comment))
		dataType = strings.ToUpper(dataType)
		if strings.HasPrefix(dataType, "TIMESTAMP") {
			dataType = "TIMESTAMP"
		}
		if comment == nil {
			comment = gdao.P("")
		}

		switch dataType {
		case "CHAR", "VARCHAR2", "VARCHAR":
			if charLength == 1 {
				fieldType = "*bool"
			} else {
				fieldType = "*string"
			}
		case "CLOB", "NCLOB", "NCHAR", "NVARCHAR2", "ROWID", "UROWID":
			fieldType = "*string"
		case "NUMBER":
			if *scale == 0 {
				fieldType = "*int64"
			} else {
				fieldType = "*float64"
			}
		case "FLOAT", "BINARY_DOUBLE":
			fieldType = "*float64"
		case "BINARY_FLOAT":
			fieldType = "*float32"
		case "DATE", "TIMESTAMP":
			fieldType = "*time.Time"
		case "BLOB", "RAW", "LONG RAW":
			fieldType = "[]byte"
		}

		f := fieldTplParam{
			Column:          column,
			FieldName:       fieldNameMapper.Convert(column),
			FieldType:       fieldType,
			IsNotNull:       isNotNull,
			HasDefaultValue: hasDefaultValue,
			Comment:         *comment,
			Valid:           fieldType != "any",
		}
		fields = append(fields, f)
	}

	rows = mustReturn(this.db.Query("SELECT comments FROM user_tab_comments WHERE table_name = :1", table))
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&tableComment)
	}
	if !exists {
		return nil, "", errors.New("\"" + table + "\" is not exists")
	}
	return fields, tableComment, nil
}

func newOracleGenerator(c GenCfg) *oracleGenerator {
	this := &oracleGenerator{}
	this.generator__ = extendGenerator_(this, c)
	return this
}
