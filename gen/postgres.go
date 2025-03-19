package gen

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jishaocong0910/gdao"
	_ "github.com/lib/pq"
)

type postgresGenerator struct {
	c        Config
	db       *sql.DB
	schema   string
	database string
}

func (g postgresGenerator) getTableInfo(table string) (bool, []*field, string) {
	var (
		exists       bool
		fields       []*field
		tableComment string
	)

	rows := mustReturn(g.db.Query("SELECT i.column_name,i.udt_name,p.attndims,p.description FROM information_schema.columns i JOIN (SELECT n.nspname,c.relname,a.attname,a.attndims,d.description FROM pg_namespace n JOIN pg_class c ON n.oid = c.relnamespace JOIN pg_attribute a ON a.attrelid = c.oid LEFT JOIN pg_description d ON d.objoid = c.oid AND d.objsubid = a.attnum WHERE a.attnum > 0 AND NOT a.attisdropped) p ON p.nspname=i.table_schema and p.relname=i.table_name and p.attname=i.column_name WHERE i.table_catalog = $1 AND i.table_schema = $2 AND i.table_name = $3", g.database, g.schema, table))
	defer rows.Close()
	for rows.Next() {
		exists = true
		var (
			column      string
			udtName     string
			attndims    int
			description *string
			fieldType   string
		)
		mustNoError(rows.Scan(&column, &udtName, &attndims, &description))
		if description == nil {
			description = gdao.Ptr("")
		}
		if udtName[:1] == "_" {
			udtName = udtName[1:]
		}
		switch udtName {
		case "int8":
			fieldType = "*int64"
		case "int4":
			fieldType = "*int32"
		case "int2":
			fieldType = "*int16"
		case "time", "timetz", "date", "timestamp", "timestamptz":
			fieldType = "*time.Time"
		case "float8", "numeric", "money":
			fieldType = "*float64"
		case "float4":
			fieldType = "*float32"
		case "bool":
			fieldType = "*bool"
		case "bytea":
			fieldType = "[]byte"
		case "bit", "varbit", "box", "bpchar", "varchar", "cidr", "circle", "inet", "interval", "json", "jsonb", "line",
			"lseg", "macaddr", "macaddr8", "path", "pg_lsn", "pg_snapshot", "point", "polygon", "text",
			"tsquery", "tsvector", "txid_snapshot", "uuid", "xml":
			fieldType = "*string"
		}
		for i := 0; i < attndims; i++ {
			if i == 0 {
				fieldType = strings.TrimPrefix(fieldType, "*")
			}
			fieldType = "[]" + fieldType
		}

		f := &field{
			Column:    column,
			FieldName: g.c.FieldNameMapper.Convert(column),
			FieldType: fieldType,
			Comment:   *description,
			Valid:     fieldType != "any",
		}
		fields = append(fields, f)
	}

	rows = mustReturn(g.db.Query("SELECT d.description FROM information_schema.tables i JOIN pg_class c ON c.relname = i.table_name LEFT JOIN pg_description d ON d.objoid = c.oid AND d.objsubid = '0' WHERE i.table_catalog = $1 AND i.table_schema = $2 AND i.table_name = $3", g.database, g.schema, table))
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&tableComment)
	}

	return exists, fields, tableComment
}

func newPostgresGenerator(c Config) postgresGenerator {
	db, err := sql.Open("postgres", c.Dsn)
	if err != nil { // coverage-ignore
		panic(fmt.Sprintf("connect db fail, dsn: %s, error: %v", c.Dsn, err))
	}

	schema := ""
	rows := mustReturn(db.Query("SELECT CURRENT_SCHEMA()"))
	if rows.Next() {
		rows.Scan(&schema)
	}
	rows.Close()

	database := ""
	rows = mustReturn(db.Query("SELECT CURRENT_DATABASE()"))
	if rows.Next() {
		rows.Scan(&database)
	}
	rows.Close()
	return postgresGenerator{c: c, db: db, schema: schema, database: database}
}
