package gen

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type sqliteGenerator struct {
	c  Config
	db *sql.DB
}

func (g sqliteGenerator) getTableInfo(table string) (bool, []*field, string) {
	var (
		exists       bool
		fields       []*field
		tableComment string
	)

	exists, st := g.getSqlTokenizer(table)
	if !exists {
		return exists, fields, tableComment
	}
	if !(st.nextUntilDelimiterAndToken("(") && st.nextToken()) {
		return exists, fields, tableComment
	}

	for {
		if st.token == ")" || st.token == "" {
			break
		}
		if strings.ToUpper(st.token) == "PRIMARY" {
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
				if !st.nextDelimiterAndToken() {
					break
				}
			} else if strings.ToUpper(st.token) == "PRIMARY" {
				isAutoIncrement = true
				if !st.nextDelimiterAndToken() {
					break
				}
			} else if st.token == "," {
				st.nextToken()
				if st.token == "--" {
					comment = g.nextComment(st)
					st.nextToken()
				}
				break
			} else if st.token == "--" {
				comment = g.nextComment(st)
				st.nextToken()
				break
			} else if st.token == ")" {
				break
			} else if !st.nextDelimiterAndToken() {
				break
			}
		}

		f := &field{
			Column:          column,
			FieldName:       g.c.FieldNameMapper.Convert(column),
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
		if !st.nextDelimiterAndToken() {
			break
		}
	}
	tableComment = g.nextComment(st)

	return exists, fields, tableComment
}

func (g *sqliteGenerator) getSqlTokenizer(table string) (bool, *stringTokenizer) {
	rows := mustReturn(g.db.Query("SELECT sql FROM sqlite_master WHERE type='table' AND name=?", table))
	defer rows.Close()
	if !rows.Next() {
		return false, nil
	}
	var sql string
	mustNoError(rows.Scan(&sql))
	return true, newStringTokenizer(sql, []rune{' ', '\t', '\n', ',', '(', ')'})
}

func (g sqliteGenerator) nextComment(st *stringTokenizer) string {
	if st.token != "--" {
		return ""
	}
	begin := st.pos + 1
	for {
		if st.token == "\n" {
			break
		}
		if !st.nextDelimiterAndToken() {
			break
		}
	}
	return string(st.chars[begin : st.pos-1])
}

func newSqliteGenerator(c Config) sqliteGenerator {
	db, err := sql.Open("sqlite3", c.Dsn)
	if err != nil { // coverage-ignore
		panic(fmt.Sprintf("connect db fail, dsn: %s, error: %v", c.Dsn, err))
	}
	return sqliteGenerator{c: c, db: db}
}

type stringTokenizer struct {
	text             string
	chars            []rune
	len              int
	c                rune
	pos              int
	end              bool
	token            string
	tokenIsDelimiter bool
	delimiters       map[rune]struct{}
}

func (st *stringTokenizer) nextDelimiterAndToken() bool {
	if st.end {
		return false
	}
	st.token = ""
	st.tokenIsDelimiter = false
	if _, ok := st.delimiters[st.c]; ok {
		st.token = string(st.c)
		st.tokenIsDelimiter = true
		st.nextChar()
	} else {
		begin := st.pos
		for st.pos < st.len && st.nextChar() {
			if _, ok := st.delimiters[st.c]; ok {
				break
			}
		}
		st.token = string(st.chars[begin:st.pos])
	}
	return true
}

func (st *stringTokenizer) nextToken() bool {
	for st.nextDelimiterAndToken() {
		cs := []rune(st.token)
		if len(cs) != 1 {
			return true
		}
		if _, ok := st.delimiters[cs[0]]; !ok {
			return true
		}
	}
	return false
}

func (st *stringTokenizer) nextUntilDelimiterAndToken(token string) bool {
	for {
		if strings.EqualFold(st.token, token) {
			return true
		}
		if !st.nextDelimiterAndToken() {
			break
		}
	}
	return false
}

func (st *stringTokenizer) nextChar() bool {
	if st.end {
		return false
	}
	st.c = 0
	st.pos++
	if st.pos == st.len {
		st.end = true
		return false
	}
	st.c = st.chars[st.pos]
	return true
}

func (st *stringTokenizer) reset() {
	st.c = 0
	st.pos = -1
	st.end = false
	st.token = ""
	st.tokenIsDelimiter = false
	st.nextChar()
	st.nextDelimiterAndToken()
}

func newStringTokenizer(text string, delimiters []rune) *stringTokenizer {
	chars := []rune(text)
	m := make(map[rune]struct{})
	for _, c := range delimiters {
		m[c] = struct{}{}
	}
	st := stringTokenizer{text: text, chars: chars, len: len(chars), pos: -1, delimiters: m}
	st.nextChar()
	st.nextDelimiterAndToken()
	return &st
}
