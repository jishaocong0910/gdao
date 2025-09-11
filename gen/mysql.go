package gen

import (
	"database/sql"
	"fmt"
	"strings"
	"text/template"

	_ "embed"

	_ "github.com/go-sql-driver/mysql"
)

//go:embed mysql_base_dao.tpl
var mysqlBaseDaoTpl string

type mySqlGenerator struct {
	baseDaoTemplate *template.Template
	c               Cfg
	db              *sql.DB
	database        string
}

func (g mySqlGenerator) getBaseDaoTemplate() *template.Template {
	return g.baseDaoTemplate
}

func (g mySqlGenerator) getTableInfo(table string) (bool, []*field, string) {
	var (
		exists       bool
		fields       []*field
		tableComment string
	)

	rows := mustReturn(g.db.Query("SELECT COLUMN_NAME, DATA_TYPE, COLUMN_TYPE, EXTRA = 'auto_increment', COLUMN_COMMENT FROM information_schema.columns WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION", g.database, table))
	defer rows.Close()
	for rows.Next() {
		exists = true
		var (
			column          string
			dataType        string
			columnType      string
			isAutoIncrement bool
			comment         string
		)
		mustNoError(rows.Scan(&column, &dataType, &columnType, &isAutoIncrement, &comment))

		f := &field{
			Column:          column,
			FieldName:       fieldNameMapper.Convert(column),
			FieldType:       "any",
			Comment:         comment,
			IsAutoIncrement: isAutoIncrement,
		}

		dataType = strings.ToLower(dataType)
		columnType = strings.ToLower(columnType)
		switch dataType {
		case "bit":
			f.FieldType = "[]uint8"
		case "tinyint":
			if strings.Contains(columnType, "unsigned") {
				f.FieldType = "*uint8"
			} else if strings.Contains(columnType, "tinyint(1)") {
				f.FieldType = "*bool"
			} else {
				f.FieldType = "*int8"
			}
		case "smallint":
			if strings.Contains(columnType, "unsigned") {
				f.FieldType = "*uint16"
			} else {
				f.FieldType = "*int16"
			}
		case "mediumint":
			if strings.Contains(columnType, "unsigned") {
				f.FieldType = "*uint32"
			} else {
				f.FieldType = "*int32"
			}
		case "int":
			if strings.Contains(columnType, "unsigned") {
				f.FieldType = "*uint32"
			} else {
				f.FieldType = "*int32"
			}
		case "bigint":
			if strings.Contains(columnType, "unsigned") {
				f.FieldType = "*uint64"
			} else {
				f.FieldType = "*int64"
			}
		case "double", "decimal":
			f.FieldType = "*float64"
		case "float":
			f.FieldType = "*float32"
		case "varchar", "char", "text", "tinytext", "mediumtext", "longtext", "enum", "json", "set", "time":
			f.FieldType = "*string"
		case "datetime", "timestamp", "date":
			f.FieldType = "*time.Time"
		case "year":
			f.FieldType = "*int64"
		case "binary", "varbinary", "geometry", "blob", "tinyblob", "mediumblob", "longblob":
			f.FieldType = "[]byte"
		}
		if f.FieldType != "any" {
			f.Valid = true
		}
		fields = append(fields, f)
	}

	rows = mustReturn(g.db.Query("SELECT TABLE_COMMENT FROM information_schema.tables WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?", g.database, table))
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&tableComment)
	}

	return exists, fields, tableComment
}

func newMySqlGenerator(c Cfg) mySqlGenerator {
	db, err := sql.Open("mysql", c.Dsn)
	if err != nil { // coverage-ignore
		panic(fmt.Sprintf("connect db fail, dsn: %s, error: %v", c.Dsn, err))
	}
	database := ""
	rows := mustReturn(db.Query("SELECT DATABASE()"))
	if rows.Next() {
		rows.Scan(&database)
	}
	rows.Close()
	t := mustReturn(template.New("").Parse(mysqlBaseDaoTpl))
	return mySqlGenerator{baseDaoTemplate: t, c: c, db: db, database: database}
}
