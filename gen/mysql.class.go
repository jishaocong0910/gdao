package gen

import (
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type mySqlGenerator struct {
	*m_Generator
	database string
}

func (this *mySqlGenerator) getEntityComment(table string) string {
	rows := mustReturn(this.db.Query("SELECT TABLE_COMMENT FROM information_schema.tables WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?", this.database, table))
	defer rows.Close()
	var comment string
	if rows.Next() {
		rows.Scan(&comment)
	}
	return comment
}

func (this *mySqlGenerator) getEntityFields(table string) []*field {
	rows := mustReturn(this.db.Query("SELECT COLUMN_NAME, DATA_TYPE, COLUMN_TYPE, EXTRA = 'auto_increment', COLUMN_COMMENT FROM information_schema.columns WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION", this.database, table))
	defer rows.Close()

	var fields []*field
	for rows.Next() {
		var (
			column          string
			dataType        string
			columnType      string
			isAutoIncrement bool
			comment         string
		)
		mustNoError(rows.Scan(&column, &dataType, &columnType, &isAutoIncrement, &comment))

		f := &field{
			Column:    column,
			FieldName: this.c.FieldNameMapper.Convert(column),
			FieldType: "any",
			Comment:   comment,
		}
		if isAutoIncrement {
			f.IsAutoIncrement = true
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
		case "varchar", "char", "text", "tinytext", "mediumtext", "longtext", "enum", "json", "set":
			f.FieldType = "*string"
		case "datetime", "timestamp", "date", "time":
			f.FieldType = "*time.Time"
		case "year":
			f.FieldType = "*int32"
		case "binary", "varbinary", "geometry", "blob", "tinyblob", "mediumblob", "longblob":
			f.FieldType = "[]byte"
		}
		if f.FieldType != "any" {
			f.Valid = true
		}
		fields = append(fields, f)
	}
	return fields
}

func newMySqlGenerator(c Config) *mySqlGenerator {
	g := &mySqlGenerator{}
	g.m_Generator = extendGenerator(g, c)

	rows := mustReturn(g.db.Query("SELECT DATABASE()"))
	if rows.Next() {
		rows.Scan(&g.database)
	}
	rows.Close()
	return g
}
