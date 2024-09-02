package gen

import (
	"strings"
)

type oracleGenerator struct {
	*m_Generator
}

func (this *oracleGenerator) getEntityComment(table string) string {
	rows := mustReturn(this.db.Query("SELECT comments FROM user_tab_comments WHERE table_name = :2", table))
	defer rows.Close()
	var comment string
	if rows.Next() {
		rows.Scan(&comment)
	}
	return comment
}

func (this *oracleGenerator) getEntityFields(table string) []*field {
	rows := mustReturn(this.db.Query(`SELECT c.column_name, c.data_type, c.data_precision , c.data_scale , c.char_length, c2.comments FROM user_tab_columns c LEFT JOIN user_col_comments c2 ON c.table_name =c2.table_name AND c.COLUMN_NAME =c2.COLUMN_NAME WHERE c.table_name = :1`, table))
	defer rows.Close()
	var fields []*field
	for rows.Next() {
		var (
			column     string
			dataType   string
			precision  *int
			scale      *int
			charLength int
			comment    *string
		)
		mustNoError(rows.Scan(&column, &dataType, &precision, &scale, &charLength, &comment))

		f := &field{
			Column:    column,
			FieldName: this.c.FieldNameMapper.Convert(column),
			FieldType: "any",
		}

		if comment != nil {
			f.Comment = *comment
		}

		if strings.HasPrefix(dataType, "TIMESTAMP") {
			dataType = "TIMESTAMP"
		}

		switch dataType {
		case "CHAR", "VARCHAR2", "VARCHAR":
			if charLength == 1 {
				f.FieldType = "*bool"
			} else {
				f.FieldType = "*string"
			}
		case "CLOB", "NCLOB", "NCHAR", "NVARCHAR2", "ROWID", "UROWID":
			f.FieldType = "*string"
		case "NUMBER":
			if *scale == 0 {
				f.FieldType = "*int64"
			} else {
				f.FieldType = "*float64"
			}
		case "FLOAT", "BINARY_DOUBLE":
			f.FieldType = "*float64"
		case "BINARY_FLOAT":
			f.FieldType = "*float32"
		case "DATE", "TIMESTAMP":
			f.FieldType = "*time.Time"
		case "BLOB", "RAW", "LONG RAW":
			f.FieldType = "[]byte"
		}
		if f.FieldType != "any" {
			f.Valid = true
		}
		fields = append(fields, f)
	}
	return fields
}

func newOracleGenerator(c Config) *oracleGenerator {
	g := &oracleGenerator{}
	g.m_Generator = extendGenerator(g, c)
	return g
}
