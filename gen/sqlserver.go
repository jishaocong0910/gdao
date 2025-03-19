package gen

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jishaocong0910/gdao"
	_ "github.com/microsoft/go-mssqldb"
)

type sqlServerGenerator struct {
	c  Config
	db *sql.DB
}

func (g sqlServerGenerator) getTableInfo(table string) (bool, []*field, string) {
	var (
		exists       bool
		fields       []*field
		tableComment string
	)

	rows := mustReturn(g.db.Query("SELECT B.name AS column_name, C.DATA_TYPE AS data_type, D.value AS comment FROM sys.tables A LEFT JOIN sys.columns B ON A.object_id = B.object_id LEFT JOIN information_schema.columns C ON B.name = C.column_name LEFT JOIN sys.extended_properties D ON B.object_id = D.major_id AND B.column_id = D.minor_id WHERE A.name = :1", table))
	defer rows.Close()
	for rows.Next() {
		exists = true
		var (
			column    string
			dataType  string
			comment   *string
			fieldType string
		)
		mustNoError(rows.Scan(&column, &dataType, &comment))
		dataType = strings.ToLower(dataType)
		if comment == nil {
			comment = gdao.Ptr("")
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
			Column:    column,
			FieldName: g.c.FieldNameMapper.Convert(column),
			FieldType: fieldType,
			Comment:   *comment,
			Valid:     fieldType != "any",
		}
		fields = append(fields, f)
	}

	rows = mustReturn(g.db.Query("SELECT * FROM sys.extended_properties WHERE major_id = object_id (:1) AND minor_id = 0", table))
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&tableComment)
	}

	return exists, fields, tableComment
}

func newSqlServerGenerator(c Config) sqlServerGenerator {
	db, err := sql.Open("mssql", c.Dsn)
	if err != nil { // coverage-ignore
		panic(fmt.Sprintf("connect db fail, dsn: %s, error: %v", c.Dsn, err))
	}
	return sqlServerGenerator{c: c, db: db}
}
