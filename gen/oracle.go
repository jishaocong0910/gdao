package gen

import (
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/jishaocong0910/gdao"
	_ "github.com/sijms/go-ora/v2"
)

//go:embed oracle_base_dao.tpl
var oracleBaseDaoTpl string

type oracleGenerator struct {
	baseDaoTemplate *template.Template
	c               Cfg
	db              *sql.DB
}

func (g oracleGenerator) getBaseDaoTemplate() *template.Template {
	return g.baseDaoTemplate
}

func (g oracleGenerator) getTableInfo(table string) (bool, []*field, string) {
	var (
		exists       bool
		fields       []*field
		tableComment string
	)

	rows := mustReturn(g.db.Query(`SELECT c.column_name, c.data_type, c.data_precision , c.data_scale , c.char_length, c2.comments FROM user_tab_columns c LEFT JOIN user_col_comments c2 ON c.table_name =c2.table_name AND c.COLUMN_NAME =c2.COLUMN_NAME WHERE c.table_name = :1 ORDER BY c.column_id`, strings.ToUpper(table)))
	defer rows.Close()
	for rows.Next() {
		exists = true
		var (
			column     string
			dataType   string
			precision  *int
			scale      *int
			charLength int
			comment    *string
		)
		mustNoError(rows.Scan(&column, &dataType, &precision, &scale, &charLength, &comment))
		dataType = strings.ToUpper(dataType)
		if strings.HasPrefix(dataType, "TIMESTAMP") {
			dataType = "TIMESTAMP"
		}
		if comment == nil {
			comment = gdao.Ptr("")
		}

		f := &field{
			Column:    column,
			FieldName: fieldNameMapper.Convert(column),
			FieldType: "any",
			Comment:   *comment,
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

	rows = mustReturn(g.db.Query("SELECT comments FROM user_tab_comments WHERE table_name = :1", table))
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&tableComment)
	}

	return exists, fields, tableComment
}

func newOracleGenerator(c Cfg) oracleGenerator {
	db, err := sql.Open("oracle", c.Dsn)
	if err != nil { // coverage-ignore
		panic(fmt.Sprintf("connect db fail, dsn: %s, error: %v", c.Dsn, err))
	}
	t := mustReturn(template.New("").Parse(oracleBaseDaoTpl))
	return oracleGenerator{baseDaoTemplate: t, c: c, db: db}
}
