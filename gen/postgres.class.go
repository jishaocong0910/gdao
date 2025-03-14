package gen

import (
	"github.com/jishaocong0910/gdao"
	_ "github.com/lib/pq"
)

type postgresGenerator struct {
	*m_Generator
	schema   string
	database string
}

func (this *postgresGenerator) existsTable(table string) bool {
	rows := mustReturn(this.db.Query("SELECT count(*) FROM information_schema.tables WHERE table_catalog = $1 AND table_schema = $2 AND table_name = $3", this.database, this.schema, table))
	defer rows.Close()
	var count int
	if rows.Next() {
		rows.Scan(&count)
	}
	return count == 1
}

func (this *postgresGenerator) getEntityComment(table string) string {
	rows := mustReturn(this.db.Query("SELECT d.description FROM information_schema.tables i JOIN pg_class c ON c.relname = i.table_name LEFT JOIN pg_description d ON d.objoid = c.oid AND d.objsubid = '0' WHERE i.table_catalog = $1 AND i.table_schema = $2 AND i.table_name = $3", this.database, this.schema, table))
	defer rows.Close()
	var comment string
	if rows.Next() {
		rows.Scan(&comment)
	}
	return comment
}

func (this *postgresGenerator) getEntityFields(table string) []*field {
	rows := mustReturn(this.db.Query("SELECT i.column_name,i.udt_name,p.attndims,p.description FROM information_schema.columns i JOIN (SELECT n.nspname,c.relname,a.attname,a.attndims,d.description FROM pg_namespace n JOIN pg_class c ON n.oid = c.relnamespace JOIN pg_attribute a ON a.attrelid = c.oid LEFT JOIN pg_description d ON d.objoid = c.oid AND d.objsubid = a.attnum WHERE a.attnum > 0 AND NOT a.attisdropped) p ON p.nspname=i.table_schema and p.relname=i.table_name and p.attname=i.column_name WHERE i.table_catalog = $1 AND i.table_schema = $2 AND i.table_name = $3", this.database, this.schema, table))
	defer rows.Close()
	var fields []*field
	for rows.Next() {
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
		case "date", "timestamp", "timestamptz":
			fieldType = "*time.Time"
		case "float8", "numeric":
			fieldType = "*float64"
		case "float4":
			fieldType = "*float32"
		case "bool":
			fieldType = "*bool"
		case "bytea":
			fieldType = "[]uint8"
		case "bit", "varbit", "box", "bpchar", "varchar", "cidr", "circle", "inet", "interval", "json", "jsonb", "line",
			"lseg", "macaddr", "macaddr8", "money", "path", "pg_lsn", "pg_snapshot", "point", "polygon", "text", "time",
			"timetz", "tsquery", "tsvector", "txid_snapshot", "uuid", "xml":
			fieldType = "*string"
		}
		for i := 0; i < attndims; i++ {
			fieldType = "[]" + fieldType
		}

		f := &field{
			Column:    column,
			FieldName: this.c.FieldNameMapper.Convert(column),
			FieldType: fieldType,
			Comment:   *description,
			Valid:     fieldType != "any",
		}
		fields = append(fields, f)
	}
	return fields
}

func newPostgresGenerator(c Config) *postgresGenerator {
	g := &postgresGenerator{}
	g.m_Generator = extendGenerator(g, c)

	rows := mustReturn(g.db.Query("SELECT CURRENT_SCHEMA()"))
	if rows.Next() {
		rows.Scan(&g.schema)
	}
	rows.Close()

	rows = mustReturn(g.db.Query("SELECT CURRENT_DATABASE()"))
	if rows.Next() {
		rows.Scan(&g.database)
	}
	rows.Close()
	return g
}
