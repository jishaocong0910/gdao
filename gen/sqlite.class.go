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

func (this *sqliteGenerator) getTableInfo(table string) (bool, []*fieldTplParam, string) {
	var (
		exists       bool
		fields       []*fieldTplParam
		tableComment string
	)

	exists, st := this.getSqlTokenizer(table)
	if !exists { // coverage-ignore
		return exists, fields, tableComment
	}
	if !(st.nextUntilDelimiterAndToken("(") && st.nextToken()) {
		return exists, fields, tableComment
	}

	for {
		if st.token == ")" || st.token == "" {
			break
		}
		if strings.ToUpper(st.token) == "PRIMARY" { // coverage-ignore
			if st.nextUntilDelimiterAndToken("(") && st.nextToken() {
				primaryKeycolumn := st.token
				for st.nextDelimiterAndToken() {
					if st.token == ")" {
						break
					}
					if !st.tokenIsDelimiter {
						primaryKeycolumn = ""
						break
					}
				}
				if primaryKeycolumn != "" {
					for _, f := range fields {
						if f.Column == primaryKeycolumn {
							f.IsAutoIncrement = true
						}
					}
				}
			}
			break
		}

		var (
			column          string
			dataType        string
			isAutoIncrement bool
			fieldType       string
			comment         string
		)
		column = st.token
		st.nextToken()
		dataType = strings.ToUpper(st.token)

		if dataType == "UNSIGNED" {
			st.nextToken()
			dataType = strings.ToUpper(st.token)
		}

		fieldType = "any"
		switch dataType {
		case "INTEGER", "INT", "TINYINT", "SMALLINT", "MEDIUMINT", "BIGINT", "UNSIGNED BIG INT", "INT2", "INT8", "TINY", "SMALL", "MEDIUM", "BIG":
			fieldType = "*int64"
		case "REAL", "NUMERIC", "DOUBLE", "FLOAT", "DECIMAL":
			fieldType = "*float64"
		case "BOOLEAN":
			fieldType = "*bool"
		case "DATE", "DATETIME":
			fieldType = "*time.Time"
		case "TEXT", "CHARACTER", "VARCHAR", "VARYING", "NCHAR", "NATIVE", "NVARCHAR", "CLOB":
			fieldType = "*string"
		case "BLOB":
			fieldType = "[]byte"
		}

		for st.nextDelimiterAndToken(); ; {
			if st.token == "(" {
				st.nextUntilDelimiterAndToken(")")
				if !st.nextDelimiterAndToken() { // coverage-ignore
					break
				}
			} else if strings.ToUpper(st.token) == "PRIMARY" {
				isAutoIncrement = true
				if !st.nextDelimiterAndToken() { // coverage-ignore
					break
				}
			} else if st.token == "," {
				st.nextToken()
				if st.token == "--" {
					comment = this.nextComment(st)
					st.nextToken()
				}
				break
			} else if st.token == "--" {
				comment = this.nextComment(st)
				st.nextToken()
				break
			} else if st.token == ")" {
				break
			} else if !st.nextDelimiterAndToken() {
				break
			}
		}

		f := &fieldTplParam{
			Column:          column,
			FieldName:       fieldNameMapper.Convert(column),
			FieldType:       fieldType,
			IsAutoIncrement: isAutoIncrement,
			Comment:         comment,
			Valid:           fieldType != "any",
		}
		fields = append(fields, f)
	}

	st.reset()
	for {
		if st.token == "(" || st.token == "--" {
			break
		}
		if !st.nextDelimiterAndToken() { // coverage-ignore
			break
		}
	}
	tableComment = this.nextComment(st)

	return exists, fields, tableComment
}

func (this *sqliteGenerator) getSqlTokenizer(table string) (bool, *stringTokenizer) {
	rows := mustReturn(this.db.Query("SELECT sql FROM sqlite_master WHERE type='table' AND name=?", table))
	defer rows.Close()
	if !rows.Next() { // coverage-ignore
		return false, nil
	}
	var sql string
	mustNoError(rows.Scan(&sql))
	return true, newStringTokenizer(sql, []rune{' ', '\t', '\n', ',', '(', ')'})
}

func (this *sqliteGenerator) nextComment(st *stringTokenizer) string {
	if st.token != "--" { // coverage-ignore
		return ""
	}
	begin := st.pos + 1
	for {
		if st.token == "\n" {
			break
		}
		if !st.nextDelimiterAndToken() { // coverage-ignore
			break
		}
	}
	return string(st.chars[begin : st.pos-1])
}

func newSqliteGenerator(c GenCfg) *sqliteGenerator {
	this := &sqliteGenerator{}
	this.generator__ = extendGenerator_(this, c)
	return this
}
