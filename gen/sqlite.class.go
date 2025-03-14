package gen

import (
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type sqliteGenerator struct {
	*m_Generator
}

func (this *sqliteGenerator) existsTable(table string) bool {
	rows := mustReturn(this.db.Query("SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", table))
	defer rows.Close()
	var count int
	if rows.Next() {
		rows.Scan(&count)
	}
	return count == 1
}

func (this *sqliteGenerator) getEntityComment(table string) string {
	st := this.getSqlTokenizer(table)
	for {
		if st.token == "(" || st.token == "--" {
			break
		}
		if !st.nextIncludeDelimiterToken() {
			break
		}
	}
	return this.getComment(st)
}

func (this *sqliteGenerator) getEntityFields(table string) []*field {
	st := this.getSqlTokenizer(table)
	if !(st.nextUntilToken("(") && st.nextToken()) {
		return nil
	}

	var fields []*field
	for {
		if st.token == ")" || st.token == "" {
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
		dataType = st.token
		st.nextIncludeDelimiterToken()

		fieldType = "any"
		switch strings.ToUpper(dataType) {
		case "INTEGER", "INT", "TINYINT", "SMALLINT", "MEDIUMINT", "BIGINT", "UNSIGNED BIG INT", "INT2", "INT8":
			fieldType = "*int64"
		case "REAL", "NUMERIC", "DOUBLE", "DOUBLE PRECISION", "FLOAT", "DECIMAL", "BOOLEAN", "DATE", "DATETIME":
			fieldType = "*float64"
		case "TEXT", "CHARACTER", "VARCHAR", "VARYING CHARACTER", "NCHAR", "NATIVE CHARACTER", "NVARCHAR", "CLOB":
			fieldType = "*string"
		case "BLOB":
			fieldType = "[]byte"
		}

		for {
			if st.token == "(" {
				st.nextUntilToken(")")
				if !st.nextIncludeDelimiterToken() {
					break
				}
			} else if strings.ToUpper(st.token) == "PRIMARY" {
				isAutoIncrement = true
				if !st.nextIncludeDelimiterToken() {
					break
				}
			} else if st.token == "," {
				st.nextToken()
				if st.token == "--" {
					comment = this.getComment(st)
					st.nextToken()
				}
				break
			} else if st.token == "--" {
				comment = this.getComment(st)
				st.nextToken()
				break
			} else if st.token == ")" {
				break
			} else if !st.nextIncludeDelimiterToken() {
				break
			}
		}

		f := &field{
			Column:          column,
			FieldName:       this.c.FieldNameMapper.Convert(column),
			FieldType:       fieldType,
			IsAutoIncrement: isAutoIncrement,
			Comment:         comment,
			Valid:           fieldType != "any",
		}
		fields = append(fields, f)
	}
	return fields
}

func (this *sqliteGenerator) getSqlTokenizer(table string) *stringTokenizer {
	rows := mustReturn(this.db.Query("SELECT sql FROM sqlite_master WHERE type='table' AND name=?", table))
	defer rows.Close()
	if !rows.Next() {
		return nil
	}
	var sql string
	mustNoError(rows.Scan(&sql))
	return newStringTokenizer(sql)
}

func (this *sqliteGenerator) getComment(st *stringTokenizer) string {
	if st.token != "--" {
		return ""
	}
	begin := st.pos + 1
	for {
		if st.token == "\n" {
			break
		}
		if !st.nextIncludeDelimiterToken() {
			break
		}
	}
	return string(st.chars[begin : st.pos-1])
}

func newSqliteGenerator(c Config) *sqliteGenerator {
	g := &sqliteGenerator{}
	g.m_Generator = extendGenerator(g, c)
	return g
}
