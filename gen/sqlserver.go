package gen

import (
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/jishaocong0910/gdao"
	_ "github.com/microsoft/go-mssqldb"
)

//go:embed sqlserver_base_dao.tpl
var sqlserverBaseDaoTpl string

type sqlServerGenerator struct {
	baseDaoTemplate *template.Template
	c               Cfg
	db              *sql.DB
}

func (g sqlServerGenerator) getBaseDaoTemplate() *template.Template {
	return g.baseDaoTemplate
}

func (g sqlServerGenerator) getTableInfo(table string) (bool, []*field, string) {
	var (
		exists       bool
		fields       []*field
		tableComment string
	)

	rows := mustReturn(g.db.Query("SELECT B.name, C.increment_value, D.data_type, E.value AS comment FROM sys.tables A LEFT JOIN sys.columns B ON A.object_id = B.object_id LEFT JOIN sys.identity_columns C ON A.object_id =C.object_id and B.name=C.name LEFT JOIN information_schema.columns D ON B.name = D.column_name LEFT JOIN sys.extended_properties E ON B.object_id = E.major_id AND B.column_id = E.minor_id WHERE A.name = :1 AND D.table_name = :2 ORDER BY D.ordinal_position", table, table))
	defer rows.Close()
	for rows.Next() {
		exists = true
		var (
			column         string
			dataType       string
			comment        *string
			fieldType      string
			incrementValue *int
		)
		mustNoError(rows.Scan(&column, &incrementValue, &dataType, &comment))
		dataType = strings.ToLower(dataType)
		if comment == nil {
			comment = gdao.Ptr("")
		}
		if incrementValue == nil {
			incrementValue = gdao.Ptr(0)
		}

		fieldType = "any"
		switch dataType {
		case "tinyint":
			fieldType = "*uint"
		case "smallint":
			fieldType = "*int16"
		case "int":
			fieldType = "*int32"
		case "bigint":
			fieldType = "*int64"
		case "bit":
			fieldType = "*bool"
		case "decimal", "numeric", "money", "smallmoney", "float":
			fieldType = "*float64"
		case "real":
			fieldType = "*float32"
		case "date", "time", "datetime2", "datetimeoffset", "datetime", "smalldatetime":
			fieldType = "*time.Time"
		case "char", "varchar", "text", "nchar", "nvarchar", "ntext", "xml", "json":
			fieldType = "*string"
		case "binary", "varbinary", "image", "geography", "geometry", "hierarchyid", "uniqueidentifier":
			fieldType = "[]byte"
		}

		f := &field{
			Column:            column,
			FieldName:         fieldNameMapper.Convert(column),
			FieldType:         fieldType,
			Comment:           *comment,
			Valid:             fieldType != "any",
			IsAutoIncrement:   *incrementValue > 0,
			AutoIncrementStep: *incrementValue,
		}
		fields = append(fields, f)
	}

	rows = mustReturn(g.db.Query("SELECT value FROM sys.extended_properties WHERE major_id = object_id (:1) AND minor_id = 0", table))
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&tableComment)
	}

	return exists, fields, tableComment
}

func newSqlServerGenerator(c Cfg) sqlServerGenerator {
	db, err := sql.Open("mssql", c.Dsn)
	if err != nil { // coverage-ignore
		panic(fmt.Sprintf("connect db fail, dsn: %s, error: %v", c.Dsn, err))
	}
	t := mustReturn(template.New("").Parse(sqlserverBaseDaoTpl))
	return sqlServerGenerator{baseDaoTemplate: t, c: c, db: db}
}
